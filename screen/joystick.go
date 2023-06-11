package screen

import rl "github.com/gen2brain/raylib-go/raylib"

func (w *Window) UpdateJoystick() {
	if w.InputDelegate == nil {
		return
	}

	var buttons uint8
	for key, button := range w.keyMap {
		if rl.IsKeyDown(key) {
			buttons |= 1 << uint8(button)
		}
	}

	w.InputDelegate(buttons)
}
