package webapp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"boki-webapps/internal/browser"
)

type App struct {
	Path        string
	Codename    string
	Name        string
	URL         string
	Icon        string
	Category    string
	BrowserName string
	BrowserExec string
	Parameters  string
	Navbar      bool
	Private     bool
	Isolated    bool
	Kind        browser.Kind
}

func AppsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local/share/applications")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func List() ([]App, error) {
	dir, err := AppsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var apps []App
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "WebApp-") || !strings.HasSuffix(name, ".desktop") {
			continue
		}
		app, err := Read(filepath.Join(dir, name))
		if err != nil || app.URL == "" {
			continue
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func Read(path string) (App, error) {
	f, err := os.Open(path)
	if err != nil {
		return App{}, err
	}
	defer f.Close()
	fields := map[string]string{}
	inEntry := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "[") {
			inEntry = line == "[Desktop Entry]"
			continue
		}
		if !inEntry {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if ok {
			fields[key] = val
		}
	}
	if err := sc.Err(); err != nil {
		return App{}, err
	}
	base := strings.TrimSuffix(filepath.Base(path), ".desktop")
	codename := strings.TrimPrefix(base, "WebApp-")
	kind := browser.KindChromium
	switch fields["X-WebApp-Kind"] {
	case "firefox":
		kind = browser.KindFirefox
	case "epiphany":
		kind = browser.KindEpiphany
	case "falkon":
		kind = browser.KindFalkon
	}
	return App{
		Path:        path,
		Codename:    codename,
		Name:        fields["Name"],
		URL:         fields["X-WebApp-URL"],
		Icon:        fields["Icon"],
		Category:    categoryFrom(fields["Categories"]),
		BrowserName: fields["X-WebApp-Browser"],
		BrowserExec: fields["X-WebApp-BrowserExec"],
		Parameters:  fields["X-WebApp-CustomParameters"],
		Navbar:      fields["X-WebApp-Navbar"] == "true",
		Private:     fields["X-WebApp-PrivateWindow"] == "true",
		Isolated:    fields["X-WebApp-Isolated"] != "false",
		Kind:        kind,
	}, nil
}

func Save(app App) (App, error) {
	if strings.TrimSpace(app.Name) == "" {
		return App{}, fmt.Errorf("name is required")
	}
	if !strings.HasPrefix(app.URL, "http://") && !strings.HasPrefix(app.URL, "https://") {
		return App{}, fmt.Errorf("url must start with http:// or https://")
	}
	if strings.TrimSpace(app.BrowserExec) == "" {
		return App{}, fmt.Errorf("choose a browser")
	}
	if strings.ContainsAny(app.URL, "\"\n") || strings.ContainsAny(app.Name, "\n") {
		return App{}, fmt.Errorf("name and url cannot contain quotes or newlines")
	}
	for _, field := range []string{app.Icon, app.Category, app.Parameters, app.BrowserName, app.BrowserExec} {
		if strings.ContainsAny(field, "\r\n") {
			return App{}, fmt.Errorf("fields cannot contain newlines")
		}
	}
	dir, err := AppsDir()
	if err != nil {
		return App{}, err
	}
	if app.Codename == "" {
		app.Codename = codename(app.Name)
	}
	if strings.Contains(app.Codename, "/") || strings.Contains(app.Codename, "..") {
		return App{}, fmt.Errorf("invalid web app id")
	}
	kept, err := keepIcon(app.Icon, app.Codename)
	if err != nil {
		return App{}, err
	}
	app.Icon = kept
	if app.Icon == "" {
		app.Icon = "web-browser"
	}
	if app.Category == "" {
		app.Category = "Network"
	}
	if err := prepareProfile(app); err != nil {
		return App{}, err
	}
	app.Path = filepath.Join(dir, "WebApp-"+app.Codename+".desktop")
	content := desktopFile(app)
	if err := os.WriteFile(app.Path, []byte(content), 0o644); err != nil {
		return App{}, err
	}
	refreshDesktop(dir)
	return app, nil
}

