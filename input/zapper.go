package input

type Zapper struct {
	lightDetected  bool
	triggerPressed bool
}

func NewZapper() *Zapper {
	return &Zapper{
		lightDetected: false,
	}
}

func (z *Zapper) PressTrigger() {
	z.triggerPressed = true
}

func (z *Zapper) TriggerPressed() bool {
	return z.triggerPressed
}

func (z *Zapper) ReleaseTrigger() {
	z.triggerPressed = false
}

func (z *Zapper) LightDetected(lightDetected bool) {
	z.lightDetected = lightDetected
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
