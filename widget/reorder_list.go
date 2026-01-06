package widget

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/internal/widget"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
)

// ReorderListItemID uniquely identifies an item within a list.
type ReorderListItemID = int

// ReorderList is a widget that pools list items for performance and
// lays the items out in a vertical direction inside of a scroller.
// By default, List requires that all items are the same size, but specific
// rows can have their heights set with SetItemHeight.
//
// Since: 1.4
type ReorderList struct {
	BaseWidget

	// Length is a callback for returning the number of items in the list.
	Length func() int `json:"-"`

	// CreateItem is a callback invoked to create a new widget to render
	// a row in the list.
	CreateItem func() fyne.CanvasObject `json:"-"`

	// UpdateItem is a callback invoked to update a list row widget
	// to display a new row in the list. The UpdateItem callback should
	// only update the given item, it should not invoke APIs that would
	// change other properties of the list itself.
	UpdateItem func(id ReorderListItemID, item fyne.CanvasObject) `json:"-"`

	// OnSelected is a callback to be notified when a given item
	// in the list has been selected.
	OnSelected func(id ReorderListItemID) `json:"-"`

	// OnSelected is a callback to be notified when a given item
	// in the list has been unselected.
	OnUnselected func(id ReorderListItemID) `json:"-"`

	// HideSeparators hides the separators between list rows
	//
	// Since: 2.5
	HideSeparators bool

	// Enable drag-and-drop of rows within the list
	EnableDragging bool

	// OnDragEnd is the callback that is invoked when a row is dragged and dropped.
	// The `draggedTo` value is the ListItemID where the dragged row
	// would be inserted, with a value of  0 meaning before the first item
	// and a value of list.Length() meaning after the last item in the list.
	OnDragEnd func(draggedFrom, draggedTo ReorderListItemID) `json:"-"`

	// OnDragBegin is the callback invoked when a row begins dragging.
	OnDragBegin func(id ReorderListItemID) `json:"-"`

	currentFocus  ReorderListItemID
	focused       bool
	scroller      *widget.Scroll
	selected      []ReorderListItemID
	itemMin       fyne.Size
	itemHeights   map[ReorderListItemID]float32
	offsetY       float32
	offsetUpdated func(fyne.Position)
}

// NewList creates and returns a list widget for displaying items in
// a vertical layout with scrolling and caching for performance.
//
// Since: 1.4
func NewReorderList(length func() int, createItem func() fyne.CanvasObject, updateItem func(ReorderListItemID, fyne.CanvasObject)) *ReorderList {
	list := &ReorderList{Length: length, CreateItem: createItem, UpdateItem: updateItem}
	list.ExtendBaseWidget(list)
	return list
}

