package main

import (
	"fmt"
	"os"

	"github.com/go-gl/glfw/v3.3/glfw"
)

var (
	text_brightness   float32 = .5
	scanline_strength float32 = .13
	noiseStrength     float32 = .03
	ambient           float32 = .12
	bloomStrength     float32 = 1.2
	bloomBrightness   float32 = 1.4
)
var showui = false
var selected int = 1
var selections = map[int]*float32{
	0: &text_brightness,
	1: &scanline_strength,
	2: &noiseStrength,
	3: &ambient,
	4: &bloomStrength,
	5: &bloomBrightness,
}

func keyCall(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyS && action == glfw.Press && mods == glfw.ModControl {
		showui = !showui
		fmt.Println("switch", showui)
	}
	if key == glfw.KeyQ && mods == glfw.ModControl && action == glfw.Press {
		os.Exit(0)
	}

	if showui {
		if key == glfw.KeyUp && action == glfw.Release {
			prevSelection()
		}
		if key == glfw.KeyDown && action == glfw.Release {
			nextSelection()
		}
		if key == glfw.KeyLeft && action == glfw.Release {
			lowerSelection()
		}
		if key == glfw.KeyRight && action == glfw.Release {
			incrementSelection()
		}
	} else {
		if key == glfw.KeyBackspace && action == glfw.Press {
			term_pty.Write([]byte{0x08})
		}

		for keycomb, str := range keymap {
			if key == keycomb.key && mods == keycomb.mod && action == glfw.Press {
				term_pty.WriteString(str)
			}
		}
	}

}

type keyCombo struct {
	key glfw.Key
	mod glfw.ModifierKey
}

