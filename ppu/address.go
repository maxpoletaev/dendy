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
	bitsCoarseX    uint16 = 0b0_000_00_00000_11111
	bitsCoarseY    uint16 = 0b0_000_00_11111_00000
	bitsNametable  uint16 = 0b0_000_11_00000_00000
	bitsNametableX uint16 = 0b0_000_01_00000_00000
	bitsNametableY uint16 = 0b0_000_10_00000_00000
	bitsFineY      uint16 = 0b0_111_00_00000_00000

	bitsInvCoarseX    = ^bitsCoarseX
	bitsInvCoarseY    = ^bitsCoarseY
	bitsInvNametable  = ^bitsNametable
	bitsInvNametableX = ^bitsNametableX
	bitsInvNametableY = ^bitsNametableY
	bitsInvFineY      = ^bitsFineY
)

func (v *vramAddr) get(mask uint16, shift uint8) uint16 {
	return (uint16(*v) & mask) >> shift
}

func (v *vramAddr) set(mask uint16, shift uint8, value uint16) {
	*v = vramAddr((uint16(*v) & mask) | (value << shift))
}

func (v *vramAddr) coarseX() uint16    { return v.get(bitsCoarseX, 0) }
func (v *vramAddr) coarseY() uint16    { return v.get(bitsCoarseY, 5) }
func (v *vramAddr) nametable() uint16  { return v.get(bitsNametable, 10) }
func (v *vramAddr) nametableX() uint16 { return v.get(bitsNametableX, 10) }
func (v *vramAddr) nametableY() uint16 { return v.get(bitsNametableY, 11) }
func (v *vramAddr) fineY() uint16      { return v.get(bitsFineY, 12) }

func (v *vramAddr) setCoarseX(x uint16)    { v.set(bitsInvCoarseX, 0, x) }
func (v *vramAddr) setCoarseY(y uint16)    { v.set(bitsInvCoarseY, 5, y) }
func (v *vramAddr) setNametable(n uint16)  { v.set(bitsInvNametable, 10, n) }
func (v *vramAddr) setNametableX(n uint16) { v.set(bitsInvNametableX, 10, n) }
func (v *vramAddr) setNametableY(n uint16) { v.set(bitsInvNametableY, 11, n) }
func (v *vramAddr) setFineY(y uint16)      { v.set(bitsInvFineY, 12, y) }

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
