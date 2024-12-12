package fyne

import "sync"

type IDirty interface {
	SetDirty()
}

// Container is a [CanvasObject] that contains a collection of child objects.
// The layout of the children is set by the specified Layout.
type Container struct {
	size     Size     // The current size of the Container
	position Position // The current position of the Container
	Hidden   bool     // Is this Container hidden

	Layout  Layout // The Layout algorithm for arranging child [CanvasObject]s
	lock    sync.Mutex
	Objects []CanvasObject // The set of [CanvasObject]s this container holds
}

// Add appends the specified object to the items this container manages.
//
// Since: 1.4
func (c *Container) Add(add CanvasObject) {
	if add == nil {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	c.Objects = append(c.Objects, add)
	c.layout()
}

// AddObject adds another [CanvasObject] to the set this Container holds.
//
// Deprecated: Use [Container.Add] instead.
func (c *Container) AddObject(o CanvasObject) {
	c.Add(o)
}

// Hide sets this container, and all its children, to be not visible.
func (c *Container) Hide() {
	if c.Hidden {
		return
	}

	c.Hidden = true
	repaint(c)
}

// MinSize calculates the minimum size of c.
// This is delegated to the [Container.Layout], if specified, otherwise it will be calculated.
func (c *Container) MinSize() Size {
	if c.Layout != nil {
		return c.Layout.MinSize(c.Objects)
	}

	minSize := NewSize(1, 1)
	for _, child := range c.Objects {
		minSize = minSize.Max(child.MinSize())
	}

	return minSize
}

// Move the container (and all its children) to a new position, relative to its parent.
func (c *Container) Move(pos Position) {
	c.position = pos
	repaint(c)
}

// Position gets the current position of c relative to its parent.
func (c *Container) Position() Position {
	return c.position
}

// Refresh causes this object to be redrawn in it's current state
func (c *Container) Refresh() {
	c.layout()

	for _, child := range c.Objects {
		child.Refresh()
	}

	// this is basically just canvas.Refresh(c) without the package loop
	o := CurrentDriver().CanvasForObject(c)
	if o == nil {
		return
	}
	o.Refresh(c)
}

// Remove updates the contents of this container to no longer include the specified object.
// This method is not intended to be used inside a loop, to remove all the elements.
// It is much more efficient to call [Container.RemoveAll) instead.
func (c *Container) Remove(rem CanvasObject) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if len(c.Objects) == 0 {
		return
	}

	for i, o := range c.Objects {
		if o != rem {
			continue
		}

		removed := make([]CanvasObject, len(c.Objects)-1)
		copy(removed, c.Objects[:i])
		copy(removed[i:], c.Objects[i+1:])

		c.Objects = removed
		c.layout()
		return
	}
}

// RemoveAll updates the contents of this container to no longer include any objects.
//
// Since: 2.2
func (c *Container) RemoveAll() {
	c.Objects = nil
	c.layout()
}

// Resize sets a new size for c.
func (c *Container) Resize(size Size) {
	if c.size == size {
		return
	}

	c.size = size
	c.layout()
}

// Show sets this container, and all its children, to be visible.
func (c *Container) Show() {
	if !c.Hidden {
		return
	}

	c.Hidden = false
}

// Size returns the current size c.
func (c *Container) Size() Size {
	return c.size
}

// Visible returns true if the container is currently visible, false otherwise.
func (c *Container) Visible() bool {
	return !c.Hidden
}

func (c *Container) layout() {
	if c.Layout == nil {
		return
	}

	c.Layout.Layout(c.Objects, c.size)
}

// repaint instructs the containing canvas to redraw, even if nothing changed.
// This method is a duplicate of what is in `canvas/canvas.go` to avoid a dependency loop or public API.
func repaint(obj *Container) {
	c := CurrentDriver().CanvasForObject(obj)
	if c != nil {
		if paint, ok := c.(IDirty); ok {
			paint.SetDirty()
		}
	}
}
