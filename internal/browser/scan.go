package browser

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Kind int

const (
	KindChromium Kind = iota
	KindFirefox
	KindEpiphany
	KindFalkon
)

func (k Kind) String() string {
	switch k {
	case KindFirefox:
		return "firefox"
	case KindEpiphany:
		return "epiphany"
	case KindFalkon:
		return "falkon"
	default:
		return "chromium"
	}
}

type Browser struct {
	Name   string
	Exec   string
	Icon   string
	Kind   Kind
	Source string
}

func Scan() ([]Browser, error) {
	dirs := applicationDirs()
	seen := map[string]bool{}
	var found []Browser
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			b, ok := fromDesktop(path)
			if !ok {
				continue
			}
			key := b.Exec
			if seen[key] {
				continue
			}
			seen[key] = true
			found = append(found, b)
		}
	}
	return found, nil
}

func applicationDirs() []string {
	home, _ := os.UserHomeDir()
	var dirs []string
	if home != "" {
		dirs = append(dirs,
			filepath.Join(home, ".local/share/applications"),
			filepath.Join(home, ".local/share/flatpak/exports/share/applications"),
		)
	}
	dataDirs := os.Getenv("XDG_DATA_DIRS")
	if dataDirs == "" {
		dataDirs = "/usr/local/share:/usr/share"
	}
	for _, dir := range strings.Split(dataDirs, ":") {
		if dir == "" {
			continue
		}
		dirs = append(dirs, filepath.Join(dir, "applications"))
	}
	dirs = append(dirs,
		"/var/lib/flatpak/exports/share/applications",
		"/usr/share/applications",
		"/usr/local/share/applications",
	)
	return unique(dirs)
}

func fromDesktop(path string) (Browser, bool) {
	fields, err := desktopEntry(path)
	if err != nil {
		return Browser{}, false
	}
	if fields["Hidden"] == "true" || fields["NoDisplay"] == "true" {
		return Browser{}, false
	}
	if fields["X-WebApp-URL"] != "" || strings.HasPrefix(filepath.Base(path), "WebApp-") {
		return Browser{}, false
	}
	if !isBrowser(fields) {
		return Browser{}, false
	}
	execLine := commandLine(fields["Exec"])
	if execLine == "" || isSiteShortcut(execLine) {
		return Browser{}, false
	}
	if !commandExists(execLine) {
		return Browser{}, false
	}
	name := fields["Name"]
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(path), ".desktop")
	}
	return Browser{
		Name:   name,
		Exec:   execLine,
		Icon:   fields["Icon"],
		Kind:   kindOf(execLine),
		Source: path,
	}, true
}

func isSiteShortcut(execLine string) bool {
	low := strings.ToLower(execLine)
	return strings.Contains(low, "--app=") ||
		strings.Contains(low, "--app ") ||
		strings.Contains(low, "--application-mode")
}

func isBrowser(fields map[string]string) bool {
	cats := ";" + fields["Categories"] + ";"
	if strings.Contains(cats, ";WebBrowser;") {
		return true
	}
	mime := fields["MimeType"]
	return strings.Contains(mime, "x-scheme-handler/http") || strings.Contains(mime, "x-scheme-handler/https")
}

func desktopEntry(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fields := map[string]string{}
	inEntry := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inEntry = line == "[Desktop Entry]"
			continue
		}
		if !inEntry {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok || strings.Contains(key, "[") {
			continue
		}
		if _, exists := fields[key]; !exists {
			fields[key] = val
		}
	}
	return fields, sc.Err()
}

func commandLine(exec string) string {
	var parts []string
	for _, tok := range splitExec(exec) {
		if tok == "" || strings.HasPrefix(tok, "%") {
			continue
		}
		parts = append(parts, tok)
	}
	return strings.Join(parts, " ")
}

func splitExec(s string) []string {
	var out []string
	var b strings.Builder
	quote := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
				continue
			}
			if c == '\\' && i+1 < len(s) {
				i++
				b.WriteByte(s[i])
				continue
			}
			b.WriteByte(c)
		case c == '\'' || c == '"':
			quote = c
		case c == ' ' || c == '\t':
			out = append(out, b.String())
			b.Reset()
		default:
			b.WriteByte(c)
		}
	}
	if b.Len() > 0 || len(out) == 0 {
		out = append(out, b.String())
	}
	return out
}

func commandExists(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	bin := fields[0]
	if strings.Contains(bin, "/") {
		info, err := os.Stat(bin)
		return err == nil && !info.IsDir()
	}
	_, err := exec.LookPath(bin)
	return err == nil
}

func kindOf(execLine string) Kind {
	low := strings.ToLower(execLine)
	switch {
	case strings.Contains(low, "epiphany") || strings.Contains(low, "org.gnome.epiphany"):
		return KindEpiphany
	case strings.Contains(low, "falkon"):
		return KindFalkon
	case strings.Contains(low, "firefox"),
		strings.Contains(low, "librewolf"),
		strings.Contains(low, "waterfox"),
		strings.Contains(low, "floorp"),
		strings.Contains(low, "zen-browser"),
		strings.Contains(low, "app.zen_browser"):
		return KindFirefox
	default:
		return KindChromium
	}
}

func unique(items []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, item := range items {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}
