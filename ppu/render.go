package ppu

import (
	"image/color"
)

type Tile struct {
	Pixels    [8][8]uint8
	PaletteID uint8
}

type Sprite struct {
	Pixels    [8][8]uint8
	PaletteID uint8
	X, Y      uint8
	FlipX     bool
	FlipY     bool
	BehindBG  bool
}

const (
	spriteAttrPalette  = 0x03 // two bits
	spriteAttrPriority = 1 << 5
	spriteAttrFlipX    = 1 << 6
	spriteAttrFlipY    = 1 << 7
)

func flipPixels(pixels [8][8]uint8, flipX, flipY bool) (flipped [8][8]uint8) {
	if !flipX && !flipY {
		return pixels
	}

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			fx, fy := x, y

			if flipX {
				fx = 7 - x
			}

			if flipY {
				fy = 7 - y
			}

			flipped[fx][fy] = pixels[x][y]
		}
	}

	return flipped
}

func (p *PPU) clearFrame(c color.RGBA) {
	for x := 0; x < 256; x++ {
		for y := 0; y < 240; y++ {
			p.Frame[x][y] = c
		}
	}
}

func (p *PPU) nameTableOffset() uint16 {
	return 0x2000 + uint16(p.Ctrl&0x03)*0x400
}

func (p *PPU) tilePatternTableOffset() uint16 {
	if p.getFlag(CtrlPatternTableSelect) {
		return 0x1000
	}

	return 0
}

func (p *PPU) spritePatternTableOffset() uint16 {
	if p.getFlag(CtrlSpritePatternAddr) {
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
	addr := p.tilePatternTableOffset() + uint16(id)*16

	for y := 0; y < 8; y++ {
		p1 := p.readVRAM(addr + uint16(y) + 0)
		p2 := p.readVRAM(addr + uint16(y) + 8)

		for x := 0; x < 8; x++ {
			px := p1 & (0x80 >> x) >> (7 - x) << 0
			px |= (p2 & (0x80 >> x) >> (7 - x)) << 1
			tile.Pixels[x][y] = px // two-bit pixel value (0-3)
		}
	}

	// two-bit palette ID (0-3)
	tile.PaletteID = attr & 0x03

	return tile
}

func (p *PPU) renderTileScanline() {
	y := p.scanline
	tileY := int(y) / 8
	offsetY := int(y) % 8

	for tileX := 0; tileX < 32; tileX++ {
		tile := p.fetchTile(tileX, tileY)

		for x := 0; x < 8; x++ {
			offsetX := tileX*8 + x

			px := tile.Pixels[x][offsetY]
			if px == 0 {
				continue
			}

			paletteID := tile.PaletteID
			p.Frame[offsetX][y] = Colors[p.PaletteTable[paletteID*4+px]]
		}
	}
}

func (p *PPU) fetchSpritePixels(id uint8) (pixels [8][8]uint8) {
	addr := p.spritePatternTableOffset() + uint16(id)*16

	for y := 0; y < 8; y++ {
		p1 := p.readVRAM(addr + uint16(y) + 0)
		p2 := p.readVRAM(addr + uint16(y) + 8)

		for x := 0; x < 8; x++ {
			px := p1 & (0x80 >> x) >> (7 - x) << 0
			px |= (p2 & (0x80 >> x) >> (7 - x)) << 1
			pixels[x][y] = px // two-bit pixel value (0-3)
		}
	}

	return pixels
}

func (p *PPU) prepareSprites() {
	scanline := p.scanline + 1
	p.spriteCount = 0

	for i := 0; i < 64; i++ {
		var (
			id   = p.OAMData[i*4+1]
			x    = p.OAMData[i*4+3]
			y    = p.OAMData[i*4+0]
			attr = p.OAMData[i*4+2]
		)

		if scanline < int(y) || scanline >= int(y)+8 {
			continue
		}

		if p.spriteCount < 8 {
			p.spriteScanline[p.spriteCount] = Sprite{
				Pixels:    p.fetchSpritePixels(id),
				PaletteID: attr & spriteAttrPalette,
				BehindBG:  attr&spriteAttrPriority != 0,
				FlipX:     attr&spriteAttrFlipX != 0,
				FlipY:     attr&spriteAttrFlipY != 0,
				Y:         y,
				X:         x,
			}
		}

		p.spriteCount++
	}

	if p.spriteCount > 8 {
		p.setFlag(StatusSpriteOverflow, true)
		p.spriteCount = 8
	}
}

func (p *PPU) renderSpriteScanline() {
	frameY := p.scanline
	if frameY > 239 {
		return
	}

	for i := 0; i < p.spriteCount; i++ {
		sprite := p.spriteScanline[i]
		if sprite.Y > 239 || sprite.Y == 0 {
			continue
		}

		pixels := flipPixels(sprite.Pixels, sprite.FlipX, sprite.FlipY)
		pixelY := p.scanline - int(sprite.Y)

		for pixelX := 0; pixelX < 8; pixelX++ {
			frameX := int(sprite.X) + pixelX
			if frameX > 255 {
				continue
			}

			px := pixels[pixelX][pixelY]
			if px == 0 {
				continue
			}

			if sprite.BehindBG && p.Frame[frameX][frameY].A != 0 {
				continue
			}

			paletteIdx := sprite.PaletteID*4 + px
			p.Frame[frameX][frameY] = Colors[p.PaletteTable[paletteIdx]]
		}
	}
}
