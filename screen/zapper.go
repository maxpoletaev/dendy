package screen

import (
	"github.com/gen2brain/raylib-go/raylib"
)

func (w *Window) getFrameMousePosition() (int, int, bool) {
	pos := rl.GetMousePosition()

	x, y := int(pos.X)/w.scale, int(pos.Y)/w.scale

	if x < 0 || x >= Width || y < 0 || y >= Height {
		return 0, 0, false
	}

	return x, y, true
}

func (w *Window) isTriggerPressed() bool {
	return rl.IsMouseButtonDown(rl.MouseLeftButton) || rl.IsMouseButtonPressed(rl.MouseLeftButton)
}

func (w *Window) UpdateZapper() {
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

	rgb := w.frame[x][y]
	brightness := (rgb.R + rgb.G + rgb.B) / 3
	w.ZapperDelegate(brightness, w.isTriggerPressed())
}
