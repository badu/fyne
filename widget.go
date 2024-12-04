package fyne

// Widget defines the standard behaviours of any widget. This extends
// [CanvasObject]. A widget behaves in the same basic way but will encapsulate
// many child objects to create the rendered widget.
type Widget interface {
	// geometry

	// MinSize returns the minimum size this object needs to be drawn.
	MinSize() Size
	// Move moves this object to the given position relative to its parent.
	// This should only be called if your object is not in a container with a layout manager.
	Move(Position)
	// Position returns the current position of the object relative to its parent.
	Position() Position
	// Resize resizes this object to the given size.
	// This should only be called if your object is not in a container with a layout manager.
	Resize(Size)
	// Size returns the current size of this object.
	Size() Size

	// visibility

	// Hide hides this object.
	Hide()
	// Visible returns whether this object is visible or not.
	Visible() bool
	// Show shows this object.
	Show()

	// Refresh must be called if this object should be redrawn because its inner state changed.
	Refresh()

	// CreateRenderer returns a new [WidgetRenderer] for this widget.
	// This should not be called by regular code, it is used internally to render a widget.
	CreateRenderer() WidgetRenderer
}

// WidgetRenderer defines the behaviour of a widget's implementation.
// This is returned from a widget's declarative object through [Widget.CreateRenderer]
// and should be exactly one instance per widget in memory.
type WidgetRenderer interface {
	// Destroy is a hook that is called when the renderer is being destroyed.
	// This happens at some time after the widget is no longer visible, and
	// once destroyed, a renderer will not be reused.
	// Renderers should dispose and clean up any related resources, if necessary.
	Destroy()
	// Layout is a hook that is called if the widget needs to be laid out.
	// This should never call [Refresh].
	Layout(Size)
	// MinSize returns the minimum size of the widget that is rendered by this renderer.
	MinSize() Size
	// Objects returns all objects that should be drawn.
	Objects() []CanvasObject
	// Refresh is a hook that is called if the widget has updated and needs to be redrawn.
	// This might trigger a [Layout].
	Refresh()
}
