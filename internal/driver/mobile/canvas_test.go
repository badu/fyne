//go:build !windows || !ci

package mobile

import (
	"image/color"
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fynecanvas "fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/mobile"
	_ "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	currentApp := fyne.CurrentApp()
	fyne.SetCurrentApp(newTestMobileApp())
	ret := m.Run()
	fyne.SetCurrentApp(currentApp)
	os.Exit(ret)
}

func Test_canvas_ChildMinSizeChangeAffectsAncestorsUpToRoot(t *testing.T) {
	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)
	leftObj1 := fynecanvas.NewRectangle(color.Black)
	leftObj1.SetMinSize(fyne.NewSize(100, 50))
	leftObj2 := fynecanvas.NewRectangle(color.Black)
	leftObj2.SetMinSize(fyne.NewSize(100, 50))
	leftCol := container.NewVBox(leftObj1, leftObj2)
	rightObj1 := fynecanvas.NewRectangle(color.Black)
	rightObj1.SetMinSize(fyne.NewSize(100, 50))
	rightObj2 := fynecanvas.NewRectangle(color.Black)
	rightObj2.SetMinSize(fyne.NewSize(100, 50))
	rightCol := container.NewVBox(rightObj1, rightObj2)
	content := container.NewHBox(leftCol, rightCol)
	c.SetContent(container.NewCenter(content))
	c.Resize(fyne.NewSize(300, 300))

	oldContentSize := fyne.NewSize(200+theme.Padding(), 100+theme.Padding())
	assert.Equal(t, oldContentSize, content.Size())

	leftObj1.SetMinSize(fyne.NewSize(110, 60))
	c.EnsureMinSize()

	expectedContentSize := oldContentSize.Add(fyne.NewSize(10, 10))
	assert.Equal(t, expectedContentSize, content.Size())
}

func Test_canvas_Dragged(t *testing.T) {
	dragged := false
	var draggedObj Draggable
	scroll := container.NewScroll(widget.NewLabel(widget.LabelWithStaticText("Hi\nHi\nHi")))
	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)
	c.SetContent(scroll)
	c.Resize(fyne.NewSize(40, 24))
	assert.Equal(t, float32(0), scroll.Offset.Y)

	c.tapDown(fyne.NewPos(32, 3), 0)
	c.tapMove(fyne.NewPos(32, 10), 0, func(wid Draggable, ev *fyne.DragEvent) {
		wid.Dragged(ev)
		dragged = true
		draggedObj = wid
	})

	assert.True(t, dragged)
	assert.Equal(t, scroll, draggedObj)
	dragged = false
	c.tapMove(fyne.NewPos(32, 5), 0, func(wid Draggable, ev *fyne.DragEvent) {
		wid.Dragged(ev)
		dragged = true
	})
	assert.True(t, dragged)
	assert.Equal(t, fyne.NewPos(0, 5), scroll.Offset)
}

func Test_canvas_DraggingOutOfWidget(t *testing.T) {
	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)
	slider := widget.NewSlider(0.0, 100.0)
	c.SetContent(container.NewGridWithRows(2, slider, widget.NewLabel(widget.LabelWithStaticText("Outside"))))
	c.Resize(fyne.NewSize(100, 200))

	assert.Zero(t, slider.Value)
	lastValue := slider.Value

	dragged := false
	c.tapDown(fyne.NewPos(23, 13), 0)
	c.tapMove(fyne.NewPos(30, 13), 0, func(wid Draggable, ev *fyne.DragEvent) {
		assert.Equal(t, slider, wid)
		wid.Dragged(ev)
		dragged = true
	})
	assert.True(t, dragged)
	assert.Greater(t, slider.Value, lastValue)
	lastValue = slider.Value

	dragged = false
	c.tapMove(fyne.NewPos(40, 120), 0, func(wid Draggable, ev *fyne.DragEvent) {
		assert.Equal(t, slider, wid)
		wid.Dragged(ev)
		dragged = true
	})
	assert.True(t, dragged)
	assert.Greater(t, slider.Value, lastValue)
}

