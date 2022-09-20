package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"unsafe"

	"github.com/fatih/color"
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func BuildCompute(comp_shader_source string) (uint32, error) {
	var compute uint32
	var err error
	// compute shader
	compute, err = compileShader(comp_shader_source, gl.COMPUTE_SHADER)
	if err != nil {
		return 0, err
	}
	// shader Program
	program := gl.CreateProgram()
	gl.AttachShader(program, compute)
	gl.LinkProgram(program)

	//Check Link Errors
	var isLinked int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &isLinked)
	if isLinked == gl.FALSE {
		var maxLength int32
		gl.GetProgramiv(compute, gl.INFO_LOG_LENGTH, &maxLength)

		infoLog := make([]uint8, maxLength+1) //[bufSize]uint8{}
		gl.GetShaderInfoLog(compute, maxLength, &maxLength, &infoLog[0])

		return 0, fmt.Errorf(";ink Result %s", string(infoLog))

	}
	return program, nil

}

// Builds the opengl program
func BuildProgram(FragSrc, VertSrc string) (uint32, error) {

	Program := gl.CreateProgram()

	var vertexShader, fragmentShader uint32
	var err error

	//Compile Vertex Shader
	vertexShader, err = compileShader(VertSrc+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		err = fmt.Errorf("vertex shader error: %s", err)
		return 0, err
	}
	gl.AttachShader(Program, vertexShader)

	//Compile Fragment Shader
	fragmentShader, err = compileShader(FragSrc+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		err = fmt.Errorf("fragment shader error: %s", err)
		return 0, err
	}
	gl.AttachShader(Program, fragmentShader)

	gl.LinkProgram(Program)

	//Release programs
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	//Check Link Errors
	var isLinked int32
	gl.GetProgramiv(Program, gl.LINK_STATUS, &isLinked)
	if isLinked == gl.FALSE {
		var maxLength int32
		gl.GetProgramiv(fragmentShader, gl.INFO_LOG_LENGTH, &maxLength)

		infoLog := make([]uint8, maxLength+1)
		gl.GetShaderInfoLog(fragmentShader, maxLength, &maxLength, &infoLog[0])

		return 0, fmt.Errorf("error linking: %s", string(infoLog))

	}
	return Program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		shaderString := "Unknown"
		if shaderType == gl.FRAGMENT_SHADER {
			shaderString = "Fragment"
		} else if shaderType == gl.VERTEX_SHADER {
			shaderString = "Vertex"
		}
		return 0, fmt.Errorf("failed to compile type %s:\nLog:\n%v", shaderString, log[:len(log)-2])
	}

	return shader, nil
}

func screenVBOVAO() (uint32, uint32) {
	var points []float32 = []float32{
		-1, 1, 0,
		-1, -1, 0,
		1, -1, 0,

		-1, 1, 0,
		1, -1, 0,
		1, 1, 0,
	}

	var vbo, vao uint32
	//Build full screen quad    ==============================================
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vbo, vao

}

func cleanupWindow(window *glfw.Window) {
	window.Destroy()
	glfw.Terminate()
}
func initWindow() *glfw.Window {
	runtime.LockOSThread()

	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	mon := glfw.GetPrimaryMonitor()

	window, err := glfw.CreateWindow(int(win_dims[0]), int(win_dims[1]), "Testing", mon, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	// Important! Call gl.Init only under the presence of an active OpenGL context,
	// i.e., after MakeContextCurrent.
	if err := gl.Init(); err != nil {
		log.Fatalln(err)
	}
	glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)

	//Pipe debug to callback
	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(DBHAndle, nil)

	return window
}
func DBHAndle(source, gltype, id, severity uint32, length int32, message string, userParam unsafe.Pointer) {
	if id == 131169 || id == 131185 || id == 131218 || id == 131204 {
		return
	}
	sources := map[uint32]string{
		gl.DEBUG_SOURCE_API:             "Source: API",
		gl.DEBUG_SOURCE_WINDOW_SYSTEM:   "Source: Window System",
		gl.DEBUG_SOURCE_SHADER_COMPILER: "Source: Shader Compiler",
		gl.DEBUG_SOURCE_THIRD_PARTY:     "Source: Third Party",
		gl.DEBUG_SOURCE_APPLICATION:     "Source: Application",
		gl.DEBUG_SOURCE_OTHER:           "Source: Other",
	}
	severities := map[uint32]string{
		gl.DEBUG_SEVERITY_HIGH:         "high",
		gl.DEBUG_SEVERITY_MEDIUM:       "medium",
		gl.DEBUG_SEVERITY_LOW:          "low",
		gl.DEBUG_SEVERITY_NOTIFICATION: "notification",
	}
	if severity == gl.DEBUG_SEVERITY_NOTIFICATION {
		return
	}
	g := color.New(color.FgBlack, color.BgGreen)
	g.Print("GL Error: ")
	log.Printf("%s type = 0x%x, severity: %s, message = %s", sources[source], gltype, severities[severity], message)
}
