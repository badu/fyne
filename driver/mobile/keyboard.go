package mobile

import (
	"fyne.io/fyne/v2"
)

// KeyboardType represents a type of virtual keyboard
type KeyboardType int32

const (
	// DefaultKeyboard is the keyboard with default input style and "return" return key
	DefaultKeyboard KeyboardType = iota
	// SingleLineKeyboard is the keyboard with default input style and "Done" return key
	SingleLineKeyboard
	// NumberKeyboard is the keyboard with number input style and "Done" return key
	NumberKeyboard
	// PasswordKeyboard is used to ensure that text is not leaked to 3rd party keyboard providers
	PasswordKeyboard
)

// Keyboardable describes any CanvasObject that needs a keyboard
type Keyboardable interface {
	// FocusGained is a hook called by the focus handling logic after this object gained the focus.
	FocusGained()
	// FocusLost is a hook called by the focus handling logic after this object lost the focus.
	FocusLost()

	// TypedRune is a hook called by the input handling logic on text input events if this object is focused.
	TypedRune(rune)
	// TypedKey is a hook called by the input handling logic on key events if this object is focused.
	TypedKey(*fyne.KeyEvent)
	Keyboard() KeyboardType
}
