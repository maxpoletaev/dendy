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
	b.WriteString(strings.Repeat(" ", 6-b.Len()))

	for i := 0; i < size; i++ {
		addr := cpu.PC + uint16(i)
		b.WriteString(fmt.Sprintf("%02X ", mem.Read(addr)))
	}

	b.WriteString(strings.Repeat(" ", 16-b.Len()))

	b.WriteString(disassemble(cpu, mem))
	b.WriteString(strings.Repeat(" ", 47-b.Len()))

	b.WriteString(fmt.Sprintf(" A:%02X", cpu.A))
	b.WriteString(fmt.Sprintf(" X:%02X", cpu.X))
	b.WriteString(fmt.Sprintf(" Y:%02X", cpu.Y))
	b.WriteString(fmt.Sprintf(" P:%02X", cpu.P))
	b.WriteString(fmt.Sprintf(" SP:%02X", cpu.SP))
	b.WriteString(fmt.Sprintf(" CYC:%3d", cpu.Cycles))

	return b.String()
}

func disassemble(cpu *CPU, mem Memory) string {
	var (
		b     strings.Builder
		instr instrInfo
		ok    bool
	)

	var (
		pc     = cpu.PC
		opcode = mem.Read(pc)
	)

	if instr, ok = instructions[opcode]; !ok {
		b.WriteString("???")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("%s ", instr.name))

	switch instr.mode {
	case AddrModeImm:
		b.WriteString(fmt.Sprintf("#$%02X", mem.Read(pc+1)))

	case AddrModeZp:
		b.WriteString(fmt.Sprintf("$%02X", mem.Read(pc+1)))

	case AddrModeZpX:
		b.WriteString(fmt.Sprintf("$%02X,X", mem.Read(pc+1)))

	case AddrModeZpY:
		b.WriteString(fmt.Sprintf("$%02X,Y", mem.Read(pc+1)))

	case AddrModeAbs:
		arg := cpu.readWord(mem, pc+1)
		b.WriteString(fmt.Sprintf("$%04X", arg))

	case AddrModeAbsX:
		arg := cpu.readWord(mem, pc+1)
		b.WriteString(fmt.Sprintf("$%04X,X", arg))

	case AddrModeAbsY:
		arg := cpu.readWord(mem, pc+1)
		b.WriteString(fmt.Sprintf("$%04X,Y", arg))

	case AddrModeInd:
		arg := cpu.readWord(mem, pc+1)
		b.WriteString(fmt.Sprintf("($%04X)", arg))

	case AddrModeIndX:
		arg := mem.Read(pc + 1)
		b.WriteString(fmt.Sprintf("($%02X,X)", arg))

	case AddrModeIndY:
		arg := mem.Read(pc + 1)
		b.WriteString(fmt.Sprintf("($%02X),Y", arg))

	case AddrModeRel:
		arg := mem.Read(pc + 1)
		b.WriteString(fmt.Sprintf("$%02X", arg))
	}

	return b.String()
}
