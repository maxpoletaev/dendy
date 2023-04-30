package ppu

import (
	"fmt"
	"image/color"

	"github.com/maxpoletaev/dendy/ines"
)

const (
	ctrlRegAddr     uint16 = 0x2000
	maskRegAddr     uint16 = 0x2001
	statusRegAddr   uint16 = 0x2002
	oamAddrRegAddr  uint16 = 0x2003
	oamDataRegAddr  uint16 = 0x2004
	scrollRegAddr   uint16 = 0x2005
	vramAddrRegAddr uint16 = 0x2006
	vramDataRegAddr uint16 = 0x2007
)

type (
	CtrlFlags   uint8
	MaskFlags   uint8
	StatusFlags uint8
)

const (
	CtrlNameTableSelect0   CtrlFlags = 1 << 0
	CtrlNameTableSelect1   CtrlFlags = 1 << 1
	CtrlIncrementMode      CtrlFlags = 1 << 2
	CtrlSpritePatternAddr  CtrlFlags = 1 << 3
	CtrlPatternTableSelect CtrlFlags = 1 << 4
	CtrlSpriteSize         CtrlFlags = 1 << 5
	CtrlSlaveMode          CtrlFlags = 1 << 6
	CtrlNMI                CtrlFlags = 1 << 7
)

const (
	MaskGrayscale       MaskFlags = 1 << 0
	MaskShowLeftBg      MaskFlags = 1 << 1
	MaskShowLeftSprites MaskFlags = 1 << 2
	MaskShowBg          MaskFlags = 1 << 3
	MaskShowSprites     MaskFlags = 1 << 4
	MaskEmphasizeRed    MaskFlags = 1 << 5
	MaskEmphasizeGreen  MaskFlags = 1 << 6
	MaskEmphasizeBlue   MaskFlags = 1 << 7
)

const (
	StatusSpriteOverflow StatusFlags = 1 << 5
	StatusSprite0Hit     StatusFlags = 1 << 6
	StatusVBlank         StatusFlags = 1 << 7
)

type PPU struct {
	cart         *ines.Cartridge // $0000-$1FFF (CHR-ROM)
	Ctrl         CtrlFlags       // $2000
	Mask         MaskFlags       // $2001
	Status       StatusFlags     // $2002
	OAMAddr      uint8           // $2003
	OAMData      [256]byte       // $2004
	NameTable    [2][1024]byte   // $2000-$2FFF
	PaletteTable [32]byte        // $3F00-$3FFF
	RequestNMI   bool
	VRAMAddr     uint16
	VBlank       bool

	// Current frame and if it is ready to be rendered.
	Frame         [256][240]color.RGBA
	FrameComplete bool

	cycle        int
	scanline     int
	addressLatch bool
	tmpVRAMAddr  uint16
	vramBuffer   uint8
}

func New(cart *ines.Cartridge) *PPU {
	return &PPU{
		cart: cart,
	}
}

func (p *PPU) getCtrlFlag(flag CtrlFlags) bool {
	return p.Ctrl&flag != 0
}

func (p *PPU) setCtrlFlag(flag CtrlFlags, value bool) {
	if value {
		p.Ctrl |= flag
		return
	}

	p.Ctrl &= ^flag
}

func (p *PPU) getMaskFlag(flag MaskFlags) bool {
	return p.Mask&flag != 0
}

func (p *PPU) setMaskFlag(flag MaskFlags, value bool) {
	if value {
		p.Mask |= flag
		return
	}

	p.Mask &= ^flag
}

func (p *PPU) getStatusFlag(flag StatusFlags) bool {
	return p.Status&flag != 0
}

func (p *PPU) setStatusFlag(flag StatusFlags, value bool) {
	if value {
		p.Status |= flag
		return
	}

	p.Status &= ^flag
}

func (p *PPU) incrementVRAMAddr() {
	if p.getCtrlFlag(CtrlIncrementMode) {
		p.VRAMAddr += 32
	} else {
		p.VRAMAddr += 1
	}
}

func (p *PPU) Reset() {
	p.Ctrl = 0
	p.Mask = 0
	p.Status = 0
	p.OAMAddr = 0
	p.VRAMAddr = 0
	p.VBlank = false
	p.RequestNMI = false

	p.FrameComplete = false
	p.addressLatch = false
	p.tmpVRAMAddr = 0
	p.vramBuffer = 0
	p.scanline = 0
	p.cycle = 0
}

