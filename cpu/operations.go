package cpu

func (cpu *CPU) lda(mem Memory, arg operand) {
	cpu.A = arg.val
	if arg.pageCross {
		cpu.Halt++
	}

	cpu.setZN(arg.val)
}

func (cpu *CPU) sta(mem Memory, arg operand) {
	mem.Write(arg.addr, cpu.A)
}

func (cpu *CPU) ldx(mem Memory, arg operand) {
	cpu.X = arg.val
	if arg.pageCross {
		cpu.Halt++
	}

	cpu.setZN(arg.val)
}

func (cpu *CPU) stx(mem Memory, arg operand) {
	mem.Write(arg.addr, cpu.X)
}

func (cpu *CPU) ldy(mem Memory, arg operand) {
	cpu.Y = arg.val
	if arg.pageCross {
		cpu.Halt++
	}

	cpu.setZN(arg.val)
}

func (cpu *CPU) sty(mem Memory, arg operand) {
	mem.Write(arg.addr, cpu.Y)
}

func (cpu *CPU) tax(mem Memory, arg operand) {
	cpu.X = cpu.A
	cpu.setZN(cpu.X)
}

func (cpu *CPU) txa(mem Memory, arg operand) {
	cpu.A = cpu.X
	cpu.setZN(cpu.A)
}

func (cpu *CPU) tay(mem Memory, arg operand) {
	cpu.Y = cpu.A
	cpu.setZN(cpu.Y)
}

func (cpu *CPU) tya(mem Memory, arg operand) {
	cpu.A = cpu.Y
	cpu.setZN(cpu.A)
}

func (cpu *CPU) tsx(mem Memory, arg operand) {
	cpu.X = cpu.SP
	cpu.setZN(cpu.X)
}

func (cpu *CPU) txs(mem Memory, arg operand) {
	cpu.SP = cpu.X
}

func (cpu *CPU) pha(mem Memory, arg operand) {
	cpu.pushByte(mem, cpu.A)
}

func (cpu *CPU) pla(mem Memory, arg operand) {
	cpu.A = cpu.popByte(mem)
	cpu.setZN(cpu.A)
}

func (cpu *CPU) php(mem Memory, arg operand) {
	cpu.pushByte(mem, uint8(cpu.P)|0x10)
}

func (cpu *CPU) plp(mem Memory, arg operand) {
	cpu.P = Flags(cpu.popByte(mem))&0xEF | 0x20
}

func (cpu *CPU) inc(mem Memory, arg operand) {
	val := arg.val + 1
	mem.Write(arg.addr, val)
	cpu.setZN(val)
}

func (cpu *CPU) inx(mem Memory, arg operand) {
	cpu.X++
	cpu.setZN(cpu.X)
}

func (cpu *CPU) iny(mem Memory, arg operand) {
	cpu.Y++
	cpu.setZN(cpu.Y)
}

func (cpu *CPU) dec(mem Memory, arg operand) {
	val := arg.val - 1
	mem.Write(arg.addr, val)
	cpu.setZN(val)
}

func (cpu *CPU) dex(mem Memory, arg operand) {
	cpu.X--
	cpu.setZN(cpu.X)
}

func (cpu *CPU) dey(mem Memory, arg operand) {
	cpu.Y--
	cpu.setZN(cpu.Y)
}

func (cpu *CPU) adc(mem Memory, arg operand) {
	val := uint16(cpu.A) + uint16(arg.val)
	if cpu.getFlag(FlagCarry) {
		val++
	}

	cpu.setFlag(FlagCarry, val > 0xFF)
	cpu.setFlag(FlagOverflow, (val^uint16(cpu.A))&(val^uint16(arg.val))&0x80 != 0)

	cpu.A = uint8(val)
	cpu.setZN(cpu.A)
}

func (cpu *CPU) sbc(mem Memory, arg operand) {
	val := uint16(cpu.A) - uint16(arg.val)
	if !cpu.getFlag(FlagCarry) {
		val--
	}

	cpu.setFlag(FlagCarry, val < 0x100)
	cpu.setFlag(FlagOverflow, (val^uint16(cpu.A))&(val^uint16(arg.val))&0x80 != 0)

	cpu.A = uint8(val)
	cpu.setZN(cpu.A)
}

func (cpu *CPU) and(mem Memory, arg operand) {
	cpu.A &= arg.val
	cpu.setZN(cpu.A)
}

func (cpu *CPU) ora(mem Memory, arg operand) {
	cpu.A |= arg.val
	cpu.setZN(cpu.A)
}

func (cpu *CPU) eor(mem Memory, arg operand) {
	cpu.A ^= arg.val
	cpu.setZN(cpu.A)
}

func (cpu *CPU) asl(mem Memory, arg operand) {
	cpu.setFlag(FlagCarry, arg.val&0x80 != 0)
	val := arg.val << 1
	cpu.setZN(val)

	if arg.addr == 0 {
		cpu.A = val
		return
	}

	mem.Write(arg.addr, val)
}

func (cpu *CPU) lsr(mem Memory, arg operand) {
	cpu.setFlag(FlagCarry, arg.val&0x01 != 0)
	val := arg.val >> 1
	cpu.setZN(val)

	if arg.addr == 0 {
		cpu.A = val
		return
	}

	mem.Write(arg.addr, val)
}

