//go:build !no_glfw && !mobile

package glfw

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2/internal/driver"
	"fyne.io/fyne/v2/test"
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	internalTest "fyne.io/fyne/v2/internal/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/stretchr/testify/assert"
)

func TestGlCanvas_ChildMinSizeChangeAffectsAncestorsUpToRoot(t *testing.T) {
	w := createWindow("Test").(*window)
	c := w.canvas
	leftObj1 := canvas.NewRectangle(color.Black)
	leftObj1.SetMinSize(fyne.NewSize(100, 50))
	leftObj2 := canvas.NewRectangle(color.Black)
	leftObj2.SetMinSize(fyne.NewSize(100, 50))
	leftCol := container.NewVBox(leftObj1, leftObj2)
	rightObj1 := canvas.NewRectangle(color.Black)
	rightObj1.SetMinSize(fyne.NewSize(100, 50))
	rightObj2 := canvas.NewRectangle(color.Black)
	rightObj2.SetMinSize(fyne.NewSize(100, 50))
	rightCol := container.NewVBox(rightObj1, rightObj2)
	content := container.NewHBox(leftCol, rightCol)
	w.SetContent(content)
	repaintWindow(w)

	oldCanvasSize := fyne.NewSize(200+3*theme.Padding(), 100+3*theme.Padding())
	assert.Equal(t, oldCanvasSize, c.Size())

	leftObj1.SetMinSize(fyne.NewSize(110, 60))
	c.Refresh(leftObj1)
	repaintWindow(w)

	expectedCanvasSize := oldCanvasSize.Add(fyne.NewSize(10, 10))
	assert.Equal(t, expectedCanvasSize, c.Size())
}

func TestGlCanvas_ChildMinSizeChangeAffectsAncestorsUpToScroll(t *testing.T) {
	w := createWindow("Test").(*window)
	c := w.canvas
	leftObj1 := canvas.NewRectangle(color.Black)
	leftObj1.SetMinSize(fyne.NewSize(50, 50))
	leftObj2 := canvas.NewRectangle(color.Black)
	leftObj2.SetMinSize(fyne.NewSize(50, 50))
	leftCol := container.NewVBox(leftObj1, leftObj2)
	rightObj1 := canvas.NewRectangle(color.Black)
	rightObj1.SetMinSize(fyne.NewSize(50, 50))
	rightObj2 := canvas.NewRectangle(color.Black)
	rightObj2.SetMinSize(fyne.NewSize(50, 50))
	rightCol := container.NewVBox(rightObj1, rightObj2)
	rightColScroll := container.NewScroll(rightCol)
	content := container.NewHBox(leftCol, rightColScroll)
	w.SetContent(content)

	oldCanvasSize := fyne.NewSize(200+3*theme.Padding(), 100+3*theme.Padding())
	w.Resize(oldCanvasSize)
	repaintWindow(w)

	// child size change affects ancestors up to scroll
	oldCanvasSize = c.Size()
	oldRightScrollSize := rightColScroll.Size()
	oldRightColSize := rightCol.Size()
	rightObj1.SetMinSize(fyne.NewSize(50, 100))
	c.Refresh(rightObj1)
	repaintWindow(w)

	assert.Equal(t, oldCanvasSize, c.Size())
	assert.Equal(t, oldRightScrollSize, rightColScroll.Size())
	expectedRightColSize := oldRightColSize.Add(fyne.NewSize(0, 50))
	assert.Equal(t, expectedRightColSize, rightCol.Size())
}

func TestGlCanvas_ChildMinSizeChangesInDifferentScrollAffectAncestorsUpToScroll(t *testing.T) {
	w := createWindow("Test").(*window)
	c := w.canvas
	leftObj1 := canvas.NewRectangle(color.Black)
	leftObj1.SetMinSize(fyne.NewSize(50, 50))
	leftObj2 := canvas.NewRectangle(color.Black)
	leftObj2.SetMinSize(fyne.NewSize(50, 50))
	leftCol := container.NewVBox(leftObj1, leftObj2)
	leftColScroll := container.NewScroll(leftCol)
	rightObj1 := canvas.NewRectangle(color.Black)
	rightObj1.SetMinSize(fyne.NewSize(50, 50))
	rightObj2 := canvas.NewRectangle(color.Black)
	rightObj2.SetMinSize(fyne.NewSize(50, 50))
	rightCol := container.NewVBox(rightObj1, rightObj2)
	rightColScroll := container.NewScroll(rightCol)
	content := container.NewHBox(leftColScroll, rightColScroll)
	w.SetContent(content)

	oldCanvasSize := fyne.NewSize(
		2*leftColScroll.MinSize().Width+3*theme.Padding(),
		leftColScroll.MinSize().Height+2*theme.Padding(),
	)
	w.Resize(oldCanvasSize)
	repaintWindow(w)

	oldLeftColSize := leftCol.Size()
	oldLeftScrollSize := leftColScroll.Size()
	oldRightColSize := rightCol.Size()
	oldRightScrollSize := rightColScroll.Size()
	leftObj2.SetMinSize(fyne.NewSize(50, 100))
	rightObj2.SetMinSize(fyne.NewSize(50, 200))
	c.Refresh(leftObj2)
	c.Refresh(rightObj2)
	repaintWindow(w)

	assert.Equal(t, oldCanvasSize, c.Size())
	assert.Equal(t, oldLeftScrollSize, leftColScroll.Size())
	assert.Equal(t, oldRightScrollSize, rightColScroll.Size())
	expectedLeftColSize := oldLeftColSize.Add(fyne.NewSize(0, 50))
	assert.Equal(t, expectedLeftColSize, leftCol.Size())
	expectedRightColSize := oldRightColSize.Add(fyne.NewSize(0, 150))
	assert.Equal(t, expectedRightColSize, rightCol.Size())
}

