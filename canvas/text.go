package canvas

import (
	"fyne.io/fyne/v2/internal/async"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Text describes a text primitive in a Fyne canvas.
// A text object can have a style set which will apply to the whole string.
// No formatting or text parsing will be performed
type Text struct {
	size     async.Size     // The current size of the canvas object
	position async.Position // The current position of the object
	Hidden   bool           // Is this object currently hidden

	min async.Size // The minimum size this object can be

	propertyLock sync.RWMutex

	Alignment fyne.TextAlign // The alignment of the text content

	Color     color.Color    // The main text draw color
	Text      string         // The string content of this Text
	TextSize  float32        // Size of the text - if the Canvas scale is 1.0 this will be equivalent to point size
	TextStyle fyne.TextStyle // The style of the text content

	// FontSource defines a resource that can be used instead of the theme for looking up the font.
	// When a font source is set the `TextStyle` may not be effective, as it will be limited to the styles
	// present in the data provided.
	//
	// Since: 2.5
	FontSource fyne.Resource
}

// Position gets the current position of this canvas object, relative to its parent.
func (t *Text) Position() fyne.Position {
	return t.position.Load()
}

// Show will set this object to be visible.
func (t *Text) Show() {
	t.propertyLock.Lock()
	t.Hidden = false
	t.propertyLock.Unlock()
}

// Size returns the current size of this canvas object.
func (t *Text) Size() fyne.Size {
	return t.size.Load()
}

// Visible returns true if this object is visible, false otherwise.
func (t *Text) Visible() bool {
	t.propertyLock.RLock()
	defer t.propertyLock.RUnlock()

	return !t.Hidden
}

// Hide will set this text to not be visible
func (t *Text) Hide() {
	t.propertyLock.Lock()
	t.Hidden = true
	t.propertyLock.Unlock()

	repaint(t)
}

// MinSize returns the minimum size of this text object based on its font size and content.
// This is normally determined by the render implementation.
func (t *Text) MinSize() fyne.Size {
	// TODO : @Badu - this might be expensive too
	s, _ := fyne.CurrentApp().Driver().RenderedTextSize(t.Text, t.TextSize, t.TextStyle, t.FontSource)
	return s
}

// Move the text to a new position, relative to its parent / canvas
func (t *Text) Move(pos fyne.Position) {
	t.position.Store(pos)
	repaint(t)
}

// Resize on a text updates the new size of this object, which may not result in a visual change, depending on alignment.
func (t *Text) Resize(size fyne.Size) {
	if size == t.Size() {
		return
	}
	t.size.Store(size)
	Refresh(t)
}

// SetMinSize has no effect as the smallest size this canvas object can be is based on its font size and content.
func (t *Text) SetMinSize(fyne.Size) {
	// no-op
}

// Refresh causes this text to be redrawn with its configured state.
func (t *Text) Refresh() {
	Refresh(t)
}

// NewText returns a new Text implementation
func NewText(text string, color color.Color) *Text {
	return &Text{
		Color:    color,
		Text:     text,
		TextSize: theme.TextSize(),
	}
}
