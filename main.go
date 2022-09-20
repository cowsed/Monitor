package main

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"os/exec"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var myFontFace font.Face

func addLabel(img *image.RGBA, x, y int, label string, col color.RGBA) {
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: myFontFace,
		Dot:  point,
	}
	d.DrawString(label)
}

var term_cells = [2]int32{50, 24}
var char_dims = [2]int32{7, 13}
var term_borders_dims = [2]int32{12, 8}
var screen_border_dims = [2]int32{32, 6}
var terminal_dims = [2]int32{term_cells[0] * char_dims[0], term_cells[1] * char_dims[1]}

// var win_dims = [2]int32{(terminal_dims[0] + term_borders_dims[0]) * 6 / 2, (terminal_dims[1] + term_borders_dims[1]) * 6 / 2}
var win_dims = [2]int32{1920, 1080}

func drawStringToImage(strs []string, operatingImage *image.RGBA, col color.RGBA) {
	for i, str := range strs {
		addLabel(operatingImage, int(term_borders_dims[0]), int(term_borders_dims[1])+13+i*13, str, col)
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

func overwriteTexWithImage(img *image.RGBA, textureHandle uint32) {
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

//go:embed Shaders/full_screen_quad.frag
var fragSrc string

//go:embed Shaders/full_screen_quad.vert
var vertSrc string

//go:embed Shaders/blur_shader.comp
var compSrc string

func doCompute(program uint32, from, to uint32, size [2]int32) {
	sz := [2]uint32{uint32(size[0]), uint32(size[1])}
	var groupSize uint32 = 16
	gl.UseProgram(program)
	gl.BindImageTexture(0, from, 0, false, 0, gl.READ_ONLY, gl.RGBA8)
	gl.BindImageTexture(1, to, 0, false, 0, gl.WRITE_ONLY, gl.RGBA8)
	gl.Uniform1f(gl.GetUniformLocation(program, gl.Str("text_brightness\x00")), bloomBrightness)

	gl.DispatchCompute((sz[0]+groupSize)/groupSize, (sz[1]+groupSize)/groupSize, 1)

	gl.MemoryBarrier(gl.SHADER_IMAGE_ACCESS_BARRIER_BIT)

}
func doDrawing(screenProg uint32, vao uint32, scanline_pos int32, texture, bloom_handle uint32) {
	gl.BindVertexArray(vao)
	gl.UseProgram(screenProg)

	gl.Uniform2i(gl.GetUniformLocation(screenProg, gl.Str("physical_border_dims\x00")), int32(screen_border_dims[0]), int32(screen_border_dims[1]))
	gl.Uniform2i(gl.GetUniformLocation(screenProg, gl.Str("screen_dims\x00")), int32(terminal_dims[0]), int32(terminal_dims[1]))
	gl.Uniform1i(gl.GetUniformLocation(screenProg, gl.Str("ScanlinePosition\x00")), scanline_pos)

	gl.Uniform1i(gl.GetUniformLocation(screenProg, gl.Str("screenImage\x00")), 0)
	gl.Uniform1i(gl.GetUniformLocation(screenProg, gl.Str("bloomImage\x00")), 1)

	gl.Uniform1f(gl.GetUniformLocation(screenProg, gl.Str("scanlineStrength\x00")), scanline_strength)
	gl.Uniform1f(gl.GetUniformLocation(screenProg, gl.Str("bloomStrength\x00")), bloomStrength)
	gl.Uniform1f(gl.GetUniformLocation(screenProg, gl.Str("text_brightness\x00")), text_brightness)
	gl.Uniform1f(gl.GetUniformLocation(screenProg, gl.Str("ambient\x00")), ambient)
	gl.Uniform1f(gl.GetUniformLocation(screenProg, gl.Str("noiseStrength\x00")), noiseStrength)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, bloom_handle)

	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}

func makeGLStuff() (uint32, uint32, uint32, *image.RGBA, uint32, uint32, uint32) {
	_, vao := screenVBOVAO()
	screenProg, err := BuildProgram(fragSrc, vertSrc)
	check(err)
	blur_program, err := BuildCompute(compSrc)
	check(err)

	operating_img := image.NewRGBA(image.Rect(0, 0, int(terminal_dims[0])+2*int(term_borders_dims[0]), int(terminal_dims[1])+2*int(term_borders_dims[1])))
	clearImage(operating_img, color.RGBA{0, 0, 0, 255})

	textHandle := texFromImage(operating_img)
	pingHandle := texFromImage(operating_img)
	pongHandle := texFromImage(operating_img)

	return vao, screenProg, blur_program, operating_img, textHandle, pingHandle, pongHandle
}

func readFromCommand(cmd *exec.Cmd, datachan chan string) {
	stdout, err := cmd.StdoutPipe()
	check(err)
	for {
		bs := make([]byte, 200)
		n, err := stdout.Read(bs)
		check(err)
		if n > 0 {
			fmt.Println("read some", string(bs))
			check(err)
			datachan <- string(bs)
		}
	}
}

// xos4 termius is a good font
func main() {
	myFontFace = basicfont.Face7x13

	window := initWindow()
	defer cleanupWindow(window)

	vao, screenProg, blur_program, operating_img, textHandle, pingHandle, pongHandle := makeGLStuff()
	var lines []string

	cmd := exec.Command("bash")

	cmd_output_chan := make(chan string)
	cmd_output_history := ""
	go readFromCommand(cmd, cmd_output_chan)

	writer, err := cmd.StdinPipe()
	check(err)
	writer.Write([]byte("ls\n"))
	cmd.Start()

	writer.Write([]byte("pwd\n"))

	var frame_num = 0
	var scanline_pos int32 = 0
	for !window.ShouldClose() {
		frame_num++
		scanline_pos += 501
		scanline_pos %= int32(terminal_dims[0] * terminal_dims[1])

		clearImage(operating_img, color.RGBA{0, 0, 0, 255})
		if showui {
			lines = MakeUI()
		} else {
			select {
			case str, ok := <-cmd_output_chan:
				if ok {
					if str != "" {
						fmt.Printf("Value %s was read.\n", str)
						cmd_output_history += str
					}
				} else {
					fmt.Println("Channel closed!")
					break
				}
			default:
			}
			lines = strings.Split(cmd_output_history, "\n")

		}

		//draw text into image, texture
		drawStringToImage(lines, operating_img, color.RGBA{255, 255, 255, 255})
		overwriteTexWithImage(operating_img, textHandle)

		// Do blurring
		doCompute(blur_program, textHandle, pingHandle, terminal_dims)
		doCompute(blur_program, pingHandle, pongHandle, terminal_dims)

		prerender()

		doDrawing(screenProg, vao, scanline_pos, textHandle, pongHandle)

		//post render
		glfw.PollEvents()
		window.SwapBuffers()
	}
}
func prerender() {
	gl.ClearColor(clear_col[0], clear_col[1], clear_col[2], clear_col[3])
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Viewport(0, 0, win_dims[0], win_dims[1])
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
