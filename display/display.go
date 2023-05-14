package display

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/maxpoletaev/dendy/input"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 240
	WindowTitle  = "Dendy Emulator"
)

type Display struct {
	frame   *[256][240]color.RGBA
	texture rl.RenderTexture2D
	joy1    *input.Joystick
	pixels  []color.RGBA
	scale   int

	sourceRec rl.Rectangle
	destRec   rl.Rectangle
}

func New(frame *[256][240]color.RGBA, joy1 *input.Joystick, scale int) *Display {
	rl.SetTraceLog(rl.LogError)
	rl.SetTargetFPS(60)

	rl.InitWindow(
		ScreenWidth*int32(scale),
		ScreenHeight*int32(scale),
		WindowTitle,
	)

	texture := rl.LoadRenderTexture(ScreenWidth, ScreenHeight)
	rl.SetTextureFilter(texture.Texture, rl.FilterPoint)

	sourceRec := rl.NewRectangle(0, 0, ScreenWidth, ScreenHeight)
	destRec := rl.NewRectangle(0, 0, float32(ScreenWidth*scale), float32(ScreenHeight*scale))

	return &Display{
		pixels:    make([]color.RGBA, ScreenWidth*ScreenHeight),
		texture:   texture,
		frame:     frame,
		scale:     scale,
		joy1:      joy1,
		sourceRec: sourceRec,
		destRec:   destRec,
	}
}

func (s *Display) Close() {
	rl.CloseWindow()
}

func (s *Display) IsRunning() bool {
	return !rl.WindowShouldClose()
}

func (s *Display) updateTexture() {
	for x := 0; x < ScreenWidth; x++ {
		for y := 0; y < ScreenHeight; y++ {
			s.pixels[x+y*ScreenWidth] = s.frame[x][y]
		}
	}

	rl.UpdateTexture(s.texture.Texture, s.pixels)
}

func (s *Display) HandleInput() {
	inputMap := map[int32]input.Button{
		rl.KeyW:          input.ButtonUp,
		rl.KeyS:          input.ButtonDown,
		rl.KeyA:          input.ButtonLeft,
		rl.KeyD:          input.ButtonRight,
		rl.KeyK:          input.ButtonA,
		rl.KeyJ:          input.ButtonB,
		rl.KeyEnter:      input.ButtonStart,
		rl.KeyRightShift: input.ButtonSelect,
	}

	for key, button := range inputMap {
		if rl.IsKeyDown(key) {
			s.joy1.Press(button)
		} else if rl.IsKeyUp(key) {
			s.joy1.Release(button)
		}
	}
}

func (s *Display) Refresh() {
	s.updateTexture()
	rl.BeginDrawing()
	defer rl.EndDrawing()

	origin := rl.NewVector2(0, 0)
	rl.ClearBackground(rl.Black)
	rl.DrawTexturePro(s.texture.Texture, s.sourceRec, s.destRec, origin, 0, rl.White)

	fps := fmt.Sprintf("%d fps", rl.GetFPS())
	rl.DrawText(fps, 6, 6, 10, rl.Black)
	rl.DrawText(fps, 5, 5, 10, rl.White)
}
