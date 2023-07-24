package ppu

import (
	"image/color"
)

type Tile struct {
	Pixels    [8][8]uint8
	PaletteID uint8
}

func (p *PPU) tilePatternTableAddr() uint16 {
	if p.getCtrl(CtrlPatternTableSelect) {
		return 0x1000
	}
	return 0
}

func (p *PPU) fetchTileLine(tileX, tileY, line int) (tile Tile) {
	nametableID := p.vramAddr.nametable()

	if tileX >= 32 {
		nametableID ^= 0x01
		tileX -= 32
	}

	nametableAddr := 0x2000 + uint16(nametableID)*0x0400
	tileID := p.readVRAM(nametableAddr + uint16(tileY)*32 + uint16(tileX))
	tileAddr := p.tilePatternTableAddr() + uint16(tileID)*16

	attrtableAddr := 0x23C0 + uint16(nametableID)*0x0400
	attrAddr := attrtableAddr + uint16(tileX)/4 + uint16(tileY)/4*8
	attr := p.readVRAM(attrAddr)

	p1 := p.readVRAM(tileAddr + uint16(line) + 0)
	p2 := p.readVRAM(tileAddr + uint16(line) + 8)

	for x := 0; x < 8; x++ {
		pixel := p1 & (0x80 >> x) >> (7 - x) << 0
		pixel |= (p2 & (0x80 >> x) >> (7 - x)) << 1
		tile.Pixels[x][line] = pixel // two-bit pixel value
	}

	// two-bit palette ID
	blockID := uint16(tileX%4/2) + uint16(tileY%4/2)*2
	tile.PaletteID = (attr >> (blockID * 2)) & 0x03

	return tile
}

func (p *PPU) readColor(pixel, paletteID uint8) color.RGBA {
	colorAddr := 0x3F00 + uint16(paletteID)*4 + uint16(pixel)
	colorIdx := p.readVRAM(colorAddr)
	return Colors[colorIdx]
}

func (p *PPU) renderTileScanline() {
	var (
		scrollX = p.vramAddr.coarseX()*8 + uint16(p.fineX)
		scrollY = p.vramAddr.coarseY()*8 + p.vramAddr.fineY()
		screenY = p.scanline + 1
	)

	var (
		tileY, pixelY = int(scrollY / 8), int(scrollY % 8)
		lastTileX     = -1
		tile          Tile
	)

	for screenX := 0; screenX < 256; screenX++ {
		if !p.getMask(MaskShowLeftTiles) && screenX < 8 {
			continue
		}

		scrolledX := screenX + int(scrollX)
		pixelX := scrolledX % 8
		tileX := scrolledX / 8

		// While staying on the same scanline, we only need to fetch a new tile when we
		// cross a tile boundary. We don’t need a full tile here either, just the line
		// we’re currently rendering.
		if tileX != lastTileX {
			tile = p.fetchTileLine(tileX, tileY, pixelY)
			lastTileX = tileX
		}

		pixel := tile.Pixels[pixelX][pixelY]
		if pixel == 0 {
			continue
		}

		p.Frame[screenX][screenY] = p.readColor(pixel, tile.PaletteID)
	}
}
