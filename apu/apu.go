package apu

import "math"

const (
	pi = float32(math.Pi)
)

var lengthTable = []byte{
	10, 254, 20, 2, 40, 4, 80, 6, 160, 8, 60, 10, 14, 12, 26, 14,
	12, 16, 24, 18, 48, 20, 96, 22, 192, 24, 72, 26, 16, 28, 32, 30,
}

type APU struct {
	Enabled    bool
	PendingIRQ bool

	mode     uint8
	time     float64
	cycle    uint64
	frame    uint64
	triangle triangle
	pulse1   square
	pulse2   square
	noise    noise
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
	a.time = 0
	a.cycle = 0
	a.frame = 0
	a.mode = 0

	a.noise.reset()
	a.pulse1.reset()
	a.pulse2.reset()
	a.triangle.reset()
}

func (a *APU) Read(addr uint16) (status byte) {
	if addr == 0x4015 {
		if a.pulse1.lengthValue > 0 {
			status |= 0b0000_0001
		}

		if a.pulse2.lengthValue > 0 {
			status |= 0b0000_0010
		}

		if a.triangle.lengthValue > 0 {
			status |= 0b0000_0100
		}

		if a.noise.lengthValue > 0 {
			status |= 0b0000_1000
		}

		if a.frameIRQ {
			status |= 0b0100_0000
			a.frameIRQ = false
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
	case addr == 0x4015:
		a.pulse1.enabled = value&0x01 != 0
		a.pulse2.enabled = value&0x02 != 0
		a.triangle.enabled = value&0x03 != 0
		a.noise.enabled = value&0x04 != 0
	case addr == 0x4017:
		a.frame = 0
		a.mode = (value & 0x80) >> 7
		a.irqDisable = value&0x40 != 0
	}
}

func (a *APU) mix(p1, p2, t, n, d float32) float32 {
	const (
		pWeight = 0.2
		tWeight = 0.2
		nWeight = 0.2
	)

	return p1*pWeight + p2*pWeight + t*tWeight + n*nWeight
}

func (a *APU) Output() float32 {
	if !a.Enabled {
		return 0
	}

	p1 := a.pulse1.output()
	p2 := a.pulse2.output()
	t := a.triangle.output()
	n := a.noise.output()
	d := float32(0.0)

	out := a.mix(p1, p2, t, n, d)
	for _, f := range a.filters {
		out = f.do(out)
	}

	return out
}

func (a *APU) Tick() {
	if !a.Enabled {
		return
	}

	// One tick is 1/1789773 seconds.
	a.time += 1.0 / 1789773.0
	t := float32(a.time)

	// Triangle is clocked at CPU speed.
	a.triangle.tickTimer(t)

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

		a.pulse1.tickTimer(t)
		a.pulse2.tickTimer(t)
		a.noise.tickTimer(t)

		a.frame++
		if a.frame == maxFrame {
			a.frame = 0
		}
	}

	a.cycle++
}
