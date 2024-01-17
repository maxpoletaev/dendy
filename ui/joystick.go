package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/maxpoletaev/dendy/input"
)

var keyMap = map[int32]input.Button{
	rl.KeyW:          input.ButtonUp,
	rl.KeyS:          input.ButtonDown,
	rl.KeyA:          input.ButtonLeft,
	rl.KeyD:          input.ButtonRight,
	rl.KeyK:          input.ButtonA,
	rl.KeyJ:          input.ButtonB,
	rl.KeyEnter:      input.ButtonStart,
	rl.KeyRightShift: input.ButtonSelect,
}

func (w *Window) UpdateJoystick() {
	if w.InputDelegate == nil {
		return
	}

	var buttons uint8

	for key, button := range keyMap {
		if rl.IsKeyDown(key) {
			buttons |= button
		}
	}

	w.InputDelegate(buttons)
}
