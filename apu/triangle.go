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
	timerLoad  uint16
	timerValue uint16

	// Length counter
	lengthValue uint8
	lengthHalt  bool

	// Linear counter
	linearEnabled bool
	linearLoad    uint8
	linearValue   uint8
	linearReload  bool
}

func (tr *triangle) reset() {
	tr.enabled = false
	tr.sample = 0
	tr.sequence = 0

	tr.timerLoad = 0
	tr.timerValue = 0

	tr.lengthValue = 0
	tr.lengthHalt = false

	tr.linearEnabled = false
	tr.linearLoad = 0
	tr.linearValue = 0
	tr.linearReload = false
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
		tr.linearValue = tr.linearLoad
	} else if tr.linearValue > 0 {
		tr.linearValue--
	}

	if tr.linearEnabled {
		tr.linearReload = false
	}
}

func (tr *triangle) tickTimer(t float32) {
	if tr.lengthValue == 0 || tr.linearValue == 0 || tr.timerLoad < 3 {
		return
	}

	if tr.timerValue > 0 {
		tr.timerValue--
	} else {
		tr.sequence = (tr.sequence + 1) & 0x1F
		tr.timerValue = tr.timerLoad

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

	return tr.sample / 15.0
}

func (tr *triangle) save(enc *gob.Encoder) error {
	return errors.Join(
		enc.Encode(tr.enabled),
		enc.Encode(tr.sample),
		enc.Encode(tr.sequence),
		enc.Encode(tr.timerLoad),
		enc.Encode(tr.timerValue),
		enc.Encode(tr.lengthValue),
		enc.Encode(tr.lengthHalt),
		enc.Encode(tr.linearEnabled),
		enc.Encode(tr.linearLoad),
		enc.Encode(tr.linearValue),
		enc.Encode(tr.linearReload),
	)
}

func (tr *triangle) load(dec *gob.Decoder) error {
	return errors.Join(
		dec.Decode(&tr.enabled),
		dec.Decode(&tr.sample),
		dec.Decode(&tr.sequence),
		dec.Decode(&tr.timerLoad),
		dec.Decode(&tr.timerValue),
		dec.Decode(&tr.lengthValue),
		dec.Decode(&tr.lengthHalt),
		dec.Decode(&tr.linearEnabled),
		dec.Decode(&tr.linearLoad),
		dec.Decode(&tr.linearValue),
		dec.Decode(&tr.linearReload),
	)
}
