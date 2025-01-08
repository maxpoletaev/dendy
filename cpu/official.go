package cpu

func nop(cpu *CPU, mem Memory, arg operand) {
	// do nothing
}

// lda loads the accumulator with a value from memory.
func lda(cpu *CPU, mem Memory, arg operand) {
	cpu.A = mem.Read(arg.addr)
	cpu.setZN(cpu.A)

	if arg.pageCross {
		cpu.Halt++
	}
}

// sta stores the accumulator in memory.
func sta(cpu *CPU, mem Memory, arg operand) {
	mem.Write(arg.addr, cpu.A)
}

// ldx loads the X register with a value from memory.
func ldx(cpu *CPU, mem Memory, arg operand) {
	cpu.X = mem.Read(arg.addr)
	cpu.setZN(cpu.X)

	if arg.pageCross {
		cpu.Halt++
	}
}

// stx stores the X register in memory.
func stx(cpu *CPU, mem Memory, arg operand) {
	mem.Write(arg.addr, cpu.X)
}

// ldy loads the Y register with a value from memory.
func ldy(cpu *CPU, mem Memory, arg operand) {
	cpu.Y = mem.Read(arg.addr)
	cpu.setZN(cpu.Y)

	if arg.pageCross {
		cpu.Halt++
	}
}

// sty stores the Y register in memory.
func sty(cpu *CPU, mem Memory, arg operand) {
	mem.Write(arg.addr, cpu.Y)
}

// tax transfers the accumulator to the X register.
func tax(cpu *CPU, mem Memory, arg operand) {
	cpu.X = cpu.A
	cpu.setZN(cpu.X)
}

// txa transfers the X register to the accumulator.
func txa(cpu *CPU, mem Memory, arg operand) {
	cpu.A = cpu.X
	cpu.setZN(cpu.A)
}

// tay transfers the accumulator to the Y register.
func tay(cpu *CPU, mem Memory, arg operand) {
	cpu.Y = cpu.A
	cpu.setZN(cpu.Y)
}

// tya transfers the Y register to the accumulator.
func tya(cpu *CPU, mem Memory, arg operand) {
	cpu.A = cpu.Y
	cpu.setZN(cpu.A)
}

// tsx transfers the stack pointer to the X register.
func tsx(cpu *CPU, mem Memory, arg operand) {
	cpu.X = cpu.SP
	cpu.setZN(cpu.X)
}

// txs transfers the X register to the stack pointer.
func txs(cpu *CPU, mem Memory, arg operand) {
	cpu.SP = cpu.X
}

// pha pushes the accumulator onto the stack.
func pha(cpu *CPU, mem Memory, arg operand) {
	cpu.pushByte(mem, cpu.A)
}

// pla pops a value from the stack into the accumulator.
func pla(cpu *CPU, mem Memory, arg operand) {
	cpu.A = cpu.popByte(mem)
	cpu.setZN(cpu.A)
}

// php pushes the processor status onto the stack.
func php(cpu *CPU, mem Memory, arg operand) {
	cpu.pushByte(mem, cpu.P|0x30)
}

// plp pops a value from the stack into the processor status.
func plp(cpu *CPU, mem Memory, arg operand) {
	cpu.P = cpu.popByte(mem)
}

// inc increments a value in memory.
func inc(cpu *CPU, mem Memory, arg operand) {
	val := mem.Read(arg.addr) + 1
	mem.Write(arg.addr, val)
	cpu.setZN(val)
}

// inx increments the X register.
func inx(cpu *CPU, mem Memory, arg operand) {
	cpu.X++
	cpu.setZN(cpu.X)
}

// iny increments the Y register.
func iny(cpu *CPU, mem Memory, arg operand) {
	cpu.Y++
	cpu.setZN(cpu.Y)
}

// dec decrements a value in memory.
func dec(cpu *CPU, mem Memory, arg operand) {
	data := mem.Read(arg.addr) - 1
	mem.Write(arg.addr, data)
	cpu.setZN(data)
}

// dex decrements the X register.
func dex(cpu *CPU, mem Memory, arg operand) {
	cpu.X--
	cpu.setZN(cpu.X)
}

