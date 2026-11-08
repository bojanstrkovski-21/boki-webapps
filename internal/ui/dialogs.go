package ui

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

func (ui *appUI) showSettings() {
	dialog := gtk.NewWindow()
	dialog.SetTransientFor(&ui.win.Window)
	dialog.SetModal(true)
	dialog.SetTitle("Settings")
	dialog.SetDefaultSize(640, 420)
	closeOnEscape(dialog)

	root := gtk.NewBox(gtk.OrientationVertical, 8)
	root.SetMarginStart(16)
	root.SetMarginEnd(16)
	root.SetMarginTop(16)
	root.SetMarginBottom(16)

	heading := gtk.NewLabel("Browsers")
	heading.SetXAlign(0)
	heading.AddCSSClass("app-title")
	root.Append(heading)

	status := gtk.NewLabel("")
	status.SetXAlign(0)
	root.Append(status)

	list := gtk.NewListBox()
	list.SetSelectionMode(gtk.SelectionNone)
	scroll := gtk.NewScrolledWindow()
	scroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scroll.SetChild(list)
	scroll.SetVExpand(true)
	root.Append(scroll)

	fill := func() {
		list.RemoveAll()
		if len(ui.browsers) == 0 {
			empty := gtk.NewLabel("No browsers found")
			empty.SetMarginTop(12)
			list.SetPlaceholder(empty)
			status.SetText("0 browsers")
			return
		}
		status.SetText(fmt.Sprintf("%d browsers", len(ui.browsers)))
		for _, b := range ui.browsers {
			row := gtk.NewListBoxRow()
			label := gtk.NewLabel(b.Name + "  —  " + b.Exec)
			label.SetXAlign(0)
			label.SetEllipsize(pango.EllipsizeEnd)
			label.SetMarginTop(6)
			label.SetMarginBottom(6)
			label.SetMarginStart(8)
			row.SetChild(label)
			list.Append(row)
		}
	}
	fill()

	buttons := gtk.NewBox(gtk.OrientationHorizontal, 8)
	buttons.SetHAlign(gtk.AlignEnd)
	scan := gtk.NewButtonWithLabel("Scan")
	scan.ConnectClicked(func() {
		ui.rescan()
		fill()
	})
	closeBtn := gtk.NewButtonWithLabel("Close")
	closeBtn.ConnectClicked(func() { dialog.Close() })
	buttons.Append(scan)
	buttons.Append(closeBtn)
	root.Append(buttons)

	dialog.SetChild(root)
	dialog.Present()
}

func (ui *appUI) showShortcuts() {
	textWindow(&ui.win.Window, "Keyboard shortcuts", strings.TrimSpace(`
Ctrl+N    Add a web app
Ctrl+F    Search
Enter     Open the selected web app
Ctrl+,    Settings
F1        Keyboard shortcuts
Ctrl+Q    Quit
Escape    Close a dialog
`))
}

func (ui *appUI) showAbout() {
	textWindow(&ui.win.Window, "About", fmt.Sprintf("Boki-WebApps\nVersion %s\n\nWeb apps open in the browser you pick.\nTheme file: ~/.config/boki-webapps/theme.css", Version))
}

func textWindow(parent *gtk.Window, title, body string) {
	dialog := gtk.NewWindow()
	if parent != nil {
		dialog.SetTransientFor(parent)
		dialog.SetModal(true)
	}
	dialog.SetTitle(title)
	dialog.SetDefaultSize(420, 240)
	closeOnEscape(dialog)

	root := gtk.NewBox(gtk.OrientationVertical, 12)
	root.SetMarginStart(16)
	root.SetMarginEnd(16)
	root.SetMarginTop(16)
	root.SetMarginBottom(16)
	label := gtk.NewLabel(body)
	label.SetXAlign(0)
	label.SetYAlign(0)
	label.SetWrap(true)
	label.SetSelectable(true)
	label.SetVExpand(true)
	root.Append(label)

	closeBtn := gtk.NewButtonWithLabel("Close")
	closeBtn.SetHAlign(gtk.AlignEnd)
	closeBtn.ConnectClicked(func() { dialog.Close() })
	root.Append(closeBtn)
	dialog.SetChild(root)
	dialog.Present()
}

func confirmWindow(parent *gtk.Window, title, body, confirmLabel string, onYes func()) {
	dialog := gtk.NewWindow()
	dialog.SetTransientFor(parent)
	dialog.SetModal(true)
	dialog.SetTitle(title)
	dialog.SetDefaultSize(380, 160)
	closeOnEscape(dialog)

	root := gtk.NewBox(gtk.OrientationVertical, 12)
	root.SetMarginStart(16)
	root.SetMarginEnd(16)
	root.SetMarginTop(16)
	root.SetMarginBottom(16)
	label := gtk.NewLabel(body)
	label.SetXAlign(0)
	label.SetWrap(true)
	label.SetVExpand(true)
	root.Append(label)

	buttons := gtk.NewBox(gtk.OrientationHorizontal, 8)
	buttons.SetHAlign(gtk.AlignEnd)
	cancel := gtk.NewButtonWithLabel("Cancel")
	cancel.ConnectClicked(func() { dialog.Close() })
	yes := gtk.NewButtonWithLabel(confirmLabel)
	yes.ConnectClicked(func() {
		dialog.Close()
		onYes()
	})
	buttons.Append(cancel)
	buttons.Append(yes)
	root.Append(buttons)
	dialog.SetChild(root)
	dialog.Present()
}

func closeOnEscape(win *gtk.Window) {
	keys := gtk.NewEventControllerKey()
	keys.ConnectKeyPressed(func(keyval, _ uint, _ gdk.ModifierType) bool {
		if keyval == gdk.KEY_Escape {
			win.Close()
			return true
		}
		return false
	})
	win.AddController(keys)
}
