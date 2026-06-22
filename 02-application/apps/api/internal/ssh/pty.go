package ssh

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// PTY represents a pseudo-terminal with full control
type PTY struct {
	Master *os.File
	Slave  *os.File
	Width  int
	Height int
}

// NewPTY creates a new pseudo-terminal (alias for OpenPTY).
func NewPTY(width, height int) (*PTY, error) {
	return OpenPTY(width, height)
}

// OpenPTY creates a new pseudo-terminal
func OpenPTY(width, height int) (*PTY, error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("PTY not supported on %s", runtime.GOOS)
	}

	// Open master
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open /dev/ptmx: %w", err)
	}

	// Grant access
	if err := grantpt(int(master.Fd())); err != nil {
		master.Close()
		return nil, fmt.Errorf("grantpt failed: %w", err)
	}

	// Unlock
	if err := unlockpt(int(master.Fd())); err != nil {
		master.Close()
		return nil, fmt.Errorf("unlockpt failed: %w", err)
	}

	// Get slave name
	slaveName, err := ptsname(int(master.Fd()))
	if err != nil {
		master.Close()
		return nil, fmt.Errorf("ptsname failed: %w", err)
	}

	// Open slave
	slave, err := os.OpenFile(slaveName, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		master.Close()
		return nil, fmt.Errorf("failed to open slave: %w", err)
	}

	// Set terminal size
	if err := setTerminalSize(master, width, height); err != nil {
		master.Close()
		slave.Close()
		return nil, fmt.Errorf("failed to set terminal size: %w", err)
	}

	return &PTY{
		Master: master,
		Slave:  slave,
		Width:  width,
		Height: height,
	}, nil
}

// Resize updates PTY dimensions
func (p *PTY) Resize(width, height int) error {
	if err := setTerminalSize(p.Master, width, height); err != nil {
		return err
	}
	p.Width = width
	p.Height = height
	return nil
}

// Close closes the PTY
func (p *PTY) Close() error {
	if p.Slave != nil {
		p.Slave.Close()
	}
	if p.Master != nil {
		p.Master.Close()
	}
	return nil
}

