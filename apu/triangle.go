package apu

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
)

type triangle struct {
	enabled  bool
	sample   uint8
	sequence uint8

	// Timer
	timerLoad uint16
	timer     uint16

	// Length counter
	lengthHalt bool
	length     uint8

	// Linear counter
	linearEnabled bool
	linearReset   bool
	linearLoad    uint8
	linear        uint8
}

func (t *triangle) reset() {
	t.enabled = false
	t.sample = 0
	t.sequence = 0

	t.timerLoad = 0
	t.timer = 0

	t.lengthHalt = false
	t.length = 0

	t.linearEnabled = false
	t.linearReset = false
	t.linearLoad = 0
	t.linear = 0
}

func (t *triangle) write(addr uint16, value byte) {
	switch addr {
	case 0x4008:
		t.linearLoad = value & 0x7F
		t.lengthHalt = value&0x80 == 0
		t.linearEnabled = value&0x80 == 0
	case 0x400A:
		t.timerLoad = t.timerLoad&0xFF00 | uint16(value)
	case 0x400B:
		t.timerLoad = t.timerLoad&0x00FF | uint16(value&0x07)<<8
		t.length = lengthTable[value>>3]
		t.linearReset = true
	}
}

func (t *triangle) tickLength() {
	if !t.lengthHalt && t.length > 0 {
		t.length--
	}
}

func (t *triangle) tickLinear() {
	if t.linearReset {
		t.linear = t.linearLoad
	} else if t.linear > 0 {
		t.linear--
	}

	if t.linearEnabled {
		t.linearReset = false
	}
}

func (t *triangle) tickTimer() {
	if t.length == 0 || t.linear == 0 || t.timerLoad < 3 {
		return
	}

	if t.timer > 0 {
		t.timer--
	} else {
		t.sequence = (t.sequence + 1) & 0x1F
		t.timer = t.timerLoad

		if t.sequence&0x10 == 0 {
			t.sample = t.sequence ^ 0x1F
		} else {
			t.sample = t.sequence
		}
	}
}

func (t *triangle) output() uint8 {
	if !t.enabled {
		return 0
	}

	return t.sample
}

func (t *triangle) saveState(w *binario.Writer) error {
	return errors.Join(
		w.WriteBool(t.enabled),
		w.WriteUint8(t.sample),
		w.WriteUint8(t.sequence),
		w.WriteUint16(t.timerLoad),
		w.WriteUint16(t.timer),
		w.WriteUint8(t.length),
		w.WriteBool(t.lengthHalt),
		w.WriteBool(t.linearEnabled),
		w.WriteUint8(t.linearLoad),
		w.WriteUint8(t.linear),
		w.WriteBool(t.linearReset),
	)
}

func (t *triangle) loadState(r *binario.Reader) error {
	return errors.Join(
		r.ReadBoolTo(&t.enabled),
		r.ReadUint8To(&t.sample),
		r.ReadUint8To(&t.sequence),
		r.ReadUint16To(&t.timerLoad),
		r.ReadUint16To(&t.timer),
		r.ReadUint8To(&t.length),
		r.ReadBoolTo(&t.lengthHalt),
		r.ReadBoolTo(&t.linearEnabled),
		r.ReadUint8To(&t.linearLoad),
		r.ReadUint8To(&t.linear),
		r.ReadBoolTo(&t.linearReset),
	)
}
