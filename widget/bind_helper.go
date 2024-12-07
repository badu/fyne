package widget

import (
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2/data/binding"
)

// basicBinder stores a DataItem and a function to be called when it changes.
// It provides a convenient way to replace data and callback independently.
type basicBinder struct {
	callback atomic.Pointer[func(binding.DataItem)]

	dataListenerPairLock sync.RWMutex
	dataListenerPair     annotatedListener // access guarded by dataListenerPairLock
}

// Bind replaces the data item whose changes are tracked by the callback function.
func (b *basicBinder) Bind(data binding.DataItem) {
	listener := binding.NewDataListener(func() { // NB: listener captures `data` but always calls the up-to-date callback
		f := b.callback.Load()
		if f == nil || *f == nil {
			return
		}

		(*f)(data)
	})
	data.AddListener(listener)
	listenerInfo := annotatedListener{
		data:     data,
		listener: listener,
	}

	b.dataListenerPairLock.Lock()
	b.unbindLocked()
	b.dataListenerPair = listenerInfo
	b.dataListenerPairLock.Unlock()
}

// CallWithData passes the currently bound data item as an argument to the
// provided function.
func (b *basicBinder) CallWithData(f func(data binding.DataItem)) {
	b.dataListenerPairLock.RLock()
	data := b.dataListenerPair.data
	b.dataListenerPairLock.RUnlock()
	f(data)
}

// SetCallback replaces the function to be called when the data changes.
func (b *basicBinder) SetCallback(f func(data binding.DataItem)) {
	b.callback.Store(&f)
}

// Unbind requests the callback to be no longer called when the previously bound
// data item changes.
func (b *basicBinder) Unbind() {
	b.dataListenerPairLock.Lock()
	b.unbindLocked()
	b.dataListenerPairLock.Unlock()
}

// unbindLocked expects the caller to hold dataListenerPairLock.
func (b *basicBinder) unbindLocked() {
	previousListener := b.dataListenerPair
	b.dataListenerPair = annotatedListener{nil, nil}

	if previousListener.listener == nil || previousListener.data == nil {
		return
	}
	previousListener.data.RemoveListener(previousListener.listener)
}

type annotatedListener struct {
	data     binding.DataItem
	listener binding.DataListener
}
