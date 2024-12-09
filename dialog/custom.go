package dialog

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// IDialog is the common API for any dialog window with a single dismiss button
type IDialog interface {
	Show()
	Hide()
	SetDismissText(label string)
	SetOnClosed(closed func())
	Refresh()
	Resize(size fyne.Size)

	// Since: 2.1
	MinSize() fyne.Size
}

// CloseOnEscape adds a key handler to a dialog.
// It enables the user to close the dialog by pressing the escape key.
//
// Note that previously defined key events will be deactivated while the dialog is open.
func CloseOnEscape(d IDialog, w fyne.Window) {
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

// CustomDialog implements a custom dialog.
//
// Since: 2.4
type CustomDialog struct {
	*Dialog
}

// NewCustom creates and returns a dialog over the specified application using custom
// content. The button will have the dismiss text set.
// The MinSize() of the CanvasObject passed will be used to set the size of the window.
func NewCustom(title, dismiss string, content fyne.CanvasObject, parent fyne.Window) *CustomDialog {
	d := &Dialog{content: content, title: title, parent: parent}

	d.dismiss = &widget.Button{Text: dismiss, OnTapped: d.Hide}
	d.create(container.NewGridWithColumns(1, d.dismiss))

	return &CustomDialog{Dialog: d}
}

// ShowCustom shows a dialog over the specified application using custom
// content. The button will have the dismiss text set.
// The MinSize() of the CanvasObject passed will be used to set the size of the window.
func ShowCustom(title, dismiss string, content fyne.CanvasObject, parent fyne.Window) {
	NewCustom(title, dismiss, content, parent).Show()
}

// NewCustomWithoutButtons creates a new custom dialog without any buttons.
// The MinSize() of the CanvasObject passed will be used to set the size of the window.
//
// Since: 2.4
func NewCustomWithoutButtons(title string, content fyne.CanvasObject, parent fyne.Window) *CustomDialog {
	d := &Dialog{content: content, title: title, parent: parent}
	d.create(container.NewGridWithColumns(1))

	return &CustomDialog{Dialog: d}
}

// SetButtons sets the row of buttons at the bottom of the dialog.
// Passing an empy slice will result in a dialog with no buttons.
//
// Since: 2.4
func (d *CustomDialog) SetButtons(buttons []fyne.CanvasObject) {
	d.dismiss = nil // New button row invalidates possible dismiss button.
	d.setButtons(container.NewGridWithRows(1, buttons...))
}

// ShowCustomWithoutButtons shows a dialog, wihout buttons, over the specified application
// using custom content.
// The MinSize() of the CanvasObject passed will be used to set the size of the window.
//
// Since: 2.4
func ShowCustomWithoutButtons(title string, content fyne.CanvasObject, parent fyne.Window) {
	NewCustomWithoutButtons(title, content, parent).Show()
}

// NewCustomConfirm creates and returns a dialog over the specified application using
// custom content. The cancel button will have the dismiss text set and the "OK" will
// use the confirm text. The response callback is called on user action.
// The MinSize() of the CanvasObject passed will be used to set the size of the window.
func NewCustomConfirm(title, confirm, dismiss string, content fyne.CanvasObject,
	callback func(bool), parent fyne.Window) *ConfirmDialog {
	d := &Dialog{content: content, title: title, parent: parent, callback: callback}

	d.dismiss = &widget.Button{Text: dismiss, Icon: theme.CancelIcon(),
		OnTapped: d.Hide,
	}
	ok := &widget.Button{Text: confirm, Icon: theme.ConfirmIcon(), Importance: widget.HighImportance,
		OnTapped: func() {
			d.hideWithResponse(true)
		},
	}
	d.create(container.NewGridWithColumns(2, d.dismiss, ok))

	return &ConfirmDialog{Dialog: d, confirm: ok}
}

// ShowCustomConfirm shows a dialog over the specified application using custom
// content. The cancel button will have the dismiss text set and the "OK" will use
// the confirm text. The response callback is called on user action.
// The MinSize() of the CanvasObject passed will be used to set the size of the window.
func ShowCustomConfirm(title, confirm, dismiss string, content fyne.CanvasObject,
	callback func(bool), parent fyne.Window) {
	NewCustomConfirm(title, confirm, dismiss, content, callback, parent).Show()
}
