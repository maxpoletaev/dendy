package ui

import rl "github.com/gen2brain/raylib-go/raylib"

type shaderFacade struct {
	shader   rl.Shader
	timeLoc  int32
	scaleLoc int32
}

func newShader(code string) *shaderFacade {
	shader := rl.LoadShaderFromMemory("", code)

	return &shaderFacade{
		shader:   shader,
		timeLoc:  rl.GetShaderLocation(shader, "time"),
		scaleLoc: rl.GetShaderLocation(shader, "scale"),
	}
}

func (s *shaderFacade) begin() {
	rl.BeginShaderMode(s.shader)
}

func (s *shaderFacade) end() {
	rl.EndShaderMode()
}

func (s *shaderFacade) unload() {
	rl.UnloadShader(s.shader)
}

func (s *shaderFacade) setTimeUniform(time float32) {
	rl.SetShaderValue(s.shader, s.timeLoc, []float32{time}, rl.ShaderUniformFloat)
}

func (s *shaderFacade) setScaleUniform(scale float32) {
	rl.SetShaderValue(s.shader, s.scaleLoc, []float32{scale}, rl.ShaderUniformFloat)
}
