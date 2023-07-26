package cpu

func (cpu *CPU) execute(mem Memory, instr Instruction, arg operand) bool {
	switch instr.Name {
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
		return false
	}

	return true
}

// lda loads the accumulator with a value from memory.
func (cpu *CPU) lda(mem Memory, arg operand) {
	cpu.A = mem.Read(arg.addr)
	cpu.setZN(cpu.A)

	if arg.pageCross {
		cpu.Halt++
	}
}

// sta stores the accumulator in memory.
func (cpu *CPU) sta(mem Memory, arg operand) {
	mem.Write(arg.addr, cpu.A)
}

// ldx loads the X register with a value from memory.
func (cpu *CPU) ldx(mem Memory, arg operand) {
	cpu.X = mem.Read(arg.addr)
	cpu.setZN(cpu.X)

	if arg.pageCross {
		cpu.Halt++
	}
}

// stx stores the X register in memory.
func (cpu *CPU) stx(mem Memory, arg operand) {
	mem.Write(arg.addr, cpu.X)
}

// ldy loads the Y register with a value from memory.
func (cpu *CPU) ldy(mem Memory, arg operand) {
	cpu.Y = mem.Read(arg.addr)
	cpu.setZN(cpu.Y)

	if arg.pageCross {
		cpu.Halt++
	}
}

// sty stores the Y register in memory.
func (cpu *CPU) sty(mem Memory, arg operand) {
	mem.Write(arg.addr, cpu.Y)
}

// tax transfers the accumulator to the X register.
func (cpu *CPU) tax(mem Memory, arg operand) {
	cpu.X = cpu.A
	cpu.setZN(cpu.X)
}

// txa transfers the X register to the accumulator.
func (cpu *CPU) txa(mem Memory, arg operand) {
	cpu.A = cpu.X
	cpu.setZN(cpu.A)
}

// tay transfers the accumulator to the Y register.
func (cpu *CPU) tay(mem Memory, arg operand) {
	cpu.Y = cpu.A
	cpu.setZN(cpu.Y)
}

// tya transfers the Y register to the accumulator.
func (cpu *CPU) tya(mem Memory, arg operand) {
	cpu.A = cpu.Y
	cpu.setZN(cpu.A)
}

// tsx transfers the stack pointer to the X register.
func (cpu *CPU) tsx(mem Memory, arg operand) {
	cpu.X = cpu.SP
	cpu.setZN(cpu.X)
}

// txs transfers the X register to the stack pointer.
func (cpu *CPU) txs(mem Memory, arg operand) {
	cpu.SP = cpu.X
}

// pha pushes the accumulator onto the stack.
func (cpu *CPU) pha(mem Memory, arg operand) {
	cpu.pushByte(mem, cpu.A)
}

// pla pops a value from the stack into the accumulator.
func (cpu *CPU) pla(mem Memory, arg operand) {
	cpu.A = cpu.popByte(mem)
	cpu.setZN(cpu.A)
}

// php pushes the processor status onto the stack.
func (cpu *CPU) php(mem Memory, arg operand) {
	cpu.pushByte(mem, uint8(cpu.P)|0x10)
}

// plp pops a value from the stack into the processor status.
func (cpu *CPU) plp(mem Memory, arg operand) {
	cpu.P = Flags(cpu.popByte(mem))&0xEF | 0x20
}

// inc increments a value in memory.
func (cpu *CPU) inc(mem Memory, arg operand) {
	val := mem.Read(arg.addr) + 1
	mem.Write(arg.addr, val)
	cpu.setZN(val)
}

// inx increments the X register.
func (cpu *CPU) inx(mem Memory, arg operand) {
	cpu.X++
	cpu.setZN(cpu.X)
}

// iny increments the Y register.
func (cpu *CPU) iny(mem Memory, arg operand) {
	cpu.Y++
	cpu.setZN(cpu.Y)
}

// dec decrements a value in memory.
func (cpu *CPU) dec(mem Memory, arg operand) {
	data := mem.Read(arg.addr) - 1
	mem.Write(arg.addr, data)
	cpu.setZN(data)
}

// dex decrements the X register.
func (cpu *CPU) dex(mem Memory, arg operand) {
	cpu.X--
	cpu.setZN(cpu.X)
}

// dey decrements the Y register.
func (cpu *CPU) dey(mem Memory, arg operand) {
	cpu.Y--
	cpu.setZN(cpu.Y)
}

