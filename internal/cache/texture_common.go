package cache

import (
	"sync"

	"fyne.io/fyne/v2"
)

var (
	textures     = make(map[fyne.CanvasObject]*textureInfo)
	texturesLock sync.RWMutex
)

// DeleteTexture deletes the texture from the cache map.
func DeleteTexture(obj fyne.CanvasObject) {
	texturesLock.Lock()
	delete(textures, obj)
	texturesLock.Unlock()
}

// GetTexture gets cached texture.
func GetTexture(obj fyne.CanvasObject) (TextureType, bool) {
	texturesLock.RLock()
	texInfo, has := textures[obj]
	texturesLock.RUnlock()
	if !has {
		return NoTexture, false
	}
	texInfo.setAlive()
	return texInfo.texture, true
}

// RangeExpiredTexturesFor range over the expired textures for the specified canvas.
//
// Note: If this is used to free textures, then it should be called inside a current
// gl context to ensure textures are deleted from gl.
func RangeExpiredTexturesFor(canvas fyne.Canvas, f func(fyne.CanvasObject)) {
	now := timeNow()
	texturesLock.Lock()
	for key, texInfo := range textures {
		if texInfo.isExpired(now) && texInfo.canvas == canvas {
			f(key.(fyne.CanvasObject))
		}
	}
	texturesLock.Unlock()
}

// RangeTexturesFor range over the textures for the specified canvas.
//
// Note: If this is used to free textures, then it should be called inside a current
// gl context to ensure textures are deleted from gl.
func RangeTexturesFor(canvas fyne.Canvas, f func(fyne.CanvasObject)) {
	texturesLock.Lock()
	for key, texInfo := range textures {
		if texInfo.canvas == canvas {
			f(key.(fyne.CanvasObject))
		}
	}
	texturesLock.Unlock()
}

// SetTexture sets cached texture.
func SetTexture(obj fyne.CanvasObject, texture TextureType, canvas fyne.Canvas) {
	texInfo := &textureInfo{texture: texture}
	texInfo.canvas = canvas
	texInfo.setAlive()
	texturesLock.Lock()
	textures[obj] = texInfo
	texturesLock.Unlock()
}

// textureCacheBase defines base texture cache object.
type textureCacheBase struct {
	expiringCacheNoLock
	canvas fyne.Canvas
}
