//go:build !ci && !wasm && test_web_driver

package app

import (
	"errors"
	"net/url"
)

func (a *FyneApp) OpenURL(url *url.URL) error {
	return errors.New("OpenURL is not supported with the test web driver.")
}
