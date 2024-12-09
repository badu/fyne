package widget

import (
	"context"
	"errors"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

var (
	ToolTipTextStyleMutex sync.Mutex
	ToolTipTextStyle      = RichTextStyle{SizeName: theme.SizeNameCaptionText}
)

func NewStack(objects ...fyne.CanvasObject) *fyne.Container {
	return &fyne.Container{Layout: layout.NewStackLayout(), Objects: objects}
}

// ToolTip is a base struct for building new tool tip supporting widgets.
// Inherit from this struct instead of from `widget.BaseWidget` to automatically
// add tool tip support to your widget.
type ToolTip struct {
	BaseWidget
	toolTipContext

	Text     string
	richtext *RichText
	wid      fyne.Widget
}

func NewToolTip(text string) *ToolTip {
	t := &ToolTip{Text: text}
	t.ExtendBaseWidget(t)
	return t
}

func (t *ToolTip) MinSize() fyne.Size {
	// zero so that ToolTip won't force a PopUp or other overlay to a larger size
	// TextMinSize returns the actual minimum size for rendering
	return fyne.NewSize(0, 0)
}

func (t *ToolTip) Resize(size fyne.Size) {
	t.updateRichText()
	t.richtext.Resize(size)
	t.BaseWidget.Resize(size)
}

func (t *ToolTip) TextMinSize() fyne.Size {
	t.updateRichText()
	return t.richtext.MinSize().Subtract(
		fyne.NewSquareSize(2 * t.Theme().Size(theme.SizeNameInnerPadding))).
		Add(fyne.NewSize(2, 8))
}

func (t *ToolTip) NonWrappingTextWidth() float32 {
	ToolTipTextStyleMutex.Lock()
	style := ToolTipTextStyle
	ToolTipTextStyleMutex.Unlock()
	th := t.Theme()
	return fyne.MeasureText(t.Text, th.Size(style.SizeName), style.TextStyle).Width + th.Size(theme.SizeNameInnerPadding)*2
}

func (t *ToolTip) updateRichText() {
	if t.richtext == nil {
		t.richtext = NewRichTextWithText(t.Text)
		t.richtext.Wrapping = fyne.TextWrapWord
	}
	ToolTipTextStyleMutex.Lock()
	style := ToolTipTextStyle
	ToolTipTextStyleMutex.Unlock()
	t.richtext.Segments[0].(*TextSegment).Text = t.Text
	t.richtext.Segments[0].(*TextSegment).Style = style
}

type toolTipRenderer struct {
	*ShadowingRenderer
	toolTip        *ToolTip
	backgroundRect canvas.Rectangle
}

func (r *toolTipRenderer) Layout(s fyne.Size) {
	r.LayoutShadow(s, fyne.NewPos(0, 0))
	r.backgroundRect.Resize(s)
	r.backgroundRect.Move(fyne.NewPos(0, 0))
	innerPad := r.toolTip.Theme().Size(theme.SizeNameInnerPadding)
	r.toolTip.richtext.Resize(s)
	r.toolTip.richtext.Move(fyne.NewPos(0, -innerPad+3))
}

func (r *toolTipRenderer) MinSize() fyne.Size {
	return r.toolTip.TextMinSize()
}

func (r *toolTipRenderer) Refresh() {
	r.ShadowingRenderer.RefreshShadow()
	th := r.toolTip.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	r.backgroundRect.FillColor = th.Color(theme.ColorNameOverlayBackground, variant)
	r.backgroundRect.StrokeColor = th.Color(theme.ColorNameInputBorder, variant)
	r.backgroundRect.StrokeWidth = th.Size(theme.SizeNameInputBorder)
	r.backgroundRect.Refresh()
	r.toolTip.updateRichText()
	r.toolTip.richtext.Refresh()

	canvas.Refresh(r.toolTip)
}

func (t *ToolTip) CreateRenderer() fyne.WidgetRenderer {
	t.updateRichText()
	r := &toolTipRenderer{toolTip: t}
	r.ShadowingRenderer = NewShadowingRenderer([]fyne.CanvasObject{&r.backgroundRect, t.richtext}, ToolTipLevel)
	return r
}

// AddWindowToolTipLayer adds a layer to the given window content for tool tips to be drawn into.
// This call is required for each new window you create to enable tool tips to be shown in it.
// It is typically invoked with `window.SetContent` as
//
//	`window.SetContent(fynetooltip.AddWindowToolTipLayer(myContent, window.Canvas()))``
//
// If the window not your main window and is closed before the app exits, it is important to call
// `DestroyWindowToolTipLayer` to release memory resources associated with the tool tip layer.
func AddWindowToolTipLayer(windowContent fyne.CanvasObject, canvas fyne.Canvas) fyne.CanvasObject {
	return NewStack(windowContent, &NewToolTipLayer(canvas).Container)
}

// DestroyWindowToolTipLayer destroys the tool tip layer for a given window canvas.
// It should be called after a window is closed to free associated memory resources.
func DestroyWindowToolTipLayer(canvas fyne.Canvas) {
	DestroyToolTipLayerForCanvas(canvas)
}

// AddPopUpToolTipLayer adds a layer to the given `*widget.PopUp` for tool tips to be drawn into.
// This call is required for each new PopUp you create to enable tool tips to be shown in it.
// It is invoked after the PopUp has been created with content, but before it is shown.
// Once the pop up is hidden and will not be shown anymore, it is important to call
// `DestroyPopUpToolTipLayer` to release memory resources associated with the tool tip layer.
// A pop up that will be shown again should not have DestroyPopUpToolTipLayer called.
func AddPopUpToolTipLayer(p *PopUp) {
	l := NewPopUpToolTipLayer(p)
	p.Content = NewStack(p.Content, &l.Container)
}

// DestroyPopUpToolTipLayer destroys the tool tip layer for a given pop up.
// It should be called after the pop up is hidden and will no longer be shown
// to free associated memory resources.
func DestroyPopUpToolTipLayer(p *PopUp) {
	DestroyToolTipLayerForPopup(p)
}

// SetToolTipTextStyle sets the TextStyle that will be used to render tool tip text.
func SetToolTipTextStyle(style fyne.TextStyle) {
	ToolTipTextStyleMutex.Lock()
	defer ToolTipTextStyleMutex.Unlock()
	ToolTipTextStyle.TextStyle = style
}

// SetToolTipTextSizeName sets the theme size name that will control the size
// of tool tip text. By default, tool tips use theme.SizeNameCaptionText.
func SetToolTipTextSizeName(sizeName fyne.ThemeSizeName) {
	ToolTipTextStyleMutex.Lock()
	defer ToolTipTextStyleMutex.Unlock()
	ToolTipTextStyle.SizeName = sizeName
}

type toolTipContext struct {
	lock                 sync.Mutex
	toolTipHandle        *ToolTipHandle
	absoluteMousePos     fyne.Position
	pendingToolTipCtx    context.Context
	pendingToolTipCancel context.CancelFunc
}

func (t *ToolTip) ExtendBaseWidget(wid fyne.Widget) {
	t.wid = wid
	t.BaseWidget.ExtendBaseWidget(wid)
}

func (t *ToolTip) SetToolTip(toolTip string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Text = toolTip
}

func (t *ToolTip) ToolTip() string {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.Text
}

func (t *ToolTip) MouseIn(e *desktop.MouseEvent) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.Text != "" {
		t.absoluteMousePos = e.AbsolutePosition
		if t.wid == nil {
			fyne.LogError("", errors.New("missing ExtendBaseWidget call for ToolTip"))
			return
		}
		t.setPendingToolTip(t.wid, t.Text)
	}
}

