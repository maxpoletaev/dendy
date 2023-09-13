package apu

import (
	"encoding/gob"
	"errors"
)

type envelope struct {
	enabled   bool
	start     bool
	loop      bool
	value     uint8
	loadValue uint8
	volume    uint8
}

func (e *envelope) reset() {
	e.enabled = false
	e.start = false
	e.loop = false
	e.value = 0
	e.loadValue = 0
	e.volume = 0
}

func (e *envelope) tick() {
	if e.start {
		e.value = e.loadValue
		e.start = false
		e.value = 15
	} else if e.value > 0 {
		e.value--
	} else {
		if e.volume > 0 {
			e.value = e.loadValue
			e.volume--
		} else if e.loop {
			e.value = e.loadValue
			e.volume = 15
		}
	}
}

func (e *envelope) save(enc *gob.Encoder) error {
	return errors.Join(
		enc.Encode(e.enabled),
		enc.Encode(e.start),
		enc.Encode(e.loop),
		enc.Encode(e.value),
		enc.Encode(e.loadValue),
		enc.Encode(e.volume),
	)
}

func (e *envelope) load(dec *gob.Decoder) error {
	return errors.Join(
		dec.Decode(&e.enabled),
		dec.Decode(&e.start),
		dec.Decode(&e.loop),
		dec.Decode(&e.value),
		dec.Decode(&e.loadValue),
		dec.Decode(&e.volume),
	)
}
