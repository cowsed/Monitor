package main

import (
	_ "embed"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/creack/pty"

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

var term_cells = [2]int32{120, 36}
var char_dims = [2]int32{7, 13}
var term_borders_dims = [2]int32{12, 8}
var screen_border_dims = [2]int32{8, 6}
var terminal_dims = [2]int32{term_cells[0] * char_dims[0], term_cells[1] * char_dims[1]}
var clear_col = []float32{1, 1, 1, 1}

var win_dims = [2]int32{(terminal_dims[0] + term_borders_dims[0]) * 2, (terminal_dims[1] + term_borders_dims[1]) * 2}

//var win_dims = [2]int32{1200, 600}

func drawStringToImage(strs []string, operatingImage *image.RGBA, col color.RGBA) {
	for i, str := range strs {
		addLabel(operatingImage, int(term_borders_dims[0]), int(term_borders_dims[1])+13+i*13, str, col)
	}
}

func terminal(ptmx *os.File, mw io.Writer) error {

	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH                        // Initial resize.
	defer func() { signal.Stop(ch); close(ch) }() // Cleanup signals when done.

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	// NOTE: The goroutine will keep reading until the next keystroke before returning.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(mw, ptmx)

	return nil
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func Strip(str string) string {
	return re.ReplaceAllString(str, "")
}

type mywriter struct {
	s string
}

func (mw *mywriter) Write(bs []byte) (int, error) {
	//mw.s += Strip(strings.ReplaceAll(string(bs), "\n", ""))
	noAnsi := Strip(string(bs))
	//rper := strings.NewReplacer("\n", "", "\t", "    ")
	//rper.Replace(Strip(string(bs)))
	for _, b := range []byte(noAnsi) {
		if b == 0x07 {
			mw.s = mw.s[:max(len(mw.s)-1, 0)]
		} else if b == 0x08 {

		} else {

			mw.s += string(b)
		}
	}
	//log.Println(string(bs))

	return len(bs), nil
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var term_writer *os.File

// xos4 termius is a good font
func main() {
	//Rendering
	myFontFace = basicfont.Face7x13

	window := initWindow()
	defer cleanupWindow(window)

	// Create arbitrary command.
	c := exec.Command("sh")

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	term_writer = ptmx
	check(err)

	var mw = &mywriter{}
	go func() {
		if err := terminal(ptmx, mw); err != nil {
			log.Fatal(err)
		}
		log.Println("Exited")
		window.SetShouldClose(true)
	}()

	vao, screenProg, blur_program, operating_img, textHandle, pingHandle, pongHandle := makeGLStuff()
	var lines []string

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

			lines = strings.Split(mw.s, "\r")
			lines = lines[max(0, len(lines)-int(term_cells[1])):]

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
