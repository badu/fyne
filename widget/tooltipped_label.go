package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
)

// TooltippedLabel widget is a label component with appropriate padding and layout.
//
// NOTE: since the tool tip label implements desktop.Hoverable while the
// standard TooltippedLabel does not, this widget may result in hover events not
// reaching the parent Hoverable widget. It provides a callback API to allow
// parent widgets to be notified of hover events received on this widget.
type TooltippedLabel struct {
	Label
	ToolTipWidgetExtend

	// Sets a callback that will be invoked for MouseIn events received
	OnMouseIn func(*desktop.MouseEvent)
	// Sets a callback that will be invoked for MouseMoved events received
	OnMouseMoved func(*desktop.MouseEvent)
	// Sets a callback that will be invoked for MouseOut events received
	OnMouseOut func()
}

// NewTooltippedLabel creates a new label widget with the set text content
func NewTooltippedLabel(text string) *TooltippedLabel {
	return NewTooltippedLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{})
}

// NewTooltippedLabelWithData returns an TooltippedLabel widget connected to the specified data source.
func NewTooltippedLabelWithData(data binding.String) *TooltippedLabel {
	label := NewTooltippedLabel("")
	label.Bind(data)

	return label
}

// NewTooltippedLabelWithStyle creates a new label widget with the set text content
func NewTooltippedLabelWithStyle(text string, alignment fyne.TextAlign, style fyne.TextStyle) *TooltippedLabel {
	l := &TooltippedLabel{
		Label: Label{
			Text:      text,
			Alignment: alignment,
			TextStyle: style,
		},
	}

	l.ExtendBaseWidget(l)
	return l
}

func (l *TooltippedLabel) ExtendBaseWidget(wid fyne.Widget) {
	l.ExtendToolTipWidget(wid)
	l.Label.ExtendBaseWidget(wid)
}

func (l *TooltippedLabel) MouseIn(e *desktop.MouseEvent) {
	l.ToolTipWidgetExtend.MouseIn(e)
	if f := l.OnMouseIn; f != nil {
		f(e)
	}
}

func (l *TooltippedLabel) MouseMoved(e *desktop.MouseEvent) {
	l.ToolTipWidgetExtend.MouseMoved(e)
	if f := l.OnMouseMoved; f != nil {
		f(e)
	}
}

func (l *TooltippedLabel) MouseOut() {
	l.ToolTipWidgetExtend.MouseOut()
	if f := l.OnMouseOut; f != nil {
		f()
	}
}
