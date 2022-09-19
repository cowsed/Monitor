package main

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var myFace font.Face

func addLabel(img *image.RGBA, x, y int, label string, col color.RGBA) {
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: myFace,
		Dot:  point,
	}
	d.DrawString(label)
}

var term_cells = [2]int{50, 24}
var char_dims = [2]int{7, 13}
var term_borders_dims = [2]int{8, 8}
var screen_border_dims = [2]int{6, 6}
var terminal_dims = [2]int{term_cells[0] * char_dims[0], term_cells[1] * char_dims[1]}

var win_dims = [2]int{(terminal_dims[0] + term_borders_dims[0]) * 5 / 2, (terminal_dims[1] + term_borders_dims[1]) * 5 / 2}

func drawStringToNewTex(strs []string, operatingImage *image.RGBA, col color.RGBA) {
	for i, str := range strs {
		addLabel(operatingImage, term_borders_dims[0], term_borders_dims[1]+13+i*13, str, col)
	}
}

var clear_col = []float32{1, 1, 1, 1}

func texFromImage(img *image.RGBA) uint32 {
	var TEXTURE_WIDTH, TEXTURE_HEIGHT int32 = int32(img.Rect.Dx()), int32(img.Rect.Dy())

	var textureHandle uint32
	gl.GenTextures(1, &textureHandle)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, textureHandle)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, TEXTURE_WIDTH, TEXTURE_HEIGHT, 0, gl.RGBA,
		gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

	return textureHandle
}

func overwriteTexFromImage(img *image.RGBA, textureHandle uint32) {
	var TEXTURE_WIDTH, TEXTURE_HEIGHT int32 = int32(img.Rect.Dx()), int32(img.Rect.Dy())

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, textureHandle)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, TEXTURE_WIDTH, TEXTURE_HEIGHT, 0, gl.RGBA,
		gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

}

var vbo, vao uint32

//go:embed Shaders/full_screen_quad.frag
var fragSrc string

//go:embed Shaders/full_screen_quad.vert
var vertSrc string

func main() {

	//ttfFile, err := os.Open("Fonts/PxPlus_IBM_VGA9.ttf")
	//check(err)
	//defer ttfFile.Close()
	//bys, err := io.ReadAll(ttfFile)
	//check(err)
	//font, err := truetype.Parse(bys)
	//check(err)
	myFace = basicfont.Face7x13

	window := initWindow()
	defer cleanupWindow(window)

	vbo, vao = screenVBOVAO()
	screenProg, err := BuildProgram(fragSrc, vertSrc)
	check(err)

	img := image.NewRGBA(image.Rect(0, 0, terminal_dims[0]+2*term_borders_dims[0], terminal_dims[1]+2*term_borders_dims[1]))
	clearImage(img, color.RGBA{0, 0, 0, 255})
	texHandle := texFromImage(img)

	var lines = []string{"", "\xDB", "this is the second line", "!@#$%^&*()-=_+", "0123456789012345678901234567890123456789", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a"}
	drawStringToNewTex(lines, img, color.RGBA{255, 255, 255, 255})
	f, err := os.Create("out.png")
	check(err)
	err = png.Encode(f, img)
	check(err)
	f.Close()

	var frame_num = 0
	var scanline_pos int32 = 0

	fmt.Println(img.Bounds())

	//xos4 termius is a good font
	for !window.ShouldClose() {
		frame_num++
		scanline_pos += 501
		scanline_pos %= int32(terminal_dims[0] * terminal_dims[1])

		clearImage(img, color.RGBA{0, 0, 0, 255})
		lines[0] = fmt.Sprint("Frame:", frame_num)
		drawStringToNewTex(lines, img, color.RGBA{255, 255, 255, 255})
		//imgHandle := texFromImage(img)
		overwriteTexFromImage(img, texHandle)

		gl.ClearColor(clear_col[0], clear_col[1], clear_col[2], clear_col[3])
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.BindVertexArray(vao)
		gl.UseProgram(screenProg)
		gl.Uniform2i(gl.GetUniformLocation(screenProg, gl.Str("screen_dims\x00")), int32(terminal_dims[0]), int32(terminal_dims[1]))
		gl.Uniform1i(gl.GetUniformLocation(screenProg, gl.Str("ScanlinePosition\x00")), scanline_pos)

		gl.BindTexture(gl.TEXTURE_2D, texHandle)

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
