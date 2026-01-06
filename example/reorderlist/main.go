package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID("test")
	w := a.NewWindow("win")

	var data []string
	for i := 0; i < 1000; i++ {
		data = append(data, fmt.Sprintf("Test list row %d", i))
	}

	l := widget.NewReorderList(
		func() int { return len(data) },
		func() fyne.CanvasObject { return widget.NewLabel(widget.LabelWithStaticText("")) },
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			co.(*widget.Label).SetText(data[lii])
		},
	)
	var selected widget.ReorderListItemID
	l.OnSelected = func(id widget.ReorderListItemID) { selected = id }
	l.EnableDragging = true
	l.OnDragBegin = l.Select
	l.OnDragEnd = func(_, insertAt widget.ReorderListItemID) {
		newData := make([]string, 0, len(data))
		newData = append(newData, data[:insertAt]...)
		newData = append(newData, data[selected])
		for i := insertAt; i < len(data); i++ {
			if i != selected {
				newData = append(newData, data[i])
			}
		}
		data = newData
		l.UnselectAll()
		l.Refresh()
	}

	w.SetContent(container.NewBorder(
		container.NewStack(
			canvas.NewRectangle(color.RGBA{R: 128, A: 255}), widget.NewLabel(widget.LabelWithStaticText("Pad"))),
		container.NewStack(
			canvas.NewRectangle(color.RGBA{R: 128, A: 255}), widget.NewLabel(widget.LabelWithStaticText("Pad"))),
		nil, nil, l),
	)
	w.Resize(fyne.NewSize(300, 400))
	w.ShowAndRun()
}
