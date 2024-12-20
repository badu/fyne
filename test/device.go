package test

import (
	"runtime"

	"fyne.io/fyne/v2"
)

type device struct {
}

func (d *device) Orientation() fyne.DeviceOrientation {
	return fyne.OrientationVertical
}

func (d *device) HasKeyboard() bool {
	return false
}

func (d *device) SystemScale() float32 {
	return d.SystemScaleForWindow(nil)
}

func (d *device) SystemScaleForWindow(fyne.Window) float32 {
	return 1
}

func (d *device) Locale() fyne.Locale {
	return "en"
}

func (*device) IsBrowser() bool {
	return runtime.GOARCH == "js" || runtime.GOOS == "js"
}
