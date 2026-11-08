package ui

import (
	"os"
	"strings"

	"boki-webapps/internal/icon"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func showIcon(preview *gtk.Image, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		preview.SetFromIconName("web-browser")
		return
	}
	if strings.Contains(value, "/") {
		if _, err := os.Stat(value); err == nil {
			preview.SetFromFile(value)
			return
		}
	}
	preview.SetFromIconName(value)
}

func openIconChoices(parent *gtk.Window, found []icon.Image, pick func(string)) {
	dialog := gtk.NewWindow()
	dialog.SetTransientFor(parent)
	dialog.SetModal(true)
	dialog.SetTitle("Choose an icon")
	dialog.SetDefaultSize(520, 180)
	closeOnEscape(dialog)

	root := gtk.NewBox(gtk.OrientationVertical, 8)
	root.SetMarginStart(12)
	root.SetMarginEnd(12)
	root.SetMarginTop(12)
	root.SetMarginBottom(12)

	customDir, _ := icon.Dir()
	custom, _ := icon.CustomImages()
	note := gtk.NewLabel("Your icons: " + customDir)
	note.SetXAlign(0)
	note.SetWrap(true)
	root.Append(note)
	if len(custom) == 0 {
		empty := gtk.NewLabel("Put PNG or SVG files in that folder, then press Choose again.")
		empty.SetXAlign(0)
		empty.SetWrap(true)
		root.Append(empty)
	}
	if len(custom) > 0 {
		root.Append(iconButtonRow(custom, pick, dialog))
	}

	if len(found) > 0 {
		site := gtk.NewLabel("From the site")
		site.SetXAlign(0)
		root.Append(site)
		root.Append(iconButtonRow(found, pick, dialog))
	}

	fileBtn := gtk.NewButtonWithLabel("Open icons folder")
	fileBtn.SetHAlign(gtk.AlignStart)
	fileBtn.ConnectClicked(func() {
		chooseIconFile(parent, func(path string) {
			pick(path)
			dialog.Close()
		})
	})
	root.Append(fileBtn)
	dialog.SetChild(root)
	dialog.Present()
}

func iconButtonRow(images []icon.Image, pick func(string), dialog *gtk.Window) *gtk.ScrolledWindow {
	row := gtk.NewBox(gtk.OrientationHorizontal, 8)
	for _, img := range images {
		button := gtk.NewButton()
		image := gtk.NewImageFromFile(img.Path)
		image.SetPixelSize(64)
		button.SetChild(image)
		button.SetTooltipText(img.Source)
		path := img.Path
		button.ConnectClicked(func() {
			pick(path)
			dialog.Close()
		})
		row.Append(button)
	}
	scroll := gtk.NewScrolledWindow()
	scroll.SetPolicy(gtk.PolicyAutomatic, gtk.PolicyNever)
	scroll.SetChild(row)
	scroll.SetMinContentHeight(80)
	return scroll
}

func chooseIconFile(parent *gtk.Window, pick func(string)) {
	chooser := gtk.NewFileChooserNative("Choose icon", parent, gtk.FileChooserActionOpen, "Choose", "Cancel")
	if dir, err := icon.Dir(); err == nil {
		_ = chooser.SetCurrentFolder(gio.NewFileForPath(dir))
	}
	filter := gtk.NewFileFilter()
	filter.SetName("Images")
	filter.AddMIMEType("image/png")
	filter.AddMIMEType("image/jpeg")
	filter.AddMIMEType("image/svg+xml")
	filter.AddMIMEType("image/webp")
	filter.AddMIMEType("image/gif")
	chooser.AddFilter(filter)
	chooser.SetFilter(filter)
	chooser.ConnectResponse(func(id int) {
		path := ""
		if id == int(gtk.ResponseAccept) {
			if file := chooser.File(); file != nil {
				path = file.Path()
			}
		}
		chooser.Destroy()
		if path != "" {
			pick(path)
		}
	})
	chooser.Show()
}
