package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	width  = 800
	height = 600
)

var (
	vertexShaderSource = `
		#version 330 core
		layout (location = 0) in vec3 aPos;
		
		void main() {
			gl_Position = vec4(aPos.x, aPos.y, aPos.z, 1.0);
		}
	` + "\x00"

	fragmentShaderSource = `
		#version 330 core
		out vec4 FragColor;
		
		uniform float time;
		
		void main() {
			float width = (sin(time) + 1.0) / 2.0; // Oscillates between 0 and 1
			if (gl_FragCoord.x / 800.0 < width) {
				FragColor = vec4(1.0, 0.0, 0.0, 1.0); // Red color for the rectangle
			} else {
				FragColor = vec4(0.0, 0.0, 0.0, 1.0); // Black color for the background
			}
		}
	` + "\x00"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Animated Rectangle", nil, nil)
	if err != nil {
		log.Fatalln("failed to create window:", err)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatalln("failed to initialize OpenGL:", err)
	}

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		log.Fatalln("failed to compile vertex shader:", err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		log.Fatalln("failed to compile fragment shader:", err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	vertices := []float32{
		-1.0, -1.0, 0.0,
		1.0, -1.0, 0.0,
		-1.0, 1.0, 0.0,
		1.0, 1.0, 0.0,
	}

	var vertexBuffer uint32
	gl.GenBuffers(1, &vertexBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

	var vertexArray uint32
	gl.GenVertexArrays(1, &vertexArray)
	gl.BindVertexArray(vertexArray)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	timeUniform := gl.GetUniformLocation(prog, gl.Str("time\x00"))

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.UseProgram(prog)
		gl.Uniform1f(timeUniform, float32(glfw.GetTime()))

		gl.BindVertexArray(vertexArray)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}
