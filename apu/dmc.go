package apu

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
)

var dmcTimerTable = []uint16{
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

	bitsLeft uint8
	shifter  uint8
	buffer   uint8
	sample   uint8
	isEmpty  bool
	isSilent bool

	dmaCallback func(addr uint16) byte
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

	d.bitsLeft = 0
	d.shifter = 0
	d.buffer = 0
	d.sample = 0
	d.isEmpty = true
	d.isSilent = true
}

func (d *dmc) write(addr uint16, value byte) {
	switch addr {
	case 0x4010:
		d.timerLoad = dmcTimerTable[value&0b1111]
		d.irqEnabled = (value>>7)&1 != 0
		d.loop = (value>>6)&1 != 0

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
		d.enabled = (value>>4)&1 != 0
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
		if !d.isSilent {
			if d.shifter&1 != 0 {
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
		d.shifter >>= 1
		d.bitsLeft--

		if d.bitsLeft == 0 {
			d.bitsLeft = 8
			d.shifter = d.buffer
			d.isSilent = d.isEmpty
			d.isEmpty = true
		}
	}

	if d.length > 0 && d.isEmpty {
		d.buffer = d.dmaCallback(d.addr)
		d.addr = (d.addr + 1) | 0x8000
		d.isEmpty = false
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

func (d *dmc) output() uint8 {
	return d.sample
}

func (d *dmc) saveState(w *binario.Writer) error {
	return errors.Join(
		w.WriteBool(d.enabled),
		w.WriteBool(d.loop),
		w.WriteBool(d.irqEnabled),
		w.WriteBool(d.irqPending),
		w.WriteUint16(d.timerLoad),
		w.WriteUint16(d.timer),
		w.WriteUint16(d.addrLoad),
		w.WriteUint16(d.addr),
		w.WriteUint16(d.lengthLoad),
		w.WriteUint16(d.length),
		w.WriteUint8(d.bitsLeft),
		w.WriteUint8(d.shifter),
		w.WriteUint8(d.buffer),
		w.WriteUint8(d.sample),
		w.WriteBool(d.isEmpty),
		w.WriteBool(d.isSilent),
	)
}

func (d *dmc) loadState(r *binario.Reader) error {
	return errors.Join(
		r.ReadBoolTo(&d.enabled),
		r.ReadBoolTo(&d.loop),
		r.ReadBoolTo(&d.irqEnabled),
		r.ReadBoolTo(&d.irqPending),
		r.ReadUint16To(&d.timerLoad),
		r.ReadUint16To(&d.timer),
		r.ReadUint16To(&d.addrLoad),
		r.ReadUint16To(&d.addr),
		r.ReadUint16To(&d.lengthLoad),
		r.ReadUint16To(&d.length),
		r.ReadUint8To(&d.bitsLeft),
		r.ReadUint8To(&d.shifter),
		r.ReadUint8To(&d.buffer),
		r.ReadUint8To(&d.sample),
		r.ReadBoolTo(&d.isEmpty),
		r.ReadBoolTo(&d.isSilent),
	)
}
