package cpu

import (
	"fmt"
	"strings"
)

func disassemble(mem Memory, addr Word) string {
	var b strings.Builder

	opcode := mem.Read(addr)

	instr, ok := opcodes[opcode]
	if !ok {
		b.WriteString(fmt.Sprintf("??? (0x%02X)", opcode))
		return b.String()
	}

	b.WriteString(fmt.Sprintf("%s ", instr.name))

	switch instr.mode {
	case AddrModeImm:
		b.WriteString(fmt.Sprintf("#$%02X", mem.Read(addr+1)))
	case AddrModeZp:
		b.WriteString(fmt.Sprintf("$%02X", mem.Read(addr+1)))
	case AddrModeZpX:
		b.WriteString(fmt.Sprintf("$%02X,X", mem.Read(addr+1)))
	case AddrModeZpY:
		b.WriteString(fmt.Sprintf("$%02X,Y", mem.Read(addr+1)))
	case AddrModeAbs:
		lo := Word(mem.Read(addr + 1))
		hi := Word(mem.Read(addr + 2))
		b.WriteString(fmt.Sprintf("$%04X", hi<<8|lo))
	case AddrModeAbsX:
		lo := Word(mem.Read(addr + 1))
		hi := Word(mem.Read(addr + 2))
		b.WriteString(fmt.Sprintf("$%04X,X", hi<<8|lo))
	case AddrModeAbsY:
		lo := Word(mem.Read(addr + 1))
		hi := Word(mem.Read(addr + 2))
		b.WriteString(fmt.Sprintf("$%04X,Y", hi<<8|lo))
	case AddrModeInd:
		lo := Word(mem.Read(addr + 1))
		hi := Word(mem.Read(addr + 2))
		b.WriteString(fmt.Sprintf("($%04X)", hi<<8|lo))
	case AddrModeIndX:
		b.WriteString(fmt.Sprintf("($%02X,X)", mem.Read(addr+1)))
	case AddrModeIndY:
		b.WriteString(fmt.Sprintf("($%02X),Y", mem.Read(addr+1)))
	case AddrModeRel:
		b.WriteString(fmt.Sprintf("$%02X", mem.Read(addr+1)))
	}

	return b.String()
}
