package ines

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
)

// Mapper4 implements the MMC3 mapper.
// https://wiki.nesdev.com/w/index.php/MMC3
type Mapper4 struct {
	rom        *ROM
	sram       [0x2000]byte
	mirror     MirrorMode
	chrBank    [8]int
	prgBank    [4]int
	registers  [8]int
	targetReg  byte
	chrMode    byte
	prgMode    byte
	irqCounter byte
	irqReload  byte
	irqEnable  bool
}

func NewMapper4(rom *ROM) *Mapper4 {
	return &Mapper4{
		rom: rom,
	}
}

func (m *Mapper4) prgOffset(idx int) int {
	if idx < 0 {
		idx = m.rom.PRGBanks*2 + idx
	}

	idx %= len(m.rom.PRG) / 0x2000
	return idx * 0x2000
}

func (m *Mapper4) chrOffset(idx int) int {
	idx %= len(m.rom.CHR) / 0x0400
	return idx * 0x0400
}

func (m *Mapper4) updateBanks() {
	switch m.prgMode {
	case 0:
		m.prgBank[0] = m.prgOffset(m.registers[6])
		m.prgBank[1] = m.prgOffset(m.registers[7])
		m.prgBank[2] = m.prgOffset(-2)
		m.prgBank[3] = m.prgOffset(-1)
	case 1:
		m.prgBank[0] = m.prgOffset(-2)
		m.prgBank[1] = m.prgOffset(m.registers[7])
		m.prgBank[2] = m.prgOffset(m.registers[6])
		m.prgBank[3] = m.prgOffset(-1)
	default:
		panic(fmt.Sprintf("mapper4: invalid prg mode %d", m.prgMode))
	}

	switch m.chrMode {
	case 0:
		m.chrBank[0] = m.chrOffset(m.registers[0] & 0xFE)
		m.chrBank[1] = m.chrOffset(m.registers[0] | 0x01)
		m.chrBank[2] = m.chrOffset(m.registers[1] & 0xFE)
		m.chrBank[3] = m.chrOffset(m.registers[1] | 0x01)
		m.chrBank[4] = m.chrOffset(m.registers[2])
		m.chrBank[5] = m.chrOffset(m.registers[3])
		m.chrBank[6] = m.chrOffset(m.registers[4])
		m.chrBank[7] = m.chrOffset(m.registers[5])
	case 1:
		m.chrBank[0] = m.chrOffset(m.registers[2])
		m.chrBank[1] = m.chrOffset(m.registers[3])
		m.chrBank[2] = m.chrOffset(m.registers[4])
		m.chrBank[3] = m.chrOffset(m.registers[5])
		m.chrBank[4] = m.chrOffset(m.registers[0] & 0xFE)
		m.chrBank[5] = m.chrOffset(m.registers[0] | 0x01)
		m.chrBank[6] = m.chrOffset(m.registers[1] & 0xFE)
		m.chrBank[7] = m.chrOffset(m.registers[1] | 0x01)
	default:
		panic(fmt.Sprintf("mapper4: invalid chr mode %d", m.chrMode))
	}
}

func (m *Mapper4) writeMirror(data byte) {
	switch data & 1 {
	case 0:
		m.mirror = MirrorVertical
	case 1:
		m.mirror = MirrorHorizontal
	}
}

func (m *Mapper4) writeRegister(addr uint16, data byte) {
	switch {
	case addr >= 0x8000 && addr <= 0x9FFF && addr%2 == 0: // bank select
		m.prgMode = (data >> 6) & 1
		m.chrMode = (data >> 7) & 1
		m.targetReg = data & 7
		m.updateBanks()
	case addr >= 0x8000 && addr <= 0x9FFF && addr%2 == 1: // bank data
		m.registers[m.targetReg] = int(data)
		m.updateBanks()
	case addr >= 0xA000 && addr <= 0xBFFF && addr%2 == 0: // mirroring
		m.writeMirror(data)
	case addr >= 0xA000 && addr <= 0xBFFF && addr%2 == 1: // prg ram protect
		// noop
	case addr >= 0xC000 && addr <= 0xDFFF && addr%2 == 0: // irq latch
		m.irqReload = data
	case addr >= 0xC000 && addr <= 0xDFFF && addr%2 == 1: // irq reload
		m.irqCounter = 0
	case addr >= 0xE000 && addr <= 0xFFFF && addr%2 == 0: // irq disable
		m.irqEnable = false
	case addr >= 0xE000 && addr <= 0xFFFF && addr%2 == 1: // irq enable
		m.irqEnable = true
	default:
		log.Printf("mapper4: invalid register write at %04X: %02X", addr, data)
	}
}

