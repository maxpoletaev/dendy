package apu

import (
	"encoding/gob"
	"errors"
	"math/bits"
)

var dmcTable = []uint16{
	214, 190, 170, 160, 143, 127, 113, 107, 95, 80, 71, 64, 53, 42, 36, 27,
}

type dmc struct {
	enabled    bool
	loop       bool
	irqEnabled bool
	irqPending bool

	timerLoad  uint16
	timer      uint16
	addrLoad   uint16
	addr       uint16
	lengthLoad uint16
	length     uint16

	bitsRemaining   uint8
	outputShifter   uint8
	buffer          uint8
	sample          uint8
	isBufferEmpty   bool
	isOutputSilence bool

	dmaRead func(addr uint16) byte
	reverse bool
}

func (d *dmc) reset() {
	d.enabled = false
	d.loop = false
	d.irqEnabled = false
	d.irqPending = false

	d.timer = 0
	d.timerLoad = 0
	d.addrLoad = 0
	d.addr = 0
	d.lengthLoad = 0
	d.length = 0

	d.bitsRemaining = 0
	d.outputShifter = 0
	d.buffer = 0
	d.sample = 0
	d.isBufferEmpty = true
	d.isOutputSilence = true
}

func (d *dmc) write(addr uint16, value byte) {
	switch addr {
	case 0x4010:
		d.timerLoad = dmcTable[value&0b00001111]
		d.irqEnabled = value&0b10000000 != 0
		d.loop = value&0b01000000 != 0

		if !d.irqEnabled {
			d.irqPending = false
		}
	case 0x4011:
		d.sample = value & 0b01111111
	case 0x4012:
		d.addrLoad = 0xC000 | uint16(value)<<6
	case 0x4013:
		d.lengthLoad = uint16(value)<<4 + 1
	case 0x4015:
		d.enabled = value&0b00010000 != 0
		d.irqPending = false

		if d.enabled {
			if d.length == 0 {
				d.length = d.lengthLoad
				d.addr = d.addrLoad
			}
		} else {
			d.length = 0
		}
	}
}

func (d *dmc) tickTimer() {
	if d.timer > 0 {
		d.timer--
	} else {
		if !d.isOutputSilence {
			if d.outputShifter&1 != 0 {
				if d.sample <= 0x7D {
					d.sample += 2
				}
			} else {
				if d.sample >= 0x02 {
					d.sample -= 2
				}
			}
		}

		d.timer = d.timerLoad
		d.outputShifter >>= 1
		d.bitsRemaining--

		if d.bitsRemaining == 0 {
			d.bitsRemaining = 8
			d.outputShifter = d.buffer
			d.isOutputSilence = d.isBufferEmpty
			d.isBufferEmpty = true
		}
	}

	if d.length > 0 && d.isBufferEmpty {
		d.buffer = d.dmaRead(d.addr)

		if d.reverse {
			// Reverse sample bits (sometimes may sound better).
			d.buffer = bits.Reverse8(d.buffer)
		}

		d.addr = (d.addr + 1) | 0x8000
		d.isBufferEmpty = false
		d.length--

		if d.length == 0 {
			if d.loop {
				d.length = d.lengthLoad
				d.addr = d.addrLoad
			}
		} else if d.irqEnabled {
			d.irqPending = true
		}
	}
}

func (d *dmc) output() float32 {
	return float32(d.sample) / 15.0
}

func (d *dmc) save(enc *gob.Encoder) error {
	return errors.Join(
		enc.Encode(d.enabled),
		enc.Encode(d.loop),
		enc.Encode(d.irqEnabled),
		enc.Encode(d.irqPending),
		enc.Encode(d.timerLoad),
		enc.Encode(d.timer),
		enc.Encode(d.addrLoad),
		enc.Encode(d.addr),
		enc.Encode(d.lengthLoad),
		enc.Encode(d.length),
		enc.Encode(d.bitsRemaining),
		enc.Encode(d.outputShifter),
		enc.Encode(d.buffer),
		enc.Encode(d.sample),
		enc.Encode(d.isBufferEmpty),
		enc.Encode(d.isOutputSilence),
	)
}

func (d *dmc) load(dec *gob.Decoder) error {
	return errors.Join(
		dec.Decode(&d.enabled),
		dec.Decode(&d.loop),
		dec.Decode(&d.irqEnabled),
		dec.Decode(&d.irqPending),
		dec.Decode(&d.timerLoad),
		dec.Decode(&d.timer),
		dec.Decode(&d.addrLoad),
		dec.Decode(&d.addr),
		dec.Decode(&d.lengthLoad),
		dec.Decode(&d.length),
		dec.Decode(&d.bitsRemaining),
		dec.Decode(&d.outputShifter),
		dec.Decode(&d.buffer),
		dec.Decode(&d.sample),
		dec.Decode(&d.isBufferEmpty),
		dec.Decode(&d.isOutputSilence),
	)
}
