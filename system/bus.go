package system

import (
	apupkg "github.com/maxpoletaev/dendy/apu"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	ppupkg "github.com/maxpoletaev/dendy/ppu"
)

// Bus represents the main CPU memory bus. It is responsible for routing memory
// read and write operations to the appropriate devices.
type Bus struct {
	ram   []byte // 2KB
	ppu   *ppupkg.PPU
	apu   *apupkg.APU
	cart  ines.Cartridge
	port1 input.Device
	port2 input.Device
}

func newBus(
	ram []uint8,
	ppu *ppupkg.PPU,
	apu *apupkg.APU,
	cart ines.Cartridge,
	port1, port2 input.Device,
) *Bus {
	return &Bus{
		ram:   ram,
		ppu:   ppu,
		apu:   apu,
		cart:  cart,
		port1: port1,
		port2: port2,
	}
}

func (b *Bus) Read(addr uint16) uint8 {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF: // Internal RAM.
		return b.ram[addr%0x0800]
	case addr >= 0x2000 && addr <= 0x3FFF: // PPU registers.
		return b.ppu.Read(addr)
	case addr >= 0x4000 && addr <= 0x4014: // Open bus.
		return 0
	case addr == 0x4015: // APU status.
		return b.apu.Read(addr)
	case addr == 0x4016: // Controller 1.
		return b.port1.Read()
	case addr == 0x4017: // Controller 2.
		return b.port2.Read()
	case addr >= 0x4018 && addr <= 0x401F: // Unused APU/IO registers.
		return 0
	default: // 0x8000-0xFFFF: Cartridge space.
		return b.cart.ReadPRG(addr)
	}
}

func (b *Bus) Write(addr uint16, data uint8) {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF: // Internal RAM.
		b.ram[addr%0x0800] = data
	case addr >= 0x2000 && addr <= 0x3FFF: // PPU registers.
		b.ppu.Write(addr, data)
	case addr >= 0x4000 && addr <= 0x4013: // APU registers.
		b.apu.Write(addr, data)
	case addr == 0x4014: // PPU OAM DMA.
		b.ppu.TransferOAM(data)
	case addr == 0x4015: // APU status.
		b.apu.Write(addr, data)
	case addr == 0x4016: // Controller strobe.
		b.port1.Write(data)
		b.port2.Write(data)
	case addr <= 0x4017: // APU frame counter.
		b.apu.Write(addr, data)
	case addr >= 0x4018 && addr <= 0x401F: // Unused APU/IO registers.
		return
	default: // 0x8000-0xFFFF: Cartridge space.
		b.cart.WritePRG(addr, data)
	}
}
