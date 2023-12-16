package apu

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
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

func (e *envelope) saveState(w *binario.Writer) error {
	return errors.Join(
		w.WriteBool(e.enabled),
		w.WriteBool(e.start),
		w.WriteBool(e.loop),
		w.WriteUint8(e.counterLoad),
		w.WriteUint8(e.counter),
	)
}

func (e *envelope) loadState(r *binario.Reader) error {
	return errors.Join(
		r.ReadBoolTo(&e.enabled),
		r.ReadBoolTo(&e.start),
		r.ReadBoolTo(&e.loop),
		r.ReadUint8To(&e.counterLoad),
		r.ReadUint8To(&e.counter),
	)
}
