package ines

// Mapper0 is the simplest mapper. It has no registers, and it only supports
// 16KB or 32KB PRG-ROM banks and 8KB CHR-ROM banks.
//
//	PRG-ROM is mapped to 0x8000-0xFFFF.
//	CHR-ROM is mapped to 0x0000-0x1FFF.
type Mapper0 struct{}

func NewMapper0() *Mapper0 {
	return &Mapper0{}
}

func (m *Mapper0) ReadPRG(cart *Cartridge, addr uint16) byte {
	addr %= uint16(len(cart.prg))
	return cart.prg[addr]
}

func (m *Mapper0) WritePRG(cart *Cartridge, addr uint16, data byte) {
	addr %= uint16(len(cart.prg))
	cart.prg[addr] = data
}

func (m *Mapper0) ReadCHR(cart *Cartridge, addr uint16) byte {
	return cart.chr[addr]
}

func (m *Mapper0) WriteCHR(cart *Cartridge, addr uint16, data byte) {
	cart.chr[addr] = data
}
