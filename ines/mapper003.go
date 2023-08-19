package ines

import (
	"encoding/gob"
	"errors"
	"log"
)

type Mapper3 struct {
	rom      *ROM
	chrBank0 int
	prgBank0 int
	prgBank1 int
}

func NewMapper3(rom *ROM) *Mapper3 {
	return &Mapper3{
		rom: rom,
	}
}

func (m *Mapper3) Reset() {
	m.chrBank0 = 0
	m.prgBank0 = 0
	m.prgBank1 = m.rom.PRGBanks - 1
}

func (m *Mapper3) Scanline() TickInfo {
	return TickInfo{}
}

func (m *Mapper3) MirrorMode() MirrorMode {
	return m.rom.MirrorMode
}

func (m *Mapper3) ReadPRG(addr uint16) byte {
	switch {
	case addr >= 0x8000 && addr <= 0xBFFF:
		idx := m.prgBank0*0x4000 + int(addr-0x8000)
		idx %= len(m.rom.PRG)
		return m.rom.PRG[idx]
	case addr >= 0xC000 && addr <= 0xFFFF:
		idx := m.prgBank1*0x4000 + int(addr-0xC000)
		idx %= len(m.rom.PRG)
		return m.rom.PRG[idx]
	default:
		return 0
	}
}

func (m *Mapper3) WritePRG(addr uint16, data byte) {
	m.chrBank0 = int(data & 0x03)
}

func (m *Mapper3) ReadCHR(addr uint16) byte {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		return m.rom.CHR[addr]
	default:
		log.Printf("[WARN] mapper3: unhandled chr read at 0x%04X", addr)
		return 0
	}
}

func (m *Mapper3) WriteCHR(addr uint16, data byte) {
	m.rom.CHR[addr] = data
}

func (m *Mapper3) Save(enc *gob.Encoder) error {
	return errors.Join(
		m.rom.Save(enc),
		enc.Encode(m.chrBank0),
		enc.Encode(m.prgBank0),
		enc.Encode(m.prgBank1),
	)
}

func (m *Mapper3) Load(dec *gob.Decoder) error {
	return errors.Join(
		m.rom.Load(dec),
		dec.Decode(&m.chrBank0),
		dec.Decode(&m.prgBank0),
		dec.Decode(&m.prgBank1),
	)
}
