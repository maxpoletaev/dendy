package ines

import "fmt"

type Mapper1 struct {
	rom        *ROM
	mirrorMode MirrorMode

	prgBank16Lo int
	prgBank16Hi int
	prgBank32   int

	chrBank4Lo int
	chrBank4Hi int
	chrBank8   int

	loadReg   byte
	ctrlReg   byte
	loadCount uint8
}

func NewMapper1(rom *ROM) *Mapper1 {
	return &Mapper1{
		rom:        rom,
		mirrorMode: rom.MirrorMode,
	}
}

func (m *Mapper1) Reset() {
	m.ctrlReg = 0x1C
	m.loadCount = 0
	m.loadReg = 0

	m.chrBank4Lo = 0
	m.chrBank4Hi = 0
	m.chrBank8 = 0

	m.prgBank16Lo = 0
	m.prgBank16Hi = 0
	m.prgBank32 = m.rom.PRGBanks - 1
}

func (m *Mapper1) isPRG32K() bool {
	return m.ctrlReg&(1<<3) == 0
}

func (m *Mapper1) isCHR8K() bool {
	return m.ctrlReg&(1<<4) == 0
}

func (m *Mapper1) MirrorMode() MirrorMode {
	return m.rom.MirrorMode
}

func (m *Mapper1) ReadPRG(addr uint16) byte {
	switch {
	case addr >= 0x6000 && addr < 0x8000:
		return 0 // SRAM not supported

	case addr >= 0x8000 && addr <= 0xFFFF:
		if m.isPRG32K() {
			return m.rom.PRG[m.prgBank32*0x8000+int(addr&0x7FFF)]
		}

		switch {
		case addr >= 0x8000 && addr <= 0xBFFF:
			return m.rom.PRG[m.prgBank16Lo*0x4000+int(addr&0x3FFF)]
		case addr >= 0xC000 && addr <= 0xFFFF:
			return m.rom.PRG[m.prgBank16Hi*0x4000+int(addr&0x3FFF)]
		}
	}

	panic(fmt.Sprintf("mapper1: unhandled read at 0x%04X", addr))
}

func (m *Mapper1) WritePRG(addr uint16, data byte) {
	if addr >= 0x8000 && addr <= 0xFFFF {
		// If the leftmost bit is set, we need to reset the shift register.
		if data&0x80 != 0 {
			m.loadReg = 0
			m.loadCount = 0
			m.ctrlReg |= 0x0C
			return
		}

		m.loadCount++
		m.loadReg >>= 1
		m.loadReg |= (data & 0x01) << 4

		// 5 bits have been loaded into the shift register. Next, depending on the
		// address of the address, we need to determine which register to update.
		if m.loadCount == 5 {
			m.writeRegister(addr, m.loadReg)
			m.loadCount = 0
			m.loadReg = 0
		}
	}
}

func (m *Mapper1) writeRegister(addr uint16, data byte) {
	switch {
	case addr <= 0x9FFF: // Control register
		m.ctrlReg = data
		m.writeMirrorMode(data)
	case addr <= 0xBFFF:
		if m.isCHR8K() {
			m.chrBank8 = int(data & 0x1E)
		} else {
			m.chrBank4Lo = int(data & 0x1F)
		}
	case addr <= 0xDFFF:
		if !m.isCHR8K() {
			m.chrBank4Hi = int(data & 0x1F)
		}
	case addr <= 0xFFFF:
		prgMode := (data >> 2) & 0x03

		switch prgMode {
		case 0, 1:
			m.prgBank32 = int(data&0x0E) >> 1
		case 2:
			m.prgBank16Lo = 0
			m.prgBank16Hi = int(data & 0x0F)
		case 3:
			m.prgBank16Lo = int(data & 0x0F)
			m.prgBank16Hi = m.rom.PRGBanks - 1
		}
	}
}

func (m *Mapper1) writeMirrorMode(data byte) {
	mode := data & 0x03

	switch mode {
	case 0:
		m.mirrorMode = MirrorSingleLo
	case 1:
		m.mirrorMode = MirrorSingleHi
	case 2:
		m.mirrorMode = MirrorVertical
	case 3:
		m.mirrorMode = MirrorHorizontal
	default:
		panic("mapper1: invalid mirror mode")
	}
}

func (m *Mapper1) ReadCHR(addr uint16) byte {
	if addr >= 0x0000 && addr <= 0x1FFF {
		if m.isCHR8K() {
			return m.rom.CHR[m.chrBank8*0x2000+int(addr&0x1FFF)]
		}

		switch {
		case addr >= 0x0000 && addr <= 0x0FFF:
			return m.rom.CHR[m.chrBank4Lo*0x1000+int(addr&0x0FFF)]
		case addr >= 0x1000 && addr <= 0x1FFF:
			return m.rom.CHR[m.chrBank4Hi*0x1000+int(addr&0x0FFF)]
		}
	}

	panic(fmt.Sprintf("mapper1: unhandled read at 0x%04X", addr))
}

func (m *Mapper1) WriteCHR(addr uint16, data byte) {

}
