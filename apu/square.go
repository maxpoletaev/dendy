package apu

var squareDuty = [4]float32{
	0.125,
	0.250,
	0.500,
	0.750,
}

func squareWaveFourier(duty, frequency, amplitude, t float32) float32 {
	sin := func(t float32) float32 {
		j := t * 0.15915
		j = j - float32(int(j))
		return 20.785 * j * (j - 0.5) * (j - 1.0)
	}

	const harmonics = 20
	var p = duty * 2.0 * pi
	var a, b, n float32

	for n = 1; n <= harmonics; n++ {
		c := n * frequency * 2.0 * pi * t
		b += -sin(c-p*n) / n
		a += -sin(c) / n
	}

	return (2.0 * amplitude / pi) * (a - b)
}

type square struct {
	enabled bool
	sample  float32
	volume  uint8
	duty    uint8
	env     envelope

	// Timer
	timerValue uint16
	timerLoad  uint16

	// Length counter
	lengthValue uint8
	lengthHalt  bool
}

func (s *square) reset() {
	s.enabled = false
	s.sample = 0
	s.volume = 0
	s.duty = 0

	s.timerValue = 0
	s.timerLoad = 0

	s.lengthValue = 0
	s.lengthHalt = false

	s.env.reset()
}

func (s *square) write(addr uint16, value byte) {
	switch addr & 0x0003 {
	case 0x0000:
		s.volume = value & 0x0F
		s.duty = value >> 6 & 0x03
		s.lengthHalt = value&0x20 != 0
	case 0x0002:
		s.timerLoad = s.timerLoad&0xFF00 | uint16(value)
	case 0x0003:
		s.timerLoad = s.timerLoad&0x00FF | uint16(value&0x07)<<8
		s.lengthValue = lengthTable[value>>3]
		s.timerValue = s.timerLoad
		s.env.start = true
	case 0x0004:
		s.env.enabled = value&0x10 != 0
		s.env.loop = value&0x20 != 0
		s.env.load = value & 0x0F
	}
}

func (s *square) tickEnvelope() {
	s.env.tick()
}

func (s *square) tickLength() {
	if !s.lengthHalt && s.lengthValue > 0 {
		s.lengthValue--
	}
}

func (s *square) tickTimer(t float32) {
	freq := 1789773.0 / (16.0 * (float32(s.timerLoad) + 1.0))
	s.sample = squareWaveFourier(squareDuty[s.duty], freq, 1.0, t)
}

func (s *square) output() float32 {
	if !s.enabled || s.lengthValue == 0 || s.timerLoad < 8 {
		return 0
	}

	vol := s.volume
	if s.env.enabled {
		vol = s.env.volume
	}

	return s.sample * float32(vol) / 15.0
}
