package ines

import (
	"errors"
	"fmt"
	"log"

	"github.com/maxpoletaev/dendy/internal/binario"
)

// Mapper1 implements the MMC1 mapper.
// https://www.nesdev.org/wiki/MMC1
type Mapper1 struct {
	rom  *ROM
	sram [0x2000]byte

	control  byte
	prgBank  byte
	chrBank0 byte
	chrBank1 byte

	shiftRegister byte
	writeCount    byte
}

func NewMapper1(rom *ROM) *Mapper1 {
	return &Mapper1{
		rom: rom,
	}
}

func (m *Mapper1) Reset() {
	m.control = 0x0C
	m.prgBank = 0
	m.chrBank0 = 0
	m.chrBank1 = 0

	m.shiftRegister = 0x10
	m.writeCount = 0
}

func (m *Mapper1) ScanlineTick() {
}

func (m *Mapper1) PendingIRQ() bool {
	return false
}

func (m *Mapper1) MirrorMode() MirrorMode {
	switch m.control & 0x03 {
	case 0:
		return MirrorSingle0
	case 1:
		return MirrorSingle1
	case 2:
		return MirrorVertical
	case 3:
		return MirrorHorizontal
	default:
		panic("mapper1: invalid mirror mode")
	}
}

func (m *Mapper1) prgMode() byte {
	// 0, 1: switch 32 KB at $8000, ignoring low bit of bank number;
	// 2: fix first bank at $8000 and switch 16 KB bank at $C000;
	// 3: fix last bank at $C000 and switch 16 KB bank at $8000
	return (m.control >> 2) & 0x03
}

func (m *Mapper1) chrMode() byte {
	// 0: switch 8 KB at a time;
	// 1: switch two separate 4 KB banks
	return (m.control >> 4) & 0x01
}

func (m *Mapper1) writeRegister(addr uint16, data byte) {
	switch {
	case addr >= 0x0000 && addr <= 0x9FFF:
		m.control = data
	case addr >= 0xA000 && addr <= 0xBFFF:
		m.chrBank0 = data & 0x1F
	case addr >= 0xC000 && addr <= 0xDFFF:
		m.chrBank1 = data & 0x1F
	case addr >= 0xE000 && addr <= 0xFFFF:
		m.prgBank = data & 0x0F
	}
}

func (m *Mapper1) loadRegister(addr uint16, data byte) {
	if data&0x80 != 0 {
		// Reset the shift register if the leftmost bit is set.
		m.shiftRegister = 0x10
		m.control |= 0x0C
		m.writeCount = 0
	} else {
		m.shiftRegister >>= 1
		m.shiftRegister |= (data & 1) << 4
		m.writeCount++

		// Once we have 5 bits written to the shift register, we can copy its contents to
		// the target register, which is determined by the address written to.
		if m.writeCount == 5 {
			m.writeRegister(addr, m.shiftRegister)
			m.shiftRegister = 0x10
			m.writeCount = 0
		}
	}
}

func (m *Mapper1) prgBankIndex() (uint, uint) {
	switch m.prgMode() {
	case 0, 1: // Switch 32 KB at $8000, ignoring low bit of bank number.
		return uint(m.prgBank & 0xFE), uint(m.prgBank | 0x01)
	case 2: // Fix first bank at $8000 and switch 16 KB bank at $C000.
		return 0, uint(m.prgBank)
	case 3: // Fix last bank at $C000 and switch 16 KB bank at $8000.
		return uint(m.prgBank), uint(m.rom.PRGBanks - 1)
	default:
		panic(fmt.Sprintf("mapper1: invalid prg mode: %d", m.prgMode()))
	}
}

func (m *Mapper1) prgOffset(idx uint) uint {
	return idx * 0x4000
}

