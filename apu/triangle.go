package apu

import (
	"encoding/gob"
	"errors"
)

type triangle struct {
	enabled  bool
	sample   float32
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

func (tr *triangle) reset() {
	tr.enabled = false
	tr.sample = 0
	tr.sequence = 0

	tr.timerLoad = 0
	tr.timer = 0

	tr.lengthHalt = false
	tr.length = 0

	tr.linearEnabled = false
	tr.linearReset = false
	tr.linearLoad = 0
	tr.linear = 0
}

func (tr *triangle) write(addr uint16, value byte) {
	switch addr {
	case 0x4008:
		tr.linearLoad = value & 0x7F
		tr.lengthHalt = value&0x80 == 0
		tr.linearEnabled = value&0x80 == 0
	case 0x400A:
		tr.timerLoad = tr.timerLoad&0xFF00 | uint16(value)
	case 0x400B:
		tr.timerLoad = tr.timerLoad&0x00FF | uint16(value&0x07)<<8
		tr.length = lengthTable[value>>3]
		tr.linearReset = true
	}
}

func (tr *triangle) tickLength() {
	if !tr.lengthHalt && tr.length > 0 {
		tr.length--
	}
}

func (tr *triangle) tickLinear() {
	if tr.linearReset {
		tr.linear = tr.linearLoad
	} else if tr.linear > 0 {
		tr.linear--
	}

	if tr.linearEnabled {
		tr.linearReset = false
	}
}

func (tr *triangle) tickTimer() {
	if tr.length == 0 || tr.linear == 0 || tr.timerLoad < 3 {
		return
	}

	if tr.timer > 0 {
		tr.timer--
	} else {
		tr.sequence = (tr.sequence + 1) & 0x1F
		tr.timer = tr.timerLoad

		if tr.sequence&0x10 == 0 {
			tr.sample = float32(tr.sequence ^ 0x1F)
		} else {
			tr.sample = float32(tr.sequence)
		}
	}
}

func (tr *triangle) output() float32 {
	if !tr.enabled {
		return 0
	}

	return tr.sample
}

func (tr *triangle) save(enc *gob.Encoder) error {
	return errors.Join(
		enc.Encode(tr.enabled),
		enc.Encode(tr.sample),
		enc.Encode(tr.sequence),
		enc.Encode(tr.timerLoad),
		enc.Encode(tr.timer),
		enc.Encode(tr.length),
		enc.Encode(tr.lengthHalt),
		enc.Encode(tr.linearEnabled),
		enc.Encode(tr.linearLoad),
		enc.Encode(tr.linear),
		enc.Encode(tr.linearReset),
	)
}

func (tr *triangle) load(dec *gob.Decoder) error {
	return errors.Join(
		dec.Decode(&tr.enabled),
		dec.Decode(&tr.sample),
		dec.Decode(&tr.sequence),
		dec.Decode(&tr.timerLoad),
		dec.Decode(&tr.timer),
		dec.Decode(&tr.length),
		dec.Decode(&tr.lengthHalt),
		dec.Decode(&tr.linearEnabled),
		dec.Decode(&tr.linearLoad),
		dec.Decode(&tr.linear),
		dec.Decode(&tr.linearReset),
	)
}
