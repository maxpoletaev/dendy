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
	Enabled  bool
	samples  chan float32
	time     float64
	cycle    uint64
	frame    uint64
	triangle triangle
	pulse1   square
	pulse2   square
	noise    noise
}

func New() *APU {
	apu := &APU{
		samples: make(chan float32, 48000),
	}

	return apu
}

func (a *APU) Reset() {
	a.time = 0
	a.cycle = 0
	a.frame = 0

	a.noise.reset()
	a.pulse1.reset()
	a.pulse2.reset()
	a.triangle.reset()
}

func (a *APU) Read(addr uint16) byte {
	return 0
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
	}
}

func (a *APU) mix(p1, p2, t, n, d float32) float32 {
	const (
		pWeight = 0.2
		tWeight = 0.3
		nWeight = 0.2
	)

	return p1*pWeight +
		p2*pWeight +
		t*tWeight +
		n*nWeight
}

func (a *APU) Output() float32 {
	p1 := a.pulse1.output()
	p2 := a.pulse2.output()
	t := a.triangle.output()
	n := a.noise.output()
	d := float32(0.0)

	return a.mix(p1, p2, t, n, d)
}

func (a *APU) Tick() {
	if !a.Enabled {
		return
	}

	a.time += 1.0 / 3.0 / 1789773.0 // Each APU tick is 1/3 CPU tick.
	t := float32(a.time)

	if a.cycle%6 == 0 {
		var (
			quarterFrame = a.frame%3729 == 0
			halfFrame    = a.frame%7457 == 0
		)

		if quarterFrame {
			a.pulse1.tickEnvelope()
			a.pulse2.tickEnvelope()
			a.noise.tickEnvelope()
			a.triangle.tickLinear()
		}

		if halfFrame {
			a.pulse1.tickLength()
			a.pulse2.tickLength()
			a.noise.tickLength()
			a.triangle.tickLength()
		}

		a.pulse1.tickTimer(t)
		a.pulse2.tickTimer(t)
		a.noise.tickTimer(t)
		a.triangle.tickTimer(t)

		a.frame++
		if a.frame == 14916 {
			a.frame = 0
		}
	}

	a.cycle++
}
