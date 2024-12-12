package fyne

import "time"

type systemTrayDriver interface {
	// CreateWindow creates a new UI Window.
	CreateWindow(string) Window
	// AllWindows returns a slice containing all app windows.
	AllWindows() []Window

	// RenderedTextSize returns the size required to render the given string of specified
	// font size and style. It also returns the height to text baseline, measured from the top.
	// If the source is specified it will be used, otherwise the current theme will be asked for the font.
	RenderedTextSize(text string, fontSize float32, style TextStyle, source Resource) (size Size, baseline float32)

	// CanvasForObject returns the canvas that is associated with a given [CanvasObject].
	CanvasForObject(CanvasObject) Canvas
	// AbsolutePositionForObject returns the position of a given [CanvasObject] relative to the top/left of a canvas.
	AbsolutePositionForObject(CanvasObject) Position

	// Device returns the device that the application is currently running on.
	Device() Device
	// Run starts the main event loop of the driver.
	Run()
	// Quit closes the driver and open windows, then exit the application.
	// On some some operating systems this does nothing, for example iOS and Android.
	Quit()

	// StartAnimation registers a new animation with this driver and requests it be started.
	StartAnimation(*Animation)
	// StopAnimation stops an animation and unregisters from this driver.
	StopAnimation(*Animation)

	// DoubleTapDelay returns the maximum duration where a second tap after a first one
	// will be considered a [DoubleTap] instead of two distinct [Tap] events.
	//
	// Since: 2.5
	DoubleTapDelay() time.Duration

	// SetDisableScreenBlanking allows an app to ask the device not to sleep/lock/blank displays
	//
	// Since: 2.5
	SetDisableScreenBlanking(bool)
	SetSystemTrayMenu(*Menu)
	SystemTrayMenu() *Menu
}

// Menu stores the information required for a standard menu.
// A menu can pop down from a [MainMenu] or could be a pop out menu.
type Menu struct {
	Label string
	Items []*MenuItem
}

// NewMenu creates a new menu given the specified label (to show in a [MainMenu]) and list of items to display.
func NewMenu(label string, items ...*MenuItem) *Menu {
	return &Menu{Label: label, Items: items}
}

// Refresh will instruct this menu to update its display.
//
// Since: 2.2
func (m *Menu) Refresh() {
	for _, w := range CurrentDriver().AllWindows() {
		main := w.MainMenu()
		if main != nil {
			for _, menu := range main.Items {
				if menu == m {
					w.SetMainMenu(main)
					break
				}
			}
		}
	}

	if d, ok := CurrentDriver().(systemTrayDriver); ok {
		if m == d.SystemTrayMenu() {
			d.SetSystemTrayMenu(m)
		}
	}
}

// MenuItem is a single item within any menu, it contains a display Label and Action function that is called when tapped.
type MenuItem struct {
	ChildMenu *Menu
	// Since: 2.1
	IsQuit      bool
	IsSeparator bool
	Label       string
	Action      func()
	// Since: 2.1
	Disabled bool
	// Since: 2.1
	Checked bool
	// Since: 2.2
	Icon Resource
	// Since: 2.2
	Shortcut Shortcut
}

// NewMenuItem creates a new menu item from the passed label and action parameters.
func NewMenuItem(label string, action func()) *MenuItem {
	return &MenuItem{Label: label, Action: action}
}

// NewMenuItemSeparator creates a menu item that is to be used as a separator.
func NewMenuItemSeparator() *MenuItem {
	return &MenuItem{IsSeparator: true, Action: func() {}}
}

// MainMenu defines the data required to show a menu bar (desktop) or other appropriate top level menu.
type MainMenu struct {
	Items []*Menu
}

// NewMainMenu creates a top level menu structure used by fyne.Window for displaying a menubar
// (or appropriate equivalent).
func NewMainMenu(items ...*Menu) *MainMenu {
	return &MainMenu{Items: items}
}

// Refresh will instruct any rendered menus using this struct to update their display.
//
// Since: 2.2
func (m *MainMenu) Refresh() {
	for _, w := range CurrentDriver().AllWindows() {
		menu := w.MainMenu()
		if menu != nil && menu == m {
			w.SetMainMenu(m)
		}
	}
}
