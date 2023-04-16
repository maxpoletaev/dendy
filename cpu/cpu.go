package cpu

import (
	"fmt"
)

type Byte uint8
type Word uint16
type AddrMode int
type Flags uint8

const (
	FlagCarry      Flags = 1 << 0
	FlagZero             = 1 << 1
	FlagIntDisable       = 1 << 2
	FlagDecimal          = 1 << 3
	FlagBreak            = 1 << 4
	FlagUnused           = 1 << 5
	FlagOverflow         = 1 << 6
	FlagNegative         = 1 << 7
)

const (
	AddrModeImp AddrMode = iota
	AddrModeAcc
	AddrModeImm
	AddrModeZp
	AddrModeZpX
	AddrModeZpY
	AddrModeAbs
	AddrModeAbsX
	AddrModeAbsY
	AddrModeInd
	AddrModeIndX
	AddrModeIndY
	AddrModeRel
)

const (
	VecNMI   Word = 0xFFFA // Non-maskable interrupt vector
	VecReset Word = 0xFFFC // Reset vector
	VecIRQ   Word = 0xFFFE // Interrupt request vector
)

type Memory interface {
	Read(addr Word) Byte
	ReadWord(addr Word) Word
	Write(addr Word, data Byte)
}

type operand struct {
	addr      Word
	val       Byte
	pageCross bool
}

var (
	opcodes map[Byte]instruction
)

func init() {
	opcodes = make(map[Byte]instruction)
	for _, instr := range instrTable {
		opcodes[instr.opcode] = instr
	}
}

type CPU struct {
	X  Byte  // X register
	Y  Byte  // Y register
	A  Byte  // Accumulator
	P  Flags // Status flags
	SP Byte  // Stack pointer
	PC Word  // Program counter

	cycles int
}

func NewCPU() *CPU {
	return &CPU{
		SP: 0xFF,
	}
}

func (cpu *CPU) getFlag(flag Flags) bool {
	return cpu.P&flag != 0
}

func (cpu *CPU) setFlag(flag Flags, value bool) {
	if value {
		cpu.P |= flag
		return
	}

	cpu.P &= ^flag
}

func (cpu *CPU) setZN(value Byte) {
	cpu.setFlag(FlagNegative, value&(1<<7) != 0)
	cpu.setFlag(FlagZero, value == 0)
}

func (cpu *CPU) pushStack(mem Memory, data Byte) {
	mem.Write(0x0100|Word(cpu.SP), data)
	cpu.SP--
}

func (cpu *CPU) popStack(mem Memory) Byte {
	cpu.SP++
	addr := 0x0100 | Word(cpu.SP)

	return mem.Read(addr)
}

func (cpu *CPU) fetchOpcode(mem Memory) Byte {
	opcode := mem.Read(cpu.PC)
	cpu.PC++

	return opcode
}

func (cpu *CPU) fetchOperand(mem Memory, mode AddrMode) operand {
	switch mode {
	case AddrModeImp:
		return operand{}
	case AddrModeAcc:
		return operand{
			val: cpu.A,
		}
	case AddrModeImm:
		val := mem.Read(cpu.PC)
		addr := cpu.PC
		cpu.PC++

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeZp:
		addr := Word(mem.Read(cpu.PC))
		val := mem.Read(addr)
		cpu.PC++

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeZpX:
		addr := Word(mem.Read(cpu.PC)) + Word(cpu.X)
		val := mem.Read(addr)
		cpu.PC++

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeZpY:
		addr := Word(mem.Read(cpu.PC)) + Word(cpu.Y)
		val := mem.Read(addr)
		cpu.PC++

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeAbs:
		lo := Word(mem.Read(cpu.PC))
		hi := Word(mem.Read(cpu.PC + 1))
		cpu.PC += 2

		addr := hi<<8 | lo
		val := mem.Read(addr)

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeAbsX:
		lo := Word(mem.Read(cpu.PC))
		hi := Word(mem.Read(cpu.PC + 1))
		cpu.PC += 2

		addr := hi<<8 | lo
		addrX := addr + Word(cpu.X)
		val := mem.Read(addrX)

		var pageCross bool
		if addrX>>8 != addr>>8 {
			pageCross = true
		}

		return operand{
			addr:      addrX,
			val:       val,
			pageCross: pageCross,
		}
	case AddrModeAbsY:
		lo := Word(mem.Read(cpu.PC))
		hi := Word(mem.Read(cpu.PC + 1))
		cpu.PC += 2

		addr := hi<<8 | lo
		addrY := addr + Word(cpu.Y)
		val := mem.Read(addrY)

		var cross bool
		if addrY>>8 != addr>>8 {
			cross = true
		}

		return operand{
			addr:      addrY,
			val:       val,
			pageCross: cross,
		}
	case AddrModeInd:
		lo := Word(mem.Read(cpu.PC))
		hi := Word(mem.Read(cpu.PC + 1))
		cpu.PC += 2

		ptrAddr := hi<<8 | lo
		lo = Word(mem.Read(ptrAddr))
		hi = Word(mem.Read(ptrAddr + 1))

		// The original 6502 has does not correctly fetch the target address if the indirect vector falls on
		// a page boundary (e.g. $XXFF where XX is any value from $00 to $FF). In this case fetches the LSB
		// from $XXFF as expected but takes the MSB from $XX00.
		if ptrAddr&0xFF == 0xFF {
			hi = Word(mem.Read(ptrAddr & 0xFF00))
		}

		addr := hi<<8 | lo
		val := mem.Read(addr)

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeIndX:
		ptrAddr := Word(mem.Read(cpu.PC)) + Word(cpu.X)
		cpu.PC++

		lo := Word(mem.Read(ptrAddr))
		hi := Word(mem.Read(ptrAddr + 1))

		addr := hi<<8 | lo
		val := mem.Read(addr)

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeIndY:
		ptrAddr := Word(mem.Read(cpu.PC))
		cpu.PC++

		lo := Word(mem.Read(ptrAddr))
		hi := Word(mem.Read(ptrAddr + 1))

		startAddr := hi<<8 | lo
		addrY := startAddr + Word(cpu.Y)
		pageCross := addrY>>8 != startAddr>>8
		val := mem.Read(addrY)

		return operand{
			addr:      addrY,
			val:       val,
			pageCross: pageCross,
		}
	case AddrModeRel:
		rel := Word(mem.Read(cpu.PC))
		cpu.PC++

		if rel&(1<<7) != 0 {
			rel |= 0xFF00
		}

		addr := cpu.PC + rel
		value := mem.Read(addr)
		pageCross := addr&0xFF00 != cpu.PC&0xFF00

		return operand{
			addr:      addr,
			val:       value,
			pageCross: pageCross,
		}
	default:
		panic(fmt.Sprintf("unhandled address mode: %d", mode))
	}
}