// NewListWithData creates a new list widget that will display the contents of the provided data.
//
// Since: 2.0
func NewListWithData(data binding.DataList, createItem func() fyne.CanvasObject, updateItem func(binding.DataItem, fyne.CanvasObject)) *ReorderList {
	l := NewReorderList(
		data.Length,
		createItem,
		func(i ReorderListItemID, o fyne.CanvasObject) {
			item, err := data.GetItem(i)
			if err != nil {
				fyne.LogError(fmt.Sprintf("Error getting data item %d", i), err)
				return
			}
			updateItem(item, o)
		})

	data.AddListener(binding.NewDataListener(l.Refresh))
	return l
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (l *ReorderList) CreateRenderer() fyne.WidgetRenderer {
	l.ExtendBaseWidget(l)

	if f := l.CreateItem; f != nil && l.itemMin.IsZero() {
		item := createItemAndApplyThemeScope(f, l)

		l.itemMin = item.MinSize()
	}

	ll := newReorderListLayout(l)
	layout := &fyne.Container{Layout: ll}
	l.scroller = widget.NewVScroll(layout)
	layout.Resize(layout.MinSize())
	objects := []fyne.CanvasObject{l.scroller, &ll.(*reorderListLayout).dragSeparator}
	return newReorderListRenderer(objects, l, l.scroller, layout)
}

// FocusGained is called after this List has gained focus.
//
// Implements: fyne.Focusable
func (l *ReorderList) FocusGained() {
	l.focused = true
	l.scrollTo(l.currentFocus)
	l.RefreshItem(l.currentFocus)
}

// FocusLost is called after this List has lost focus.
//
// Implements: fyne.Focusable
func (l *ReorderList) FocusLost() {
	l.focused = false
	l.RefreshItem(l.currentFocus)
}

// MinSize returns the size that this widget should not shrink below.
func (l *ReorderList) MinSize() fyne.Size {
	l.ExtendBaseWidget(l)
	return l.BaseWidget.MinSize()
}

// RefreshItem refreshes a single item, specified by the item ID passed in.
//
// Since: 2.4
func (l *ReorderList) RefreshItem(id ReorderListItemID) {
	if l.scroller == nil {
		return
	}
	l.BaseWidget.Refresh()
	lo := l.scroller.Content.(*fyne.Container).Layout.(*reorderListLayout)
	item, ok := lo.searchVisible(lo.visible, id)
	if ok {
		lo.setupListItem(item, id, l.focused && l.currentFocus == id)
	}
}

// Returns the item that is currently bound to the given ID,
// or none of the ID is currently out of the visible range of the list.
//
// Since: Not a core Fyne list API
func (l *ReorderList) ItemForID(id ReorderListItemID) fyne.CanvasObject {
	lo := l.scroller.Content.(*fyne.Container).Layout.(*reorderListLayout)
	item, ok := lo.searchVisible(lo.visible, id)
	if ok {
		return item.child
	}
	return nil
}

// SetItemHeight supports changing the height of the specified list item. Items normally take the height of the template
// returned from the CreateItem callback. The height parameter uses the same units as a fyne.Size type and refers
// to the internal content height not including the divider size.
//
// Since: 2.3
func (l *ReorderList) SetItemHeight(id ReorderListItemID, height float32) {
	if l.itemHeights == nil {
		l.itemHeights = make(map[ReorderListItemID]float32)
	}

	refresh := l.itemHeights[id] != height
	l.itemHeights[id] = height

	if refresh {
		l.RefreshItem(id)
	}
}

func (l *ReorderList) scrollTo(id ReorderListItemID) {
	if l.scroller == nil {
		return
	}

	separatorThickness := l.Theme().Size(theme.SizeNamePadding)
	y := float32(0)
	lastItemHeight := l.itemMin.Height
	if len(l.itemHeights) == 0 {
		y = (float32(id) * l.itemMin.Height) + (float32(id) * separatorThickness)
	} else {
		i := 0
		for ; i < id; i++ {
			height := l.itemMin.Height
			if h, ok := l.itemHeights[i]; ok {
				height = h
			}

			y += height + separatorThickness
		}
		lastItemHeight = l.itemMin.Height
		if h, ok := l.itemHeights[i]; ok {
			lastItemHeight = h
		}
	}

	if y < l.scroller.Offset.Y {
		l.scroller.Offset.Y = y
	} else if y+l.itemMin.Height > l.scroller.Offset.Y+l.scroller.Size().Height {
		l.scroller.Offset.Y = y + lastItemHeight - l.scroller.Size().Height
	}
	l.offsetUpdated(l.scroller.Offset)
}

// Resize is called when this list should change size. We refresh to ensure invisible items are drawn.
func (l *ReorderList) Resize(s fyne.Size) {
	l.BaseWidget.Resize(s)
	if l.scroller == nil {
		return
	}

	l.offsetUpdated(l.scroller.Offset)
	l.scroller.Content.(*fyne.Container).Layout.(*reorderListLayout).updateList(true)
}

// Select add the item identified by the given ID to the selection.
func (l *ReorderList) Select(id ReorderListItemID) {
	if len(l.selected) > 0 && id == l.selected[0] {
		return
	}
	length := 0
	if f := l.Length; f != nil {
		length = f()
	}
	if id < 0 || id >= length {
		return
	}
	old := l.selected
	l.selected = []ReorderListItemID{id}
	defer func() {
		if f := l.OnUnselected; f != nil && len(old) > 0 {
			f(old[0])
		}
		if f := l.OnSelected; f != nil {
			f(id)
		}
	}()
	l.scrollTo(id)
	l.Refresh()
}

// ScrollTo scrolls to the item represented by id
//
// Since: 2.1
func (l *ReorderList) ScrollTo(id ReorderListItemID) {
	length := 0
	if f := l.Length; f != nil {
		length = f()
	}
	if id < 0 || id >= length {
		return
	}
	l.scrollTo(id)
	l.Refresh()
}

// ScrollToBottom scrolls to the end of the list
//
// Since: 2.1
func (l *ReorderList) ScrollToBottom() {
	l.scroller.ScrollToBottom()
	l.offsetUpdated(l.scroller.Offset)
}

// ScrollToTop scrolls to the start of the list
//
// Since: 2.1
func (l *ReorderList) ScrollToTop() {
	l.scroller.ScrollToTop()
	l.offsetUpdated(l.scroller.Offset)
}

// ScrollToOffset scrolls the list to the given offset position.
//
// Since: 2.5
func (l *ReorderList) ScrollToOffset(offset float32) {
	if l.scroller == nil {
		return
	}
	if offset < 0 {
		offset = 0
	}
	contentHeight := l.contentMinSize().Height
	if l.Size().Height >= contentHeight {
		return // content fully visible - no need to scroll
	}
	if offset > contentHeight {
		offset = contentHeight
	}
	l.scroller.ScrollToOffset(fyne.NewPos(0, offset))
	l.offsetUpdated(l.scroller.Offset)
}

// GetScrollOffset returns the current scroll offset position
//
// Since: 2.5
func (l *ReorderList) GetScrollOffset() float32 {
	return l.offsetY
}

// TypedKey is called if a key event happens while this List is focused.
//
// Implements: fyne.Focusable
func (l *ReorderList) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeySpace:
		l.Select(l.currentFocus)
	case fyne.KeyDown:
		if f := l.Length; f != nil && l.currentFocus >= f()-1 {
			return
		}
		l.RefreshItem(l.currentFocus)
		l.currentFocus++
		l.scrollTo(l.currentFocus)
		l.RefreshItem(l.currentFocus)
	case fyne.KeyUp:
		if l.currentFocus <= 0 {
			return
		}
		l.RefreshItem(l.currentFocus)
		l.currentFocus--
		l.scrollTo(l.currentFocus)
		l.RefreshItem(l.currentFocus)
	}
}

