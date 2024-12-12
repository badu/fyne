package binding

// DataTreeRootID const is the value used as ID for the root of any tree binding.
const DataTreeRootID = ""

// DataTree is the base interface for all bindable data trees.
//
// Since: 2.4
type DataTree interface {
	// AddListener attaches a new change listener to this DataItem.
	// Listeners are called each time the data inside this DataItem changes.
	// Additionally the listener will be triggered upon successful connection to get the current value.
	AddListener(DataListener)
	// RemoveListener will detach the specified change listener from the DataItem.
	// Disconnected listener will no longer be triggered when changes occur.
	RemoveListener(DataListener)
	GetItem(id string) (DataItem, error)
	ChildIDs(string) []string
}

type treeBase struct {
	base

	ids   map[string][]string
	items map[string]DataItem
}

// GetItem returns the DataItem at the specified id.
func (t *treeBase) GetItem(id string) (DataItem, error) {
	t.propertiesLock.RLock()
	defer t.propertiesLock.RUnlock()

	if item, ok := t.items[id]; ok {
		return item, nil
	}

	return nil, errOutOfBounds
}

// ChildIDs returns the ordered IDs of items in this data tree that are children of the specified ID.
func (t *treeBase) ChildIDs(id string) []string {
	t.propertiesLock.RLock()
	defer t.propertiesLock.RUnlock()

	if ids, ok := t.ids[id]; ok {
		return ids
	}

	return []string{}
}

func (t *treeBase) appendItem(i DataItem, id, parent string) {
	t.items[id] = i
	ids, ok := t.ids[parent]
	if !ok {
		ids = make([]string, 0)
	}

	for _, in := range ids {
		if in == id {
			return
		}
	}
	t.ids[parent] = append(ids, id)
}

func (t *treeBase) deleteItem(id, parent string) {
	delete(t.items, id)

	ids, ok := t.ids[parent]
	if !ok {
		return
	}

	off := -1
	for i, id2 := range ids {
		if id2 == id {
			off = i
			break
		}
	}
	if off == -1 {
		return
	}
	t.ids[parent] = append(ids[:off], ids[off+1:]...)
}

func parentIDFor(id string, ids map[string][]string) string {
	for parent, list := range ids {
		for _, child := range list {
			if child == id {
				return parent
			}
		}
	}

	return ""
}
