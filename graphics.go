package main

import (
	_ "embed"
	"image"
	"image/color"

	"github.com/go-gl/gl/v4.6-core/gl"
)

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
