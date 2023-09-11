package apu

func triangleWave(phase float32) float32 {
	abs := func(x float32) float32 {
		if x < 0 {
			return -x
		}
		return x
	}

	phase = phase - float32(int(phase))
	return 4.0 * (abs(phase-0.5) - 0.25)
}

type triangle struct {
	enabled  bool
	sample   float32
	sequence uint8

	// Timer
	timerLoad  uint16
	timerValue uint16

	// Length counter
	lengthValue uint8
	lengthHalt  bool

	// Linear counter
	linearEnabled bool
	linearLoad    uint8
	linearValue   uint8
	linearReload  bool
}

func (tr *triangle) reset() {
	tr.enabled = false
	tr.sample = 0
	tr.sequence = 0

	tr.timerLoad = 0
	tr.timerValue = 0

	tr.lengthValue = 0
	tr.lengthHalt = false

	tr.linearLoad = 0
	tr.linearValue = 0
	tr.linearReload = false
}

func (tr *triangle) write(addr uint16, value byte) {
	switch addr {
	case 0x4008:
		tr.linearLoad = value & 0x7F
		tr.lengthHalt = value&0x80 == 0
		tr.linearEnabled = value&0x80 == 0
	case 0x400A:
		tr.timerLoad = tr.timerLoad&0xFF00 | uint16(value)
		tr.timerValue = tr.timerLoad
	case 0x400B:
		tr.timerLoad = tr.timerLoad&0x00FF | uint16(value&0x07)<<8
		tr.lengthValue = lengthTable[value>>3]
		tr.timerValue = tr.timerLoad
		tr.linearReload = true
	}
}

func (tr *triangle) tickLength() {
	if !tr.lengthHalt && tr.lengthValue > 0 {
		tr.lengthValue--
	}
}

func (tr *triangle) tickLinear() {
	if tr.linearReload {
		tr.linearValue = tr.linearLoad
	} else if tr.linearValue > 0 {
		tr.linearValue--
	}

	if tr.linearEnabled {
		tr.linearReload = false
	}
}

func (tr *triangle) tickTimer(t float32) {
	if tr.lengthValue == 0 || tr.linearValue == 0 || tr.timerValue < 3 {
		return
	}

	freq := 1789773.0 / (32.0 * (float32(tr.timerValue) + 1.0))
	tr.sample = triangleWave(t * freq)
}

func (tr *triangle) output() float32 {
	if !tr.enabled {
		return 0
	}

	return tr.sample
}
