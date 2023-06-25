package ppu

import (
	"fmt"
	"image/color"
	"log"

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
	Frame            [256][240]color.RGBA
	RequestNMI       bool
	ScanlineComplete bool
	FrameComplete    bool
	NoSpriteLimit    bool

	cart         ines.Cartridge // $0000-$1FFF (CHR-ROM)
	ctrl         CtrlFlags      // $2000
	mask         MaskFlags      // $2001
	status       StatusFlags    // $2002
	oamAddr      uint8          // $2003
	oamData      [256]byte      // $2004
	scrollX      uint8          // $2005 (first write)
	scrollY      uint8          // $2005 (second write)
	nameTable    [2][1024]byte  // $2000-$2FFF
	paletteTable [32]byte       // $3F00-$3FFF

	vramAddr   uint16
	tmpAddr    uint16
	vramBuffer uint8
	addrLatch  bool

	spriteCount    int
	spriteScanline [64]Sprite

	cycle      int
	scanline   int
	skipRender bool
}

func New(cart ines.Cartridge) *PPU {
	return &PPU{
		cart: cart,
	}
}

// DisableRender disables the actual rendering of the frame but keeps the
// other operations intact. This is useful for fast-forwarding the game
// state where we don’t need to display the frames.
func (p *PPU) DisableRender() {
	p.skipRender = true
}

// EnableRender enables the rendering of the frame.
func (p *PPU) EnableRender() {
	p.skipRender = false
}

func (p *PPU) getStatus(flag StatusFlags) bool {
	return p.status&flag != 0
}

func (p *PPU) setStatus(flag StatusFlags, value bool) {
	if value {
		p.status |= flag
	} else {
		p.status &= ^flag
	}
}

func (p *PPU) getCtrl(flag CtrlFlags) bool {
	return p.ctrl&flag != 0
}

func (p *PPU) getMask(flag MaskFlags) bool {
	return p.mask&flag != 0
}

func (p *PPU) incrementAddr() {
	if p.getCtrl(CtrlIncrementMode) {
		p.vramAddr += 32
	} else {
		p.vramAddr += 1
	}
}

func (p *PPU) Reset() {
	p.ctrl = 0
	p.mask = 0
	p.status = 0
	p.oamAddr = 0
	p.vramAddr = 0
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
		status := p.status
		p.addrLatch = false
		p.setStatus(StatusVBlank, false)
		return uint8(status)&0xE0 | p.vramBuffer&0x1F

	case oamDataRegAddr:
		data := p.oamData[p.oamAddr]
		if p.oamAddr&0x03 == 0x02 {
			data &= 0xE3
		}
		return data

	case vramDataRegAddr:
		// Palette reads are not delayed.
		if p.vramAddr >= 0x3F00 {
			data := p.readVRAM(p.vramAddr)
			p.incrementAddr()
			return data
		}

		// Reads from pattern tables are delayed by one cycle.
		data := p.vramBuffer
		p.vramBuffer = p.readVRAM(p.vramAddr)
		p.incrementAddr()
		return data

	default:
		return 0
	}
}

func (p *PPU) Write(addr uint16, data uint8) {
	switch addr % 0x2008 {
	case ctrlRegAddr:
		p.ctrl = CtrlFlags(data)
	case maskRegAddr:
		p.mask = MaskFlags(data)
	case oamAddrRegAddr:
		p.oamAddr = data
	case oamDataRegAddr:
		p.oamData[p.oamAddr] = data
		p.oamAddr++
	case vramAddrRegAddr:
		if !p.addrLatch {
			p.tmpAddr = uint16(data)
			p.addrLatch = true
		} else {
			p.vramAddr = p.tmpAddr<<8 | uint16(data)
			p.addrLatch = false
			p.tmpAddr = 0
		}
	case scrollRegAddr:
		if !p.addrLatch {
			p.scrollX = data
			p.addrLatch = true
		} else {
			p.scrollY = data
			p.addrLatch = false
		}
	case vramDataRegAddr:
		p.writeVRAM(p.vramAddr, data)
		p.incrementAddr()
	}
}

func (p *PPU) WriteOAM(data byte) {
	p.oamData[p.oamAddr] = data
	p.oamAddr++
}

