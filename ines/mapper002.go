package ines

import "fmt"

type Mapper2 struct {
	rom       *ROM
	prgBankLo int
	prgBankHi int
}

func NewMapper2(rom *ROM) *Mapper2 {
	return &Mapper2{
		rom: rom,
	}
}

func (m *Mapper2) Reset() {
	m.prgBankLo = 0
	m.prgBankHi = m.rom.PRGBanks - 1
}

func (m *Mapper2) MirrorMode() MirrorMode {
	return m.rom.MirrorMode
}

func (m *Mapper2) ReadPRG(addr uint16) byte {
	switch {
	case addr >= 0x8000 && addr <= 0xBFFF:
		// Read from the first 16KB PRG-ROM bank (fixed).
		idx := m.prgBankLo*0x4000 + int(addr-0x8000)
		return m.rom.PRG[idx]
	case addr >= 0xC000 && addr <= 0xFFFF:
		// Read from the last 16KB PRG-ROM bank (switchable).
		idx := m.prgBankHi*0x4000 + int(addr-0xC000)
		return m.rom.PRG[idx]
	default:
		panic(fmt.Sprintf("mapper2: unhandled read at 0x%04X", addr))
	}
}

func (m *Mapper2) WritePRG(addr uint16, data byte) {
	// The lower 4 bits of the data written to 0x8000-0xFFFF select the
	// 16KB PRG-ROM bank at 0xC000-0xFFFF. The upper 4 bits are ignored.
	m.prgBankLo = int(data & 0x0F)
}

func (m *Mapper2) ReadCHR(addr uint16) byte {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		return m.rom.CHR[addr]
	default:
		panic(fmt.Sprintf("mapper2: unhandled read at 0x%04X", addr))
	}
}

func (m *Mapper2) WriteCHR(addr uint16, data byte) {
	m.rom.CHR[addr] = data
}