// TypedRune is called if a text event happens while this List is focused.
//
// Implements: fyne.Focusable
func (l *ReorderList) TypedRune(_ rune) {
	// intentionally left blank
}

// Unselect removes the item identified by the given ID from the selection.
func (l *ReorderList) Unselect(id ReorderListItemID) {
	if len(l.selected) == 0 || l.selected[0] != id {
		return
	}

	l.selected = nil
	l.Refresh()
	if f := l.OnUnselected; f != nil {
		f(id)
	}
}

// UnselectAll removes all items from the selection.
//
// Since: 2.1
func (l *ReorderList) UnselectAll() {
	if len(l.selected) == 0 {
		return
	}

	selected := l.selected
	l.selected = nil
	l.Refresh()
	if f := l.OnUnselected; f != nil {
		for _, id := range selected {
			f(id)
		}
	}
}

func (l *ReorderList) contentMinSize() fyne.Size {
	separatorThickness := l.Theme().Size(theme.SizeNamePadding)
	if l.Length == nil {
		return fyne.NewSize(0, 0)
	}
	items := l.Length()

	if len(l.itemHeights) == 0 {
		return fyne.NewSize(l.itemMin.Width,
			(l.itemMin.Height+separatorThickness)*float32(items)-separatorThickness)
	}

	height := float32(0)
	totalCustom := 0
	templateHeight := l.itemMin.Height
	for id, itemHeight := range l.itemHeights {
		if id < items {
			totalCustom++
			height += itemHeight
		}
	}
	height += float32(items-totalCustom) * templateHeight

	return fyne.NewSize(l.itemMin.Width, height+separatorThickness*float32(items-1))
}

func (l *reorderListLayout) calculateDragSeparatorY(thickness float32) float32 {
	if l.list.scroller.Size().Height <= 0 {
		return 0
	}

	relY := l.dragRelativeY
	if relY < 0 {
		relY = 0
	} else if h := l.list.Size().Height; relY > h {
		relY = h
	}

	numItems := 0.0
	if l.list.Length != nil {
		numItems = float64(l.list.Length())
	}
	if len(l.list.itemHeights) == 0 {
		padding := theme.Padding()
		paddedItemHeight := l.list.itemMin.Height + padding
		beforeItem := math.Round(float64(relY+l.list.offsetY) / float64(paddedItemHeight))
		if beforeItem > numItems {
			beforeItem = numItems
		}
		y := float32(beforeItem)*paddedItemHeight - padding/2 - thickness
		l.dragInsertAt = ReorderListItemID(beforeItem)
		return y
	}
	// TODO: support item heights
	return 0
}