func Delete(app App) error {
	if app.Codename == "" || strings.Contains(app.Codename, "/") || strings.Contains(app.Codename, "..") {
		return fmt.Errorf("invalid web app id")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	ice := filepath.Join(home, ".local/share/ice")
	for _, dir := range []string{
		filepath.Join(ice, "profiles", app.Codename),
		filepath.Join(ice, "firefox", app.Codename),
		filepath.Join(ice, "falkon", app.Codename),
		filepath.Join(ice, "epiphany", "org.gnome.Epiphany.WebApp_"+app.Codename),
	} {
		_ = os.RemoveAll(dir)
	}
	if app.Path != "" {
		_ = os.Remove(app.Path)
		refreshDesktop(filepath.Dir(app.Path))
	}
	if home != "" {
		matches, _ := filepath.Glob(filepath.Join(home, ".local/share/boki-webapps/icons", app.Codename+".*"))
		for _, match := range matches {
			_ = os.Remove(match)
		}
	}
	return nil
}

func keepIcon(icon, codename string) (string, error) {
	icon = strings.TrimSpace(icon)
	if icon == "" || !filepath.IsAbs(icon) {
		return icon, nil
	}
	info, err := os.Stat(icon)
	if err != nil || info.IsDir() {
		return "", fmt.Errorf("icon file not found")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local/share/boki-webapps/icons")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(icon))
	if ext == "" {
		ext = ".png"
	}
	dest := filepath.Join(dir, codename+ext)
	if icon == dest {
		return dest, nil
	}
	in, err := os.Open(icon)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}
	return dest, nil
}

func Launch(app App) error {
	if app.Path == "" {
		return fmt.Errorf("web app has no desktop file")
	}
	return exec.Command("gio", "launch", app.Path).Start()
}

func desktopFile(app App) string {
	var b strings.Builder
	write := func(key, val string) {
		fmt.Fprintf(&b, "%s=%s\n", key, val)
	}
	b.WriteString("[Desktop Entry]\n")
	write("Version", "1.0")
	write("Name", app.Name)
	write("Comment", "Web App")
	write("Exec", execLine(app))
	write("Terminal", "false")
	write("Type", "Application")
	write("Icon", app.Icon)
	write("Categories", "GTK;"+app.Category+";")
	write("StartupWMClass", "WebApp-"+app.Codename)
	write("StartupNotify", "true")
	write("X-WebApp-Browser", app.BrowserName)
	write("X-WebApp-BrowserExec", app.BrowserExec)
	write("X-WebApp-URL", app.URL)
	write("X-WebApp-CustomParameters", app.Parameters)
	write("X-WebApp-Navbar", strconv.FormatBool(app.Navbar))
	write("X-WebApp-PrivateWindow", strconv.FormatBool(app.Private))
	write("X-WebApp-Isolated", strconv.FormatBool(app.Isolated))
	write("X-WebApp-Kind", app.Kind.String())
	return b.String()
}

func execLine(app App) string {
	class := "WebApp-" + app.Codename
	switch app.Kind {
	case browser.KindFirefox:
		profile := filepath.Join(mustHome(), ".local/share/ice/firefox", app.Codename)
		line := fmt.Sprintf("sh -c 'XAPP_FORCE_GTKWINDOW_ICON=%q %s --class %s --name %s --profile %s --no-remote", app.Icon, app.BrowserExec, class, class, profile)
		if app.Private {
			line += " --private-window"
		}
		if app.Parameters != "" {
			line += " " + app.Parameters
		}
		return line + " " + strconv.Quote(app.URL) + "'"
	case browser.KindEpiphany:
		profile := filepath.Join(mustHome(), ".local/share", "org.gnome.Epiphany.WebApp_"+app.Codename)
		line := fmt.Sprintf("%s --application-mode --profile=%s %s", app.BrowserExec, strconv.Quote(profile), strconv.Quote(app.URL))
		if app.Parameters != "" {
			line += " " + app.Parameters
		}
		return line
	case browser.KindFalkon:
		line := app.BrowserExec + " --wmclass=" + class
		if app.Isolated {
			line += " --profile=" + app.Codename
		}
		if app.Private {
			line += " --private-browsing"
		}
		if app.Parameters != "" {
			line += " " + app.Parameters
		}
		return line + " --no-remote " + strconv.Quote(app.URL)
	default:
		line := fmt.Sprintf("%s --app=%s --class=%s --name=%s", app.BrowserExec, strconv.Quote(app.URL), class, class)
		if app.Isolated {
			profile := filepath.Join(mustHome(), ".local/share/ice/profiles", app.Codename)
			line += " --user-data-dir=" + profile
		}
		if app.Private {
			if strings.Contains(strings.ToLower(app.BrowserName), "edge") {
				line += " --inprivate"
			} else {
				line += " --incognito"
			}
		}
		if app.Parameters != "" {
			line += " " + app.Parameters
		}
		return line
	}
}

func prepareProfile(app App) error {
	home := mustHome()
	switch app.Kind {
	case browser.KindFirefox:
		dst := filepath.Join(home, ".local/share/ice/firefox", app.Codename)
		chromeDir := filepath.Join(dst, "chrome")
		if err := os.MkdirAll(chromeDir, 0o755); err != nil {
			return err
		}
		userJS := filepath.Join(dst, "user.js")
		if _, err := os.Stat(userJS); os.IsNotExist(err) {
			src := "/usr/share/webapp-manager/firefox/profile"
			if st, srcErr := os.Stat(src); srcErr == nil && st.IsDir() {
				if err := copyTree(src, dst); err != nil {
					return err
				}
			}
			if _, err := os.Stat(userJS); os.IsNotExist(err) {
				body := "user_pref(\"toolkit.legacyUserProfileCustomizations.stylesheets\", true);\nuser_pref(\"browser.shell.checkDefaultBrowser\", false);\n"
				if err := os.WriteFile(userJS, []byte(body), 0o644); err != nil {
					return err
				}
			}
		}
		css := filepath.Join(chromeDir, "userChrome.css")
		if app.Navbar {
			return os.WriteFile(css, nil, 0o644)
		}
		info, err := os.Stat(css)
		if err != nil || info.Size() == 0 {
			return os.WriteFile(css, []byte(hiddenNavbarCSS), 0o644)
		}
		return nil
	case browser.KindEpiphany:
		ice := filepath.Join(home, ".local/share/ice/epiphany", "org.gnome.Epiphany.WebApp_"+app.Codename)
		link := filepath.Join(home, ".local/share", "org.gnome.Epiphany.WebApp_"+app.Codename)
		if err := os.MkdirAll(ice, 0o755); err != nil {
			return err
		}
		_ = os.WriteFile(filepath.Join(ice, ".app"), nil, 0o644)
		if _, err := os.Lstat(link); err != nil {
			return os.Symlink(ice, link)
		}
		return nil
	case browser.KindFalkon:
		ice := filepath.Join(home, ".local/share/ice/falkon", app.Codename)
		link := filepath.Join(home, ".config/falkon/profiles", app.Codename)
		if err := os.MkdirAll(ice, 0o755); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
			return err
		}
		if _, err := os.Lstat(link); err != nil {
			return os.Symlink(ice, link)
		}
		return nil
	default:
		if app.Isolated {
			return os.MkdirAll(filepath.Join(home, ".local/share/ice/profiles", app.Codename), 0o755)
		}
		return nil
	}
}

const hiddenNavbarCSS = `#nav-bar { visibility: collapse; }
#nav-bar:hover, #nav-bar:focus-within { visibility: visible; }
`

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}

func codename(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		b.WriteString("Web")
	}
	suffix := 0
	if n, err := os.Open("/dev/urandom"); err == nil {
		buf := make([]byte, 2)
		_, _ = io.ReadFull(n, buf)
		n.Close()
		suffix = (int(buf[0])<<8 | int(buf[1])) % 10000
	}
	return fmt.Sprintf("%s%04d", b.String(), suffix)
}

func categoryFrom(categories string) string {
	for _, part := range strings.Split(categories, ";") {
		switch part {
		case "GTK", "Network", "":
			continue
		default:
			return part
		}
	}
	return "Network"
}

func refreshDesktop(dir string) {
	if _, err := exec.LookPath("update-desktop-database"); err == nil {
		_ = exec.Command("update-desktop-database", dir).Run()
	}
}

func mustHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}
