package browser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFromDesktopFindsBrowserAndSkipsActions(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "brave-origin")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	desktop := filepath.Join(dir, "brave.desktop")
	body := "[Desktop Entry]\nName=Brave Origin\nExec=" + bin + " %U\nCategories=Network;WebBrowser;\n\n[Desktop Action new]\nName=New Window\nExec=/no/such/binary\n"
	if err := os.WriteFile(desktop, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := fromDesktop(desktop)
	if !ok {
		t.Fatal("browser was not detected")
	}
	if got.Name != "Brave Origin" || got.Exec != bin || got.Kind != KindChromium {
		t.Fatalf("got %+v", got)
	}
}

func TestFromDesktopSkipsWebAppAndEditors(t *testing.T) {
	dir := t.TempDir()
	webapp := filepath.Join(dir, "WebApp-Youtube1234.desktop")
	if err := os.WriteFile(webapp, []byte("[Desktop Entry]\nName=Youtube\nExec=/usr/bin/true\nCategories=Network;WebBrowser;\nX-WebApp-URL=https://youtube.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := fromDesktop(webapp); ok {
		t.Fatal("created web app was treated as a browser")
	}

	editor := filepath.Join(dir, "editor.desktop")
	if err := os.WriteFile(editor, []byte("[Desktop Entry]\nName=Editor\nExec=/usr/bin/true\nMimeType=text/html;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := fromDesktop(editor); ok {
		t.Fatal("text editor was treated as a browser")
	}

	shortcut := filepath.Join(dir, "ChatGPT.desktop")
	if err := os.WriteFile(shortcut, []byte("[Desktop Entry]\nName=ChatGPT\nExec=/usr/bin/true --app=https://chatgpt.com\nCategories=Network;WebBrowser;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := fromDesktop(shortcut); ok {
		t.Fatal("site shortcut was treated as a browser")
	}
}

func TestLiveScanFindsInstalledBrowsers(t *testing.T) {
	if os.Getenv("BOKI_LIVE_SCAN") != "1" {
		t.Skip("set BOKI_LIVE_SCAN=1 to check this machine")
	}
	found, err := Scan()
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range found {
		t.Log(b.Name, b.Exec, b.Kind)
	}
	for _, part := range []string{"firefox", "chromium", "vivaldi", "brave-origin"} {
		ok := false
		for _, b := range found {
			if strings.Contains(strings.ToLower(b.Exec), part) {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("missing %s", part)
		}
	}
}

func TestKindOf(t *testing.T) {
	if kindOf("/usr/lib/firefox/firefox") != KindFirefox {
		t.Fatal("firefox")
	}
	if kindOf("brave-origin") != KindChromium {
		t.Fatal("brave")
	}
	if kindOf("flatpak run org.gnome.Epiphany") != KindEpiphany {
		t.Fatal("epiphany")
	}
}
