package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"boki-webapps/internal/browser"
	"boki-webapps/internal/icon"
	"boki-webapps/internal/theme"
	"boki-webapps/internal/webapp"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

const (
	AppID   = "com.boki.WebApps"
	Version = "0.1.0"
)

type appUI struct {
	app      *gtk.Application
	win      *gtk.ApplicationWindow
	search   *gtk.SearchEntry
	list     *gtk.ListBox
	apps     []webapp.App
	visible  []webapp.App
	browsers []browser.Browser
	css      *gtk.CSSProvider
	monitors []gio.FileMonitorrer
}

func Run() int {
	application := gtk.NewApplication(AppID, gio.ApplicationFlagsNone)
	var ui *appUI
	application.ConnectActivate(func() {
		if ui != nil {
			ui.win.Present()
			return
		}
		ui = newAppUI(application)
		if os.Getenv("BOKI_SMOKE") == "1" {
			glib.TimeoutAdd(1200, func() {
				application.Quit()
			})
		}
	})
	return application.Run(os.Args)
}

func newAppUI(application *gtk.Application) *appUI {
	ui := &appUI{app: application}
	ui.win = gtk.NewApplicationWindow(application)
	ui.win.SetTitle("Boki-WebApps")
	ui.win.SetDefaultSize(920, 480)
	ui.win.SetIconName("web-browser")

	header := gtk.NewHeaderBar()
	header.SetShowTitleButtons(false)

	menuBtn := gtk.NewMenuButton()
	menuBtn.SetIconName("open-menu-symbolic")
	menuBtn.SetPrimary(true)
	menuBtn.AddCSSClass("flat")
	menuBtn.SetTooltipText("Main menu")
	menu := gio.NewMenu()
	menu.Append("Settings", "app.settings")
	menu.Append("Keyboard shortcuts", "app.shortcuts")
	menu.Append("About", "app.about")
	menu.Append("Quit", "app.quit")
	menuBtn.SetMenuModel(menu)
	header.PackStart(menuBtn)

	title := gtk.NewLabel("Boki-WebApps")
	title.AddCSSClass("app-title")
	title.SetHExpand(true)
	title.SetHAlign(gtk.AlignCenter)
	header.SetTitleWidget(title)

	closeBtn := gtk.NewButtonWithLabel("×")
	closeBtn.AddCSSClass("flat")
	closeBtn.AddCSSClass("window-x")
	closeBtn.SetTooltipText("Close")
	closeBtn.ConnectClicked(func() { ui.win.Close() })

	maxBtn := gtk.NewButtonFromIconName("window-maximize-symbolic")
	maxBtn.AddCSSClass("flat")
	maxBtn.SetTooltipText("Maximize")
	maxBtn.ConnectClicked(func() {
		if ui.win.IsMaximized() {
			ui.win.Unmaximize()
		} else {
			ui.win.Maximize()
		}
	})

	minBtn := gtk.NewButtonFromIconName("window-minimize-symbolic")
	minBtn.AddCSSClass("flat")
	minBtn.SetTooltipText("Minimize")
	minBtn.ConnectClicked(func() { ui.win.Minimize() })

	header.PackEnd(closeBtn)
	header.PackEnd(maxBtn)
	header.PackEnd(minBtn)
	ui.win.SetTitlebar(header)

	ui.addAction("settings", ui.showSettings)
	ui.addAction("shortcuts", ui.showShortcuts)
	ui.addAction("about", ui.showAbout)
	ui.addAction("quit", application.Quit)
	ui.addAction("add", func() { ui.edit(nil) })
	ui.addAction("search", func() { ui.search.GrabFocus() })
	application.SetAccelsForAction("app.add", []string{"<Control>n"})
	application.SetAccelsForAction("app.search", []string{"<Control>f"})
	application.SetAccelsForAction("app.settings", []string{"<Control>comma"})
	application.SetAccelsForAction("app.shortcuts", []string{"F1"})
	application.SetAccelsForAction("app.quit", []string{"<Control>q"})

	root := gtk.NewBox(gtk.OrientationVertical, 0)
	bar := gtk.NewBox(gtk.OrientationHorizontal, 16)
	bar.SetHAlign(gtk.AlignCenter)
	bar.SetMarginTop(18)
	bar.SetMarginBottom(14)

	addBtn := gtk.NewButtonFromIconName("list-add-symbolic")
	addBtn.AddCSSClass("circular")
	addBtn.SetTooltipText("Add web app")
	addBtn.ConnectClicked(func() { ui.edit(nil) })
	bar.Append(addBtn)

	ui.search = gtk.NewSearchEntry()
	ui.search.SetPlaceholderText("Search")
	ui.search.SetSizeRequest(460, -1)
	ui.search.ConnectSearchChanged(func() { ui.rebuild() })
	bar.Append(ui.search)
	root.Append(bar)

	ui.list = gtk.NewListBox()
	ui.list.SetSelectionMode(gtk.SelectionBrowse)
	ui.list.AddCSSClass("webapps")
	ui.list.SetMarginStart(28)
	ui.list.SetMarginEnd(28)
	ui.list.SetMarginBottom(12)
	ui.list.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		i := row.Index()
		if i < 0 || i >= len(ui.visible) {
			return
		}
		if err := webapp.Launch(ui.visible[i]); err != nil {
			textWindow(&ui.win.Window, "Could not open", err.Error())
		}
	})

	scroll := gtk.NewScrolledWindow()
	scroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scroll.SetChild(ui.list)
	scroll.SetVExpand(true)
	root.Append(scroll)
	ui.win.SetChild(root)

	_, _ = icon.Dir()
	ui.loadTheme()
	ui.watchTheme()
	ui.rescan()
	ui.reloadApps()
	ui.win.Present()
	return ui
}

