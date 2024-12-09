package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// TooltippedButton widget has a text label and triggers an event func when clicked
type TooltippedButton struct {
	Button
	ToolTipWidgetExtend
}

// NewTooltippedButton creates a new button widget with the set label and tap handler
func NewTooltippedButton(text string, onTapped func()) *TooltippedButton {
	return NewTooltippedButtonWithIcon(text, nil, onTapped)
}

// NewTooltippedButtonWithIcon creates a new button widget with the specified label, themed icon and tap handler
func NewTooltippedButtonWithIcon(text string, icon fyne.Resource, onTapped func()) *TooltippedButton {
	b := &TooltippedButton{
		Button: Button{
			Text:     text,
			Icon:     icon,
			OnTapped: onTapped,
		},
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *TooltippedButton) ExtendBaseWidget(wid fyne.Widget) {
	b.ExtendToolTipWidget(wid)
	b.Button.ExtendBaseWidget(wid)
}

func (b *TooltippedButton) MouseIn(e *desktop.MouseEvent) {
	b.ToolTipWidgetExtend.MouseIn(e)
	b.Button.MouseIn(e)
}

func (b *TooltippedButton) MouseOut() {
	b.ToolTipWidgetExtend.MouseOut()
	b.Button.MouseOut()
}

func (b *TooltippedButton) MouseMoved(e *desktop.MouseEvent) {
	b.ToolTipWidgetExtend.MouseMoved(e)
	b.Button.MouseMoved(e)
}
