Project discovery
---

[x] Don't break the contract removal (e.g. `var _ fyne.Widget = (*Accordion)(nil)`). We are NOT at kinder-garden.
[x] Define interfaces where they are used. Use Go language when writing Go language.
Notes:
* There are interfaces that have methods that return interfaces:
```go
type ToolbarItem interface {
      ToolbarObject() fyne.CanvasObject
}
// where fyne.CanvasObject is...
type CanvasObject interface {
	MinSize() Size
	Move(Position)
	Position() Position
	Resize(Size)
	Size() Size
	Hide()
	Visible() bool
	Show()
	Refresh()
}
```
* There are embedded interfaces as well:
```go
type Widget interface {
    CanvasObject
}
// where CanvasObject is described above
```

[ ] Exported function with the unexported return type `func NewGLDriver() *gLDriver`. Don't just use Go, abuse it!
[ ] Remove `// Deprecated: `. We need simpler picture.
[ ] Too much getters and setters. Make code easier to read.

```go
type Vector2 interface {
	Components() (float32, float32)
	IsZero() bool
}
// and implementation
func (v Delta) Components() (float32, float32) {
    return v.DX, v.DY
}
// if the properties are public, we complicating ourselves?
```

[ ] I am NOT sure that we got binding right.
[x] Goland doesn't know what this is `queueItem(key.(DataListener).DataChanged)`. We should help poor thing.
[ ] The hardest one : the composition everywhere (e.g. /fyne/internal/driver/glfw/window_desktop.go contains keyboard, mouse, scale, state, menu, resources, and more)