func (m *Mapper4) Reset() {
	m.mirror = MirrorHorizontal
	m.registers = [8]int{}
	m.chrBank = [8]int{}
	m.prgMode = 0
	m.chrMode = 0

	m.prgBank[0] = 0 * 0x2000
	m.prgBank[1] = 1 * 0x2000
	m.prgBank[2] = m.prgOffset(-2)
	m.prgBank[3] = m.prgOffset(-1)
}

func (m *Mapper4) Scanline() (t TickInfo) {
	if m.irqCounter == 0 {
		m.irqCounter = m.irqReload
	} else {
		m.irqCounter--
		if m.irqCounter == 0 {
			t.RequestIRQ = m.irqEnable
		}
	}

	return
}

func (m *Mapper4) MirrorMode() MirrorMode {
	return m.mirror
}

func (m *Mapper4) ReadPRG(addr uint16) byte {
	switch {
	case addr >= 0x6000 && addr <= 0x7FFF:
		return m.sram[addr-0x6000]
	case addr >= 0x8000 && addr <= 0xFFFF:
		bank := (addr - 0x8000) / 0x2000
		offset := int(addr-0x8000) % 0x2000
		return m.rom.PRG[m.prgBank[bank]+offset]
	default:
		log.Printf("[WARN] mapper4: unhandled prg read at %04X", addr)
		return 0
	}
}

func (m *Mapper4) WritePRG(addr uint16, data byte) {
	switch {
	case addr >= 0x6000 && addr <= 0x7FFF:
		m.sram[addr-0x6000] = data
	case addr >= 0x8000 && addr <= 0xFFFF:
		m.writeRegister(addr, data)
	default:
		log.Printf("[WARN] mapper4: unhandled prg write at %04X", addr)
	}
}

func (m *Mapper4) ReadCHR(addr uint16) byte {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		bank := int(addr / 0x0400)
		offset := int(addr % 0x0400)
		return m.rom.CHR[m.chrBank[bank]+offset]
	default:
		log.Printf("[WARN] mapper4: invalid chr read at %04X", addr)
		return 0

	}
}

func (m *Mapper4) WriteCHR(addr uint16, data byte) {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		bank := int(addr / 0x0400)
		offset := int(addr % 0x0400)
		m.rom.CHR[m.chrBank[bank]+offset] = data
	default:
		log.Printf("[WARN] mapper4: unhandled chr write at %04X", addr)
	}
}

func (m *Mapper4) Save(enc *gob.Encoder) error {
	return errors.Join(
		m.rom.SaveCRC(enc),
		enc.Encode(m.sram),
		enc.Encode(m.mirror),
		enc.Encode(m.prgMode),
		enc.Encode(m.chrMode),
		enc.Encode(m.targetReg),
		enc.Encode(m.registers),
		enc.Encode(m.chrBank),
		enc.Encode(m.prgBank),
		enc.Encode(m.irqEnable),
		enc.Encode(m.irqCounter),
		enc.Encode(m.irqReload),
	)
}

func (m *Mapper4) Load(dec *gob.Decoder) error {
	return errors.Join(
		m.rom.LoadCRC(dec),
		dec.Decode(&m.sram),
		dec.Decode(&m.mirror),
		dec.Decode(&m.prgMode),
		dec.Decode(&m.chrMode),
		dec.Decode(&m.targetReg),
		dec.Decode(&m.registers),
		dec.Decode(&m.chrBank),
		dec.Decode(&m.prgBank),
		dec.Decode(&m.irqEnable),
		dec.Decode(&m.irqCounter),
		dec.Decode(&m.irqReload),
	)
}