// setTerminalSize sets the terminal dimensions using TIOCSWINSZ
func setTerminalSize(f *os.File, width, height int) error {
	ws := &unix.Winsize{
		Row: uint16(height),
		Col: uint16(width),
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		unix.TIOCSWINSZ,
		uintptr(unsafe.Pointer(ws)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// grantpt grants access to the slave pseudo-terminal
func grantpt(fd int) error {
	_, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	return err
}

// unlockpt unlocks the slave pseudo-terminal
func unlockpt(fd int) error {
	var unlock int32
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		unix.TIOCSPTLCK,
		uintptr(unsafe.Pointer(&unlock)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// ptsname returns the name of the slave pseudo-terminal
func ptsname(fd int) (string, error) {
	n, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/dev/pts/%d", n), nil
}

// ScreenBuffer represents a terminal screen buffer
type ScreenBuffer struct {
	cells    [][]Cell
	width    int
	height   int
	cursorX  int
	cursorY  int
	scrollback [][]Cell
	maxScrollback int
	mu       sync.RWMutex
}

// Cell represents a single terminal cell
type Cell struct {
	Char       rune
	FgColor    [3]byte // RGB
	BgColor    [3]byte // RGB
	Bold       bool
	Italic     bool
	Underline  bool
	Blink      bool
	Reverse    bool
	Hidden     bool
}

// Default cell colors
var (
	DefaultFg = [3]byte{0xC0, 0xC0, 0xC0} // Light gray
	DefaultBg = [3]byte{0x00, 0x00, 0x00} // Black
)

// NewScreenBuffer creates a new screen buffer
func NewScreenBuffer(width, height, maxScrollback int) *ScreenBuffer {
	if maxScrollback < 100 {
		maxScrollback = 10000
	}

	cells := make([][]Cell, height)
	for i := range cells {
		cells[i] = make([]Cell, width)
		for j := range cells[i] {
			cells[i][j] = Cell{
				Char:    ' ',
				FgColor: DefaultFg,
				BgColor: DefaultBg,
			}
		}
	}

	return &ScreenBuffer{
		cells:         cells,
		width:         width,
		height:        height,
		maxScrollback: maxScrollback,
		scrollback:    make([][]Cell, 0, maxScrollback),
	}
}

// Resize resizes the screen buffer
func (sb *ScreenBuffer) Resize(width, height int) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Save current content to scrollback if height decreases
	if height < sb.height {
		for i := 0; i < sb.height-height && i < sb.height; i++ {
			sb.addToScrollback(sb.cells[i])
		}
	}

	newCells := make([][]Cell, height)
	for i := range newCells {
		newCells[i] = make([]Cell, width)
		for j := range newCells[i] {
			if i < sb.height && j < sb.width {
				newCells[i][j] = sb.cells[i][j]
			} else {
				newCells[i][j] = Cell{
					Char:    ' ',
					FgColor: DefaultFg,
					BgColor: DefaultBg,
				}
			}
		}
	}

	sb.cells = newCells
	sb.width = width
	sb.height = height

	// Clamp cursor
	if sb.cursorX >= width {
		sb.cursorX = width - 1
	}
	if sb.cursorY >= height {
		sb.cursorY = height - 1
	}
}

// addToScrollback adds a line to scrollback buffer
func (sb *ScreenBuffer) addToScrollback(line []Cell) {
	if len(sb.scrollback) >= sb.maxScrollback {
		sb.scrollback = sb.scrollback[1:]
	}
	// Copy line
	copied := make([]Cell, len(line))
	copy(copied, line)
	sb.scrollback = append(sb.scrollback, copied)
}

// WriteChar writes a character at current cursor position
func (sb *ScreenBuffer) WriteChar(ch rune) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.cursorY >= sb.height {
		// Scroll up
		sb.addToScrollback(sb.cells[0])
		copy(sb.cells, sb.cells[1:])
		sb.cursorY = sb.height - 1
		// Clear new line
		for i := range sb.cells[sb.cursorY] {
			sb.cells[sb.cursorY][i] = Cell{Char: ' ', FgColor: DefaultFg, BgColor: DefaultBg}
		}
	}

	if sb.cursorX < sb.width {
		sb.cells[sb.cursorY][sb.cursorX] = Cell{
			Char:    ch,
			FgColor: DefaultFg,
			BgColor: DefaultBg,
		}
		sb.cursorX++
	}
}

// MoveCursor moves cursor to position
func (sb *ScreenBuffer) MoveCursor(x, y int) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	
	if x >= 0 && x < sb.width {
		sb.cursorX = x
	}
	if y >= 0 && y < sb.height {
		sb.cursorY = y
	}
}

// Clear clears the screen
func (sb *ScreenBuffer) Clear() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	for i := range sb.cells {
		for j := range sb.cells[i] {
			sb.cells[i][j] = Cell{Char: ' ', FgColor: DefaultFg, BgColor: DefaultBg}
		}
	}
	sb.cursorX = 0
	sb.cursorY = 0
}

// GetContent returns current screen content as string
func (sb *ScreenBuffer) GetContent() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	var result strings.Builder
	for i := 0; i < sb.height; i++ {
		for j := 0; j < sb.width; j++ {
			result.WriteRune(sb.cells[i][j].Char)
		}
		if i < sb.height-1 {
			result.WriteByte('\n')
		}
	}
	return result.String()
}

// TerminalEmulator handles ANSI escape sequences
type TerminalEmulator struct {
	screen    *ScreenBuffer
	fgColor   [3]byte
	bgColor   [3]byte
	bold      bool
	italic    bool
	underline bool
	reverse   bool
	escapeSeq []byte
	inEscape  bool
	mu        sync.RWMutex
}

