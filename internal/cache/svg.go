package cache

import (
	"image"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

var (
	svgs     = make(map[string]*svgInfo)
	svgsLock sync.RWMutex
)

// GetSvg gets svg image from cache if it exists.
func GetSvg(name string, o fyne.CanvasObject, w, h int) *image.NRGBA {
	svgsLock.RLock()
	svginfo, has := svgs[overriddenName(name, o)]
	svgsLock.RUnlock()
	if !has {
		return nil
	}

	if svginfo.w != w || svginfo.h != h {
		return nil
	}

	svginfo.setAlive()
	return svginfo.pix
}

// SetSvg sets a svg into the cache map.
func SetSvg(name string, o fyne.CanvasObject, pix *image.NRGBA, w int, h int) {
	sinfo := &svgInfo{
		pix: pix,
		w:   w,
		h:   h,
	}
	sinfo.setAlive()

	svgsLock.Lock()
	svgs[overriddenName(name, o)] = sinfo
	svgsLock.Unlock()
}

type svgInfo struct {
	// An svgInfo can be accessed from different goroutines, e.g., systray.
	// Use expiringCache instead of expiringCacheNoLock.
	expiringCache
	pix  *image.NRGBA
	w, h int
}

// destroyExpiredSvgs destroys expired svgs cache data.
func destroyExpiredSvgs(now time.Time) {
	svgsLock.Lock()
	for key, sInfo := range svgs {
		if sInfo.isExpired(now) {
			delete(svgs, key)
		}
	}
	svgsLock.Unlock()
}

func overriddenName(name string, o fyne.CanvasObject) string {
	if o == nil { // for overridden themes get the cache key right
		return name
	}

	overrides.mu.RLock()
	scope, has := overrides.m[o]
	overrides.mu.RUnlock()
	if has { // not overridden in parent
		return scope.cacheID + name
	}
	return name
}
