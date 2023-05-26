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

const (
	KeySpace = rl.KeySpace
	KeyF     = rl.KeyF
)

type Display struct {
	ShowFPS bool

	keyMap  map[int32]input.Button
	frame   *[256][240]color.RGBA
	texture rl.RenderTexture2D
	joy1    *input.Joystick
	zap     *input.Zapper
	pixels  []color.RGBA
	scale   int

	sourceRec rl.Rectangle
	destRec   rl.Rectangle
}

func Show(frame *[256][240]color.RGBA, joy1 *input.Joystick, zap *input.Zapper, scale int) *Display {
	rl.SetTraceLog(rl.LogNone)
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
		sourceRec: sourceRec,
		keyMap:    keyMap,
		destRec:   destRec,
		joy1:      joy1,
		zap:       zap,
	}
}

func (d *Display) Close() {
	rl.CloseWindow()
}

func (d *Display) ShouldClose() bool {
	return rl.WindowShouldClose()
}

func (d *Display) updateTexture() {
	for x := 0; x < ScreenWidth; x++ {
		for y := 0; y < ScreenHeight; y++ {
			d.pixels[x+y*ScreenWidth] = d.frame[x][y]
		}
	}

	rl.UpdateTexture(d.texture.Texture, d.pixels)
}

func (d *Display) handleJoystick() {
	for key, button := range d.keyMap {
		if rl.IsKeyDown(key) {
			d.joy1.Press(button)
		} else if rl.IsKeyUp(key) {
			d.joy1.Release(button)
		}
	}
}

func (d *Display) handleZapper() {
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		d.zap.PressTrigger()
		return
	}

	d.zap.ReleaseTrigger()
	pos := rl.GetMousePosition()

	x, y := int(pos.X)/d.scale, int(pos.Y)/d.scale
	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return
	}

	rgb := d.frame[x][y]
	hit := rgb.R > 100 && rgb.G > 100 && rgb.B > 100

	d.zap.LightDetected(hit)
}

func (d *Display) HandleInput() {
	d.handleJoystick()
	d.handleZapper()
}

func (d *Display) NoSignal() {
	rl.BeginDrawing()
	defer rl.EndDrawing()

	rl.ClearBackground(rl.Blue)
	rl.DrawText("NO SIGNAL", 20, 20, 30, rl.Gold)
}

func (d *Display) Noop() {
	rl.BeginDrawing()
	rl.EndDrawing()
}

func (d *Display) Refresh() {
	d.updateTexture()

	rl.BeginDrawing()
	defer rl.EndDrawing()

	origin := rl.NewVector2(0, 0)
	rl.ClearBackground(rl.Black)
	rl.DrawTexturePro(d.texture.Texture, d.sourceRec, d.destRec, origin, 0, rl.White)

	if d.ShowFPS {
		fps := fmt.Sprintf("%d fps", rl.GetFPS())
		rl.DrawText(fps, 6, 6, 10, rl.Black)
		rl.DrawText(fps, 5, 5, 10, rl.White)
	}
}

func (d *Display) KeyPressed(key int32) bool {
	return rl.IsKeyPressed(key)
}
