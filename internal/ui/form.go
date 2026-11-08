package ui

import (
	"context"
	"strings"
	"time"

	"boki-webapps/internal/browser"
	"boki-webapps/internal/icon"
	"boki-webapps/internal/webapp"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func (ui *appUI) edit(existing *webapp.App) {
	browsers := append([]browser.Browser(nil), ui.browsers...)
	if existing != nil && existing.BrowserExec != "" && !hasExec(browsers, existing.BrowserExec) {
		browsers = append([]browser.Browser{{
			Name: existing.BrowserName,
			Exec: existing.BrowserExec,
			Icon: existing.Icon,
			Kind: existing.Kind,
		}}, browsers...)
	}

	dialog := gtk.NewWindow()
	dialog.SetTransientFor(&ui.win.Window)
	dialog.SetModal(true)
	dialog.SetTitle("Add")
	if existing != nil {
		dialog.SetTitle("Edit")
	}
	dialog.SetDefaultSize(540, 460)
	closeOnEscape(dialog)

	root := gtk.NewBox(gtk.OrientationVertical, 10)
	root.SetMarginStart(16)
	root.SetMarginEnd(16)
	root.SetMarginTop(16)
	root.SetMarginBottom(16)

	name := gtk.NewEntry()
	name.SetPlaceholderText("Youtube")
	url := gtk.NewEntry()
	url.SetPlaceholderText("https://youtube.com")
	iconEntry := gtk.NewEntry()
	iconEntry.SetPlaceholderText("From the site, or choose another")
	iconEntry.SetHExpand(true)
	preview := gtk.NewImage()
	preview.SetPixelSize(40)
	category := gtk.NewEntry()
	category.SetPlaceholderText("Network")
	params := gtk.NewEntry()
	params.SetPlaceholderText("Extra browser arguments")

	labels := make([]string, 0, len(browsers))
	for _, b := range browsers {
		labels = append(labels, b.Name)
	}
	if len(labels) == 0 {
		labels = []string{"No browsers found"}
	}
	choose := gtk.NewDropDownFromStrings(labels)
	choose.SetHExpand(true)

	isolated := gtk.NewCheckButtonWithLabel("Isolated profile")
	isolated.SetActive(true)
	private := gtk.NewCheckButtonWithLabel("Private window")
	navbar := gtk.NewCheckButtonWithLabel("Navigation bar (Firefox)")

	if existing != nil {
		name.SetText(existing.Name)
		url.SetText(existing.URL)
		iconEntry.SetText(existing.Icon)
		category.SetText(existing.Category)
		params.SetText(existing.Parameters)
		isolated.SetActive(existing.Isolated)
		private.SetActive(existing.Private)
		navbar.SetActive(existing.Navbar)
		for i, b := range browsers {
			if b.Exec == existing.BrowserExec {
				choose.SetSelected(uint(i))
				break
			}
		}
	}
	showIcon(preview, iconEntry.Text())

	syncNavbar := func() {
		b, ok := browserAt(browsers, choose)
		navbar.SetSensitive(ok && b.Kind == browser.KindFirefox)
	}
	choose.NotifyProperty("selected", syncNavbar)
	syncNavbar()

	lookupGen := 0
	var found []icon.Image
	iconStatus := gtk.NewLabel("")
	iconStatus.SetXAlign(0)

	chooseIcon := gtk.NewButtonWithLabel("Choose")
	chooseIcon.ConnectClicked(func() {
		lookupGen++
		openIconChoices(&ui.win.Window, found, func(path string) {
			iconEntry.SetText(path)
			showIcon(preview, path)
			iconStatus.SetText("")
		})
	})
	iconRow := gtk.NewBox(gtk.OrientationHorizontal, 8)
	iconRow.SetHExpand(true)
	iconRow.Append(preview)
	iconRow.Append(iconEntry)
	iconRow.Append(chooseIcon)

	scheduleLookup := func() {
		lookupGen++
		gen := lookupGen
		page := strings.TrimSpace(url.Text())
		glib.TimeoutAdd(600, func() bool {
			if gen != lookupGen {
				return false
			}
			if icon.Normalize(page) == "" {
				return false
			}
			iconStatus.SetText("Looking up the site icon…")
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
				defer cancel()
				images, err := icon.Lookup(ctx, page)
				glib.IdleAdd(func() {
					if gen != lookupGen {
						return
					}
					if err != nil || len(images) == 0 {
						iconStatus.SetText("No site icon found. Choose a picture.")
						return
					}
					found = images
					iconEntry.SetText(images[0].Path)
					showIcon(preview, images[0].Path)
					iconStatus.SetText("")
				})
			}()
			return false
		})
	}
	url.ConnectChanged(scheduleLookup)
	dialog.ConnectCloseRequest(func() bool {
		lookupGen++
		return false
	})

	for _, row := range [][2]any{
		{"Name", name},
		{"URL", url},
		{"Icon", iconRow},
		{"Category", category},
		{"Browser", choose},
		{"Parameters", params},
	} {
		root.Append(fieldRow(row[0].(string), row[1].(gtk.Widgetter)))
	}
	root.Append(isolated)
	root.Append(private)
	root.Append(navbar)

	errLabel := gtk.NewLabel("")
	errLabel.AddCSSClass("error")
	errLabel.SetXAlign(0)
	errLabel.SetWrap(true)
	root.Append(iconStatus)
	root.Append(errLabel)

	actions := gtk.NewBox(gtk.OrientationHorizontal, 8)
	actions.SetHAlign(gtk.AlignEnd)
	actions.SetMarginTop(8)
	cancel := gtk.NewButtonWithLabel("Cancel")
	cancel.ConnectClicked(func() { dialog.Close() })
	save := gtk.NewButtonWithLabel("Save")
	save.ConnectClicked(func() {
		b, ok := browserAt(browsers, choose)
		if !ok {
			errLabel.SetText("No browser found. Open Settings and scan.")
			return
		}
		app := webapp.App{
			Name:        strings.TrimSpace(name.Text()),
			URL:         strings.TrimSpace(url.Text()),
			Icon:        strings.TrimSpace(iconEntry.Text()),
			Category:    strings.TrimSpace(category.Text()),
			BrowserName: b.Name,
			BrowserExec: b.Exec,
			Parameters:  strings.TrimSpace(params.Text()),
			Navbar:      navbar.Active() && b.Kind == browser.KindFirefox,
			Private:     private.Active(),
			Isolated:    isolated.Active(),
			Kind:        b.Kind,
		}
		if existing != nil {
			app.Codename = existing.Codename
			app.Path = existing.Path
		}
		if strings.ContainsAny(app.Category, ";") {
			errLabel.SetText("Category cannot contain a semicolon.")
			return
		}
		saved, err := webapp.Save(app)
		if err != nil {
			errLabel.SetText(err.Error())
			return
		}
		if existing != nil {
			*existing = saved
		}
		dialog.Close()
		ui.reloadApps()
	})
	actions.Append(cancel)
	actions.Append(save)
	root.Append(actions)

	name.SetActivatesDefault(true)
	url.SetActivatesDefault(true)
	dialog.SetDefaultWidget(save)
	dialog.SetChild(root)
	dialog.Present()
}

func fieldRow(title string, widget gtk.Widgetter) *gtk.Box {
	row := gtk.NewBox(gtk.OrientationHorizontal, 12)
	label := gtk.NewLabel(title)
	label.SetXAlign(0)
	label.SetSizeRequest(110, -1)
	row.Append(label)
	if w, ok := widget.(*gtk.Entry); ok {
		w.SetHExpand(true)
	}
	row.Append(widget)
	return row
}

func browserAt(browsers []browser.Browser, choose *gtk.DropDown) (browser.Browser, bool) {
	if len(browsers) == 0 {
		return browser.Browser{}, false
	}
	i := choose.Selected()
	if i == gtk.INVALID_LIST_POSITION || int(i) >= len(browsers) {
		return browser.Browser{}, false
	}
	return browsers[i], true
}

func hasExec(browsers []browser.Browser, exec string) bool {
	for _, b := range browsers {
		if b.Exec == exec {
			return true
		}
	}
	return false
}
