package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

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

var dims = [2]int{512, 512}
var char_dims = [2]int{7, 13}

func drawStringToNewTex(str string, operatingImage *image.RGBA, col color.RGBA) {
	addLabel(operatingImage, 0, 13, str, col)
}

func main() {
	img := image.NewRGBA(image.Rect(0, 0, dims[0], dims[1]))

	for i := 0; i < 60; i++ {
		col := color.RGBA{0, 0, 0, 255}

		drawStringToNewTex(fmt.Sprint("frame:", i-1,"\nwow"), img, col)
		col = color.RGBA{200, 100, 0, 255}
		drawStringToNewTex(fmt.Sprint("frame:", i,"\nwow"), img, col)
	}
	f, err := os.Create("out.png")
	check(err)
	defer f.Close()
	err = png.Encode(f, img)
	check(err)
}



func check(err error) {
	if err != nil {
		panic(err)
	}
}
