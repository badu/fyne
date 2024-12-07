package tutorials

import (
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_demo/data"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// containerScreen loads a tab panel for containers
func containerScreen(_ fyne.Window) fyne.CanvasObject {
	content := container.NewBorder(
		widget.NewLabel(
			widget.LabelWithStaticText("Top"),
			widget.LabelWithAlignment(fyne.TextAlignCenter),
			widget.LabelWithStyle(fyne.TextStyle{}),
		),
		widget.NewLabel(
			widget.LabelWithStaticText("Bottom"),
			widget.LabelWithAlignment(fyne.TextAlignCenter),
			widget.LabelWithStyle(fyne.TextStyle{}),
		),
		widget.NewLabel(widget.LabelWithStaticText("Left")),
		widget.NewLabel(widget.LabelWithStaticText("Right")),
		widget.NewLabel(widget.LabelWithStaticText("Border Container")))
	return container.NewCenter(content)
}

func makeAppTabsTab(_ fyne.Window) fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Tab 1", widget.NewLabel(widget.LabelWithStaticText("Content of tab 1"))),
		container.NewTabItem("Tab 2 bigger", widget.NewLabel(widget.LabelWithStaticText("Content of tab 2"))),
		container.NewTabItem("Tab 3", widget.NewLabel(widget.LabelWithStaticText("Content of tab 3"))),
	)
	for i := 4; i <= 12; i++ {
		tabs.Append(container.NewTabItem(fmt.Sprintf("Tab %d", i), widget.NewLabel(widget.LabelWithStaticText(fmt.Sprintf("Content of tab %d", i)))))
	}
	locations := makeTabLocationSelect(tabs.SetTabLocation)
	return container.NewBorder(locations, nil, nil, nil, tabs)
}

func makeBorderLayout(_ fyne.Window) fyne.CanvasObject {
	top := makeCell()
	bottom := makeCell()
	left := makeCell()
	right := makeCell()
	middle := widget.NewLabel(
		widget.LabelWithStaticText("BorderLayout"),
		widget.LabelWithAlignment(fyne.TextAlignCenter),
		widget.LabelWithStyle(fyne.TextStyle{}),
	)

	return container.NewBorder(top, bottom, left, right, middle)
}

func makeBoxLayout(_ fyne.Window) fyne.CanvasObject {
	top := makeCell()
	bottom := makeCell()
	middle := widget.NewLabel(widget.LabelWithStaticText("BoxLayout"))
	center := makeCell()
	right := makeCell()

	col := container.NewVBox(top, middle, bottom)

	return container.NewHBox(col, center, right)
}

func makeButtonList(count int) []fyne.CanvasObject {
	var items []fyne.CanvasObject
	for i := 1; i <= count; i++ {
		index := i // capture
		items = append(items, widget.NewButton(
			widget.ButtonWithLabel("Button "+strconv.Itoa(index)),
			widget.ButtonWithCallback(
				func() {
					fmt.Println("Tapped", index)
				},
			),
		),
		)
	}

	return items
}

func makeCell() fyne.CanvasObject {
	rect := canvas.NewRectangle(&color.NRGBA{R: 128, G: 128, B: 128, A: 255})
	rect.SetMinSize(fyne.NewSize(30, 30))
	return rect
}

func makeCenterLayout(_ fyne.Window) fyne.CanvasObject {
	middle := widget.NewButton(widget.ButtonWithLabel("CenterLayout"))

	return container.NewCenter(middle)
}

func makeDocTabsTab(_ fyne.Window) fyne.CanvasObject {
	tabs := container.NewDocTabs(
		container.NewTabItem("Doc 1", widget.NewLabel(widget.LabelWithStaticText("Content of document 1"))),
		container.NewTabItem("Doc 2 bigger", widget.NewLabel(widget.LabelWithStaticText("Content of document 2"))),
		container.NewTabItem("Doc 3", widget.NewLabel(widget.LabelWithStaticText("Content of document 3"))),
	)
	i := 3
	tabs.CreateTab = func() *container.TabItem {
		i++
		return container.NewTabItem(fmt.Sprintf("Doc %d", i), widget.NewLabel(
			widget.LabelWithStaticText(fmt.Sprintf("Content of document %d", i)),
		),
		)
	}
	locations := makeTabLocationSelect(tabs.SetTabLocation)
	return container.NewBorder(locations, nil, nil, nil, tabs)
}

func makeGridLayout(_ fyne.Window) fyne.CanvasObject {
	box1 := makeCell()
	box2 := widget.NewLabel(widget.LabelWithStaticText("Grid"))
	box3 := makeCell()
	box4 := makeCell()

	return container.NewGridWithColumns(2,
		box1, box2, box3, box4)
}

func makeInnerWindowTab(_ fyne.Window) fyne.CanvasObject {
	label := widget.NewLabel(widget.LabelWithStaticText("Window content for inner demo"))
	win1 := container.NewInnerWindow(
		"Inner Demo",
		container.NewVBox(
			label,
			widget.NewButton(
				widget.ButtonWithLabel("Tap Me"),
				widget.ButtonWithCallback(
					func() {
						label.SetText("Tapped")
					},
				),
			),
		),
	)
	win1.Icon = data.FyneLogo

	win2 := container.NewInnerWindow("Inner2", widget.NewLabel(widget.LabelWithStaticText("Win 2")))

	multi := container.NewMultipleWindows()
	multi.Windows = []*container.InnerWindow{win1, win2}
	return multi
}

func makeScrollTab(_ fyne.Window) fyne.CanvasObject {
	hlist := makeButtonList(20)
	vlist := makeButtonList(50)

	horiz := container.NewHScroll(container.NewHBox(hlist...))
	vert := container.NewVScroll(container.NewVBox(vlist...))

	return container.NewAdaptiveGrid(2,
		container.NewBorder(horiz, nil, nil, nil, vert),
		makeScrollBothTab())
}

func makeScrollBothTab() fyne.CanvasObject {
	logo := canvas.NewImageFromResource(data.FyneLogo)
	logo.SetMinSize(fyne.NewSize(800, 800))

	scroll := container.NewScroll(logo)
	scroll.Resize(fyne.NewSize(400, 400))

	return scroll
}

func makeSplitTab(_ fyne.Window) fyne.CanvasObject {
	left := widget.NewEntry(widget.EntryWithMultiline())
	left.Wrapping = fyne.TextWrapWord
	left.SetText("Long text is looooooooooooooong")
	right := container.NewVSplit(
		widget.NewLabel(widget.LabelWithStaticText("Label")),
		widget.NewButton(
			widget.ButtonWithLabel("Button"),
			widget.ButtonWithCallback(
				func() { fmt.Println("button tapped!") },
			),
		),
	)
	return container.NewHSplit(container.NewVScroll(left), right)
}

func makeTabLocationSelect(callback func(container.TabLocation)) *widget.Select {
	locations := widget.NewSelect(
		widget.SelectWithOptions("Top", "Bottom", "Leading", "Trailing"),
		widget.SelectWithCallback(
			func(s string) {
				callback(map[string]container.TabLocation{
					"Top":      container.TabLocationTop,
					"Bottom":   container.TabLocationBottom,
					"Leading":  container.TabLocationLeading,
					"Trailing": container.TabLocationTrailing,
				}[s])
			},
		),
	)
	locations.SetSelected("Top")
	return locations
}
