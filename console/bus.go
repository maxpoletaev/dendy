package console

import (
	"errors"
	"fmt"
	"io"

	apupkg "github.com/maxpoletaev/dendy/apu"
	cpupkg "github.com/maxpoletaev/dendy/cpu"
	ppupkg "github.com/maxpoletaev/dendy/ppu"

	"github.com/maxpoletaev/dendy/disasm"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
)

type Bus struct {
	RAM    [2048]uint8
	CPU    *cpupkg.CPU
	PPU    *ppupkg.PPU
	APU    *apupkg.APU
	Cart   ines.Cartridge
	Joy1   *input.Joystick
	Joy2   *input.Joystick
	Zapper *input.Zapper

	scanlineComplete bool
	frameComplete    bool

	DisasmWriter  io.StringWriter
	DisasmEnabled bool

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
	case addr >= 0x0000 && addr <= 0x1FFF: // Internal RAM.
		return b.RAM[addr%0x0800]
	case addr >= 0x2000 && addr <= 0x3FFF: // PPU registers.
		return b.PPU.Read(addr)
	case addr >= 0x4000 && addr <= 0x4014: // Open bus.
		return 0
	case addr == 0x4015: // APU status.
		return b.APU.Read(addr)
	case addr == 0x4016: // Controller 1.
		return b.Joy1.Read()
	case addr == 0x4017: // Controller 2.
		if b.Zapper != nil {
			return b.Zapper.Read()
		}
		return b.Joy2.Read()
	case addr >= 0x4018 && addr <= 0x401F: // Unused APU/IO registers.
		return 0
	default: // Cartridge space.
		return b.Cart.ReadPRG(addr)
	}
}

func (b *Bus) writeStrobe(data uint8) {
	if b.Joy1 != nil {
		b.Joy1.Write(data)
	}

	if b.Joy2 != nil {
		b.Joy2.Write(data)
	}
}

func (b *Bus) Write(addr uint16, data uint8) {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF: // Internal RAM.
		b.RAM[addr%0x0800] = data
	case addr >= 0x2000 && addr <= 0x3FFF: // PPU registers.
		b.PPU.Write(addr, data)
	case addr >= 0x4000 && addr <= 0x4013: // APU registers.
		b.APU.Write(addr, data)
	case addr == 0x4014: // PPU OAM DMA.
		b.transferOAM(data)
	case addr == 0x4015: // APU status.
		b.APU.Write(addr, data)
	case addr == 0x4016: // Controller strobe.
		b.writeStrobe(data)
	case addr <= 0x4017: // APU frame counter.
		b.APU.Write(addr, data)
	case addr >= 0x4018 && addr <= 0x401F: // Unused APU/IO registers.
		return
	default: // Cartridge space.
		b.Cart.WritePRG(addr, data)
	}
}

func (b *Bus) Reset() {
	b.Cart.Reset()
	b.CPU.Reset(b)
	b.PPU.Reset()
	b.APU.Reset()

	if b.Joy1 != nil {
		b.Joy1.Reset()
	}

	if b.Joy2 != nil {
		b.Joy2.Reset()
	}

	if b.Zapper != nil {
		b.Zapper.Reset()
	}

	b.cycles = 0
	b.frameComplete = false
	b.scanlineComplete = false
}

func (b *Bus) disassemble() {
	if b.DisasmWriter == nil {
		return
	}

	_, err1 := b.DisasmWriter.WriteString(disasm.DebugStep(b, b.CPU))
	_, err2 := b.DisasmWriter.WriteString("\n")

	if err := errors.Join(err1, err2); err != nil {
		panic(fmt.Sprintf("error writing disassembly: %v", err))
	}
}

func (b *Bus) Tick() {
	b.cycles++

	if b.cycles%3 == 0 {
		instructionComplete := b.CPU.Tick(b)
		if b.DisasmEnabled && instructionComplete {
			b.disassemble()
		}

		b.APU.Tick()
		if b.APU.PendingIRQ() {
			b.CPU.TriggerIRQ()
		}
	}

	b.PPU.Tick()
	if b.PPU.PendingNMI() {
		b.CPU.TriggerNMI()
	}

	if b.PPU.ScanlineComplete() {
		b.scanlineComplete = true

		b.Cart.ScanlineTick()
		if b.Cart.PendingIRQ() {
			b.CPU.TriggerIRQ()
		}
	}

	if b.PPU.FrameComplete() {
		b.frameComplete = true
	}
}

func (b *Bus) ScanlineComplete() (v bool) {
	v, b.scanlineComplete = b.scanlineComplete, false
	return v
}

func (b *Bus) FrameComplete() (v bool) {
	v, b.frameComplete = b.frameComplete, false
	return v
}
