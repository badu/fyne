package mobile

import (
	"context"
	"image"
	"image/color"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	fyneCanvas "fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/internal"
	"fyne.io/fyne/v2/internal/app"
	"fyne.io/fyne/v2/internal/async"
	"fyne.io/fyne/v2/internal/cache"
	intdriver "fyne.io/fyne/v2/internal/driver"
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

type mobileCanvas struct {
	sync.RWMutex

	OnFocus   func(obj fyne.Focusable)
	OnUnfocus func()

	impl SizeableCanvas

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

	mWindowHeadTree, contentTree, menuTree *renderCacheTree

	content        fyne.CanvasObject
	device         *device
	initialized    bool
	lastTapDown    map[int]time.Time
	lastTapDownPos map[int]fyne.Position
	menu           fyne.CanvasObject
	padded         bool
	scale          float32
	size           fyne.Size
	touched        map[int]mobile.Touchable
	windowHead     fyne.CanvasObject

	dragOffset fyne.Position
	dragStart  fyne.Position
	dragging   Draggable

	onTypedKey  func(event *fyne.KeyEvent)
	onTypedRune func(rune)

	touchCancelFunc context.CancelFunc
	touchLastTapped fyne.CanvasObject
	touchTapCount   int
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
	o.renderCaches = append(o.renderCaches, &renderCacheTree{root: &intdriver.RenderCacheNode{Obj: overlay}})
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
	root *intdriver.RenderCacheNode
}

func (c *mobileCanvas) updateLayout(objToLayout fyne.CanvasObject) {
	switch cont := objToLayout.(type) {
	case *fyne.Container:
		if cont.Layout != nil {
			layout := cont.Layout
			objects := cont.Objects
			c.RUnlock()
			layout.Layout(objects, cont.Size())
			c.RLock()
		}
	case fyne.Widget:
		renderer := cache.Renderer(cont)
		c.RUnlock()
		renderer.Layout(cont.Size())
		c.RLock()
	}
}

func newMobileCanvas(dev fyne.Device) fyne.Canvas {
	d, _ := dev.(*device)
	ret := &mobileCanvas{
		OnFocus:        handleKeyboard,
		OnUnfocus:      hideVirtualKeyboard,
		device:         d,
		lastTapDown:    make(map[int]time.Time),
		lastTapDownPos: make(map[int]fyne.Position),
		padded:         true,
		scale:          dev.SystemScaleForWindow(nil), // we don't need a window parameter on mobile,
		touched:        make(map[int]mobile.Touchable),
	}
	ret.Initialize(ret, ret.overlayChanged)
	return ret
}

// AddShortcut adds a shortcut to the canvas.
func (c *mobileCanvas) AddShortcut(shortcut fyne.Shortcut, handler func(shortcut fyne.Shortcut)) {
	c.shortcut.AddShortcut(shortcut, handler)
}

func (c *mobileCanvas) DrawDebugOverlay(obj fyne.CanvasObject, pos fyne.Position, size fyne.Size) {
	switch obj.(type) {
	case fyne.Widget:
		r := fyneCanvas.NewRectangle(color.Transparent)
		r.StrokeColor = color.NRGBA{R: 0xcc, G: 0x33, B: 0x33, A: 0xff}
		r.StrokeWidth = 1
		r.Resize(obj.Size())
		c.Painter().Paint(r, pos, size)

		t := fyneCanvas.NewText(reflect.ValueOf(obj).Elem().Type().Name(), r.StrokeColor)
		t.TextSize = 10
		c.Painter().Paint(t, pos.AddXY(2, 2), size)
	case *fyne.Container:
		r := fyneCanvas.NewRectangle(color.Transparent)
		r.StrokeColor = color.NRGBA{R: 0x33, G: 0x33, B: 0xcc, A: 0xff}
		r.StrokeWidth = 1
		r.Resize(obj.Size())
		c.Painter().Paint(r, pos, size)
	}
}

// EnsureMinSize ensure canvas min size.
//
// This function uses lock.
func (c *mobileCanvas) EnsureMinSize() bool {
	if c.impl.Content() == nil {
		return false
	}
	windowNeedsMinSizeUpdate := false
	csize := c.impl.Size()
	min := c.impl.MinSize()

	c.RLock()
	defer c.RUnlock()

	var parentNeedingUpdate *intdriver.RenderCacheNode

	ensureMinSize := func(node *intdriver.RenderCacheNode, pos fyne.Position) {
		obj := node.Obj
		cache.SetCanvasForObject(obj, c.impl, func() {
			if img, ok := obj.(*fyneCanvas.Image); ok {
				c.RUnlock()
				img.Refresh() // this may now have a different texScale
				c.RLock()
			}
		})

		if parentNeedingUpdate == node {
			c.updateLayout(obj)
			parentNeedingUpdate = nil
		}

		c.RUnlock()
		if !obj.Visible() {
			c.RLock()
			return
		}
		minSize := obj.MinSize()
		c.RLock()

		minSizeChanged := node.MinSize != minSize
		if minSizeChanged {
			node.MinSize = minSize
			if node.Parent != nil {
				parentNeedingUpdate = node.Parent
			} else {
				windowNeedsMinSizeUpdate = true
				c.RUnlock()
				size := obj.Size()
				c.RLock()
				expectedSize := minSize.Max(size)
				if expectedSize != size && size != csize {
					c.RUnlock()
					obj.Resize(expectedSize)
					c.RLock()
				} else {
					c.updateLayout(obj)
				}
			}
		}
	}
	c.WalkTrees(nil, ensureMinSize)

	shouldResize := windowNeedsMinSizeUpdate && (csize.Width < min.Width || csize.Height < min.Height)
	if shouldResize {
		c.RUnlock()
		c.impl.Resize(csize.Max(min))
		c.RLock()
	}
	return windowNeedsMinSizeUpdate
}

// Focus makes the provided item focused.
func (c *mobileCanvas) Focus(obj fyne.Focusable) {
	focusMgr := c.focusManager()
	if focusMgr != nil && focusMgr.Focus(obj) { // fast path – probably >99.9% of all cases
		if c.OnFocus != nil {
			c.OnFocus(obj)
		}
		return
	}

	c.RLock()
	focusMgrs := append([]*app.FocusManager{c.contentFocusMgr, c.menuFocusMgr}, c.overlays.ListFocusManagers()...)
	c.RUnlock()

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
func (c *mobileCanvas) Focused() fyne.Focusable {
	mgr := c.focusManager()
	if mgr == nil {
		return nil
	}
	return mgr.Focused()
}

// FocusGained signals to the manager that its content got focus.
// Valid only on Desktop.
func (c *mobileCanvas) FocusGained() {
	mgr := c.focusManager()
	if mgr == nil {
		return
	}
	mgr.FocusGained()
}

// FocusLost signals to the manager that its content lost focus.
// Valid only on Desktop.
func (c *mobileCanvas) FocusLost() {
	mgr := c.focusManager()
	if mgr == nil {
		return
	}
	mgr.FocusLost()
}

// FocusNext focuses the next focusable item.
func (c *mobileCanvas) FocusNext() {
	mgr := c.focusManager()
	if mgr == nil {
		return
	}
	mgr.FocusNext()
}

// FocusPrevious focuses the previous focusable item.
func (c *mobileCanvas) FocusPrevious() {
	mgr := c.focusManager()
	if mgr == nil {
		return
	}
	mgr.FocusPrevious()
}

// FreeDirtyTextures frees dirty textures and returns the number of freed textures.
func (c *mobileCanvas) FreeDirtyTextures() (freed uint64) {
	freeObject := func(object fyne.CanvasObject) {
		freeWalked := func(obj fyne.CanvasObject, _ fyne.Position, _ fyne.Position, _ fyne.Size) bool {
			// No image refresh while recursing to avoid double texture upload.
			if _, ok := obj.(*fyneCanvas.Image); ok {
				return false
			}
			if c.painter != nil {
				c.painter.Free(obj)
			}
			return false
		}

		// Image.Refresh will trigger a refresh specific to the object, while recursing on parent widget would just lead to
		// a double texture upload.
		if img, ok := object.(*fyneCanvas.Image); ok {
			if c.painter != nil {
				c.painter.Free(img)
			}
		} else {
			intdriver.WalkCompleteObjectTree(object, freeWalked, nil)
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
		cache.RangeExpiredTexturesFor(c.impl, c.painter.Free)
	}
	return
}

// Initialize initializes the canvas.
func (c *mobileCanvas) Initialize(impl SizeableCanvas, onOverlayChanged func()) {
	c.impl = impl
	c.refreshQueue = async.NewCanvasObjectQueue()
	c.overlays = &overlayStack{
		OverlayStack: internal.OverlayStack{
			OnChange: onOverlayChanged,
			Canvas:   impl,
		},
	}
}

// ObjectTrees return canvas object trees.
//
// This function uses lock.
func (c *mobileCanvas) ObjectTrees() []fyne.CanvasObject {
	c.RLock()
	var content, menu fyne.CanvasObject
	if c.contentTree != nil && c.contentTree.root != nil {
		content = c.contentTree.root.Obj
	}
	if c.menuTree != nil && c.menuTree.root != nil {
		menu = c.menuTree.root.Obj
	}
	c.RUnlock()
	trees := make([]fyne.CanvasObject, 0, len(c.Overlays().List())+2)
	trees = append(trees, content)
	if menu != nil {
		trees = append(trees, menu)
	}
	trees = append(trees, c.Overlays().List()...)
	return trees
}

// Overlays returns the overlay stack.
func (c *mobileCanvas) Overlays() fyne.OverlayStack {
	// we don't need to lock here, because overlays never changes
	return c.overlays
}

// Painter returns the canvas painter.
func (c *mobileCanvas) Painter() gl.Painter {
	return c.painter
}

// Refresh refreshes a canvas object.
func (c *mobileCanvas) Refresh(obj fyne.CanvasObject) {
	walkNeeded := false
	switch obj.(type) {
	case *fyne.Container:
		walkNeeded = true
	case fyne.Widget:
		walkNeeded = true
	}

	if walkNeeded {
		intdriver.WalkCompleteObjectTree(obj, func(co fyne.CanvasObject, p1, p2 fyne.Position, s fyne.Size) bool {
			if i, ok := co.(*fyneCanvas.Image); ok {
				i.Refresh()
			}
			return false
		}, nil)
	}

	c.refreshQueue.In(obj)
	c.SetDirty()
}

// RemoveShortcut removes a shortcut from the canvas.
func (c *mobileCanvas) RemoveShortcut(shortcut fyne.Shortcut) {
	c.shortcut.RemoveShortcut(shortcut)
}

// SetContentTreeAndFocusMgr sets content tree and focus manager.
//
// This function does not use the canvas lock.
func (c *mobileCanvas) SetContentTreeAndFocusMgr(content fyne.CanvasObject) {
	c.contentTree = &renderCacheTree{root: &intdriver.RenderCacheNode{Obj: content}}
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
func (c *mobileCanvas) CheckDirtyAndClear() bool {
	return c.dirty.Swap(false)
}

// SetDirty sets canvas dirty flag atomically.
func (c *mobileCanvas) SetDirty() {
	c.dirty.Store(true)
}

// SetMenuTreeAndFocusMgr sets menu tree and focus manager.
//
// This function does not use the canvas lock.
func (c *mobileCanvas) SetMenuTreeAndFocusMgr(menu fyne.CanvasObject) {
	c.menuTree = &renderCacheTree{root: &intdriver.RenderCacheNode{Obj: menu}}
	if menu != nil {
		c.menuFocusMgr = app.NewFocusManager(menu)
	} else {
		c.menuFocusMgr = nil
	}
}

// SetMobileWindowHeadTree sets window head tree.
//
// This function does not use the canvas lock.
func (c *mobileCanvas) SetMobileWindowHeadTree(head fyne.CanvasObject) {
	c.mWindowHeadTree = &renderCacheTree{root: &intdriver.RenderCacheNode{Obj: head}}
}

// SetPainter sets the canvas painter.
func (c *mobileCanvas) SetPainter(p gl.Painter) {
	c.painter = p
}

// TypedShortcut handle the registered shortcut.
func (c *mobileCanvas) TypedShortcut(shortcut fyne.Shortcut) {
	c.shortcut.TypedShortcut(shortcut)
}

// Unfocus unfocuses all the objects in the canvas.
func (c *mobileCanvas) Unfocus() {
	mgr := c.focusManager()
	if mgr == nil {
		return
	}
	if mgr.Focus(nil) && c.OnUnfocus != nil {
		c.OnUnfocus()
	}
}

// WalkTrees walks over the trees.
func (c *mobileCanvas) WalkTrees(
	beforeChildren func(*intdriver.RenderCacheNode, fyne.Position),
	afterChildren func(*intdriver.RenderCacheNode, fyne.Position),
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

func (c *mobileCanvas) focusManager() *app.FocusManager {
	if focusMgr := c.overlays.TopFocusManager(); focusMgr != nil {
		return focusMgr
	}
	c.RLock()
	defer c.RUnlock()
	if c.isMenuActive() {
		return c.menuFocusMgr
	}
	return c.contentFocusMgr
}

func (c *mobileCanvas) isMenuActive() bool {
	if c.menuTree == nil || c.menuTree.root == nil || c.menuTree.root.Obj == nil {
		return false
	}
	menu := c.menuTree.root.Obj
	if am, ok := menu.(activatableMenu); ok {
		return am.IsActive()
	}
	return true
}

func (c *mobileCanvas) walkTree(
	tree *renderCacheTree,
	beforeChildren func(*intdriver.RenderCacheNode, fyne.Position),
	afterChildren func(*intdriver.RenderCacheNode, fyne.Position),
) {
	tree.Lock()
	defer tree.Unlock()
	var node, parent, prev *intdriver.RenderCacheNode
	node = tree.root

	bc := func(obj fyne.CanvasObject, pos fyne.Position, _ fyne.Position, _ fyne.Size) bool {
		if node != nil && node.Obj != obj {
			if parent.FirstChild == node {
				parent.FirstChild = nil
			}
			node = nil
		}
		if node == nil {
			node = &intdriver.RenderCacheNode{Parent: parent, Obj: obj}
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
	intdriver.WalkVisibleObjectTree(tree.root.Obj, bc, ac)
}

func (c *mobileCanvas) Capture() image.Image {
	return c.Painter().Capture(c)
}

func (c *mobileCanvas) Content() fyne.CanvasObject {
	return c.content
}

func (c *mobileCanvas) InteractiveArea() (fyne.Position, fyne.Size) {
	var pos fyne.Position
	var size fyne.Size
	if c.device == nil {
		// running in test mode
		size = c.Size()
	} else {
		safeLeft := float32(c.device.safeLeft) / c.scale
		safeTop := float32(c.device.safeTop) / c.scale
		safeRight := float32(c.device.safeRight) / c.scale
		safeBottom := float32(c.device.safeBottom) / c.scale
		pos = fyne.NewPos(safeLeft, safeTop)
		size = c.size.SubtractWidthHeight(safeLeft+safeRight, safeTop+safeBottom)
	}
	if c.windowHeadIsDisplacing() {
		offset := c.windowHead.MinSize().Height
		pos = pos.AddXY(0, offset)
		size = size.SubtractWidthHeight(0, offset)
	}
	return pos, size
}

func (c *mobileCanvas) MinSize() fyne.Size {
	return c.size // TODO check
}

func (c *mobileCanvas) OnTypedKey() func(*fyne.KeyEvent) {
	return c.onTypedKey
}

func (c *mobileCanvas) OnTypedRune() func(rune) {
	return c.onTypedRune
}

func (c *mobileCanvas) PixelCoordinateForPosition(pos fyne.Position) (int, int) {
	return int(float32(pos.X) * c.scale), int(float32(pos.Y) * c.scale)
}

func (c *mobileCanvas) Resize(size fyne.Size) {
	if size == c.size {
		return
	}

	c.sizeContent(size)
}

func (c *mobileCanvas) Scale() float32 {
	return c.scale
}

func (c *mobileCanvas) SetContent(content fyne.CanvasObject) {
	c.setContent(content)
	c.sizeContent(c.Size()) // fixed window size for mobile, cannot stretch to new content
	c.SetDirty()
}

func (c *mobileCanvas) SetOnTypedKey(typed func(*fyne.KeyEvent)) {
	c.onTypedKey = typed
}

func (c *mobileCanvas) SetOnTypedRune(typed func(rune)) {
	c.onTypedRune = typed
}

func (c *mobileCanvas) Size() fyne.Size {
	return c.size
}

func (c *mobileCanvas) applyThemeOutOfTreeObjects() {
	if c.menu != nil {
		app.ApplyThemeTo(c.menu, c) // Ensure our menu gets the theme change message as it's out-of-tree
	}
	if c.windowHead != nil {
		app.ApplyThemeTo(c.windowHead, c) // Ensure our child windows get the theme change message as it's out-of-tree
	}
}

func (c *mobileCanvas) findObjectAtPositionMatching(pos fyne.Position, test func(object fyne.CanvasObject) bool) (fyne.CanvasObject, fyne.Position, int) {
	if c.menu != nil {
		return intdriver.FindObjectAtPositionMatching(pos, test, c.Overlays().Top(), c.menu)
	}

	return intdriver.FindObjectAtPositionMatching(pos, test, c.Overlays().Top(), c.windowHead, c.content)
}

func (c *mobileCanvas) overlayChanged() {
	handleKeyboard(c.Focused())
	c.SetDirty()
}

func (c *mobileCanvas) setContent(content fyne.CanvasObject) {
	c.content = content
	c.SetContentTreeAndFocusMgr(content)
}

func (c *mobileCanvas) setMenu(menu fyne.CanvasObject) {
	c.menu = menu
	c.SetMenuTreeAndFocusMgr(menu)
}

func (c *mobileCanvas) setWindowHead(head fyne.CanvasObject) {
	if c.padded {
		head = container.NewPadded(head)
	}
	c.windowHead = head
	c.SetMobileWindowHeadTree(head)
}

func (c *mobileCanvas) sizeContent(size fyne.Size) {
	if c.content == nil { // window may not be configured yet
		return
	}

	c.size = size
	areaPos, areaSize := c.InteractiveArea()

	if c.windowHead != nil {
		var headSize fyne.Size
		headPos := areaPos
		if c.windowHeadIsDisplacing() {
			headSize = fyne.NewSize(areaSize.Width, c.windowHead.MinSize().Height)
			headPos = headPos.SubtractXY(0, headSize.Height)
		} else {
			headSize = c.windowHead.MinSize()
		}
		c.windowHead.Resize(headSize)
		c.windowHead.Move(headPos)
	}

	for _, overlay := range c.Overlays().List() {
		if p, ok := overlay.(*widget.PopUp); ok {
			// TODO: remove this when #707 is being addressed.
			// “Notifies” the PopUp of the canvas size change.
			p.Refresh()
		} else {
			overlay.Resize(areaSize)
			overlay.Move(areaPos)
		}
	}

	if c.padded {
		c.content.Resize(areaSize.Subtract(fyne.NewSize(theme.Padding()*2, theme.Padding()*2)))
		c.content.Move(areaPos.Add(fyne.NewPos(theme.Padding(), theme.Padding())))
	} else {
		c.content.Resize(areaSize)
		c.content.Move(areaPos)
	}
}

func (c *mobileCanvas) tapDown(pos fyne.Position, tapID int) {
	c.lastTapDown[tapID] = time.Now()
	c.lastTapDownPos[tapID] = pos
	c.dragging = nil

	co, objPos, layer := c.findObjectAtPositionMatching(pos, func(object fyne.CanvasObject) bool {
		switch object.(type) {
		case mobile.Touchable, fyne.Focusable:
			return true
		}

		return false
	})

	if wid, ok := co.(mobile.Touchable); ok {
		touchEv := &mobile.TouchEvent{}
		touchEv.Position = objPos
		touchEv.AbsolutePosition = pos
		wid.TouchDown(touchEv)
		c.touched[tapID] = wid
	}

	if layer != 1 { // 0 - overlay, 1 - window head / menu, 2 - content
		if wid, ok := co.(fyne.Focusable); !ok || wid != c.Focused() {
			c.Unfocus()
		}
	}
}

func (c *mobileCanvas) tapMove(pos fyne.Position, tapID int,
	dragCallback func(Draggable, *fyne.DragEvent)) {
	previousPos := c.lastTapDownPos[tapID]
	deltaX := pos.X - previousPos.X
	deltaY := pos.Y - previousPos.Y

	if c.dragging == nil && (math.Abs(float64(deltaX)) < tapMoveThreshold && math.Abs(float64(deltaY)) < tapMoveThreshold) {
		return
	}
	c.lastTapDownPos[tapID] = pos

	co, objPos, _ := c.findObjectAtPositionMatching(pos, func(object fyne.CanvasObject) bool {
		if _, ok := object.(Draggable); ok {
			return true
		} else if _, ok := object.(mobile.Touchable); ok {
			return true
		}

		return false
	})

	if c.touched[tapID] != nil {
		if touch, ok := co.(mobile.Touchable); !ok || c.touched[tapID] != touch {
			touchEv := &mobile.TouchEvent{}
			touchEv.Position = objPos
			touchEv.AbsolutePosition = pos
			c.touched[tapID].TouchCancel(touchEv)
			c.touched[tapID] = nil
		}
	}

	if c.dragging == nil {
		if drag, ok := co.(Draggable); ok {
			c.dragging = drag
			c.dragOffset = previousPos.Subtract(objPos)
			c.dragStart = co.Position()
		} else {
			return
		}
	}

	ev := &fyne.DragEvent{}
	draggedObjDelta := c.dragStart.Subtract(c.dragging.(fyne.CanvasObject).Position())
	ev.Position = pos.Subtract(c.dragOffset).Add(draggedObjDelta)
	ev.Dragged = fyne.Delta{DX: deltaX, DY: deltaY}

	dragCallback(c.dragging, ev)
}

func (c *mobileCanvas) tapUp(pos fyne.Position, tapID int,
	tapCallback func(Tappable, *fyne.PointEvent),
	tapAltCallback func(SecondaryTappable, *fyne.PointEvent),
	doubleTapCallback func(DoubleTappable, *fyne.PointEvent),
	dragCallback func(Draggable)) {

	if c.dragging != nil {
		dragCallback(c.dragging)

		c.dragging = nil
		return
	}

	duration := time.Since(c.lastTapDown[tapID])

	if c.menu != nil && c.Overlays().Top() == nil && pos.X > c.menu.Size().Width {
		c.menu.Hide()
		c.menu.Refresh()
		c.setMenu(nil)
		return
	}

	co, objPos, _ := c.findObjectAtPositionMatching(pos, func(object fyne.CanvasObject) bool {
		if _, ok := object.(Tappable); ok {
			return true
		} else if _, ok := object.(SecondaryTappable); ok {
			return true
		} else if _, ok := object.(mobile.Touchable); ok {
			return true
		} else if _, ok := object.(DoubleTappable); ok {
			return true
		}

		return false
	})

	if wid, ok := co.(mobile.Touchable); ok {
		touchEv := &mobile.TouchEvent{}
		touchEv.Position = objPos
		touchEv.AbsolutePosition = pos
		wid.TouchUp(touchEv)
		c.touched[tapID] = nil
	}

	ev := &fyne.PointEvent{
		Position:         objPos,
		AbsolutePosition: pos,
	}

	if duration < tapSecondaryDelay {
		_, doubleTap := co.(DoubleTappable)
		if doubleTap {
			c.touchTapCount++
			c.touchLastTapped = co
			if c.touchCancelFunc != nil {
				c.touchCancelFunc()
				return
			}
			go c.waitForDoubleTap(co, ev, tapCallback, doubleTapCallback)
		} else {
			if wid, ok := co.(Tappable); ok {
				tapCallback(wid, ev)
			}
		}
	} else {
		if wid, ok := co.(SecondaryTappable); ok {
			tapAltCallback(wid, ev)
		}
	}
}

func (c *mobileCanvas) waitForDoubleTap(co fyne.CanvasObject, ev *fyne.PointEvent, tapCallback func(Tappable, *fyne.PointEvent), doubleTapCallback func(DoubleTappable, *fyne.PointEvent)) {
	var ctx context.Context
	ctx, c.touchCancelFunc = context.WithDeadline(context.TODO(), time.Now().Add(tapDoubleDelay))
	defer c.touchCancelFunc()
	<-ctx.Done()
	if c.touchTapCount == 2 && c.touchLastTapped == co {
		if wid, ok := co.(DoubleTappable); ok {
			doubleTapCallback(wid, ev)
		}
	} else {
		if wid, ok := co.(Tappable); ok {
			tapCallback(wid, ev)
		}
	}
	c.touchTapCount = 0
	c.touchCancelFunc = nil
	c.touchLastTapped = nil
}

func (c *mobileCanvas) windowHeadIsDisplacing() bool {
	if c.windowHead == nil {
		return false
	}

	chromeBox := c.windowHead.(*fyne.Container)
	if c.padded {
		chromeBox = chromeBox.Objects[0].(*fyne.Container) // the padded container
	}
	return len(chromeBox.Objects) > 1
}