func (ui *appUI) addAction(name string, fn func()) {
	action := gio.NewSimpleAction(name, nil)
	action.ConnectActivate(func(*glib.Variant) { fn() })
	ui.app.AddAction(action)
}

func (ui *appUI) rescan() {
	found, err := browser.Scan()
	if err != nil {
		textWindow(&ui.win.Window, "Scan failed", err.Error())
		return
	}
	ui.browsers = found
}

func (ui *appUI) reloadApps() {
	apps, err := webapp.List()
	if err != nil {
		textWindow(&ui.win.Window, "Could not read web apps", err.Error())
		return
	}
	sort.Slice(apps, func(i, j int) bool {
		left := strings.ToLower(apps[i].Name)
		right := strings.ToLower(apps[j].Name)
		if left == right {
			return apps[i].Codename < apps[j].Codename
		}
		return left < right
	})
	ui.apps = apps
	ui.rebuild()
}

func (ui *appUI) rebuild() {
	ui.list.RemoveAll()
	ui.visible = nil
	query := strings.ToLower(strings.TrimSpace(ui.search.Text()))
	for _, app := range ui.apps {
		if query != "" &&
			!strings.Contains(strings.ToLower(app.Name), query) &&
			!strings.Contains(strings.ToLower(app.URL), query) {
			continue
		}
		ui.visible = append(ui.visible, app)
		ui.list.Append(ui.row(app))
	}
	placeholder := "No web apps yet"
	if query != "" {
		placeholder = "No matches"
	}
	label := gtk.NewLabel(placeholder)
	label.SetMarginTop(24)
	ui.list.SetPlaceholder(label)
}

func (ui *appUI) row(app webapp.App) *gtk.ListBoxRow {
	row := gtk.NewListBoxRow()
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetMarginTop(6)
	box.SetMarginBottom(6)

	mark := gtk.NewImage()
	mark.SetPixelSize(36)
	mark.SetVAlign(gtk.AlignCenter)
	showIcon(mark, app.Icon)
	box.Append(mark)

	line := gtk.NewLabel(app.Name + " - " + app.URL)
	line.AddCSSClass("webapp-line")
	line.SetXAlign(0)
	line.SetHAlign(gtk.AlignStart)
	line.SetEllipsize(pango.EllipsizeEnd)
	line.SetMaxWidthChars(72)
	line.SetTooltipText("Double-click to open")
	box.Append(line)

	gap := gtk.NewLabel("")
	gap.SetHExpand(true)
	box.Append(gap)

	edit := gtk.NewButtonFromIconName("document-edit-symbolic")
	edit.AddCSSClass("flat")
	edit.AddCSSClass("webapp-edit")
	edit.SetTooltipText("Edit")
	current := app
	edit.ConnectClicked(func() { ui.edit(&current) })
	box.Append(edit)

	remove := gtk.NewButtonWithLabel("×")
	remove.AddCSSClass("flat")
	remove.AddCSSClass("webapp-x")
	remove.SetTooltipText("Remove")
	remove.ConnectClicked(func() {
		confirmWindow(&ui.win.Window, "Remove web app", "Remove "+current.Name+"?", "Remove", func() {
			if err := webapp.Delete(current); err != nil {
				textWindow(&ui.win.Window, "Could not remove", err.Error())
				return
			}
			ui.reloadApps()
		})
	})
	box.Append(remove)

	row.SetChild(box)
	return row
}

func (ui *appUI) loadTheme() {
	if ui.css == nil {
		ui.css = gtk.NewCSSProvider()
		if display := gdk.DisplayGetDefault(); display != nil {
			gtk.StyleContextAddProviderForDisplay(display, ui.css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
		}
	}
	path, err := theme.Path()
	if err != nil {
		ui.css.LoadFromString(theme.DefaultCSS)
		return
	}
	ui.css.LoadFromPath(path)
}

func (ui *appUI) watchTheme() {
	dir, err := theme.Dir()
	if err != nil {
		return
	}
	ui.watchPath(dir, true)
	ui.watchPath(filepath.Join(dir, theme.FileName), false)
}

func (ui *appUI) watchPath(path string, directory bool) {
	gfile := gio.NewFileForPath(path)
	var (
		mon gio.FileMonitorrer
		err error
	)
	if directory {
		mon, err = gfile.MonitorDirectory(context.Background(), gio.FileMonitorWatchMoves)
	} else {
		mon, err = gfile.MonitorFile(context.Background(), gio.FileMonitorNone)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "boki-webapps: theme watch %s: %v\n", path, err)
		return
	}
	ui.monitors = append(ui.monitors, mon)
	fm, ok := mon.(*gio.FileMonitor)
	if !ok {
		fmt.Fprintf(os.Stderr, "boki-webapps: theme watch %s: unexpected monitor %T\n", path, mon)
		return
	}
	fm.ConnectChanged(func(file, other gio.Filer, event gio.FileMonitorEvent) {
		if !themeEvent(file, other, event) {
			return
		}
		glib.IdleAdd(func() { ui.loadTheme() })
	})
}

func themeEvent(file, other gio.Filer, event gio.FileMonitorEvent) bool {
	if !isTheme(file) && !isTheme(other) {
		return false
	}
	switch event {
	case gio.FileMonitorEventChangesDoneHint,
		gio.FileMonitorEventCreated,
		gio.FileMonitorEventRenamed,
		gio.FileMonitorEventMovedIn:
		return true
	default:
		return false
	}
}

func isTheme(file gio.Filer) bool {
	if file == nil {
		return false
	}
	return filepath.Base(file.Path()) == theme.FileName || file.Basename() == theme.FileName
}
