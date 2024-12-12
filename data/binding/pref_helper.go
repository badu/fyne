package binding

import (
	"sync"

	"fyne.io/fyne/v2"
)

type preferenceItem interface {
	checkForChange()
}

type providerReplacer interface {
	replaceProvider(fyne.Preferences)
}

type preferenceBindings struct {
	//mu    sync.RWMutex
	items map[string]preferenceItem
}

func (b *preferenceBindings) getItem(key string) preferenceItem {
	//b.mu.RLock()
	val, has := b.items[key]
	if !has {
		return nil
	}
	//b.mu.RUnlock()
	return val
}

func (b *preferenceBindings) list() []preferenceItem {
	//b.mu.Lock()
	ret := make([]preferenceItem, 0, len(b.items))
	for _, val := range b.items {
		ret = append(ret, val)
	}
	//b.mu.Unlock()
	return ret
}

func (b *preferenceBindings) setItem(key string, item preferenceItem) {
	//b.mu.Lock()
	b.items[key] = item
	//b.mu.Unlock()
}

type preferencesMap struct {
	mu    sync.RWMutex
	prefs map[fyne.Preferences]*preferenceBindings

	appPrefs fyne.Preferences // the main application prefs, to check if it changed...
}

func newPreferencesMap() *preferencesMap {
	return &preferencesMap{prefs: make(map[fyne.Preferences]*preferenceBindings)}
}

func (m *preferencesMap) ensurePreferencesAttached(p fyne.Preferences) *preferenceBindings {
	m.mu.RLock()
	binds, has := m.prefs[p]
	m.mu.RUnlock()
	if has {
		return binds
	}

	binds = &preferenceBindings{items: make(map[string]preferenceItem)}

	m.mu.Lock()
	m.prefs[p] = binds
	m.mu.Unlock()

	p.AddChangeListener(func() { m.preferencesChanged(fyne.CurrentApp().Preferences()) })
	return binds
}

func (m *preferencesMap) getBindings(p fyne.Preferences) *preferenceBindings {
	if p == fyne.CurrentApp().Preferences() {
		if m.appPrefs == nil {
			m.appPrefs = p
		} else if m.appPrefs != p {
			m.migratePreferences(m.appPrefs, p)
		}
	}

	m.mu.RLock()
	binds, has := m.prefs[p]
	m.mu.RUnlock()

	if !has {
		return nil
	}
	return binds
}

func (m *preferencesMap) preferencesChanged(p fyne.Preferences) {
	binds := m.getBindings(p)
	if binds == nil {
		return
	}
	for _, item := range binds.list() {
		item.checkForChange()
	}
}

func (m *preferencesMap) migratePreferences(src, dst fyne.Preferences) {
	m.mu.RLock()
	old, hasOld := m.prefs[src]
	m.mu.RUnlock()
	if !hasOld {
		return
	}

	m.mu.Lock()
	m.prefs[dst] = old
	delete(m.prefs, src)
	m.mu.Unlock()

	m.appPrefs = dst

	binds := m.getBindings(dst)
	if binds == nil {
		return
	}
	for _, b := range binds.list() {
		if backed, ok := b.(providerReplacer); ok {
			backed.replaceProvider(dst)
		}
	}

	m.preferencesChanged(dst)
}
