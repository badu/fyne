package widget

import "fyne.io/fyne/v2"

// Scrollable describes any [CanvasObject] that can also be scrolled.
// This is mostly used to implement the widget.ScrollContainer.
type Scrollable interface {
	Scrolled(*fyne.ScrollEvent)
}

// SecondaryTappable describes a [CanvasObject] that can be right-clicked or long-tapped.
type SecondaryTappable interface {
	TappedSecondary(*fyne.PointEvent)
}

// Tabbable describes any object that needs to accept the Tab key presses.
//
// Since: 2.1
type Tabbable interface {
	// AcceptsTab is a hook called by the key press handling logic.
	// If it returns true then the Tab key events will be sent using TypedKey.
	AcceptsTab() bool
}

// Tappable describes any [CanvasObject] that can also be tapped.
// This should be implemented by buttons etc that wish to handle pointer interactions.
type Tappable interface {
	Tapped(*fyne.PointEvent)
}

// Disableable describes any [CanvasObject] that can be disabled.
// This is primarily used with objects that also implement the Tappable interface.
type Disableable interface {
	Enable()
	Disable()
	Disabled() bool
}

// DoubleTappable describes any [CanvasObject] that can also be double tapped.
type DoubleTappable interface {
	DoubleTapped(*fyne.PointEvent)
}

// Shortcutable describes any [CanvasObject] that can respond to shortcut commands (quit, cut, copy, and paste).
type Shortcutable interface {
	TypedShortcut(fyne.Shortcut)
}