func TestGlCanvas_Content(t *testing.T) {
	content := &canvas.Circle{}
	w := createWindow("Test")
	w.SetContent(content)

	assert.Equal(t, content, w.Content())
}

func TestGlCanvas_ContentChangeWithoutMinSizeChangeDoesNotLayout(t *testing.T) {
	w := createWindow("Test").(*window)
	c := w.canvas
	leftObj1 := canvas.NewRectangle(color.Black)
	leftObj1.SetMinSize(fyne.NewSize(50, 50))
	leftObj2 := canvas.NewRectangle(color.Black)
	leftObj2.SetMinSize(fyne.NewSize(50, 50))
	leftCol := container.NewVBox(leftObj1, leftObj2)
	rightObj1 := canvas.NewRectangle(color.Black)
	rightObj1.SetMinSize(fyne.NewSize(50, 50))
	rightObj2 := canvas.NewRectangle(color.Black)
	rightObj2.SetMinSize(fyne.NewSize(50, 50))
	rightCol := container.NewVBox(rightObj1, rightObj2)
	content := container.NewWithoutLayout(leftCol, rightCol)
	layout := &recordingLayout{}
	content.Layout = layout
	w.SetContent(content)

	repaintWindow(w)
	// clear the recorded layouts
	for layout.popLayoutEvent() != nil {
	}
	assert.Nil(t, layout.popLayoutEvent())

	leftObj1.FillColor = color.White
	rightObj1.FillColor = color.White
	rightObj2.FillColor = color.White
	c.Refresh(leftObj1)
	c.Refresh(rightObj1)
	c.Refresh(rightObj2)

	assert.Nil(t, layout.popLayoutEvent())
}

func TestGlCanvas_Focus(t *testing.T) {
	w := createWindow("Test")
	w.SetPadded(false)
	c := w.Canvas().(*glCanvas)

	ce := &focusable{id: "ce1"}
	content := container.NewVBox(ce)
	me := &focusable{id: "o2e1"}
	menuOverlay := container.NewVBox(me)
	o1e := &focusable{id: "o1e1"}
	overlay1 := container.NewVBox(o1e)
	o2e := &focusable{id: "o2e1"}
	overlay2 := container.NewVBox(o2e)
	w.SetContent(content)
	c.setMenuOverlay(menuOverlay)
	c.Overlays().Add(overlay1)
	c.Overlays().Add(overlay2)

	c.Focus(ce)
	assert.True(t, ce.focused, "focuses content object even if content is not in focus")

	c.Focus(me)
	assert.True(t, me.focused, "focuses menu object even if menu is not in focus")
	assert.True(t, ce.focused, "does not affect focus on other layer")

	c.Focus(o1e)
	assert.True(t, o1e.focused, "focuses overlay object even if menu is not in focus")
	assert.True(t, me.focused, "does not affect focus on other layer")

	c.Focus(o2e)
	assert.True(t, o2e.focused)
	assert.True(t, o1e.focused, "does not affect focus on other layer")

	foreign := &focusable{id: "o2e1"}
	c.Focus(foreign)
	assert.False(t, foreign.focused, "does not focus foreign object")
	assert.True(t, o2e.focused)
}

func TestGlCanvas_Focus_BeforeVisible(t *testing.T) {
	w := createWindow("Test")
	w.SetPadded(false)
	e := widget.NewEntry()
	c := w.Canvas().(*glCanvas)
	c.Focus(e) // this crashed in the past
}

func TestGlCanvas_Focus_SetContent(t *testing.T) {
	w := createWindow("Test")
	w.SetPadded(false)
	e := widget.NewEntry()
	w.SetContent(container.NewHBox(e))
	c := w.Canvas().(*glCanvas)
	c.Focus(e)
	assert.Equal(t, e, c.Focused())

	w.SetContent(container.NewVBox(e))
	assert.Equal(t, e, c.Focused())
}

func TestGlCanvas_FocusHandlingWhenAddingAndRemovingOverlays(t *testing.T) {
	w := createWindow("Test")
	w.SetPadded(false)
	c := w.Canvas().(*glCanvas)

	ce1 := &focusable{id: "ce1"}
	ce2 := &focusable{id: "ce2"}
	content := container.NewVBox(ce1, ce2)
	o1e1 := &focusable{id: "o1e1"}
	o1e2 := &focusable{id: "o1e2"}
	overlay1 := container.NewVBox(o1e1, o1e2)
	o2e1 := &focusable{id: "o2e1"}
	o2e2 := &focusable{id: "o2e2"}
	overlay2 := container.NewVBox(o2e1, o2e2)
	w.SetContent(content)

	assert.Nil(t, c.Focused())

	c.FocusPrevious()
	assert.Equal(t, ce2, c.Focused())
	assert.True(t, ce2.focused)

	c.Overlays().Add(overlay1)
	ctxt := "adding overlay changes focus handler but does not remove focus from content"
	assert.Nil(t, c.Focused(), ctxt)
	assert.True(t, ce2.focused, ctxt)

	c.FocusNext()
	ctxt = "changing focus affects overlay instead of content"
	assert.Equal(t, o1e1, c.Focused(), ctxt)
	assert.False(t, ce1.focused, ctxt)
	assert.True(t, ce2.focused, ctxt)
	assert.True(t, o1e1.focused, ctxt)

	c.Overlays().Add(overlay2)
	ctxt = "adding overlay changes focus handler but does not remove focus from previous overlay"
	assert.Nil(t, c.Focused(), ctxt)
	assert.True(t, o1e1.focused, ctxt)

	c.FocusPrevious()
	ctxt = "changing focus affects top overlay only"
	assert.Equal(t, o2e2, c.Focused(), ctxt)
	assert.True(t, o1e1.focused, ctxt)
	assert.False(t, o1e2.focused, ctxt)
	assert.True(t, o2e2.focused, ctxt)

	c.FocusNext()
	assert.Equal(t, o2e1, c.Focused())
	assert.False(t, o2e2.focused)
	assert.True(t, o2e1.focused)

	c.Overlays().Remove(overlay2)
	ctxt = "removing overlay restores focus handler from previous overlay but does not remove focus from removed overlay"
	assert.Equal(t, o1e1, c.Focused(), ctxt)
	assert.True(t, o2e1.focused, ctxt)
	assert.False(t, o2e2.focused, ctxt)
	assert.True(t, o1e1.focused, ctxt)

	c.FocusPrevious()
	assert.Equal(t, o1e2, c.Focused())
	assert.False(t, o1e1.focused)
	assert.True(t, o1e2.focused)

	c.Overlays().Remove(overlay1)
	ctxt = "removing last overlay restores focus handler from content but does not remove focus from removed overlay"
	assert.Equal(t, ce2, c.Focused(), ctxt)
	assert.False(t, o1e1.focused, ctxt)
	assert.True(t, o1e2.focused, ctxt)
	assert.True(t, ce2.focused, ctxt)
}

