package terminal

import (
	"image/color"
	"io"
	"math"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

const (
	bufLen = 32768 // 32KB buffer for output, to align with modern L1 cache
)

// Config is the state of a terminal, updated upon certain actions or commands.
// Use Terminal.OnConfigure hook to register for changes.
type Config struct {
	Title         string
	Rows, Columns uint
}

type charSet int

const (
	charSetANSII charSet = iota
	charSetDECSpecialGraphics
	charSetAlternate
)

// Terminal is a terminal widget that loads a shell and handles input/output.
type Terminal struct {
	widget.BaseWidget

	pty io.Closer
	in  io.WriteCloser
	out io.Reader

	currentFG   color.Color
	currentBG   color.Color
	mouseCursor desktop.Cursor

	printer                Printer
	readWriterConfigurator ReadWriterConfigurator

	clipboard fyne.Clipboard
	content   *widget.TermGrid

	cursor      *canvas.Rectangle
	cursorMoved func()

	onMouseDown func(int, fyne.KeyModifier, fyne.Position)
	onMouseUp   func(int, fyne.KeyModifier, fyne.Position)

	selStart, selEnd *position
	state            *parseState
	cmd              *exec.Cmd
	fyne.ShortcutHandler
	startDir string

	listeners               []chan Config
	printData               []byte
	config                  Config
	cursorRow, cursorCol    int
	savedRow, savedCol      int
	scrollTop, scrollBottom int

	g0Charset    charSet
	g1Charset    charSet
	listenerLock sync.Mutex

	keyboardState struct {
		shiftPressed bool
		ctrlPressed  bool
		altPressed   bool
	}

	bell         bool
	bold         bool
	debug        bool
	focused      bool
	cursorHidden bool
	bufferMode   bool // buffer mode is an xterm extension that impacts control keys
	useG1CharSet bool

	blockMode          bool
	highlightBitMask   uint8
	selecting          bool
	newLineMode        bool // new line mode or line feed mode
	bracketedPasteMode bool
	blinking           bool
}

// Printer is used for spooling print data when its received.
type Printer interface {
	Print([]byte)
}

// PrinterFunc is a helper function to enable easy implementation of printers.
type PrinterFunc func([]byte)

// Print calls the PrinterFunc.
func (p PrinterFunc) Print(d []byte) {
	p(d)
}

// Cursor is used for displaying a specific cursor.
func (t *Terminal) Cursor() desktop.Cursor {
	return t.mouseCursor
}

// AcceptsTab indicates that this widget will use the Tab key (avoids loss of focus).
func (t *Terminal) AcceptsTab() bool {
	return true
}

// AddListener registers a new outgoing channel that will have our Config sent each time it changes.
func (t *Terminal) AddListener(listener chan Config) {
	t.listenerLock.Lock()
	defer t.listenerLock.Unlock()

	t.listeners = append(t.listeners, listener)
}

// MinSize provides a size large enough that a terminal could technically funcion.
func (t *Terminal) MinSize() fyne.Size {
	s := guessCellSize()
	return fyne.NewSize(s.Width*2.5, s.Height*1.2) // just enough to get a terminal init
}

// MouseDown handles the down action for desktop mouse events.
func (t *Terminal) MouseDown(ev *desktop.MouseEvent) {
	if t.hasSelectedText() {
		t.copySelectedText(t.clipboard)
		t.clearSelectedText()
	}
	if ev.Button == desktop.MouseButtonSecondary {
		t.pasteText(t.clipboard)
	}

	if t.onMouseDown == nil {
		return
	}

	if ev.Button == desktop.MouseButtonPrimary {
		t.onMouseDown(1, ev.Modifier, ev.Position)
	} else if ev.Button == desktop.MouseButtonSecondary {
		t.onMouseDown(2, ev.Modifier, ev.Position)
	}
}

// MouseUp handles the up action for desktop mouse events.
func (t *Terminal) MouseUp(ev *desktop.MouseEvent) {

	if t.onMouseDown == nil {
		return
	}

	if ev.Button == desktop.MouseButtonPrimary {
		t.onMouseUp(1, ev.Modifier, ev.Position)
	} else if ev.Button == desktop.MouseButtonSecondary {
		t.onMouseUp(2, ev.Modifier, ev.Position)
	}
}

// DoubleTapped handles the double tapped event.
func (t *Terminal) DoubleTapped(pe *fyne.PointEvent) {
	pos := t.sanitizePosition(pe.Position)
	termPos := t.getTermPosition(*pos)
	row, col := termPos.Row, termPos.Col

	if t.hasSelectedText() {
		t.clearSelectedText()
	}

	if row < 1 || row > len(t.content.Rows) {
		return
	}

	rowContent := t.content.Rows[row-1].Cells

	if col < 0 || col >= len(rowContent) {
		return // No valid character under the cursor, do nothing
	}

	start, end := col-1, col-1

	if !unicode.IsLetter(rowContent[start].Rune) && !unicode.IsDigit(rowContent[start].Rune) {
		return
	}

	for start > 0 && (unicode.IsLetter(rowContent[start-1].Rune) || unicode.IsDigit(rowContent[start-1].Rune)) {
		start--
	}
	if start < len(rowContent) && !unicode.IsLetter(rowContent[start].Rune) && !unicode.IsDigit(rowContent[start].Rune) {
		start++
	}
	for end < len(rowContent) && (unicode.IsLetter(rowContent[end].Rune) || unicode.IsDigit(rowContent[end].Rune)) {
		end++
	}
	if start == end {
		return
	}

	t.selStart = &position{Row: row, Col: start + 1}
	t.selEnd = &position{Row: row, Col: end}

	t.highlightSelectedText()
}

// RemoveListener de-registers a Config channel and closes it
func (t *Terminal) RemoveListener(listener chan Config) {
	t.listenerLock.Lock()
	defer t.listenerLock.Unlock()

	for i, l := range t.listeners {
		if l == listener {
			if i < len(t.listeners)-1 {
				t.listeners = append(t.listeners[:i], t.listeners[i+1:]...)
			} else {
				t.listeners = t.listeners[:i]
			}
			close(l)
			return
		}
	}
}

// Resize is called when this terminal widget has been resized.
// It ensures that the virtual terminal is within the bounds of the widget.
func (t *Terminal) Resize(s fyne.Size) {
	cellSize := guessCellSize()
	cols := uint(math.Floor(float64(s.Width) / float64(cellSize.Width)))
	rows := uint(math.Floor(float64(s.Height) / float64(cellSize.Height)))
	if (t.config.Columns == cols) && (t.config.Rows == rows) {
		return
	}

	t.BaseWidget.Resize(s)
	t.content.Resize(fyne.NewSize(float32(cols)*cellSize.Width, float32(rows)*cellSize.Height))

	oldRows := int(t.config.Rows)
	t.config.Columns, t.config.Rows = cols, rows
	if t.scrollBottom == 0 || t.scrollBottom == oldRows-1 {
		t.scrollBottom = int(t.config.Rows) - 1
	}
	t.onConfigure()

	go t.updatePTYSize()
}

// SetDebug turns on output about terminal codes and other errors if the parameter is `true`.
func (t *Terminal) SetDebug(debug bool) {
	t.debug = debug
}

// SetStartDir can be called before one of the Run calls to specify the initial directory.
func (t *Terminal) SetStartDir(path string) {
	t.startDir = path
}

// Tapped makes sure we ask for focus if user taps us.
func (t *Terminal) Tapped(ev *fyne.PointEvent) {
	fyne.CurrentDriver().CanvasForObject(t).Focus(t)
}

// Text returns the contents of the buffer as a single string joined with `\n` (no style information).
func (t *Terminal) Text() string {
	return t.content.Text()
}

// ExitCode returns the exit code from the terminal's shell.
// Returns -1 if called before shell was started or before shell exited.
// Also returns -1 if shell was terminated by a signal.
func (t *Terminal) ExitCode() int {
	if t.cmd == nil {
		return -1
	}
	return t.cmd.ProcessState.ExitCode()
}

// TouchCancel handles the tap action for mobile apps that lose focus during tap.
func (t *Terminal) TouchCancel(ev *mobile.TouchEvent) {
	if t.onMouseUp != nil {
		t.onMouseUp(1, 0, ev.Position)
	}
}

// TouchDown handles the down action for mobile touch events.
func (t *Terminal) TouchDown(ev *mobile.TouchEvent) {
	if t.onMouseDown != nil {
		t.onMouseDown(1, 0, ev.Position)
	}
}

// TouchUp handles the up action for mobile touch events.
func (t *Terminal) TouchUp(ev *mobile.TouchEvent) {
	if t.onMouseUp != nil {
		t.onMouseUp(1, 0, ev.Position)
	}
}

func (t *Terminal) onConfigure() {
	t.listenerLock.Lock()
	for _, l := range t.listeners {
		select {
		case l <- t.config:
		default:
			// channel blocked, might be closed
		}
	}
	t.listenerLock.Unlock()
}

func (t *Terminal) open() error {
	in, out, pty, err := t.startPTY()
	if err != nil {
		return err
	}

	t.in, t.out = in, out
	if t.readWriterConfigurator != nil {
		t.out, t.in = t.readWriterConfigurator.SetupReadWriter(out, in)
	}

	t.pty = pty

	t.updatePTYSize()
	return nil
}

// Exit requests that this terminal exits.
// If there are embedded shells it will exit the child one only.
func (t *Terminal) Exit() {
	_, _ = t.Write([]byte{0x4})
}

func (t *Terminal) close() error {
	if t.in != t.pty {
		_ = t.in.Close() // we may already be closed
	}
	if t.pty == nil {
		return nil
	}

	return t.pty.Close()
}

// don't call often - should we cache?
func guessCellSize() fyne.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	min := cell.MinSize()
	return fyne.NewSize(float32(math.Round(float64(min.Width))), float32(math.Round(float64(min.Height))))
}

