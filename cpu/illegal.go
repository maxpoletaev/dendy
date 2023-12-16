package cpu

func (cpu *CPU) executeIllegal(mem Memory, instr *Instruction, arg operand) bool {
	switch instr.ID {
	case XDCP:
		cpu.dcp(mem, arg)
	case XISB:
		cpu.isb(mem, arg)
	case XSLO:
		cpu.slo(mem, arg)
	case XRLA:
		cpu.rla(mem, arg)
	case XSRE:
		cpu.sre(mem, arg)
	case XRRA:
		cpu.rra(mem, arg)
	case XLAX:
		cpu.lax(mem, arg)
	case XSAX:
		cpu.sax(mem, arg)
	case XSBC: // USBC
		cpu.sbc(mem, arg)
	case XNOP:
		cpu.nop(mem, arg)
	default:
		return false
	}

	return true
}

func (cpu *CPU) nop(mem Memory, arg operand) {
	if arg.pageCross {
		cpu.Halt += 1
	}
}

// dcp is dec + cmp
func (cpu *CPU) dcp(mem Memory, arg operand) {
	data := mem.Read(arg.addr) - 1
	mem.Write(arg.addr, data)

	data2 := uint16(cpu.A) - uint16(data)
	cpu.setFlag(flagCarry, data2 < 0x100)
	cpu.setZN(uint8(data2))
}

// isb is inc + sbc
func (cpu *CPU) isb(mem Memory, arg operand) {
	var (
		data = mem.Read(arg.addr) + 1
		a    = uint16(cpu.A)
		b    = uint16(data)
	)

	r := a - b - uint16(1-cpu.carried())
	overflow := (a^b)&0x80 != 0 && (a^r)&0x80 != 0

	mem.Write(arg.addr, data)
	cpu.setFlag(flagCarry, r < 0x100)
	cpu.setFlag(flagOverflow, overflow)
	cpu.A = uint8(r)
	cpu.setZN(cpu.A)
}

// lax is lda + ldx
func (cpu *CPU) lax(mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	cpu.A, cpu.X = data, data
	cpu.setZN(cpu.X)

	if arg.pageCross {
		cpu.Halt += 1
	}
}

// rla is rol + and
func (cpu *CPU) rla(mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	carr := cpu.carried()

	cpu.setFlag(flagCarry, data&0x80 != 0)
	data = (data << 1) | carr
	mem.Write(arg.addr, data)
	cpu.A &= data
	cpu.setZN(cpu.A)
}

// sax is sta + stx
func (cpu *CPU) sax(mem Memory, arg operand) {
	data := cpu.A & cpu.X
	mem.Write(arg.addr, data)
}

// slo is asl + ora
func (cpu *CPU) slo(mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	cpu.setFlag(flagCarry, data&0x80 != 0)

	data <<= 1
	mem.Write(arg.addr, data)

	cpu.A |= data
	cpu.setZN(cpu.A)
}

// sre is lsr + eor
func (cpu *CPU) sre(mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	cpu.setFlag(flagCarry, data&0x01 != 0)

	data >>= 1
	mem.Write(arg.addr, data)

	cpu.A ^= data
	cpu.setZN(cpu.A)
}

// rra is ror + adc
func (cpu *CPU) rra(mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	carr := cpu.carried()

	// ror
	cpu.setFlag(flagCarry, data&0x01 != 0)
	data = data>>1 | carr<<7
	mem.Write(arg.addr, data)

	// adc
	a, b := uint16(cpu.A), uint16(data)
	r := a + b + uint16(cpu.carried())
	overflow := (a^b)&0x80 == 0 && (a^r)&0x80 != 0

	cpu.setFlag(flagOverflow, overflow)
	cpu.setFlag(flagCarry, r > 0xFF)
	cpu.A = uint8(r)
	cpu.setZN(cpu.A)
}
