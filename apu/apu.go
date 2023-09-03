package apu

type APU struct {
	samples chan float32
	time    float32
	cycle   uint64
	frame   uint64
	pulse1  Square
	pulse2  Square
}

func New() *APU {
	return &APU{
		samples: make(chan float32, 48000),
	}
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
	case addr == 0x4015:
		a.pulse1.enabled = value&0x01 == 1
		a.pulse2.enabled = value&0x02 == 2
	}
}

func (a *APU) Output() float32 {
	p1 := a.pulse1.output()
	p2 := a.pulse2.output()
	return (p1-0.5)*0.5 + (p2-0.5)*0.5
}

func (a *APU) Tick() {
	a.time += 0.333333 / 1789773.0
	if a.time >= 1.0 {
		a.time -= 1.0
	}

	if a.cycle%6 == 0 {
		a.pulse1.tick(a.time)
		a.pulse2.tick(a.time)
		a.frame++
	}

	a.cycle++
}
