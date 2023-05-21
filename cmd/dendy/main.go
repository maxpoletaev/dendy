package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/maxpoletaev/dendy/cpu"
	"github.com/maxpoletaev/dendy/display"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/input"
	"github.com/maxpoletaev/dendy/ppu"
)

type Bus struct {
	screen *display.Display
	cart   ines.Cartridge
	joy1   *input.Joystick
	ram    [2048]uint8
	cpu    *cpu.CPU
	ppu    *ppu.PPU
	cycles uint64
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
	case addr <= 0x4017: // APU and I/O registers.
		return 0
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

func (b *Bus) transferOAM(addr uint8) {
	base := uint16(addr) << 8

	for i := uint16(0); i < 256; i++ {
		b.ppu.OAMData[b.ppu.OAMAddr] = b.Read(base + i)
		b.ppu.OAMAddr++
	}

	b.cpu.Halt += 513
	if b.cpu.Halt%2 == 1 {
		b.cpu.Halt++
	}
}

func (b *Bus) Reset() {
	b.cart.Reset()
	b.cpu.Reset(b)
	b.ppu.Reset()
	b.cycles = 0
}

func (b *Bus) Tick() {
	b.cycles++
	b.ppu.Tick()

	if b.cycles%3 == 0 {
		b.cpu.Tick(b)
	}

	if b.ppu.RequestNMI {
		b.ppu.RequestNMI = false
		b.cpu.TriggerNMI()
	}

	if b.ppu.FrameComplete {
		b.ppu.FrameComplete = false
		b.screen.HandleInput()
		b.screen.Refresh()
	}
}

func main() {
	var (
		stepMode bool
		disasm   bool
	)

	flag.BoolVar(&disasm, "disasm", false, "enable cpu disassembler")
	flag.BoolVar(&stepMode, "step", false, "step mode")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Usage: dendy <rom>")
		os.Exit(1)
	}

	cart, err := ines.Load(flag.Arg(0))
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to open rom: %s", err))
		os.Exit(1)
	}

	var (
		c = cpu.New()
		p = ppu.New(cart)
		j = input.NewJoystick()
		d = display.New(&p.Frame, j, 2)
	)

	c.AllowIllegal = true

	if disasm {
		c.EnableDisasm = true
	}

	bus := &Bus{
		cart:   cart,
		screen: d,
		cpu:    c,
		ppu:    p,
		joy1:   j,
	}

	bus.Reset()
	d.Refresh()

	for d.IsRunning() {
		bus.Tick()
	}
}
