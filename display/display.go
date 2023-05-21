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
	keyMap  map[int32]input.Button
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

	keyMap := map[int32]input.Button{
		rl.KeyW:          input.ButtonUp,
		rl.KeyS:          input.ButtonDown,
		rl.KeyA:          input.ButtonLeft,
		rl.KeyD:          input.ButtonRight,
		rl.KeyK:          input.ButtonA,
		rl.KeyJ:          input.ButtonB,
		rl.KeyEnter:      input.ButtonStart,
		rl.KeyRightShift: input.ButtonSelect,
	}

	return &Display{
		pixels:    make([]color.RGBA, ScreenWidth*ScreenHeight),
		texture:   texture,
		frame:     frame,
		scale:     scale,
		joy1:      joy1,
		sourceRec: sourceRec,
		keyMap:    keyMap,
		destRec:   destRec,
	}
}

func (d *Display) Close() {
	rl.CloseWindow()
}

func (d *Display) IsRunning() bool {
	return !rl.WindowShouldClose()
}

func (d *Display) updateTexture() {
	for x := 0; x < ScreenWidth; x++ {
		for y := 0; y < ScreenHeight; y++ {
			d.pixels[x+y*ScreenWidth] = d.frame[x][y]
		}
	}

	rl.UpdateTexture(d.texture.Texture, d.pixels)
}

func (d *Display) HandleInput() {
	for key, button := range d.keyMap {
		if rl.IsKeyDown(key) {
			d.joy1.Press(button)
		} else if rl.IsKeyUp(key) {
			d.joy1.Release(button)
		}
	}
}

func (d *Display) Refresh() {
	d.updateTexture()
	rl.BeginDrawing()
	defer rl.EndDrawing()

	origin := rl.NewVector2(0, 0)
	rl.ClearBackground(rl.Black)
	rl.DrawTexturePro(d.texture.Texture, d.sourceRec, d.destRec, origin, 0, rl.White)

	fps := fmt.Sprintf("%d fps", rl.GetFPS())
	rl.DrawText(fps, 6, 6, 10, rl.Black)
	rl.DrawText(fps, 5, 5, 10, rl.White)
}
