package ppu

import (
	"fmt"
	"image/color"
	"log"

	"github.com/maxpoletaev/dendy/ines"
)

type (
	CtrlFlags   = uint8
	MaskFlags   = uint8
	StatusFlags = uint8
)

const (
	CtrlIncrementMode      CtrlFlags = 1 << 2
	CtrlSpritePatternAddr  CtrlFlags = 1 << 3
	CtrlPatternTableSelect CtrlFlags = 1 << 4
	CtrlSpriteSize         CtrlFlags = 1 << 5
	CtrlSlaveMode          CtrlFlags = 1 << 6
	CtrlNMI                CtrlFlags = 1 << 7
)

const (
	MaskGrayscale       MaskFlags = 1 << 0
	MaskShowLeftTiles   MaskFlags = 1 << 1
	MaskShowLeftSprites MaskFlags = 1 << 2
	MaskShowBackground  MaskFlags = 1 << 3
	MaskShowSprites     MaskFlags = 1 << 4
	MaskEmphasizeRed    MaskFlags = 1 << 5
	MaskEmphasizeGreen  MaskFlags = 1 << 6
	MaskEmphasizeBlue   MaskFlags = 1 << 7
)

const (
	StatusSpriteOverflow StatusFlags = 1 << 5
	StatusSpriteZeroHit  StatusFlags = 1 << 6
	StatusVBlank         StatusFlags = 1 << 7
)

type (
	dmaFunc func(addr uint16, data []byte, size int)
)

type PPU struct {
	Frame       [256][240]color.RGBA
	transparent [256][240]bool

	NoSpriteLimit bool
	FastForward   bool
	dmaCopy       dmaFunc

	cart         ines.Cartridge // $0000-$1FFF (CHR-ROM)
	ctrl         CtrlFlags      // $2000
	mask         MaskFlags      // $2001
	status       StatusFlags    // $2002
	oamAddr      uint8          // $2003
	oamData      [256]byte      // $2004
	nameTable    [2][1024]byte  // $2000-$2FFF
	paletteTable [32]byte       // $3F00-$3FFF

	vramAddr   vramAddr
	tmpAddr    vramAddr
	vramBuffer uint8
	addrLatch  bool
	fineX      uint8
	oddFrame   bool

	spriteCount    int
	spriteScanline [64]Sprite

	cycle            int
	scanline         int
	pendingNMI       bool
	scanlineComplete bool
	frameComplete    bool
}

func New(cart ines.Cartridge) *PPU {
	return &PPU{
		cart: cart,
	}
}

