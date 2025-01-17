package glfont

import (
	"fmt"
	"fyne.io/fyne/v2/cmd/glfw4/glutils"
	"github.com/go-gl/gl/v4.1-core/gl"
	"os"
)

// Direction represents the direction in which strings should be rendered.
type Direction uint8

// Known directions.
const (
	LeftToRight Direction = iota // E.g.: Latin
	RightToLeft                  // E.g.: Arabic
	TopToBottom                  // E.g.: Chinese
)

// A Font allows rendering of text to an OpenGL context.
type Font struct {
	fontChar []*character
	vao      uint32
	vbo      uint32
	program  uint32
	texture  uint32 // Holds the glyph texture id.
	color    fcolor
}

type fcolor struct {
	r float32
	g float32
	b float32
	a float32
}

// LoadFont loads the specified font at the given scale.
func LoadFont(file string, scale int32, windowWidth, windowHeight float64) (*Font, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	// Configure the default font vertex and fragment shaders
	program, err := glutils.BasicProgram(vertexFontShader, fragmentFontShader)
	if err != nil {
		panic(err)
	}

	// Activate corresponding render state
	gl.UseProgram(program)

	//set screen resolution
	resUniform := gl.GetUniformLocation(program, gl.Str("resolution\x00"))
	gl.Uniform2f(resUniform, float32(windowWidth), float32(windowHeight))

	return LoadTrueTypeFont(program, fd, scale, 32, 127, LeftToRight)
}

// SetColor allows you to set the text color to be used when you draw the text
func (f *Font) SetColor(red float32, green float32, blue float32, alpha float32) {
	f.color.r = red
	f.color.g = green
	f.color.b = blue
	f.color.a = alpha
}

func (f *Font) Resize(windowWidth, windowHeight float64) {
	// Activate corresponding render state
	gl.UseProgram(f.program)

	//set screen resolution
	resUniform := gl.GetUniformLocation(f.program, gl.Str("resolution\x00"))
	gl.Uniform2f(resUniform, float32(windowWidth), float32(windowHeight))
}

// Printf draws a string to the screen, takes a list of arguments like printf
func (f *Font) Printf(x, y float32, scale float32, fs string, argv ...interface{}) error {

	indices := []rune(fmt.Sprintf(fs, argv...))

	if len(indices) == 0 {
		return nil
	}

	lowChar := rune(32)

	//setup blending mode
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Activate corresponding render state
	gl.UseProgram(f.program)
	//set text color
	gl.Uniform4f(gl.GetUniformLocation(f.program, gl.Str("textColor\x00")), f.color.r, f.color.g, f.color.b, f.color.a)
	//set screen resolution
	//resUniform := gl.GetUniformLocation(f.program, gl.Str("resolution\x00"))
	//gl.Uniform2f(resUniform, float32(2560), float32(1440))

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindVertexArray(f.vao)

	// Iterate through all characters in string
	for i := range indices {

		//get rune
		runeIndex := indices[i]

		//skip runes that are not in font chacter range
		if int(runeIndex)-int(lowChar) > len(f.fontChar) || runeIndex < lowChar {
			continue
		}

		//find rune in fontChar list
		ch := f.fontChar[runeIndex-lowChar]

		//calculate position and size for current rune
		xpos := x + float32(ch.bearingH)*scale
		ypos := y - float32(ch.height-ch.bearingV)*scale
		w := float32(ch.width) * scale
		h := float32(ch.height) * scale

		//set quad positions
		var x1 = xpos
		var x2 = xpos + w
		var y1 = ypos
		var y2 = ypos + h

		//setup quad array
		var vertices = []float32{
			//  X, Y, Z, U, V
			// Front
			x1, y1, 0.0, 0.0,
			x2, y1, 1.0, 0.0,
			x1, y2, 0.0, 1.0,
			x1, y2, 0.0, 1.0,
			x2, y1, 1.0, 0.0,
			x2, y2, 1.0, 1.0}

		// Render glyph texture over quad
		gl.BindTexture(gl.TEXTURE_2D, ch.textureID)
		// Update content of VBO memory
		gl.BindBuffer(gl.ARRAY_BUFFER, f.vbo)

		//BufferSubData(target Enum, offset int, data []byte)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, gl.Ptr(vertices)) // Be sure to use glBufferSubData and not glBufferData
		// Render quad
		gl.DrawArrays(gl.TRIANGLES, 0, 24)

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		// Now advance cursors for next glyph (note that advance is number of 1/64 pixels)
		x += float32((ch.advance >> 6)) * scale // Bitshift by 6 to get value in pixels (2^6 = 64 (divide amount of 1/64th pixels by 64 to get amount of pixels))

	}

	//clear opengl textures and programs
	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.UseProgram(0)
	gl.Disable(gl.BLEND)

	return nil
}

var fragmentFontShader = `#version 150 core
in vec2 fragTexCoord;
out vec4 outputColor;

uniform sampler2D tex;
uniform vec4 textColor;

void main()
{
    vec4 sampled = vec4(1.0, 1.0, 1.0, texture(tex, fragTexCoord).r);
    outputColor = textColor * sampled;
}` + "\x00"

var vertexFontShader = `#version 150 core

//vertex position
in vec2 vert;

//pass through to fragTexCoord
in vec2 vertTexCoord;

//window res
uniform vec2 resolution;

//pass to frag
out vec2 fragTexCoord;

void main() {
   // convert the rectangle from pixels to 0.0 to 1.0
   vec2 zeroToOne = vert / resolution;

   // convert from 0->1 to 0->2
   vec2 zeroToTwo = zeroToOne * 2.0;

   // convert from 0->2 to -1->+1 (clipspace)
   vec2 clipSpace = zeroToTwo - 1.0;

   fragTexCoord = vertTexCoord;

   gl_Position = vec4(clipSpace * vec2(1, -1), 0, 1);
}` + "\x00"
