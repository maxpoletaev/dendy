package ines

import (
	"encoding/gob"
	"fmt"
)

// Mapper0 is the simplest mapper. It has no registers, and it only supports
// 16KB or 32KB PRG-ROM banks and 8KB CHR-ROM banks.
//
//	PRG-ROM is mapped to 0x8000-0xFFFF.
//	CHR-ROM is mapped to 0x0000-0x1FFF.
type Mapper0 struct {
	rom *ROM
}

func NewMapper0(cart *ROM) *Mapper0 {
	return &Mapper0{
		rom: cart,
	}
}

func (m *Mapper0) Reset() {

}

func (m *Mapper0) MirrorMode() MirrorMode {
	return m.rom.MirrorMode
}

func (m *Mapper0) ReadPRG(addr uint16) byte {
	switch {
	case addr >= 0x8000 && addr <= 0xFFFF:
		idx := addr % uint16(len(m.rom.PRG))
		return m.rom.PRG[idx]
	default:
		panic(fmt.Sprintf("mapper0: unhandled read at 0x%04X", addr))
	}
}

func (m *Mapper0) WritePRG(addr uint16, data byte) {

}

func (m *Mapper0) ReadCHR(addr uint16) byte {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		return m.rom.CHR[addr]
	}

	panic(fmt.Sprintf("mapper0: unhandled read at 0x%04X", addr))
}

func (m *Mapper0) WriteCHR(addr uint16, data byte) {

}

func (m *Mapper0) Save(enc *gob.Encoder) error {
	return nil
}

func (m *Mapper0) Load(dec *gob.Decoder) error {
	return nil
}
