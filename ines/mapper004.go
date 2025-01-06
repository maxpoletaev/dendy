package ines

import (
	"errors"
	"fmt"
	"log"

	"github.com/maxpoletaev/dendy/internal/binario"
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
	irqEnable  bool
	irqCounter byte
	irqReload  byte
	irqPending bool
}

func NewMapper4(rom *ROM) *Mapper4 {
	return &Mapper4{
		rom: rom,
	}
}

func (m *Mapper4) Reset() {
	m.mirror = MirrorHorizontal

	m.registers = [8]int{}
	m.chrBank = [8]int{}
	m.prgBank = [4]int{}
	m.targetReg = 0
	m.chrMode = 0
	m.prgMode = 0

	m.irqPending = false
	m.irqEnable = false
	m.irqCounter = 0
	m.irqReload = 0

	m.updateBanks()
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

func (m *Mapper4) ScanlineTick() {
	if m.irqCounter == 0 {
		m.irqCounter = m.irqReload
	} else {
		m.irqCounter--
		if m.irqCounter == 0 {
			m.irqPending = m.irqEnable
		}
	}
}

func (m *Mapper4) PendingIRQ() (v bool) {
	v, m.irqPending = m.irqPending, false
	return v
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
	if !m.rom.chrRAM {
		log.Printf("[WARN] mapper4: write to read-only chr at %04X", addr)
		return
	}

	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		bank := int(addr / 0x0400)
		offset := int(addr % 0x0400)
		m.rom.CHR[m.chrBank[bank]+offset] = data
	default:
		log.Printf("[WARN] mapper4: unhandled chr write at %04X", addr)
	}
}

func (m *Mapper4) SaveState(w *binario.Writer) error {
	err := errors.Join(
		m.rom.SaveState(w),
		w.WriteByteSlice(m.sram[:]),
		w.WriteUint8(m.mirror),
		w.WriteUint8(m.prgMode),
		w.WriteUint8(m.chrMode),
		w.WriteUint8(m.targetReg),
		w.WriteBool(m.irqEnable),
		w.WriteUint8(m.irqCounter),
		w.WriteUint8(m.irqReload),
	)

	if err != nil {
		return err
	}

	for i := range m.registers {
		if err := w.WriteUint32(uint32(m.registers[i])); err != nil {
			return err
		}
	}

	for i := range m.chrBank {
		if err := w.WriteUint32(uint32(m.chrBank[i])); err != nil {
			return err
		}
	}

	for i := range m.prgBank {
		if err := w.WriteUint32(uint32(m.prgBank[i])); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mapper4) LoadState(r *binario.Reader) error {
	err := errors.Join(
		m.rom.LoadState(r),
		r.ReadByteSliceTo(m.sram[:]),
		r.ReadUint8To(&m.mirror),
		r.ReadUint8To(&m.prgMode),
		r.ReadUint8To(&m.chrMode),
		r.ReadUint8To(&m.targetReg),
		r.ReadBoolTo(&m.irqEnable),
		r.ReadUint8To(&m.irqCounter),
		r.ReadUint8To(&m.irqReload),
	)

	for i := range m.registers {
		val, err := r.ReadUint32()
		if err != nil {
			return err
		}

		m.registers[i] = int(val)
	}

	for i := range m.chrBank {
		val, err := r.ReadUint32()
		if err != nil {
			return err
		}

		m.chrBank[i] = int(val)
	}

	for i := range m.prgBank {
		val, err := r.ReadUint32()
		if err != nil {
			return err
		}

		m.prgBank[i] = int(val)
	}

	return err
}