func (cpu *CPU) Reset(mem Memory) {
	cpu.PC = mem.ReadWord(VecReset)
	cpu.SP = 0xFF
	cpu.A = 0
	cpu.X = 0
	cpu.Y = 0
	cpu.P = 0

	cpu.cycles = 0
}

func (cpu *CPU) runOpcode(mem Memory, instr instruction, arg operand) {
	switch instr.name {
	case "NOP":
		// do nothing
	case "LDA":
		cpu.lda(mem, arg)
	case "STA":
		cpu.sta(mem, arg)
	case "LDX":
		cpu.ldx(mem, arg)
	case "STX":
		cpu.stx(mem, arg)
	case "LDY":
		cpu.ldy(mem, arg)
	case "STY":
		cpu.sty(mem, arg)
	case "TAX":
		cpu.tax(mem, arg)
	case "TXA":
		cpu.txa(mem, arg)
	case "TAY":
		cpu.tay(mem, arg)
	case "TYA":
		cpu.tya(mem, arg)
	case "TSX":
		cpu.tsx(mem, arg)
	case "TXS":
		cpu.txs(mem, arg)
	case "PHA":
		cpu.pha(mem, arg)
	case "PLA":
		cpu.pla(mem, arg)
	case "PHP":
		cpu.php(mem, arg)
	case "PLP":
		cpu.plp(mem, arg)
	case "ADC":
		cpu.adc(mem, arg)
	case "SBC":
		cpu.sbc(mem, arg)
	case "AND":
		cpu.and(mem, arg)
	case "ORA":
		cpu.ora(mem, arg)
	case "EOR":
		cpu.eor(mem, arg)
	case "CMP":
		cpu.cmp(mem, arg)
	case "CPX":
		cpu.cpx(mem, arg)
	case "CPY":
		cpu.cpy(mem, arg)
	case "INC":
		cpu.inc(mem, arg)
	case "DEC":
		cpu.dec(mem, arg)
	case "INX":
		cpu.inx(mem, arg)
	case "DEX":
		cpu.dex(mem, arg)
	case "INY":
		cpu.iny(mem, arg)
	case "DEY":
		cpu.dey(mem, arg)
	case "ASL":
		cpu.asl(mem, arg)
	case "LSR":
		cpu.lsr(mem, arg)
	case "ROL":
		cpu.rol(mem, arg)
	case "ROR":
		cpu.ror(mem, arg)
	case "BIT":
		cpu.bit(mem, arg)
	case "BCC":
		cpu.bcc(mem, arg)
	case "BCS":
		cpu.bcs(mem, arg)
	case "BEQ":
		cpu.beq(mem, arg)
	case "BMI":
		cpu.bmi(mem, arg)
	case "BNE":
		cpu.bne(mem, arg)
	case "BPL":
		cpu.bpl(mem, arg)
	case "BVC":
		cpu.bvc(mem, arg)
	case "BVS":
		cpu.bvs(mem, arg)
	case "CLC":
		cpu.clc(mem, arg)
	case "CLD":
		cpu.cld(mem, arg)
	case "CLI":
		cpu.cli(mem, arg)
	case "CLV":
		cpu.clv(mem, arg)
	case "SEC":
		cpu.sec(mem, arg)
	case "SED":
		cpu.sed(mem, arg)
	case "SEI":
		cpu.sei(mem, arg)
	case "BRK":
		cpu.brk(mem, arg)
	case "RTI":
		cpu.rti(mem, arg)
	case "JMP":
		cpu.jmp(mem, arg)
	case "JSR":
		cpu.jsr(mem, arg)
	case "RTS":
		cpu.rts(mem, arg)
	default:
		panic(fmt.Sprintf("unhandled instruction: %s", instr.name))
	}
}

// Tick executes a single CPU cycle, returning true if the CPU has finished executing the current instruction.
func (cpu *CPU) Tick(mem Memory) bool {
	if cpu.cycles > 0 {
		cpu.cycles--
		return cpu.cycles == 0
	}

	var (
		opcode = cpu.fetchOpcode(mem)
		instr  instruction
		ok     bool
	)

	if instr, ok = opcodes[opcode]; !ok {
		panic(fmt.Sprintf("invalid opcode: %02X", opcode))
	}

	opr := cpu.fetchOperand(mem, instr.mode)
	cpu.setFlag(FlagUnused, true) // must always be set
	cpu.runOpcode(mem, instr, opr)
	cpu.cycles += instr.cost - 1

	return false
}
