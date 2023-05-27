package input

import "image/color"

type Zapper struct {
	lightDetected  bool
	triggerPressed bool
}

func NewZapper() *Zapper {
	return &Zapper{
		lightDetected: false,
	}
}

func (z *Zapper) PullTrigger() {
	z.triggerPressed = true
}

func (z *Zapper) ReleaseTrigger() {
	z.triggerPressed = false
}

func (z *Zapper) DetectLight(rgb color.RGBA) bool {
	hit := rgb.R > 250 && rgb.G > 250 && rgb.B > 250
	z.lightDetected = hit
	return hit
}

func (z *Zapper) ResetSensor() {
	z.lightDetected = false
}

func (z *Zapper) Read() (value byte) {
	if !z.lightDetected {
		value |= 1 << 3
	}

	if z.triggerPressed {
		value |= 1 << 4
	}

	return value
}
