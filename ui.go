package main

import (
	"fmt"

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
var showui = true
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
	lines = append(lines, "Frame")
	lines = append(lines, "[Settings]:")
	lines = append(lines, fmt.Sprintf("Text Brightness: <%2.3f>", text_brightness))
	lines = append(lines, fmt.Sprintf("Scanline Strength: <%2.3f>", scanline_strength))
	lines = append(lines, fmt.Sprintf("Noise Strength: <%2.3f>", noiseStrength))
	lines = append(lines, fmt.Sprintf("Ambient Light: <%2.3f>", ambient))
	lines = append(lines, fmt.Sprintf("Bloom Strength: <%2.3f>", bloomStrength))
	lines = append(lines, fmt.Sprintf("Bloom Brightness: <%2.3f>", bloomBrightness))

	lines[selected+2] = "> " + lines[selected+2]

	return lines
}