func (p *PPU) nameTableIdx(addr uint16) int {
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
	case ines.MirrorSingle0:
		return 0
	case ines.MirrorSingle1:
		return 1
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
		idx := p.nameTableIdx(addr)
		return p.nameTable[idx][addr%1024]
	}

	if addr <= 0x3FFF {
		if addr == 0x3F10 || addr == 0x3F14 || addr == 0x3F18 || addr == 0x3F1C {
			addr -= 0x10 // mirrors $3F00/$3F04/$3F08/$3F0C
		}

		idx := (addr - 0x3F00) % 32
		value := p.paletteTable[idx]

		if p.getMask(MaskGrayscale) {
			value &= 0x30
		}

		return value
	}

	log.Printf("[WARN] read from invalid vram address: %04X", addr)
	return 0
}

func (p *PPU) writeVRAM(addr uint16, data uint8) {
	if addr <= 0x1FFF {
		p.cart.WriteCHR(addr, data)
		return
	}

	if addr <= 0x3EFF {
		addr = addr & 0x2FFF
		idx := p.nameTableIdx(addr)
		p.nameTable[idx][addr%1024] = data

		return
	}

	if addr <= 0x3FFF {
		if addr == 0x3F10 || addr == 0x3F14 || addr == 0x3F18 || addr == 0x3F1C {
			addr -= 0x10 // mirrors $3F00/$3F04/$3F08/$3F0C
		}

		idx := (addr - 0x3F00) % 32
		p.paletteTable[idx] = data

		return
	}

	log.Printf("[WARN] write to invalid vram address: %04X", addr)
}

func (p *PPU) checkSpriteZeroHit() bool {
	if !p.getMask(MaskShowSprites) || !p.getMask(MaskShowBackground) || p.getStatus(StatusSpriteZeroHit) {
		return false
	}

	spriteX, spriteY := int(p.oamData[3]), int(p.oamData[0])
	frameY, frameX := p.scanline, p.cycle
	spriteY += 2 // Not sure why.

	// Check if the scanline is within the sprite's horizontal range.
	if frameX < spriteX || frameX >= spriteX+8 {
		return false
	}

	// Check if the scanline is within the sprite's vertical range.
	if frameY < spriteY || frameY >= spriteY+p.spriteHeight() {
		return false
	}

	spritePixel := p.fetchSprite(0).Pixels[frameX-spriteX][frameY-spriteY]
	//tilePixel := p.fetchTile(frameX/8, frameY/8).Pixels[frameX%8][frameY%8]

	return spritePixel != 0 // && tilePixel != 0
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

func (p *PPU) render() {
	if p.getMask(MaskShowBackground) {
		p.renderTileScanline()
	}

	if p.getMask(MaskShowSprites) {
		p.renderSpriteScanline()
	}
}

func (p *PPU) Tick() {
	if p.scanline == -1 {
		// Start of pre-render scanline, clear the frame with the backdrop color and
		// reset the PPU status flags.
		if p.cycle == 1 {
			p.clearFrame(p.backdropColor())
			p.setStatus(StatusSpriteOverflow, false)
			p.setStatus(StatusSpriteZeroHit, false)
			p.setStatus(StatusVBlank, false)
		}

		// End of pre-render scanline, prepare the sprites for the first visible scanline.
		if p.cycle == 340 {
			if !p.skipRender {
				p.prepareSprites()
			}
		}
	}

	if p.scanline >= 0 && p.scanline <= 239 {
		if p.cycle >= 1 && p.cycle <= 256 {
			if p.checkSpriteZeroHit() {
				p.setStatus(StatusSpriteZeroHit, true)
			}
		}

		// End of visible part of a scanline, render of what is on it.
		if p.cycle == 257 {
			p.ScanlineComplete = true

			if !p.skipRender {
				p.render()
			}
		}

		// Prepare the sprites for the next scanline.
		if p.cycle == 340 {
			if !p.skipRender {
				p.prepareSprites()
			}
		}
	}

	// End of visible scanlines and start of vertical blank. Set the vblank flag and
	// trigger the CPU interrupt if the NMI flag is set.
	if p.scanline == 241 && p.cycle == 1 {
		p.setStatus(StatusVBlank, true)

		if p.getCtrl(CtrlNMI) {
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
