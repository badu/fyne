package widget

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// TooltippedHyperlink widget is a text component with appropriate padding and layout.
// When clicked, the default web browser should open with a URL
type TooltippedHyperlink struct {
	Hyperlink
	ToolTipWidgetExtend
}

// NewTooltippedHyperlink creates a new hyperlink widget with the set text content
func NewTooltippedHyperlink(text string, url *url.URL) *TooltippedHyperlink {
	return NewTooltippedHyperlinkWithStyle(text, url, fyne.TextAlignLeading, fyne.TextStyle{})
}

// NewTooltippedHyperlinkWithStyle creates a new hyperlink widget with the set text content
func NewTooltippedHyperlinkWithStyle(text string, url *url.URL, alignment fyne.TextAlign, style fyne.TextStyle) *TooltippedHyperlink {
	l := &TooltippedHyperlink{
		Hyperlink: Hyperlink{
			Text:      text,
			URL:       url,
			Alignment: alignment,
			TextStyle: style,
		},
	}

	l.ExtendBaseWidget(l)
	return l
}

func (l *TooltippedHyperlink) ExtendBaseWidget(wid fyne.Widget) {
	l.ExtendToolTipWidget(wid)
	l.Hyperlink.ExtendBaseWidget(wid)
}

func (l *TooltippedHyperlink) MouseIn(e *desktop.MouseEvent) {
	l.ToolTipWidgetExtend.MouseIn(e)
	l.Hyperlink.MouseIn(e)
}

func (l *TooltippedHyperlink) MouseOut() {
	l.ToolTipWidgetExtend.MouseOut()
	l.Hyperlink.MouseOut()
}

func (l *TooltippedHyperlink) MouseMoved(e *desktop.MouseEvent) {
	l.ToolTipWidgetExtend.MouseMoved(e)
	l.Hyperlink.MouseMoved(e)
}
