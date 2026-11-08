package icon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Image struct {
	Source string
	Path   string
	Score  int
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local/share/boki-webapps/icons")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func CustomImages() ([]Image, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var found []Image
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(entry.Name())) {
		case ".png", ".jpg", ".jpeg", ".svg", ".webp", ".gif":
			found = append(found, Image{
				Source: entry.Name(),
				Path:   filepath.Join(dir, entry.Name()),
			})
		}
	}
	for i := 1; i < len(found); i++ {
		j := i
		for j > 0 && found[j].Source < found[j-1].Source {
			found[j], found[j-1] = found[j-1], found[j]
			j--
		}
	}
	return found, nil
}

func Lookup(ctx context.Context, page string) ([]Image, error) {
	page = Normalize(page)
	if page == "" {
		return nil, fmt.Errorf("url must start with http:// or https://")
	}
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 12 * time.Second}
	links := []candidate{{URL: originOf(page) + "/favicon.ico", Source: "favicon", Score: 20}}
	if body, err := get(ctx, client, page, 1<<20); err == nil {
		links = append(linksFromHTML(page, string(body)), links...)
	}
	if host := hostOf(page); host != "" {
		links = append(links, candidate{
			URL:    "https://www.google.com/s2/favicons?sz=128&domain=" + url.QueryEscape(host),
			Source: "fallback",
			Score:  10,
		})
	}
	sortCandidates(links)

	var found []Image
	seen := map[string]bool{}
	for _, link := range links {
		if len(found) >= 6 || ctx.Err() != nil {
			break
		}
		if link.URL == "" || seen[link.URL] {
			continue
		}
		seen[link.URL] = true
		body, err := get(ctx, client, link.URL, 2<<20)
		if err != nil {
			continue
		}
		path, ok := saveImage(dir, body)
		if !ok {
			continue
		}
		pixels := area(path)
		if pixels > 256*256 {
			pixels = 256 * 256
		}
		found = append(found, Image{
			Source: link.Source,
			Path:   path,
			Score:  link.Score*100000 + pixels,
		})
	}
	sortImages(found)
	return found, nil
}

type candidate struct {
	URL    string
	Source string
	Score  int
}

func linksFromHTML(page, html string) []candidate {
	base, err := url.Parse(page)
	if err != nil {
		return nil
	}
	var out []candidate
	rest := html
	for {
		start := strings.IndexByte(rest, '<')
		if start < 0 {
			break
		}
		rest = rest[start:]
		end := strings.IndexByte(rest, '>')
		if end < 0 {
			break
		}
		tag := rest[:end+1]
		rest = rest[end+1:]
		name, attrs := parseTag(tag)
		switch name {
		case "link":
			rel := strings.ToLower(attrs["rel"])
			href := resolve(base, attrs["href"])
			if href == "" || !strings.Contains(rel, "icon") {
				continue
			}
			score := 70
			source := "icon"
			if strings.Contains(rel, "apple") {
				score = 100
				source = "apple-touch-icon"
			}
			out = append(out, candidate{URL: href, Source: source, Score: score})
		case "meta":
			key := strings.ToLower(attrs["property"] + " " + attrs["name"])
			content := resolve(base, attrs["content"])
			if content == "" {
				continue
			}
			if strings.Contains(key, "og:image") || strings.Contains(key, "tileimage") || strings.Contains(key, "logo") {
				out = append(out, candidate{URL: content, Source: "page-image", Score: 40})
			}
		}
	}
	return out
}

func parseTag(tag string) (string, map[string]string) {
	s := strings.TrimPrefix(tag, "<")
	s = strings.TrimSuffix(s, ">")
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "!") {
		return "", nil
	}
	name, s := takeToken(s)
	attrs := map[string]string{}
	for s != "" {
		var key string
		key, s = takeToken(s)
		if key == "" || key == "/" {
			break
		}
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, "=") {
			attrs[strings.ToLower(key)] = ""
			continue
		}
		s = strings.TrimSpace(s[1:])
		var val string
		val, s = takeValue(s)
		attrs[strings.ToLower(key)] = val
	}
	return strings.ToLower(name), attrs
}

func takeToken(s string) (string, string) {
	s = strings.TrimSpace(s)
	i := 0
	for i < len(s) && s[i] != ' ' && s[i] != '\t' && s[i] != '=' && s[i] != '/' && s[i] != '>' {
		i++
	}
	return s[:i], strings.TrimSpace(s[i:])
}