func Test_canvas_Focusable(t *testing.T) {
	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)
	content := newFocusableEntry(c)
	c.SetContent(content)
	content.Resize(fyne.NewSize(25, 25))

	pos := fyne.NewPos(10, 10)
	c.tapDown(pos, 0)
	c.tapUp(pos, 0, func(wid Tappable, ev *fyne.PointEvent) {
		wid.Tapped(ev)
	}, nil, nil, nil)
	time.Sleep(tapDoubleDelay + 150*time.Millisecond)
	assert.Equal(t, 1, content.focusedTimes)
	assert.Equal(t, 0, content.unfocusedTimes)

	c.tapDown(pos, 1)
	c.tapUp(pos, 1, func(wid Tappable, ev *fyne.PointEvent) {
		wid.Tapped(ev)
	}, nil, nil, nil)
	time.Sleep(tapDoubleDelay + 150*time.Millisecond)
	assert.Equal(t, 1, content.focusedTimes)
	assert.Equal(t, 0, content.unfocusedTimes)

	c.Focus(content)
	assert.Equal(t, 1, content.focusedTimes)
	assert.Equal(t, 0, content.unfocusedTimes)

	c.Unfocus()
	assert.Equal(t, 1, content.focusedTimes)
	assert.Equal(t, 1, content.unfocusedTimes)

	content.Disable()
	c.Focus(content)
	assert.Equal(t, 1, content.focusedTimes)
	assert.Equal(t, 1, content.unfocusedTimes)

	c.tapDown(fyne.NewPos(10, 10), 2)
	assert.Equal(t, 1, content.focusedTimes)
	assert.Equal(t, 1, content.unfocusedTimes)
}

func Test_canvas_InteractiveArea(t *testing.T) {
	dev := &device{
		safeTop:    17,
		safeLeft:   42,
		safeBottom: 71,
		safeRight:  24,
	}
	scale := dev.SystemScaleForWindow(nil)

	c := newMobileCanvas(dev)
	c.SetContent(fynecanvas.NewRectangle(color.Black))
	canvasSize := fyne.NewSize(600, 800)
	c.(*mobileCanvas).Resize(canvasSize)

	t.Run("for canvas with size", func(t *testing.T) {
		pos, size := c.InteractiveArea()
		assert.Equal(t, fyne.NewPos(float32(dev.safeLeft)/scale, float32(dev.safeTop)/scale), pos)
		assert.Equal(t, canvasSize.SubtractWidthHeight(float32(dev.safeLeft+dev.safeRight)/scale, float32(dev.safeTop+dev.safeBottom)/scale), size)
	})

	t.Run("when canvas size has changed", func(t *testing.T) {
		changedCanvasSize := fyne.NewSize(800, 600)
		c.(*mobileCanvas).Resize(changedCanvasSize)
		pos, size := c.InteractiveArea()
		assert.Equal(t, fyne.NewPos(float32(dev.safeLeft)/scale, float32(dev.safeTop)/scale), pos)
		assert.Equal(t, changedCanvasSize.SubtractWidthHeight(float32(dev.safeLeft+dev.safeRight)/scale, float32(dev.safeTop+dev.safeBottom)/scale), size)
	})

	t.Run("when canvas got hovering window head", func(t *testing.T) {
		c.(*mobileCanvas).Resize(canvasSize)
		hoveringMenu := fynecanvas.NewRectangle(color.Black)
		hoveringMenu.SetMinSize(fyne.NewSize(17, 17))
		windowHead := container.NewHBox(hoveringMenu) // a single object in window head is considered to be the hovering menu
		c.(*mobileCanvas).setWindowHead(windowHead)
		pos, size := c.InteractiveArea()
		assert.Equal(t, fyne.NewPos(float32(dev.safeLeft)/scale, float32(dev.safeTop)/scale), pos)
		assert.Equal(t, canvasSize.SubtractWidthHeight(float32(dev.safeLeft+dev.safeRight)/scale, float32(dev.safeTop+dev.safeBottom)/scale), size)
	})

	t.Run("when canvas got displacing window head", func(t *testing.T) {
		c.(*mobileCanvas).Resize(canvasSize)
		menu := fynecanvas.NewRectangle(color.Black)
		menu.SetMinSize(fyne.NewSize(17, 17))
		title := fynecanvas.NewRectangle(color.Black)
		title.SetMinSize(fyne.NewSize(42, 20))
		expectedOffset := 20 + 2*theme.Padding()
		windowHead := container.NewHBox(menu, title) // two or more objects in window head are considered to be the window title bar
		c.(*mobileCanvas).setWindowHead(windowHead)
		pos, size := c.InteractiveArea()
		assert.Equal(t, fyne.NewPos(float32(dev.safeLeft)/scale, float32(dev.safeTop)/scale+expectedOffset), pos)
		assert.Equal(t, canvasSize.SubtractWidthHeight(float32(dev.safeLeft+dev.safeRight)/scale, float32(dev.safeTop+dev.safeBottom)/scale+expectedOffset), size)
	})
}

func Test_canvas_PixelCoordinateAtPosition(t *testing.T) {
	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)

	pos := fyne.NewPos(4, 4)
	c.scale = 2.5
	x, y := c.PixelCoordinateForPosition(pos)
	assert.Equal(t, 10, x)
	assert.Equal(t, 10, y)
}

