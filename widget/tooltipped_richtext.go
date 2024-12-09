package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/internal/widget"
)

// TooltippedRichText represents the base element for a rich text-based widget.
//
// NOTE: since the tool tip TooltippedRichText implements desktop.Hoverable while the
// standard TooltippedRichText does not, this widget may result in hover events not
// reaching the parent Hoverable widget. It provides a callback API to allow
// parent widgets to be notified of hover events received on this widget.
type TooltippedRichText struct {
	RichText
	ToolTipWidgetExtend

	// Sets a callback that will be invoked for MouseIn events received
	OnMouseIn func(*desktop.MouseEvent)
	// Sets a callback that will be invoked for MouseMoved events received
	OnMouseMoved func(*desktop.MouseEvent)
	// Sets a callback that will be invoked for MouseOut events received
	OnMouseOut func()
}

// NewTooltippedRichText returns a new TooltippedRichText widget that renders the given text and segments.
// If no segments are specified it will be converted to a single segment using the default text settings.
func NewTooltippedRichText(segments ...RichTextSegment) *TooltippedRichText {
	t := &TooltippedRichText{RichText: RichText{Segments: segments}}
	t.Scroll = widget.ScrollNone
	t.ExtendBaseWidget(t)
	return t
}

// NewTooltippedRichTextWithText returns a new TooltippedRichText widget that renders the given text.
// The string will be converted to a single text segment using the default text settings.
func NewTooltippedRichTextWithText(text string) *TooltippedRichText {
	return NewTooltippedRichText(&TextSegment{
		Style: RichTextStyleInline,
		Text:  text,
	})
}

func (r *TooltippedRichText) ExtendBaseWidget(wid fyne.Widget) {
	r.ExtendToolTipWidget(wid)
	r.RichText.ExtendBaseWidget(wid)
}

func (r *TooltippedRichText) MouseIn(e *desktop.MouseEvent) {
	r.ToolTipWidgetExtend.MouseIn(e)
	if f := r.OnMouseIn; f != nil {
		f(e)
	}
}

func (r *TooltippedRichText) MouseMoved(e *desktop.MouseEvent) {
	r.ToolTipWidgetExtend.MouseMoved(e)
	if f := r.OnMouseMoved; f != nil {
		f(e)
	}
}

func (r *TooltippedRichText) MouseOut() {
	r.ToolTipWidgetExtend.MouseOut()
	if f := r.OnMouseOut; f != nil {
		f()
	}
}
