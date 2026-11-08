package webapp

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"boki-webapps/internal/browser"
)

func TestSaveAndRead(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	app, err := Save(App{
		Name:        "Youtube",
		URL:         "https://youtube.com",
		BrowserName: "Brave Origin",
		BrowserExec: "brave-origin",
		Icon:        "brave-origin",
		Category:    "Network",
		Isolated:    true,
		Kind:        browser.KindChromium,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`WebApp-Youtube\d{4}\.desktop$`).MatchString(app.Path) {
		t.Fatal(app.Path)
	}
	body, err := os.ReadFile(app.Path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "--app=\"https://youtube.com\"") || !strings.Contains(text, "--user-data-dir=") {
		t.Fatal(text)
	}
	got, err := Read(app.Path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Youtube" || got.URL != "https://youtube.com" || got.BrowserExec != "brave-origin" {
		t.Fatalf("%+v", got)
	}
	listed, err := List()
	if err != nil || len(listed) != 1 {
		t.Fatalf("list %v %v", listed, err)
	}
	if err := Delete(got); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(app.Path); !os.IsNotExist(err) {
		t.Fatal("desktop file still exists")
	}
	if _, err := os.Stat(filepath.Join(home, ".local/share/ice/profiles", got.Codename)); !os.IsNotExist(err) {
		t.Fatal("profile still exists")
	}
}

func TestFirefoxNavbarToggle(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	app, err := Save(App{
		Name:        "Docs",
		URL:         "https://example.com",
		BrowserName: "Firefox",
		BrowserExec: "firefox",
		Kind:        browser.KindFirefox,
	})
	if err != nil {
		t.Fatal(err)
	}
	cssPath := filepath.Join(home, ".local/share/ice/firefox", app.Codename, "chrome", "userChrome.css")
	body, err := os.ReadFile(cssPath)
	if err != nil || strings.TrimSpace(string(body)) == "" {
		t.Fatalf("hidden navbar css: %v %q", err, body)
	}
	desktop, err := os.ReadFile(app.Path)
	if err != nil || !strings.Contains(string(desktop), "--profile ") {
		t.Fatalf("exec: %v %s", err, desktop)
	}
	app.Navbar = true
	if _, err := Save(app); err != nil {
		t.Fatal(err)
	}
	body, err = os.ReadFile(cssPath)
	if err != nil || len(strings.TrimSpace(string(body))) != 0 {
		t.Fatalf("navbar on: %v %q", err, body)
	}
	app.Navbar = false
	app.Private = true
	if _, err := Save(app); err != nil {
		t.Fatal(err)
	}
	body, err = os.ReadFile(cssPath)
	if err != nil || !strings.Contains(string(body), "#nav-bar") {
		t.Fatalf("navbar off: %v %q", err, body)
	}
	desktop, err = os.ReadFile(app.Path)
	if err != nil || !strings.Contains(string(desktop), "--private-window") {
		t.Fatalf("private: %v %s", err, desktop)
	}
}
