package widget

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
)

// Slider is a variant of the Fyne Slider widget that also displays the current value.
type SimpleSlider struct {
	BaseWidget

	OnChangeEnded func(float64)

	data   binding.Float
	label  *Label
	layout fyne.Layout
	slider *Slider
}

// NewSimpleSlider returns a new instance of a [Slider] widget.
func NewSimpleSlider(min, max float64) *SimpleSlider {
	d := binding.NewFloat()
	w := &SimpleSlider{
		label:  NewLabel(LabelWithBindedText(binding.FloatToStringWithFormat(d, "%v"))),
		slider: NewSliderWithData(min, max, d),
		data:   d,
	}
	w.updateLayout()
	w.label.Alignment = fyne.TextAlignTrailing
	w.slider.OnChangeEnded = func(v float64) {
		if w.OnChangeEnded == nil {
			return
		}
		w.OnChangeEnded(v)
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetStep sets a custom step for a slider.
func (w *SimpleSlider) SetStep(step float64) {
	w.slider.Step = step
	w.updateLayout()
}

func (w *SimpleSlider) updateLayout() {
	x := NewLabel(LabelWithStaticText(fmt.Sprintf("%v", w.slider.Max+w.slider.Step)))
	minW := x.MinSize().Width
	w.layout = layout.NewColumns(minW, minW)
}

// Value returns the current value of a slider.
func (w *SimpleSlider) Value() float64 {
	return w.slider.Value
}

// SetValue set the value of a slider.
func (w *SimpleSlider) SetValue(v float64) {
	w.slider.SetValue(float64(v))
}

func (w *SimpleSlider) CreateRenderer() fyne.WidgetRenderer {
	c := &fyne.Container{Layout: w.layout, Objects: []fyne.CanvasObject{w.label, w.slider}}
	return NewSimpleRenderer(c)
}