// NewTerminalEmulator creates a new terminal emulator
func NewTerminalEmulator(width, height int) *TerminalEmulator {
	return &TerminalEmulator{
		screen:    NewScreenBuffer(width, height, 10000),
		fgColor:   DefaultFg,
		bgColor:   DefaultBg,
		escapeSeq: make([]byte, 0, 32),
	}
}

// Process processes input data through the terminal emulator
func (te *TerminalEmulator) Process(data []byte) {
	te.mu.Lock()
	defer te.mu.Unlock()

	for _, b := range data {
		if te.inEscape {
			te.escapeSeq = append(te.escapeSeq, b)
			if te.isEscapeComplete() {
				te.processEscapeSequence()
				te.inEscape = false
				te.escapeSeq = te.escapeSeq[:0]
			}
		} else if b == 0x1b { // ESC
			te.inEscape = true
			te.escapeSeq = append(te.escapeSeq, b)
		} else if b == '\r' {
			te.screen.cursorX = 0
		} else if b == '\n' {
			te.screen.cursorY++
			te.screen.cursorX = 0
			if te.screen.cursorY >= te.screen.height {
				te.screen.addToScrollback(te.screen.cells[0])
				copy(te.screen.cells, te.screen.cells[1:])
				te.screen.cursorY = te.screen.height - 1
				for i := range te.screen.cells[te.screen.cursorY] {
					te.screen.cells[te.screen.cursorY][i] = Cell{Char: ' ', FgColor: te.fgColor, BgColor: te.bgColor}
				}
			}
		} else if b == '\t' {
			// Tab - move to next tab stop (every 8 chars)
			te.screen.cursorX = ((te.screen.cursorX / 8) + 1) * 8
			if te.screen.cursorX >= te.screen.width {
				te.screen.cursorX = te.screen.width - 1
			}
		} else if b == '\b' {
			// Backspace
			if te.screen.cursorX > 0 {
				te.screen.cursorX--
			}
		} else if b >= 32 && b < 127 {
			// Printable ASCII
			if te.screen.cursorY < te.screen.height && te.screen.cursorX < te.screen.width {
				te.screen.cells[te.screen.cursorY][te.screen.cursorX] = Cell{
					Char:    rune(b),
					FgColor: te.fgColor,
					BgColor: te.bgColor,
					Bold:    te.bold,
					Italic:  te.italic,
					Underline: te.underline,
					Reverse: te.reverse,
				}
				te.screen.cursorX++
			}
		}
	}
}

// isEscapeComplete checks if escape sequence is complete
func (te *TerminalEmulator) isEscapeComplete() bool {
	if len(te.escapeSeq) < 2 {
		return false
	}

	// CSI sequence: ESC [ ... letter
	if te.escapeSeq[1] == '[' {
		last := te.escapeSeq[len(te.escapeSeq)-1]
		return (last >= 'A' && last <= 'Z') || (last >= 'a' && last <= 'z') || last == '~'
	}

	return true
}

// processEscapeSequence processes a complete escape sequence
func (te *TerminalEmulator) processEscapeSequence() {
	if len(te.escapeSeq) < 2 {
		return
	}

	switch te.escapeSeq[1] {
	case '[':
		te.processCSISequence()
	case ']':
		// OSC - Operating System Command
		// Ignore for now
	case '(':
		// Set character set
		// Ignore for now
	}
}

