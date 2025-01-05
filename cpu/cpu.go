package cpu

import (
	"fmt"
)

type Flags = uint8

const (
	flagCarry     Flags = 1 << 0
	flagZero            = 1 << 1
	flagInterrupt       = 1 << 2
	flagDecimal         = 1 << 3
	flagBreak           = 1 << 4
	flagOverflow        = 1 << 6
	flagNegative        = 1 << 7
)

type Interrupt = uint8

const (
	interruptNMI Interrupt = iota + 1
	interruptIRQ
)

const (
	vecNMI   uint16 = 0xFFFA // Non-maskable interrupt vector
	vecReset uint16 = 0xFFFC // Reset vector
	vecIRQ   uint16 = 0xFFFE // Interrupt request vector
)

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

	Cycles uint64 // Number of cycles executed
	Halt   int    // Number of cycles to wait

	interrupt Interrupt
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

// setZN sets the zero and negative flags based on the given value, which is
// assumed to be the result of an operation.
func (cpu *CPU) setZN(value uint8) {
	cpu.setFlag(flagZero, value == 0)
	cpu.setFlag(flagNegative, value&0x80 != 0)
}

// carried returns 1 if the carry flag is set, otherwise 0.
func (cpu *CPU) carried() uint8 {
	if cpu.getFlag(flagCarry) {
		return 1
	}

	return 0
}

// pushByte pushes a byte onto the stack.
func (cpu *CPU) pushByte(mem Memory, data uint8) {
	mem.Write(0x0100|uint16(cpu.SP), data)
	cpu.SP--
}

// popByte pops a byte from the stack.
func (cpu *CPU) popByte(mem Memory) uint8 {
	cpu.SP++
	return mem.Read(0x0100 | uint16(cpu.SP))
}

// pushWord pushes a word onto the stack in little-endian order.
func (cpu *CPU) pushWord(mem Memory, data uint16) {
	cpu.pushByte(mem, uint8(data>>8))
	cpu.pushByte(mem, uint8(data))
}

// popWord pops a word from the stack in little-endian order.
func (cpu *CPU) popWord(mem Memory) uint16 {
	lo := uint16(cpu.popByte(mem))
	hi := uint16(cpu.popByte(mem))

	return hi<<8 | lo
}

// fetchOpcode reads the next opcode from memory and increments the program
// counter.
func (cpu *CPU) fetchOpcode(mem Memory) uint8 {
	opcode := mem.Read(cpu.PC)
	cpu.PC++

	return opcode
}

// Reset resets the CPU to its initial state. To match the behaviour of the real
// CPU, the next 6 cycles are skipped after a reset.
func (cpu *CPU) Reset(mem Memory) {
	cpu.PC = readWord(mem, vecReset)
	cpu.SP = 0xFD
	cpu.P = 0x24
	cpu.A = 0
	cpu.X = 0
	cpu.Y = 0

	cpu.Cycles = 0
	cpu.Halt = 6

	cpu.interrupt = 0
}

func (cpu *CPU) nmi(mem Memory) {
	cpu.pushWord(mem, cpu.PC)
	cpu.pushByte(mem, cpu.P)
	cpu.setFlag(flagInterrupt, true)
	cpu.PC = readWord(mem, vecNMI)
	cpu.Halt += 7
}

// TriggerNMI triggers a non-maskable interrupt on the next CPU cycle.
func (cpu *CPU) TriggerNMI() {
	cpu.interrupt = interruptNMI
}

func (cpu *CPU) irq(mem Memory) {
	cpu.pushWord(mem, cpu.PC)
	cpu.pushByte(mem, cpu.P)
	cpu.setFlag(flagInterrupt, true)
	cpu.PC = readWord(mem, vecIRQ)
	cpu.Halt += 7
}

// TriggerIRQ triggers an interrupt on the next CPU cycle.
// If the interrupt flag is set, the interrupt is ignored.
func (cpu *CPU) TriggerIRQ() {
	if cpu.getFlag(flagInterrupt) {
		return
	}

	cpu.interrupt = interruptIRQ
}

// Tick executes a single CPU cycle, returning true if the CPU has finished
// executing the current instruction.
func (cpu *CPU) Tick(mem Memory) bool {
	cpu.Cycles++

	if cpu.Halt > 0 {
		cpu.Halt--
		return cpu.Halt == 0
	}

	switch cpu.interrupt {
	case interruptIRQ:
		cpu.irq(mem)
		cpu.interrupt = 0
	case interruptNMI:
		cpu.nmi(mem)
		cpu.interrupt = 0
	}

	var (
		opcode = cpu.fetchOpcode(mem)
		instr  = &Opcodes[opcode]
	)

	if instr.Size == 0 {
		panic(fmt.Sprintf("unknown opcode: 0x%02X", opcode))
	}

	arg := cpu.fetchOperand(mem, instr.AddrMode)

	instr.handler(cpu, mem, arg)

	cpu.Halt += instr.Cycles - 1

	return false
}
