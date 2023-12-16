package apu

import (
	"encoding/gob"
	"errors"
)

type envelope struct {
	enabled     bool
	start       bool
	loop        bool
	counterLoad uint8
	counter     uint8
	volume      uint8
}

func (e *envelope) reset() {
	e.enabled = false
	e.start = false
	e.loop = false
	e.counterLoad = 0
	e.volume = 0
	e.counter = 0
}

func (e *envelope) tick() {
	if e.start {
		e.counter = e.counterLoad
		e.volume = 0x0F
		e.start = false
	} else {
		if e.counter > 0 {
			e.counter--
		} else {
			e.counter = e.counterLoad

			if e.volume > 0 {
				e.volume--
			} else if e.loop {
				e.volume = 0x0F
			}
		}
	}
}

func (e *envelope) save(enc *gob.Encoder) error {
	return errors.Join(
		enc.Encode(e.enabled),
		enc.Encode(e.start),
		enc.Encode(e.loop),
		enc.Encode(e.counter),
		enc.Encode(e.counterLoad),
		enc.Encode(e.volume),
	)
}

func (e *envelope) load(dec *gob.Decoder) error {
	return errors.Join(
		dec.Decode(&e.enabled),
		dec.Decode(&e.start),
		dec.Decode(&e.loop),
		dec.Decode(&e.counter),
		dec.Decode(&e.counterLoad),
		dec.Decode(&e.volume),
	)
}
