package ui

import (
	"image/color"

	"github.com/gen2brain/raylib-go/raylib"

	"github.com/maxpoletaev/dendy/ppu"
)

func (w *Window) getFrameMousePosition() (int, int, bool) {
	pos := rl.GetMousePosition()

	x := int(pos.X) / w.scale
	if x < 0 || x >= ppu.FrameWidth {
		return 0, 0, false
	}

	y := int(pos.Y) / w.scale
	if y < 0 || y >= ppu.FrameHeight {
		return 0, 0, false
	}

	return x, y, true
}

func (w *Window) isTriggerPressed() bool {
	return rl.IsMouseButtonDown(rl.MouseLeftButton) ||
		rl.IsMouseButtonPressed(rl.MouseLeftButton)
}

func (w *Window) UpdateZapper(ppuFrame []color.RGBA) {
	if w.ZapperDelegate == nil {
		return
	}

	var (
		x, y int
		ok   bool
	)

	if x, y, ok = w.getFrameMousePosition(); !ok {
		w.ZapperDelegate(0, w.isTriggerPressed())
		return
	}

	rgb := ppuFrame[y*ppu.FrameWidth+x]

	brightness := (rgb.R + rgb.G + rgb.B) / 3

	w.ZapperDelegate(brightness, w.isTriggerPressed())
}
