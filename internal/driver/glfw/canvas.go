package glfw

import (
	"image"
	"image/color"
	"math"
	"reflect"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/internal"
	"fyne.io/fyne/v2/internal/app"
	"fyne.io/fyne/v2/internal/async"
	"fyne.io/fyne/v2/internal/build"
	"fyne.io/fyne/v2/internal/cache"
	"fyne.io/fyne/v2/internal/driver"
	"fyne.io/fyne/v2/internal/painter/gl"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SizeableCanvas defines a canvas with size related functions.
type SizeableCanvas interface {
	Content() fyne.CanvasObject
	SetContent(fyne.CanvasObject)

	Refresh(fyne.CanvasObject)

	// Focus makes the provided item focused.
	// The item has to be added to the contents of the canvas before calling this.
	Focus(fyne.Focusable)
	// FocusNext focuses the next focusable item.
	// If no item is currently focused, the first focusable item is focused.
	// If the last focusable item is currently focused, the first focusable item is focused.
	//
	// Since: 2.0
	FocusNext()
	// FocusPrevious focuses the previous focusable item.
	// If no item is currently focused, the last focusable item is focused.
	// If the first focusable item is currently focused, the last focusable item is focused.
	//
	// Since: 2.0
	FocusPrevious()
	Unfocus()
	Focused() fyne.Focusable

	// Size returns the current size of this canvas
	Size() fyne.Size
	// Scale returns the current scale (multiplication factor) this canvas uses to render
	// The pixel size of a [CanvasObject] can be found by multiplying by this value.
	Scale() float32

	// Overlays returns the overlay stack.
	Overlays() fyne.OverlayStack

	OnTypedRune() func(rune)
	SetOnTypedRune(func(rune))
	OnTypedKey() func(*fyne.KeyEvent)
	SetOnTypedKey(func(*fyne.KeyEvent))
	AddShortcut(shortcut fyne.Shortcut, handler func(shortcut fyne.Shortcut))
	RemoveShortcut(shortcut fyne.Shortcut)

	Capture() image.Image

	// PixelCoordinateForPosition returns the x and y pixel coordinate for a given position on this canvas.
	// This can be used to find absolute pixel positions or pixel offsets relative to an object top left.
	PixelCoordinateForPosition(fyne.Position) (int, int)

	// InteractiveArea returns the position and size of the central interactive area.
	// Operating system elements may overlap the portions outside this area and widgets should avoid being outside.
	//
	// Since: 1.4
	InteractiveArea() (fyne.Position, fyne.Size)

	Resize(fyne.Size)
	MinSize() fyne.Size
}

type activatableMenu interface {
	IsActive() bool
}

type overlayStack struct {
	internal.OverlayStack

	propertyLock sync.RWMutex
	renderCaches []*renderCacheTree
}

func (o *overlayStack) Add(overlay fyne.CanvasObject) {
	if overlay == nil {
		return
	}
	o.propertyLock.Lock()
	defer o.propertyLock.Unlock()
	o.add(overlay)
}

func (o *overlayStack) Remove(overlay fyne.CanvasObject) {
	if overlay == nil || len(o.List()) == 0 {
		return
	}
	o.propertyLock.Lock()
	defer o.propertyLock.Unlock()
	o.remove(overlay)
}

func (o *overlayStack) add(overlay fyne.CanvasObject) {
	o.renderCaches = append(o.renderCaches, &renderCacheTree{root: &driver.RenderCacheNode{Obj: overlay}})
	o.OverlayStack.Add(overlay)
}

func (o *overlayStack) remove(overlay fyne.CanvasObject) {
	o.OverlayStack.Remove(overlay)
	overlayCount := len(o.List())
	o.renderCaches[overlayCount] = nil // release memory reference to removed element
	o.renderCaches = o.renderCaches[:overlayCount]
}

type renderCacheTree struct {
	sync.RWMutex
	root *driver.RenderCacheNode
}

// WithContext allows drivers to execute within another context.
// Mostly this helps GLFW code execute within the painter's GL context.
type WithContext interface {
	RunWithContext(f func())
	RescaleContext()
	Context() any
}

type glCanvas struct {
	Mu sync.RWMutex

	OnFocus   func(obj fyne.Focusable)
	OnUnfocus func()

	contentFocusMgr *app.FocusManager
	menuFocusMgr    *app.FocusManager
	overlays        *overlayStack

	shortcut fyne.ShortcutHandler

	painter gl.Painter

	// Any object that requestes to enter to the refresh queue should
	// not be omitted as it is always a rendering task's decision
	// for skipping frames or drawing calls.
	//
	// If an object failed to ender the refresh queue, the object may
	// disappear or blink from the view at any frames. As of this reason,
	// the refreshQueue is an unbounded queue which is bale to cache
	// arbitrary number of fyne.CanvasObject for the rendering.
	refreshQueue *async.CanvasObjectQueue
	dirty        atomic.Bool

	mWindowHeadTree *renderCacheTree
	contentTree     *renderCacheTree
	menuTree        *renderCacheTree

	content fyne.CanvasObject
	menu    fyne.CanvasObject
	padded  bool
	size    fyne.Size

	onTypedRune func(rune)
	onTypedKey  func(*fyne.KeyEvent)
	onKeyDown   func(*fyne.KeyEvent)
	onKeyUp     func(*fyne.KeyEvent)
	// shortcut    fyne.ShortcutHandler

	scale, detectedScale, texScale float32

	context         WithContext
	webExtraWindows *container.MultipleWindows
}

// AddShortcut adds a shortcut to the canvas.
func (c *glCanvas) AddShortcut(shortcut fyne.Shortcut, handler func(shortcut fyne.Shortcut)) {
	c.shortcut.AddShortcut(shortcut, handler)
}

func (c *glCanvas) DrawDebugOverlay(obj fyne.CanvasObject, pos fyne.Position, size fyne.Size) {
	switch obj.(type) {
	case fyne.Widget:
		r := canvas.NewRectangle(color.Transparent)
		r.StrokeColor = color.NRGBA{R: 0xcc, G: 0x33, B: 0x33, A: 0xff}
		r.StrokeWidth = 1
		r.Resize(obj.Size())
		c.Painter().Paint(r, pos, size)

		t := canvas.NewText(reflect.ValueOf(obj).Elem().Type().Name(), r.StrokeColor)
		t.TextSize = 10
		c.Painter().Paint(t, pos.AddXY(2, 2), size)
	case *fyne.Container:
		r := canvas.NewRectangle(color.Transparent)
		r.StrokeColor = color.NRGBA{R: 0x33, G: 0x33, B: 0xcc, A: 0xff}
		r.StrokeWidth = 1
		r.Resize(obj.Size())
		c.Painter().Paint(r, pos, size)
	}
}

// EnsureMinSize ensure canvas min size.
//
// This function uses lock.
func (c *glCanvas) EnsureMinSize() bool {
	if c.Content() == nil {
		return false
	}
	windowNeedsMinSizeUpdate := false
	csize := c.Size()
	min := c.MinSize()

	c.Mu.RLock()
	defer c.Mu.RUnlock()

	var parentNeedingUpdate *driver.RenderCacheNode

	ensureMinSize := func(node *driver.RenderCacheNode, pos fyne.Position) {
		obj := node.Obj
		cache.SetCanvasForObject(obj, c, func() {
			if img, ok := obj.(*canvas.Image); ok {
				c.Mu.RUnlock()
				img.Refresh() // this may now have a different texScale
				c.Mu.RLock()
			}
		})

		if parentNeedingUpdate == node {
			c.updateLayout(obj)
			parentNeedingUpdate = nil
		}

		c.Mu.RUnlock()
		if !obj.Visible() {
			c.Mu.RLock()
			return
		}
		minSize := obj.MinSize()
		c.Mu.RLock()

		minSizeChanged := node.MinSize != minSize
		if minSizeChanged {
			node.MinSize = minSize
			if node.Parent != nil {
				parentNeedingUpdate = node.Parent
			} else {
				windowNeedsMinSizeUpdate = true
				c.Mu.RUnlock()
				size := obj.Size()
				c.Mu.RLock()
				expectedSize := minSize.Max(size)
				if expectedSize != size && size != csize {
					c.Mu.RUnlock()
					obj.Resize(expectedSize)
					c.Mu.RLock()
				} else {
					c.updateLayout(obj)
				}
			}
		}
	}
	c.WalkTrees(nil, ensureMinSize)

	shouldResize := windowNeedsMinSizeUpdate && (csize.Width < min.Width || csize.Height < min.Height)
	if shouldResize {
		c.Mu.RUnlock()
		c.Resize(csize.Max(min))
		c.Mu.RLock()
	}
	return windowNeedsMinSizeUpdate
}

// Focus makes the provided item focused.
func (c *glCanvas) Focus(obj fyne.Focusable) {
	focusMgr := c.focusManager()
	if focusMgr != nil && focusMgr.Focus(obj) { // fast path – probably >99.9% of all cases
		if c.OnFocus != nil {
			c.OnFocus(obj)
		}
		return
	}

	c.Mu.RLock()
	focusMgrs := append([]*app.FocusManager{c.contentFocusMgr, c.menuFocusMgr}, c.overlays.ListFocusManagers()...)
	c.Mu.RUnlock()

	for _, mgr := range focusMgrs {
		if mgr == nil {
			continue
		}
		if focusMgr != mgr {
			if mgr.Focus(obj) {
				if c.OnFocus != nil {
					c.OnFocus(obj)
				}
				return
			}
		}
	}

	fyne.LogError("Failed to focus object which is not part of the canvas’ content, menu or overlays.", nil)
}

// Focused returns the current focused object.
func (c *glCanvas) Focused() fyne.Focusable {
	mgr := c.focusManager()
	if mgr == nil {
		return nil
	}
	return mgr.Focused()
}

// FocusGained signals to the manager that its content got focus.
// Valid only on Desktop.
func (c *glCanvas) FocusGained() {
	mgr := c.focusManager()
	if mgr == nil {
		return
	}
	mgr.FocusGained()
}

// FocusLost signals to the manager that its content lost focus.
// Valid only on Desktop.
func (c *glCanvas) FocusLost() {
	mgr := c.focusManager()
	if mgr == nil {
		return
	}
	mgr.FocusLost()
}

// FocusNext focuses the next focusable item.
func (c *glCanvas) FocusNext() {
	mgr := c.focusManager()
	if mgr == nil {
		return
	}
	mgr.FocusNext()
}

// FocusPrevious focuses the previous focusable item.
func (c *glCanvas) FocusPrevious() {
	mgr := c.focusManager()
	if mgr == nil {
		return
	}
	mgr.FocusPrevious()
}

// FreeDirtyTextures frees dirty textures and returns the number of freed textures.
func (c *glCanvas) FreeDirtyTextures() (freed uint64) {
	freeObject := func(object fyne.CanvasObject) {
		freeWalked := func(obj fyne.CanvasObject, _ fyne.Position, _ fyne.Position, _ fyne.Size) bool {
			// No image refresh while recursing to avoid double texture upload.
			if _, ok := obj.(*canvas.Image); ok {
				return false
			}
			if c.painter != nil {
				c.painter.Free(obj)
			}
			return false
		}

		// Image.Refresh will trigger a refresh specific to the object, while recursing on parent widget would just lead to
		// a double texture upload.
		if img, ok := object.(*canvas.Image); ok {
			if c.painter != nil {
				c.painter.Free(img)
			}
		} else {
			driver.WalkCompleteObjectTree(object, freeWalked, nil)
		}
	}

	// Within a frame, refresh tasks are requested from the Refresh method,
	// and we desire to clear out all requested operations within a frame.
	// See https://github.com/fyne-io/fyne/issues/2548.
	tasksToDo := c.refreshQueue.Len()

	shouldFilterDuplicates := tasksToDo > 200 // filtering has overhead, not worth enabling for few tasks
	var refreshSet map[fyne.CanvasObject]struct{}
	if shouldFilterDuplicates {
		refreshSet = make(map[fyne.CanvasObject]struct{})
	}

	for c.refreshQueue.Len() > 0 {
		object := c.refreshQueue.Out()
		if !shouldFilterDuplicates {
			freed++
			freeObject(object)
		} else {
			refreshSet[object] = struct{}{}
			tasksToDo--
			if tasksToDo == 0 {
				shouldFilterDuplicates = false // stop collecting messages to avoid starvation
				for object := range refreshSet {
					freed++
					freeObject(object)
				}
			}
		}
	}

	if c.painter != nil {
		cache.RangeExpiredTexturesFor(c, c.painter.Free)
	}
	return
}

// Initialize initializes the canvas.
func (c *glCanvas) Initialize() {
	c.refreshQueue = async.NewCanvasObjectQueue()
	c.overlays = &overlayStack{
		OverlayStack: internal.OverlayStack{
			OnChange: c.overlayChanged,
			Canvas:   c,
		},
	}
}

// ObjectTrees return canvas object trees.
//
// This function uses lock.
func (c *glCanvas) ObjectTrees() []fyne.CanvasObject {
	c.Mu.RLock()
	var content, menu fyne.CanvasObject
	if c.contentTree != nil && c.contentTree.root != nil {
		content = c.contentTree.root.Obj
	}
	if c.menuTree != nil && c.menuTree.root != nil {
		menu = c.menuTree.root.Obj
	}
	c.Mu.RUnlock()
	trees := make([]fyne.CanvasObject, 0, len(c.Overlays().List())+2)
	trees = append(trees, content)
	if menu != nil {
		trees = append(trees, menu)
	}
	trees = append(trees, c.Overlays().List()...)
	return trees
}

// Overlays returns the overlay stack.
func (c *glCanvas) Overlays() fyne.OverlayStack {
	// we don't need to lock here, because overlays never changes
	return c.overlays
}

// Painter returns the canvas painter.
func (c *glCanvas) Painter() gl.Painter {
	return c.painter
}

// Refresh refreshes a canvas object.
func (c *glCanvas) Refresh(obj fyne.CanvasObject) {
	walkNeeded := false
	switch obj.(type) {
	case *fyne.Container:
		walkNeeded = true
	case fyne.Widget:
		walkNeeded = true
	}

	if walkNeeded {
		driver.WalkCompleteObjectTree(obj, func(co fyne.CanvasObject, p1, p2 fyne.Position, s fyne.Size) bool {
			if i, ok := co.(*canvas.Image); ok {
				i.Refresh()
			}
			return false
		}, nil)
	}

	c.refreshQueue.In(obj)
	c.SetDirty()
}

// RemoveShortcut removes a shortcut from the canvas.
func (c *glCanvas) RemoveShortcut(shortcut fyne.Shortcut) {
	c.shortcut.RemoveShortcut(shortcut)
}

// SetContentTreeAndFocusMgr sets content tree and focus manager.
//
// This function does not use the canvas lock.
func (c *glCanvas) SetContentTreeAndFocusMgr(content fyne.CanvasObject) {
	c.contentTree = &renderCacheTree{root: &driver.RenderCacheNode{Obj: content}}
	var focused fyne.Focusable
	if c.contentFocusMgr != nil {
		focused = c.contentFocusMgr.Focused() // keep old focus if possible
	}
	c.contentFocusMgr = app.NewFocusManager(content)
	if focused != nil {
		c.contentFocusMgr.Focus(focused)
	}
}

// CheckDirtyAndClear returns true if the canvas is dirty and
// clears the dirty state atomically.
func (c *glCanvas) CheckDirtyAndClear() bool {
	return c.dirty.Swap(false)
}

// SetDirty sets canvas dirty flag atomically.
func (c *glCanvas) SetDirty() {
	c.dirty.Store(true)
}

// SetMenuTreeAndFocusMgr sets menu tree and focus manager.
//
// This function does not use the canvas lock.
func (c *glCanvas) SetMenuTreeAndFocusMgr(menu fyne.CanvasObject) {
	c.menuTree = &renderCacheTree{root: &driver.RenderCacheNode{Obj: menu}}
	if menu != nil {
		c.menuFocusMgr = app.NewFocusManager(menu)
	} else {
		c.menuFocusMgr = nil
	}
}

// SetMobileWindowHeadTree sets window head tree.
//
// This function does not use the canvas lock.
func (c *glCanvas) SetMobileWindowHeadTree(head fyne.CanvasObject) {
	c.mWindowHeadTree = &renderCacheTree{root: &driver.RenderCacheNode{Obj: head}}
}

// SetPainter sets the canvas painter.
func (c *glCanvas) SetPainter(p gl.Painter) {
	c.painter = p
}

// TypedShortcut handle the registered shortcut.
func (c *glCanvas) TypedShortcut(shortcut fyne.Shortcut) {
	c.shortcut.TypedShortcut(shortcut)
}

// Unfocus unfocuses all the objects in the canvas.
func (c *glCanvas) Unfocus() {
	mgr := c.focusManager()
	if mgr == nil {
		return
	}
	if mgr.Focus(nil) && c.OnUnfocus != nil {
		c.OnUnfocus()
	}
}

// WalkTrees walks over the trees.
func (c *glCanvas) WalkTrees(
	beforeChildren func(*driver.RenderCacheNode, fyne.Position),
	afterChildren func(*driver.RenderCacheNode, fyne.Position),
) {
	c.walkTree(c.contentTree, beforeChildren, afterChildren)
	if c.mWindowHeadTree != nil && c.mWindowHeadTree.root.Obj != nil {
		c.walkTree(c.mWindowHeadTree, beforeChildren, afterChildren)
	}
	if c.menuTree != nil && c.menuTree.root.Obj != nil {
		c.walkTree(c.menuTree, beforeChildren, afterChildren)
	}
	for _, tree := range c.overlays.renderCaches {
		if tree != nil {
			c.walkTree(tree, beforeChildren, afterChildren)
		}
	}
}

func (c *glCanvas) focusManager() *app.FocusManager {
	if focusMgr := c.overlays.TopFocusManager(); focusMgr != nil {
		return focusMgr
	}
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	if c.isMenuActive() {
		return c.menuFocusMgr
	}
	return c.contentFocusMgr
}

func (c *glCanvas) isMenuActive() bool {
	if c.menuTree == nil || c.menuTree.root == nil || c.menuTree.root.Obj == nil {
		return false
	}
	menu := c.menuTree.root.Obj
	if am, ok := menu.(activatableMenu); ok {
		return am.IsActive()
	}
	return true
}

func (c *glCanvas) walkTree(
	tree *renderCacheTree,
	beforeChildren func(*driver.RenderCacheNode, fyne.Position),
	afterChildren func(*driver.RenderCacheNode, fyne.Position),
) {
	tree.Lock()
	defer tree.Unlock()
	var node, parent, prev *driver.RenderCacheNode
	node = tree.root

	bc := func(obj fyne.CanvasObject, pos fyne.Position, _ fyne.Position, _ fyne.Size) bool {
		if node != nil && node.Obj != obj {
			if parent.FirstChild == node {
				parent.FirstChild = nil
			}
			node = nil
		}
		if node == nil {
			node = &driver.RenderCacheNode{Parent: parent, Obj: obj}
			if parent.FirstChild == nil {
				parent.FirstChild = node
			} else {
				prev.NextSibling = node
			}
		}
		if prev != nil && prev.Parent != parent {
			prev = nil
		}

		if beforeChildren != nil {
			beforeChildren(node, pos)
		}

		parent = node
		node = parent.FirstChild
		return false
	}
	ac := func(obj fyne.CanvasObject, pos fyne.Position, _ fyne.CanvasObject) {
		node = parent
		parent = node.Parent
		if prev != nil && prev.Parent != parent {
			prev.NextSibling = nil
		}

		if afterChildren != nil {
			afterChildren(node, pos)
		}

		prev = node
		node = node.NextSibling
	}
	driver.WalkVisibleObjectTree(tree.root.Obj, bc, ac)
}

func (c *glCanvas) updateLayout(objToLayout fyne.CanvasObject) {
	switch cont := objToLayout.(type) {
	case *fyne.Container:
		if cont.Layout != nil {
			layout := cont.Layout
			objects := cont.Objects
			c.Mu.RUnlock()
			layout.Layout(objects, cont.Size())
			c.Mu.RLock()
		}
	case fyne.Widget:
		renderer := cache.Renderer(cont)
		c.Mu.RUnlock()
		renderer.Layout(cont.Size())
		c.Mu.RLock()
	}
}

func (c *glCanvas) Capture() image.Image {
	var img image.Image
	runOnDraw(c.context.(*window), func() {
		img = c.Painter().Capture(c)
	})
	return img
}

func (c *glCanvas) Content() fyne.CanvasObject {
	c.Mu.RLock()
	retval := c.content
	c.Mu.RUnlock()
	return retval
}

func (c *glCanvas) DismissMenu() bool {
	c.Mu.RLock()
	menu := c.menu
	c.Mu.RUnlock()
	if menu != nil && menu.(*MenuBar).IsActive() {
		menu.(*MenuBar).Toggle()
		return true
	}
	return false
}

func (c *glCanvas) InteractiveArea() (fyne.Position, fyne.Size) {
	return fyne.Position{}, c.Size()
}

func (c *glCanvas) MinSize() fyne.Size {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	return c.canvasSize(c.content.MinSize())
}

func (c *glCanvas) OnKeyDown() func(*fyne.KeyEvent) {
	return c.onKeyDown
}

func (c *glCanvas) OnKeyUp() func(*fyne.KeyEvent) {
	return c.onKeyUp
}

func (c *glCanvas) OnTypedKey() func(*fyne.KeyEvent) {
	return c.onTypedKey
}

func (c *glCanvas) OnTypedRune() func(rune) {
	return c.onTypedRune
}

func (c *glCanvas) Padded() bool {
	return c.padded
}

func (c *glCanvas) PixelCoordinateForPosition(pos fyne.Position) (int, int) {
	c.Mu.RLock()
	texScale := c.texScale
	scale := c.scale
	c.Mu.RUnlock()
	multiple := scale * texScale
	scaleInt := func(x float32) int {
		return int(math.Round(float64(x * multiple)))
	}

	return scaleInt(pos.X), scaleInt(pos.Y)
}

func (c *glCanvas) Resize(size fyne.Size) {
	// This might not be the ideal solution, but it effectively avoid the first frame to be blurry due to the
	// rounding of the size to the loower integer when scale == 1. It does not affect the other cases as far as we tested.
	// This can easily be seen with fyne/cmd/hello and a scale == 1 as the text will happear blurry without the following line.
	nearestSize := fyne.NewSize(float32(math.Ceil(float64(size.Width))), float32(math.Ceil(float64(size.Height))))

	c.Mu.Lock()
	c.size = nearestSize
	c.Mu.Unlock()

	if c.webExtraWindows != nil {
		c.webExtraWindows.Resize(size)
	}
	for _, overlay := range c.Overlays().List() {
		if p, ok := overlay.(*widget.PopUp); ok {
			// TODO: remove this when #707 is being addressed.
			// “Notifies” the PopUp of the canvas size change.
			p.Refresh()
		} else {
			overlay.Resize(nearestSize)
		}
	}

	c.Mu.RLock()
	content := c.content
	contentSize := c.contentSize(nearestSize)
	contentPos := c.contentPos()
	menu := c.menu
	menuHeight := c.menuHeight()
	c.Mu.RUnlock()

	content.Resize(contentSize)
	content.Move(contentPos)

	if menu != nil {
		menu.Refresh()
		menu.Resize(fyne.NewSize(nearestSize.Width, menuHeight))
	}
}

func (c *glCanvas) Scale() float32 {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	return c.scale
}

func (c *glCanvas) SetContent(content fyne.CanvasObject) {
	content.Resize(content.MinSize()) // give it the space it wants then calculate the real min

	c.Mu.Lock()
	// the pass above makes some layouts wide enough to wrap, so we ask again what the true min is.
	newSize := c.size.Max(c.canvasSize(content.MinSize()))

	c.setContent(content)
	c.Mu.Unlock()

	c.Resize(newSize)
	c.SetDirty()
}

func (c *glCanvas) SetOnKeyDown(typed func(*fyne.KeyEvent)) {
	c.onKeyDown = typed
}

func (c *glCanvas) SetOnKeyUp(typed func(*fyne.KeyEvent)) {
	c.onKeyUp = typed
}

func (c *glCanvas) SetOnTypedKey(typed func(*fyne.KeyEvent)) {
	c.onTypedKey = typed
}

func (c *glCanvas) SetOnTypedRune(typed func(rune)) {
	c.onTypedRune = typed
}

func (c *glCanvas) SetPadded(padded bool) {
	c.Mu.Lock()
	content := c.content
	c.padded = padded
	pos := c.contentPos()
	c.Mu.Unlock()

	content.Move(pos)
}

func (c *glCanvas) reloadScale() {
	w := c.context.(*window)
	w.viewLock.RLock()
	windowVisible := w.visible
	w.viewLock.RUnlock()
	if !windowVisible {
		return
	}

	c.Mu.Lock()
	c.scale = w.calculatedScale()
	c.Mu.Unlock()
	c.SetDirty()

	c.context.RescaleContext()
}

func (c *glCanvas) Size() fyne.Size {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	return c.size
}

func (c *glCanvas) ToggleMenu() {
	c.Mu.RLock()
	menu := c.menu
	c.Mu.RUnlock()
	if menu != nil {
		menu.(*MenuBar).Toggle()
	}
}

func (c *glCanvas) buildMenu(w *window, m *fyne.MainMenu) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.setMenuOverlay(nil)
	if m == nil {
		return
	}
	if hasNativeMenu() {
		setupNativeMenu(w, m)
	} else {
		c.setMenuOverlay(buildMenuOverlay(m, w))
	}
}

