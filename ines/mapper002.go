package ines

import (
	"errors"
	"log"

	"github.com/maxpoletaev/dendy/internal/binario"
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
	if !m.rom.chrRAM {
		log.Printf("[WARN] mapper2: write to read-only chr at %04X", addr)
		return
	}

	m.rom.CHR[addr] = data
}

func (m *Mapper2) SaveState(w *binario.Writer) error {
	return errors.Join(
		m.rom.SaveState(w),
		w.WriteUint64(uint64(m.prgBank0)),
		w.WriteUint64(uint64(m.prgBank1)),
	)
}

func (m *Mapper2) LoadState(w *binario.Reader) error {
	var (
		prgBank0 uint64
		prgBank1 uint64
	)

	err := errors.Join(
		m.rom.LoadState(w),
		w.ReadUint64To(&prgBank0),
		w.ReadUint64To(&prgBank1),
	)

	m.prgBank0 = int(prgBank0)
	m.prgBank1 = int(prgBank1)

	return err
}
