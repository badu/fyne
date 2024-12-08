// Package dialog contains helpers for building dialogs with the Fyne GUI toolkit.
package dialog

import (
	"fyne.io/fyne/v2"
)

// Dialog is the common API for any dialog window with a single dismiss button
type Dialog interface {
	Show()
	Hide()
	SetDismissText(label string)
	SetOnClosed(closed func())
	Refresh()
	Resize(size fyne.Size)

	// Since: 2.1
	MinSize() fyne.Size
}

// AddDialogKeyHandler adds a key handler to a dialog.
// It enables the user to close the dialog by pressing the escape key.
//
// Note that previously defined key events will be deactivated while the dialog is open.
func AddDialogKeyHandler(d Dialog, w fyne.Window) {
	originalEvent := w.Canvas().OnTypedKey() // == nil when not set
	w.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		if d == nil {
			return
		}
		if ke.Name == fyne.KeyEscape {
			d.Hide()
		}
	})
	d.SetOnClosed(func() {
		w.Canvas().SetOnTypedKey(originalEvent)
	})
}
