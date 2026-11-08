package icon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCustomImages(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "youtube.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("no"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := CustomImages()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !strings.HasSuffix(got[0].Path, "youtube.png") {
		t.Fatalf("%+v", got)
	}
}

func TestLinksFromHTML(t *testing.T) {
	html := `<html><head>
<link rel="stylesheet" href="/site.css">
<link rel="apple-touch-icon" href="/touch.png">
<link rel="icon" href="https://cdn.example/icon.png">
<meta property="og:image" content="/og.jpg">
</head></html>`
	got := linksFromHTML("https://youtube.com/feed", html)
	want := []string{
		"https://youtube.com/touch.png",
		"https://cdn.example/icon.png",
		"https://youtube.com/og.jpg",
	}
	if len(got) != len(want) {
		t.Fatalf("%+v", got)
	}
	for i := range want {
		if got[i].URL != want[i] {
			t.Fatalf("got %s want %s", got[i].URL, want[i])
		}
	}
	if got[0].Source != "apple-touch-icon" || got[0].Score < got[1].Score {
		t.Fatalf("%+v", got)
	}
}

func TestNormalize(t *testing.T) {
	if Normalize("youtube.com") != "https://youtube.com" {
		t.Fatal(Normalize("youtube.com"))
	}
	if Normalize("file:///tmp/x") != "" || Normalize("https://you") != "" {
		t.Fatal("rejected url")
	}
}

func TestSavePNGAndICO(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	png := []byte("\x89PNG\r\n\x1a\nrest")
	path, ok := saveImage(mustDir(t), png)
	if !ok || !strings.HasSuffix(path, ".png") {
		t.Fatal(path, ok)
	}
	ico := make([]byte, 22)
	ico[2] = 1
	ico[4] = 1
	off := 22
	ico[18] = byte(off)
	body := append(ico, png...)
	got, ok := pngInsideICO(body)
	if !ok || string(got) != string(png) {
		t.Fatalf("%v %q", ok, got)
	}
}

func mustDir(t *testing.T) string {
	t.Helper()
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir)); err != nil {
		t.Fatal(err)
	}
	return dir
}
