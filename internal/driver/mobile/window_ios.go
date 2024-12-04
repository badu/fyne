//go:build ios

package mobile

import (
	fyneDriver "fyne.io/fyne/v2/driver"
)

func (w *window) RunNative(fn func(context any)) {
	fn(&fyneDriver.UnknownContext{})
}
