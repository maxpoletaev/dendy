package ui

import (
	"image/color"
	"log"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/maxpoletaev/dendy/ppu"
	"github.com/maxpoletaev/dendy/shaders"
)

func toGrayscale(c color.RGBA) color.RGBA {
	gray := uint8(float64(c.R)*0.3 + float64(c.G)*0.59 + float64(c.B)*0.11)
	return color.RGBA{R: gray, G: gray, B: gray, A: c.A}
}

type Window struct {
	ZapperDelegate func(brightness uint8, trigger bool)
	InputDelegate  func(buttons uint8)
	MuteDelegate   func()
	ResyncDelegate func()
	ResetDelegate  func()
	RewindDelegate func()
	ShowPing       bool
	ShowFPS        bool
	FPS            int

	viewport    rl.RenderTexture2D
	shader      *shaderFacade
	remotePing  int64
	shouldClose bool
	grayscale   bool
	scale       int
	width       int
	height      int
}

func CreateWindow(scale int, verbose bool) *Window {
	if !verbose {
		rl.SetTraceLogLevel(rl.LogWarning)
	}

	windowWidth := ppu.FrameWidth * scale
	windowHeight := ppu.FrameHeight * scale

	rl.InitWindow(int32(windowWidth), int32(windowHeight), "Dendy Emulator")
	rl.SetExitKey(0) // disable exit on ESC

	viewport := rl.LoadRenderTexture(ppu.FrameWidth, ppu.FrameHeight)
	rl.SetTextureFilter(viewport.Texture, rl.FilterPoint)

	return &Window{
		viewport: viewport,
		scale:    scale,
		width:    windowWidth,
		height:   windowHeight,
	}
}

func (w *Window) EnableCRT() {
	if w.scale == 1 {
		log.Printf("[WARN] CRT effect cannot be used with scale 1 (not enough pixels)")
		return
	}

	w.shader = newShader(shaders.ScanlineFragment)
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
	if w.shader != nil {
		w.shader.unload()
	}

	rl.UnloadRenderTexture(w.viewport)
	rl.CloseWindow()
}

func (w *Window) ShouldClose() bool {
	return w.shouldClose || rl.WindowShouldClose()
}

func (w *Window) SetPingInfo(pingMs int64) {
	w.remotePing = pingMs
}

func (w *Window) drawTextWithShadow(text string, x int32, y int32, size int32, colour rl.Color) {
	rl.DrawText(text, x+1, y+1, size, rl.Black)
	rl.DrawText(text, x, y, size, colour)
}

func (w *Window) updateTexture(ppuFrame []color.RGBA) {
	if w.grayscale {
		for i, c := range ppuFrame {
			ppuFrame[i] = toGrayscale(c)
		}
	}

	rl.UpdateTexture(w.viewport.Texture, ppuFrame)
}

func (w *Window) drawScreen() {
	if w.shader != nil {
		w.shader.setTimeUniform(float32(rl.GetTime()))
		w.shader.setScaleUniform(float32(w.scale))

		w.shader.begin()
		defer w.shader.end()
	}

	rl.DrawTexturePro(
		w.viewport.Texture,
		rl.Rectangle{
			Width:  float32(w.viewport.Texture.Width),
			Height: float32(w.viewport.Texture.Height),
		},
		rl.Rectangle{
			Width:  float32(w.width),
			Height: float32(w.height),
		},
		rl.Vector2{
			X: 0,
			Y: 0,
		},
		0,
		rl.White,
	)
}

func (w *Window) drawHUD() {
	var offsetY int32

	if w.ShowFPS {
		textY := offsetY + 5
		fpsText := strconv.Itoa(int(rl.GetFPS())) + " fps"
		w.drawTextWithShadow(fpsText, 6, textY, 10, rl.White)
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

		pingText := strconv.Itoa(int(w.remotePing)) + " ms"
		w.drawTextWithShadow(pingText, 6, textY, 10, colour)
	}
}

func (w *Window) Refresh(ppuFrame []color.RGBA) {
	w.updateTexture(ppuFrame)

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)

	w.drawScreen()
	w.drawHUD()

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

	case rl.IsKeyPressed(rl.KeyM):
		if w.MuteDelegate != nil {
			w.MuteDelegate()
		}

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

	case w.isModifierPressed() && rl.IsKeyPressed(rl.KeyZ):
		if w.RewindDelegate != nil {
			w.RewindDelegate()
		}
	}
}
