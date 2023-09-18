package apu

import (
	"encoding/gob"
	"errors"
)

var noiseTable = [16]uint16{
	0, 4, 8, 16, 32, 64, 96, 128,
	160, 202, 254, 380, 508, 1016, 2034, 4068,
}

type noise struct {
	enabled  bool
	sample   float32
	seq      uint16
	mode6    bool
	volume   uint8
	envelope envelope

	// Timer
	timerLoad  uint16
	timerValue uint16

	// Length counter
	lengthValue uint8
	lengthHalt  bool
}

func (n *noise) reset() {
	n.enabled = false
	n.sample = 0
	n.seq = 1
	n.mode6 = false
	n.volume = 0
	n.envelope.reset()

	n.timerLoad = 0
	n.timerValue = 0

	n.lengthValue = 0
	n.lengthHalt = false
}

func (n *noise) write(addr uint16, value byte) {
	switch addr {
	case 0x400E:
		n.timerLoad = noiseTable[value&0x0F]
		n.mode6 = value&0x80 != 0
	case 0x400C:
		n.lengthHalt = value&0x20 != 0
		n.volume = value & 0x0F

	case 0x400F:
		n.lengthValue = lengthTable[value>>3]
	}
}

func (n *noise) tickEnvelope() {
	n.envelope.tick()
}

func (n *noise) tickLength() {
	if !n.lengthHalt && n.lengthValue > 0 {
		n.lengthValue--
	}
}

func (n *noise) tickTimer(t float32) {
	if n.timerValue == 0 {
		shift := 1
		if n.mode6 {
			shift = 6
		}

		a := n.seq & 1
		b := (n.seq >> shift) & 1
		n.seq = n.seq>>1 | (a^b)<<14
		n.sample = float32(n.seq & 1)
		n.timerValue = n.timerLoad
	} else {
		n.timerValue--
	}
}

func (n *noise) output() float32 {
	if !n.enabled || n.lengthValue == 0 {
		return 0
	}

	return n.sample * float32(n.volume) / 15.0
}

func (n *noise) save(enc *gob.Encoder) error {
	return errors.Join(
		n.envelope.save(enc),
		enc.Encode(n.enabled),
		enc.Encode(n.sample),
		enc.Encode(n.seq),
		enc.Encode(n.mode6),
		enc.Encode(n.volume),
		enc.Encode(n.timerLoad),
		enc.Encode(n.timerValue),
		enc.Encode(n.lengthValue),
		enc.Encode(n.lengthHalt),
	)
}

func (n *noise) load(dec *gob.Decoder) error {
	return errors.Join(
		n.envelope.load(dec),
		dec.Decode(&n.enabled),
		dec.Decode(&n.sample),
		dec.Decode(&n.seq),
		dec.Decode(&n.mode6),
		dec.Decode(&n.volume),
		dec.Decode(&n.timerLoad),
		dec.Decode(&n.timerValue),
		dec.Decode(&n.lengthValue),
		dec.Decode(&n.lengthHalt),
	)
}