func (p *PPU) Read(addr uint16) uint8 {
	switch addr % 0x2008 {
	case statusRegAddr:
		p.VBlank = false
		p.addressLatch = false
		return uint8(p.Status)
	case oamDataRegAddr:
		data := p.OAMData[p.OAMAddr]
		if p.OAMAddr&0x03 == 0x02 {
			data &= 0xE3
		}
		return data
	case vramDataRegAddr:
		data := p.vramBuffer
		if p.VRAMAddr >= 0x3F00 {
			data = p.readVRAM(p.VRAMAddr)
		}
		p.vramBuffer = p.readVRAM(p.VRAMAddr)
		p.incrementVRAMAddr()
		return data
	default:
		return 0
	}
}

func (p *PPU) Write(addr uint16, data uint8) {
	switch addr % 0x2008 {
	case ctrlRegAddr:
		p.Ctrl = CtrlFlags(data)
	case maskRegAddr:
		p.Mask = MaskFlags(data)
	case oamAddrRegAddr:
		p.OAMAddr = data
	case oamDataRegAddr:
		p.OAMData[p.OAMAddr] = data
		p.OAMAddr++
	case vramAddrRegAddr:
		if !p.addressLatch {
			p.tmpVRAMAddr = uint16(data&0x3F) << 8
			p.addressLatch = true
			break
		}

		p.tmpVRAMAddr |= uint16(data)
		p.VRAMAddr = p.tmpVRAMAddr
		p.addressLatch = false
	case vramDataRegAddr:
		p.writeVRAM(p.VRAMAddr, data)
		p.incrementVRAMAddr()
	}
}

func (p *PPU) resolveNametableIdx(addr uint16) int {
	if p.cart.Mirror == ines.Horizontal {
		switch {
		case addr >= 0x2000 && addr <= 0x23FF: // Nametable 0
			return 0
		case addr >= 0x2400 && addr <= 0x27FF: // Nametable 1
			return 0
		case addr >= 0x2800 && addr <= 0x2BFF: // Nametable 2
			return 1
		case addr >= 0x2C00 && addr <= 0x2FFF: // Nametable 3
			return 1
		}
	}

	if p.cart.Mirror == ines.Vertical {
		switch {
		case addr >= 0x2000 && addr <= 0x23FF: // Nametable 0
			return 0
		case addr >= 0x2400 && addr <= 0x27FF: // Nametable 1
			return 1
		case addr >= 0x2800 && addr <= 0x2BFF: // Nametable 2
			return 0
		case addr >= 0x2C00 && addr <= 0x2FFF: // Nametable 3
			return 1
		}
	}

	// Should never happen, only in case there is a bug in the code above.
	panic(fmt.Sprintf("mirroring error: addr=%04X, mode=%d", addr, p.cart.Mirror))
}

func (p *PPU) applyGrayscaleIfSet(v uint8) uint8 {
	if p.Mask&MaskGrayscale != 0 {
		return v & 0x30
	}

	return v
}

func (p *PPU) readVRAM(addr uint16) uint8 {
	if addr <= 0x1FFF {
		return p.cart.ReadCHR(addr)
	}

	if addr <= 0x3EFF {
		addr = addr % 0x2FFF
		idx := p.resolveNametableIdx(addr)
		return p.NameTable[idx][addr%1024]
	}

	if addr <= 0x3FFF {
		addr = addr % 0x3F1F
		value := p.PaletteTable[addr%32]
		return p.applyGrayscaleIfSet(value)
	}

	panic(fmt.Sprintf("invalid vram address: %04X", addr))
}

func (p *PPU) writeVRAM(addr uint16, data uint8) {
	if addr <= 0x1FFF {
		p.cart.WriteCHR(addr, data)
		return
	}

	if addr <= 0x3EFF {
		addr = addr % 0x2FFF
		idx := p.resolveNametableIdx(addr)
		p.NameTable[idx][addr%1024] = data
		return
	}

	if addr <= 0x3FFF {
		addr = addr % 0x3F1F
		p.PaletteTable[addr%32] = data
		return
	}

	panic(fmt.Sprintf("invalid vram address: %04X", addr))
}

func (p *PPU) Tick() {
	// End of vblank, start of visible scanline.
	if p.scanline == -1 && p.cycle == 1 {
		p.setStatusFlag(StatusVBlank, false)
	}

	// End of visible scanline, start of vblank.
	if p.scanline == 241 && p.cycle == 1 {
		p.setStatusFlag(StatusVBlank, true)

		// Trigger the CPU interrupt.
		if p.getCtrlFlag(CtrlNMI) {
			p.RequestNMI = true
		}
	}

	if p.cycle++; p.cycle >= 341 {
		p.cycle = 0
		p.scanline++

		if p.scanline >= 261 {
			p.scanline = -1
			p.FrameComplete = true
			p.renderBackground()
			p.printNamenable()
		}
	}
}
