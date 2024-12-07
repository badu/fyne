package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
)

// Label widget is a label component with appropriate padding and layout.
type Label struct {
	BaseWidget
	Text      string
	Alignment fyne.TextAlign // The alignment of the text
	Wrapping  fyne.TextWrap  // The wrapping of the text
	TextStyle fyne.TextStyle // The style of the label text

	// The truncation mode of the text
	//
	// Since: 2.4
	Truncation fyne.TextTruncation
	// Importance informs how the label should be styled, i.e. warning or disabled
	//
	// Since: 2.4
	Importance Importance

	provider *RichText
	binder   basicBinder
}

type LabelOption func(*Label)

func LabelWithStaticText(text string) LabelOption {
	return func(l *Label) {
		l.Text = text
	}
}

func LabelWithBindedText(data binding.String) LabelOption {
	return func(l *Label) {
		l.Bind(data)
	}
}

func LabelWithAlignment(alignment fyne.TextAlign) LabelOption {
	return func(l *Label) {
		l.Alignment = alignment
	}
}

func LabelWithStyle(style fyne.TextStyle) LabelOption {
	return func(l *Label) {
		l.TextStyle = style
	}
}

func LabelWithTruncation(truncation fyne.TextTruncation) LabelOption {
	return func(l *Label) {
		l.Truncation = truncation
	}
}

func LabelWithWrapping(wrapping fyne.TextWrap) LabelOption {
	return func(l *Label) {
		l.Wrapping = wrapping
	}
}

// NewLabel creates a new label widget with the set text content
func NewLabel(options ...LabelOption) *Label {
	result := &Label{}
	for _, opt := range options {
		opt(result)
	}
	result.ExtendBaseWidget(result)
	return result
}

// Bind connects the specified data source to this Label.
// The current value will be displayed and any changes in the data will cause the widget to update.
//
// Since: 2.0
func (l *Label) Bind(data binding.String) {
	l.binder.SetCallback(l.updateFromData) // This could only be done once, maybe in ExtendBaseWidget?
	l.binder.Bind(data)
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (l *Label) CreateRenderer() fyne.WidgetRenderer {
	l.provider = NewRichTextWithText(l.Text)
	l.ExtendBaseWidget(l)
	l.syncSegments()

	return NewSimpleRenderer(l.provider)
}

// MinSize returns the size that this label should not shrink below.
//
// Implements: fyne.Widget
func (l *Label) MinSize() fyne.Size {
	l.ExtendBaseWidget(l)
	return l.BaseWidget.MinSize()
}

// Refresh triggers a redraw of the label.
//
// Implements: fyne.Widget
func (l *Label) Refresh() {
	if l.provider == nil { // not created until visible
		return
	}
	l.syncSegments()
	l.provider.Refresh()
	l.BaseWidget.Refresh()
}

// Resize sets a new size for the label.
// This should only be called if it is not in a container with a layout manager.
//
// Implements: fyne.Widget
func (l *Label) Resize(s fyne.Size) {
	l.BaseWidget.Resize(s)
	if l.provider != nil {
		l.provider.Resize(s)
	}
}

// SetText sets the text of the label
func (l *Label) SetText(text string) {
	l.propertyLock.Lock()
	l.Text = text
	l.propertyLock.Unlock()
	l.Refresh()
}

// Unbind disconnects any configured data source from this Label.
// The current value will remain at the last value of the data source.
//
// Since: 2.0
func (l *Label) Unbind() {
	l.binder.Unbind()
}

func (l *Label) syncSegments() {
	l.propertyLock.RLock()
	defer l.propertyLock.RUnlock()

	l.provider.Wrapping = l.Wrapping
	l.provider.Truncation = l.Truncation

	l.provider.Segments[0].(*TextSegment).Style = RichTextStyle{
		Alignment: l.Alignment,
		Inline:    true,
		TextStyle: l.TextStyle,
	}
	l.provider.Segments[0].(*TextSegment).Text = l.Text

	switch l.Importance {
	case LowImportance:
		l.provider.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameDisabled
	case MediumImportance:
		l.provider.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameForeground
	case HighImportance:
		l.provider.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNamePrimary
	case DangerImportance:
		l.provider.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameError
	case WarningImportance:
		l.provider.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameWarning
	case SuccessImportance:
		l.provider.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameSuccess
	default:
		l.provider.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameForeground
	}

}

func (l *Label) updateFromData(data binding.DataItem) {
	if data == nil {
		return
	}

	textSource, ok := data.(binding.String)
	if !ok {
		return
	}

	val, err := textSource.Get()
	if err != nil {
		fyne.LogError("Error getting current data value", err)
		return
	}

	l.SetText(val)
}
