package apu

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
)

var lengthTable = []byte{
	10, 254, 20, 2, 40, 4, 80, 6, 160, 8, 60, 10, 14, 12, 26, 14,
	12, 16, 24, 18, 48, 20, 96, 22, 192, 24, 72, 26, 16, 28, 32, 30,
}

type APU struct {
	Enabled    bool
	PendingIRQ bool

	mode     uint8
	cycle    uint64
	frame    uint64 // not the same as ppu frame
	pulse1   square
	pulse2   square
	noise    noise
	dmc      dmc
	triangle triangle
	filters  []*filter

	irqDisable bool
	frameIRQ   bool
}

func New() *APU {
	return &APU{
		Enabled: true,
		filters: []*filter{
			highPassFilter(44100.0, 90.0),
			lowPassFilter(44100.0, 14000.0),
		},
	}
}

func (a *APU) Reset() {
	a.mode = 0
	a.cycle = 0
	a.frame = 0

	a.dmc.reset()
	a.noise.reset()
	a.pulse1.reset()
	a.pulse1.isPulse1 = true
	a.pulse2.reset()
	a.triangle.reset()
}

func (a *APU) Read(addr uint16) (status byte) {
	if addr == 0x4015 {
		if a.pulse1.length > 0 {
			status |= 1 << 0
		}

		if a.pulse2.length > 0 {
			status |= 1 << 1
		}

		if a.triangle.length > 0 {
			status |= 1 << 2
		}

		if a.noise.length > 0 {
			status |= 1 << 3
		}

		if a.dmc.length > 0 {
			status |= 1 << 4
		}

		if a.frameIRQ {
			status |= 1 << 6
			a.frameIRQ = false
		}

		if a.dmc.irqPending {
			status |= 1 << 7
		}
	}

	return status
}

func (a *APU) Write(addr uint16, value byte) {
	switch {
	case addr >= 0x4000 && addr <= 0x4003:
		a.pulse1.write(addr, value)
	case addr >= 0x4004 && addr <= 0x4007:
		a.pulse2.write(addr, value)
	case addr >= 0x4008 && addr <= 0x400B:
		a.triangle.write(addr, value)
	case addr >= 0x400C && addr <= 0x400F:
		a.noise.write(addr, value)
	case addr >= 0x4010 && addr <= 0x4013:
		a.dmc.write(addr, value)
	case addr == 0x4015:
		a.pulse1.enabled = value&0x01 != 0
		a.pulse2.enabled = value&0x02 != 0
		a.triangle.enabled = value&0x03 != 0
		a.noise.enabled = value&0x04 != 0
		a.dmc.write(addr, value)
	case addr == 0x4017:
		a.frame = 0
		a.mode = (value & 0x80) >> 7
		a.irqDisable = value&0x40 != 0
	}
}

func (a *APU) mix(p1, p2, t, n, d uint8) float32 {
	tndOut := 0.00851*float32(t) + 0.00494*float32(n) + 0.00335*float32(d)
	pulseOut := 0.00752 * (float32(p1) + float32(p2))
	return pulseOut + tndOut
}

func (a *APU) Output() float32 {
	if !a.Enabled {
		return 0
	}

	p1 := a.pulse1.output()
	p2 := a.pulse2.output()
	t := a.triangle.output()
	n := a.noise.output()
	d := a.dmc.output()

	out := a.mix(p1, p2, t, n, d)
	for _, f := range a.filters {
		out = f.do(out)
	}

	return clamp(out, -1, 1)
}

func (a *APU) Tick() {
	if !a.Enabled {
		return
	}

	// Triangle is clocked at CPU speed.
	a.triangle.tickTimer()

	// Everything else is clocked at half CPU speed.
	if a.cycle%2 == 0 {
		var quarterFrame, halfFrame bool
		var maxFrame uint64

		if a.mode == 0 {
			quarterFrame = a.frame == 3728 || a.frame == 7456 || a.frame == 11185 || a.frame == 14914
			halfFrame = a.frame == 7456 || a.frame == 14914
			maxFrame = 14915
		} else {
			quarterFrame = a.frame == 3728 || a.frame == 7456 || a.frame == 11185 || a.frame == 18640
			halfFrame = a.frame == 7456 || a.frame == 18640
			maxFrame = 18641
		}

		if quarterFrame {
			a.pulse1.tickEnvelope()
			a.pulse2.tickEnvelope()
			a.noise.tickEnvelope()
			a.triangle.tickLinear()
		}

		if halfFrame {
			a.pulse1.tickLength()
			a.pulse1.tickSweep()
			a.pulse2.tickLength()
			a.pulse2.tickSweep()
			a.noise.tickLength()
			a.triangle.tickLength()

			if a.mode == 0 && !a.irqDisable {
				a.PendingIRQ = true
				a.frameIRQ = true
			}
		}

		a.pulse1.tickTimer()
		a.pulse2.tickTimer()
		a.noise.tickTimer()
		a.dmc.tickTimer()

		a.frame++
		if a.frame == maxFrame {
			a.frame = 0
		}
	}

	a.cycle++
}

func (a *APU) SetDMACallback(cb func(addr uint16) byte) {
	a.dmc.dmaCallback = cb
}

func (a *APU) SaveState(w *binario.Writer) error {
	return errors.Join(
		a.pulse1.saveState(w),
		a.pulse2.saveState(w),
		a.triangle.saveState(w),
		a.noise.saveState(w),
		a.dmc.saveState(w),
		w.WriteBool(a.PendingIRQ),
		w.WriteUint8(a.mode),
		w.WriteUint64(a.cycle),
		w.WriteUint64(a.frame),
		w.WriteBool(a.irqDisable),
		w.WriteBool(a.frameIRQ),
	)
}

func (a *APU) LoadState(r *binario.Reader) error {
	return errors.Join(
		a.pulse1.loadState(r),
		a.pulse2.loadState(r),
		a.triangle.loadState(r),
		a.noise.loadState(r),
		a.dmc.loadState(r),
		r.ReadBoolTo(&a.PendingIRQ),
		r.ReadUint8To(&a.mode),
		r.ReadUint64To(&a.cycle),
		r.ReadUint64To(&a.frame),
		r.ReadBoolTo(&a.irqDisable),
		r.ReadBoolTo(&a.frameIRQ),
	)
}

func clamp(v float32, min, max float32) float32 {
	if v < min {
		return min
	} else if v > max {
		return max
	}
	return v
}
