package apu

import (
	"encoding/gob"
	"errors"
)

var squareDuty = [4]float32{
	0.125,
	0.250,
	0.500,
	0.750,
}

func squareWave(duty, frequency, amplitude, t float32) float32 {
	phase := t*frequency - float32(int(t*frequency))
	if phase < duty {
		return amplitude
	}

	return -amplitude
}

// squareWaveFourier is a forier series approximation of a square wave, ported
// from the olcNES emulator. It might sound better, but apparently much slower,
// even with sine approximation. It’s here just for reference.
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
	enabled  bool
	sample   float32
	volume   uint8
	mode2    bool
	duty     uint8
	envelope envelope

	// Timer
	timerValue uint16
	timerLoad  uint16

	// Length counter
	lengthValue uint8
	lengthHalt  bool

	// Sweep
	sweepEnabled bool
	sweepValue   uint8
	sweepLoad    uint8
	sweepNegate  bool
	sweepShift   uint8
	sweepReload  bool
}

func (s *square) reset() {
	s.enabled = false
	s.sample = 0
	s.volume = 0
	s.mode2 = false
	s.duty = 0
	s.envelope.reset()

	s.timerValue = 0
	s.timerLoad = 0

	s.lengthValue = 0
	s.lengthHalt = false

	s.sweepEnabled = false
	s.sweepValue = 0
	s.sweepLoad = 0
	s.sweepNegate = false
	s.sweepShift = 0
	s.sweepReload = false
}

func (s *square) write(addr uint16, value byte) {
	switch addr & 0x0003 {
	case 0x0000:
		s.volume = value & 0b1111
		s.duty = value >> 6 & 0b11
		s.lengthHalt = (value>>5)&1 == 1
		s.envelope.loop = (value>>5)&1 == 1
		s.envelope.enabled = (value>>4)&1 == 0
		s.envelope.counterLoad = value & 0b1111
		s.envelope.start = true
	case 0x0001:
		s.sweepEnabled = value&0x80 != 0
		s.sweepNegate = value&0x08 != 0
		s.sweepLoad = value >> 4 & 0x07
		s.sweepShift = value & 0x07
		s.sweepReload = true
	case 0x0002:
		s.timerLoad = s.timerLoad&0xFF00 | uint16(value)
		s.timerValue = s.timerLoad
	case 0x0003:
		s.timerLoad = s.timerLoad&0x00FF | uint16(value&0x07)<<8
		s.lengthValue = lengthTable[value>>3]
		s.timerValue = s.timerLoad
	}
}

func (s *square) tickEnvelope() {
	s.envelope.tick()
}

func (s *square) tickSweep() {
	if s.sweepReload {
		s.sweepValue = s.sweepLoad
		s.sweepReload = false
	} else {
		if s.sweepValue > 0 {
			s.sweepValue--
		} else {
			s.sweepValue = s.sweepLoad

			if s.sweepEnabled && s.sweepShift > 0 {
				if s.sweepNegate {
					s.timerValue -= s.timerValue >> s.sweepShift

					if !s.mode2 {
						s.timerValue++
					}
				} else {
					s.timerValue += s.timerValue >> s.sweepShift
				}
			}
		}
	}
}

func (s *square) tickLength() {
	if !s.lengthHalt && s.lengthValue > 0 {
		s.lengthValue--
	}
}

func (s *square) tickTimer(t float32) {
	freq := 1789773.0 / (16.0 * (float32(s.timerValue) + 1.0))
	s.sample = squareWave(squareDuty[s.duty], freq, 1.0, t)
}

func (s *square) output() float32 {
	if !s.enabled || s.lengthValue == 0 || s.timerValue < 8 {
		return 0
	}

	vol := s.volume
	if s.envelope.enabled {
		vol = s.envelope.volume
	}

	return s.sample * float32(vol) / 15.0
}

func (s *square) save(enc *gob.Encoder) error {
	return errors.Join(
		s.envelope.save(enc),
		enc.Encode(s.enabled),
		enc.Encode(s.sample),
		enc.Encode(s.volume),
		enc.Encode(s.duty),
		enc.Encode(s.timerValue),
		enc.Encode(s.timerLoad),
		enc.Encode(s.lengthValue),
		enc.Encode(s.lengthHalt),
		enc.Encode(s.sweepEnabled),
		enc.Encode(s.sweepValue),
		enc.Encode(s.sweepLoad),
		enc.Encode(s.sweepNegate),
		enc.Encode(s.sweepShift),
		enc.Encode(s.sweepReload),
	)
}

func (s *square) load(dec *gob.Decoder) error {
	return errors.Join(
		s.envelope.load(dec),
		dec.Decode(&s.enabled),
		dec.Decode(&s.sample),
		dec.Decode(&s.volume),
		dec.Decode(&s.duty),
		dec.Decode(&s.timerValue),
		dec.Decode(&s.timerLoad),
		dec.Decode(&s.lengthValue),
		dec.Decode(&s.lengthHalt),
		dec.Decode(&s.sweepEnabled),
		dec.Decode(&s.sweepValue),
		dec.Decode(&s.sweepLoad),
		dec.Decode(&s.sweepNegate),
		dec.Decode(&s.sweepShift),
		dec.Decode(&s.sweepReload),
	)
}