func TestGlCanvas_InsufficientSizeDoesntTriggerResizeIfSizeIsAlreadyMaxedOut(t *testing.T) {
	w := createWindow("Test").(*window)
	c := w.Canvas().(*glCanvas)
	canvasSize := fyne.NewSize(200, 100)
	w.Resize(canvasSize)
	ensureCanvasSize(t, w, canvasSize)
	popUpContent := canvas.NewRectangle(color.Black)
	popUpContent.SetMinSize(fyne.NewSize(1000, 10))
	popUp := widget.NewPopUp(popUpContent, c)

	// This is because of a bug in PopUp size handling that will be fixed later.
	// This line will vanish then.
	popUp.Resize(popUpContent.MinSize().Add(fyne.NewSize(theme.Padding()*2, theme.Padding()*2)))

	assert.Equal(t, fyne.NewSize(1000, 10), popUpContent.Size())
	assert.Equal(t, fyne.NewSize(1000, 10).Add(fyne.NewSize(theme.Padding()*2, theme.Padding()*2)), popUp.MinSize())
	assert.Equal(t, canvasSize, popUp.Size())

	repaintWindow(w)

	assert.Equal(t, fyne.NewSize(1000, 10), popUpContent.Size())
	assert.Equal(t, canvasSize, popUp.Size())
}

func TestGlCanvas_MinSizeShrinkTriggersLayout(t *testing.T) {
	w := createWindow("Test").(*window)
	c := w.Canvas().(*glCanvas)
	leftObj1 := canvas.NewRectangle(color.Black)
	leftObj1.SetMinSize(fyne.NewSize(100, 50))
	leftObj2 := canvas.NewRectangle(color.Black)
	leftObj2.SetMinSize(fyne.NewSize(100, 50))
	leftCol := container.NewVBox(leftObj1, leftObj2)
	rightObj1 := canvas.NewRectangle(color.Black)
	rightObj1.SetMinSize(fyne.NewSize(100, 50))
	rightObj2 := canvas.NewRectangle(color.Black)
	rightObj2.SetMinSize(fyne.NewSize(100, 50))
	rightCol := container.NewVBox(rightObj1, rightObj2)
	content := container.NewHBox(leftCol, rightCol)
	w.SetContent(content)

	oldCanvasSize := fyne.NewSize(200+3*theme.Padding(), 100+3*theme.Padding())
	assert.Equal(t, oldCanvasSize, c.Size())
	repaintWindow(w)

	oldRightColSize := rightCol.Size()
	leftObj1.SetMinSize(fyne.NewSize(90, 40))
	rightObj1.SetMinSize(fyne.NewSize(80, 30))
	rightObj2.SetMinSize(fyne.NewSize(80, 20))
	c.Refresh(leftObj1)
	c.Refresh(rightObj1)
	c.Refresh(rightObj2)
	repaintWindow(w)

	assert.Equal(t, oldCanvasSize, c.Size())
	expectedRightColSize := oldRightColSize.Subtract(fyne.NewSize(20, 0))
	assert.Equal(t, expectedRightColSize, rightCol.Size())
	assert.Equal(t, fyne.NewSize(100, 40), leftObj1.Size())
	assert.Equal(t, fyne.NewSize(80, 30), rightObj1.Size())
	assert.Equal(t, fyne.NewSize(80, 20), rightObj2.Size())
}

func TestGlCanvas_NilContent(t *testing.T) {
	w := createWindow("Test")

	assert.NotNil(t, w.Content()) // never a nil canvas so we have a sensible fallback
}

func TestGlCanvas_PixelCoordinateAtPosition(t *testing.T) {
	w := createWindow("Test").(*window)
	c := w.Canvas().(*glCanvas)

	pos := fyne.NewPos(4, 4)
	c.Mu.Lock()
	c.scale = 2.5
	c.Mu.Unlock()
	x, y := c.PixelCoordinateForPosition(pos)
	assert.Equal(t, int(10*c.texScale), x)
	assert.Equal(t, int(10*c.texScale), y)

	c.Mu.Lock()
	c.texScale = 2.0
	c.Mu.Unlock()
	x, y = c.PixelCoordinateForPosition(pos)
	assert.Equal(t, 20, x)
	assert.Equal(t, 20, y)
}

func TestGlCanvas_Resize(t *testing.T) {
	w := createWindow("Test")
	w.SetPadded(false)

	content := widget.NewLabel(widget.LabelWithStaticText("Content"))
	w.SetContent(content)
	ensureCanvasSize(t, w.(*window), fyne.NewSize(69, 36))

	size := fyne.NewSize(200, 100)
	assert.NotEqual(t, size, content.Size())

	w.Resize(size)
	ensureCanvasSize(t, w.(*window), size)
	assert.Equal(t, size, content.Size())
}

