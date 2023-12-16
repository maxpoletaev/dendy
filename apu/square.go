package apu

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
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

func (s *square) saveState(w *binario.Writer) error {
	return errors.Join(
		s.envelope.saveState(w),
		w.WriteBool(s.enabled),
		w.WriteBool(s.isPulse1),
		w.WriteUint8(s.sample),
		w.WriteUint8(s.volume),
		w.WriteUint8(s.duty),
		w.WriteUint8(s.dutyBit),
		w.WriteUint16(s.timerLoad),
		w.WriteUint16(s.timer),
		w.WriteUint8(s.length),
		w.WriteBool(s.lengthHalt),
		w.WriteBool(s.sweepEnabled),
		w.WriteBool(s.sweepNegate),
		w.WriteUint8(s.sweepShift),
		w.WriteUint8(s.sweepLoad),
		w.WriteBool(s.sweepReload),
		w.WriteUint8(s.sweep),
	)
}

func (s *square) loadState(r *binario.Reader) error {
	return errors.Join(
		s.envelope.loadState(r),
		r.ReadBoolTo(&s.enabled),
		r.ReadBoolTo(&s.isPulse1),
		r.ReadUint8To(&s.sample),
		r.ReadUint8To(&s.volume),
		r.ReadUint8To(&s.duty),
		r.ReadUint8To(&s.dutyBit),
		r.ReadUint16To(&s.timerLoad),
		r.ReadUint16To(&s.timer),
		r.ReadUint8To(&s.length),
		r.ReadBoolTo(&s.lengthHalt),
		r.ReadBoolTo(&s.sweepEnabled),
		r.ReadBoolTo(&s.sweepNegate),
		r.ReadUint8To(&s.sweepShift),
		r.ReadUint8To(&s.sweepLoad),
		r.ReadBoolTo(&s.sweepReload),
		r.ReadUint8To(&s.sweep),
	)
}
