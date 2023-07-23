package ppu

// vramAddr represents a 15-bit video ram address, that either can be manipulated
// as a whole, or partially as a coarse X/Y, fine Y, or nametable address
// component. The layout is as follows:
//
// yyy NN YYYYY XXXXX
// ||| || ||||| +++++-- coarse X scroll
// ||| || +++++-------- coarse Y scroll
// ||| ++-------------- nametable select
// +++----------------- fine Y scroll
//
// The original idea belongs to a guy with nickname loopy from the NesDev mailing
// list, and can be found in a document called "The skinny on NES scrolling"
// dated 13 Apr 1999. Here is the thread on the nesdev forum with some
// clarifications: https://forums.nesdev.org/viewtopic.php?t=664
//
// The code that increments scroll X/Y coordinates is basically a copy-paste from
// the nesdev wiki: https://www.nesdev.org/wiki/PPU_scrolling#Wrapping_around
type vramAddr uint16

const (
	bitsCoarseX    uint16 = 0b0000000000011111
	bitsCoarseY    uint16 = 0b0000001111100000
	bitsFineY      uint16 = 0b0111000000000000
	bitsNametable  uint16 = 0b0000110000000000
	bitsNametableX uint16 = 0b0000010000000000
	bitsNametableY uint16 = 0b0000100000000000
)

func (v *vramAddr) coarseX() uint16    { return (uint16(*v) & bitsCoarseX) >> 0 }
func (v *vramAddr) coarseY() uint16    { return (uint16(*v) & bitsCoarseY) >> 5 }
func (v *vramAddr) nametable() uint16  { return (uint16(*v) & bitsNametable) >> 10 }
func (v *vramAddr) nametableX() uint16 { return (uint16(*v) & bitsNametableX) >> 10 }
func (v *vramAddr) nametableY() uint16 { return (uint16(*v) & bitsNametableY) >> 11 }
func (v *vramAddr) fineY() uint16      { return (uint16(*v) & bitsFineY) >> 12 }

func (v *vramAddr) setCoarseX(x uint16)    { *v = vramAddr((uint16(*v) & ^bitsCoarseX) | (x << 0)) }
func (v *vramAddr) setCoarseY(y uint16)    { *v = vramAddr((uint16(*v) & ^bitsCoarseY) | (y << 5)) }
func (v *vramAddr) setNametable(n uint16)  { *v = vramAddr((uint16(*v) & ^bitsNametable) | (n << 10)) }
func (v *vramAddr) setNametableX(n uint16) { *v = vramAddr((uint16(*v) & ^bitsNametableX) | (n << 10)) }
func (v *vramAddr) setNametableY(n uint16) { *v = vramAddr((uint16(*v) & ^bitsNametableY) | (n << 11)) }
func (v *vramAddr) setFineY(y uint16)      { *v = vramAddr((uint16(*v) & ^bitsFineY) | (y << 12)) }

func (v *vramAddr) flipNametableX() { *v ^= 0x0400 }
func (v *vramAddr) flipNametableY() { *v ^= 0x0800 }

func (v *vramAddr) incrementX() {
	if v.coarseX() == 31 {
		v.flipNametableX()
		v.setCoarseX(0)
	} else {
		*v++
	}
}

func (v *vramAddr) incrementY() {
	if v.fineY() < 7 {
		*v += 0x1000
	} else {
		y := v.coarseY()
		v.setFineY(0)
		if y == 29 {
			v.flipNametableY()
			y = 0
		} else if y == 31 {
			y = 0
		} else {
			y++
		}
		v.setCoarseY(y)
	}
}