func (p *PPU) Reset() {
	p.ctrl = 0
	p.mask = 0
	p.status = 0
	p.oamAddr = 0

	p.vramAddr = 0
	p.tmpAddr = 0
	p.vramBuffer = 0
	p.addrLatch = false
	p.fineX = 0
	p.oddFrame = false

	p.spriteCount = 0

	p.cycle = 0
	p.scanline = 0
	p.scanlineComplete = false
	p.pendingNMI = false
	p.frameComplete = false
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

func (p *PPU) Read(addr uint16) uint8 {
	addr = 0x2000 + (addr % 8)

	switch addr {
	case 0x2002:
		// We only use the top 3 bits of the status register, and the rest are filled
		// with noise from the bottom 5 bits of the vram buffer. It also clears the
		// address latch and vblank flag.
		status := p.status
		p.addrLatch = false
		p.setStatus(StatusVBlank, false)
		return uint8(status)&0xE0 | p.vramBuffer&0x1F

	case 0x2004:
		data := p.oamData[p.oamAddr]
		if p.oamAddr&0x03 == 0x02 {
			data &= 0xE3
		}
		return data

	case 0x2007:
		// Palette reads are not delayed.
		if p.vramAddr >= 0x3F00 {
			data := p.readVRAM(uint16(p.vramAddr))
			p.incrementAddr()
			return data
		}

		// Reads from pattern tables are delayed by one cycle.
		data := p.vramBuffer
		p.vramBuffer = p.readVRAM(uint16(p.vramAddr))
		p.incrementAddr()
		return data

	default:
		return 0
	}
}

func (p *PPU) Write(addr uint16, data uint8) {
	addr = 0x2000 + (addr % 8)

	switch addr {
	case 0x2000:
		p.ctrl = CtrlFlags(data)
		p.tmpAddr.setNametable(uint16(data) & 0x03)
	case 0x2001:
		p.mask = MaskFlags(data)
	case 0x2003:
		p.oamAddr = data
	case 0x2004:
		p.oamData[p.oamAddr] = data
		p.oamAddr++
	case 0x2005:
		if !p.addrLatch {
			p.tmpAddr.setCoarseX(uint16(data) >> 3)
			p.fineX = data & 0x07
			p.addrLatch = true
		} else {
			p.tmpAddr.setCoarseY(uint16(data) >> 3)
			p.tmpAddr.setFineY(uint16(data) & 0x07)
			p.addrLatch = false
		}
	case 0x2006:
		if !p.addrLatch {
			p.tmpAddr = vramAddr(data)
			p.addrLatch = true
		} else {
			p.vramAddr = p.tmpAddr<<8 | vramAddr(data)
			p.addrLatch = false
			p.tmpAddr = 0
		}
	case 0x2007:
		writeAddr := uint16(p.vramAddr) % 0x4000
		p.writeVRAM(writeAddr, data)
		p.incrementAddr()
	}
}

func (p *PPU) SetDMACallback(callback dmaFunc) {
	p.dmaCopy = callback
}

func (p *PPU) TransferOAM(pageAddr uint8) {
	addr := uint16(pageAddr) << 8
	p.dmaCopy(addr, p.oamData[:], len(p.oamData))
}

// nameTableIdx returns the index of the nametable (0 or 1) for the given vram
// address, based on the cartridge’s mirroring mode.
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

// readVRAM returns the value at the given VRAM address. Depending on the
// address, it may read from the cartridge’s CHR-ROM, PPU nametables, or
// palette table.
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

// clearFrame fills the frame with the given color.
func (p *PPU) clearFrame(c color.RGBA) {
	if p.FastForward {
		return
	}

	// Incremental copy optimization.
	// See https://gist.github.com/taylorza/df2f89d5f9ab3ffd06865062a4cf015d

	p.Frame[0][0] = c

	for j := 1; j < 240; j *= 2 {
		copy(p.Frame[0][j:], p.Frame[0][:j])
	}

	for i := 1; i < 256; i *= 2 {
		copy(p.Frame[i:], p.Frame[:i])
	}
}

func (p *PPU) backdropColor() color.RGBA {
	idx := p.readVRAM(0x3F00)
	return Colors[idx]
}

func (p *PPU) renderScanline() {
	if p.FastForward {
		return
	}

	if p.getMask(MaskShowBackground) {
		p.renderTileScanline()
	}

	if p.getMask(MaskShowSprites) {
		p.renderSpriteScanline()
	}
}

func (p *PPU) renderingEnabled() bool {
	return p.getMask(MaskShowBackground) || p.getMask(MaskShowSprites)
}

func (p *PPU) checkSpriteZeroHit() bool {
	spriteX, spriteY := int(p.oamData[3]), int(p.oamData[0])
	frameY, frameX := p.scanline, p.cycle
	spriteY += 1

	// Check if the scanline is within the sprite's horizontal range.
	if frameX < spriteX || frameX >= spriteX+8 {
		return false
	}

	// Check if the scanline is within the sprite's vertical range.
	if frameY < spriteY || frameY >= spriteY+p.spriteHeight() {
		return false
	}

	pixelX, pixelY := frameX%8, frameY%8
	spritePixel := p.fetchSpritePixel(0, frameX-spriteX, frameY-spriteY)
	tilePixel := p.fetchTileLine(frameX/8, frameY/8, pixelY).Pixels[pixelX][pixelY]

	return spritePixel != 0 && tilePixel != 0
}

func (p *PPU) Tick() {
	// Pre-render + visible scanlines.
	if p.scanline >= -1 && p.scanline <= 238 {
		if p.scanline == -1 && p.cycle == 1 {
			p.setStatus(StatusSpriteOverflow, false)
			p.setStatus(StatusSpriteZeroHit, false)
			p.setStatus(StatusVBlank, false)
			p.clearFrame(p.backdropColor())
		}

		// Skip the first cycle of the first scanline on odd frames.
		if p.scanline == 0 {
			if p.cycle == 0 && p.oddFrame {
				p.cycle = 1
			}
		}

		// Manual sprite zero hit detection during fast-forward.
		if p.FastForward && p.cycle >= 0 && p.cycle <= 255 {
			if p.renderingEnabled() && !p.getStatus(StatusSpriteZeroHit) {
				p.setStatus(StatusSpriteZeroHit, p.checkSpriteZeroHit())
			}
		}

		// Increment scrollX every 8 cycles (tile width).
		if p.cycle%8 == 0 && p.cycle <= 256 {
			if p.renderingEnabled() {
				p.vramAddr.incrementX()
			}
		}

		// Increment scrollY at the end of each scanline.
		if p.cycle == 256 {
			if p.renderingEnabled() {
				p.vramAddr.incrementY()
			}
		}

		// At the end of each scanline, reset scrollX to the initial position from tmpAddr.
		// Then render it and prepare the sprites for the next scanline.
		if p.cycle == 257 {
			if p.renderingEnabled() {
				p.vramAddr.setNametableX(p.tmpAddr.nametableX())
				p.vramAddr.setCoarseX(p.tmpAddr.coarseX())
			}

			if p.scanline >= 0 {
				p.renderScanline()
			}

			if p.renderingEnabled() {
				p.evaluateSprites()
			}

			p.scanlineComplete = true
		}

		// During cycles 280-304 of the pre-render scanline, vertical scroll
		// bits are copied multiple times. But I guess it's fine to do it once?
		if p.scanline == -1 && p.cycle == 280 {
			if p.renderingEnabled() {
				p.vramAddr.setNametableY(p.tmpAddr.nametableY())
				p.vramAddr.setCoarseY(p.tmpAddr.coarseY())
				p.vramAddr.setFineY(p.tmpAddr.fineY())
			}
		}

		// X is incremented on cycles 328 and 336 as well.
		if p.cycle == 328 || p.cycle == 336 {
			if p.renderingEnabled() {
				p.vramAddr.incrementX()
			}
		}
	}

	// Start of vertical blank.
	if p.scanline == 241 {
		if p.cycle == 1 {
			p.setStatus(StatusVBlank, true)
			p.frameComplete = true

			if p.getCtrl(CtrlNMI) {
				p.pendingNMI = true
			}
		}
	}

	// Endless loop of scanlines and cycles.
	if p.cycle++; p.cycle == 341 {
		p.cycle = 0
		p.scanline++

		if p.scanline == 261 {
			p.oddFrame = !p.oddFrame
			p.scanline = -1
		}
	}
}

func (p *PPU) PendingNMI() (v bool) {
	v, p.pendingNMI = p.pendingNMI, false
	return v
}

func (p *PPU) ScanlineComplete() (v bool) {
	v = p.scanlineComplete
	p.scanlineComplete = false
	return v
}

func (p *PPU) FrameComplete() (v bool) {
	v = p.frameComplete
	p.frameComplete = false
	return v
}