func (t *Terminal) run() {
	buf := make([]byte, bufLen)
	var leftOver []byte
	for {
		num, err := t.out.Read(buf)
		if err != nil {
			if t.cmd != nil {
				// wait for cmd (shell) to exit, populates ProcessState.ExitCode
				t.cmd.Wait()
			}
			// this is the pre-go 1.13 way to check for the read failing (terminal closed)
			if err.Error() == "EOF" {
				break // term exit on macOS
			} else if err, ok := err.(*os.PathError); ok && err.Err.Error() == "input/output error" {
				break // broken pipe, terminal exit
			}

			fyne.LogError("pty read error", err)
		}

		lenLeftOver := len(leftOver)
		fullBuf := buf
		if lenLeftOver > 0 {
			fullBuf = append(leftOver, buf[:num]...)
			num += lenLeftOver
		}
		leftOver = t.handleOutput(fullBuf[:num])
		if len(leftOver) == 0 {
			t.Refresh()
		}
	}
}

// RunLocalShell starts the terminal by loading a shell and starting to process the input/output.
func (t *Terminal) RunLocalShell() error {
	for t.config.Columns == 0 { // don't load the TTY until our output is configured
		time.Sleep(time.Millisecond * 50)
	}
	err := t.open()
	if err != nil {
		return err
	}

	t.run()

	return t.close()
}

