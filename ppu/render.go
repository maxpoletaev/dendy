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
	Back      bool
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

func (p *PPU) nameTableAddr() uint16 {
	return 0x2000 + uint16(p.Ctrl&CtrlNameTableSelect)*0x400
}

func (p *PPU) tilePatternTableAddr() uint16 {
	if p.getFlag(CtrlPatternTableSelect) {
		return 0x1000
	}

	return 0
}

func (p *PPU) spritePatternTableAddr() uint16 {
	if p.getFlag(CtrlSpritePatternAddr) {
		return 0x1000
	}

	return 0
}

func (p *PPU) fetchTile(tileX, tileY int) (tile Tile) {
	offset := p.nameTableAddr()

	tileX += int(p.ScrollX / 8)
	if tileX >= 32 {
		offset ^= 0x0400
		tileX -= 32
	}

	tileY += int(p.ScrollY / 8)
	if tileY >= 30 {
		offset ^= 0x0800
		tileY -= 30
	}

	attr := p.readVRAM(offset + 0x03C0 + uint16(tileX)/4 + uint16(tileY)/4*8)
	id := p.readVRAM(offset + uint16(tileY)*32 + uint16(tileX))
	addr := p.tilePatternTableAddr() + uint16(id)*16

	for y := 0; y < 8; y++ {
		p1 := p.readVRAM(addr + uint16(y) + 0)
		p2 := p.readVRAM(addr + uint16(y) + 8)

		for x := 0; x < 8; x++ {
			pixel := p1 & (0x80 >> x) >> (7 - x) << 0
			pixel |= (p2 & (0x80 >> x) >> (7 - x)) << 1
			tile.Pixels[x][y] = pixel // two-bit pixel value (0-3)
		}
	}

	// two-bit palette ID (0-3)
	blockId := uint16(tileX%4/2) + uint16(tileY%4/2)*2
	tile.PaletteID = (attr >> (blockId * 2)) & 0x03

	return tile
}

func (p *PPU) renderTileScanline() {
	var (
		fineX = int(p.ScrollX % 8)
		fineY = int(p.ScrollY % 8)
	)

	var (
		frameY = (p.scanline + fineY) % 248
		tileY  = int(frameY) / 8
		pixelY = int(frameY) % 8
	)

	for tileX := 0; tileX < 32; tileX++ {
		tile := p.fetchTile(tileX, tileY)

		for pixelX := 0; pixelX < 8; pixelX++ {
			frameX := tileX*8 + pixelX

			pixel := tile.Pixels[pixelX][pixelY]
			if pixel == 0 {
				continue
			}

			// To simulate scrolling, we need to offset the tile's position by the fine
			// scroll values. This is not how the PPU does it, but it seems to work.
			x, y := frameX-fineX, frameY-fineY
			if x < 0 || y < 0 || x >= 256 || y >= 240 {
				continue
			}

			addr := 0x3F00 + uint16(tile.PaletteID)*4 + uint16(pixel)
			p.Frame[x][y] = Colors[p.readVRAM(addr)]
		}
	}

	// Simulate the smooth wrap-around effect by rendering the first 8 pixels of the
	// tiles from the next name table for the rightmost 8 pixels of the frame, if we
	// are scrolling horizontally. This is a hack, but it works for most games.
	if fineX > 0 {
		tile := p.fetchTile(32, tileY)

		for pixelX := 0; pixelX < fineX; pixelX++ {
			pixel := tile.Pixels[pixelX][pixelY]
			if pixel == 0 {
				continue
			}

			addr := 0x3F00 + uint16(tile.PaletteID)*4 + uint16(pixel)
			offsetX := 32*8 - fineX + pixelX

			x, y := offsetX, frameY-fineY
			if x < 0 || y < 0 || x >= 256 || y >= 240 {
				continue
			}

			p.Frame[x][y] = Colors[p.readVRAM(addr)]
		}
	}
}

func (p *PPU) fetchSpritePixels(id uint8) (pixels [8][8]uint8) {
	addr := p.spritePatternTableAddr() + uint16(id)*16

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
			y    = p.OAMData[i*4+0]
			id   = p.OAMData[i*4+1]
			attr = p.OAMData[i*4+2]
			x    = p.OAMData[i*4+3]
		)

		if scanline < int(y) || scanline >= int(y)+8 {
			continue
		}

		if p.spriteCount < 8 {
			p.spriteScanline[p.spriteCount] = Sprite{
				Pixels:    p.fetchSpritePixels(id),
				PaletteID: attr & spriteAttrPalette,
				Back:      attr&spriteAttrPriority != 0,
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

	var (
		bgColor = Colors[p.readVRAM(0x3F00)]
	)

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

			// Sprite is behind the background, so don't render.
			if sprite.Back && p.Frame[frameX][frameY] != bgColor {
				continue
			}

			addr := 0x3F10 + uint16(sprite.PaletteID)*4 + uint16(px)
			p.Frame[frameX][frameY] = Colors[p.readVRAM(addr)]
		}
	}
}
