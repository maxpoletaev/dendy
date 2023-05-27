package cpu

type Memory interface {
	Read(addr uint16) uint8
	Write(addr uint16, data uint8)
}

// readWord reads a word from memory in little-endian order.
func readWord(mem Memory, addr uint16) uint16 {
	lo := uint16(mem.Read(addr))
	hi := uint16(mem.Read(addr + 1))
	return hi<<8 | lo
}

// readWordBug reads a word from memory, simulating the 6502 bug where the high
// byte is read from the same page as the address, but the low byte is read from
// the next page.
func readWordBug(mem Memory, addr uint16) uint16 {
	addr2 := addr&0xFF00 | uint16(uint8(addr)+1)

	lo := uint16(mem.Read(addr))
	hi := uint16(mem.Read(addr2))

	return hi<<8 | lo
}
