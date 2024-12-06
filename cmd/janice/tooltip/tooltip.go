package tooltip

import (
	"fyne.io/fyne/v2/container"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/janice/tooltip/shadow"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	ToolTipTextStyleMutex sync.Mutex
	ToolTipTextStyle      = widget.RichTextStyle{SizeName: theme.SizeNameCaptionText}
)

type ToolTip struct {
	widget.BaseWidget

	Text string

	richtext *widget.RichText
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
		t.richtext = widget.NewRichTextWithText(t.Text)
		t.richtext.Wrapping = fyne.TextWrapWord
	}
	ToolTipTextStyleMutex.Lock()
	style := ToolTipTextStyle
	ToolTipTextStyleMutex.Unlock()
	t.richtext.Segments[0].(*widget.TextSegment).Text = t.Text
	t.richtext.Segments[0].(*widget.TextSegment).Style = style
}

type toolTipRenderer struct {
	*shadow.ShadowingRenderer
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
	r.ShadowingRenderer = shadow.NewShadowingRenderer([]fyne.CanvasObject{&r.backgroundRect, t.richtext}, shadow.ToolTipLevel)
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
	return container.NewStack(windowContent, &NewToolTipLayer(canvas).Container)
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
func AddPopUpToolTipLayer(p *widget.PopUp) {
	l := NewPopUpToolTipLayer(p)
	p.Content = container.NewStack(p.Content, &l.Container)
}

// DestroyPopUpToolTipLayer destroys the tool tip layer for a given pop up.
// It should be called after the pop up is hidden and will no longer be shown
// to free associated memory resources.
func DestroyPopUpToolTipLayer(p *widget.PopUp) {
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
