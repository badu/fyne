package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type textEdit struct {
	cursorRow, cursorCol *widget.Label
	entry                *widget.Entry
	window               fyne.Window
	changed              binding.Bool

	uri fyne.URI
}

func (e *textEdit) updateStatus() {
	e.cursorRow.SetText(fmt.Sprintf("%d", e.entry.CursorRow+1))
	e.cursorCol.SetText(fmt.Sprintf("%d", e.entry.CursorColumn+1))
}

func (e *textEdit) cut() {
	e.entry.TypedShortcut(&fyne.ShortcutCut{Clipboard: e.window.Clipboard()})
}

func (e *textEdit) copy() {
	e.entry.TypedShortcut(&fyne.ShortcutCopy{Clipboard: e.window.Clipboard()})
}

func (e *textEdit) paste() {
	e.entry.TypedShortcut(&fyne.ShortcutPaste{Clipboard: e.window.Clipboard()})
}

func (e *textEdit) buildToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderOpenIcon(), e.open),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), e.save),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			e.entry.SetText("set text")
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentCutIcon(), e.cut),
		widget.NewToolbarAction(theme.ContentCopyIcon(), e.copy),
		widget.NewToolbarAction(theme.ContentPasteIcon(), e.paste),
		widget.NewToolbarSeparator(),
	)
}

// makeUI loads a new text editor
func (e *textEdit) makeUI() fyne.CanvasObject {
	e.entry = widget.NewEntry(widget.EntryWithMultiline())
	e.cursorRow = widget.NewLabel(widget.LabelWithStaticText("1"))
	e.cursorCol = widget.NewLabel(widget.LabelWithStaticText("1"))

	e.entry.OnCursorChanged = e.updateStatus
	e.entry.OnChanged = func(s string) {
		e.changed.Set(true)
	}

	toolbar := e.buildToolbar()
	status := container.NewHBox(layout.NewSpacer(),
		widget.NewLabel(widget.LabelWithStaticText("Cursor Row:")), e.cursorRow,
		widget.NewLabel(widget.LabelWithStaticText("Col:")), e.cursorCol)
	return container.NewBorder(toolbar, status, nil, nil, container.NewScroll(e.entry))
}
