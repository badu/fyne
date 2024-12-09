package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
)

// TooltippedCheck widget has a text label and a checked (or unchecked) icon and triggers an event func when toggled
type TooltippedCheck struct {
	Check
	ToolTipWidgetExtend
}

// NewTooltippedCheck creates a new check widget with the set label and change handler
func NewTooltippedCheck(label string, changed func(bool)) *TooltippedCheck {
	c := &TooltippedCheck{
		Check: Check{
			Text:      label,
			OnChanged: changed,
		},
	}
	c.ExtendBaseWidget(c)
	return c
}

// NewTooltippedCheckWithData returns a check widget connected with the specified data source.
func NewTooltippedCheckWithData(label string, data binding.Bool) *TooltippedCheck {
	check := NewTooltippedCheck(label, nil)
	check.Bind(data)

	return check
}

func (c *TooltippedCheck) ExtendBaseWidget(wid fyne.Widget) {
	c.ExtendToolTipWidget(wid)
	c.Check.ExtendBaseWidget(wid)
}

func (c *TooltippedCheck) MouseIn(e *desktop.MouseEvent) {
	c.ToolTipWidgetExtend.MouseIn(e)
	c.Check.MouseIn(e)
}

func (c *TooltippedCheck) MouseMoved(e *desktop.MouseEvent) {
	c.ToolTipWidgetExtend.MouseMoved(e)
	c.Check.MouseMoved(e)
}

func (c *TooltippedCheck) MouseOut() {
	c.ToolTipWidgetExtend.MouseOut()
	c.Check.MouseOut()
}
