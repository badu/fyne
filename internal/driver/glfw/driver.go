// Package glfw provides a full Fyne desktop driver that uses the system OpenGL libraries.
// This supports Windows, Mac OS X and Linux using the gl and glfw packages from go-gl.
package glfw

import (
	"bytes"
	"image"
	"os"
	"runtime"
	"sync"

	intRepo "fyne.io/fyne/v2/internal/repository"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/animation"
	"fyne.io/fyne/v2/internal/app"
	"fyne.io/fyne/v2/internal/cache"
	"fyne.io/fyne/v2/internal/driver"
	"fyne.io/fyne/v2/internal/painter"
	"fyne.io/fyne/v2/storage/repository"

	"github.com/fyne-io/image/ico"
)

// mainGoroutineID stores the main goroutine ID.
// This ID must be initialized in main.init because
// a main goroutine may not equal to 1 due to the
// influence of a garbage collector.
var mainGoroutineID uint64

var curWindow *window

// A workaround on Apple M1/M2, just use 1 thread until fixed upstream.
const drawOnMainThread bool = runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"

type GLDriver struct {
	windowLock   sync.RWMutex
	windows      []fyne.Window
	done         chan struct{}
	drawDone     chan struct{}
	waitForStart chan struct{}

	animation animation.Runner

	currentKeyModifiers fyne.KeyModifier // desktop driver only

	trayStart, trayStop func()     // shut down the system tray, if used
	systrayMenu         *fyne.Menu // cache the menu set so we know when to refresh
}

func toOSIcon(icon []byte) ([]byte, error) {
	if runtime.GOOS != "windows" {
		return icon, nil
	}

	img, _, err := image.Decode(bytes.NewReader(icon))
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	err = ico.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *GLDriver) RenderedTextSize(text string, textSize float32, style fyne.TextStyle, source fyne.Resource) (size fyne.Size, baseline float32) {
	return painter.RenderedTextSize(text, textSize, style, source)
}

// CanvasForObject returns the canvas for the specified object.
func CanvasForObject(obj fyne.CanvasObject) fyne.Canvas {
	return cache.GetCanvasForObject(obj)
}

func (d *GLDriver) CanvasForObject(obj fyne.CanvasObject) fyne.Canvas {
	return CanvasForObject(obj)
}

func (d *GLDriver) AbsolutePositionForObject(co fyne.CanvasObject) fyne.Position {
	c := CanvasForObject(co)
	if c == nil {
		return fyne.NewPos(0, 0)
	}

	glc := c.(*glCanvas)
	return driver.AbsolutePositionForObject(co, glc.ObjectTrees())
}

func (d *GLDriver) Device() fyne.Device {
	return &glDevice{}
}

func (d *GLDriver) Quit() {
	if curWindow != nil {
		if f := fyne.CurrentApp().Lifecycle().(*app.Lifecycle).OnExitedForeground(); f != nil {
			curWindow.QueueEvent(f)
		}
		curWindow = nil
		if d.trayStop != nil {
			d.trayStop()
		}
	}

	// Only call close once to avoid panic.
	if running.CompareAndSwap(true, false) {
		close(d.done)
	}
}

func (d *GLDriver) addWindow(w *window) {
	d.windowLock.Lock()
	defer d.windowLock.Unlock()
	d.windows = append(d.windows, w)
}

// a trivial implementation of "focus previous" - return to the most recently opened, or master if set.
// This may not do the right thing if your app has 3 or more windows open, but it was agreed this was not much
// of an issue, and the added complexity to track focus was not needed at this time.
func (d *GLDriver) focusPreviousWindow() {
	d.windowLock.RLock()
	wins := d.windows
	d.windowLock.RUnlock()

	var chosen *window
	for _, w := range wins {
		win := w.(*window)
		if !win.visible {
			continue
		}
		chosen = win
		if win.master {
			break
		}
	}

	if chosen == nil || chosen.view() == nil {
		return
	}
	chosen.RequestFocus()
}

func (d *GLDriver) windowList() []fyne.Window {
	d.windowLock.RLock()
	defer d.windowLock.RUnlock()
	return d.windows
}

func (d *GLDriver) initFailed(msg string, err error) {
	logError(msg, err)

	if !running.Load() {
		d.Quit()
	} else {
		os.Exit(1)
	}
}

func (d *GLDriver) Run() {
	if goroutineID() != mainGoroutineID {
		panic("Run() or ShowAndRun() must be called from main goroutine")
	}

	go d.catchTerm()
	d.runGL()

	// Ensure lifecycle events run to completion before the app exits
	l := fyne.CurrentApp().Lifecycle().(*app.Lifecycle)
	l.WaitForEvents()
	l.DestroyEventQueue()
}

func (d *GLDriver) SetDisableScreenBlanking(disable bool) {
	setDisableScreenBlank(disable)
}

// NewGLDriver sets up a new Driver instance implemented using the GLFW Go library and OpenGL bindings.
func NewGLDriver() *GLDriver {
	repository.Register("file", intRepo.NewFileRepository())

	return &GLDriver{
		done:         make(chan struct{}),
		drawDone:     make(chan struct{}, 1),
		waitForStart: make(chan struct{}),
	}
}
