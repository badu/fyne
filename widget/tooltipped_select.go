package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// TooltippedSelect widget has a list of options, with the current one shown, and triggers an event func when clicked
type TooltippedSelect struct {
	Select
	ToolTipWidgetExtend
}

// NewTooltippedSelect creates a new select widget with the set list of options and changes handler
func NewTooltippedSelect(options []string, changed func(string)) *TooltippedSelect {
	s := &TooltippedSelect{
		Select: Select{
			OnChanged:   changed,
			Options:     options,
			PlaceHolder: "(TooltippedSelect one)",
		},
	}
	s.ExtendBaseWidget(s)
	return s
}

func (s *TooltippedSelect) ExtendBaseWidget(wid fyne.Widget) {
	s.ExtendToolTipWidget(wid)
	s.Select.ExtendBaseWidget(wid)
}

func (s *TooltippedSelect) MouseIn(e *desktop.MouseEvent) {
	s.ToolTipWidgetExtend.MouseIn(e)
	s.Select.MouseIn(e)
}

func (s *TooltippedSelect) MouseMoved(e *desktop.MouseEvent) {
	s.ToolTipWidgetExtend.MouseMoved(e)
	s.Select.MouseMoved(e)
}

func (s *TooltippedSelect) MouseOut() {
	s.ToolTipWidgetExtend.MouseOut()
	s.Select.MouseOut()
}
