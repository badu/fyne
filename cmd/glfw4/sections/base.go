package sections

import (
	"errors"

	"fyne.io/fyne/v2/cmd/glfw4/glutils"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// WIDTH is the width of the window
var WIDTH = 1280.0

// HEIGHT is the height of the window
var HEIGHT = 1024.0

var Ratio = float32(WIDTH / HEIGHT)

// Slide is the most basic slide. it has to setup, update, draw and close
type Slide interface {
	Init(a ...interface{}) error
	InitGL() error
	Update()
	Draw()
	Close()
	GetHeader() string
	GetSubHeader() string
	SetName(s string)
	GetName() string
	GetColorHex() string
	HandleKeyboard(k glfw.Key, s int, a glfw.Action, m glfw.ModifierKey, keys map[glfw.Key]bool)
	HandleMousePosition(xpos, ypos float64)
	HandleScroll(xoff, yoff float64)
	HandleFiles(names []string)
	DrawText() bool
}

// BaseSlide is the base implementation of Slide with the min required fields
type BaseSlide struct {
	Slide
	Name     string
	Color    glutils.Color
	Color32  glutils.Color32
	ColorHex string
}

func (s *BaseSlide) GetHeader() string {
	return s.Name
}

func (s *BaseSlide) GetSubHeader() string {
	return ""
}

func (s *BaseSlide) SetName(n string) {
	s.Name = n
}

func (s *BaseSlide) GetColorHex() string {
	return s.ColorHex
}

func (s *BaseSlide) InitGL() error {
	return nil
}

func (s *BaseSlide) Update() {

}

func (s *BaseSlide) Draw() {

}

func (s *BaseSlide) Close() {

}

func (s *BaseSlide) HandleKeyboard(k glfw.Key, sc int, a glfw.Action, mk glfw.ModifierKey, keys map[glfw.Key]bool) {

}

func (s *BaseSlide) HandleMousePosition(xpos, ypos float64) {

}

func (s *BaseSlide) HandleScroll(xoff, yoff float64) {

}

func (s *BaseSlide) HandleFiles(names []string) {

}

type BaseSketch struct {
	BaseSlide
}

func (b *BaseSketch) Init(a ...interface{}) error {
	c, ok := a[1].(glutils.Color)
	if ok == false {
		return errors.New("first argument isnt a color")
	}
	b.Color = c
	b.Color32 = c.To32()
	b.ColorHex = glutils.Rgb2Hex(c)
	return nil
}

func (b *BaseSketch) DrawText() bool {
	return true
}

// returns the index of an object within a slice. returns -1 if it doesnt exist.
func SlidePosition(slice []Slide, value Slide) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}
