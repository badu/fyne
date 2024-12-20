package main

import (
	"fmt"
	gltext "fyne.io/fyne/v2/text"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/math/fixed"
	"math"
	"os"
	"runtime"
	"time"
)

var useStrictCoreProfile = (runtime.GOOS == "darwin")

func main() {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		panic("glfw error")
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	if useStrictCoreProfile {
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	}
	glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True)

	window, err := glfw.CreateWindow(1280, 960, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("Opengl version", version)

	// code from here
	gltext.IsDebug = true

	var font *gltext.Font
	config, err := gltext.LoadTruetypeFontConfig("fontconfigs", "font_1_honokamin")
	if err == nil {
		font, err = gltext.NewFont(config)
		if err != nil {
			panic(err)
		}
		fmt.Println("Font loaded from disk...")
	} else {
		fd, err := os.Open("./cmd/glfw3/luxisr.ttf")
		if err != nil {
			panic(err)
		}
		defer fd.Close()

		// Japanese character ranges
		// http://www.rikai.com/library/kanjitables/kanji_codes.unicode.shtml
		runeRanges := make(gltext.RuneRanges, 0)
		runeRanges = append(runeRanges, gltext.RuneRange{Low: 32, High: 128})
		runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x3000, High: 0x3030})
		runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x3040, High: 0x309f})
		runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x30a0, High: 0x30ff})
		runeRanges = append(runeRanges, gltext.RuneRange{Low: 0x4e00, High: 0x9faf})
		runeRanges = append(runeRanges, gltext.RuneRange{Low: 0xff00, High: 0xffef})

		scale := fixed.Int26_6(32)
		runesPerRow := fixed.Int26_6(128)
		config, err = gltext.NewTruetypeFontConfig(fd, scale, runeRanges, runesPerRow, 5)
		if err != nil {
			panic(err)
		}
		err = config.Save("fontconfigs", "font_1_honokamin")
		if err != nil {
			panic(err)
		}
		font, err = gltext.NewFont(config)
		if err != nil {
			panic(err)
		}
	}

	width, height := window.GetSize()
	font.ResizeWindow(float32(width), float32(height))

	str0 := "! \" # $ ' ( ) + , - . / 0123456789 : "
	str1 := "; <=> ? @ ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	str2 := "[^_`] abcdefghijklmnopqrstuvwxyz {|}"
	str3 := "大好きどなにｂｃｄｆｇｈｊｋｌｍｎｐｑ"

	scaleMin, scaleMax := float32(1.0), float32(1.1)
	strs := []string{str0, str1, str2, str3}
	txts := []*gltext.Text{}
	for _, str := range strs {
		text := gltext.NewText(font, scaleMin, scaleMax)
		text.SetString(str)
		text.SetColor(mgl32.Vec3{1, 1, 1})
		text.FadeOutPerFrame = 0.01
		if gltext.IsDebug {
			for _, s := range str {
				fmt.Printf("%c: %d\n", s, rune(s))
			}
		}
		txts = append(txts, text)
	}

	start := time.Now()

	gl.ClearColor(0.4, 0.4, 0.4, 0.0)
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Position can be set freely
		for index, text := range txts {
			text.SetPosition(mgl32.Vec2{0, float32(index * 50)})
			text.Draw()

			// just for illustrative purposes
			// i imagine that user interaction of some sort will trigger these rather than a moment in time

			// fade out
			if math.Floor(time.Now().Sub(start).Seconds()) == 5 {
				text.BeginFadeOut()
			}

			// show text
			if math.Floor(time.Now().Sub(start).Seconds()) == 10 {
				text.Show()
			}

			// hide
			if math.Floor(time.Now().Sub(start).Seconds()) == 15 {
				text.Hide()
			}
		}

		window.SwapBuffers()
		glfw.PollEvents()
	}
	for _, text := range txts {
		text.Release()
	}
	font.Release()
}
