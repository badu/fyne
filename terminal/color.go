package terminal

import (
	"image/color"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	basicColors = []color.Color{
		color.Black,
		&color.RGBA{R: 170, A: 255},
		&color.RGBA{G: 170, A: 255},
		&color.RGBA{R: 170, G: 170, A: 255},
		&color.RGBA{B: 170, A: 255},
		&color.RGBA{R: 170, B: 170, A: 255},
		&color.RGBA{G: 255, B: 255, A: 255},
		&color.RGBA{R: 170, G: 170, B: 170, A: 255},
	}
	brightColors = []color.Color{
		&color.RGBA{R: 85, G: 85, B: 85, A: 255},
		&color.RGBA{R: 255, G: 85, B: 85, A: 255},
		&color.RGBA{R: 85, G: 255, B: 85, A: 255},
		&color.RGBA{R: 255, G: 255, B: 85, A: 255},
		&color.RGBA{R: 85, G: 85, B: 255, A: 255},
		&color.RGBA{R: 255, G: 85, B: 255, A: 255},
		&color.RGBA{R: 85, G: 255, B: 255, A: 255},
		&color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
	colourBands = []uint8{
		0x00,
		0x5f,
		0x87,
		0xaf,
		0xd7,
		0xff,
	}
)

func (t *Terminal) handleColorEscape(message string) {
	if message == "" || message == "0" {
		t.currentBG = nil
		t.currentFG = nil
		t.bold = false
		t.blinking = false
		return
	}
	modes := strings.Split(message, ";")
	for i := 0; i < len(modes); i++ {
		mode := modes[i]
		if mode == "" {
			continue
		}

		if (mode == "38" || mode == "48") && i+1 < len(modes) {
			nextMode := modes[i+1]
			if nextMode == "5" && i+2 < len(modes) {
				t.handleColorModeMap(mode, modes[i+2])
				i += 2
			} else if nextMode == "2" && i+4 < len(modes) {
				t.handleColorModeRGB(mode, modes[i+2], modes[i+3], modes[i+4])
				i += 4
			}
		} else {
			t.handleColorMode(mode)
		}
	}
}

func (t *Terminal) handleColorMode(modeStr string) {
	mode, err := strconv.Atoi(modeStr)
	if err != nil {
		fyne.LogError("Failed to parse color mode: "+modeStr, err)
		return
	}
	switch mode {
	case 0:
		t.currentBG, t.currentFG = nil, nil
		t.bold = false
		t.blinking = false
	case 1:
		t.bold = true
	case 4, 24: //italic
	case 5:
		t.blinking = true
	case 7: // reverse
		bg, fg := t.currentBG, t.currentFG
		if fg == nil {
			t.currentBG = theme.Color(theme.ColorNameForeground)
		} else {
			t.currentBG = fg
		}
		if bg == nil {
			t.currentFG = theme.Color(theme.ColorNameDisabledButton)
		} else {
			t.currentFG = bg
		}
	case 27: // reverse off
		bg, fg := t.currentBG, t.currentFG
		if fg != nil {
			t.currentBG = nil
		} else {
			t.currentBG = fg
		}
		if bg != nil {
			t.currentFG = nil
		} else {
			t.currentFG = bg
		}
	case 30, 31, 32, 33, 34, 35, 36, 37:
		t.currentFG = basicColors[mode-30]
	case 39:
		t.currentFG = nil
	case 40, 41, 42, 43, 44, 45, 46, 47:
		t.currentBG = basicColors[mode-40]
	case 49:
		t.currentBG = nil
	case 90, 91, 92, 93, 94, 95, 96, 97:
		t.currentFG = brightColors[mode-90]
	case 100, 101, 102, 103, 104, 105, 106, 107:
		t.currentBG = brightColors[mode-100]
	default:
		if t.debug {
			log.Println("Unsupported graphics mode", mode)
		}
	}
}

func (t *Terminal) handleColorModeMap(mode, ids string) {
	var c color.Color
	id, err := strconv.Atoi(ids)
	if err != nil {
		if t.debug {
			log.Println("Invalid color map ID", ids)
		}
		return
	}
	if id <= 7 {
		c = basicColors[id]
	} else if id <= 15 {
		c = brightColors[id-8]
	} else if id <= 231 {
		id -= 16
		b := id % 6
		id = (id - b) / 6
		g := id % 6
		r := (id - g) / 6
		c = &color.RGBA{R: colourBands[r], G: colourBands[g], B: colourBands[b], A: 255}
	} else if id <= 255 {
		id -= 232
		inc := 256 / 24
		y := id * inc
		c = &color.Gray{Y: uint8(y)}
	} else if t.debug {
		log.Println("Invalid colour map ID", id)
	}

	if mode == "38" {
		t.currentFG = c
	} else if mode == "48" {
		t.currentBG = c
	}
}

func (t *Terminal) handleColorModeRGB(mode, rs, gs, bs string) {
	r, _ := strconv.Atoi(rs)
	g, _ := strconv.Atoi(gs)
	b, _ := strconv.Atoi(bs)
	c := &color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}

	if mode == "38" {
		t.currentFG = c
	} else if mode == "48" {
		t.currentBG = c
	}
}
