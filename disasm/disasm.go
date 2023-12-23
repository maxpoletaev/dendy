package disasm

import (
	"fmt"
	"strings"

	cpupkg "github.com/maxpoletaev/dendy/cpu"
)

func instrSize(opcode byte) int {
	instr := cpupkg.Opcodes[opcode]
	if instr.Size == 0 {
		return 1
	}

	return instr.Size
}

func readWord(mem cpupkg.Memory, addr uint16) uint16 {
	lo := uint16(mem.Read(addr))
	hi := uint16(mem.Read(addr + 1))
	return hi<<8 | lo
}

// DebugStep returns a string containing the current CPU state and the
// disassembled instruction at the current PC. The format of the string is
// designed to be similar to the output of the Nintendulator NES emulator, for
// easy comparison with the golden log files.
func DebugStep(mem cpupkg.Memory, cpu *cpupkg.CPU) string {
	var b strings.Builder

	opcode := mem.Read(cpu.PC)
	size := instrSize(opcode)

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
	disassemble(&b, mem, cpu.PC)
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
func disassemble(b *strings.Builder, mem cpupkg.Memory, pc uint16) {
	opcode := mem.Read(pc)

	instr := cpupkg.Opcodes[opcode]
	if instr.Size == 0 {
		b.WriteString("???")
		return
	}

	var arg uint16
	if instr.Size == 2 {
		arg = uint16(mem.Read(pc + 1))
	} else if instr.Size == 3 {
		arg = readWord(mem, pc+1)
	}

	b.WriteString(fmt.Sprintf("%s ", instr.Name))

	// Note for future me: You should not try to read a memory values here, because
	// reading from some addresses (e.g. PPU registers) can have side effects.

	switch instr.AddrMode {
	case cpupkg.AddrModeAcc:
		b.WriteString("A")
	case cpupkg.AddrModeImm:
		b.WriteString(fmt.Sprintf("#$%02X", arg))
	case cpupkg.AddrModeZp:
		b.WriteString(fmt.Sprintf("$%02X", arg))
	case cpupkg.AddrModeZpX:
		b.WriteString(fmt.Sprintf("$%02X,X", arg))
	case cpupkg.AddrModeZpY:
		b.WriteString(fmt.Sprintf("$%02X,Y", arg))
	case cpupkg.AddrModeAbs:
		b.WriteString(fmt.Sprintf("$%04X", arg))
	case cpupkg.AddrModeAbsX:
		b.WriteString(fmt.Sprintf("$%04X,X", arg))
	case cpupkg.AddrModeAbsY:
		b.WriteString(fmt.Sprintf("$%04X,Y", arg))
	case cpupkg.AddrModeInd:
		b.WriteString(fmt.Sprintf("($%04X)", arg))
	case cpupkg.AddrModeIndX:
		b.WriteString(fmt.Sprintf("($%02X,X)", arg))
	case cpupkg.AddrModeIndY:
		b.WriteString(fmt.Sprintf("($%02X),Y", arg))
	case cpupkg.AddrModeRel:
		b.WriteString(fmt.Sprintf("$%02X", arg))
	case cpupkg.AddrModeImp:
		// Do nothing.
	}
}
