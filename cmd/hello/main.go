// Package main loads a very basic Hello World graphical application.
package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello")

	hello := widget.NewLabel(widget.LabelWithStaticText("Hello Fyne!"))
	w.SetContent(
		container.NewVBox(
			hello,
			widget.NewButton(
				widget.ButtonWithLabel("Hi!"),
				widget.ButtonWithCallback(
					func() {
						hello.SetText("Welcome 😀")
					},
				),
			),
		),
	)

	w.ShowAndRun()
}
