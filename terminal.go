package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/faiface/beep"
	"golang.org/x/image/colornames"
)

type TermColor struct {
	foreground int
	background int
}

func (tc TermColor) ForegroundRGBA() color.RGBA {
	return colornames.Wheat
}
func (tc TermColor) BackgroundRGBA() color.RGBA {
	return colornames.Black
}

type Cell struct {
	style TermColor
	char  string
}
type termHandler struct {
	buffer, alternate [][]Cell
	useAlternate      bool

	cursorX, cursorY int

	charWidth, charHeight int

	defFG, defBG  int
	cursorEnabled bool

	bellSound beep.StreamSeekCloser
}

func safePrintAns(ansi string) {
	s := ""
	for _, b := range ansi {
		if b == 0x1b {
			s += "ESC"
		} else {
			s += string(b)
		}
	}
	log.Println(s)
}
func NewTerminal(width, height int, char_width, char_height int) *termHandler {
	buf := make([][]Cell, height)
	alt := make([][]Cell, height)
	for i := range buf {
		buf[i] = make([]Cell, width)
		alt[i] = make([]Cell, width)
	}

	mw := &termHandler{
		buffer:        buf,
		alternate:     alt,
		useAlternate:  false,
		cursorX:       0,
		cursorY:       0,
		charWidth:     char_width,
		charHeight:    char_height,
		cursorEnabled: true,
		defFG:         10,
		defBG:         12,
	}
	return mw
}

func fillRect(rect image.Rectangle, img *image.RGBA, color color.RGBA) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.SetRGBA(x, y, color)
		}
	}
}

func (th termHandler) DrawToImage(img *image.RGBA) {
	opBuffer := th.buffer //buffer to operate on
	//if th.useAlternate {
	//	opBuffer = th.buffer
	//}

	for y := range opBuffer {
		for x, cell := range opBuffer[y] {
			startx := x*th.charWidth + int(term_borders_dims[0])
			starty := y*th.charHeight + int(term_borders_dims[1])
			fg, bg := cell.style.ForegroundRGBA(), cell.style.BackgroundRGBA()
			// Cursor
			if th.cursorEnabled {
				if x == th.cursorX && y == th.cursorY {
					fg, bg = bg, fg
				}
			}

			fillRect(image.Rect(startx, starty, startx+th.charWidth, starty+th.charHeight-1), img, bg)
			if cell.char == ("\x00") {
				continue
			}
			addLabel(img, startx, starty+th.charHeight-2, cell.char, fg)

		}
	}

}

func (mw *termHandler) WriteChar(x, y int, ch string) {
	mw.buffer[y][x].char = ch
}
func (mw *termHandler) SetCursor(x, y int) {
	mw.cursorX = x
	mw.cursorY = y
}
func (mw *termHandler) eraseBuffer() {
	for y := range mw.buffer {
		mw.eraseLine(y)
	}
}
func (mw *termHandler) eraseLine(y int) {
	if y < 0 || y >= len(mw.buffer) {
		return
	}
	for x := range mw.buffer[y] {
		mw.buffer[y][x].style.background = mw.defBG
		mw.buffer[y][x].style.foreground = mw.defFG
		mw.buffer[y][x].char = ""
	}
}

func (mw *termHandler) eraseAfterCursor() {
	for y := mw.cursorY; y < len(mw.buffer); y++ {
		startx := 0
		if y == mw.cursorY {
			startx = mw.cursorX
		}
		for x := startx; x < len(mw.buffer[0]); x++ {
			mw.buffer[y][x].style.background = mw.defBG
			mw.buffer[y][x].style.foreground = mw.defFG
			mw.buffer[y][x].char = ""
		}
	}
}

