//go:build wasm || test_web_driver

package glfw

import (
	"time"

	"fyne.io/fyne/v2"
)

const webDefaultDoubleTapDelay = 300 * time.Millisecond

func (d *GLDriver) SetSystemTrayMenu(m *fyne.Menu) {
	// no-op for mobile apps using this driver
}

func (d *GLDriver) catchTerm() {
}

func setDisableScreenBlank(disable bool) {
	// awaiting complete support for WakeLock
}

func (g *GLDriver) DoubleTapDelay() time.Duration {
	return webDefaultDoubleTapDelay
}
