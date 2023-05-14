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

// debugStep returns a string containing the current CPU state and the
// disassembled instruction at the current PC. The format of the string is
// designed to be similar to the output of the Nintendulator NES emulator, for
// easy comparison with the golden log files.
func debugStep(mem Memory, cpu *CPU) string {
	var b strings.Builder

	opcode := mem.Read(cpu.PC)
	size := getInstrSize(opcode)

	// PC
	b.WriteString(fmt.Sprintf("%04X", cpu.PC))
	b.WriteString(strings.Repeat(" ", 6-b.Len()))

	// Instruction bytes.
	for i := 0; i < size; i++ {
		addr := cpu.PC + uint16(i)
		b.WriteString(fmt.Sprintf("%02X ", mem.Read(addr)))
	}

	// Pad out to 16 bytes.
	b.WriteString(strings.Repeat(" ", 16-b.Len()))

	// Instruction disassembly.
	b.WriteString(disassemble(mem, cpu.PC))
	b.WriteString(strings.Repeat(" ", 47-b.Len()))

	// CPU state.
	b.WriteString(fmt.Sprintf(" A:%02X", cpu.A))
	b.WriteString(fmt.Sprintf(" X:%02X", cpu.X))
	b.WriteString(fmt.Sprintf(" Y:%02X", cpu.Y))
	b.WriteString(fmt.Sprintf(" P:%02X", cpu.P))
	b.WriteString(fmt.Sprintf(" SP:%02X", cpu.SP))
	b.WriteString(fmt.Sprintf(" CYC:%3d", cpu.Cycles))

	return b.String()
}

// disassemble returns a string containing the disassembled instruction at the
// current PC. In the case of an unknown opcode, it returns "???".
func disassemble(mem Memory, pc uint16) string {
	opcode := mem.Read(pc)
	instr, ok := instructions[opcode]

	if !ok {
		return "???"
	}

	var arg uint16
	if instr.size == 2 {
		arg = uint16(mem.Read(pc + 1))
	} else if instr.size == 3 {
		arg = readWord(mem, pc+1)
	}

	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("%s ", instr.name))

	switch instr.mode {
	case AddrModeAcc:
		b.WriteString("A")
	case AddrModeImm:
		b.WriteString(fmt.Sprintf("#$%02X", arg))
	case AddrModeZp:
		b.WriteString(fmt.Sprintf("$%02X", arg))
	case AddrModeZpX:
		b.WriteString(fmt.Sprintf("$%02X,X", arg))
	case AddrModeZpY:
		b.WriteString(fmt.Sprintf("$%02X,Y", arg))
	case AddrModeAbs:
		b.WriteString(fmt.Sprintf("$%04X", arg))
	case AddrModeAbsX:
		b.WriteString(fmt.Sprintf("$%04X,X", arg))
	case AddrModeAbsY:
		b.WriteString(fmt.Sprintf("$%04X,Y", arg))
	case AddrModeInd:
		b.WriteString(fmt.Sprintf("($%04X)", arg))
	case AddrModeIndX:
		b.WriteString(fmt.Sprintf("($%02X,X)", arg))
	case AddrModeIndY:
		b.WriteString(fmt.Sprintf("($%02X),Y", arg))
	case AddrModeRel:
		b.WriteString(fmt.Sprintf("$%02X", arg))
	}

	return b.String()
}
