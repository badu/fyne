package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	rowHeight    = 30
	columnWidth  = 100
	visibleRows  = 10 // Number of visible rows
	totalRows    = 1000
	totalColumns = 5
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Virtualized Table")

	// Data for the table (2D array simulation)
	data := make([][]string, totalRows)
	for i := range data {
		data[i] = make([]string, totalColumns)
		for j := range data[i] {
			data[i][j] = fmt.Sprintf("Cell %d,%d", i, j)
		}
	}

	// Table headers
	headers := []string{"Column 1", "Column 2", "Column 3", "Column 4", "Column 5"}

	createCell := func() fyne.CanvasObject { // Create a new cell
		return widget.NewLabel(widget.LabelWithStaticText("")) // Initially empty label for each cell
	}
	updateCell := func(tcell widget.TableCellID, obj fyne.CanvasObject) {
		rowIndex := tcell.Row
		colIndex := tcell.Col
		if rowIndex < totalRows {
			obj.(*widget.Label).SetText(data[rowIndex][colIndex])
		}
	}

	// Create a table widget
	table := widget.NewTableWithHeaders(
		func() (int, int) { return visibleRows, len(headers) }, // Number of visible rows
		createCell,
		updateCell,
	)

	// Scroll container for the table
	scroll := container.NewVScroll(table)
	//scroll.SetMinSize(fyne.NewSize(columnWidth*totalColumns, rowHeight*totalRows))

	// Scroll offset
	offsetY := 0 // Row offset
	// Listen for scrolling to update the visible rows
	scroll.OnScrolled = func(offset fyne.Position) {
		newOffsetY := int(offset.Y) / rowHeight
		if newOffsetY != offsetY {
			offsetY = newOffsetY
			table.Refresh() // Refresh the table to load the new set of rows
		}
	}

	// Set content in the window
	myWindow.SetContent(scroll)
	myWindow.Resize(fyne.NewSize(columnWidth*totalColumns, rowHeight*visibleRows))
	myWindow.ShowAndRun()
}
