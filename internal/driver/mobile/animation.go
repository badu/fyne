package mobile

import "fyne.io/fyne/v2"

func (d *MobileDriver) StartAnimation(a *fyne.Animation) {
	d.animation.Start(a)
}

func (d *MobileDriver) StopAnimation(a *fyne.Animation) {
	d.animation.Stop(a)
}
