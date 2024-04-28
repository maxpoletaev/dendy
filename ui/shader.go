package ui

import rl "github.com/gen2brain/raylib-go/raylib"

type shaderFacade struct {
	shader  rl.Shader
	timeLoc int32
}

func newShader(code string) *shaderFacade {
	shader := rl.LoadShaderFromMemory("", code)
	timeLoc := rl.GetShaderLocation(shader, "time")
	return &shaderFacade{shader, timeLoc}
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