// processCSISequence processes CSI escape sequences
func (te *TerminalEmulator) processCSISequence() {
	seq := string(te.escapeSeq[2:]) // Skip ESC [
	if len(seq) == 0 {
		return
	}

	cmd := seq[len(seq)-1]
	params := seq[:len(seq)-1]

	switch cmd {
	case 'm': // SGR - Select Graphic Rendition
		te.processSGR(params)
	case 'H', 'f': // Cursor Position
		te.processCursorPosition(params)
	case 'A': // Cursor Up
		n := parseParam(params, 1)
		te.screen.cursorY -= n
		if te.screen.cursorY < 0 {
			te.screen.cursorY = 0
		}
	case 'B': // Cursor Down
		n := parseParam(params, 1)
		te.screen.cursorY += n
		if te.screen.cursorY >= te.screen.height {
			te.screen.cursorY = te.screen.height - 1
		}
	case 'C': // Cursor Forward
		n := parseParam(params, 1)
		te.screen.cursorX += n
		if te.screen.cursorX >= te.screen.width {
			te.screen.cursorX = te.screen.width - 1
		}
	case 'D': // Cursor Backward
		n := parseParam(params, 1)
		te.screen.cursorX -= n
		if te.screen.cursorX < 0 {
			te.screen.cursorX = 0
		}
	case 'J': // Erase Display
		mode := parseParam(params, 0)
		te.eraseDisplay(mode)
	case 'K': // Erase Line
		mode := parseParam(params, 0)
		te.eraseLine(mode)
	case 'r': // Set Scroll Region
		// Ignore for now
	}
}

// processSGR processes Select Graphic Rendition
func (te *TerminalEmulator) processSGR(params string) {
	if params == "" {
		// Reset
		te.fgColor = DefaultFg
		te.bgColor = DefaultBg
		te.bold = false
		te.italic = false
		te.underline = false
		te.reverse = false
		return
	}

	codes := parseParams(params)
	for i := 0; i < len(codes); i++ {
		code := codes[i]
		switch {
		case code == 0:
			te.fgColor = DefaultFg
			te.bgColor = DefaultBg
			te.bold = false
			te.italic = false
			te.underline = false
			te.reverse = false
		case code == 1:
			te.bold = true
		case code == 3:
			te.italic = true
		case code == 4:
			te.underline = true
		case code == 7:
			te.reverse = true
		case code >= 30 && code <= 37:
			te.fgColor = ansiColor(code - 30)
		case code >= 40 && code <= 47:
			te.bgColor = ansiColor(code - 40)
		case code == 38 && i+2 < len(codes) && codes[i+1] == 5:
			// 256 color foreground
			te.fgColor = xterm256Color(codes[i+2])
			i += 2
		case code == 48 && i+2 < len(codes) && codes[i+1] == 5:
			// 256 color background
			te.bgColor = xterm256Color(codes[i+2])
			i += 2
		case code >= 90 && code <= 97:
			// Bright foreground
			te.fgColor = ansiBrightColor(code - 90)
		case code >= 100 && code <= 107:
			// Bright background
			te.bgColor = ansiBrightColor(code - 100)
		}
	}
}

// processCursorPosition processes cursor position sequence
func (te *TerminalEmulator) processCursorPosition(params string) {
	codes := parseParams(params)
	row, col := 1, 1
	if len(codes) > 0 {
		row = codes[0]
	}
	if len(codes) > 1 {
		col = codes[1]
	}
	te.screen.MoveCursor(col-1, row-1)
}

// eraseDisplay erases part or all of the display
func (te *TerminalEmulator) eraseDisplay(mode int) {
	switch mode {
	case 0: // From cursor to end
		for j := te.screen.cursorX; j < te.screen.width; j++ {
			te.screen.cells[te.screen.cursorY][j] = Cell{Char: ' ', FgColor: DefaultFg, BgColor: DefaultBg}
		}
		for i := te.screen.cursorY + 1; i < te.screen.height; i++ {
			for j := 0; j < te.screen.width; j++ {
				te.screen.cells[i][j] = Cell{Char: ' ', FgColor: DefaultFg, BgColor: DefaultBg}
			}
		}
	case 1: // From beginning to cursor
		for j := 0; j <= te.screen.cursorX; j++ {
			te.screen.cells[te.screen.cursorY][j] = Cell{Char: ' ', FgColor: DefaultFg, BgColor: DefaultBg}
		}
		for i := 0; i < te.screen.cursorY; i++ {
			for j := 0; j < te.screen.width; j++ {
				te.screen.cells[i][j] = Cell{Char: ' ', FgColor: DefaultFg, BgColor: DefaultBg}
			}
		}
	case 2, 3: // Entire display
		te.screen.Clear()
	}
}