func (cpu *CPU) rol(mem Memory, arg operand) {
	val := arg.val << 1
	if cpu.getFlag(FlagCarry) {
		val |= 0x01
	}

	cpu.setFlag(FlagCarry, arg.val&0x80 != 0)
	cpu.setZN(val)

	if arg.addr == 0 {
		cpu.A = val
		return
	}

	mem.Write(arg.addr, val)
}

func (cpu *CPU) ror(mem Memory, arg operand) {
	val := arg.val >> 1
	if cpu.getFlag(FlagCarry) {
		val |= 0x80
	}

	cpu.setFlag(FlagCarry, arg.val&0x01 != 0)
	cpu.setZN(val)

	if arg.addr == 0 {
		cpu.A = val
		return
	}

	mem.Write(arg.addr, val)
}

func (cpu *CPU) bit(mem Memory, arg operand) {
	cpu.setFlag(FlagZero, cpu.A&arg.val == 0)
	cpu.setFlag(FlagOverflow, arg.val&(1<<6) != 0)
	cpu.setFlag(FlagNegative, arg.val&(1<<7) != 0)
}

func (cpu *CPU) cmp(mem Memory, arg operand) {
	val := uint16(cpu.A) - uint16(arg.val)
	cpu.setFlag(FlagCarry, val < 0x100)
	cpu.setZN(uint8(val))

	if arg.pageCross {
		cpu.Halt++
	}
}

func (cpu *CPU) cpx(mem Memory, arg operand) {
	val := uint16(cpu.X) - uint16(arg.val)
	cpu.setFlag(FlagCarry, val < 0x100)
	cpu.setZN(uint8(val))

	if arg.pageCross {
		cpu.Halt++
	}
}

func (cpu *CPU) cpy(mem Memory, arg operand) {
	val := uint16(cpu.Y) - uint16(arg.val)
	cpu.setFlag(FlagCarry, val < 0x100)
	cpu.setZN(uint8(val))

	if arg.pageCross {
		cpu.Halt++
	}
}

func (cpu *CPU) jmp(mem Memory, arg operand) {
	cpu.PC = arg.addr
}

func (cpu *CPU) jsr(mem Memory, arg operand) {
	retAddr := cpu.PC - 1
	cpu.pushByte(mem, uint8(retAddr>>8))
	cpu.pushByte(mem, uint8(retAddr))
	cpu.PC = arg.addr
}

func (cpu *CPU) rts(mem Memory, arg operand) {
	lo := uint16(cpu.popByte(mem))
	hi := uint16(cpu.popByte(mem))
	addr := hi<<8 | lo
	cpu.PC = addr + 1
}

func (cpu *CPU) bcc(mem Memory, arg operand) {
	if !cpu.getFlag(FlagCarry) {
		cpu.PC = arg.addr
		cpu.Halt += 2

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bcs(mem Memory, arg operand) {
	if cpu.getFlag(FlagCarry) {
		cpu.PC = arg.addr
		cpu.Halt += 2

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) beq(mem Memory, arg operand) {
	if cpu.getFlag(FlagZero) {
		cpu.PC = arg.addr
		cpu.Halt += 2

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bmi(mem Memory, arg operand) {
	if cpu.getFlag(FlagNegative) {
		cpu.PC = arg.addr
		cpu.Halt += 2

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bne(mem Memory, arg operand) {
	if !cpu.getFlag(FlagZero) {
		cpu.PC = arg.addr
		cpu.Halt += 2

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bpl(mem Memory, arg operand) {
	if !cpu.getFlag(FlagNegative) {
		cpu.PC = arg.addr
		cpu.Halt += 2

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bvc(mem Memory, arg operand) {
	if !cpu.getFlag(FlagOverflow) {
		cpu.PC = arg.addr
		cpu.Halt += 2

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) bvs(mem Memory, arg operand) {
	if cpu.getFlag(FlagOverflow) {
		cpu.PC = arg.addr
		cpu.Halt += 2

		if arg.pageCross {
			cpu.Halt += 2
		}
	}
}

func (cpu *CPU) brk(mem Memory, arg operand) {
	cpu.pushByte(mem, uint8(cpu.PC>>8))
	cpu.pushByte(mem, uint8(cpu.PC))
	cpu.pushByte(mem, uint8(cpu.P))
	cpu.setFlag(FlagBreak, true)
	cpu.PC = cpu.readWord(mem, VecIRQ)
}

func (cpu *CPU) clc(mem Memory, arg operand) {
	cpu.setFlag(FlagCarry, false)
}

func (cpu *CPU) cld(mem Memory, arg operand) {
	cpu.setFlag(FlagDecimal, false)
}

func (cpu *CPU) cli(mem Memory, arg operand) {
	cpu.setFlag(FlagInterrupt, false)
}

func (cpu *CPU) clv(mem Memory, arg operand) {
	cpu.setFlag(FlagOverflow, false)
}

func (cpu *CPU) sec(mem Memory, arg operand) {
	cpu.setFlag(FlagCarry, true)
}

func (cpu *CPU) sed(mem Memory, arg operand) {
	cpu.setFlag(FlagDecimal, true)
}

func (cpu *CPU) sei(mem Memory, arg operand) {
	cpu.setFlag(FlagInterrupt, true)
}

func (cpu *CPU) rti(mem Memory, arg operand) {
	cpu.P = Flags(cpu.popByte(mem))&0xEF | 0x20
	cpu.PC = cpu.popWord(mem)
}