// RunWithConnection starts the terminal by connecting to an external resource like an SSH connection.
func (t *Terminal) RunWithConnection(in io.WriteCloser, out io.Reader) error {
	for t.config.Columns == 0 { // don't load the TTY until our output is configured
		time.Sleep(time.Millisecond * 50)
	}
	t.in, t.out = in, out
	if t.readWriterConfigurator != nil {
		t.out, t.in = t.readWriterConfigurator.SetupReadWriter(out, in)
	}

	t.run()

	return t.close()
}

// Write is used to send commands into an open terminal connection.
// Errors will be returned if the connection is not established, has closed, or there was a problem in transmission.
func (t *Terminal) Write(b []byte) (int, error) {
	if t.in == nil {
		return 0, io.EOF
	}

	return t.in.Write(b)
}

func (t *Terminal) setupShortcuts() {
	var paste fyne.Shortcut
	paste = &desktop.CustomShortcut{KeyName: fyne.KeyV, Modifier: fyne.KeyModifierShift | fyne.KeyModifierShortcutDefault}
	if runtime.GOOS == "darwin" {
		paste = &fyne.ShortcutPaste{} // we look up clipboard later
	}
	t.ShortcutHandler.AddShortcut(paste,
		func(_ fyne.Shortcut) {
			t.pasteText(t.clipboard)
		})
	var shortcutCopy fyne.Shortcut
	shortcutCopy = &desktop.CustomShortcut{KeyName: fyne.KeyC, Modifier: fyne.KeyModifierShift | fyne.KeyModifierShortcutDefault}
	if runtime.GOOS == "darwin" {
		shortcutCopy = &fyne.ShortcutCopy{} // we look up clipboard later
	}

	t.ShortcutHandler.AddShortcut(shortcutCopy,
		func(_ fyne.Shortcut) {
			t.copySelectedText(t.clipboard)
		})
}

