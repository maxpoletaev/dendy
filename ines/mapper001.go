package ines

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
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
	m.shiftRegister = 0x10
	m.control = 0x0C
	m.prgBank = 0
	m.chrBank0 = 0
	m.chrBank1 = 0
	m.writeCount = 0
}

func (m *Mapper1) Scanline() TickInfo {
	return TickInfo{}
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
	// 0: switch 32 KB at $8000, ignoring low bit of bank number;
	// 1: fix first bank at $8000 and switch 16 KB bank at $C000;
	// 2: fix last bank at $C000 and switch 16 KB bank at $8000
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

func (m *Mapper1) prgBankIndex() (int, int) {
	switch m.prgMode() {
	case 0, 1: // Switch 32 KB at $8000, ignoring low bit of bank number.
		return int(m.prgBank & 0xFE), int(m.prgBank | 0x01)
	case 2: // Fix first bank at $8000 and switch 16 KB bank at $C000.
		return 0, int(m.prgBank)
	case 3: // Fix last bank at $C000 and switch 16 KB bank at $8000.
		return int(m.prgBank), m.rom.PRGBanks - 1
	default:
		panic(fmt.Sprintf("mapper1: invalid prg mode: %d", m.prgMode()))
	}
}

func (m *Mapper1) prgOffset(idx int) int {
	return idx * 0x4000
}

func (m *Mapper1) ReadPRG(addr uint16) byte {
	bank0, bank1 := m.prgBankIndex()

	switch {
	case addr >= 0x6000 && addr <= 0x7FFF: // PRG-RAM
		return m.sram[addr-0x6000]
	case addr >= 0x8000 && addr <= 0xBFFF: // PRG-ROM, bank 0
		relAddr := int((addr - 0x8000) % 0x4000)
		return m.rom.PRG[m.prgOffset(bank0)+relAddr]
	case addr >= 0xC000 && addr <= 0xFFFF: // PRG-ROM, bank 1
		relAddr := int((addr - 0x8000) % 0x4000)
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

func (m *Mapper1) chrBankIndex() (int, int) {
	switch m.chrMode() {
	case 0: // Switch 8 KB at a time.
		return int(m.chrBank0 & 0xFE), int(m.chrBank0 | 0x01)
	case 1: // Switch two separate 4 KB banks.
		return int(m.chrBank0), int(m.chrBank1)
	default:
		panic(fmt.Sprintf("mapper1: invalid chr mode: %d", m.chrMode()))
	}
}

func (m *Mapper1) chrOffset(idx int) int {
	return idx * 0x1000
}

func (m *Mapper1) ReadCHR(addr uint16) byte {
	bank0, bank1 := m.chrBankIndex()
	relAddr := int(addr % 0x1000)

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
	bank0, bank1 := m.chrBankIndex()
	relAddr := int(addr % 0x1000)

	switch {
	case addr >= 0x0000 && addr <= 0x0FFF: // CHR-RAM, bank 0
		m.rom.CHR[m.chrOffset(bank0)+relAddr] = data
	case addr >= 0x1000 && addr <= 0x1FFF: // CHR-RAM, bank 1
		m.rom.CHR[m.chrOffset(bank1)+relAddr] = data
	default:
		log.Printf("mapper1: unhandled chr write at %04X", addr)
	}
}

func (m *Mapper1) Save(enc *gob.Encoder) error {
	return errors.Join(
		m.rom.Save(enc),
		enc.Encode(m.sram),
		enc.Encode(m.control),
		enc.Encode(m.chrBank0),
		enc.Encode(m.chrBank1),
		enc.Encode(m.prgBank),
		enc.Encode(m.shiftRegister),
		enc.Encode(m.writeCount),
	)
}

func (m *Mapper1) Load(dec *gob.Decoder) error {
	return errors.Join(
		m.rom.Load(dec),
		dec.Decode(&m.sram),
		dec.Decode(&m.control),
		dec.Decode(&m.chrBank0),
		dec.Decode(&m.chrBank1),
		dec.Decode(&m.prgBank),
		dec.Decode(&m.shiftRegister),
		dec.Decode(&m.writeCount),
	)
}
