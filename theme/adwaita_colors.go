package theme

// This file is generated by adwaita_theme_generator.go
// Please do not edit manually, use:
// go generate ./theme/...
//
// The colors are taken from: https://gnome.pages.gitlab.gnome.org/libadwaita/doc/1.0/named-colors.html

import (
	"image/color"

	"fyne.io/fyne/v2"
)

var adwaitaDarkScheme = map[fyne.ThemeColorName]color.Color{
	ColorBlue:                  color.NRGBA{R: 0x35, G: 0x84, B: 0xe4, A: 0xff}, // Adwaita color name @blue_3
	ColorBrown:                 color.NRGBA{R: 0x98, G: 0x6a, B: 0x44, A: 0xff}, // Adwaita color name @brown_3
	ColorGray:                  color.NRGBA{R: 0x5e, G: 0x5c, B: 0x64, A: 0xff}, // Adwaita color name @dark_2
	ColorGreen:                 color.NRGBA{R: 0x26, G: 0xa2, B: 0x69, A: 0xff}, // Adwaita color name @green_5
	ColorNameBackground:        color.NRGBA{R: 0x24, G: 0x24, B: 0x24, A: 0xff}, // Adwaita color name @window_bg_color
	ColorNameButton:            color.NRGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, // Adwaita color name @headerbar_bg_color
	ColorNameError:             color.NRGBA{R: 0xc0, G: 0x1c, B: 0x28, A: 0xff}, // Adwaita color name @error_bg_color
	ColorNameForeground:        color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, // Adwaita color name @window_fg_color
	ColorNameInputBackground:   color.NRGBA{R: 0x1e, G: 0x1e, B: 0x1e, A: 0xff}, // Adwaita color name @view_bg_color
	ColorNameMenuBackground:    color.NRGBA{R: 0x38, G: 0x38, B: 0x38, A: 0xff}, // Adwaita color name @popover_bg_color
	ColorNameOverlayBackground: color.NRGBA{R: 0x1e, G: 0x1e, B: 0x1e, A: 0xff}, // Adwaita color name @view_bg_color
	ColorNamePrimary:           color.NRGBA{R: 0x35, G: 0x84, B: 0xe4, A: 0xff}, // Adwaita color name @accent_bg_color
	ColorNameScrollBar:         color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x5b}, // Adwaita color name @light_1
	ColorNameSelection:         color.NRGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xff}, // Adwaita color name @headerbar_bg_color
	ColorNameShadow:            color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x5b}, // Adwaita color name @shade_color
	ColorNameSuccess:           color.NRGBA{R: 0x26, G: 0xa2, B: 0x69, A: 0xff}, // Adwaita color name @success_bg_color
	ColorNameWarning:           color.NRGBA{R: 0xcd, G: 0x93, B: 0x09, A: 0xff}, // Adwaita color name @warning_bg_color
	ColorOrange:                color.NRGBA{R: 0xff, G: 0x78, B: 0x00, A: 0xff}, // Adwaita color name @orange_3
	ColorPurple:                color.NRGBA{R: 0x91, G: 0x41, B: 0xac, A: 0xff}, // Adwaita color name @purple_3
	ColorRed:                   color.NRGBA{R: 0xc0, G: 0x1c, B: 0x28, A: 0xff}, // Adwaita color name @red_4
	ColorYellow:                color.NRGBA{R: 0xf6, G: 0xd3, B: 0x2d, A: 0xff}, // Adwaita color name @yellow_3
}

var adwaitaLightScheme = map[fyne.ThemeColorName]color.Color{
	ColorBlue:                  color.NRGBA{R: 0x35, G: 0x84, B: 0xe4, A: 0xff}, // Adwaita color name @blue_3
	ColorBrown:                 color.NRGBA{R: 0x98, G: 0x6A, B: 0x44, A: 0xff}, // Adwaita color name @brown_3
	ColorGray:                  color.NRGBA{R: 0x5e, G: 0x5C, B: 0x64, A: 0xff}, // Adwaita color name @dark_2
	ColorGreen:                 color.NRGBA{R: 0x2e, G: 0xC2, B: 0x7e, A: 0xff}, // Adwaita color name @green_4
	ColorNameBackground:        color.NRGBA{R: 0xfa, G: 0xFA, B: 0xfa, A: 0xff}, // Adwaita color name @window_bg_color
	ColorNameButton:            color.NRGBA{R: 0xeb, G: 0xEB, B: 0xeb, A: 0xff}, // Adwaita color name @headerbar_bg_color
	ColorNameError:             color.NRGBA{R: 0xe0, G: 0x1B, B: 0x24, A: 0xff}, // Adwaita color name @error_bg_color
	ColorNameForeground:        color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xcc}, // Adwaita color name @window_fg_color
	ColorNameInputBackground:   color.NRGBA{R: 0xff, G: 0xFF, B: 0xff, A: 0xff}, // Adwaita color name @view_bg_color
	ColorNameMenuBackground:    color.NRGBA{R: 0xff, G: 0xFF, B: 0xff, A: 0xff}, // Adwaita color name @popover_bg_color
	ColorNameOverlayBackground: color.NRGBA{R: 0xff, G: 0xFF, B: 0xff, A: 0xff}, // Adwaita color name @view_bg_color
	ColorNamePrimary:           color.NRGBA{R: 0x35, G: 0x84, B: 0xe4, A: 0xff}, // Adwaita color name @accent_bg_color
	ColorNameScrollBar:         color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x5b}, // Adwaita color name @dark_5
	ColorNameSelection:         color.NRGBA{R: 0xeb, G: 0xEB, B: 0xeb, A: 0xff}, // Adwaita color name @headerbar_bg_color
	ColorNameShadow:            color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x11}, // Adwaita color name @shade_color
	ColorNameSuccess:           color.NRGBA{R: 0x2e, G: 0xC2, B: 0x7e, A: 0xff}, // Adwaita color name @success_bg_color
	ColorNameWarning:           color.NRGBA{R: 0xe5, G: 0xA5, B: 0x0a, A: 0xff}, // Adwaita color name @warning_bg_color
	ColorOrange:                color.NRGBA{R: 0xff, G: 0x78, B: 0x00, A: 0xff}, // Adwaita color name @orange_3
	ColorPurple:                color.NRGBA{R: 0x91, G: 0x41, B: 0xac, A: 0xff}, // Adwaita color name @purple_3
	ColorRed:                   color.NRGBA{R: 0xe0, G: 0x1B, B: 0x24, A: 0xff}, // Adwaita color name @red_3
	ColorYellow:                color.NRGBA{R: 0xf6, G: 0xD3, B: 0x2d, A: 0xff}, // Adwaita color name @yellow_3
}
