package ppu

import (
	"image/color"
)

const (
	spriteAttrPalette  = 0x03 // two bits
	spriteAttrPriority = 1 << 5
	spriteAttrFlipX    = 1 << 6
	spriteAttrFlipY    = 1 << 7
)

type Sprite struct {
	Index     int
	Pixels    [8][16]uint8
	PaletteID uint8
	X, Y      uint8
	FlipX     bool
	FlipY     bool
	Behind    bool
}

// flipPixels flips the given pixels horizontally and/or vertically depending on the flip flags.
func flipPixels(pixels [8][16]uint8, flipX, flipY bool, height int) (flipped [8][16]uint8) {
	if !flipX && !flipY {
		return pixels
	}

	for y := 0; y < height; y++ {
		for x := 0; x < 8; x++ {
			fx, fy := x, y

			if flipX {
				fx = 7 - x
			}

			if flipY {
				fy = height - y - 1
			}

			flipped[fx][fy] = pixels[x][y]
		}
	}

	return flipped
}

// spritePatternTableOffset returns the address offset in VRAM for the sprite pattern table.
func (p *PPU) spritePatternTableOffset() uint16 {
	if p.getCtrl(CtrlSpritePatternAddr) {
		return 0x1000
	}

	return 0
}

// spriteHeight returns the height of sprites in pixels (8 or 16).
func (p *PPU) spriteHeight() int {
	if p.getCtrl(CtrlSpriteSize) {
		return 16
	}

	return 8
}

// spriteAddr returns the address in VRAM for the given sprite ID and y coordinate.
func (p *PPU) spriteAddr(tableOffset uint16, spriteID uint8, y, height int) uint16 {
	if height == 16 {
		table := spriteID & 0x01
		tile := spriteID & 0xFE
		if y >= 8 {
			tile++
		}

		return uint16(table)*0x1000 + uint16(tile)*16 + uint16(y&7)
	}

	return tableOffset + uint16(spriteID)*16 + uint16(y)
}

// fetchSpritePixel returns the pixel value (0-3) for the given sprite at the
// given x/y coordinates. This is used for sprite zero hit detection, when we
// don’t need the whole sprite data, and just want to know if it’s visible.
func (p *PPU) fetchSpritePixel(idx int, x, y int) uint8 {
	var (
		tableAddr = p.spritePatternTableOffset()
		spriteID  = p.oamData[idx*4+1]
		height    = p.spriteHeight()
	)

	addr := p.spriteAddr(tableAddr, spriteID, y, height)
	p1 := p.readVRAM(addr + 0)
	p2 := p.readVRAM(addr + 8)

	px := p1 & (0x80 >> x) >> (7 - x) << 0
	px |= (p2 & (0x80 >> x) >> (7 - x)) << 1

	return px
}

// fetchSprite returns the sprite data for the given sprite index.
func (p *PPU) fetchSprite(idx int) Sprite {
	var (
		spriteID  = p.oamData[idx*4+1]
		attr      = p.oamData[idx*4+2]
		spriteX   = p.oamData[idx*4+3]
		spriteY   = p.oamData[idx*4+0]
		height    = p.spriteHeight()
		tableAddr = p.spritePatternTableOffset()
	)

	sprite := Sprite{
		Index:     idx,
		PaletteID: attr & spriteAttrPalette,
		Behind:    attr&spriteAttrPriority != 0,
		FlipX:     attr&spriteAttrFlipX != 0,
		FlipY:     attr&spriteAttrFlipY != 0,
		Y:         spriteY,
		X:         spriteX,
	}

	for y := 0; y < height; y++ {
		addr := p.spriteAddr(tableAddr, spriteID, y, height)
		p1 := p.readVRAM(addr + 0)
		p2 := p.readVRAM(addr + 8)

		for x := 0; x < 8; x++ {
			px := p1 & (0x80 >> x) >> (7 - x) << 0
			px |= (p2 & (0x80 >> x) >> (7 - x)) << 1
			sprite.Pixels[x][y] = px // two-bit pixel value (0-3)
		}
	}

	return sprite
}

// evaluateSprites checks which sprites will be visible on the next scanline, and
// stores them in the p.spriteScanline array.
func (p *PPU) evaluateSprites() {
	height := p.spriteHeight()
	scanline := p.scanline + 1
	p.spriteCount = 0

	for i := 0; i < 64; i++ {
		spriteY := int(p.oamData[i*4+0])

		if scanline < spriteY || scanline >= spriteY+height {
			continue
		}

		if p.spriteCount == 8 {
			p.setStatus(StatusSpriteOverflow, true)

			if !p.NoSpriteLimit {
				break
			}
		}

		p.spriteScanline[p.spriteCount] = p.fetchSprite(i)

		p.spriteCount++
	}
}

// readSpriteColor returns the RGBA color for the given pixel value and palette ID.
func (p *PPU) readSpriteColor(pixel, paletteID uint8) color.RGBA {
	colorAddr := 0x3F10 + uint16(paletteID)*4 + uint16(pixel)
	colorIdx := p.readVRAM(colorAddr)
	return Colors[colorIdx%64]
}

// renderSpriteScanline renders the sprites currently in the p.spriteScanline array.
func (p *PPU) renderSpriteScanline() {
	frameY := p.scanline
	if frameY > 239 {
		return
	}

	var (
		height = p.spriteHeight()
	)

	for i := p.spriteCount - 1; i >= 0; i-- {
		sprite := p.spriteScanline[i]

		pixels := flipPixels(sprite.Pixels, sprite.FlipX, sprite.FlipY, height)
		pixelY := p.scanline - int(sprite.Y)

		// FIXME: sometimes crashes due to pixelY being negative or > 15
		//  (e.g. ninja cat, when you climb the first ladder).
		if pixelY < 0 || pixelY >= height {
			continue
		}

		for pixelX := 0; pixelX < 8; pixelX++ {
			frameX := int(sprite.X) + pixelX
			if frameX > 255 {
				continue
			}

			if !p.getMask(MaskShowLeftSprites) && frameX < 8 {
				continue
			}

			px := pixels[pixelX][pixelY]
			if px == 0 {
				continue
			}

			// Sprite zero hit detection.
			if sprite.Index == 0 && !p.transparent[frameX][frameY] {
				p.setStatus(StatusSpriteZeroHit, true)
			}

			// Sprite is behind the background, so don't render.
			if sprite.Behind && !p.transparent[frameX][frameY] {
				continue
			}

			p.Frame[frameX][frameY] = p.readSpriteColor(px, sprite.PaletteID)
		}
	}
}
