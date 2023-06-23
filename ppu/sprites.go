package ppu

const (
	spriteAttrPalette  = 0x03 // two bits
	spriteAttrPriority = 1 << 5
	spriteAttrFlipX    = 1 << 6
	spriteAttrFlipY    = 1 << 7
)

type Sprite struct {
	Pixels    [8][16]uint8
	PaletteID uint8
	X, Y      uint8
	FlipX     bool
	FlipY     bool
	Back      bool
}

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

func (p *PPU) spritePatternTableAddr() uint16 {
	if p.getCtrl(CtrlSpritePatternAddr) {
		return 0x1000
	}

	return 0
}

func (p *PPU) spriteHeight() int {
	if p.getCtrl(CtrlSpriteSize) {
		return 16
	}

	return 8
}

func (p *PPU) fetchSprite(idx int) Sprite {
	var (
		id      = p.oamData[idx*4+1]
		attr    = p.oamData[idx*4+2]
		spriteX = p.oamData[idx*4+3]
		spriteY = p.oamData[idx*4+0]
		height  = p.spriteHeight()
	)

	sprite := Sprite{
		PaletteID: attr & spriteAttrPalette,
		Back:      attr&spriteAttrPriority != 0,
		FlipX:     attr&spriteAttrFlipX != 0,
		FlipY:     attr&spriteAttrFlipY != 0,
		Y:         spriteY,
		X:         spriteX,
	}

	var addr uint16

	for y := 0; y < height; y++ {
		if height == 16 {
			table := id & 1
			tile := id & 0xFE
			if y >= 8 {
				tile++
			}
			addr = uint16(table)*0x1000 + uint16(tile)*16 + uint16(y&7)
		} else {
			addr = p.spritePatternTableAddr() + uint16(id)*16 + uint16(y)
		}

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

func (p *PPU) prepareSprites() {
	scanline := p.scanline + 1
	height := p.spriteHeight()
	p.spriteCount = 0

	for i := 0; i < 64; i++ {
		spriteY := int(p.oamData[i*4+0])
		if scanline < spriteY || scanline >= spriteY+height {
			continue
		}

		if p.spriteCount < 8 {
			p.spriteScanline[p.spriteCount] = p.fetchSprite(i)
		}

		p.spriteCount++
	}

	if p.spriteCount > 8 {
		p.setStatus(StatusSpriteOverflow, true)
		p.spriteCount = 8
	}
}

func (p *PPU) renderSpriteScanline() {
	frameY := p.scanline
	if frameY > 239 {
		return
	}

	var (
		bgColor = p.backdropColor()
		height  = p.spriteHeight()
	)

	for i := p.spriteCount - 1; i >= 0; i-- {
		sprite := p.spriteScanline[i]

		if sprite.Y > 239 || sprite.Y == 0 {
			continue
		}

		pixels := flipPixels(sprite.Pixels, sprite.FlipX, sprite.FlipY, height)
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
