package nes

import (
	cpupkg "github.com/maxpoletaev/dendy/cpu"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	ppupkg "github.com/maxpoletaev/dendy/ppu"
)

type TickInfo struct {
	InstrComplete    bool
	ScanlineComplete bool
	FrameComplete    bool
}

type Bus struct {
	RAM    [2048]uint8
	CPU    *cpupkg.CPU
	PPU    *ppupkg.PPU
	Cart   ines.Cartridge
	Joy1   *input.Joystick
	Joy2   *input.Joystick
	Zap    *input.Zapper
	cycles uint64
}

func (b *Bus) transferOAM(addr uint8) {
	memAddr := uint16(addr) << 8
	for i := uint16(0); i < 256; i++ {
		b.PPU.WriteOAM(b.Read(memAddr + i))
	}

	b.CPU.Halt += 513
	if b.CPU.Halt%2 == 1 {
		b.CPU.Halt++
	}
}

func (b *Bus) Read(addr uint16) uint8 {
	switch {
	case addr <= 0x1FFF: // Internal RAM.
		return b.RAM[addr%0x0800]
	case addr <= 0x3FFF: // PPU registers.
		return b.PPU.Read(addr)
	case addr == 0x4014: // PPU OAM DMA.
		return b.PPU.Read(addr)
	case addr == 0x4016: // Controller 1.
		return b.Joy1.Read()
	case addr <= 0x4017: // Controller 2 or Zapper.
		if b.Joy2 != nil {
			return b.Joy2.Read()
		} else if b.Zap != nil {
			return b.Zap.Read()
		}
		return 0
	case addr <= 0x401F: // APU and I/O functionality.
		return 0
	default: // Cartridge space.
		return b.Cart.ReadPRG(addr)
	}
}

func (b *Bus) Write(addr uint16, data uint8) {
	switch {
	case addr <= 0x1FFF: // Internal RAM.
		b.RAM[addr%0x0800] = data
	case addr <= 0x3FFF: // PPU registers.
		b.PPU.Write(addr, data)
	case addr == 0x4014: // PPU OAM direct access.
		b.transferOAM(data)
	case addr == 0x4016: // Controller strobe.
		if b.Joy1 != nil {
			b.Joy1.Write(data)
		}
		if b.Joy2 != nil {
			b.Joy2.Write(data)
		}
	case addr <= 0x4017: // APU and I/O registers.
		return
	case addr <= 0x401F: // APU and I/O functionality.
		return
	default: // Cartridge space.
		b.Cart.WritePRG(addr, data)
	}
}

func (b *Bus) Reset() {
	b.Cart.Reset()
	b.CPU.Reset(b)
	b.PPU.Reset()
	b.cycles = 0
}

func (b *Bus) Tick() (r TickInfo) {
	b.cycles++
	b.PPU.Tick()

	if b.cycles%3 == 0 {
		// CPU runs 3x slower than PPU.
		r.InstrComplete = b.CPU.Tick(b)
	}

	if b.PPU.RequestNMI {
		b.PPU.RequestNMI = false
		b.CPU.TriggerNMI()
	}

	if b.PPU.ScanlineComplete {
		b.PPU.ScanlineComplete = false
		r.ScanlineComplete = true

		// Some mappers require scanline IRQs.
		if t := b.Cart.Scanline(); t.IRQ {
			b.CPU.TriggerIRQ()
		}
	}

	if b.PPU.FrameComplete {
		b.PPU.FrameComplete = false
		r.FrameComplete = true
	}

	return r
}
