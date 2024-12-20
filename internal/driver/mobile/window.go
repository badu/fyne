package mobile

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/internal/async"
	"fyne.io/fyne/v2/internal/cache"
	intdriver "fyne.io/fyne/v2/internal/driver"

	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"sync"
)

var DonePool = sync.Pool{
	New: func() any {
		return make(chan struct{})
	},
}

type window struct {
	eventQueue *async.UnboundedFuncChan

	title              string
	visible            bool
	onClosed           func()
	onCloseIntercepted func()
	isChild            bool

	clipboard fyne.Clipboard
	canvas    *mobileCanvas
	icon      fyne.Resource
	menu      *fyne.MainMenu
	handle    uintptr // the window handle - currently just Android
}

// DestroyEventQueue destroys the event queue.
func (w *window) DestroyEventQueue() {
	w.eventQueue.Close()
}

// InitEventQueue initializes the event queue.
func (w *window) InitEventQueue() {
	// This channel should be closed when the window is closed.
	w.eventQueue = async.NewUnboundedFuncChan()
}

// QueueEvent uses this method to queue up a callback that handles an event. This ensures
// user interaction events for a given window are processed in order.
func (w *window) QueueEvent(fn func()) {
	w.eventQueue.In() <- fn
}

// RunEventQueue runs the event queue. This should called inside a go routine.
// This function blocks.
func (w *window) RunEventQueue() {
	for fn := range w.eventQueue.Out() {
		fn()
	}
}

// WaitForEvents wait for all the events.
func (w *window) WaitForEvents() {
	done := DonePool.Get().(chan struct{})
	defer DonePool.Put(done)

	w.eventQueue.In() <- func() { done <- struct{}{} }
	<-done
}

func (w *window) Title() string {
	return w.title
}

func (w *window) SetTitle(title string) {
	w.title = title
}

func (w *window) FullScreen() bool {
	return true
}

func (w *window) SetFullScreen(bool) {
	// no-op
}

func (w *window) Resize(size fyne.Size) {
	w.Canvas().(*mobileCanvas).Resize(size)
}

func (w *window) RequestFocus() {
	// no-op - we cannot change which window is focused
}

func (w *window) FixedSize() bool {
	return true
}

func (w *window) SetFixedSize(bool) {
	// no-op - all windows are fixed size
}

func (w *window) CenterOnScreen() {
	// no-op
}

func (w *window) Padded() bool {
	return w.canvas.padded
}

func (w *window) SetPadded(padded bool) {
	w.canvas.padded = padded
}

func (w *window) Icon() fyne.Resource {
	if w.icon == nil {
		return fyne.CurrentApp().Icon()
	}

	return w.icon
}

func (w *window) SetIcon(icon fyne.Resource) {
	w.icon = icon
}

func (w *window) SetMaster() {
	// no-op on mobile
}

func (w *window) MainMenu() *fyne.MainMenu {
	return w.menu
}

func (w *window) SetMainMenu(menu *fyne.MainMenu) {
	w.menu = menu
}

func (w *window) SetOnClosed(callback func()) {
	w.onClosed = callback
}

func (w *window) SetCloseIntercept(callback func()) {
	w.onCloseIntercepted = callback
}

func (w *window) SetOnDropped(dropped func(fyne.Position, []fyne.URI)) {
	// not implemented yet
}

func (w *window) Show() {
	menu := fyne.CurrentDriver().(*MobileDriver).findMenu(w)
	menuButton := w.newMenuButton(menu)
	if menu == nil {
		menuButton.Hide()
	}

	if w.isChild {
		exit := widget.NewButton(
			widget.ButtonWithIcon(theme.CancelIcon()),
			widget.ButtonWithCallback(func() {
				w.tryClose()
			},
			),
		)
		title := widget.NewLabel(widget.LabelWithStaticText(w.title))
		title.Alignment = fyne.TextAlignCenter
		w.canvas.setWindowHead(container.NewHBox(menuButton,
			layout.NewSpacer(), title, layout.NewSpacer(), exit))
		w.canvas.Resize(w.canvas.size)
	} else {
		w.canvas.setWindowHead(container.NewHBox(menuButton))
	}
	w.visible = true

	if w.Content() != nil {
		w.Content().Refresh()
		w.Content().Show()
	}
}

func (w *window) Hide() {
	w.visible = false

	if w.Content() != nil {
		w.Content().Hide()
	}
}

func (w *window) tryClose() {
	if w.onCloseIntercepted != nil {
		w.QueueEvent(w.onCloseIntercepted)
		return
	}

	w.Close()
}

func (w *window) Close() {
	d := fyne.CurrentDriver().(*MobileDriver)
	pos := -1
	for i, win := range d.windows {
		if win == w {
			pos = i
		}
	}
	if pos != -1 {
		d.windows = append(d.windows[:pos], d.windows[pos+1:]...)
	}

	cache.RangeTexturesFor(w.canvas, w.canvas.Painter().Free)

	w.canvas.WalkTrees(nil, func(node *intdriver.RenderCacheNode, _ fyne.Position) {
		if wid, ok := node.Obj.(fyne.Widget); ok {
			cache.DestroyRenderer(wid)
		}
	})

	w.QueueEvent(func() {
		cache.CleanCanvas(w.canvas)
	})

	// Call this in a go routine, because this function could be called
	// inside a button which callback would be queued in this event queue
	// and it will lead to a deadlock if this is performed in the same go
	// routine.
	go w.DestroyEventQueue()

	if w.onClosed != nil {
		w.onClosed()
	}
}

func (w *window) ShowAndRun() {
	w.Show()
	fyne.CurrentApp().Run()
}

func (w *window) Content() fyne.CanvasObject {
	return w.canvas.Content()
}

func (w *window) SetContent(content fyne.CanvasObject) {
	w.canvas.SetContent(content)
}

func (w *window) Canvas() fyne.Canvas {
	return w.canvas
}

func (w *window) Clipboard() fyne.Clipboard {
	if w.clipboard == nil {
		w.clipboard = &mobileClipboard{}
	}
	return w.clipboard
}

func (w *window) RunWithContext(f func()) {
	//	ctx, _ = e.DrawContext.(gl.Context)

	f()
}

func (w *window) RescaleContext() {
	// TODO
}

func (w *window) Context() any {
	return fyne.CurrentDriver().(*MobileDriver).glctx
}
