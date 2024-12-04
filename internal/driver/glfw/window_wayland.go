//go:build wayland && (linux || freebsd || openbsd || netbsd)

package glfw

import (
	"unsafe"

	"fyne.io/fyne/v2/driver"
)

func (w *window) RunNative(f func(any)) {
	runOnMain(func() {
		var waylandSurface uintptr
		if v := w.view(); v != nil {
			waylandSurface = uintptr(unsafe.Pointer(v.GetWaylandWindow()))
		}
		f(driver.WaylandWindowContext{
			WaylandSurface: waylandSurface,
		})
	})
}
