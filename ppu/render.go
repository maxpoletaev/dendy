package ppu

import (
	"fmt"
	"image/color"
)

type Tile struct {
	Pixels    [8][8]uint8
	PaletteID uint8
}

func (p *PPU) nameTableOffset() uint16 {
	return 0x2000 + uint16(p.Ctrl&0x03)*0x400
}

func (p *PPU) patternTableOffset() uint16 {
	if p.getFlag(CtrlPatternTableSelect) {
		return 0x1000
	}

	return 0
}

func (p *PPU) tileID(tileX, tileY int) uint8 {
	return p.readVRAM(p.nameTableOffset() + uint16(tileY)*32 + uint16(tileX))
}

func (p *PPU) tileAttr(tileX, tileY int) uint8 {
	return p.readVRAM(p.nameTableOffset() + 0x03C0 + uint16(tileY)/32*8 + uint16(tileX)/32)
}

func (p *PPU) fetchTile(tileX, tileY int) (tile Tile) {
	id := p.tileID(tileX, tileY)
	attr := p.tileAttr(tileX, tileY)
	addr := p.patternTableOffset() + uint16(id)*16

	for y := 0; y < 8; y++ {
		plane1 := p.readVRAM(addr + uint16(y) + 0)
		plane2 := p.readVRAM(addr + uint16(y) + 8)

		for x := 0; x < 8; x++ {
			px := plane1 & (0x80 >> x) >> (7 - x) << 0
			px |= (plane2 & (0x80 >> x) >> (7 - x)) << 1
			tile.Pixels[x][y] = px // two-bit pixel value (0-3)
		}
	}

	// two-bit palette ID (0-3)
	tile.PaletteID = attr & 0x03

	return tile
}

func (p *PPU) renderTile(tile Tile, tileX, tileY int) {
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			px := tile.Pixels[x][y]
			if px == 0 {
				continue
			}

			rgb := Colors[p.PaletteTable[px+tile.PaletteID*4]]
			p.Frame[tileX*8+x][tileY*8+y] = rgb
		}
	}
}

func (p *PPU) clearFrame(c color.RGBA) {
	for x := 0; x < 256; x++ {
		for y := 0; y < 240; y++ {
			p.Frame[x][y] = c
		}
	}
}

func (p *PPU) renderBackground() {
	for tileY := 0; tileY < 30; tileY++ {
		for tileX := 0; tileX < 32; tileX++ {
			tile := p.fetchTile(tileX, tileY)
			p.renderTile(tile, tileX, tileY)
		}
	}
}

func (p *PPU) printNamenable() {
	for tileY := 0; tileY < 32; tileY++ {
		for tileX := 0; tileX < 32; tileX++ {
			tileID := p.NameTable[0][tileY*32+tileX]
			fmt.Printf("%02X ", tileID)
		}

		fmt.Println()
	}

	fmt.Println()
}
