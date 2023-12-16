package apu

import (
	"encoding/gob"
	"errors"
)

var squareDutyTable = [4]byte{
	0b01000000,
	0b01100000,
	0b01111000,
	0b10011111,
}

type square struct {
	enabled  bool
	isPulse1 bool
	sample   uint8
	volume   uint8
	duty     uint8
	dutyBit  uint8
	envelope envelope

	// Timer
	timerLoad uint16
	timer     uint16

	// Length counter
	lengthHalt bool
	length     uint8

	// Sweep
	sweepEnabled bool
	sweepLoad    uint8
	sweepNegate  bool
	sweepShift   uint8
	sweepReload  bool
	sweep        uint8
}

func (s *square) reset() {
	s.enabled = false
	s.sample = 0
	s.volume = 0
	s.isPulse1 = false
	s.duty = 0
	s.envelope.reset()

	s.timer = 0
	s.timerLoad = 0

	s.length = 0
	s.lengthHalt = false

	s.sweepEnabled = false
	s.sweepReload = false
	s.sweepNegate = false
	s.sweepShift = 0
	s.sweepLoad = 0
	s.sweep = 0
}

func (s *square) write(addr uint16, value byte) {
	switch addr & 0x0003 {
	case 0x0000:
		s.volume = value & 0b1111
		s.lengthHalt = (value>>5)&1 == 1
		s.envelope.loop = (value>>5)&1 == 1
		s.duty = squareDutyTable[(value>>6)&0b11]
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
	case 0x0003:
		s.timerLoad = s.timerLoad&0x00FF | uint16(value&0x07)<<8
		s.length = lengthTable[value>>3]
		s.envelope.start = true
		s.dutyBit = 0
	}
}

func (s *square) tickEnvelope() {
	s.envelope.tick()
}

func (s *square) tickSweep() {
	if s.sweepReload {
		s.sweep = s.sweepLoad
		s.sweepReload = false
	} else {
		if s.sweep > 0 {
			s.sweep--
		} else {
			s.sweep = s.sweepLoad

			if s.sweepEnabled && s.sweepShift > 0 {
				if s.sweepNegate {
					s.timerLoad -= s.timerLoad >> s.sweepShift
					if s.isPulse1 {
						s.timerLoad++
					}
				} else {
					s.timerLoad += s.timerLoad >> s.sweepShift
				}
			}
		}
	}
}

func (s *square) tickLength() {
	if !s.lengthHalt && s.length > 0 {
		s.length--
	}
}

func (s *square) tickTimer() {
	if s.timer > 0 {
		s.timer--
	} else {
		s.timer = s.timerLoad
		s.dutyBit = (s.dutyBit + 1) & 0b111
		s.sample = (s.duty >> s.dutyBit) & 1
	}
}

func (s *square) output() uint8 {
	if !s.enabled || s.length == 0 || s.timer < 8 {
		return 0
	}

	vol := s.volume
	if s.envelope.enabled {
		vol = s.envelope.volume
	}

	return s.sample * vol
}

func (s *square) save(enc *gob.Encoder) error {
	return errors.Join(
		s.envelope.save(enc),
		enc.Encode(s.enabled),
		enc.Encode(s.sample),
		enc.Encode(s.volume),
		enc.Encode(s.duty),
		enc.Encode(s.dutyBit),
		enc.Encode(s.timer),
		enc.Encode(s.timerLoad),
		enc.Encode(s.length),
		enc.Encode(s.lengthHalt),
		enc.Encode(s.sweepEnabled),
		enc.Encode(s.sweep),
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
		dec.Decode(&s.dutyBit),
		dec.Decode(&s.timer),
		dec.Decode(&s.timerLoad),
		dec.Decode(&s.length),
		dec.Decode(&s.lengthHalt),
		dec.Decode(&s.sweepEnabled),
		dec.Decode(&s.sweep),
		dec.Decode(&s.sweepLoad),
		dec.Decode(&s.sweepNegate),
		dec.Decode(&s.sweepShift),
		dec.Decode(&s.sweepReload),
	)
}
