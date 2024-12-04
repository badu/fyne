//go:build wasm || test_web_driver

package glfw

import (
	glfw "github.com/fyne-io/glfw-js"
)

// clipboard represents the system clipboard
type clipboard struct {
	window *glfw.Window
}

// Content returns the clipboard content
func (c *clipboard) Content() string {
	content := ""
	runOnMain(func() {
		content, _ = c.window.GetClipboardString()
	})
	return content
}

// SetContent sets the clipboard content
func (c *clipboard) SetContent(content string) {
	runOnMain(func() {
		c.window.SetClipboardString(content)
	})
}
