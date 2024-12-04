package tutorials

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_demo/data"
	"fyne.io/fyne/v2/container"
)

func rgbGradient(x, y, w, h int) color.Color {
	g := int(float32(x) / float32(w) * float32(255))
	b := int(float32(y) / float32(h) * float32(255))

	return color.NRGBA{R: uint8(255 - b), G: uint8(g), B: uint8(b), A: 0xff}
}

// canvasScreen loads a graphics example panel for the demo app
func canvasScreen(_ fyne.Window) fyne.CanvasObject {
	gradient := canvas.NewHorizontalGradient(color.NRGBA{R: 0x80, A: 0xff}, color.NRGBA{G: 0x80, A: 0xff})
	go func() {
		for {
			time.Sleep(time.Second)

			gradient.Angle += 45
			if gradient.Angle >= 360 {
				gradient.Angle -= 360
			}
			canvas.Refresh(gradient)
		}
	}()

	return container.NewGridWrap(fyne.NewSize(90, 90),
		canvas.NewImageFromResource(data.FyneLogo),
		&canvas.Rectangle{FillColor: color.NRGBA{R: 0x80, A: 0xff},
			StrokeColor: color.NRGBA{R: 255, G: 120, B: 0, A: 255},
			StrokeWidth: 1},
		&canvas.Rectangle{
			FillColor:    color.NRGBA{R: 255, G: 200, B: 0, A: 180},
			StrokeColor:  color.NRGBA{R: 255, G: 120, B: 0, A: 255},
			StrokeWidth:  4.0,
			CornerRadius: 20},
		&canvas.Line{StrokeColor: color.NRGBA{B: 0x80, A: 0xff}, StrokeWidth: 5},
		&canvas.Circle{StrokeColor: color.NRGBA{B: 0x80, A: 0xff},
			FillColor:   color.NRGBA{R: 0x30, G: 0x30, B: 0x30, A: 0x60},
			StrokeWidth: 2},
		canvas.NewText("Text", color.NRGBA{G: 0x80, A: 0xff}),
		canvas.NewRasterWithPixels(rgbGradient),
		gradient,
		canvas.NewRadialGradient(color.NRGBA{R: 0x80, A: 0xff}, color.NRGBA{G: 0x80, B: 0x80, A: 0xff}),
	)
}