// eraseLine erases part or all of the line
func (te *TerminalEmulator) eraseLine(mode int) {
	switch mode {
	case 0: // From cursor to end
		for j := te.screen.cursorX; j < te.screen.width; j++ {
			te.screen.cells[te.screen.cursorY][j] = Cell{Char: ' ', FgColor: DefaultFg, BgColor: DefaultBg}
		}
	case 1: // From beginning to cursor
		for j := 0; j <= te.screen.cursorX; j++ {
			te.screen.cells[te.screen.cursorY][j] = Cell{Char: ' ', FgColor: DefaultFg, BgColor: DefaultBg}
		}
	case 2: // Entire line
		for j := 0; j < te.screen.width; j++ {
			te.screen.cells[te.screen.cursorY][j] = Cell{Char: ' ', FgColor: DefaultFg, BgColor: DefaultBg}
		}
	}
}

// parseParam parses a single parameter
func parseParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	var val int
	fmt.Sscanf(s, "%d", &val)
	if val == 0 {
		return defaultVal
	}
	return val
}

// parseParams parses multiple parameters separated by semicolons
func parseParams(s string) []int {
	if s == "" {
		return []int{}
	}
	parts := strings.Split(s, ";")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		var val int
		fmt.Sscanf(p, "%d", &val)
		result = append(result, val)
	}
	return result
}

// ansiColor returns ANSI 8-color palette
func ansiColor(n int) [3]byte {
	colors := [][3]byte{
		{0x00, 0x00, 0x00}, // Black
		{0xCD, 0x00, 0x00}, // Red
		{0x00, 0xCD, 0x00}, // Green
		{0xCD, 0xCD, 0x00}, // Yellow
		{0x00, 0x00, 0xEE}, // Blue
		{0xCD, 0x00, 0xCD}, // Magenta
		{0x00, 0xCD, 0xCD}, // Cyan
		{0xE5, 0xE5, 0xE5}, // White
	}
	if n >= 0 && n < len(colors) {
		return colors[n]
	}
	return DefaultFg
}

// ansiBrightColor returns ANSI bright colors
func ansiBrightColor(n int) [3]byte {
	colors := [][3]byte{
		{0x7F, 0x7F, 0x7F}, // Bright Black
		{0xFF, 0x00, 0x00}, // Bright Red
		{0x00, 0xFF, 0x00}, // Bright Green
		{0xFF, 0xFF, 0x00}, // Bright Yellow
		{0x5C, 0x5C, 0xFF}, // Bright Blue
		{0xFF, 0x00, 0xFF}, // Bright Magenta
		{0x00, 0xFF, 0xFF}, // Bright Cyan
		{0xFF, 0xFF, 0xFF}, // Bright White
	}
	if n >= 0 && n < len(colors) {
		return colors[n]
	}
	return DefaultFg
}

// xterm256Color returns xterm 256 color
func xterm256Color(n int) [3]byte {
	if n < 16 {
		if n < 8 {
			return ansiColor(n)
		}
		return ansiBrightColor(n - 8)
	}
	if n < 232 {
		// 6x6x6 color cube
		n -= 16
		r := (n / 36) * 51
		g := ((n % 36) / 6) * 51
		b := (n % 6) * 51
		return [3]byte{byte(r), byte(g), byte(b)}
	}
	// Grayscale
	gray := (n - 232) * 10 + 8
	return [3]byte{byte(gray), byte(gray), byte(gray)}
}

// GetScreen returns the screen buffer for reading
func (te *TerminalEmulator) GetScreen() *ScreenBuffer {
	return te.screen
}

// Resize resizes the terminal
func (te *TerminalEmulator) Resize(width, height int) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.screen.Resize(width, height)
}
