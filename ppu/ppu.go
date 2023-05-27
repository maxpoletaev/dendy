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
	CtrlNameTableSelect              = 0x03 // two bits
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
	StatusSpriteZeroHit  StatusFlags = 1 << 6
	StatusVBlank         StatusFlags = 1 << 7
)

type PPU struct {
	cart         ines.Cartridge // $0000-$1FFF (CHR-ROM)
	Ctrl         CtrlFlags      // $2000
	Mask         MaskFlags      // $2001
	Status       StatusFlags    // $2002
	OAMAddr      uint8          // $2003
	OAMData      [256]byte      // $2004
	ScrollX      uint8          // $2005 (first write)
	ScrollY      uint8          // $2005 (second write)
	NameTable    [2][1024]byte  // $2000-$2FFF
	PaletteTable [32]byte       // $3F00-$3FFF

	Frame         [256][240]color.RGBA
	RequestNMI    bool
	FrameComplete bool

	VRAMAddr  uint16
	tmpAddr   uint16
	addrLatch bool

	spriteCount    int
	spriteScanline [8]Sprite

	cycle      int
	scanline   int
	vramBuffer uint8
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
	p.addrLatch = false
	p.tmpAddr = 0
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
		p.addrLatch = false
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
		if !p.addrLatch {
			p.tmpAddr = uint16(data)
			p.addrLatch = true
		} else {
			p.VRAMAddr = p.tmpAddr<<8 | uint16(data)
			p.addrLatch = false
			p.tmpAddr = 0
		}
	case scrollRegAddr:
		if !p.addrLatch {
			p.ScrollX = data
			p.addrLatch = true
		} else {
			p.ScrollY = data
			p.addrLatch = false
		}
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
		if addr == 0x3F10 || addr == 0x3F14 || addr == 0x3F18 || addr == 0x3F1C {
			addr -= 0x10 // mirrors $3F00/$3F04/$3F08/$3F0C
		}

		idx := (addr - 0x3F00) % 32
		value := p.PaletteTable[idx]

		if p.getFlag(MaskGrayscale) {
			value &= 0x30
		}

		return value
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
		if addr == 0x3F10 || addr == 0x3F14 || addr == 0x3F18 || addr == 0x3F1C {
			addr -= 0x10 // mirrors $3F00/$3F04/$3F08/$3F0C
		}

		idx := (addr - 0x3F00) % 32
		p.PaletteTable[idx] = data

		return
	}

	panic(fmt.Sprintf("invalid vram address: %04X", addr))
}

func (p *PPU) spriteZeroHit() bool {
	if p.getFlag(MaskShowSprites) && p.getFlag(MaskShowBackground) {
		spriteY := int(p.OAMData[0])
		if p.scanline == spriteY+8 {
			return true
		}
	}

	return false
}

func (p *PPU) clearFrame(c color.RGBA) {
	for x := 0; x < 256; x++ {
		for y := 0; y < 240; y++ {
			p.Frame[x][y] = c
		}
	}
}

func (p *PPU) backdropColor() color.RGBA {
	idx := p.readVRAM(0x3F00)
	return Colors[idx]
}

func (p *PPU) Tick() {
	if p.scanline == -1 {
		// Start of pre-render scanline, clear the frame with the backdrop color and
		// reset the PPU status flags.
		if p.cycle == 1 {
			p.clearFrame(p.backdropColor())
			p.setFlag(StatusSpriteOverflow, false)
			p.setFlag(StatusSpriteZeroHit, false)
			p.setFlag(StatusVBlank, false)
		}

		// End of pre-render scanline, prepare the sprites for the first visible scanline.
		if p.cycle == 340 {
			p.prepareSprites()
		}
	}

	if p.scanline >= 0 && p.scanline <= 239 {
		if p.spriteZeroHit() {
			p.setFlag(StatusSpriteZeroHit, true)
		}

		// End of visible scanline, render the tiles and sprites, and prepare the sprites
		// for the next scanline. This is a simplification of the PPU's behaviour, but
		// should produce visually identical results for most games that don't rely on
		// mid-scanline changes to scroll values, etc.
		if p.cycle == 340 {
			if p.getFlag(MaskShowBackground) {
				p.renderTileScanline()
			}

			if p.getFlag(MaskShowSprites) {
				p.renderSpriteScanline()
			}

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