func Test_canvas_ResizeWithModalPopUpOverlay(t *testing.T) {
	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)

	c.SetContent(widget.NewLabel(widget.LabelWithStaticText("Content")))

	popup := widget.NewModalPopUp(widget.NewLabel(widget.LabelWithStaticText("PopUp")), c)
	popupBgSize := fyne.NewSize(200, 200)
	popup.Show()
	popup.Resize(popupBgSize)

	canvasSize := fyne.NewSize(600, 700)
	c.Resize(canvasSize)

	// get popup content padding dynamically
	popupContentPadding := popup.MinSize().Subtract(popup.Content.MinSize())

	assert.Equal(t, popupBgSize.Subtract(popupContentPadding), popup.Content.Size())
	assert.Equal(t, canvasSize, popup.Size())
}

func Test_canvas_Tappable(t *testing.T) {
	content := &touchableLabel{Label: widget.NewLabel(widget.LabelWithStaticText("Hi\nHi\nHi"))}
	content.ExtendBaseWidget(content)
	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)
	c.SetContent(content)
	c.Resize(fyne.NewSize(36, 24))
	content.Resize(fyne.NewSize(24, 24))

	c.tapDown(fyne.NewPos(15, 15), 0)
	assert.True(t, content.down)

	c.tapUp(fyne.NewPos(15, 15), 0, func(wid Tappable, ev *fyne.PointEvent) {
	}, func(wid SecondaryTappable, ev *fyne.PointEvent) {
	}, func(wid DoubleTappable, ev *fyne.PointEvent) {
	}, func(wid Draggable) {
	})
	assert.True(t, content.up)

	c.tapDown(fyne.NewPos(15, 15), 0)
	c.tapMove(fyne.NewPos(35, 15), 0, func(wid Draggable, ev *fyne.DragEvent) {
		wid.Dragged(ev)
	})
	assert.True(t, content.cancel)
}

func Test_canvas_Tapped(t *testing.T) {
	tapped := false
	altTapped := false
	buttonTap := false
	var pointEvent *fyne.PointEvent
	var tappedObj Tappable
	button := widget.NewButton(
		widget.ButtonWithLabel("Test"),
		widget.ButtonWithCallback(
			func() {
				buttonTap = true
			},
		),
	)
	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)
	c.SetContent(button)
	c.Resize(fyne.NewSize(36, 24))
	button.Move(fyne.NewPos(3, 3))

	tapPos := fyne.NewPos(6, 6)
	c.tapDown(tapPos, 0)
	c.tapUp(tapPos, 0, func(wid Tappable, ev *fyne.PointEvent) {
		tapped = true
		tappedObj = wid
		pointEvent = ev
		wid.Tapped(ev)
	}, func(wid SecondaryTappable, ev *fyne.PointEvent) {
		altTapped = true
		wid.TappedSecondary(ev)
	}, func(wid DoubleTappable, ev *fyne.PointEvent) {
		wid.DoubleTapped(ev)
	}, func(wid Draggable) {
	})

	assert.True(t, tapped, "tap primary")
	assert.False(t, altTapped, "don't tap secondary")
	assert.True(t, buttonTap, "button should be tapped")
	assert.Equal(t, button, tappedObj)
	if assert.NotNil(t, pointEvent) {
		assert.Equal(t, fyne.NewPos(6, 6), pointEvent.AbsolutePosition)
		assert.Equal(t, fyne.NewPos(3, 3), pointEvent.Position)
	}
}

func Test_canvas_TappedAndDoubleTapped(t *testing.T) {
	tapped := 0
	but := newDoubleTappableButton()
	but.OnTapped = func() {
		tapped = 1
	}
	but.onDoubleTap = func() {
		tapped = 2
	}

	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)
	c.SetContent(container.NewStack(but))
	c.Resize(fyne.NewSize(36, 24))

	simulateTap(c)
	time.Sleep(700 * time.Millisecond)
	assert.Equal(t, 1, tapped)

	simulateTap(c)
	simulateTap(c)
	time.Sleep(700 * time.Millisecond)
	assert.Equal(t, 2, tapped)
}

