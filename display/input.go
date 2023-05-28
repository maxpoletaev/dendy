package display

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/maxpoletaev/dendy/input"
)

func (w *Window) getFrameMousePosition() (int, int, bool) {
	pos := rl.GetMousePosition()
	x, y := int(pos.X)/w.scale, int(pos.Y)/w.scale

	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return 0, 0, false
	}

	return x, y, true
}

func (w *Window) UpdateZapper(zap *input.Zapper) {
	var (
		x, y int
		ok   bool
	)

	if x, y, ok = w.getFrameMousePosition(); !ok {
		zap.SetBrightness(0)
		return
	}

	rgb := w.frame[x][y]
	zap.SetBrightness((rgb.R + rgb.G + rgb.B) / 3)

	if rl.IsMouseButtonDown(rl.MouseLeftButton) ||
		rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		zap.PullTrigger()
	} else {
		zap.ReleaseTrigger()
	}
}

func (w *Window) UpdateJoystick(joy *input.Joystick) {
	for key, button := range w.keyMap {
		if rl.IsKeyDown(key) {
			joy.Press(button)
		} else {
			joy.Release(button)
		}
	}
}

func (w *Window) HandleHotKeys() {
	if rl.IsKeyPressed(rl.KeyF1) {
		w.toggleSlowMode()
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

func (w *Window) KeyPressed(key int32) bool {
	return rl.IsKeyPressed(key)
}
