package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// TooltipSlider is a widget that can slide between two fixed values.
type TooltipSlider struct {
	widget.Slider
	ToolTipWidgetExtend
}

// NewTooltipSlider returns a basic slider.
func NewTooltipSlider(min, max float64) *TooltipSlider {
	slider := &TooltipSlider{
		Slider: widget.Slider{
			Value:       0,
			Min:         min,
			Max:         max,
			Step:        1,
			Orientation: widget.Horizontal,
		},
	}
	slider.ExtendBaseWidget(slider)
	return slider
}

// NewSliderWithData returns a slider connected with the specified data source.
func NewSliderWithData(min, max float64, data binding.Float) *TooltipSlider {
	slider := NewTooltipSlider(min, max)
	slider.Bind(data)

	return slider
}

func (s *TooltipSlider) ExtendBaseWidget(wid fyne.Widget) {
	s.ExtendToolTipWidget(wid)
	s.Slider.ExtendBaseWidget(wid)
}

func (s *TooltipSlider) MouseIn(e *desktop.MouseEvent) {
	s.ToolTipWidgetExtend.MouseIn(e)
	s.Slider.MouseIn(e)
}

func (s *TooltipSlider) MouseMoved(e *desktop.MouseEvent) {
	s.ToolTipWidgetExtend.MouseMoved(e)
	s.Slider.MouseMoved(e)
}

func (s *TooltipSlider) MouseOut() {
	s.ToolTipWidgetExtend.MouseOut()
	s.Slider.MouseOut()
}