func takeValue(s string) (string, string) {
	if s == "" {
		return "", s
	}
	if s[0] == '"' || s[0] == '\'' {
		q := s[0]
		s = s[1:]
		i := strings.IndexByte(s, q)
		if i < 0 {
			return s, ""
		}
		return s[:i], strings.TrimSpace(s[i+1:])
	}
	i := 0
	for i < len(s) && s[i] != ' ' && s[i] != '\t' && s[i] != '>' {
		i++
	}
	return s[:i], strings.TrimSpace(s[i:])
}

func resolve(base *url.URL, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.HasPrefix(ref, "data:") {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(u)
	if abs.Scheme != "http" && abs.Scheme != "https" {
		return ""
	}
	return abs.String()
}

func Normalize(page string) string {
	page = strings.TrimSpace(page)
	if page == "" {
		return ""
	}
	if !strings.Contains(page, "://") {
		page = "https://" + page
	}
	u, err := url.Parse(page)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	if !strings.Contains(u.Hostname(), ".") {
		return ""
	}
	return u.String()
}

func originOf(page string) string {
	u, err := url.Parse(page)
	if err != nil {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func hostOf(page string) string {
	u, err := url.Parse(page)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func get(ctx context.Context, client *http.Client, raw string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Boki-WebApps")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, limit))
}

func saveImage(dir string, body []byte) (string, bool) {
	ext, data, ok := sniff(body)
	if !ok {
		return "", false
	}
	sum := sha256.Sum256(data)
	path := filepath.Join(dir, hex.EncodeToString(sum[:8])+ext)
	if _, err := os.Stat(path); err == nil {
		return path, true
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", false
	}
	return path, true
}

func sniff(body []byte) (string, []byte, bool) {
	if len(body) >= 8 && string(body[:8]) == "\x89PNG\r\n\x1a\n" {
		return ".png", body, true
	}
	if len(body) >= 3 && body[0] == 0xff && body[1] == 0xd8 && body[2] == 0xff {
		return ".jpg", body, true
	}
	if len(body) >= 6 && (string(body[:6]) == "GIF87a" || string(body[:6]) == "GIF89a") {
		return ".gif", body, true
	}
	if len(body) >= 12 && string(body[:4]) == "RIFF" && string(body[8:12]) == "WEBP" {
		return ".webp", body, true
	}
	trim := strings.TrimSpace(string(body))
	low := strings.ToLower(trim)
	if strings.Contains(low, "<svg") && !strings.Contains(trim, "\x00") && len(body) < 500_000 {
		return ".svg", []byte(trim), true
	}
	if png, ok := pngInsideICO(body); ok {
		return ".png", png, true
	}
	return "", nil, false
}

func pngInsideICO(body []byte) ([]byte, bool) {
	if len(body) < 22 || body[0] != 0 || body[1] != 0 || body[2] != 1 {
		return nil, false
	}
	count := int(body[4]) | int(body[5])<<8
	best := []byte(nil)
	for i := 0; i < count; i++ {
		off := 6 + i*16
		if off+16 > len(body) {
			break
		}
		start := int(body[off+12]) | int(body[off+13])<<8 | int(body[off+14])<<16 | int(body[off+15])<<24
		if start < 0 || start+8 > len(body) {
			continue
		}
		if string(body[start:start+8]) != "\x89PNG\r\n\x1a\n" {
			continue
		}
		next := len(body)
		if i+1 < count {
			noff := 6 + (i+1)*16
			if noff+16 <= len(body) {
				n := int(body[noff+12]) | int(body[noff+13])<<8 | int(body[noff+14])<<16 | int(body[noff+15])<<24
				if n > start && n <= len(body) {
					next = n
				}
			}
		}
		chunk := body[start:next]
		if len(chunk) > len(best) {
			best = chunk
		}
	}
	if best == nil {
		return nil, false
	}
	return append([]byte(nil), best...), true
}

func area(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 {
		return 0
	}
	return cfg.Width * cfg.Height
}

func sortCandidates(links []candidate) {
	for i := 1; i < len(links); i++ {
		j := i
		for j > 0 && links[j].Score > links[j-1].Score {
			links[j], links[j-1] = links[j-1], links[j]
			j--
		}
	}
}

func sortImages(found []Image) {
	for i := 1; i < len(found); i++ {
		j := i
		for j > 0 && found[j].Score > found[j-1].Score {
			found[j], found[j-1] = found[j-1], found[j]
			j--
		}
	}
}
