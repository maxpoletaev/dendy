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
	MaskGrayscale          MaskFlags = 1 << 0
	MaskShowLeftBackground MaskFlags = 1 << 1
	MaskShowLeftSprites    MaskFlags = 1 << 2
	MaskShowBackground     MaskFlags = 1 << 3
	MaskShowSprites        MaskFlags = 1 << 4
	MaskEmphasizeRed       MaskFlags = 1 << 5
	MaskEmphasizeGreen     MaskFlags = 1 << 6
	MaskEmphasizeBlue      MaskFlags = 1 << 7
)

const (
	StatusSpriteOverflow StatusFlags = 1 << 5
	StatusSprite0Hit     StatusFlags = 1 << 6
	StatusVBlank         StatusFlags = 1 << 7
)

type PPU struct {
	cart         ines.Cartridge // $0000-$1FFF (CHR-ROM)
	Ctrl         CtrlFlags      // $2000
	Mask         MaskFlags      // $2001
	Status       StatusFlags    // $2002
	OAMAddr      uint8          // $2003
	OAMData      [256]byte      // $2004
	NameTable    [2][1024]byte  // $2000-$2FFF
	PaletteTable [32]byte       // $3F00-$3FFF
	VRAMAddr     uint16

	Frame         [256][240]color.RGBA
	FrameComplete bool
	RequestNMI    bool

	spriteCount    int
	spriteScanline [8]Sprite

	cycle        int
	scanline     int
	addressLatch bool
	tmpVRAMAddr  uint16
	vramBuffer   uint8
}

func New(cart ines.Cartridge) *PPU {
	return &PPU{
		cart: cart,
	}
}

func (p *PPU) getFlag(flag any) bool {
	switch f := flag.(type) {
	case CtrlFlags:
		return p.Ctrl&f != 0
	case MaskFlags:
		return p.Mask&f != 0
	case StatusFlags:
		return p.Status&f != 0
	default:
		panic(fmt.Sprintf("unknown flag type %T", f))
	}
}

func (p *PPU) setFlag(flag any, value bool) {
	switch f := flag.(type) {
	case CtrlFlags:
		if value {
			p.Ctrl |= f
			return
		}

		p.Ctrl &= ^f
	case MaskFlags:
		if value {
			p.Mask |= f
			return
		}

		p.Mask &= ^f
	case StatusFlags:
		if value {
			p.Status |= f
			return
		}

		p.Status &= ^f
	default:
		panic(fmt.Sprintf("unknown flag type %T", f))
	}
}

func (p *PPU) incrementVRAMAddr() {
	if p.getFlag(CtrlIncrementMode) {
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
		// We only use the top 3 bits of the status register, and the rest are filled
		// with noise from the bottom 5 bits of the vram buffer. It also clears the
		// address latch and vblank flag.
		status := p.Status
		p.addressLatch = false
		p.setFlag(StatusVBlank, false)
		return uint8(status)&0xE0 | p.vramBuffer&0x1F

	case oamDataRegAddr:
		data := p.OAMData[p.OAMAddr]
		if p.OAMAddr&0x03 == 0x02 {
			data &= 0xE3
		}
		return data

	case vramDataRegAddr:
		// Palette reads are not delayed.
		if p.VRAMAddr >= 0x3F00 {
			data := p.readVRAM(p.VRAMAddr)
			p.vramBuffer = p.readVRAM(p.VRAMAddr - 0x1000)
			p.incrementVRAMAddr()
			return data
		}

		// Reads from pattern tables are delayed by one cycle.
		data := p.vramBuffer
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
			p.tmpVRAMAddr = uint16(data)
			p.addressLatch = true
			break
		}

		p.VRAMAddr = p.tmpVRAMAddr<<8 | uint16(data)
		p.addressLatch = false
		p.tmpVRAMAddr = 0
	case vramDataRegAddr:
		p.writeVRAM(p.VRAMAddr, data)
		p.incrementVRAMAddr()
	}
}

func (p *PPU) NameTableIdx(addr uint16) int {
	switch p.cart.MirrorMode() {
	case ines.MirrorHorizontal:
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
	case ines.MirrorVertical:
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
	default:
		mode := p.cart.MirrorMode()
		panic(fmt.Sprintf("invalid mirroring mode: %d", mode))
	}

	panic(fmt.Sprintf("invalid nametable address: %04X", addr))
}

func (p *PPU) applyGrayscaleIfSet(v uint8) uint8 {
	if p.getFlag(MaskGrayscale) {
		return v & 0x30
	}

	return v
}

func (p *PPU) readVRAM(addr uint16) uint8 {
	if addr <= 0x1FFF {
		return p.cart.ReadCHR(addr)
	}

	if addr <= 0x3EFF {
		addr = addr & 0x2FFF
		idx := p.NameTableIdx(addr)
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
		addr = addr & 0x2FFF
		idx := p.NameTableIdx(addr)
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
	if p.scanline == -1 {
		// Start of pre-render scanline, clear the frame and reset the status flags.
		if p.cycle == 1 {
			p.clearFrame(color.RGBA{0, 0, 0, 0xFF})
			p.setFlag(StatusSpriteOverflow, false)
			p.setFlag(StatusSprite0Hit, false)
			p.setFlag(StatusVBlank, false)
		}

		// End of pre-render scanline, prepare the sprites for the first visible scanline.
		if p.cycle == 340 {
			p.prepareSprites()
		}
	}

	// Visible scanlines. At the end of each scanline, render the tiles and sprites,
	// and prepare the sprites for the next scanline. This is not how the real PPU
	// works, but it's good enough for now.
	if p.scanline >= 0 && p.scanline <= 239 {
		if p.cycle == 340 {
			p.renderTileScanline()
			p.renderSpriteScanline()
			p.prepareSprites()
		}
	}

	// End of visible scanlines and start of vertical blank. Set the vblank flag and
	// trigger the CPU interrupt if the NMI flag is set.
	if p.scanline == 241 && p.cycle == 1 {
		p.setFlag(StatusVBlank, true)

		if p.getFlag(CtrlNMI) {
			p.RequestNMI = true
		}

		p.FrameComplete = true
	}

	// Endless loop of scanlines and cycles.
	if p.cycle++; p.cycle >= 341 {
		p.cycle = 0
		p.scanline++

		if p.scanline >= 261 {
			p.scanline = -1
		}
	}
}