// canvasSize computes the needed canvas size for the given content size
func (c *glCanvas) canvasSize(contentSize fyne.Size) fyne.Size {
	canvasSize := contentSize.Add(fyne.NewSize(0, c.menuHeight()))
	if c.Padded() {
		return canvasSize.Add(fyne.NewSquareSize(theme.Padding() * 2))
	}
	return canvasSize
}

func (c *glCanvas) contentPos() fyne.Position {
	contentPos := fyne.NewPos(0, c.menuHeight())
	if c.Padded() {
		return contentPos.Add(fyne.NewSquareOffsetPos(theme.Padding()))
	}
	return contentPos
}

func (c *glCanvas) contentSize(canvasSize fyne.Size) fyne.Size {
	contentSize := fyne.NewSize(canvasSize.Width, canvasSize.Height-c.menuHeight())
	if c.Padded() {
		return contentSize.Subtract(fyne.NewSquareSize(theme.Padding() * 2))
	}
	return contentSize
}

func (c *glCanvas) menuHeight() float32 {
	if c.menu == nil {
		return 0 // no menu or native menu -> does not consume space on the canvas
	}

	return c.menu.MinSize().Height
}

func (c *glCanvas) overlayChanged() {
	c.SetDirty()
}

func (c *glCanvas) paint(size fyne.Size) {
	clips := &internal.ClipStack{}
	if c.Content() == nil {
		return
	}
	c.Painter().Clear()

	paint := func(node *driver.RenderCacheNode, pos fyne.Position) {
		obj := node.Obj
		if _, ok := obj.(Scrollable); ok {
			inner := clips.Push(pos, obj.Size())
			c.Painter().StartClipping(inner.Rect())
		}
		if size.Width <= 0 || size.Height <= 0 { // iconifying on Windows can do bad things
			return
		}
		c.Painter().Paint(obj, pos, size)
	}
	afterPaint := func(node *driver.RenderCacheNode, pos fyne.Position) {
		if _, ok := node.Obj.(Scrollable); ok {
			clips.Pop()
			if top := clips.Top(); top != nil {
				c.Painter().StartClipping(top.Rect())
			} else {
				c.Painter().StopClipping()
			}
		}

		if build.Mode == fyne.BuildDebug {
			c.DrawDebugOverlay(node.Obj, pos, size)
		}
	}
	c.WalkTrees(paint, afterPaint)
}

func (c *glCanvas) setContent(content fyne.CanvasObject) {
	c.content = content
	c.SetContentTreeAndFocusMgr(content)
}

func (c *glCanvas) setMenuOverlay(b fyne.CanvasObject) {
	c.menu = b
	c.SetMenuTreeAndFocusMgr(b)

	if c.menu != nil && !c.size.IsZero() {
		c.content.Resize(c.contentSize(c.size))
		c.content.Move(c.contentPos())

		c.menu.Refresh()
		c.menu.Resize(fyne.NewSize(c.size.Width, c.menu.MinSize().Height))
	}
}

func (c *glCanvas) applyThemeOutOfTreeObjects() {
	c.Mu.RLock()
	menu := c.menu
	padded := c.padded
	c.Mu.RUnlock()
	if menu != nil {
		app.ApplyThemeTo(menu, c) // Ensure our menu gets the theme change message as it's out-of-tree
	}

	c.SetPadded(padded) // refresh the padding for potential theme differences
}

func newCanvas() *glCanvas {
	c := &glCanvas{scale: 1.0, texScale: 1.0, padded: true}
	c.Initialize()
	c.setContent(&canvas.Rectangle{FillColor: theme.Color(theme.ColorNameBackground)})
	return c
}
