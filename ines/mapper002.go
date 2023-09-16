package ines

import (
	"encoding/gob"
	"errors"
	"log"
)

type Mapper2 struct {
	rom      *ROM
	prgBank0 int
	prgBank1 int
}

func NewMapper2(rom *ROM) *Mapper2 {
	return &Mapper2{
		rom: rom,
	}
}

func (m *Mapper2) Reset() {
	m.prgBank0 = 0
	m.prgBank1 = m.rom.PRGBanks - 1
}

func (m *Mapper2) ScanlineTick() {
}

func (m *Mapper2) PendingIRQ() bool {
	return false
}

func (m *Mapper2) MirrorMode() MirrorMode {
	return m.rom.MirrorMode
}

func (m *Mapper2) ReadPRG(addr uint16) byte {
	switch {
	case addr >= 0x8000 && addr <= 0xBFFF:
		// Read from the first 16KB PRG-ROM bank (fixed).
		idx := m.prgBank0*0x4000 + int(addr-0x8000)
		idx %= len(m.rom.PRG)
		return m.rom.PRG[idx]
	case addr >= 0xC000 && addr <= 0xFFFF:
		// Read from the last 16KB PRG-ROM bank (switchable).
		idx := m.prgBank1*0x4000 + int(addr-0xC000)
		idx %= len(m.rom.PRG)
		return m.rom.PRG[idx]
	default:
		log.Printf("[WARN] mapper2: unhandled prg read at 0x%04X", addr)
		return 0
	}
}

func (m *Mapper2) WritePRG(addr uint16, data byte) {
	// The lower 4 bits of the data written to 0x8000-0xFFFF select the
	// 16KB PRG-ROM bank at 0xC000-0xFFFF. The upper 4 bits are ignored.
	m.prgBank0 = int(data & 0x0F)
}

func (m *Mapper2) ReadCHR(addr uint16) byte {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		return m.rom.CHR[addr]
	default:
		log.Printf("[WARN] mapper2: unhandled chr read at 0x%04X", addr)
		return 0
	}
}

func (m *Mapper2) WriteCHR(addr uint16, data byte) {
	m.rom.CHR[addr] = data
}

func (m *Mapper2) Save(enc *gob.Encoder) error {
	return errors.Join(
		m.rom.Save(enc),
		enc.Encode(m.prgBank0),
		enc.Encode(m.prgBank1),
	)
}

func (m *Mapper2) Load(dec *gob.Decoder) error {
	return errors.Join(
		m.rom.Load(dec),
		dec.Decode(&m.prgBank0),
		dec.Decode(&m.prgBank1),
	)
}
