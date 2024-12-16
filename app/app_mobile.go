//go:build !ci && (android || ios || mobile)

package app

import (
	internalapp "fyne.io/fyne/v2/internal/app"
	"fyne.io/fyne/v2/internal/driver/mobile"
)

// NewWithID returns a new app instance using the appropriate runtime driver.
// The ID string should be globally unique to this app.
func NewWithID(id string) *FyneApp {
	d := mobile.NewGoMobileDriver()
	a := newAppWithDriver(d, id)
	d.(mobile.ConfiguredDriver).SetOnConfigurationChanged(func(c *mobile.Configuration) {
		internalapp.SystemTheme = c.SystemTheme

		a.Settings().(*settings).setupTheme()
	})
	return a
}
