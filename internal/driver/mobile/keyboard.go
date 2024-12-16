package mobile

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/internal/driver/mobile/app"
)

func hideVirtualKeyboard() {
	if d, ok := fyne.CurrentDriver().(*MobileDriver); ok {
		if d.app == nil { // not yet running
			return
		}

		d.app.HideVirtualKeyboard()
	}
}

func handleKeyboard(obj fyne.Focusable) {
	isDisabled := false
	if disWid, ok := obj.(Disableable); ok {
		isDisabled = disWid.Disabled()
	}
	if obj != nil && !isDisabled {
		if keyb, ok := obj.(mobile.Keyboardable); ok {
			showVirtualKeyboard(keyb.Keyboard())
		} else {
			showVirtualKeyboard(mobile.DefaultKeyboard)
		}
	} else {
		hideVirtualKeyboard()
	}
}

func showVirtualKeyboard(keyboard mobile.KeyboardType) {
	if d, ok := fyne.CurrentDriver().(*MobileDriver); ok {
		if d.app == nil { // not yet running
			fyne.LogError("Cannot show keyboard before app is running", nil)
			return
		}

		d.app.ShowVirtualKeyboard(app.KeyboardType(keyboard))
	}
}
