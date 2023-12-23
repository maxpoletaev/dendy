package cpu

func xnop(cpu *CPU, mem Memory, arg operand) {
	if arg.pageCross {
		cpu.Halt += 1
	}
}

// xsbc is the same as official sbc
func xsbc(cpu *CPU, mem Memory, arg operand) {
	sbc(cpu, mem, arg)
}

// xdcp is dec + cmp
func xdcp(cpu *CPU, mem Memory, arg operand) {
	data := mem.Read(arg.addr) - 1
	mem.Write(arg.addr, data)

	data2 := uint16(cpu.A) - uint16(data)
	cpu.setFlag(flagCarry, data2 < 0x100)
	cpu.setZN(uint8(data2))
}

// xisb is inc + sbc
func xisb(cpu *CPU, mem Memory, arg operand) {
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

// xlax is lda + ldx
func xlax(cpu *CPU, mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	cpu.A, cpu.X = data, data
	cpu.setZN(cpu.X)

	if arg.pageCross {
		cpu.Halt += 1
	}
}

// xrla is rol + and
func xrla(cpu *CPU, mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	carr := cpu.carried()

	cpu.setFlag(flagCarry, data&0x80 != 0)
	data = (data << 1) | carr
	mem.Write(arg.addr, data)
	cpu.A &= data
	cpu.setZN(cpu.A)
}

// xsax is sta + stx
func xsax(cpu *CPU, mem Memory, arg operand) {
	data := cpu.A & cpu.X
	mem.Write(arg.addr, data)
}

// xslo is asl + ora
func xslo(cpu *CPU, mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	cpu.setFlag(flagCarry, data&0x80 != 0)

	data <<= 1
	mem.Write(arg.addr, data)

	cpu.A |= data
	cpu.setZN(cpu.A)
}

// xsre is lsr + eor
func xsre(cpu *CPU, mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	cpu.setFlag(flagCarry, data&0x01 != 0)

	data >>= 1
	mem.Write(arg.addr, data)

	cpu.A ^= data
	cpu.setZN(cpu.A)
}

// xrra is ror + adc
func xrra(cpu *CPU, mem Memory, arg operand) {
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
