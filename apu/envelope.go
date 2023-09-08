package apu

type envelope struct {
	enabled bool
	start   bool
	loop    bool
	value   uint8
	load    uint8
	volume  uint8
}

func (e *envelope) reset() {
	e.enabled = false
	e.start = false
	e.loop = false
	e.value = 0
	e.load = 0
	e.volume = 0
}

func (e *envelope) tick() {
	if e.start {
		e.value = e.load
		e.start = false
		e.value = 15
	} else if e.value > 0 {
		e.value--
	} else {
		if e.volume > 0 {
			e.value = e.load
			e.volume--
		} else if e.loop {
			e.value = e.load
			e.volume = 15
		}
	}
}
