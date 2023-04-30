package cpu

import (
	"fmt"
	"strings"
)

func getInstrSize(opcode byte) int {
	if instr, ok := instructions[opcode]; ok {
		return instr.size
	}

	return 1
}

func debugStep(mem Memory, cpu *CPU) string {
	var b strings.Builder

	opcode := mem.Read(cpu.PC)
	size := getInstrSize(opcode)

	b.WriteString(fmt.Sprintf("%04X", cpu.PC))
	b.WriteString(strings.Repeat(" ", 8-b.Len()))

	for i := 0; i < size; i++ {
		addr := cpu.PC + uint16(i)
		b.WriteString(fmt.Sprintf("%02X ", mem.Read(addr)))
	}

	b.WriteString(strings.Repeat(" ", 20-b.Len()))

	b.WriteString(disassemble(mem, cpu.PC))
	b.WriteString(strings.Repeat(" ", 40-b.Len()))

	b.WriteString(fmt.Sprintf(" A:%02X", cpu.A))
	b.WriteString(fmt.Sprintf(" X:%02X", cpu.X))
	b.WriteString(fmt.Sprintf(" Y:%02X", cpu.Y))
	b.WriteString(fmt.Sprintf(" P:%08b", cpu.P))
	b.WriteString(fmt.Sprintf(" SP:%02X", cpu.SP))
	b.WriteString(fmt.Sprintf(" CYC:%3d", cpu.Cycles))

	return b.String()
}

func disassemble(mem Memory, addr uint16) string {
	var (
		b     strings.Builder
		instr instrInfo
		ok    bool
	)

	opcode := mem.Read(addr)

	if instr, ok = instructions[opcode]; !ok {
		b.WriteString("???")
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
		lo := uint16(mem.Read(addr + 1))
		hi := uint16(mem.Read(addr + 2))
		b.WriteString(fmt.Sprintf("$%04X", hi<<8|lo))
	case AddrModeAbsX:
		lo := uint16(mem.Read(addr + 1))
		hi := uint16(mem.Read(addr + 2))
		b.WriteString(fmt.Sprintf("$%04X,X", hi<<8|lo))
	case AddrModeAbsY:
		lo := uint16(mem.Read(addr + 1))
		hi := uint16(mem.Read(addr + 2))
		b.WriteString(fmt.Sprintf("$%04X,Y", hi<<8|lo))
	case AddrModeInd:
		lo := uint16(mem.Read(addr + 1))
		hi := uint16(mem.Read(addr + 2))
		b.WriteString(fmt.Sprintf("($%04X)", hi<<8|lo))
	case AddrModeIndX:
		b.WriteString(fmt.Sprintf("($%02X,X)", mem.Read(addr+1)))
	case AddrModeIndY:
		b.WriteString(fmt.Sprintf("($%02X),Y", mem.Read(addr+1)))
	case AddrModeRel:
		relAddr := mem.Read(addr + 1)
		targetAddr := addr + 2 + uint16(relAddr)
		b.WriteString(fmt.Sprintf("$%02X", relAddr))
		b.WriteString(fmt.Sprintf(" [$%04X]", targetAddr))
	}

	return b.String()
}
