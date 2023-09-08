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
	enabled bool
	duty    uint8
	sample  float32
	phase   float32

	// Timer
	timerPeriod uint16
	timerValue  uint16

	// Length counter
	lengthValue uint8
	lengthHalt  bool

	// Linear counter
	linearPeriod uint8
	linearValue  uint8
	linearReload bool
}

func (tr *triangle) reset() {
	tr.enabled = false
	tr.sample = 0
	tr.duty = 0

	tr.timerPeriod = 0
	tr.timerValue = 0

	tr.lengthValue = 0
	tr.lengthHalt = false

	tr.linearPeriod = 0
	tr.linearValue = 0
	tr.linearReload = false
}

func (tr *triangle) write(addr uint16, value byte) {
	switch addr {
	case 0x4008:
		tr.lengthHalt = value&0x80 != 0
		tr.linearPeriod = value & 0x7F
	case 0x400A:
		tr.timerPeriod = tr.timerPeriod&0xFF00 | uint16(value)
	case 0x400B:
		tr.timerPeriod = tr.timerPeriod&0x00FF | uint16(value&0x07)<<8
		tr.lengthValue = lengthTable[value>>3]
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
		tr.linearValue = tr.linearPeriod
	} else if tr.linearValue > 0 {
		tr.linearValue--
	}

	if !tr.lengthHalt {
		tr.linearReload = false
	}
}

func (tr *triangle) tickTimer(t float32) {
	freq := 1789773.0 / (32.0 * (float32(tr.timerPeriod) + 1.0))
	tr.sample = triangleWave(t * freq)
}

func (tr *triangle) output() float32 {
	if !tr.enabled || tr.lengthValue == 0 || tr.linearValue == 0 || tr.timerPeriod < 3 {
		return 0
	}

	return tr.sample
}
