//go:build android

package mobile

import (
	fyneDriver "fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/internal/driver/mobile/app"
)

func (w *window) RunNative(f func(context any)) {
	app.RunOnJVM(func(vm, env, ctx uintptr) error {
		data := &fyneDriver.AndroidWindowContext{
			NativeWindow: w.handle,
			AndroidContext: fyneDriver.AndroidContext{
				VM:  vm,
				Env: env,
				Ctx: ctx,
			},
		}
		f(data)
		return nil
	})
}
