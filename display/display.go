package display

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 240
)

type Display struct {
	frame   *[256][240]color.RGBA
	texture rl.RenderTexture2D
	pixels  []color.RGBA
	scale   int

	sourceRec rl.Rectangle
	destRec   rl.Rectangle
}

func New(frame *[256][240]color.RGBA, scale int) *Display {
	rl.SetTargetFPS(60)
	rl.SetTraceLog(rl.LogError)
	rl.InitWindow(ScreenWidth*int32(scale), ScreenHeight*int32(scale), "Dendy Emulator")

	texture := rl.LoadRenderTexture(ScreenWidth, ScreenHeight)
	rl.SetTextureFilter(texture.Texture, rl.FilterPoint)

	sourceRec := rl.NewRectangle(0, 0, ScreenWidth, ScreenHeight)
	destRec := rl.NewRectangle(0, 0, float32(ScreenWidth*scale), float32(ScreenHeight*scale))

	return &Display{
		pixels:    make([]color.RGBA, ScreenWidth*ScreenHeight),
		texture:   texture,
		frame:     frame,
		scale:     scale,
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

func (s *Display) drawDirect() {
	rl.BeginDrawing()
	defer rl.EndDrawing()

	rl.ClearBackground(rl.Black)
	for x := 0; x < ScreenWidth; x++ {
		for y := 0; y < ScreenHeight; y++ {
			rl.DrawPixel(int32(x), int32(y), rl.Color{
				R: s.frame[x][y].R,
				G: s.frame[x][y].G,
				B: s.frame[x][y].B,
				A: s.frame[x][y].A,
			})
		}
	}

	fps := fmt.Sprintf("%d fps", rl.GetFPS())
	rl.DrawText(fps, 6, 6, 10, rl.Black)
	rl.DrawText(fps, 5, 5, 10, rl.White)
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
