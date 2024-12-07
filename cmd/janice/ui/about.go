package ui

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	kxdialog "fyne.io/fyne/v2/cmd/janice/dialog"
)

func (u *UI) showAboutDialog() {
	data := u.app.Metadata()
	current := data.Version
	x, err := url.Parse(data.Custom["Website"])
	if err != nil || x.Path == "" {
		x, _ = url.Parse(websiteURL)
	}
	c := container.NewVBox(
		widget.NewRichTextFromMarkdown(
			fmt.Sprintf("## %s\n\n"+
				"**Version:** %s\n\n"+
				"(c) 2024 Erik Kalkoken", data.Name, current),
		),
		widget.NewLabel(widget.LabelWithStaticText("A desktop app for viewing large JSON files.")),
		widget.NewHyperlink(widget.HyperlinkWithLabel("Website"), widget.HyperlinkWithURL(x)),
	)
	d := dialog.NewCustom("About", "OK", c, u.window)
	kxdialog.AddDialogKeyHandler(d, u.window)
	d.Show()
}
