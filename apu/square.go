package apu

import (
	"math"
)

var dutyTable = [4]uint8{
	0b01000000,
	0b01100000,
	0b01111000,
	0b10011111,
}

var dutyRatio = [4]float32{
	0.125,
	0.250,
	0.500,
	0.750,
}

func fastsin(t float32) float32 {
	j := t * 0.15915
	j = j - float32(int(j))
	return 20.785 * j * (j - 0.5) * (j - 1.0)
}

func squareWave(dutyratio, frequency, amplitude float32, t float32) float32 {
	const pi = float32(math.Pi)
	const harmonics = 20

	var p = dutyratio * 2.0 * pi
	var a, b, n float32

	for n = 1; n <= harmonics; n++ {
		c := n * frequency * 2.0 * pi * t
		b += -fastsin(c-p*n) / n
		a += -fastsin(c) / n
	}

	return (2.0 * amplitude / pi) * (a - b)
}

type Square struct {
	enabled  bool
	sequence uint8
	reload   uint16
	timer    uint16
	sample   float32
	volume   uint8
	duty     uint8
}

func (s *Square) write(addr uint16, value byte) {
	switch addr & 0x0003 {
	case 0x0000:
		s.volume = value & 0x0F
		s.duty = value >> 6 & 0x03
		s.sequence = dutyTable[s.duty]
	case 0x0002:
		s.reload = s.reload&0xFF00 | uint16(value)
	case 0x0003:
		s.reload = s.reload&0x00FF | uint16(value&0x07)<<8
		s.timer = s.reload
	}
}

func (s *Square) tick(t float32) {
	frequency := 1789773.0 / (16.0 * (float32(s.reload) + 1.0))
	s.sample = squareWave(dutyRatio[s.duty], frequency, 1.0, t)
}

func (s *Square) output() float32 {
	if !s.enabled {
		return 0
	}

	return s.sample * float32(s.volume)
}
