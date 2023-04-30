package ppu

import "fmt"

type Tile struct {
	Pixels    [8][8]uint8
	PaletteID uint8
}

func (p *PPU) getNameTableOffset() uint16 {
	return 0x2000 + uint16(p.Ctrl&0x03)*0x400
}

func (p *PPU) getPatternTableOffset() uint16 {
	return 0x1000 * uint16(p.Ctrl&CtrlPatternTableSelect>>4)
}

func (p *PPU) getTileID(tileX, tileY int) uint8 {
	return p.readVRAM(p.getNameTableOffset() + uint16(tileY)*30 + uint16(tileX))
}

func (p *PPU) getTileAttribute(tileX, tileY int) uint8 {
	return p.readVRAM(p.getNameTableOffset() + 0x03C0 + uint16(tileX)/32*8 + uint16(tileY)/32)
}

func (p *PPU) fetchTile(tileX, tileY int) (tile Tile) {
	id := p.getTileID(tileX, tileY)
	attr := p.getTileAttribute(tileX, tileY)
	addr := p.getPatternTableOffset() + uint16(id)*16

	for y := 0; y < 8; y++ {
		plane1 := p.readVRAM(addr + uint16(y) + 0)
		plane2 := p.readVRAM(addr + uint16(y) + 8)

		for x := 0; x < 8; x++ {
			px := plane1 & (0x80 >> x) >> (7 - x) << 0
			px |= (plane2 & (0x80 >> x) >> (7 - x)) << 1
			tile.Pixels[x][y] = px // 0, 1, 2, 3
		}
	}

	tile.PaletteID = attr & 0x03

	return tile
}

func (p *PPU) renderTile(tile Tile, tileX, tileY int) {
	var (
		rgb uint32
		px  uint8
	)

	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			px = tile.Pixels[x][y]
			if px == 0 {
				continue
			}

			switch px {
			case 0:
				rgb = Colors[1]
			case 1:
				rgb = Colors[4]
			case 2:
				rgb = Colors[7]
			case 3:
				rgb = Colors[10]
			default:
				panic("invalid pixel")
			}

			//clr := Palette[p.PaletteTable[px+tile.PaletteID*4]]
			p.Frame[tileX*8+x][tileY*8+y] = toRGBA(rgb)
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
	for tileY := 0; tileY < 30; tileY++ {
		for tileX := 0; tileX < 32; tileX++ {
			tileID := p.getTileID(tileX, tileY)
			fmt.Printf("%02X ", tileID)
		}

		fmt.Println()
	}

	fmt.Println()
}
