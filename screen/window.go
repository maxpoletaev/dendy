package screen

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	Width     = 256
	Height    = 240
	FrameRate = 60
	Title     = "Dendy Emulator"
)

type Window struct {
	ZapperDelegate func(brightness uint8, trigger bool)
	InputDelegate  func(buttons uint8)
	ResetDelegate  func()
	ShowFPS        bool
	FPS            int

	frame    *[256][240]color.RGBA
	texture  rl.RenderTexture2D
	pixels   []color.RGBA
	slowMode bool
	scale    int

	sourceRec rl.Rectangle
	destRec   rl.Rectangle
}

func Show(frame *[256][240]color.RGBA, scale int) *Window {
	rl.SetTraceLog(rl.LogNone)
	rl.SetTargetFPS(FrameRate)
	rl.InitWindow(Width*int32(scale), Height*int32(scale), Title)

	texture := rl.LoadRenderTexture(Width, Height)
	rl.SetTextureFilter(texture.Texture, rl.FilterPoint)

	sourceRec := rl.NewRectangle(0, 0, Width, Height)
	destRec := rl.NewRectangle(0, 0, float32(Width*scale), float32(Height*scale))

	return &Window{
		pixels:    make([]color.RGBA, Width*Height),
		texture:   texture,
		frame:     frame,
		scale:     scale,
		sourceRec: sourceRec,
		destRec:   destRec,
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
		rl.SetTargetFPS(FrameRate)
	}
}

func (w *Window) Close() {
	rl.CloseWindow()
}

func (w *Window) ShouldClose() bool {
	return rl.WindowShouldClose()
}

func (w *Window) updateTexture() {
	for x := 0; x < Width; x++ {
		for y := 0; y < Height; y++ {
			w.pixels[x+y*Width] = w.frame[x][y]
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

func (w *Window) isCtrlPressed() bool {
	ctrl := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	super := rl.IsKeyDown(rl.KeyLeftSuper) || rl.IsKeyDown(rl.KeyRightSuper)
	return super || ctrl
}

func (w *Window) HandleHotKeys() {
	switch {
	case rl.IsKeyPressed(rl.KeyF1):
		w.ToggleSlowMode()

	case rl.IsKeyPressed(rl.KeyF12):
		rl.TakeScreenshot("screenshot.png")

	case w.isCtrlPressed() && rl.IsKeyPressed(rl.KeyR):
		if w.ResetDelegate != nil {
			w.ResetDelegate()
		}
	}
}