// TODO: this can be removed when #707 is addressed
func TestGlCanvas_ResizeWithOtherOverlay(t *testing.T) {
	w := createWindow("Test")
	w.SetPadded(false)

	content := widget.NewLabel(widget.LabelWithStaticText("Content"))
	over := widget.NewLabel(widget.LabelWithStaticText("Over"))
	w.SetContent(content)
	w.Canvas().Overlays().Add(over)
	ensureCanvasSize(t, w.(*window), fyne.NewSize(69, 36))
	// TODO: address #707; overlays should always be canvas size
	over.Resize(w.Canvas().Size())

	size := fyne.NewSize(200, 100)
	assert.NotEqual(t, size, content.Size())
	assert.NotEqual(t, size, over.Size())

	w.Resize(size)
	ensureCanvasSize(t, w.(*window), size)
	assert.Equal(t, size, content.Size(), "canvas content is resized")
	assert.Equal(t, size, over.Size(), "canvas overlay is resized")
}

func TestGlCanvas_ResizeWithOverlays(t *testing.T) {
	w := createWindow("Test")
	w.SetPadded(false)

	content := widget.NewLabel(widget.LabelWithStaticText("Content"))
	o1 := widget.NewLabel(widget.LabelWithStaticText("o1"))
	o2 := widget.NewLabel(widget.LabelWithStaticText("o2"))
	o3 := widget.NewLabel(widget.LabelWithStaticText("o3"))
	w.SetContent(content)
	w.Canvas().Overlays().Add(o1)
	w.Canvas().Overlays().Add(o2)
	w.Canvas().Overlays().Add(o3)
	ensureCanvasSize(t, w.(*window), fyne.NewSize(69, 36))

	size := fyne.NewSize(200, 100)
	assert.NotEqual(t, size, content.Size())
	assert.NotEqual(t, size, o1.Size())
	assert.NotEqual(t, size, o2.Size())
	assert.NotEqual(t, size, o3.Size())

	w.Resize(size)
	ensureCanvasSize(t, w.(*window), size)
	assert.Equal(t, size, content.Size(), "canvas content is resized")
	assert.Equal(t, size, o1.Size(), "canvas overlay 1 is resized")
	assert.Equal(t, size, o2.Size(), "canvas overlay 2 is resized")
	assert.Equal(t, size, o3.Size(), "canvas overlay 3 is resized")
}

// TODO: this can be removed when #707 is addressed
func TestGlCanvas_ResizeWithPopUpOverlay(t *testing.T) {
	w := createWindow("Test")
	w.SetPadded(false)

	content := widget.NewLabel(widget.LabelWithStaticText("Content"))
	over := widget.NewPopUp(widget.NewLabel(widget.LabelWithStaticText("Over")), w.Canvas())
	w.SetContent(content)
	over.Show()
	ensureCanvasSize(t, w.(*window), fyne.NewSize(69, 36))

	size := fyne.NewSize(200, 100)
	overContentSize := over.Content.Size()
	assert.NotZero(t, overContentSize)
	assert.NotEqual(t, size, content.Size())
	assert.NotEqual(t, size, over.Size())
	assert.NotEqual(t, size, overContentSize)

	w.Resize(size)
	ensureCanvasSize(t, w.(*window), size)
	assert.Equal(t, size, content.Size(), "canvas content is resized")
	assert.Equal(t, size, over.Size(), "canvas overlay is resized")
	assert.Equal(t, overContentSize, over.Content.Size(), "canvas overlay content is _not_ resized")
}

func TestGlCanvas_ResizeWithModalPopUpOverlay(t *testing.T) {
	w := createWindow("Test")
	w.SetPadded(false)

	content := widget.NewLabel(widget.LabelWithStaticText("Content"))
	w.SetContent(content)

	popup := widget.NewModalPopUp(widget.NewLabel(widget.LabelWithStaticText("PopUp")), w.Canvas())
	popupBgSize := fyne.NewSize(975, 575)
	popup.Show()
	popup.Resize(popupBgSize)
	ensureCanvasSize(t, w.(*window), fyne.NewSize(69, 36))

	winSize := fyne.NewSize(1000, 600)
	w.Resize(winSize)
	ensureCanvasSize(t, w.(*window), winSize)

	// get popup content padding dynamically
	popupContentPadding := popup.MinSize().Subtract(popup.Content.MinSize())

	assert.Equal(t, popupBgSize.Subtract(popupContentPadding), popup.Content.Size())
	assert.Equal(t, winSize, popup.Size())
}

func TestGlCanvas_Scale(t *testing.T) {
	w := createWindow("Test").(*window)
	c := w.Canvas().(*glCanvas)

	c.Mu.Lock()
	c.scale = 2.5
	c.Mu.Unlock()
	assert.Equal(t, 5, int(2*c.Scale()))
}

func TestGlCanvas_SetContent(t *testing.T) {
	fyne.CurrentApp().Settings().SetTheme(internalTest.DarkTheme(theme.DefaultTheme()))
	var menuHeight float32
	if hasNativeMenu() {
		menuHeight = 0
	} else {
		menuHeight = NewMenuBar(fyne.NewMainMenu(fyne.NewMenu("Test", fyne.NewMenuItem("Empty", func() {}))), nil).MinSize().Height
	}
	tests := []struct {
		name               string
		padding            bool
		menu               bool
		expectedPad        float32
		expectedMenuHeight float32
	}{
		{"window without padding", false, false, 0, 0},
		{"window with padding", true, false, theme.Padding(), 0},
		{"window with menu without padding", false, true, 0, menuHeight},
		{"window with menu and padding", true, true, theme.Padding(), menuHeight},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := createWindow("Test").(*window)
			w.SetPadded(tt.padding)
			if tt.menu {
				w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("Test", fyne.NewMenuItem("Test", func() {}))))
			}
			ensureCanvasSize(t, w, fyne.NewSize(69, 37))
			content := canvas.NewCircle(color.Black)
			canvasSize := float32(200)
			w.SetContent(content)
			w.Resize(fyne.NewSize(canvasSize, canvasSize))
			ensureCanvasSize(t, w, fyne.NewSize(canvasSize, canvasSize))

			newContent := canvas.NewCircle(color.White)
			assert.Equal(t, fyne.NewPos(0, 0), newContent.Position())
			assert.Equal(t, fyne.NewSize(0, 0), newContent.Size())
			w.SetContent(newContent)
			assert.Equal(t, fyne.NewPos(tt.expectedPad, tt.expectedPad+tt.expectedMenuHeight), newContent.Position())
			assert.Equal(t, fyne.NewSize(canvasSize-2*tt.expectedPad, canvasSize-2*tt.expectedPad-tt.expectedMenuHeight), newContent.Size())
		})
	}
}