func (mw *termHandler) HandleEscape(code string) {
	code = code[1:]

	if code == "[?2004h" || code == "[?2004l" {
		//ignore
		return
	}
	//UP
	if code == "[A" {
		mw.cursorY--
		mw.safeCursor()
		return
	}
	if code[:1] == "[" && code[len(code)-1:] == "A" {
		deltay := 0
		n, err := fmt.Sscanf(code, "[%dA", &deltay)
		if n != 1 || err != nil {
			safePrintAns(code)
			log.Fatal(err)
		}
		mw.cursorY -= deltay
		mw.safeCursor()
		return
	}
	//DOWN
	if code == "[B" {
		mw.cursorY++
		mw.safeCursor()
		return
	}
	if code[:1] == "[" && code[len(code)-1:] == "B" {
		deltay := 0
		n, err := fmt.Sscanf(code, "[%dB", &deltay)
		if n != 1 || err != nil {
			safePrintAns(code)
			log.Fatal(err)
		}
		mw.cursorY += deltay
		mw.safeCursor()
		return
	}
	//RIGHT
	if code == "[C" {
		mw.cursorX++
		mw.safeCursor()
		return
	}
	if code[:1] == "[" && code[len(code)-1:] == "C" {
		deltax := 0
		n, err := fmt.Sscanf(code, "[%dC", &deltax)
		if n != 1 || err != nil {
			safePrintAns(code)
			log.Fatal(err)
		}
		mw.cursorX += deltax
		mw.safeCursor()
		return
	}
	//LEFT
	if code == "[D" {
		mw.cursorX--
		mw.safeCursor()
		return
	}
	if code[:1] == "[" && code[len(code)-1:] == "D" {
		deltay := 0
		n, err := fmt.Sscanf(code, "[%dD", &deltay)
		if n != 1 || err != nil {
			safePrintAns(code)
			log.Fatal(err)
		}
		mw.cursorX -= deltay
		mw.safeCursor()
		return
	}

	if code == "[J" || code == "[0J" {
		mw.eraseAfterCursor()
		return
	}
	if code == "[2J" || code == "[3J" {
		mw.eraseBuffer()
		return
	}
	if code == "[H" {
		mw.SetCursor(0, 0)
		return
	}
	if code[len(code)-1:] == "H" && code[:1] == "[" {
		y := 0
		x := 0
		n, err := fmt.Sscanf(code, "[%d;%d", &x, &y)
		if n != 2 || err != nil {
			log.Println("Not read")
			log.Fatal(err)
		}
		mw.cursorX = x
		mw.cursorY = y
		mw.safeCursor()
		return
	}
	if code == "[?25h" {
		mw.cursorEnabled = true
		return
	}
	if code == "[?25l" {
		mw.cursorEnabled = false
		return
	}
	if code == "[?1049h" {
		//use alternate
		mw.useAlternate = true
		return
	}
	if code == "[?1049l" {
		//use default
		mw.useAlternate = false
		return
	}

	if code == "[K" || code == "[0K" {
		mw.eraseAfterCursor()
		return
	}
	if code == "[2K" {
		mw.eraseLine(mw.cursorY)
		return
	}

	if code[:1] == "[" && code[len(code)-1:] == "G" {
		x := 0
		n, err := fmt.Sscanf(code, "[%dG", &x)
		if n != 1 || err != nil {
			log.Println("Not read")
			safePrintAns(code)
			log.Fatal(err)
		}
		mw.cursorX = x
		mw.safeCursor()
		return
	}

	safePrintAns(code)
}
func (mw *termHandler) safeCursor() {
	if mw.cursorX < 0 {
		mw.cursorX = 0
	}
	if mw.cursorY < 0 {
		mw.cursorY = 0
	}

	if mw.cursorX > len(mw.buffer[0]) {
		mw.cursorX = len(mw.buffer[0]) - 1
	}
	if mw.cursorY > len(mw.buffer) {
		mw.cursorY = len(mw.buffer) - 1
	}
}
func (mw *termHandler) ScrollDown() {
	mw.buffer = append(mw.buffer[1:], make([]Cell, len(mw.buffer[0])))
}
func (mw *termHandler) Write(bs []byte) (int, error) {
	line := string(bs)
	ansiIndices := FindAnsiIndex(line)
	var ansiIndex = 0
	for i := 0; i < len(line); i++ {
		//handle then skip ansi escape codes
		if ansiIndex < len(ansiIndices) { //Only care if any are left
			startAnsi := ansiIndices[ansiIndex][0]
			endAnsi := ansiIndices[ansiIndex][1]
			if i == startAnsi { //check if this position is the start of an escape code
				code := line[i:endAnsi]
				mw.HandleEscape(code)

				//safePrintAns("Line: " + line)
				//log.Println(ansiIndices)
				i = endAnsi
				ansiIndex++
				if ansiIndex < len(ansiIndices) && i == ansiIndices[ansiIndex][0] {
					i -= 1
					//handle that next iteration
				}
				continue
			}
		}
		var TabSize = 8
		char := line[i]
		//Special codes
		if char == '\n' {
			mw.cursorY++
			if mw.cursorY >= len(mw.buffer) {
				mw.ScrollDown()
				mw.cursorY--
			}
			continue
		} else if char == '\r' {
			mw.cursorX = 0
			continue
		} else if char == '\t' {
			toAdd := TabSize - (mw.cursorX % TabSize)
			mw.cursorX += toAdd
			mw.safeCursor()
			continue
		} else if char == 0x07 {

			continue
		} else if char == 0x08 {
			//Backspace
			log.Println("Backspace")
			mw.cursorX -= 1
			mw.safeCursor()
			continue
		}

		//knownset := " \"'!@#$%^&*(){}[]/\\|"
		//bounds check buffer access
		if mw.cursorY < len(mw.buffer) {
			if mw.cursorX < len(mw.buffer[0]) { //no line wrapping yet
				if mw.useAlternate {
					mw.alternate[mw.cursorY][mw.cursorX].char = string(line[i])
				} else {
					mw.buffer[mw.cursorY][mw.cursorX].char = string(line[i])
				}
				mw.cursorX++

			}
		}

	}

	//log.Println("Cursor x,y", mw.cursorX, mw.cursorY)
	return len(bs), nil
}
