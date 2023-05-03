package cpu

import (
	"fmt"
)

type (
	Flags    uint8
	AddrMode int
)

const (
	FlagCarry     Flags = 1 << 0
	FlagZero            = 1 << 1
	FlagInterrupt       = 1 << 2
	FlagDecimal         = 1 << 3
	FlagBreak           = 1 << 4
	FlagUnused          = 1 << 5
	FlagOverflow        = 1 << 6
	FlagNegative        = 1 << 7
)

const (
	VecNMI   uint16 = 0xFFFA // Non-maskable interrupt vector
	VecReset uint16 = 0xFFFC // Reset vector
	VecIRQ   uint16 = 0xFFFE // Interrupt request vector
)

type instrInfo struct {
	name   string
	opcode uint8
	mode   AddrMode
	size   int
	cost   int
}

type Memory interface {
	Read(addr uint16) uint8
	Write(addr uint16, data uint8)
}

type operand struct {
	mode      AddrMode
	addr      uint16
	pageCross bool
}

type CPU struct {
	X  uint8  // X register
	Y  uint8  // Y register
	A  uint8  // Accumulator
	P  Flags  // Status flags
	SP uint8  // Stack pointer
	PC uint16 // Program counter

	EnableDisasm bool   // Enable disassembler
	AllowIllegal bool   // Handle illegal opcodes
	Cycles       uint64 // Number of cycles executed
	Halt         int    // Number of cycles to halt

	interrupt func(mem Memory)
}

func New() *CPU {
	return &CPU{}
}

func (cpu *CPU) getFlag(flag Flags) bool {
	return cpu.P&flag != 0
}

func (cpu *CPU) setFlag(flag Flags, value bool) {
	if value {
		cpu.P |= flag
		return
	}

	cpu.P &= 0xFF - flag
}

func (cpu *CPU) setZN(value uint8) {
	cpu.setFlag(FlagZero, value == 0)
	cpu.setFlag(FlagNegative, value&0x80 != 0)
}

func (cpu *CPU) carried() uint8 {
	if cpu.getFlag(FlagCarry) {
		return 1
	}

	return 0
}

func (cpu *CPU) pushByte(mem Memory, data uint8) {
	mem.Write(0x0100|uint16(cpu.SP), data)
	cpu.SP--
}

func (cpu *CPU) popByte(mem Memory) uint8 {
	cpu.SP++
	return mem.Read(0x0100 | uint16(cpu.SP))
}

func (cpu *CPU) pushWord(mem Memory, data uint16) {
	cpu.pushByte(mem, uint8(data>>8))
	cpu.pushByte(mem, uint8(data))
}

func (cpu *CPU) popWord(mem Memory) uint16 {
	lo := uint16(cpu.popByte(mem))
	hi := uint16(cpu.popByte(mem))

	return hi<<8 | lo
}

func (cpu *CPU) readWord(mem Memory, addr uint16) uint16 {
	lo := uint16(mem.Read(addr))
	hi := uint16(mem.Read(addr + 1))
	return hi<<8 | lo
}

func (cpu *CPU) readWordBug(mem Memory, addr uint16) uint16 {
	addr2 := addr&0xFF00 | uint16(uint8(addr)+1)

	lo := uint16(mem.Read(addr))
	hi := uint16(mem.Read(addr2))

	return hi<<8 | lo
}

func (cpu *CPU) fetchOpcode(mem Memory) uint8 {
	opcode := mem.Read(cpu.PC)
	cpu.PC++

	return opcode
}

func (cpu *CPU) Reset(mem Memory) {
	cpu.PC = cpu.readWord(mem, VecReset)
	cpu.SP = 0xFD
	cpu.P = 0x24
	cpu.A = 0
	cpu.X = 0
	cpu.Y = 0

	cpu.Halt = 0
}

func (cpu *CPU) TriggerNMI() {
	cpu.interrupt = func(mem Memory) {
		cpu.pushWord(mem, cpu.PC)
		cpu.pushByte(mem, uint8(cpu.P))
		cpu.setFlag(FlagInterrupt, true)
		cpu.PC = cpu.readWord(mem, VecNMI)
		cpu.Halt += 7
	}
}

func (cpu *CPU) TriggerIRQ() {
	if cpu.getFlag(FlagInterrupt) {
		return
	}

	cpu.interrupt = func(mem Memory) {
		cpu.pushWord(mem, cpu.PC)
		cpu.pushByte(mem, uint8(cpu.P))
		cpu.setFlag(FlagInterrupt, true)
		cpu.PC = cpu.readWord(mem, VecIRQ)
		cpu.Halt += 7
	}
}

// Tick executes a single CPU cycle, returning true if the CPU has finished executing the current instruction.
func (cpu *CPU) Tick(mem Memory) bool {
	cpu.Cycles++

	if cpu.interrupt != nil {
		cpu.interrupt(mem)
		cpu.interrupt = nil
		return false
	}

	if cpu.Halt > 0 {
		cpu.Halt--
		return cpu.Halt == 0
	}

	if cpu.EnableDisasm {
		fmt.Println(debugStep(mem, cpu))
	}

	var (
		opcode = cpu.fetchOpcode(mem)
		instr  instrInfo
		ok     bool
	)

	if instr, ok = instructions[opcode]; !ok {
		panic(fmt.Sprintf("unknown opcode: %02X", opcode))
	}

	opr := cpu.fetchOperand(mem, instr.mode)
	ok = cpu.execute(mem, instr, opr)
	if !ok && cpu.AllowIllegal {
		ok = cpu.executeIllegal(mem, instr, opr)
	}

	if !ok {
		panic(fmt.Sprintf("invalid instruction: %s", instr.name))
	}

	cpu.Halt += instr.cost - 1

	return false
}