func (t *ToolTip) MouseOut() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.cancelToolTip()
}

func (t *ToolTip) MouseMoved(e *desktop.MouseEvent) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.absoluteMousePos = e.AbsolutePosition
}

// ToolTipWidgetExtend is a struct for extending existing widgets for tool tip support.
// Use this to extend existing widgets for tool tip support. When creating an extended
// widget with ToolTipWidgetExtend you must override ExtendBaseWidget to call both the
// ExtendBaseWidget implementation of the parent widget, and ExtendToolTipWidget.
type ToolTipWidgetExtend struct {
	toolTipContext

	// Obj is the widget this ToolTipWidgetExtend is embedded in; set by ExtendToolTipWidget
	Obj fyne.CanvasObject

	toolTip string
}

func (t *ToolTipWidgetExtend) SetToolTip(toolTip string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.toolTip = toolTip
}

func (t *ToolTipWidgetExtend) ToolTip() string {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.toolTip
}

// ExtendToolTipWidget sets up a tool tip extended widget.
func (t *ToolTipWidgetExtend) ExtendToolTipWidget(wid fyne.Widget) {
	t.Obj = wid
}

func (t *ToolTipWidgetExtend) MouseIn(e *desktop.MouseEvent) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.toolTip != "" {
		t.absoluteMousePos = e.AbsolutePosition
		t.setPendingToolTip(t.Obj, t.toolTip)
	}
}