// dey decrements the Y register.
func dey(cpu *CPU, mem Memory, arg operand) {
	cpu.Y--
	cpu.setZN(cpu.Y)
}

// adc adds a value from memory to the accumulator with carry. The carry flag is
// set if the result is greater than 255. The overflow flag is set if the result
// is greater than 127 or less than -128 (incorrect sign bit).
func adc(cpu *CPU, mem Memory, arg operand) {
	var (
		a = uint16(cpu.A)
		b = uint16(mem.Read(arg.addr))
	)

	r := a + b + uint16(cpu.carried())
	overflow := (a^b)&0x80 == 0 && (a^r)&0x80 != 0

	cpu.setFlag(flagCarry, r > 0xFF)
	cpu.setFlag(flagOverflow, overflow)
	cpu.A = uint8(r)
	cpu.setZN(cpu.A)

	if arg.pageCross {
		cpu.Halt++
	}
}

func sbc(cpu *CPU, mem Memory, arg operand) {
	var (
		a = uint16(cpu.A)
		b = uint16(mem.Read(arg.addr))
	)

	r := a - b - uint16(1-cpu.carried())
	overflow := (a^b)&0x80 != 0 && (a^r)&0x80 != 0

	cpu.setFlag(flagCarry, r < 0x100)
	cpu.setFlag(flagOverflow, overflow)
	cpu.A = uint8(r)
	cpu.setZN(cpu.A)

	if arg.pageCross {
		cpu.Halt++
	}
}

func and(cpu *CPU, mem Memory, arg operand) {
	cpu.A &= mem.Read(arg.addr)
	cpu.setZN(cpu.A)

	if arg.pageCross {
		cpu.Halt++
	}
}

func ora(cpu *CPU, mem Memory, arg operand) {
	cpu.A |= mem.Read(arg.addr)
	cpu.setZN(cpu.A)

	if arg.pageCross {
		cpu.Halt++
	}
}

func eor(cpu *CPU, mem Memory, arg operand) {
	cpu.A ^= mem.Read(arg.addr)
	cpu.setZN(cpu.A)

	if arg.pageCross {
		cpu.Halt++
	}
}

func asl(cpu *CPU, mem Memory, arg operand) {
	if arg.mode == AddrModeAcc {
		data := cpu.A
		cpu.setFlag(flagCarry, data&0x80 != 0)
		data <<= 1
		cpu.setZN(data)
		cpu.A = data
	} else {
		data := mem.Read(arg.addr)
		cpu.setFlag(flagCarry, data&0x80 != 0)
		data <<= 1
		cpu.setZN(data)
		mem.Write(arg.addr, data)
	}
}

func lsr(cpu *CPU, mem Memory, arg operand) {
	if arg.mode == AddrModeAcc {
		data := cpu.A
		cpu.setFlag(flagCarry, data&0x01 != 0)
		data >>= 1
		cpu.setZN(data)
		cpu.A = data
	} else {
		data := mem.Read(arg.addr)
		cpu.setFlag(flagCarry, data&0x01 != 0)
		data >>= 1
		cpu.setZN(data)
		mem.Write(arg.addr, data)
	}
}

func rol(cpu *CPU, mem Memory, arg operand) {
	if arg.mode == AddrModeAcc {
		data := cpu.A
		carr := cpu.carried()
		cpu.setFlag(flagCarry, data&0x80 != 0)
		data = data<<1 | carr
		cpu.setZN(data)
		cpu.A = data
	} else {
		data := mem.Read(arg.addr)
		carr := cpu.carried()
		cpu.setFlag(flagCarry, data&0x80 != 0)
		data = data<<1 | carr
		cpu.setZN(data)
		mem.Write(arg.addr, data)
	}
}

func ror(cpu *CPU, mem Memory, arg operand) {
	if arg.mode == AddrModeAcc {
		data := cpu.A
		carr := cpu.carried()
		cpu.setFlag(flagCarry, data&0x01 != 0)
		data = data>>1 | carr<<7
		cpu.setZN(data)
		cpu.A = data
	} else {
		data := mem.Read(arg.addr)
		carr := cpu.carried()
		cpu.setFlag(flagCarry, data&0x01 != 0)
		data = data>>1 | carr<<7
		cpu.setZN(data)
		mem.Write(arg.addr, data)
	}
}

