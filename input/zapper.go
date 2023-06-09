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

func (z *Zapper) PullTrigger() {
	z.triggerPressed = true
}

func (z *Zapper) ReleaseTrigger() {
	z.triggerPressed = false
}

func (z *Zapper) Update(brightness uint8, trigger bool) {
	z.lightDetected = brightness > 64
	z.triggerPressed = trigger
}

func (z *Zapper) Read() (value byte) {
	if z.triggerPressed {
		value |= 1 << 4
	}

	if !z.lightDetected {
		value |= 1 << 3
	}

	return value
}