func (t *ToolTipWidgetExtend) MouseOut() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.cancelToolTip()
}

func (t *ToolTipWidgetExtend) MouseMoved(e *desktop.MouseEvent) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.absoluteMousePos = e.AbsolutePosition
}

func (t *toolTipContext) setPendingToolTip(wid fyne.CanvasObject, toolTipText string) {
	ctx, cancel := context.WithCancel(context.Background())
	t.pendingToolTipCtx, t.pendingToolTipCancel = ctx, cancel

	go func() {
		<-time.After(NextToolTipDelayTime())
		select {
		case <-ctx.Done():
			return
		default:
			t.lock.Lock()
			defer t.lock.Unlock()
			t.cancelToolTip() // don't leak ctx resources
			pos := t.absoluteMousePos
			canvas := fyne.CurrentApp().Driver().CanvasForObject(wid)
			t.toolTipHandle = ShowToolTipAtMousePosition(canvas, pos, toolTipText)
		}
	}()
}

func (t *toolTipContext) cancelToolTip() {
	if t.pendingToolTipCancel != nil {
		t.pendingToolTipCancel()
		t.pendingToolTipCancel = nil
		t.pendingToolTipCtx = nil
	}
	if t.toolTipHandle != nil {
		HideToolTip(t.toolTipHandle)
		t.toolTipHandle = nil
	}
}

const (
	initialToolTipDelay             = 750 * time.Millisecond
	subsequentToolTipDelay          = 300 * time.Millisecond
	subsequentToolTipDelayValidTime = 1500 * time.Millisecond
)

var (
	toolTipLayers             = make(map[fyne.Canvas]*ToolTipLayer)
	lastToolTipShownUnixMilli int64
)

func NextToolTipDelayTime() time.Duration {
	if time.Now().UnixMilli()-lastToolTipShownUnixMilli < subsequentToolTipDelayValidTime.Milliseconds() {
		return subsequentToolTipDelay
	}
	return initialToolTipDelay
}

type ToolTipHandle struct {
	canvas  fyne.Canvas
	overlay fyne.CanvasObject
}

type ToolTipLayer struct {
	Container fyne.Container
	overlays  map[fyne.CanvasObject]*ToolTipLayer
}

func NewToolTipLayer(canvas fyne.Canvas) *ToolTipLayer {
	t := &ToolTipLayer{}
	toolTipLayers[canvas] = t
	return t
}

func DestroyToolTipLayerForCanvas(canvas fyne.Canvas) {
	delete(toolTipLayers, canvas)
}

