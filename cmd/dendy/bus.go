package main

import (
	cpu2 "github.com/maxpoletaev/dendy/cpu"
	"github.com/maxpoletaev/dendy/display"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	ppu2 "github.com/maxpoletaev/dendy/ppu"
)

type Bus struct {
	screen *display.Window
	cart   ines.Cartridge
	joy1   *input.Joystick
	zap    *input.Zapper
	ram    [2048]uint8
	cpu    *cpu2.CPU
	ppu    *ppu2.PPU
	cycles uint64
}

func (b *Bus) transferOAM(addr uint8) {
	var (
		oamAddr = uint16(b.ppu.OAMAddr)
		memAddr = uint16(addr) << 8
	)

	for i := uint16(0); i < 256; i++ {
		b.ppu.OAMData[oamAddr+i] = b.Read(memAddr + i)
	}

	b.cpu.Halt += 513
	if b.cpu.Halt%2 == 1 {
		b.cpu.Halt++
	}
}

func (b *Bus) Read(addr uint16) uint8 {
	switch {
	case addr <= 0x1FFF: // Internal RAM.
		addr = addr % 0x0800
		return b.ram[addr]
	case addr <= 0x3FFF: // PPU registers.
		return b.ppu.Read(addr)
	case addr == 0x4014: // PPU OAM DMA.
		return b.ppu.Read(addr)
	case addr == 0x4016: // Controller 1.
		return b.joy1.Read()
	case addr <= 0x4017: // Zapper.
		return b.zap.Read()
	case addr <= 0x401F: // APU and I/O functionality.
		return 0
	default: // Cartridge space.
		return b.cart.ReadPRG(addr)
	}
}

func (b *Bus) Write(addr uint16, data uint8) {
	switch {
	case addr <= 0x1FFF: // Internal RAM.
		addr = addr % 0x0800
		b.ram[addr] = data
	case addr <= 0x3FFF: // PPU registers.
		b.ppu.Write(addr, data)
	case addr == 0x4014: // PPU OAM direct access.
		b.transferOAM(data)
	case addr == 0x4016: // Controller strobe.
		b.joy1.Write(data)
	case addr <= 0x4017: // APU and I/O registers.
		return
	case addr <= 0x401F: // APU and I/O functionality.
		return
	default: // Cartridge space.
		b.cart.WritePRG(addr, data)
	}
}

func (b *Bus) Reset() {
	b.cart.Reset()
	b.cpu.Reset(b)
	b.ppu.Reset()
	b.cycles = 0
}

func (b *Bus) Tick() (instrComplete, frameComplete bool) {
	b.cycles++
	b.ppu.Tick()

	// CPU runs 3 times slower than PPU.
	if b.cycles%3 == 0 {
		instrComplete = b.cpu.Tick(b)
	}

	// Trigger the CPU NMI if the PPU has requested it.
	if b.ppu.RequestNMI {
		b.ppu.RequestNMI = false
		b.cpu.TriggerNMI()
	}

	// Refresh the screen if a frame has completed.
	if b.ppu.FrameComplete {
		frameComplete = true
		b.ppu.FrameComplete = false
		b.screen.HandleInput()
		b.screen.Refresh()
	}

	return instrComplete, frameComplete
}