type recordingLayout struct {
	layoutEvents []any
}

func (l *recordingLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.layoutEvents = append(l.layoutEvents, size)
}

func (l *recordingLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(6, 9)
}

func (l *recordingLayout) popLayoutEvent() (e any) {
	e, l.layoutEvents = pop(l.layoutEvents)
	return
}

func TestCanvas_walkTree(t *testing.T) {
	test.NewTempApp(t)

	leftObj1 := canvas.NewRectangle(color.Gray16{Y: 1})
	leftObj2 := canvas.NewRectangle(color.Gray16{Y: 2})
	leftCol := container.NewWithoutLayout(leftObj1, leftObj2)
	rightObj1 := canvas.NewRectangle(color.Gray16{Y: 10})
	rightObj2 := canvas.NewRectangle(color.Gray16{Y: 20})
	rightCol := container.NewWithoutLayout(rightObj1, rightObj2)
	content := container.NewWithoutLayout(leftCol, rightCol)
	content.Move(fyne.NewPos(17, 42))
	leftCol.Move(fyne.NewPos(300, 400))
	leftObj1.Move(fyne.NewPos(1, 2))
	leftObj2.Move(fyne.NewPos(20, 30))
	rightObj1.Move(fyne.NewPos(500, 600))
	rightObj2.Move(fyne.NewPos(60, 70))
	rightCol.Move(fyne.NewPos(7, 8))

	tree := &renderCacheTree{root: &driver.RenderCacheNode{Obj: content}}
	c := &glCanvas{}
	c.Initialize(nil, func() {})
	c.SetContentTreeAndFocusMgr(&canvas.Rectangle{FillColor: theme.Color(theme.ColorNameBackground)})

	type nodeInfo struct {
		obj                                     fyne.CanvasObject
		lastBeforeCallIndex, lastAfterCallIndex int
	}
	updateInfoBefore := func(node *driver.RenderCacheNode, index int) {
		pd, _ := node.PainterData.(nodeInfo)
		if (pd != nodeInfo{}) && pd.obj != node.Obj {
			panic("node cache does not match node obj - nodes should not be reused for different objects")
		}
		pd.obj = node.Obj
		pd.lastBeforeCallIndex = index
		node.PainterData = pd
	}
	updateInfoAfter := func(node *driver.RenderCacheNode, index int) {
		pd := node.PainterData.(nodeInfo)
		if pd.obj != node.Obj {
			panic("node cache does not match node obj - nodes should not be reused for different objects")
		}
		pd.lastAfterCallIndex = index
		node.PainterData = pd
	}

	//
	// test that first walk calls the hooks correctly
	//
	type beforeCall struct {
		obj    fyne.CanvasObject
		parent fyne.CanvasObject
		pos    fyne.Position
	}
	var beforeCalls []beforeCall
	type afterCall struct {
		obj    fyne.CanvasObject
		parent fyne.CanvasObject
	}
	var afterCalls []afterCall

	var i int
	c.walkTree(tree, func(node *driver.RenderCacheNode, pos fyne.Position) {
		var parent fyne.CanvasObject
		if node.Parent != nil {
			parent = node.Parent.Obj
		}
		i++
		updateInfoBefore(node, i)
		beforeCalls = append(beforeCalls, beforeCall{obj: node.Obj, parent: parent, pos: pos})
	}, func(node *driver.RenderCacheNode, _ fyne.Position) {
		var parent fyne.CanvasObject
		if node.Parent != nil {
			parent = node.Parent.Obj
		}
		i++
		updateInfoAfter(node, i)
		node.MinSize.Height = node.Obj.Position().Y
		afterCalls = append(afterCalls, afterCall{obj: node.Obj, parent: parent})
	})

	assert.Equal(t, []beforeCall{
		{obj: content, pos: fyne.NewPos(17, 42)},
		{obj: leftCol, parent: content, pos: fyne.NewPos(317, 442)},
		{obj: leftObj1, parent: leftCol, pos: fyne.NewPos(318, 444)},
		{obj: leftObj2, parent: leftCol, pos: fyne.NewPos(337, 472)},
		{obj: rightCol, parent: content, pos: fyne.NewPos(24, 50)},
		{obj: rightObj1, parent: rightCol, pos: fyne.NewPos(524, 650)},
		{obj: rightObj2, parent: rightCol, pos: fyne.NewPos(84, 120)},
	}, beforeCalls, "calls before children hook with the correct node and position")
	assert.Equal(t, []afterCall{
		{obj: leftObj1, parent: leftCol},
		{obj: leftObj2, parent: leftCol},
		{obj: leftCol, parent: content},
		{obj: rightObj1, parent: rightCol},
		{obj: rightObj2, parent: rightCol},
		{obj: rightCol, parent: content},
		{obj: content},
	}, afterCalls, "calls after children hook with the correct node")

	//
	// test that second walk gives access to the cache
	//
	var secondRunBeforePainterData []nodeInfo
	var secondRunAfterPainterData []nodeInfo
	var nodes []*driver.RenderCacheNode

	c.walkTree(tree, func(node *driver.RenderCacheNode, pos fyne.Position) {
		secondRunBeforePainterData = append(secondRunBeforePainterData, node.PainterData.(nodeInfo))
		nodes = append(nodes, node)
	}, func(node *driver.RenderCacheNode, _ fyne.Position) {
		secondRunAfterPainterData = append(secondRunAfterPainterData, node.PainterData.(nodeInfo))
	})

	assert.Equal(t, []nodeInfo{
		{obj: content, lastBeforeCallIndex: 1, lastAfterCallIndex: 14},
		{obj: leftCol, lastBeforeCallIndex: 2, lastAfterCallIndex: 7},
		{obj: leftObj1, lastBeforeCallIndex: 3, lastAfterCallIndex: 4},
		{obj: leftObj2, lastBeforeCallIndex: 5, lastAfterCallIndex: 6},
		{obj: rightCol, lastBeforeCallIndex: 8, lastAfterCallIndex: 13},
		{obj: rightObj1, lastBeforeCallIndex: 9, lastAfterCallIndex: 10},
		{obj: rightObj2, lastBeforeCallIndex: 11, lastAfterCallIndex: 12},
	}, secondRunBeforePainterData, "second run uses cached nodes")
	assert.Equal(t, []nodeInfo{
		{obj: leftObj1, lastBeforeCallIndex: 3, lastAfterCallIndex: 4},
		{obj: leftObj2, lastBeforeCallIndex: 5, lastAfterCallIndex: 6},
		{obj: leftCol, lastBeforeCallIndex: 2, lastAfterCallIndex: 7},
		{obj: rightObj1, lastBeforeCallIndex: 9, lastAfterCallIndex: 10},
		{obj: rightObj2, lastBeforeCallIndex: 11, lastAfterCallIndex: 12},
		{obj: rightCol, lastBeforeCallIndex: 8, lastAfterCallIndex: 13},
		{obj: content, lastBeforeCallIndex: 1, lastAfterCallIndex: 14},
	}, secondRunAfterPainterData, "second run uses cached nodes")
	leftObj1Node := nodes[2]
	leftObj2Node := nodes[3]
	assert.Equal(t, leftObj2Node, leftObj1Node.NextSibling, "correct sibling relation")
	assert.Nil(t, leftObj2Node.NextSibling, "no surplus nodes")
	rightObj1Node := nodes[5]
	rightObj2Node := nodes[6]
	assert.Equal(t, rightObj2Node, rightObj1Node.NextSibling, "correct sibling relation")
	rightColNode := nodes[4]
	assert.Nil(t, rightColNode.NextSibling, "no surplus nodes")

	//
	// test that removal, replacement and adding at the end of a children list works
	//
	deleteAt(leftCol, 1)
	leftNewObj2 := canvas.NewRectangle(color.Gray16{Y: 3})
	leftCol.Add(leftNewObj2)
	deleteAt(rightCol, 1)
	thirdCol := container.NewVBox()
	content.Add(thirdCol)
	var thirdRunBeforePainterData []nodeInfo
	var thirdRunAfterPainterData []nodeInfo

	i = 0
	c.walkTree(tree, func(node *driver.RenderCacheNode, pos fyne.Position) {
		i++
		updateInfoBefore(node, i)
		thirdRunBeforePainterData = append(thirdRunBeforePainterData, node.PainterData.(nodeInfo))
	}, func(node *driver.RenderCacheNode, _ fyne.Position) {
		i++
		updateInfoAfter(node, i)
		thirdRunAfterPainterData = append(thirdRunAfterPainterData, node.PainterData.(nodeInfo))
	})

	assert.Equal(t, []nodeInfo{
		{obj: content, lastBeforeCallIndex: 1, lastAfterCallIndex: 14},
		{obj: leftCol, lastBeforeCallIndex: 2, lastAfterCallIndex: 7},
		{obj: leftObj1, lastBeforeCallIndex: 3, lastAfterCallIndex: 4},
		{obj: leftNewObj2, lastBeforeCallIndex: 5, lastAfterCallIndex: 0}, // new node for replaced obj
		{obj: rightCol, lastBeforeCallIndex: 8, lastAfterCallIndex: 13},
		{obj: rightObj1, lastBeforeCallIndex: 9, lastAfterCallIndex: 10},
		{obj: thirdCol, lastBeforeCallIndex: 12, lastAfterCallIndex: 0}, // new node for third column
	}, thirdRunBeforePainterData, "third run uses cached nodes if possible")
	assert.Equal(t, []nodeInfo{
		{obj: leftObj1, lastBeforeCallIndex: 3, lastAfterCallIndex: 4},
		{obj: leftNewObj2, lastBeforeCallIndex: 5, lastAfterCallIndex: 6}, // new node for replaced obj
		{obj: leftCol, lastBeforeCallIndex: 2, lastAfterCallIndex: 7},
		{obj: rightObj1, lastBeforeCallIndex: 9, lastAfterCallIndex: 10},
		{obj: rightCol, lastBeforeCallIndex: 8, lastAfterCallIndex: 11},
		{obj: thirdCol, lastBeforeCallIndex: 12, lastAfterCallIndex: 13}, // new node for third column
		{obj: content, lastBeforeCallIndex: 1, lastAfterCallIndex: 14},
	}, thirdRunAfterPainterData, "third run uses cached nodes if possible")
	assert.NotEqual(t, leftObj2Node, leftObj1Node.NextSibling, "new node for replaced object")
	assert.Nil(t, rightObj1Node.NextSibling, "node for removed object has been removed, too")
	assert.NotNil(t, rightColNode.NextSibling, "new node for new object")

	//
	// test that insertion at the beginnning or in the middle of a children list
	// removes all following siblings and their subtrees
	//
	leftNewObj2a := canvas.NewRectangle(color.Gray16{Y: 4})
	insert(leftCol, leftNewObj2a, 1)
	rightNewObj0 := canvas.NewRectangle(color.Gray16{Y: 30})
	Prepend(rightCol, rightNewObj0)
	var fourthRunBeforePainterData []nodeInfo
	var fourthRunAfterPainterData []nodeInfo
	nodes = []*driver.RenderCacheNode{}

	i = 0
	c.walkTree(tree, func(node *driver.RenderCacheNode, pos fyne.Position) {
		i++
		updateInfoBefore(node, i)
		fourthRunBeforePainterData = append(fourthRunBeforePainterData, node.PainterData.(nodeInfo))
		nodes = append(nodes, node)
	}, func(node *driver.RenderCacheNode, _ fyne.Position) {
		i++
		updateInfoAfter(node, i)
		fourthRunAfterPainterData = append(fourthRunAfterPainterData, node.PainterData.(nodeInfo))
	})

	assert.Equal(t, []nodeInfo{
		{obj: content, lastBeforeCallIndex: 1, lastAfterCallIndex: 14},
		{obj: leftCol, lastBeforeCallIndex: 2, lastAfterCallIndex: 7},
		{obj: leftObj1, lastBeforeCallIndex: 3, lastAfterCallIndex: 4},
		{obj: leftNewObj2a, lastBeforeCallIndex: 5, lastAfterCallIndex: 0}, // new node for inserted obj
		{obj: leftNewObj2, lastBeforeCallIndex: 7, lastAfterCallIndex: 0},  // new node because of tail cut
		{obj: rightCol, lastBeforeCallIndex: 10, lastAfterCallIndex: 11},
		{obj: rightNewObj0, lastBeforeCallIndex: 11, lastAfterCallIndex: 0}, // new node for inserted obj
		{obj: rightObj1, lastBeforeCallIndex: 13, lastAfterCallIndex: 0},    // new node because of tail cut
		{obj: thirdCol, lastBeforeCallIndex: 16, lastAfterCallIndex: 13},
	}, fourthRunBeforePainterData, "fourth run uses cached nodes if possible")
	assert.Equal(t, []nodeInfo{
		{obj: leftObj1, lastBeforeCallIndex: 3, lastAfterCallIndex: 4},
		{obj: leftNewObj2a, lastBeforeCallIndex: 5, lastAfterCallIndex: 6},
		{obj: leftNewObj2, lastBeforeCallIndex: 7, lastAfterCallIndex: 8},
		{obj: leftCol, lastBeforeCallIndex: 2, lastAfterCallIndex: 9},
		{obj: rightNewObj0, lastBeforeCallIndex: 11, lastAfterCallIndex: 12},
		{obj: rightObj1, lastBeforeCallIndex: 13, lastAfterCallIndex: 14},
		{obj: rightCol, lastBeforeCallIndex: 10, lastAfterCallIndex: 15},
		{obj: thirdCol, lastBeforeCallIndex: 16, lastAfterCallIndex: 17},
		{obj: content, lastBeforeCallIndex: 1, lastAfterCallIndex: 18},
	}, fourthRunAfterPainterData, "fourth run uses cached nodes if possible")
	// check cache tree integrity
	// content node
	assert.Equal(t, content, nodes[0].Obj)
	assert.Equal(t, leftCol, nodes[0].FirstChild.Obj)
	assert.Nil(t, nodes[0].NextSibling)
	// leftCol node
	assert.Equal(t, leftCol, nodes[1].Obj)
	assert.Equal(t, leftObj1, nodes[1].FirstChild.Obj)
	assert.Equal(t, rightCol, nodes[1].NextSibling.Obj)
	// leftObj1 node
	assert.Equal(t, leftObj1, nodes[2].Obj)
	assert.Nil(t, nodes[2].FirstChild)
	assert.Equal(t, leftNewObj2a, nodes[2].NextSibling.Obj)
	// leftNewObj2a node
	assert.Equal(t, leftNewObj2a, nodes[3].Obj)
	assert.Nil(t, nodes[3].FirstChild)
	assert.Equal(t, leftNewObj2, nodes[3].NextSibling.Obj)
	// leftNewObj2 node
	assert.Equal(t, leftNewObj2, nodes[4].Obj)
	assert.Nil(t, nodes[4].FirstChild)
	assert.Nil(t, nodes[4].NextSibling)
	// rightCol node
	assert.Equal(t, rightCol, nodes[5].Obj)
	assert.Equal(t, rightNewObj0, nodes[5].FirstChild.Obj)
	assert.Equal(t, thirdCol, nodes[5].NextSibling.Obj)
	// rightNewObj0 node
	assert.Equal(t, rightNewObj0, nodes[6].Obj)
	assert.Nil(t, nodes[6].FirstChild)
	assert.Equal(t, rightObj1, nodes[6].NextSibling.Obj)
	// rightObj1 node
	assert.Equal(t, rightObj1, nodes[7].Obj)
	assert.Nil(t, nodes[7].FirstChild)
	assert.Nil(t, nodes[7].NextSibling)
	// thirdCol node
	assert.Equal(t, thirdCol, nodes[8].Obj)
	assert.Nil(t, nodes[8].FirstChild)
	assert.Nil(t, nodes[8].NextSibling)

	//
	// test that removal at the beginning or in the middle of a children list
	// removes all following siblings and their subtrees
	//
	deleteAt(leftCol, 1)
	deleteAt(rightCol, 0)
	var fifthRunBeforePainterData []nodeInfo
	var fifthRunAfterPainterData []nodeInfo
	nodes = []*driver.RenderCacheNode{}

	i = 0
	c.walkTree(tree, func(node *driver.RenderCacheNode, pos fyne.Position) {
		i++
		updateInfoBefore(node, i)
		fifthRunBeforePainterData = append(fifthRunBeforePainterData, node.PainterData.(nodeInfo))
		nodes = append(nodes, node)
	}, func(node *driver.RenderCacheNode, _ fyne.Position) {
		i++
		updateInfoAfter(node, i)
		fifthRunAfterPainterData = append(fifthRunAfterPainterData, node.PainterData.(nodeInfo))
	})

	assert.Equal(t, []nodeInfo{
		{obj: content, lastBeforeCallIndex: 1, lastAfterCallIndex: 18},
		{obj: leftCol, lastBeforeCallIndex: 2, lastAfterCallIndex: 9},
		{obj: leftObj1, lastBeforeCallIndex: 3, lastAfterCallIndex: 4},
		{obj: leftNewObj2, lastBeforeCallIndex: 5, lastAfterCallIndex: 0}, // new node because of tail cut
		{obj: rightCol, lastBeforeCallIndex: 8, lastAfterCallIndex: 15},
		{obj: rightObj1, lastBeforeCallIndex: 9, lastAfterCallIndex: 0}, // new node because of tail cut
		{obj: thirdCol, lastBeforeCallIndex: 12, lastAfterCallIndex: 17},
	}, fifthRunBeforePainterData, "fifth run uses cached nodes if possible")
	assert.Equal(t, []nodeInfo{
		{obj: leftObj1, lastBeforeCallIndex: 3, lastAfterCallIndex: 4},
		{obj: leftNewObj2, lastBeforeCallIndex: 5, lastAfterCallIndex: 6},
		{obj: leftCol, lastBeforeCallIndex: 2, lastAfterCallIndex: 7},
		{obj: rightObj1, lastBeforeCallIndex: 9, lastAfterCallIndex: 10},
		{obj: rightCol, lastBeforeCallIndex: 8, lastAfterCallIndex: 11},
		{obj: thirdCol, lastBeforeCallIndex: 12, lastAfterCallIndex: 13},
		{obj: content, lastBeforeCallIndex: 1, lastAfterCallIndex: 14},
	}, fifthRunAfterPainterData, "fifth run uses cached nodes if possible")
	// check cache tree integrity
	// content node
	assert.Equal(t, content, nodes[0].Obj)
	assert.Equal(t, leftCol, nodes[0].FirstChild.Obj)
	assert.Nil(t, nodes[0].NextSibling)
	// leftCol node
	assert.Equal(t, leftCol, nodes[1].Obj)
	assert.Equal(t, leftObj1, nodes[1].FirstChild.Obj)
	assert.Equal(t, rightCol, nodes[1].NextSibling.Obj)
	// leftObj1 node
	assert.Equal(t, leftObj1, nodes[2].Obj)
	assert.Nil(t, nodes[2].FirstChild)
	assert.Equal(t, leftNewObj2, nodes[2].NextSibling.Obj)
	// leftNewObj2 node
	assert.Equal(t, leftNewObj2, nodes[3].Obj)
	assert.Nil(t, nodes[3].FirstChild)
	assert.Nil(t, nodes[3].NextSibling)
	// rightCol node
	assert.Equal(t, rightCol, nodes[4].Obj)
	assert.Equal(t, rightObj1, nodes[4].FirstChild.Obj)
	assert.Equal(t, thirdCol, nodes[4].NextSibling.Obj)
	// rightObj1 node
	assert.Equal(t, rightObj1, nodes[5].Obj)
	assert.Nil(t, nodes[5].FirstChild)
	assert.Nil(t, nodes[5].NextSibling)
	// thirdCol node
	assert.Equal(t, thirdCol, nodes[6].Obj)
	assert.Nil(t, nodes[6].FirstChild)
	assert.Nil(t, nodes[6].NextSibling)
}