// fills l.visibleRowHeights and also returns offY and minRow
func (l *reorderListLayout) calculateVisibleRowHeights(itemHeight float32, length int, th fyne.Theme) (offY float32, minRow int) {
	rowOffset := float32(0)
	isVisible := false
	l.visibleRowHeights = l.visibleRowHeights[:0]

	if l.list.scroller.Size().Height <= 0 {
		return
	}

	padding := th.Size(theme.SizeNamePadding)

	if len(l.list.itemHeights) == 0 {
		paddedItemHeight := itemHeight + padding

		offY = float32(math.Floor(float64(l.list.offsetY/paddedItemHeight))) * paddedItemHeight
		minRow = int(math.Floor(float64(offY / paddedItemHeight)))
		maxRow := int(math.Ceil(float64((offY + l.list.scroller.Size().Height) / paddedItemHeight)))

		if minRow > length-1 {
			minRow = length - 1
		}
		if minRow < 0 {
			minRow = 0
			offY = 0
		}

		if maxRow > length-1 {
			maxRow = length - 1
		}

		for i := 0; i <= maxRow-minRow; i++ {
			l.visibleRowHeights = append(l.visibleRowHeights, itemHeight)
		}
		return
	}

	for i := 0; i < length; i++ {
		height := itemHeight
		if h, ok := l.list.itemHeights[i]; ok {
			height = h
		}

		if rowOffset <= l.list.offsetY-height-padding {
			// before scroll
		} else if rowOffset <= l.list.offsetY {
			minRow = i
			offY = rowOffset
			isVisible = true
		}
		if rowOffset >= l.list.offsetY+l.list.scroller.Size().Height {
			break
		}

		rowOffset += height + padding
		if isVisible {
			l.visibleRowHeights = append(l.visibleRowHeights, height)
		}
	}
	return
}

const (
	// max speed (in units per frame) that the list will scroll when dragging above or below
	maxScrollSpeed = 500
	minScrollSpeed = 3
	// how far to drag above or below the top/bottom of the list to reach the max scroll speed
	scrollAccelerateRange = 250
)

func (l *reorderListLayout) onRowDragged(id ReorderListItemID, e *fyne.DragEvent) {
	if !l.list.EnableDragging {
		return
	}
	startedDrag := false
	if l.draggingRow < 0 /*no drag in progress*/ {
		l.draggingRow = id
		startedDrag = true
	}

	listPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(l.list.scroller)
	// TODO: this may break if the list itself is positioned outside the window viewport?
	// don't worry about it now
	l.dragRelativeY = e.AbsolutePosition.Y - listPos.Y

	animationSpeedCurve := func(x float32) float32 {
		// scale to domain: x_: [0, 1]
		x_ := math.Min(math.Abs(float64(x)), scrollAccelerateRange) / scrollAccelerateRange
		// quadratic, modified by minScrollSpeed
		return float32(math.Max(x_*x_*maxScrollSpeed, minScrollSpeed))
	}

	// distance from top or bottom of list that starts to trigger scrolling animation
	scrollStartThreshold := l.list.itemMin.Height / 2

	if topThresh := l.dragRelativeY - scrollStartThreshold; topThresh < 0 {
		l.scrollAnimSpeed = -animationSpeedCurve(topThresh)
		l.ensureStartDragAnim()
	} else if bottmThresh := l.list.Size().Height - scrollStartThreshold; l.dragRelativeY > bottmThresh {
		l.scrollAnimSpeed = animationSpeedCurve(l.dragRelativeY - bottmThresh)
		l.ensureStartDragAnim()
	} else {
		l.ensureStopDragAnim()
	}

	l.updateDragSeparator()
	if startedDrag && l.list.OnDragBegin != nil {
		l.list.OnDragBegin(l.draggingRow)
	}
}

func (l *reorderListLayout) onDragEnd() {
	startRow := l.draggingRow
	l.ensureStopDragAnim()
	l.draggingRow = -1
	l.dragSeparator.Hide()
	if l.list.OnDragEnd != nil {
		l.list.OnDragEnd(startRow, l.dragInsertAt)
	}
}

