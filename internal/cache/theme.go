package cache

import (
	"strconv"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
)

type overridesMap struct {
	m  map[fyne.CanvasObject]*overrideScope
	mu sync.RWMutex
}

var (
	overrides     = overridesMap{m: make(map[fyne.CanvasObject]*overrideScope)}
	overrideCount = atomic.Uint32{}
)

type overrideScope struct {
	th      fyne.Theme
	cacheID string
}

// OverrideTheme allows an app to specify that a single object should use a different theme to the app.
// This should be used sparingly to avoid a jarring user experience.
// If the object is a container it will theme the children, if it is a canvas primitive it will do nothing.
//
// Since: 2.5
func OverrideTheme(o fyne.CanvasObject, th fyne.Theme) {
	id := overrideCount.Add(1)
	s := &overrideScope{th: th, cacheID: strconv.Itoa(int(id))}
	overrideTheme(o, s)
}

func OverrideThemeMatchingScope(o, parent fyne.CanvasObject) bool {
	overrides.mu.RLock()
	scope, has := overrides.m[parent]
	overrides.mu.RUnlock()
	if !has { // not overridden in parent
		return false
	}

	overrideTheme(o, scope)
	return true
}

func WidgetScopeID(o fyne.CanvasObject) string {
	overrides.mu.RLock()
	scope, has := overrides.m[o]
	overrides.mu.RUnlock()
	if !has {
		return ""
	}

	return scope.cacheID
}

func WidgetTheme(o fyne.CanvasObject) fyne.Theme {
	if o == nil {
		return nil
	}

	overrides.mu.RLock()
	scope, has := overrides.m[o]
	overrides.mu.RUnlock()
	if !has { // not overridden in parent
		return nil
	}
	return scope.th
}

func overrideContainer(c *fyne.Container, s *overrideScope) {
	for _, o := range c.Objects {
		overrideTheme(o, s)
	}
}

func overrideTheme(o fyne.CanvasObject, s *overrideScope) {
	switch c := o.(type) {
	case fyne.Widget:
		overrideWidget(c, s)
	case *fyne.Container:
		overrideContainer(c, s)
	default:
		overrides.mu.Lock()
		overrides.m[c] = s
		overrides.mu.Unlock()
	}
}

func overrideWidget(w fyne.Widget, s *overrideScope) {
	ResetThemeCaches()

	overrides.mu.Lock()
	overrides.m[w] = s
	overrides.mu.Unlock()

	r := Renderer(w)
	if r == nil {
		return
	}

	for _, o := range r.Objects() {
		overrideTheme(o, s)
	}
}
