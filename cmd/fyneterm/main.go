package main

import (
	"embed"
	"flag"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyneterm/data"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/terminal"
	"fyne.io/fyne/v2/theme"
)

const termOverlay = fyne.ThemeColorName("termOver")

var sizer *termTheme

//go:embed translation
var translations embed.FS

func setupListener(t *terminal.Terminal, w fyne.Window) {
	listen := make(chan terminal.Config)
	go func() {
		for {
			config := <-listen

			if config.Title == "" {
				w.SetTitle(termTitle())
			} else {
				w.SetTitle(termTitle() + ": " + config.Title)
			}
		}
	}()
	t.AddListener(listen)
}

func termTitle() string {
	return lang.L("Title")
}

func guessCellSize() fyne.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	return cell.MinSize()
}

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Show terminal debug messages")
	flag.Parse()

	lang.AddTranslationsFS(translations, "translation")

	app := app.New()
	app.SetIcon(data.Icon)
	sizer = newTermTheme()
	app.Settings().SetTheme(sizer)
	window := newTerminalWindow(app, sizer, debug)
	window.ShowAndRun()
}

func newTerminalWindow(app fyne.App, appTheme fyne.Theme, debug bool) fyne.Window {
	window := app.NewWindow(termTitle())
	window.SetPadded(false)

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	img := canvas.NewImageFromResource(data.FyneLogo)
	img.FillMode = canvas.ImageFillContain
	over := canvas.NewRectangle(appTheme.Color(termOverlay, app.Settings().ThemeVariant()))

	ch := make(chan fyne.Settings)
	go func() {
		for {
			<-ch

			bg.FillColor = theme.Color(theme.ColorNameBackground)
			bg.Refresh()
			over.FillColor = appTheme.Color(termOverlay, app.Settings().ThemeVariant())
			over.Refresh()
		}
	}()
	app.Settings().AddChangeListener(ch)

	terminalWidget := terminal.New()
	terminalWidget.SetDebug(debug)
	terminalWidget.SetClipboard(window.Clipboard())
	setupListener(terminalWidget, window)
	window.SetContent(container.NewStack(bg, img, over, terminalWidget))

	cellSize := guessCellSize()
	window.Resize(fyne.NewSize(cellSize.Width*80, cellSize.Height*24))
	window.Canvas().Focus(terminalWidget)

	terminalWidget.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift},
		func(_ fyne.Shortcut) {
			w := newTerminalWindow(app, appTheme, debug)
			w.Show()
		})
	terminalWidget.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyEqual, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift},
		func(_ fyne.Shortcut) {
			sizer.fontSize++
			app.Settings().SetTheme(sizer)
			terminalWidget.Refresh()
			terminalWidget.Resize(terminalWidget.Size())
		})
	terminalWidget.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyMinus, Modifier: fyne.KeyModifierControl},
		func(_ fyne.Shortcut) {
			sizer.fontSize--
			app.Settings().SetTheme(sizer)
			terminalWidget.Refresh()
			terminalWidget.Resize(terminalWidget.Size())
		})

	go func() {
		err := terminalWidget.RunLocalShell()
		if err != nil {
			fyne.LogError("Failure in terminal", err)
		}
		window.Close()
	}()

	return window
}
