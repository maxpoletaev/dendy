package screen

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

type Window struct {
	ZapperDelegate func(brightness uint8, trigger bool)
	InputDelegate  func(buttons uint8)
	slowMode       bool
	ShowFPS        bool

	keyMap  map[int32]input.Button
	frame   *[256][240]color.RGBA
	texture rl.RenderTexture2D
	pixels  []color.RGBA
	scale   int

	sourceRec rl.Rectangle
	destRec   rl.Rectangle
}

func Show(frame *[256][240]color.RGBA, scale int) *Window {
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

	return &Window{
		pixels:    make([]color.RGBA, ScreenWidth*ScreenHeight),
		texture:   texture,
		frame:     frame,
		scale:     scale,
		sourceRec: sourceRec,
		destRec:   destRec,
		keyMap:    keyMap,
	}
}

func (w *Window) SetTitle(title string) {
	rl.SetWindowTitle(title)
}

func (w *Window) ToggleSlowMode() {
	w.slowMode = !w.slowMode
	w.ShowFPS = true

	if w.slowMode {
		rl.SetTargetFPS(10)
	} else {
		rl.SetTargetFPS(60)
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

func (w *Window) InFocus() bool {
	return rl.IsWindowFocused()
}

func (w *Window) HandleHotKeys() {
	if rl.IsKeyPressed(rl.KeyF1) {
		w.ToggleSlowMode()
	}

	if rl.IsKeyPressed(rl.KeyF12) {
		rl.TakeScreenshot("screenshot.png")
	}
}

func (w *Window) IsResetPressed() bool {
	super := rl.IsKeyDown(rl.KeyLeftSuper) || rl.IsKeyDown(rl.KeyRightSuper)
	ctrl := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	return (super || ctrl) && rl.IsKeyPressed(rl.KeyR)
}