func Test_canvas_TappedMulti(t *testing.T) {
	buttonTap := false
	button := widget.NewButton(
		widget.ButtonWithLabel("Test"),
		widget.ButtonWithCallback(
			func() {
				buttonTap = true
			},
		),
	)
	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)
	c.SetContent(button)
	c.Resize(fyne.NewSize(36, 24))
	button.Move(fyne.NewPos(3, 3))

	tapPos := fyne.NewPos(6, 6)
	c.tapDown(tapPos, 0)
	c.tapUp(tapPos, 1, func(wid Tappable, ev *fyne.PointEvent) { // different tapID
		wid.Tapped(ev)
	}, func(wid SecondaryTappable, ev *fyne.PointEvent) {
	}, func(wid DoubleTappable, ev *fyne.PointEvent) {
		wid.DoubleTapped(ev)
	}, func(wid Draggable) {
	})

	assert.False(t, buttonTap, "button should not be tapped")
}

func Test_canvas_TappedSecondary(t *testing.T) {
	var pointEvent *fyne.PointEvent
	var altTappedObj SecondaryTappable
	obj := &tappableLabel{}
	obj.ExtendBaseWidget(obj)
	c := newMobileCanvas(fyne.CurrentDevice()).(*mobileCanvas)
	c.SetContent(obj)
	c.Resize(fyne.NewSize(36, 24))
	obj.Move(fyne.NewPos(3, 3))

	tapPos := fyne.NewPos(6, 6)
	c.tapDown(tapPos, 0)
	time.Sleep(310 * time.Millisecond)
	c.tapUp(tapPos, 0, func(wid Tappable, ev *fyne.PointEvent) {
		obj.tap = true
		wid.Tapped(ev)
	}, func(wid SecondaryTappable, ev *fyne.PointEvent) {
		obj.altTap = true
		altTappedObj = wid
		pointEvent = ev
		wid.TappedSecondary(ev)
	}, func(wid DoubleTappable, ev *fyne.PointEvent) {
		wid.DoubleTapped(ev)
	}, func(wid Draggable) {
	})

	assert.False(t, obj.tap, "don't tap primary")
	assert.True(t, obj.altTap, "tap secondary")
	assert.Equal(t, obj, altTappedObj)
	if assert.NotNil(t, pointEvent) {
		assert.Equal(t, fyne.NewPos(6, 6), pointEvent.AbsolutePosition)
		assert.Equal(t, fyne.NewPos(3, 3), pointEvent.Position)
	}
}

type touchableLabel struct {
	*widget.Label
	down, up, cancel bool
}

func (t *touchableLabel) TouchDown(event *mobile.TouchEvent) {
	t.down = true
}

func (t *touchableLabel) TouchUp(event *mobile.TouchEvent) {
	t.up = true
}

func (t *touchableLabel) TouchCancel(event *mobile.TouchEvent) {
	t.cancel = true
}

type tappableLabel struct {
	widget.Label
	tap, altTap bool
}

func (t *tappableLabel) Tapped(_ *fyne.PointEvent) {
	t.tap = true
}

func (t *tappableLabel) TappedSecondary(_ *fyne.PointEvent) {
	t.altTap = true
}

type focusableEntry struct {
	widget.Entry
	c fyne.Canvas

	focusedTimes   int
	unfocusedTimes int
}

func newFocusableEntry(c fyne.Canvas) *focusableEntry {
	entry := &focusableEntry{c: c}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (f *focusableEntry) FocusGained() {
	f.focusedTimes++
	f.Entry.FocusGained()
}

func (f *focusableEntry) FocusLost() {
	f.unfocusedTimes++
	f.Entry.FocusLost()
}

func (f *focusableEntry) Tapped(ev *fyne.PointEvent) {
	f.c.Focus(f)
}

type doubleTappableButton struct {
	widget.Button

	onDoubleTap func()
}

func (t *doubleTappableButton) DoubleTapped(_ *fyne.PointEvent) {
	t.onDoubleTap()
}

func newDoubleTappableButton() *doubleTappableButton {
	but := &doubleTappableButton{}
	but.ExtendBaseWidget(but)

	return but
}

func simulateTap(c *mobileCanvas) {
	c.tapDown(fyne.NewPos(15, 15), 0)
	time.Sleep(50 * time.Millisecond)
	c.tapUp(fyne.NewPos(15, 15), 0, func(wid Tappable, ev *fyne.PointEvent) {
		wid.Tapped(ev)
	}, func(wid SecondaryTappable, ev *fyne.PointEvent) {
	}, func(wid DoubleTappable, ev *fyne.PointEvent) {
		wid.DoubleTapped(ev)
	}, func(wid Draggable) {
	})
}

type mobileApp struct {
	fyne.App
	driver fyne.Driver
}

func (a *mobileApp) Driver() fyne.Driver {
	return a.driver
}

func newTestMobileApp() fyne.App {
	return &mobileApp{
		App:    fyne.CurrentApp(),
		driver: NewGoMobileDriver(),
	}
}