// adc adds a value from memory to the accumulator with carry. The carry flag is
// set if the result is greater than 255. The overflow flag is set if the result
// is greater than 127 or less than -128 (incorrect sign bit).
func (cpu *CPU) adc(mem Memory, arg operand) {
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

func (cpu *CPU) sbc(mem Memory, arg operand) {
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

func (cpu *CPU) and(mem Memory, arg operand) {
	cpu.A &= mem.Read(arg.addr)
	cpu.setZN(cpu.A)

	if arg.pageCross {
		cpu.Halt++
	}
}

func (cpu *CPU) ora(mem Memory, arg operand) {
	cpu.A |= mem.Read(arg.addr)
	cpu.setZN(cpu.A)

	if arg.pageCross {
		cpu.Halt++
	}
}

func (cpu *CPU) eor(mem Memory, arg operand) {
	cpu.A ^= mem.Read(arg.addr)
	cpu.setZN(cpu.A)

	if arg.pageCross {
		cpu.Halt++
	}
}

func (cpu *CPU) asl(mem Memory, arg operand) {
	var (
		write = func(v uint8) { mem.Write(arg.addr, v) }
		read  = func() uint8 { return mem.Read(arg.addr) }
	)

	if arg.mode == AddrModeAcc {
		write = func(v uint8) { cpu.A = v }
		read = func() uint8 { return cpu.A }
	}

	data := read()
	cpu.setFlag(flagCarry, data&0x80 != 0)
	data <<= 1
	cpu.setZN(data)
	write(data)
}

func (cpu *CPU) lsr(mem Memory, arg operand) {
	var (
		write = func(v uint8) { mem.Write(arg.addr, v) }
		read  = func() uint8 { return mem.Read(arg.addr) }
	)

	if arg.mode == AddrModeAcc {
		write = func(v uint8) { cpu.A = v }
		read = func() uint8 { return cpu.A }
	}

	data := read()
	cpu.setFlag(flagCarry, data&0x01 != 0)
	data >>= 1
	cpu.setZN(data)
	write(data)
}

func (cpu *CPU) rol(mem Memory, arg operand) {
	var (
		write = func(v uint8) { mem.Write(arg.addr, v) }
		read  = func() uint8 { return mem.Read(arg.addr) }
	)

	if arg.mode == AddrModeAcc {
		write = func(v uint8) { cpu.A = v }
		read = func() uint8 { return cpu.A }
	}

	data := read()
	carr := cpu.carried()

	cpu.setFlag(flagCarry, data&0x80 != 0)
	data = data<<1 | carr
	cpu.setZN(data)
	write(data)
}

func (cpu *CPU) ror(mem Memory, arg operand) {
	var (
		write = func(v uint8) { mem.Write(arg.addr, v) }
		read  = func() uint8 { return mem.Read(arg.addr) }
	)

	if arg.mode == AddrModeAcc {
		write = func(v uint8) { cpu.A = v }
		read = func() uint8 { return cpu.A }
	}

	data := read()
	carr := cpu.carried()

	cpu.setFlag(flagCarry, data&0x01 != 0)
	data = data>>1 | carr<<7
	cpu.setZN(data)
	write(data)
}

func (cpu *CPU) bit(mem Memory, arg operand) {
	data := mem.Read(arg.addr)
	cpu.setFlag(flagZero, cpu.A&data == 0)
	cpu.setFlag(flagOverflow, data&(1<<6) != 0)
	cpu.setFlag(flagNegative, data&(1<<7) != 0)
}

func (cpu *CPU) cmp(mem Memory, arg operand) {
	data := uint16(cpu.A) - uint16(mem.Read(arg.addr))
	cpu.setFlag(flagCarry, data < 0x100)
	cpu.setZN(uint8(data))

	if arg.pageCross {
		cpu.Halt++
	}
}

func (cpu *CPU) cpx(mem Memory, arg operand) {
	data := uint16(cpu.X) - uint16(mem.Read(arg.addr))
	cpu.setFlag(flagCarry, data < 0x100)
	cpu.setZN(uint8(data))

	if arg.pageCross {
		cpu.Halt++
	}
}

func (cpu *CPU) cpy(mem Memory, arg operand) {
	data := uint16(cpu.Y) - uint16(mem.Read(arg.addr))
	cpu.setFlag(flagCarry, data < 0x100)
	cpu.setZN(uint8(data))

	if arg.pageCross {
		cpu.Halt++
	}
}

func (cpu *CPU) jmp(mem Memory, arg operand) {
	cpu.PC = arg.addr
}

func (cpu *CPU) jsr(mem Memory, arg operand) {
	cpu.pushWord(mem, cpu.PC-1)
	cpu.PC = arg.addr
}

func (cpu *CPU) rts(mem Memory, arg operand) {
	addr := cpu.popWord(mem)
	cpu.PC = addr + 1
}

func (cpu *CPU) bcc(mem Memory, arg operand) {
	if !cpu.getFlag(flagCarry) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bcs(mem Memory, arg operand) {
	if cpu.getFlag(flagCarry) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) beq(mem Memory, arg operand) {
	if cpu.getFlag(flagZero) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bmi(mem Memory, arg operand) {
	if cpu.getFlag(flagNegative) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bne(mem Memory, arg operand) {
	if !cpu.getFlag(flagZero) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bpl(mem Memory, arg operand) {
	if !cpu.getFlag(flagNegative) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bvc(mem Memory, arg operand) {
	if !cpu.getFlag(flagOverflow) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bvs(mem Memory, arg operand) {
	if cpu.getFlag(flagOverflow) {
		cpu.PC = arg.addr
		cpu.Halt += 1

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) brk(mem Memory, arg operand) {
	cpu.pushWord(mem, cpu.PC)
	cpu.pushByte(mem, uint8(cpu.P&flagBreak))
	cpu.PC = readWord(mem, vecIRQ)
}

func (cpu *CPU) clc(mem Memory, arg operand) {
	cpu.setFlag(flagCarry, false)
}

func (cpu *CPU) cld(mem Memory, arg operand) {
	cpu.setFlag(flagDecimal, false)
}

func (cpu *CPU) cli(mem Memory, arg operand) {
	cpu.setFlag(flagInterrupt, false)
}

func (cpu *CPU) clv(mem Memory, arg operand) {
	cpu.setFlag(flagOverflow, false)
}

func (cpu *CPU) sec(mem Memory, arg operand) {
	cpu.setFlag(flagCarry, true)
}

func (cpu *CPU) sed(mem Memory, arg operand) {
	cpu.setFlag(flagDecimal, true)
}

func (cpu *CPU) sei(mem Memory, arg operand) {
	cpu.setFlag(flagInterrupt, true)
}

func (cpu *CPU) rti(mem Memory, arg operand) {
	cpu.P = Flags(cpu.popByte(mem))&0xEF | 0x20
	cpu.setFlag(flagBreak, false)
	cpu.PC = cpu.popWord(mem)
}
