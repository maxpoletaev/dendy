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

func toGrayscale(c color.RGBA) color.RGBA {
	gray := uint8(float64(c.R)*0.3 + float64(c.G)*0.59 + float64(c.B)*0.11)
	return color.RGBA{R: gray, G: gray, B: gray, A: c.A}
}

type Window struct {
	ZapperDelegate func(brightness uint8, trigger bool)
	InputDelegate  func(buttons uint8)
	ResyncDelegate func()
	ResetDelegate  func()
	ShowPing       bool
	ShowFPS        bool
	FPS            int

	remotePing  int64
	shouldClose bool
	frame       *[width][height]color.RGBA
	texture     rl.RenderTexture2D
	pixels      []color.RGBA
	grayscale   bool
	scale       int

	sourceRec rl.Rectangle
	targetRec rl.Rectangle
}

func CreateWindow(frame *[width][height]color.RGBA, scale int, verbose bool) *Window {
	if !verbose {
		rl.SetTraceLogLevel(rl.LogNone)
	}

	rl.InitWindow(width*int32(scale), height*int32(scale), "")
	rl.SetExitKey(0) // disable exit on ESC

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

func (w *Window) SetGrayscale(grayscale bool) {
	w.grayscale = grayscale
}

func (w *Window) Close() {
	rl.CloseWindow()
}

func (w *Window) ShouldClose() bool {
	return w.shouldClose || rl.WindowShouldClose()
}

func (w *Window) updateTexture() {
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			px := w.frame[x][y]
			if w.grayscale {
				px = toGrayscale(px)
			}

			w.pixels[x+y*width] = px
		}
	}

	rl.UpdateTexture(w.texture.Texture, w.pixels)
}

func (w *Window) SetPingInfo(pingMs int64) {
	w.remotePing = pingMs
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

	if w.ShowPing && w.remotePing > 0 {
		textY := offsetY + 5
		colour := rl.Green

		if w.remotePing > 150 {
			colour = rl.Red
		} else if w.remotePing > 100 {
			colour = rl.Yellow
		}

		ping := fmt.Sprintf("%d ms", w.remotePing)
		w.drawTextWithShadow(ping, 6, textY, 10, colour)
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

	case w.isModifierPressed() && rl.IsKeyPressed(rl.KeyQ):
		w.shouldClose = true

	case w.isModifierPressed() && rl.IsKeyPressed(rl.KeyR):
		if w.ResetDelegate != nil {
			w.ResetDelegate()
		}

	case w.isModifierPressed() && rl.IsKeyPressed(rl.KeyX):
		if w.ResyncDelegate != nil {
			w.ResyncDelegate()
		}
	}
}
