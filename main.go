package main

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func addLabel(img *image.RGBA, x, y int, label string, col color.RGBA) {
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

var dims = [2]int{512 * 2, 256 * 2}
var terminal_dims = [2]int{512, 256}
var char_dims = [2]int{7, 13}

func drawStringToNewTex(str string, operatingImage *image.RGBA, col color.RGBA) {
	addLabel(operatingImage, 0, 13, str, col)
}

var clear_col = []float32{1, 1, 1, 1}
var lines = []string{"This is the first line", "this is the second line"}

func texFromImage(img *image.RGBA) uint32 {
	var TEXTURE_WIDTH, TEXTURE_HEIGHT int32 = int32(img.Rect.Dx()), int32(img.Rect.Dy())

	var textureHandle uint32
	gl.GenTextures(1, &textureHandle)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, textureHandle)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, TEXTURE_WIDTH, TEXTURE_HEIGHT, 0, gl.RGBA,
		gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

	return textureHandle
}

var vbo, vao uint32

//go:embed Shaders/full_screen_quad.frag
var fragSrc string

//go:embed Shaders/full_screen_quad.vert
var vertSrc string

func main() {
	window := initWindow()
	defer cleanupWindow(window)

	glfw.SwapInterval(0)

	vbo, vao = screenVBOVAO()
	screenProg, err := BuildProgram(fragSrc, vertSrc)
	check(err)

	img := image.NewRGBA(image.Rect(0, 0, terminal_dims[0], terminal_dims[1]))
	clearImage(img, color.RGBA{0, 0, 0, 255})
	var frame_num = 0
	for !window.ShouldClose() {
		frame_num++

		clearImage(img, color.RGBA{0, 0, 0, 255})
		drawStringToNewTex(fmt.Sprint("Frame:", frame_num), img, color.RGBA{0, 0, 255, 255})
		imgHandle := texFromImage(img)

		gl.ClearColor(clear_col[0], clear_col[1], clear_col[2], clear_col[3])
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.BindVertexArray(vao)
		gl.UseProgram(screenProg)

		gl.BindTexture(gl.TEXTURE_2D, imgHandle)

		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		glfw.PollEvents()
		window.SwapBuffers()
	}
}

func clearImage(img *image.RGBA, col color.RGBA) {
	for i := range img.Pix {
		switch i % 4 {
		case 0:
			img.Pix[i] = col.R

		case 1:
			img.Pix[i] = col.G
		case 2:
			img.Pix[i] = col.B
		case 3:
			img.Pix[i] = col.A
		}
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
