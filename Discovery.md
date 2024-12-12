The rules of refactoring and discovery
---

1. Check errors in caller, if callee doesn't return an error. This comes from "Always alert as close to the app as possible"
2. Always expand small payloads to their components
3. Always use small interface for arguments of a function
4. Always start with the model (data) and it's transformation
5. Always check for initialization of structs (with no "constructors"), to have field names: it will help you reordering fields. In general, you should have "constructors"
6. Exported functions should never return unexported types
7. If you need a getter / setter, you hide business logic from the users / developers. A public property + additional calls to inform the interested parties, should do
8. Consumer of a channel should create it and hand it over to producer
9. Code duplication is not as bad as confusing code. Overall, the code looks good, except the parts where it is avoiding duplication
10. Always go for immutability. If the pointer doesn't belong to the struct, it shows at least two cases : a hidden change in a foreign function or a shared value. Find a better way to share values
11. If inside a function, we check for implementation satisfies an interface, do NOT alter the methods of that interface (e.g. `type selectableText interface {
    Enable()
    Disable()
    Disabled() bool
    SelectedText() string
    }`) - removing one of the functions will break the code
12. If you have multiple constructors, it's a sign you should use functional options
13. Backward compatibility, even for settings, has real costs
14. Don't export properties / methods / functions for the sake of tests

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