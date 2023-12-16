package ines

import (
	"log"

	"github.com/maxpoletaev/dendy/internal/binario"
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

func (m *Mapper0) ScanlineTick() {
}

func (m *Mapper0) PendingIRQ() bool {
	return false
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
		log.Printf("[WARN] mapper0: unhandled prg read at %04X", addr)
		return 0
	}
}

func (m *Mapper0) WritePRG(addr uint16, data byte) {
	log.Printf("[WARN] mapper0: write to read-only prg at %04X", addr)
}

func (m *Mapper0) ReadCHR(addr uint16) byte {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		return m.rom.CHR[addr]
	default:
		log.Printf("[WARN] mapper0: unhandled chr read at %04X", addr)
		return 0
	}
}

func (m *Mapper0) WriteCHR(addr uint16, data byte) {
	log.Printf("[WARN] mapper0: write to read-only chr at %04X", addr)
}

func (m *Mapper0) SaveState(w *binario.Writer) error {
	return m.rom.SaveState(w)
}

func (m *Mapper0) LoadState(r *binario.Reader) error {
	return m.rom.LoadState(r)
}
