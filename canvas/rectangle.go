package canvas

import (
	"fyne.io/fyne/v2/internal/async"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
)

// Rectangle describes a colored rectangle primitive in a Fyne canvas
type Rectangle struct {
	size     async.Size     // The current size of the canvas object
	position async.Position // The current position of the object
	Hidden   bool           // Is this object currently hidden

	min async.Size // The minimum size this object can be

	propertyLock sync.RWMutex

	FillColor   color.Color // The rectangle fill color
	StrokeColor color.Color // The rectangle stroke color
	StrokeWidth float32     // The stroke width of the rectangle
	// The radius of the rectangle corners
	//
	// Since: 2.4
	CornerRadius float32
}

// Hide will set this rectangle to not be visible
func (r *Rectangle) Hide() {
	r.propertyLock.Lock()
	r.Hidden = true
	r.propertyLock.Unlock()

	repaint(r)
}

// MinSize returns the specified minimum size, if set, or {1, 1} otherwise.
func (r *Rectangle) MinSize() fyne.Size {
	min := r.min.Load()
	if min.IsZero() {
		return fyne.Size{Width: 1, Height: 1}
	}

	return min
}

// Move the rectangle to a new position, relative to its parent / canvas
func (r *Rectangle) Move(pos fyne.Position) {
	r.position.Store(pos)

	repaint(r)
}

// Position gets the current position of this canvas object, relative to its parent.
func (r *Rectangle) Position() fyne.Position {
	return r.position.Load()
}

// Refresh causes this rectangle to be redrawn with its configured state.
func (r *Rectangle) Refresh() {
	Refresh(r)
}

// Resize on a rectangle updates the new size of this object.
// If it has a stroke width this will cause it to Refresh.
func (r *Rectangle) Resize(size fyne.Size) {
	if size == r.Size() {
		return
	}

	r.size.Store(size)
	if r.StrokeWidth == 0 {
		return
	}

	Refresh(r)
}

// SetMinSize specifies the smallest size this object should be.
func (r *Rectangle) SetMinSize(size fyne.Size) {
	r.min.Store(size)
}

// Show will set this object to be visible.
func (r *Rectangle) Show() {
	r.propertyLock.Lock()
	r.Hidden = false
	r.propertyLock.Unlock()
}

// Size returns the current size of this canvas object.
func (r *Rectangle) Size() fyne.Size {
	return r.size.Load()
}

// Visible returns true if this object is visible, false otherwise.
func (r *Rectangle) Visible() bool {
	r.propertyLock.RLock()
	defer r.propertyLock.RUnlock()

	return !r.Hidden
}

// NewRectangle returns a new Rectangle instance
func NewRectangle(color color.Color) *Rectangle {
	return &Rectangle{
		FillColor: color,
	}
}
