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
	AddrModeImp AddrMode = iota + 1
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
	addr      uint16
	val       uint8
	pageCross bool
}

type CPU struct {
	X  uint8  // X register
	Y  uint8  // Y register
	A  uint8  // Accumulator
	P  Flags  // Status flags
	SP uint8  // Stack pointer
	PC uint16 // Program counter

	Halt         int    // Number of cycles to halt
	AllowIllegal bool   // Allow invalid opcodes
	EnableDisasm bool   // Enable disassembler
	Cycles       uint64 // Number of cycles executed
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

func (cpu *CPU) fetchOpcode(mem Memory) uint8 {
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
		addr := uint16(mem.Read(cpu.PC))
		val := mem.Read(addr)
		cpu.PC++

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeZpX:
		addr := uint16(mem.Read(cpu.PC)) + uint16(cpu.X)
		val := mem.Read(addr)
		cpu.PC++

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeZpY:
		addr := uint16(mem.Read(cpu.PC)) + uint16(cpu.Y)
		val := mem.Read(addr)
		cpu.PC++

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeAbs:
		lo := uint16(mem.Read(cpu.PC))
		hi := uint16(mem.Read(cpu.PC + 1))
		cpu.PC += 2

		addr := hi<<8 | lo
		val := mem.Read(addr)

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeAbsX:
		lo := uint16(mem.Read(cpu.PC))
		hi := uint16(mem.Read(cpu.PC + 1))
		cpu.PC += 2

		addr := hi<<8 | lo
		addrX := addr + uint16(cpu.X)
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
		lo := uint16(mem.Read(cpu.PC))
		hi := uint16(mem.Read(cpu.PC + 1))
		cpu.PC += 2

		addr := hi<<8 | lo
		addrY := addr + uint16(cpu.Y)
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
		lo := uint16(mem.Read(cpu.PC))
		hi := uint16(mem.Read(cpu.PC + 1))
		cpu.PC += 2

		ptrAddr := hi<<8 | lo
		lo = uint16(mem.Read(ptrAddr))
		hi = uint16(mem.Read(ptrAddr + 1))

		// The original 6502 has does not correctly fetch the target address if the indirect vector falls on
		// a page boundary (e.g. $XXFF where XX is any value from $00 to $FF). In this case fetches the LSB
		// from $XXFF as expected but takes the MSB from $XX00.
		if ptrAddr&0xFF == 0xFF {
			hi = uint16(mem.Read(ptrAddr & 0xFF00))
		}

		addr := hi<<8 | lo
		val := mem.Read(addr)

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeIndX:
		ptrAddr := uint16(mem.Read(cpu.PC)) + uint16(cpu.X)
		cpu.PC++

		lo := uint16(mem.Read(ptrAddr))
		hi := uint16(mem.Read(ptrAddr + 1))

		addr := hi<<8 | lo
		val := mem.Read(addr)

		return operand{
			addr: addr,
			val:  val,
		}
	case AddrModeIndY:
		ptrAddr := uint16(mem.Read(cpu.PC))
		cpu.PC++

		lo := uint16(mem.Read(ptrAddr))
		hi := uint16(mem.Read(ptrAddr + 1))

		startAddr := hi<<8 | lo
		addrY := startAddr + uint16(cpu.Y)
		pageCross := addrY>>8 != startAddr>>8
		val := mem.Read(addrY)

		return operand{
			addr:      addrY,
			val:       val,
			pageCross: pageCross,
		}
	case AddrModeRel:
		rel := uint16(mem.Read(cpu.PC))
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
		panic(fmt.Sprintf("invalid addressing mode: %d", mode))
	}
}

func (cpu *CPU) Reset(mem Memory) {
	cpu.PC = cpu.readWord(mem, VecReset)
	cpu.SP = 0xFF
	cpu.A = 0
	cpu.X = 0
	cpu.Y = 0
	cpu.P = 0

	cpu.Halt = 0
}

func (cpu *CPU) execute(mem Memory, instr instrInfo, arg operand) {
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
	case "???":
		if !cpu.AllowIllegal {
			panic(fmt.Sprintf("illegal instruction: %02X", instr.opcode))
		}
	default:
		panic(fmt.Sprintf("unhandled instruction: %s (%02X)", instr.name, instr.opcode))
	}
}

func (cpu *CPU) Interrupt(mem Memory) {
	cpu.pushWord(mem, cpu.PC)
	cpu.pushByte(mem, uint8(cpu.P))
	cpu.setFlag(FlagBreak, false)
	cpu.setFlag(FlagUnused, true)
	cpu.setFlag(FlagInterrupt, true)
	cpu.PC = cpu.readWord(mem, VecNMI)
}

// Tick executes a single CPU cycle, returning true if the CPU has finished executing the current instruction.
func (cpu *CPU) Tick(mem Memory) bool {
	cpu.Cycles++

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
	cpu.setFlag(FlagUnused, true) // must always be set
	cpu.execute(mem, instr, opr)
	cpu.Halt += instr.cost - 1

	return false
}