func (l *reorderListLayout) ensureStartDragAnim() {
	if l.dragScrollAnim == nil {
		l.dragScrollAnim = fyne.NewAnimation(math.MaxInt64 /*until stopped*/, func(_ float32) {
			l.list.scroller.Scrolled(&fyne.ScrollEvent{Scrolled: fyne.Delta{DY: -l.scrollAnimSpeed}})
		})
		l.dragScrollAnim.Start()
	}
}

func (l *reorderListLayout) ensureStopDragAnim() {
	if l.dragScrollAnim != nil {
		l.dragScrollAnim.Stop()
		l.dragScrollAnim = nil
	}
}

type reorderlistRenderer struct {
	objects  []fyne.CanvasObject
	list     *ReorderList
	scroller *widget.Scroll
	layout   *fyne.Container
}

func newReorderListRenderer(objects []fyne.CanvasObject, l *ReorderList, scroller *widget.Scroll, layout *fyne.Container) *reorderlistRenderer {
	lr := &reorderlistRenderer{objects: objects, list: l, scroller: scroller, layout: layout}
	lr.scroller.OnScrolled = l.offsetUpdated
	return lr
}

func (l *reorderlistRenderer) Layout(size fyne.Size) {
	l.scroller.Resize(size)
}

func (l *reorderlistRenderer) MinSize() fyne.Size {
	return l.scroller.MinSize().Max(l.list.itemMin)
}

func (l *reorderlistRenderer) Refresh() {
	if f := l.list.CreateItem; f != nil {
		item := createItemAndApplyThemeScope(f, l.list)
		l.list.itemMin = item.MinSize()
	}
	l.Layout(l.list.Size())
	l.scroller.Refresh()
	layout := l.layout.Layout.(*reorderListLayout)
	layout.dragSeparator.FillColor = theme.ForegroundColor()
	layout.dragSeparator.Refresh()
	layout.updateList(false)

	for _, s := range layout.separators {
		s.Refresh()
	}
	canvas.Refresh(l.list)
}

func (l *reorderlistRenderer) Objects() []fyne.CanvasObject {
	return l.objects
}

func (l *reorderlistRenderer) Destroy() {}

type reorderListItem struct {
	BaseWidget

	id                ReorderListItemID
	onTapped          func()
	background        *canvas.Rectangle
	listLayout        *reorderListLayout
	child             fyne.CanvasObject
	hovered, selected bool
}