func TestCanvas_OverlayStack(t *testing.T) {
	o := &overlayStack{}
	a := canvas.NewRectangle(color.Black)
	b := canvas.NewCircle(color.Black)
	c := canvas.NewRectangle(color.White)
	o.Add(a)
	o.Add(b)
	o.Add(c)
	assert.Equal(t, 3, len(o.List()))
	o.Remove(c)
	assert.Equal(t, 2, len(o.List()))
	o.Remove(a)
	assert.Equal(t, 0, len(o.List()))
}

func deleteAt(c *fyne.Container, index int) {
	if index < len(c.Objects)-1 {
		c.Objects = append(c.Objects[:index], c.Objects[index+1:]...)
	} else {
		c.Objects = c.Objects[:index]
	}
	c.Refresh()
}

func insert(c *fyne.Container, object fyne.CanvasObject, index int) {
	tail := append([]fyne.CanvasObject{object}, c.Objects[index:]...)
	c.Objects = append(c.Objects[:index], tail...)
	c.Refresh()
}

func Prepend(c *fyne.Container, object fyne.CanvasObject) {
	c.Objects = append([]fyne.CanvasObject{object}, c.Objects...)
	c.Refresh()
}

func TestRefreshCount(t *testing.T) { // Issue 2548.
	var (
		c              = &glCanvas{}
		errCh          = make(chan error)
		freed   uint64 = 0
		refresh uint64 = 1000
	)
	c.Initialize(nil, func() {})
	for i := uint64(0); i < refresh; i++ {
		c.Refresh(canvas.NewRectangle(color.Gray16{Y: 1}))
	}

	go func() {
		freed = c.FreeDirtyTextures()
		if freed == 0 {
			errCh <- errors.New("expected to free dirty textures but actually not freed")
			return
		}
		errCh <- nil
	}()
	err := <-errCh
	if err != nil {
		t.Fatal(err)
	}
	if freed != refresh {
		t.Fatalf("FreeDirtyTextures left refresh tasks behind in a frame, got %v, want %v", freed, refresh)
	}
}

func BenchmarkRefresh(b *testing.B) {
	c := &glCanvas{}
	c.Initialize(nil, func() {})

	for i := uint64(1); i < 1<<15; i *= 2 {
		b.Run(fmt.Sprintf("#%d", i), func(b *testing.B) {
			b.ReportAllocs()

			for j := 0; j < b.N; j++ {
				for n := uint64(0); n < i; n++ {
					c.Refresh(canvas.NewRectangle(color.Black))
				}
				c.FreeDirtyTextures()
			}
		})
	}
}