func (t *Terminal) startingDir() string {
	if t.startDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}

	return t.startDir
}

// New sets up a new terminal instance with the bash shell
func New() *Terminal {
	t := &Terminal{
		mouseCursor:      desktop.DefaultCursor,
		highlightBitMask: 0x55,
	}
	t.ExtendBaseWidget(t)
	t.content = widget.NewTermGrid()
	t.setupShortcuts()

	return t
}

// sanitizePosition ensures that the given position p is within the bounds of the terminal.
// If the position is outside the bounds, it adjusts the coordinates to the nearest valid values.
// The adjusted position is then returned.
func (t *Terminal) sanitizePosition(p fyne.Position) *fyne.Position {
	size := t.Size()
	width, height := size.Width, size.Height
	if p.X < 0 {
		p.X = 0
	} else if p.X > width {
		p.X = width
	}

	if p.Y < 0 {
		p.Y = 0
	} else if p.Y > height {
		p.Y = height
	}

	return &p
}

// Dragged is called by fyne when the left mouse is down and moved whilst over the widget.
func (t *Terminal) Dragged(d *fyne.DragEvent) {
	pos := t.sanitizePosition(d.Position)
	if !t.selecting {
		if t.keyboardState.altPressed {
			t.blockMode = true
		}
		p := t.getTermPosition(*pos)
		t.selStart = &p
		t.selEnd = nil
	}
	// clear any previous selection
	sr, sc, er, ec := t.getSelectedRange()
	widget.ClearHighlightRange(t.content, t.blockMode, sr, sc, er, ec)

	// make sure that x,y,x1,y1 are always positive
	t.selecting = true
	t.mouseCursor = desktop.TextCursor
	p := t.getTermPosition(*pos)
	t.selEnd = &p
	t.highlightSelectedText()
}

// DragEnd is called by fyne when the left mouse is released after a Drag event.
func (t *Terminal) DragEnd() {
	t.selecting = false
}

// SetReadWriter sets the readWriterConfigurator function that will be used when creating a new terminal.
// The readWriterConfigurator function is responsible for setting up the I/O readers and writers.
func (t *Terminal) SetReadWriter(mw ReadWriterConfigurator) {
	t.readWriterConfigurator = mw
}

func (t *Terminal) SetClipboard(clipboard fyne.Clipboard) {
	t.clipboard = clipboard
}

// ReadWriterConfigurator is an interface that defines the methods required to set up
// the input (reader) and output (writer) streams for the terminal.
// Implementations of this interface can modify or wrap the reader and writer.
type ReadWriterConfigurator interface {
	// SetupReadWriter configures the input and output streams for the terminal.
	// It takes an input reader (r) and an output writer (w) as arguments.
	// The function returns a possibly modified reader and writer that
	// the terminal will use for I/O operations.
	SetupReadWriter(r io.Reader, w io.WriteCloser) (io.Reader, io.WriteCloser)
}

// ReadWriterConfiguratorFunc is a function type that matches the signature of the
// SetupReadWriter method in the Middleware interface.
type ReadWriterConfiguratorFunc func(r io.Reader, w io.WriteCloser) (io.Reader, io.WriteCloser)

// SetupReadWriter allows ReadWriterConfiguratorFunc to satisfy the Middleware interface.
// It calls the ReadWriterConfiguratorFunc itself.
func (m ReadWriterConfiguratorFunc) SetupReadWriter(r io.Reader, w io.WriteCloser) (io.Reader, io.WriteCloser) {
	return m(r, w)
}