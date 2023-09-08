package ui

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	width  = 256
	height = 240
)

type Window struct {
	ZapperDelegate func(brightness uint8, trigger bool)
	InputDelegate  func(buttons uint8)
	ResetDelegate  func()
	ShowPing       bool
	ShowFPS        bool
	FPS            int

	latency int64
	frame   *[width][height]color.RGBA
	texture rl.RenderTexture2D
	pixels  []color.RGBA
	scale   int

	sourceRec rl.Rectangle
	targetRec rl.Rectangle
}

func CreateWindow(frame *[width][height]color.RGBA, scale int, verbose bool) *Window {
	if !verbose {
		rl.SetTraceLog(rl.LogNone)
	}

	rl.InitWindow(width*int32(scale), height*int32(scale), "")
	texture := rl.LoadRenderTexture(width, height)
	rl.SetTextureFilter(texture.Texture, rl.FilterPoint)

	sourceRec := rl.NewRectangle(0, 0, width, height)
	targetRec := rl.NewRectangle(0, 0, float32(width*scale), float32(height*scale))

	return &Window{
		pixels:    make([]color.RGBA, width*height),
		texture:   texture,
		frame:     frame,
		scale:     scale,
		sourceRec: sourceRec,
		targetRec: targetRec,
	}
}

func (w *Window) SetTitle(title string) {
	rl.SetWindowTitle(title)
}

func (w *Window) SetFrameRate(fps int) {
	rl.SetTargetFPS(int32(fps))
}

func (w *Window) Close() {
	rl.CloseWindow()
}

func (w *Window) ShouldClose() bool {
	return rl.WindowShouldClose()
}

func (w *Window) updateTexture() {
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			w.pixels[x+y*width] = w.frame[x][y]
		}
	}

	rl.UpdateTexture(w.texture.Texture, w.pixels)
}

func (w *Window) SetLatencyInfo(latency int64) {
	w.latency = latency
}

func (w *Window) drawTextWithShadow(text string, x int32, y int32, size int32, colour rl.Color) {
	rl.DrawText(text, x+1, y+1, size, rl.Black)
	rl.DrawText(text, x, y, size, colour)
}

func (w *Window) Refresh() {
	w.updateTexture()
	rl.BeginDrawing()

	origin := rl.NewVector2(0, 0)
	rl.ClearBackground(rl.Black)
	rl.DrawTexturePro(w.texture.Texture, w.sourceRec, w.targetRec, origin, 0, rl.White)

	var offsetY int32

	if w.ShowFPS {
		textY := offsetY + 5
		fps := fmt.Sprintf("%d fps", rl.GetFPS())
		w.drawTextWithShadow(fps, 6, textY, 10, rl.White)
		offsetY += 10
	}

	if w.ShowPing && w.latency > 0 {
		textY := offsetY + 5
		colour := rl.Green

		if w.latency > 150 {
			colour = rl.Red
		} else if w.latency > 100 {
			colour = rl.Yellow
		}

		latency := fmt.Sprintf("%d ms", w.latency)
		w.drawTextWithShadow(latency, 6, textY, 10, colour)
	}

	rl.EndDrawing()
}

func (w *Window) InFocus() bool {
	return rl.IsWindowFocused()
}

func (w *Window) isModifierPressed() bool {
	ctrl := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	super := rl.IsKeyDown(rl.KeyLeftSuper) || rl.IsKeyDown(rl.KeyRightSuper)
	return super || ctrl
}

func (w *Window) HandleHotKeys() {
	switch {
	case rl.IsKeyPressed(rl.KeyF12):
		rl.TakeScreenshot("screenshot.png")

	case w.isModifierPressed() && rl.IsKeyPressed(rl.KeyR):
		if w.ResetDelegate != nil {
			w.ResetDelegate()
		}
	}
}
