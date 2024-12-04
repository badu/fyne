package binding

// DataList is the base interface for all bindable data lists.
//
// Since: 2.0
type DataList interface {
	// AddListener attaches a new change listener to this DataItem.
	// Listeners are called each time the data inside this DataItem changes.
	// Additionally the listener will be triggered upon successful connection to get the current value.
	AddListener(DataListener)
	// RemoveListener will detach the specified change listener from the DataItem.
	// Disconnected listener will no longer be triggered when changes occur.
	RemoveListener(DataListener)
	GetItem(index int) (DataItem, error)
	Length() int
}

type listBase struct {
	base
	items []DataItem
}

// GetItem returns the DataItem at the specified index.
func (b *listBase) GetItem(i int) (DataItem, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if i < 0 || i >= len(b.items) {
		return nil, errOutOfBounds
	}

	return b.items[i], nil
}

// Length returns the number of items in this data list.
func (b *listBase) Length() int {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return len(b.items)
}

func (b *listBase) appendItem(i DataItem) {
	b.items = append(b.items, i)
}

func (b *listBase) deleteItem(i int) {
	b.items = append(b.items[:i], b.items[i+1:]...)
}