func (m *Mapper1) ReadPRG(addr uint16) byte {
	bank0, bank1 := m.prgBankIndex()

	switch {
	case addr >= 0x6000 && addr <= 0x7FFF: // PRG-RAM
		return m.sram[addr-0x6000]
	case addr >= 0x8000 && addr <= 0xBFFF: // PRG-ROM, bank 0
		relAddr := uint((addr - 0x8000) % 0x4000)
		return m.rom.PRG[m.prgOffset(bank0)+relAddr]
	case addr >= 0xC000 && addr <= 0xFFFF: // PRG-ROM, bank 1
		relAddr := uint((addr - 0x8000) % 0x4000)
		return m.rom.PRG[m.prgOffset(bank1)+relAddr]
	default:
		log.Printf("[WARN] mapper1: unhandled prg read at %04X", addr)
		return 0
	}
}

func (m *Mapper1) WritePRG(addr uint16, data byte) {
	switch {
	case addr >= 0x6000 && addr <= 0x7FFF: // PRG-RAM
		m.sram[addr-0x6000] = data
	case addr >= 0x8000 && addr <= 0xFFFF: // PRG-ROM (registers)
		m.loadRegister(addr, data)
	default:
		log.Printf("[WARN] mapper1: unhandled prg write at %04X", addr)
	}
}

func (m *Mapper1) chrBankIndex() (uint, uint) {
	switch m.chrMode() {
	case 0: // Switch 8 KB at a time.
		return uint(m.chrBank0 & 0xFE), uint(m.chrBank0 | 0x01)
	case 1: // Switch two separate 4 KB banks.
		return uint(m.chrBank0), uint(m.chrBank1)
	default:
		panic(fmt.Sprintf("mapper1: invalid chr mode: %d", m.chrMode()))
	}
}

func (m *Mapper1) chrOffset(idx uint) uint {
	return idx * 0x1000
}

func (m *Mapper1) ReadCHR(addr uint16) byte {
	bank0, bank1 := m.chrBankIndex()
	relAddr := uint(addr % 0x1000)

	switch {
	case addr >= 0x0000 && addr <= 0x0FFF: // CHR-RAM, bank 0
		return m.rom.CHR[m.chrOffset(bank0)+relAddr]
	case addr >= 0x1000 && addr <= 0x1FFF: // CHR-RAM, bank 1
		return m.rom.CHR[m.chrOffset(bank1)+relAddr]
	default:
		log.Printf("[WARN] mapper1: unhandled chr read at %04X", addr)
		return 0
	}
}

func (m *Mapper1) WriteCHR(addr uint16, data byte) {
	if !m.rom.chrRAM {
		log.Printf("[WARN] mapper1: write to read-only chr at %04X", addr)
		return
	}

	bank0, bank1 := m.chrBankIndex()
	relAddr := uint(addr % 0x1000)

	switch {
	case addr >= 0x0000 && addr <= 0x0FFF: // CHR-RAM, bank 0
		m.rom.CHR[m.chrOffset(bank0)+relAddr] = data
	case addr >= 0x1000 && addr <= 0x1FFF: // CHR-RAM, bank 1
		m.rom.CHR[m.chrOffset(bank1)+relAddr] = data
	default:
		log.Printf("mapper1: unhandled chr write at %04X", addr)
	}
}

func (m *Mapper1) SaveState(w *binario.Writer) error {
	return errors.Join(
		m.rom.SaveState(w),
		w.WriteByteSlice(m.sram[:]),
		w.WriteUint8(m.control),
		w.WriteUint8(m.chrBank0),
		w.WriteUint8(m.chrBank1),
		w.WriteUint8(m.prgBank),
		w.WriteUint8(m.shiftRegister),
		w.WriteUint8(m.writeCount),
	)
}

func (m *Mapper1) LoadState(r *binario.Reader) error {
	return errors.Join(
		m.rom.LoadState(r),
		r.ReadByteSliceTo(m.sram[:]),
		r.ReadUint8To(&m.control),
		r.ReadUint8To(&m.chrBank0),
		r.ReadUint8To(&m.chrBank1),
		r.ReadUint8To(&m.prgBank),
		r.ReadUint8To(&m.shiftRegister),
		r.ReadUint8To(&m.writeCount),
	)
}
