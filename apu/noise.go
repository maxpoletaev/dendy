package apu

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
)

var noiseTable = [16]uint16{
	0, 4, 8, 16, 32, 64, 96, 128,
	160, 202, 254, 380, 508, 1016, 2034, 4068,
}

type noise struct {
	enabled  bool
	sample   uint8
	seq      uint16
	mode6    bool
	volume   uint8
	envelope envelope

	// Timer
	timerLoad uint16
	timer     uint16

	// Length counter
	lengthHalt bool
	length     uint8
}

func (n *noise) reset() {
	n.enabled = false
	n.sample = 0
	n.seq = 1
	n.mode6 = false
	n.volume = 0
	n.envelope.reset()

	n.timerLoad = 0
	n.timer = 0

	n.length = 0
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
		n.length = lengthTable[value>>3]
	}
}

func (n *noise) tickEnvelope() {
	n.envelope.tick()
}

func (n *noise) tickLength() {
	if !n.lengthHalt && n.length > 0 {
		n.length--
	}
}

func (n *noise) tickTimer() {
	if n.timer == 0 {
		shift := 1
		if n.mode6 {
			shift = 6
		}

		a := n.seq & 1
		b := (n.seq >> shift) & 1
		n.seq = n.seq>>1 | (a^b)<<14
		n.sample = uint8(n.seq & 1)
		n.timer = n.timerLoad
	} else {
		n.timer--
	}
}

func (n *noise) output() uint8 {
	if !n.enabled || n.length == 0 {
		return 0
	}

	return n.sample * n.volume
}

func (n *noise) saveState(w *binario.Writer) error {
	return errors.Join(
		n.envelope.saveState(w),
		w.WriteBool(n.enabled),
		w.WriteUint8(n.sample),
		w.WriteUint16(n.seq),
		w.WriteBool(n.mode6),
		w.WriteUint8(n.volume),
		w.WriteUint16(n.timerLoad),
		w.WriteUint16(n.timer),
		w.WriteUint8(n.length),
		w.WriteBool(n.lengthHalt),
	)
}

func (n *noise) loadState(r *binario.Reader) error {
	return errors.Join(
		n.envelope.loadState(r),
		r.ReadBoolTo(&n.enabled),
		r.ReadUint8To(&n.sample),
		r.ReadUint16To(&n.seq),
		r.ReadBoolTo(&n.mode6),
		r.ReadUint8To(&n.volume),
		r.ReadUint16To(&n.timerLoad),
		r.ReadUint16To(&n.timer),
		r.ReadUint8To(&n.length),
		r.ReadBoolTo(&n.lengthHalt),
	)
}