func bit(cpu *CPU, mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	cpu.setFlag(flagZero, cpu.A&data == 0)
	cpu.setFlag(flagOverflow, data&(1<<6) != 0)
	cpu.setFlag(flagNegative, data&(1<<7) != 0)
}

func cmp(cpu *CPU, mem Memory, arg operand) {
	data := uint16(cpu.A) - uint16(mem.Read(arg.addr))
	cpu.setFlag(flagCarry, data < 0x100)
	cpu.setZN(uint8(data))

	if arg.pageCross {
		cpu.Halt++
	}
}

func cpx(cpu *CPU, mem Memory, arg operand) {
	data := uint16(cpu.X) - uint16(mem.Read(arg.addr))
	cpu.setFlag(flagCarry, data < 0x100)
	cpu.setZN(uint8(data))

	if arg.pageCross {
		cpu.Halt++
	}
}

func cpy(cpu *CPU, mem Memory, arg operand) {
	data := uint16(cpu.Y) - uint16(mem.Read(arg.addr))
	cpu.setFlag(flagCarry, data < 0x100)
	cpu.setZN(uint8(data))

	if arg.pageCross {
		cpu.Halt++
	}
}

func jmp(cpu *CPU, mem Memory, arg operand) {
	cpu.PC = arg.addr
}

func jsr(cpu *CPU, mem Memory, arg operand) {
	cpu.pushWord(mem, cpu.PC-1)
	cpu.PC = arg.addr
}

func rts(cpu *CPU, mem Memory, arg operand) {
	addr := cpu.popWord(mem)
	cpu.PC = addr + 1
}

func bcc(cpu *CPU, mem Memory, arg operand) {
	if !cpu.getFlag(flagCarry) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func bcs(cpu *CPU, mem Memory, arg operand) {
	if cpu.getFlag(flagCarry) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func beq(cpu *CPU, mem Memory, arg operand) {
	if cpu.getFlag(flagZero) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func bmi(cpu *CPU, mem Memory, arg operand) {
	if cpu.getFlag(flagNegative) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func bne(cpu *CPU, mem Memory, arg operand) {
	if !cpu.getFlag(flagZero) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func bpl(cpu *CPU, mem Memory, arg operand) {
	if !cpu.getFlag(flagNegative) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func bvc(cpu *CPU, mem Memory, arg operand) {
	if !cpu.getFlag(flagOverflow) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func bvs(cpu *CPU, mem Memory, arg operand) {
	if cpu.getFlag(flagOverflow) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func brk(cpu *CPU, mem Memory, arg operand) {
	cpu.pushWord(mem, cpu.PC+1)
	cpu.pushByte(mem, cpu.P|0x30)
	cpu.setFlag(flagInterrupt, true)
	cpu.PC = readWord(mem, vecIRQ)
}

func clc(cpu *CPU, mem Memory, arg operand) {
	cpu.setFlag(flagCarry, false)
}

func cld(cpu *CPU, mem Memory, arg operand) {
	cpu.setFlag(flagDecimal, false)
}

func cli(cpu *CPU, mem Memory, arg operand) {
	cpu.setFlag(flagInterrupt, false)
}

func clv(cpu *CPU, mem Memory, arg operand) {
	cpu.setFlag(flagOverflow, false)
}

func sec(cpu *CPU, mem Memory, arg operand) {
	cpu.setFlag(flagCarry, true)
}

func sed(cpu *CPU, mem Memory, arg operand) {
	cpu.setFlag(flagDecimal, true)
}

func sei(cpu *CPU, mem Memory, arg operand) {
	cpu.setFlag(flagInterrupt, true)
}

func rti(cpu *CPU, mem Memory, arg operand) {
	cpu.P = Flags(cpu.popByte(mem))&0xEF | 0x20
	cpu.setFlag(flagBreak, false)
	cpu.PC = cpu.popWord(mem)
}