var keymap = map[keyCombo]string{
	{glfw.KeyLeft, 0}:  "\x1b[1D",
	{glfw.KeyRight, 0}: "\x1b[1C",
	{glfw.KeyUp, 0}:    "\x1b[1A",
	{glfw.KeyDown, 0}:  "\x1b[1B",

	{glfw.KeyEnter, 0}:             "\r",
	{glfw.KeyTab, 0}:               "\t",
	{glfw.KeySpace, 0}:             " ",
	{glfw.KeyMinus, 0}:             "-",
	{glfw.KeyMinus, glfw.ModShift}: "_",

	{glfw.KeyEqual, 0}:             "=",
	{glfw.KeyEqual, glfw.ModShift}: "+",

	{glfw.KeyComma, 0}:              ",",
	{glfw.KeyPeriod, 0}:             ".",
	{glfw.KeyComma, glfw.ModShift}:  ",",
	{glfw.KeyPeriod, glfw.ModShift}: ">",

	{glfw.KeySlash, 0}:             "/",
	{glfw.KeySlash, glfw.ModShift}: "?",

	{glfw.KeyApostrophe, 0}:             "'",
	{glfw.KeyApostrophe, glfw.ModShift}: "\"",

	{glfw.KeyBackslash, 0}:             "\\",
	{glfw.KeyBackslash, glfw.ModShift}: "|",

	{glfw.KeyLeftBracket, 0}:             "[",
	{glfw.KeyLeftBracket, glfw.ModShift}: "{",

	{glfw.KeyRightBracket, 0}:             "]",
	{glfw.KeyRightBracket, glfw.ModShift}: "}",

	{glfw.KeySemicolon, 0}:             ";",
	{glfw.KeySemicolon, glfw.ModShift}: ":",

	{glfw.KeyC, glfw.ModControl}: string([]byte{0x03}),
	{glfw.KeyD, glfw.ModControl}: string([]byte{0x04}),
	{glfw.KeyZ, glfw.ModControl}: string([]byte{0x04}),

	{glfw.KeySlash, 0}: "/",

	{glfw.Key0, 0}: "0",
	{glfw.Key1, 0}: "1",
	{glfw.Key2, 0}: "2",
	{glfw.Key3, 0}: "3",
	{glfw.Key4, 0}: "4",
	{glfw.Key5, 0}: "5",
	{glfw.Key6, 0}: "6",
	{glfw.Key7, 0}: "7",
	{glfw.Key8, 0}: "8",
	{glfw.Key9, 0}: "9",

	{glfw.Key0, glfw.ModShift}: ")",
	{glfw.Key1, glfw.ModShift}: "!",
	{glfw.Key2, glfw.ModShift}: "@",
	{glfw.Key3, glfw.ModShift}: "#",
	{glfw.Key4, glfw.ModShift}: "$",
	{glfw.Key5, glfw.ModShift}: "%",
	{glfw.Key6, glfw.ModShift}: "^",
	{glfw.Key7, glfw.ModShift}: "&",
	{glfw.Key8, glfw.ModShift}: "*",
	{glfw.Key9, glfw.ModShift}: "(",

	{glfw.KeyA, 0}: "a",
	{glfw.KeyB, 0}: "b",
	{glfw.KeyC, 0}: "c",
	{glfw.KeyD, 0}: "d",
	{glfw.KeyE, 0}: "e",
	{glfw.KeyF, 0}: "f",
	{glfw.KeyG, 0}: "g",
	{glfw.KeyH, 0}: "h",
	{glfw.KeyI, 0}: "i",
	{glfw.KeyJ, 0}: "j",
	{glfw.KeyK, 0}: "k",
	{glfw.KeyL, 0}: "l",
	{glfw.KeyM, 0}: "m",
	{glfw.KeyN, 0}: "n",
	{glfw.KeyO, 0}: "o",
	{glfw.KeyP, 0}: "p",
	{glfw.KeyQ, 0}: "q",
	{glfw.KeyR, 0}: "r",
	{glfw.KeyS, 0}: "s",
	{glfw.KeyT, 0}: "t",
	{glfw.KeyU, 0}: "u",
	{glfw.KeyV, 0}: "v",
	{glfw.KeyW, 0}: "w",
	{glfw.KeyX, 0}: "x",
	{glfw.KeyY, 0}: "y",
	{glfw.KeyZ, 0}: "z",

	{glfw.KeyA, glfw.ModShift}: "A",
	{glfw.KeyB, glfw.ModShift}: "B",
	{glfw.KeyC, glfw.ModShift}: "C",
	{glfw.KeyD, glfw.ModShift}: "D",
	{glfw.KeyE, glfw.ModShift}: "E",
	{glfw.KeyF, glfw.ModShift}: "F",
	{glfw.KeyG, glfw.ModShift}: "G",
	{glfw.KeyH, glfw.ModShift}: "H",
	{glfw.KeyI, glfw.ModShift}: "I",
	{glfw.KeyJ, glfw.ModShift}: "J",
	{glfw.KeyK, glfw.ModShift}: "K",
	{glfw.KeyL, glfw.ModShift}: "L",
	{glfw.KeyM, glfw.ModShift}: "M",
	{glfw.KeyN, glfw.ModShift}: "N",
	{glfw.KeyO, glfw.ModShift}: "O",
	{glfw.KeyP, glfw.ModShift}: "P",
	{glfw.KeyQ, glfw.ModShift}: "Q",
	{glfw.KeyR, glfw.ModShift}: "R",
	{glfw.KeyS, glfw.ModShift}: "S",
	{glfw.KeyT, glfw.ModShift}: "T",
	{glfw.KeyU, glfw.ModShift}: "U",
	{glfw.KeyV, glfw.ModShift}: "V",
	{glfw.KeyW, glfw.ModShift}: "W",
	{glfw.KeyX, glfw.ModShift}: "X",
	{glfw.KeyY, glfw.ModShift}: "Y",
	{glfw.KeyZ, glfw.ModShift}: "Z",
}

func incrementSelection() {
	sel := selections[selected]
	if sel == nil {
		return
	}
	*sel += float32(.01)
}
func lowerSelection() {
	sel := selections[selected]
	if sel == nil {
		return
	}
	*sel -= float32(.01)
}

func nextSelection() {
	if selected < len(selections)-1 {
		selected++
	}
}
func prevSelection() {
	if selected > 0 {
		selected--
	}
}
func MakeUI() []string {
	lines := []string{}
	lines = append(lines, "[Settings]:")
	lines = append(lines, fmt.Sprintf("Text Brightness: <%2.3f>", text_brightness))
	lines = append(lines, fmt.Sprintf("Scanline Strength: <%2.3f>", scanline_strength))
	lines = append(lines, fmt.Sprintf("Noise Strength: <%2.3f>", noiseStrength))
	lines = append(lines, fmt.Sprintf("Ambient Light: <%2.3f>", ambient))
	lines = append(lines, fmt.Sprintf("Bloom Strength: <%2.3f>", bloomStrength))
	lines = append(lines, fmt.Sprintf("Bloom Brightness: <%2.3f>", bloomBrightness))

	lines[selected+1] = "> " + lines[min(selected+1, len(lines)-1)]

	return lines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
