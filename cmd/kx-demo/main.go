// Demo is a Fyne app for demonstrating the features provided by the fyne-kx library.
package main

import (
	"fmt"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	kxdialog "fyne.io/fyne/v2/cmd/janice/dialog"
	kxlayout "fyne.io/fyne/v2/cmd/janice/layout"
	kxmodal "fyne.io/fyne/v2/cmd/janice/modal"
	kxwidget "fyne.io/fyne/v2/cmd/janice/widget"
)

func main() {
	app := app.New()
	w := app.NewWindow("KX Demo")
	tabs := container.NewAppTabs(
		container.NewTabItem("Dialogs", makeDialogs(w)),
		container.NewTabItem("Layouts", makeLayouts()),
		container.NewTabItem("Modals", makeModals(w)),
		container.NewTabItem("Widgets", makeWidgets(w)),
	)
	tabs.SetTabLocation(container.TabLocationLeading)
	tabs.SelectIndex(3)

	w.SetContent(container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		tabs,
	))
	w.Resize(fyne.NewSize(600, 500))
	w.ShowAndRun()
}

func makeDialogs(w fyne.Window) fyne.CanvasObject {
	c := container.NewVBox(
		widget.NewButton("Information Dialog with key handler", func() {
			d := dialog.NewInformation("Info", "You can close this dialog with the Escape key.", w)
			kxdialog.AddDialogKeyHandler(d, w)
			d.Show()
		}),
		widget.NewButton("Confirm Dialog with key handler", func() {
			d := dialog.NewConfirm("Confirm", "You can close this dialog with the Escape key.", func(b bool) {
				fmt.Printf("Confirm dialog: %v\n", b)
			}, w)
			kxdialog.AddDialogKeyHandler(d, w)
			d.Show()
		}),
	)
	return c
}

func makeLayouts() fyne.CanvasObject {
	layout := kxlayout.NewColumns(150, 100, 50)
	makeBox := func(h float32) fyne.CanvasObject {
		x := canvas.NewRectangle(theme.Color(theme.ColorNameInputBorder))
		w := rand.Float32()*100 + 50
		x.SetMinSize(fyne.NewSize(w, h))
		return x
	}
	c := container.NewVBox(
		container.New(layout, makeBox(50), makeBox(50), makeBox(50)),
		container.New(layout, makeBox(150), makeBox(150), makeBox(150)),
		container.New(layout, makeBox(30), makeBox(30), makeBox(30)),
	)
	x := widget.NewLabel("Columns")
	x.TextStyle.Bold = true
	return container.NewBorder(
		container.NewVBox(x, widget.NewSeparator()),
		nil,
		nil,
		nil,
		c,
	)
}

func makeWidgets(w fyne.Window) fyne.CanvasObject {
	badge := kxwidget.NewBadge("1234")
	img := kxwidget.NewTappableImage(theme.FyneLogo(), func() {
		d := dialog.NewInformation("TappableImage", "tapped", w)
		kxdialog.AddDialogKeyHandler(d, w)
		d.Show()
	})
	img.SetFillMode(canvas.ImageFillContain)
	img.SetMinSize(fyne.NewSize(100, 100))
	icon := kxwidget.NewTappableIcon(theme.AccountIcon(), func() {
		d := dialog.NewInformation("TappableIcon", "tapped", w)
		kxdialog.AddDialogKeyHandler(d, w)
		d.Show()
	})
	label := kxwidget.NewTappableLabel("Tap me", func() {
		d := dialog.NewInformation("TappableLabel", "tapped", w)
		kxdialog.AddDialogKeyHandler(d, w)
		d.Show()
	})
	slider := kxwidget.NewSlider(0, 100)
	slider.SetValue(25)

	textForBool := func(b bool) string {
		if b {
			return "on"
		}
		return "off"
	}
	switchLabel1 := widget.NewLabel("")
	switch1 := kxwidget.NewSwitch(func(on bool) {
		switchLabel1.SetText(textForBool(on))
	})
	switch1.On = true
	switchLabel1.Text = textForBool(switch1.State())
	switch1Box := container.NewHBox(switch1, switchLabel1)

	switchLabel2 := widget.NewLabel("")
	switch2 := kxwidget.NewSwitch(func(on bool) {
		switchLabel2.SetText(textForBool(on))
	})
	switchLabel2.Text = textForBool(switch2.State())
	switch2Box := container.NewHBox(switch2, switchLabel2)

	switch3 := kxwidget.NewSwitch(nil)
	switch3.On = true
	switch3.Disable()
	switch4 := kxwidget.NewSwitch(nil)
	switch4.Disable()
	addLabel := func(c fyne.CanvasObject, text string) fyne.CanvasObject {
		return container.NewHBox(c, widget.NewLabel(text))
	}

	f := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Badge", Widget: badge},
			{Text: "", Widget: container.NewPadded()},
			{Text: "Slider", Widget: slider},
			{Text: "", Widget: container.NewPadded()},
			{Text: "Switch", Widget: container.NewVBox(
				switch1Box,
				switch2Box,
				addLabel(switch3, "on disabled"),
				addLabel(switch4, "off disabled"),
			)},
			{Text: "", Widget: container.NewPadded()},
			{Text: "TappableIcon", Widget: container.NewHBox(icon)},
			{Text: "", Widget: container.NewPadded()},
			{Text: "TappableImage", Widget: container.NewHBox(img)},
			{Text: "", Widget: container.NewPadded()},
			{Text: "TappableLabel", Widget: label},
		},
	}
	return f
}

func makeModals(w fyne.Window) *fyne.Container {
	b1 := widget.NewButton("ProgressModal", func() {
		m := kxmodal.NewProgress("ProgressModal", "Please wait...", func(progress binding.Float) error {
			for i := 1; i < 50; i++ {
				progress.Set(float64(i))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		}, 50, w)
		m.Start()
	})

	b2 := widget.NewButton("ProgressCancelModal", func() {
		m := kxmodal.NewProgressWithCancel("ProgressCancelModal", "Please wait...", func(progress binding.Float, canceled chan struct{}) error {
			ticker := time.NewTicker(100 * time.Millisecond)
			for i := 1; i < 50; i++ {
				progress.Set(float64(i))
				select {
				case <-canceled:
					return nil
				case <-ticker.C:
				}
			}
			return nil
		}, 50, w)
		m.Start()
	})

	b3 := widget.NewButton("ProgressInfiniteModal", func() {
		m := kxmodal.NewProgressInfinite("ProgressInfiniteModal", "Please wait...", func() error {
			time.Sleep(3 * time.Second)
			return nil
		}, w)
		m.Start()
	})

	b4 := widget.NewButton("ProgressInfiniteCancelModal", func() {
		m := kxmodal.NewProgressInfiniteWithCancel("ProgressInfiniteCancelModal", "Please wait...", func(canceled chan struct{}) error {
			ticker := time.NewTicker(100 * time.Millisecond)
			for i := 1; i < 50; i++ {
				select {
				case <-canceled:
					return nil
				case <-ticker.C:
				}
			}
			return nil
		}, w)
		m.Start()
	})

	return container.NewVBox(b1, b2, b3, b4)
}
