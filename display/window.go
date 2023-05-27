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
	KeyEnter = rl.KeyEnter
)

type Window struct {
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

func Show(frame *[256][240]color.RGBA, joy1 *input.Joystick, zap *input.Zapper, scale int) *Window {
	rl.SetTraceLog(rl.LogNone)
	rl.SetTargetFPS(60) // PAL

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

	return &Window{
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

func (w *Window) Close() {
	rl.CloseWindow()
}

func (w *Window) ShouldClose() bool {
	return rl.WindowShouldClose()
}

func (w *Window) updateTexture() {
	for x := 0; x < ScreenWidth; x++ {
		for y := 0; y < ScreenHeight; y++ {
			w.pixels[x+y*ScreenWidth] = w.frame[x][y]
		}
	}

	rl.UpdateTexture(w.texture.Texture, w.pixels)
}

func (w *Window) handleJoystick() {
	for key, button := range w.keyMap {
		if rl.IsKeyDown(key) {
			w.joy1.Press(button)
		} else if rl.IsKeyUp(key) {
			w.joy1.Release(button)
		}
	}
}

func (w *Window) handleZapper() {
	pos := rl.GetMousePosition()
	x, y := int(pos.X)/w.scale, int(pos.Y)/w.scale

	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return
	}

	// Check if the gun is over the light spot.
	w.zap.DetectLight(w.frame[x][y])

	// Check if the trigger is pressed. We need to reset the light sensor
	// state for the frame when the trigger was pulled, so it doesn't
	// catch any light from the game (say, light background).
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		w.zap.ResetSensor()
		w.zap.PullTrigger()
	} else {
		w.zap.ReleaseTrigger()
	}
}

func (w *Window) HandleInput() {
	w.handleJoystick()
	w.handleZapper()

	if rl.IsKeyPressed(rl.KeyF12) {
		rl.TakeScreenshot("screenshot.png")
	}
}

func (w *Window) Refresh() {
	w.updateTexture()

	rl.BeginDrawing()
	defer rl.EndDrawing()

	origin := rl.NewVector2(0, 0)
	rl.ClearBackground(rl.Black)
	rl.DrawTexturePro(w.texture.Texture, w.sourceRec, w.destRec, origin, 0, rl.White)

	if w.ShowFPS {
		fps := fmt.Sprintf("%d fps", rl.GetFPS())
		rl.DrawText(fps, 6, 6, 10, rl.Black)
		rl.DrawText(fps, 5, 5, 10, rl.White)
	}
}

func (w *Window) IsResetPressed() bool {
	super := rl.IsKeyDown(rl.KeyLeftSuper) || rl.IsKeyDown(rl.KeyRightSuper)
	ctrl := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	return (super || ctrl) && rl.IsKeyPressed(rl.KeyR)
}

func (w *Window) KeyPressed(key int32) bool {
	return rl.IsKeyPressed(key)
}

func (w *Window) InFocus() bool {
	return rl.IsWindowFocused()
}
