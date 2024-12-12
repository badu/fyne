package tutorials

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

func windowScreen(_ fyne.Window) fyne.CanvasObject {
	var visibilityWindow fyne.Window = nil
	var visibilityState = false

	windowGroup := container.NewVBox(
		widget.NewButton(
			widget.ButtonWithLabel("New window"),
			widget.ButtonWithCallback(
				func() {
					w := fyne.CurrentApp().NewWindow("Hello")
					w.SetContent(widget.NewLabel(widget.LabelWithStaticText("Hello World!")))
					w.Show()
				},
			),
		),
		widget.NewButton(
			widget.ButtonWithLabel("Fixed size window"),
			widget.ButtonWithCallback(
				func() {
					w := fyne.CurrentApp().NewWindow("Fixed")
					w.SetContent(container.NewCenter(widget.NewLabel(widget.LabelWithStaticText("Hello World!"))))

					w.Resize(fyne.NewSize(240, 180))
					w.SetFixedSize(true)
					w.Show()
				},
			),
		),
		widget.NewButton(
			widget.ButtonWithLabel("Toggle between fixed/not fixed window size"),
			widget.ButtonWithCallback(
				func() {
					w := fyne.CurrentApp().NewWindow("Toggle fixed size")
					w.SetContent(container.NewCenter(widget.NewCheck(
						widget.CheckWithLabel("Fixed size"),
						widget.CheckWithCallback(
							func(toggle bool) {
								if toggle {
									w.Resize(fyne.NewSize(240, 180))
								}
								w.SetFixedSize(toggle)
							},
						),
					),
					),
					)
					w.Show()
				},
			),
		),
		widget.NewButton(
			widget.ButtonWithLabel("Centered window"),
			widget.ButtonWithCallback(
				func() {
					w := fyne.CurrentApp().NewWindow("Central")
					w.SetContent(container.NewCenter(widget.NewLabel(widget.LabelWithStaticText("Hello World!"))))

					w.CenterOnScreen()
					w.Show()
				},
			),
		),
		widget.NewButton(
			widget.ButtonWithLabel("Show/Hide window"),
			widget.ButtonWithCallback(
				func() {
					if visibilityWindow == nil {
						visibilityWindow = fyne.CurrentApp().NewWindow("Hello")
						visibilityWindow.SetContent(widget.NewLabel(widget.LabelWithStaticText("Hello World!")))
						visibilityWindow.Show()
						visibilityWindow.SetOnClosed(func() {
							visibilityWindow = nil
							visibilityState = !visibilityState
						})
					}
					if visibilityState {
						visibilityWindow.Hide()
					} else {
						visibilityWindow.Show()
					}
					visibilityState = !visibilityState
				},
			),
		),
	)

	drv := fyne.CurrentDriver()
	if drv, ok := drv.(desktop.Driver); ok {
		windowGroup.Objects = append(windowGroup.Objects,
			widget.NewButton(
				widget.ButtonWithLabel("Splash Window (only use on start)"),
				widget.ButtonWithCallback(
					func() {
						w := drv.CreateSplashWindow()
						w.SetContent(
							widget.NewLabel(
								widget.LabelWithStaticText("Hello World!\n\nMake a splash!"),
								widget.LabelWithAlignment(fyne.TextAlignCenter),
								widget.LabelWithStyle(fyne.TextStyle{Bold: true}),
							),
						)
						w.Show()

						go func() {
							time.Sleep(time.Second * 3)
							w.Close()
						}()
					},
				),
			),
		)
	}

	otherGroup := widget.NewCard(widget.CardWithTitle("Other"),
		widget.CardWithContent(
			widget.NewButton(
				widget.ButtonWithLabel("Notification"),
				widget.ButtonWithCallback(
					func() {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Fyne Demo",
							Content: "Testing notifications...",
						})
					},
				),
			),
		),
	)

	return container.NewVBox(widget.NewCard(widget.CardWithTitle("Windows"), widget.CardWithContent(windowGroup)), otherGroup)
}
