package glfw

import "fyne.io/fyne/v2"

func (d *GLDriver) StartAnimation(a *fyne.Animation) {
	d.animation.Start(a)
}

func (d *GLDriver) StopAnimation(a *fyne.Animation) {
	d.animation.Stop(a)
}
