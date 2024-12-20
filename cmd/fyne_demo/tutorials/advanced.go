package tutorials

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

func scaleString(c fyne.Canvas) string {
	return strconv.FormatFloat(float64(c.Scale()), 'f', 2, 32)
}

func texScaleString(c fyne.Canvas) string {
	pixels, _ := c.PixelCoordinateForPosition(fyne.NewPos(1, 1))
	texScale := float32(pixels) / c.Scale()
	return strconv.FormatFloat(float64(texScale), 'f', 2, 32)
}

func prependTo(g *fyne.Container, s string) {
	g.Objects = append([]fyne.CanvasObject{widget.NewLabel(widget.LabelWithStaticText(s))}, g.Objects...)
	g.Refresh()
}

func setScaleText(scale, tex *widget.Label, win fyne.Window) {
	for scale.Visible() {
		scale.SetText(scaleString(win.Canvas()))
		tex.SetText(texScaleString(win.Canvas()))

		time.Sleep(time.Second)
	}
}

// advancedScreen loads a panel that shows details and settings that are a bit
// more detailed than normally needed.
func advancedScreen(win fyne.Window) fyne.CanvasObject {
	scale := widget.NewLabel(widget.LabelWithStaticText(""))
	tex := widget.NewLabel(widget.LabelWithStaticText(""))

	screen := widget.NewCard(
		widget.CardWithTitle("Screen info"),
		widget.CardWithContent(
			widget.NewForm(
				&widget.FormItem{Text: "Scale", Widget: scale},
				&widget.FormItem{Text: "Texture Scale", Widget: tex},
			),
		),
	)

	go setScaleText(scale, tex, win)

	label := widget.NewLabel(widget.LabelWithStaticText("Just type..."))
	generic := container.NewVBox()
	desk := container.NewVBox()

	genericCard := widget.NewCard(widget.CardWithSubtitle("Generic"), widget.CardWithContent(container.NewVScroll(generic)))
	deskCard := widget.NewCard(widget.CardWithSubtitle("Desktop"), widget.CardWithContent(container.NewVScroll(desk)))

	win.Canvas().SetOnTypedRune(func(r rune) {
		prependTo(generic, "Rune: "+string(r))
	})
	win.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		prependTo(generic, "Key : "+string(ev.Name))
	})
	if deskCanvas, ok := win.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(func(ev *fyne.KeyEvent) {
			prependTo(desk, "KeyDown: "+string(ev.Name))
		})
		deskCanvas.SetOnKeyUp(func(ev *fyne.KeyEvent) {
			prependTo(desk, "KeyUp  : "+string(ev.Name))
		})
	}

	buildStatus, ok := fyne.CurrentApp().Metadata().Custom["HelperText"]
	if !ok {
		buildStatus = "No helper text provided at compile time."
	}
	labelBuildStatus := widget.NewLabel(widget.LabelWithStaticText(buildStatus))

	return container.NewHBox(
		container.NewVBox(screen,
			widget.NewButton(
				widget.ButtonWithLabel("Custom Theme"),
				widget.ButtonWithCallback(
					func() {
						fyne.CurrentSettings().SetTheme(newCustomTheme())
					},
				),
			),
			widget.NewButton(
				widget.ButtonWithLabel("Fullscreen"),
				widget.ButtonWithCallback(
					func() {
						win.SetFullScreen(!win.FullScreen())
					},
				),
			),
		),
		container.NewBorder(label, labelBuildStatus, nil, nil,
			container.NewGridWithColumns(2, genericCard, deskCard),
		),
	)
}
