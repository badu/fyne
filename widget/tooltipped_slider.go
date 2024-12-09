package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
)

// TooltippedSlider is a widget that can slide between two fixed values.
type TooltippedSlider struct {
	Slider
	ToolTipWidgetExtend
}

// NewTooltippedSlider returns a basic slider.
func NewTooltippedSlider(min, max float64) *TooltippedSlider {
	slider := &TooltippedSlider{
		Slider: Slider{
			Value:       0,
			Min:         min,
			Max:         max,
			Step:        1,
			Orientation: Horizontal,
		},
	}
	slider.ExtendBaseWidget(slider)
	return slider
}

// NewTooltipSliderWithData returns a slider connected with the specified data source.
func NewTooltipSliderWithData(min, max float64, data binding.Float) *TooltippedSlider {
	slider := NewTooltippedSlider(min, max)
	slider.Bind(data)

	return slider
}

func (s *TooltippedSlider) ExtendBaseWidget(wid fyne.Widget) {
	s.ExtendToolTipWidget(wid)
	s.Slider.ExtendBaseWidget(wid)
}

func (s *TooltippedSlider) MouseIn(e *desktop.MouseEvent) {
	s.ToolTipWidgetExtend.MouseIn(e)
	s.Slider.MouseIn(e)
}

func (s *TooltippedSlider) MouseMoved(e *desktop.MouseEvent) {
	s.ToolTipWidgetExtend.MouseMoved(e)
	s.Slider.MouseMoved(e)
}

func (s *TooltippedSlider) MouseOut() {
	s.ToolTipWidgetExtend.MouseOut()
	s.Slider.MouseOut()
}
