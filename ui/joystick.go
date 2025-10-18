package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/maxpoletaev/dendy/input"
)

const (
	turboRate      = 4
	gamepadIndex   = 0
	stickDeadzone  = 0.25
	stickThreshold = 0.5
)

type keyMapping struct {
	key    int32
	button input.Button
	turbo  bool
}

var keyboardMappings = []keyMapping{
	{rl.KeyW, input.ButtonUp, false},
	{rl.KeyS, input.ButtonDown, false},
	{rl.KeyA, input.ButtonLeft, false},
	{rl.KeyD, input.ButtonRight, false},
	{rl.KeyJ, input.ButtonB, false},
	{rl.KeyK, input.ButtonA, false},
	{rl.KeyU, input.ButtonB, true},
	{rl.KeyI, input.ButtonA, true},
	{rl.KeyEnter, input.ButtonStart, false},
	{rl.KeyRightShift, input.ButtonSelect, false},
}

var gamepadMappings = []keyMapping{
	{rl.GamepadButtonLeftFaceUp, input.ButtonUp, false},
	{rl.GamepadButtonLeftFaceDown, input.ButtonDown, false},
	{rl.GamepadButtonLeftFaceLeft, input.ButtonLeft, false},
	{rl.GamepadButtonLeftFaceRight, input.ButtonRight, false},
	{rl.GamepadButtonRightFaceDown, input.ButtonA, false},
	{rl.GamepadButtonRightFaceLeft, input.ButtonB, false},
	{rl.GamepadButtonRightFaceRight, input.ButtonB, true},
	{rl.GamepadButtonRightFaceUp, input.ButtonA, true},
	{rl.GamepadButtonMiddleRight, input.ButtonStart, false},
	{rl.GamepadButtonMiddleLeft, input.ButtonSelect, false},
}

func (w *Window) UpdateJoystick() {
	if w.InputDelegate == nil {
		return
	}

	w.turboCounter++
	if w.turboCounter >= turboRate {
		w.turboCounter = 0
	}

	var buttons uint8

	if w.gamepadAvailable {
		buttons = readAnalogInput(buttons)

		for _, km := range gamepadMappings {
			if km.turbo && w.turboCounter != 0 {
				continue
			}
			if rl.IsGamepadButtonDown(gamepadIndex, km.key) {
				buttons |= km.button
			}
		}
	}

	for _, km := range keyboardMappings {
		if km.turbo && w.turboCounter != 0 {
			continue
		}
		if rl.IsKeyDown(km.key) {
			buttons |= km.button
		}
	}

	w.InputDelegate(buttons)
}

func readAnalogInput(buttons uint8) uint8 {
	leftX := rl.GetGamepadAxisMovement(gamepadIndex, rl.GamepadAxisLeftX)
	leftY := rl.GetGamepadAxisMovement(gamepadIndex, rl.GamepadAxisLeftY)

	// Apply deadzone
	if leftX < stickDeadzone && leftX > -stickDeadzone {
		leftX = 0
	}
	if leftY < stickDeadzone && leftY > -stickDeadzone {
		leftY = 0
	}

	// Convert analog input to digital directional buttons
	if leftX < -stickThreshold {
		buttons |= input.ButtonLeft
	}
	if leftX > stickThreshold {
		buttons |= input.ButtonRight
	}
	if leftY < -stickThreshold {
		buttons |= input.ButtonUp
	}
	if leftY > stickThreshold {
		buttons |= input.ButtonDown
	}

	return buttons
}
