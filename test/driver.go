package test

import (
	"image"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/driver"
	"fyne.io/fyne/v2/internal/painter"
	"fyne.io/fyne/v2/internal/painter/software"
	intRepo "fyne.io/fyne/v2/internal/repository"
	"fyne.io/fyne/v2/storage/repository"
)

// SoftwarePainter describes a simple type that can render canvases
type SoftwarePainter interface {
	Paint(fyne.Canvas) image.Image
}

type testDriver struct {
	device       device
	painter      SoftwarePainter
	windows      []fyne.Window
	windowsMutex sync.RWMutex
}

// NewDriver sets up and registers a new dummy driver for test purpose
func NewDriver() fyne.Driver {
	drv := &testDriver{windowsMutex: sync.RWMutex{}}
	repository.Register("file", intRepo.NewFileRepository())

	httpHandler := intRepo.NewHTTPRepository()
	repository.Register("http", httpHandler)
	repository.Register("https", httpHandler)

	// make a single dummy window for rendering tests
	drv.CreateWindow("")

	return drv
}

// NewDriverWithPainter creates a new dummy driver that will pass the given
// painter to all canvases created
func NewDriverWithPainter(painter SoftwarePainter) fyne.Driver {
	return &testDriver{painter: painter}
}

func (d *testDriver) AbsolutePositionForObject(co fyne.CanvasObject) fyne.Position {
	c := d.CanvasForObject(co)
	if c == nil {
		return fyne.NewPos(0, 0)
	}

	tc := c.(*testCanvas)
	pos := driver.AbsolutePositionForObject(co, tc.objectTrees())
	inset, _ := c.InteractiveArea()
	return pos.Subtract(inset)
}

func (d *testDriver) AllWindows() []fyne.Window {
	d.windowsMutex.RLock()
	defer d.windowsMutex.RUnlock()
	return d.windows
}

func (d *testDriver) CanvasForObject(fyne.CanvasObject) fyne.Canvas {
	d.windowsMutex.RLock()
	defer d.windowsMutex.RUnlock()
	// cheating: probably the last created window is meant
	return d.windows[len(d.windows)-1].Canvas()
}

func (d *testDriver) CreateWindow(title string) fyne.Window {
	c := NewCanvas().(*testCanvas)
	if d.painter != nil {
		c.painter = d.painter
	} else {
		c.painter = software.NewPainter()
	}

	w := &window{canvas: c, driver: d, title: title}

	d.windowsMutex.Lock()
	d.windows = append(d.windows, w)
	d.windowsMutex.Unlock()
	return w
}

func (d *testDriver) Device() fyne.Device {
	return &d.device
}

// RenderedTextSize looks up how bit a string would be if drawn on screen
func (d *testDriver) RenderedTextSize(text string, size float32, style fyne.TextStyle, source fyne.Resource) (fyne.Size, float32) {
	return painter.RenderedTextSize(text, size, style, source)
}

func (d *testDriver) Run() {
	// no-op
}

func (d *testDriver) StartAnimation(a *fyne.Animation) {
	// currently no animations in test app, we just initialise it and leave
	a.Tick(1.0)
}

func (d *testDriver) StopAnimation(a *fyne.Animation) {
	// currently no animations in test app, do nothing
}

func (d *testDriver) Quit() {
	// no-op
}

func (d *testDriver) removeWindow(w *window) {
	d.windowsMutex.Lock()
	i := 0
	for _, win := range d.windows {
		if win == w {
			break
		}
		i++
	}

	copy(d.windows[i:], d.windows[i+1:])
	d.windows[len(d.windows)-1] = nil // Allow the garbage collector to reclaim the memory.
	d.windows = d.windows[:len(d.windows)-1]

	d.windowsMutex.Unlock()
}

func (d *testDriver) DoubleTapDelay() time.Duration {
	return 300 * time.Millisecond
}

func (d *testDriver) SetDisableScreenBlanking(_ bool) {
	// no-op for test
}
