//go:generate go run gen.go

// Package binding provides support for binding data to widgets.
package binding

import (
	"errors"
	"reflect"
	"sync"

	"fyne.io/fyne/v2"
)

var (
	errKeyNotFound = errors.New("key not found")
	errOutOfBounds = errors.New("index out of bounds")
	errParseFailed = errors.New("format did not match 1 value")

	// As an optimisation we connect any listeners asking for the same key, so that there is only 1 per preference item.
	prefBinds = newPreferencesMap()
)

// DataItem is the base interface for all bindable data items.
//
// Since: 2.0
type DataItem interface {
	// AddListener attaches a new change listener to this DataItem.
	// Listeners are called each time the data inside this DataItem changes.
	// Additionally the listener will be triggered upon successful connection to get the current value.
	AddListener(DataListener)
	// RemoveListener will detach the specified change listener from the DataItem.
	// Disconnected listener will no longer be triggered when changes occur.
	RemoveListener(DataListener)
}

// DataListener is any object that can register for changes in a bindable DataItem.
// See NewDataListener to define a new listener using just an inline function.
//
// Since: 2.0
type DataListener interface {
	DataChanged()
}

// NewDataListener is a helper function that creates a new listener type from a simple callback function.
//
// Since: 2.0
func NewDataListener(fn func()) DataListener {
	return &listener{fn}
}

type listener struct {
	callback func()
}

func (l *listener) DataChanged() {
	l.callback()
}

type base struct {
	listeners      map[DataListener]struct{}
	propertiesLock sync.RWMutex
}

// AddListener allows a data listener to be informed of changes to this item.
func (b *base) AddListener(l DataListener) {
	if b.listeners == nil {
		b.listeners = make(map[DataListener]struct{})
	}

	b.propertiesLock.Lock()
	b.listeners[l] = struct{}{}
	b.propertiesLock.Unlock()

	queueItem(l.DataChanged)
}

// RemoveListener should be called if the listener is no longer interested in being informed of data change events.
func (b *base) RemoveListener(l DataListener) {
	b.propertiesLock.Lock()
	_, has := b.listeners[l]
	if has {
		delete(b.listeners, l)
	}
	b.propertiesLock.Unlock()
}

func (b *base) trigger() {
	for key := range b.listeners {
		queueItem(key.DataChanged)
	}
}

// Untyped supports binding a any value.
//
// Since: 2.1
type Untyped interface {
	// AddListener attaches a new change listener to this DataItem.
	// Listeners are called each time the data inside this DataItem changes.
	// Additionally the listener will be triggered upon successful connection to get the current value.
	AddListener(DataListener)
	// RemoveListener will detach the specified change listener from the DataItem.
	// Disconnected listener will no longer be triggered when changes occur.
	RemoveListener(DataListener)
	Get() (any, error)
	Set(any) error
}

// NewUntyped returns a bindable any value that is managed internally.
//
// Since: 2.1
func NewUntyped() Untyped {
	var blank any = nil
	v := &blank
	return &boundUntyped{val: reflect.ValueOf(v).Elem()}
}

type boundUntyped struct {
	base

	val reflect.Value
}

func (b *boundUntyped) Get() (any, error) {
	b.propertiesLock.RLock()
	defer b.propertiesLock.RUnlock()

	return b.val.Interface(), nil
}

func (b *boundUntyped) Set(val any) error {
	b.propertiesLock.Lock()
	defer b.propertiesLock.Unlock()
	if b.val.Interface() == val {
		return nil
	}
	if val == nil {
		zeroValue := reflect.Zero(b.val.Type())
		b.val.Set(zeroValue)
	} else {
		b.val.Set(reflect.ValueOf(val))
	}

	b.trigger()
	return nil
}

// ExternalUntyped supports binding a any value to an external value.
//
// Since: 2.1
type ExternalUntyped interface {
	// AddListener attaches a new change listener to this DataItem.
	// Listeners are called each time the data inside this DataItem changes.
	// Additionally the listener will be triggered upon successful connection to get the current value.
	AddListener(DataListener)
	// RemoveListener will detach the specified change listener from the DataItem.
	// Disconnected listener will no longer be triggered when changes occur.
	RemoveListener(DataListener)
	Get() (any, error)
	Set(any) error
	Reload() error
}

// BindUntyped returns a bindable any value that is bound to an external type.
// The parameter must be a pointer to the type you wish to bind.
//
// Since: 2.1
func BindUntyped(v any) ExternalUntyped {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		fyne.LogError("Invalid type passed to BindUntyped, must be a pointer", nil)
		v = nil
	}

	if v == nil {
		var blank any
		v = &blank // never allow a nil value pointer
	}

	b := &boundExternalUntyped{}
	b.val = reflect.ValueOf(v).Elem()
	b.old = b.val.Interface()
	return b
}

type boundExternalUntyped struct {
	boundUntyped

	old any
}

func (b *boundExternalUntyped) Set(val any) error {
	b.propertiesLock.Lock()
	defer b.propertiesLock.Unlock()
	if b.old == val {
		return nil
	}
	b.val.Set(reflect.ValueOf(val))
	b.old = val

	b.trigger()
	return nil
}

func (b *boundExternalUntyped) Reload() error {
	return b.Set(b.val.Interface())
}