func newReorderListItem(child fyne.CanvasObject, listLayout *reorderListLayout, tapped func()) *reorderListItem {
	li := &reorderListItem{
		listLayout: listLayout,
		child:      child,
		onTapped:   tapped,
	}

	li.ExtendBaseWidget(li)
	return li
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (li *reorderListItem) CreateRenderer() fyne.WidgetRenderer {
	li.ExtendBaseWidget(li)
	th := li.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	li.background = canvas.NewRectangle(th.Color(theme.ColorNameHover, v))
	li.background.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
	li.background.Hide()

	container := &fyne.Container{Layout: layout.NewStackLayout(), Objects: []fyne.CanvasObject{li.background, li.child}}

	return NewSimpleRenderer(container)
}

// MinSize returns the size that this widget should not shrink below.
func (li *reorderListItem) MinSize() fyne.Size {
	li.ExtendBaseWidget(li)
	return li.BaseWidget.MinSize()
}

// MouseIn is called when a desktop pointer enters the widget.
func (li *reorderListItem) MouseIn(*desktop.MouseEvent) {
	if li.listLayout.draggingRow >= 0 {
		return
	}
	li.hovered = true
	li.Refresh()
}

// MouseMoved is called when a desktop pointer hovers over the widget.
func (li *reorderListItem) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut is called when a desktop pointer exits the widget.
func (li *reorderListItem) MouseOut() {
	li.hovered = false
	li.Refresh()
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler.
func (li *reorderListItem) Tapped(*fyne.PointEvent) {
	if li.onTapped != nil {
		li.selected = true
		li.Refresh()
		li.onTapped()
	}
}

func (li *reorderListItem) Refresh() {
	th := li.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	li.background.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
	if li.selected {
		li.background.FillColor = th.Color(theme.ColorNameSelection, v)
		li.background.Show()
	} else if li.hovered {
		li.background.FillColor = th.Color(theme.ColorNameHover, v)
		li.background.Show()
	} else {
		li.background.Hide()
	}
	li.background.Refresh()
	canvas.Refresh(li)
}

func (li *reorderListItem) Dragged(e *fyne.DragEvent) {
	li.listLayout.onRowDragged(li.id, e)
}

func (li *reorderListItem) DragEnd() {
	li.listLayout.onDragEnd()
}

type reorderListItemAndID struct {
	item *reorderListItem
	id   ReorderListItemID
}

// thickness: theme.SeparatorThicknessSize() * dragSeparatorThicknessMultiplier
const dragSeparatorThicknessMultiplier = 1.5

type reorderListLayout struct {
	list          *ReorderList
	separators    []fyne.CanvasObject
	children      []fyne.CanvasObject
	dragSeparator canvas.Rectangle

	itemPool          sync.Pool
	visible           []reorderListItemAndID
	wasVisible        []reorderListItemAndID
	visibleRowHeights []float32

	draggingRow     ReorderListItemID // -1 if no drag
	dragRelativeY   float32           // 0 == top of list widget
	dragInsertAt    ReorderListItemID
	dragScrollAnim  *fyne.Animation
	scrollAnimSpeed float32
}

func newReorderListLayout(list *ReorderList) fyne.Layout {
	l := &reorderListLayout{list: list, draggingRow: -1}
	l.dragSeparator.FillColor = theme.ForegroundColor()
	list.offsetUpdated = l.offsetUpdated
	return l
}

func (l *reorderListLayout) Layout([]fyne.CanvasObject, fyne.Size) {
	l.updateList(true)
}

func (l *reorderListLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return l.list.contentMinSize()
}

func (l *reorderListLayout) getItem() *reorderListItem {
	item := l.itemPool.Get()
	if item == nil {
		if f := l.list.CreateItem; f != nil {
			item2 := createItemAndApplyThemeScope(f, l.list)
			item = newReorderListItem(item2, l, nil)
		}
	}
	return item.(*reorderListItem)
}

func (l *reorderListLayout) offsetUpdated(pos fyne.Position) {
	if l.list.offsetY == pos.Y {
		return
	}
	l.list.offsetY = pos.Y
	if l.draggingRow >= 0 {
		l.updateDragSeparator()
	}
	l.updateList(true)
}

func (l *reorderListLayout) setupListItem(li *reorderListItem, id ReorderListItemID, focus bool) {
	li.id = id
	previousIndicator := li.selected
	li.selected = false
	for _, s := range l.list.selected {
		if id == s {
			li.selected = true
			break
		}
	}
	if focus {
		li.hovered = true
		li.Refresh()
	} else if previousIndicator != li.selected || li.hovered {
		li.hovered = false
		li.Refresh()
	}
	if f := l.list.UpdateItem; f != nil {
		f(id, li.child)
	}
	li.onTapped = func() {
		if !fyne.CurrentDevice().IsMobile() {
			canvas := fyne.CurrentApp().Driver().CanvasForObject(l.list)
			if canvas != nil {
				canvas.Focus(l.list)
			}

			l.list.currentFocus = id
		}

		l.list.Select(id)
	}
}

func (l *reorderListLayout) updateList(newOnly bool) {
	th := l.list.Theme()
	separatorThickness := th.Size(theme.SizeNamePadding)
	width := l.list.Size().Width
	length := 0
	if f := l.list.Length; f != nil {
		length = f()
	}
	if l.list.UpdateItem == nil {
		fyne.LogError("Missing UpdateCell callback required for List", nil)
	}

	// l.wasVisible now represents the currently visible items, while
	// l.visible will be updated to represent what is visible *after* the update
	l.wasVisible = append(l.wasVisible, l.visible...)
	l.visible = l.visible[:0]

	offY, minRow := l.calculateVisibleRowHeights(l.list.itemMin.Height, length, th)
	if len(l.visibleRowHeights) == 0 && length > 0 { // we can't show anything until we have some dimensions
		return
	}

	oldChildrenLen := len(l.children)
	l.children = l.children[:0]

	y := offY
	for index, itemHeight := range l.visibleRowHeights {
		row := index + minRow
		size := fyne.NewSize(width, itemHeight)

		c, ok := l.searchVisible(l.wasVisible, row)
		if !ok {
			c = l.getItem()
			if c == nil {
				continue
			}
			c.Resize(size)
		}

		c.Move(fyne.NewPos(0, y))
		c.Resize(size)

		y += itemHeight + separatorThickness
		l.visible = append(l.visible, reorderListItemAndID{id: row, item: c})
		l.children = append(l.children, c)
	}
	l.nilOldSliceData(l.children, len(l.children), oldChildrenLen)

	for _, wasVis := range l.wasVisible {
		if _, ok := l.searchVisible(l.visible, wasVis.id); !ok {
			l.itemPool.Put(wasVis.item)
		}
	}

	l.updateSeparators()

	c := l.list.scroller.Content.(*fyne.Container)
	oldObjLen := len(c.Objects)
	c.Objects = c.Objects[:0]
	c.Objects = append(c.Objects, l.children...)
	c.Objects = append(c.Objects, l.separators...)
	l.nilOldSliceData(c.Objects, len(c.Objects), oldObjLen)

	if newOnly {
		for _, vis := range l.visible {
			if _, ok := l.searchVisible(l.wasVisible, vis.id); !ok {
				l.setupListItem(vis.item, vis.id, l.list.focused && l.list.currentFocus == vis.id)
			}
		}
	} else {
		for _, vis := range l.visible {
			l.setupListItem(vis.item, vis.id, l.list.focused && l.list.currentFocus == vis.id)
		}
	}

	// we don't need wasVisible now until next call to update
	// nil out all references before truncating slice
	for i := 0; i < len(l.wasVisible); i++ {
		l.wasVisible[i].item = nil
	}
	l.wasVisible = l.wasVisible[:0]
}

func (l *reorderListLayout) updateDragSeparator() {
	listSize := l.list.Size()
	thickness := theme.SeparatorThicknessSize() * dragSeparatorThicknessMultiplier
	l.dragSeparator.Resize(fyne.NewSize(listSize.Width, thickness))
	sepY := l.calculateDragSeparatorY(thickness) - l.list.offsetY
	padding := theme.Padding()
	if sepY > listSize.Height+padding || sepY < -padding {
		// use margin of [-padding, padding] make sure
		// it can be shown above/below first and last items
		l.dragSeparator.Hide()
		return
	}
	l.dragSeparator.Move(fyne.NewPos(0, sepY))
	l.dragSeparator.Show()
}

func (l *reorderListLayout) updateSeparators() {
	if l.draggingRow >= 0 {
		l.updateDragSeparator()
	}
	if l.list.HideSeparators {
		l.separators = nil
		return
	}
	if lenChildren := len(l.children); lenChildren > 1 {
		if lenSep := len(l.separators); lenSep > lenChildren {
			l.separators = l.separators[:lenChildren]
		} else {
			for i := lenSep; i < lenChildren; i++ {

				sep := NewSeparator()
				//if cache.OverrideThemeMatchingScope(sep, l.list) {
				sep.Refresh()
				//}

				l.separators = append(l.separators, sep)
			}
		}
	} else {
		l.separators = nil
	}

	th := l.list.Theme()
	separatorThickness := th.Size(theme.SizeNameSeparatorThickness)
	dividerOff := (th.Size(theme.SizeNamePadding) + separatorThickness) / 2
	for i, child := range l.children {
		if i == 0 {
			continue
		}
		l.separators[i].Move(fyne.NewPos(0, child.Position().Y-dividerOff))
		l.separators[i].Resize(fyne.NewSize(l.list.Size().Width, separatorThickness))
		l.separators[i].Show()
	}
}

// invariant: visible is in ascending order of IDs
func (l *reorderListLayout) searchVisible(visible []reorderListItemAndID, id ReorderListItemID) (*reorderListItem, bool) {
	ln := len(visible)
	idx := sort.Search(ln, func(i int) bool { return visible[i].id >= id })
	if idx < ln && visible[idx].id == id {
		return visible[idx].item, true
	}
	return nil, false
}

func (l *reorderListLayout) nilOldSliceData(objs []fyne.CanvasObject, len, oldLen int) {
	if oldLen > len {
		objs = objs[:oldLen] // gain view into old data
		for i := len; i < oldLen; i++ {
			objs[i] = nil
		}
	}
}

func createReorderItemAndApplyThemeScope(f func() fyne.CanvasObject, _ fyne.Widget) fyne.CanvasObject {
	item := f()
	//if !cache.OverrideThemeMatchingScope(item, scope) {
	//return item
	//}

	item.Refresh()
	return item
}