func NewPopUpToolTipLayer(popUp *PopUp) *ToolTipLayer {
	ct := toolTipLayers[popUp.Canvas]
	if ct == nil {
		fyne.LogError("", errors.New("no tool tip layer created for parent canvas"))
		return nil
	}
	t := &ToolTipLayer{}
	if ct.overlays == nil {
		ct.overlays = make(map[fyne.CanvasObject]*ToolTipLayer)
	}
	ct.overlays[popUp] = t
	return t
}

func DestroyToolTipLayerForPopup(popUp *PopUp) {
	ct := toolTipLayers[popUp.Canvas]
	if ct != nil {
		delete(ct.overlays, popUp)
	}
}

func ShowToolTipAtMousePosition(canvas fyne.Canvas, pos fyne.Position, text string) *ToolTipHandle {
	if canvas == nil {
		fyne.LogError("", errors.New("no canvas associated with tool tip widget"))
		return nil
	}

	lastToolTipShownUnixMilli = time.Now().UnixMilli()
	overlay := canvas.Overlays().Top()
	handle := &ToolTipHandle{canvas: canvas, overlay: overlay}
	tl := findToolTipLayer(handle, true)
	if tl == nil {
		return nil
	}

	t := NewToolTip(text)
	tl.Container.Objects = []fyne.CanvasObject{t}

	var zeroPos fyne.Position
	if pop, ok := overlay.(*PopUp); ok && pop != nil {
		zeroPos = pop.Content.Position()
	} else {
		zeroPos = fyne.CurrentApp().Driver().AbsolutePositionForObject(&tl.Container)
	}

	sizeAndPositionToolTip(zeroPos, pos.Subtract(zeroPos), t, canvas)
	tl.Container.Refresh()
	return handle
}

func HideToolTip(handle *ToolTipHandle) {
	if handle == nil {
		return
	}
	tl := findToolTipLayer(handle, false)
	if tl == nil {
		return
	}
	tl.Container.Objects = nil
	tl.Container.Refresh()
}

func findToolTipLayer(handle *ToolTipHandle, logErr bool) *ToolTipLayer {
	tl := toolTipLayers[handle.canvas]
	if tl == nil {
		if logErr {
			fyne.LogError("", errors.New("no tool tip layer created for window canvas"))
		}
		return nil
	}
	if handle.overlay != nil {
		tl = tl.overlays[handle.overlay]
		if tl == nil {
			if logErr {
				fyne.LogError("", errors.New("no tool tip layer created for current overlay"))
			}
			return nil
		}
	}
	return tl
}

const (
	maxToolTipWidth = 600
	belowMouseDist  = 16
	aboveMouseDist  = 8
)

func sizeAndPositionToolTip(zeroPos, relPos fyne.Position, t *ToolTip, canvas fyne.Canvas) {
	canvasSize := canvas.Size()
	canvasPad := theme.Padding()

	// calculate width of tooltip
	w := fyne.Min(t.NonWrappingTextWidth(), fyne.Min(canvasSize.Width-canvasPad*2, maxToolTipWidth))
	t.Resize(fyne.NewSize(w, 1)) // set up to get min height with wrapping at width w
	t.Resize(fyne.NewSize(w, t.TextMinSize().Height))

	// if would overflow the right edge of the window, move back to the left
	if rightEdge := relPos.X + zeroPos.X + w; rightEdge > canvasSize.Width-canvasPad {
		relPos.X -= rightEdge - canvasSize.Width + canvasPad
	}

	// if would overflow the bottom of the window, move above mouse
	if bottomEdge := relPos.Y + zeroPos.Y + t.Size().Height + belowMouseDist; bottomEdge > canvasSize.Height-canvasPad {
		relPos.Y -= t.Size().Height + aboveMouseDist
	} else {
		relPos.Y += belowMouseDist
	}

	t.Move(relPos)
}
