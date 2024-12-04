package dialog

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func Test_colorPreview_Color(t *testing.T) {
	test.NewTempApp(t)

	preview := newColorPreview(color.RGBA{R: 53, G: 113, B: 233, A: 255})
	preview.SetColor(color.RGBA{R: 90, G: 206, B: 80, A: 180})
	window := test.NewTempWindow(t, preview)
	padding := theme.Padding() * 2
	window.Resize(fyne.NewSize(128+padding, 64+padding))

	test.AssertRendersToImage(t, "color/preview_color.png", window.Canvas())
}
