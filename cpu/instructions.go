package cpu

type instruction struct {
	name   string
	opcode Byte
	mode   AddrMode
	size   int
	cost   int
}

var instrTable = []instruction{
	// NOP - No Operation
	{"NOP", OpNopImp, AddrModeImp, 1, 2},

	// LDA - Load Accumulator
	{"LDA", OpLdaImm, AddrModeImm, 2, 2},
	{"LDA", OpLdaZp, AddrModeZp, 2, 3},
	{"LDA", OpLdaZpX, AddrModeZpX, 2, 4},
	{"LDA", OpLdaAbs, AddrModeAbs, 3, 4},
	{"LDA", OpLdaAbsX, AddrModeAbsX, 3, 4},
	{"LDA", OpLdaAbsY, AddrModeAbsY, 3, 4},
	{"LDA", OpLdaIndX, AddrModeIndX, 2, 6},
	{"LDA", OpLdaIndY, AddrModeIndY, 2, 5},

	// STA - Store Accumulator
	{"STA", OpStaZp, AddrModeZp, 2, 3},
	{"STA", OpStaZpX, AddrModeZpX, 2, 4},
	{"STA", OpStaAbs, AddrModeAbs, 3, 4},
	{"STA", OpStaAbsX, AddrModeAbsX, 3, 5},
	{"STA", OpStaAbsY, AddrModeAbsY, 3, 5},
	{"STA", OpStaIndX, AddrModeIndX, 2, 6},
	{"STA", OpStaIndY, AddrModeIndY, 2, 6},

	// LDX - Load X Register
	{"LDX", OpLdxImm, AddrModeImm, 2, 2},
	{"LDX", OpLdxZp, AddrModeZp, 2, 3},
	{"LDX", OpLdxZpY, AddrModeZpY, 2, 4},
	{"LDX", OpLdxAbs, AddrModeAbs, 3, 4},
	{"LDX", OpLdxAbsY, AddrModeAbsY, 3, 4},

	// STX - Store X Register
	{"STX", OpStxZp, AddrModeZp, 2, 3},
	{"STX", OpStxZpY, AddrModeZpY, 2, 4},
	{"STX", OpStxAbs, AddrModeAbs, 3, 4},

	// LDY - Load Y Register
	{"LDY", OpLdyImm, AddrModeImm, 2, 2},
	{"LDY", OpLdyZp, AddrModeZp, 2, 3},
	{"LDY", OpLdyZpX, AddrModeZpX, 2, 4},
	{"LDY", OpLdyAbs, AddrModeAbs, 3, 4},
	{"LDY", OpLdyAbsX, AddrModeAbsX, 3, 4},

	// STY - Store Y Register
	{"STY", OpStyZp, AddrModeZp, 2, 3},
	{"STY", OpStyZpX, AddrModeZpX, 2, 4},
	{"STY", OpStyAbs, AddrModeAbs, 3, 4},

	// TAX - Transfer Accumulator to X
	{"TAX", OpTaxImp, AddrModeImp, 1, 2},

	// TAY - Transfer Accumulator to Y
	{"TAY", OpTayImp, AddrModeImp, 1, 2},

	// TXA - Transfer X to Accumulator
	{"TXA", OpTxaImp, AddrModeImp, 1, 2},

	// TYA - Transfer Y to Accumulator
	{"TYA", OpTyaImp, AddrModeImp, 1, 2},

	// TSX - Transfer Stack Pointer to X
	{"TSX", OpTsxImp, AddrModeImp, 1, 2},

	// TXS - Transfer X to Stack Pointer
	{"TXS", OpTxsImp, AddrModeImp, 1, 2},

	// PHA - Push Accumulator
	{"PHA", OpPhaImp, AddrModeImp, 1, 3},

	// PHP - Push Processor Status
	{"PHP", OpPhpImp, AddrModeImp, 1, 3},

	// PLA - Pull Accumulator
	{"PLA", OpPlaImp, AddrModeImp, 1, 4},

	// PLP - Pull Processor Status
	{"PLP", OpPlpImp, AddrModeImp, 1, 4},

	// ADC - Add with Carry
	{"ADC", OpAdcImm, AddrModeImm, 2, 2},
	{"ADC", OpAdcZp, AddrModeZp, 2, 3},
	{"ADC", OpAdcZpX, AddrModeZpX, 2, 4},
	{"ADC", OpAdcAbs, AddrModeAbs, 3, 4},
	{"ADC", OpAdcAbsX, AddrModeAbsX, 3, 4},
	{"ADC", OpAdcAbsY, AddrModeAbsY, 3, 4},
	{"ADC", OpAdcIndX, AddrModeIndX, 2, 6},
	{"ADC", OpAdcIndY, AddrModeIndY, 2, 5},

	// SBC - Subtract with Carry
	{"SBC", OpSbcImm, AddrModeImm, 2, 2},
	{"SBC", OpSbcZp, AddrModeZp, 2, 3},
	{"SBC", OpSbcZpX, AddrModeZpX, 2, 4},
	{"SBC", OpSbcAbs, AddrModeAbs, 3, 4},
	{"SBC", OpSbcAbsX, AddrModeAbsX, 3, 4},
	{"SBC", OpSbcAbsY, AddrModeAbsY, 3, 4},
	{"SBC", OpSbcIndX, AddrModeIndX, 2, 6},
	{"SBC", OpSbcIndY, AddrModeIndY, 2, 5},

	// AND - Logical AND
	{"AND", OpAndImm, AddrModeImm, 2, 2},
	{"AND", OpAndZp, AddrModeZp, 2, 3},
	{"AND", OpAndZpX, AddrModeZpX, 2, 4},
	{"AND", OpAndAbs, AddrModeAbs, 3, 4},
	{"AND", OpAndAbsX, AddrModeAbsX, 3, 4},
	{"AND", OpAndAbsY, AddrModeAbsY, 3, 4},
	{"AND", OpAndIndX, AddrModeIndX, 2, 6},
	{"AND", OpAndIndY, AddrModeIndY, 2, 5},

	// ORA - Logical OR
	{"ORA", OpOraImm, AddrModeImm, 2, 2},
	{"ORA", OpOraZp, AddrModeZp, 2, 3},
	{"ORA", OpOraZpX, AddrModeZpX, 2, 4},
	{"ORA", OpOraAbs, AddrModeAbs, 3, 4},
	{"ORA", OpOraAbsX, AddrModeAbsX, 3, 4},
	{"ORA", OpOraAbsY, AddrModeAbsY, 3, 4},
	{"ORA", OpOraIndX, AddrModeIndX, 2, 6},
	{"ORA", OpOraIndY, AddrModeIndY, 2, 5},

	// EOR - Exclusive OR
	{"EOR", OpEorImm, AddrModeImm, 2, 2},
	{"EOR", OpEorZp, AddrModeZp, 2, 3},
	{"EOR", OpEorZpX, AddrModeZpX, 2, 4},
	{"EOR", OpEorAbs, AddrModeAbs, 3, 4},
	{"EOR", OpEorAbsX, AddrModeAbsX, 3, 4},
	{"EOR", OpEorAbsY, AddrModeAbsY, 3, 4},
	{"EOR", OpEorIndX, AddrModeIndX, 2, 6},
	{"EOR", OpEorIndY, AddrModeIndY, 2, 5},

	// CMP - Compare
	{"CMP", OpCmpImm, AddrModeImm, 2, 2},
	{"CMP", OpCmpZp, AddrModeZp, 2, 3},
	{"CMP", OpCmpZpX, AddrModeZpX, 2, 4},
	{"CMP", OpCmpAbs, AddrModeAbs, 3, 4},
	{"CMP", OpCmpAbsX, AddrModeAbsX, 3, 4},
	{"CMP", OpCmpAbsY, AddrModeAbsY, 3, 4},
	{"CMP", OpCmpIndX, AddrModeIndX, 2, 6},
	{"CMP", OpCmpIndY, AddrModeIndY, 2, 5},

	// CPX - Compare X Register
	{"CPX", OpCpxImm, AddrModeImm, 2, 2},
	{"CPX", OpCpxZp, AddrModeZp, 2, 3},
	{"CPX", OpCpxAbs, AddrModeAbs, 3, 4},

	// CPY - Compare Y Register
	{"CPY", OpCpyImm, AddrModeImm, 2, 2},
	{"CPY", OpCpyZp, AddrModeZp, 2, 3},
	{"CPY", OpCpyAbs, AddrModeAbs, 3, 4},

	// BIT - Bit Test
	{"BIT", OpBitZp, AddrModeZp, 2, 3},
	{"BIT", OpBitAbs, AddrModeAbs, 3, 4},

	// ASL - Arithmetic Shift Left
	{"ASL", OpAslAcc, AddrModeAcc, 1, 2},
	{"ASL", OpAslZp, AddrModeZp, 2, 5},
	{"ASL", OpAslZpX, AddrModeZpX, 2, 6},
	{"ASL", OpAslAbs, AddrModeAbs, 3, 6},
	{"ASL", OpAslAbsX, AddrModeAbsX, 3, 7},

	// LSR - Logical Shift Right
	{"LSR", OpLsrAcc, AddrModeAcc, 1, 2},
	{"LSR", OpLsrZp, AddrModeZp, 2, 5},
	{"LSR", OpLsrZpX, AddrModeZpX, 2, 6},
	{"LSR", OpLsrAbs, AddrModeAbs, 3, 6},
	{"LSR", OpLsrAbsX, AddrModeAbsX, 3, 7},

	// ROL - Rotate Left
	{"ROL", OpRolAcc, AddrModeAcc, 1, 2},
	{"ROL", OpRolZp, AddrModeZp, 2, 5},
	{"ROL", OpRolZpX, AddrModeZpX, 2, 6},
	{"ROL", OpRolAbs, AddrModeAbs, 3, 6},
	{"ROL", OpRolAbsX, AddrModeAbsX, 3, 7},

	// ROR - Rotate Right
	{"ROR", OpRorAcc, AddrModeAcc, 1, 2},
	{"ROR", OpRorZp, AddrModeZp, 2, 5},
	{"ROR", OpRorZpX, AddrModeZpX, 2, 6},
	{"ROR", OpRorAbs, AddrModeAbs, 3, 6},
	{"ROR", OpRorAbsX, AddrModeAbsX, 3, 7},

	// INC - Increment Memory
	{"INC", OpIncZp, AddrModeZp, 2, 5},
	{"INC", OpIncZpX, AddrModeZpX, 2, 6},
	{"INC", OpIncAbs, AddrModeAbs, 3, 6},
	{"INC", OpIncAbsX, AddrModeAbsX, 3, 7},

	// DEC - Decrement Memory
	{"DEC", OpDecZp, AddrModeZp, 2, 5},
	{"DEC", OpDecZpX, AddrModeZpX, 2, 6},
	{"DEC", OpDecAbs, AddrModeAbs, 3, 6},
	{"DEC", OpDecAbsX, AddrModeAbsX, 3, 7},

	// JMP - Jump
	{"JMP", OpJmpAbs, AddrModeAbs, 3, 3},
	{"JMP", OpJmpInd, AddrModeInd, 3, 5},

	// JSR - Jump to Subroutine
	{"JSR", OpJsrAbs, AddrModeAbs, 3, 6},

	// RTS - Return from Subroutine
	{"RTS", OpRtsImp, AddrModeImp, 1, 6},

	// RTI - Return from Interrupt
	{"RTI", OpRtiImp, AddrModeImp, 1, 6},

	// BRK - Force Interrupt
	{"BRK", OpBrkImp, AddrModeImp, 1, 7},

	// NOP - No Operation
	{"BPL", OpBplRel, AddrModeRel, 2, 2},

	// BMI - Branch if Minus
	{"BMI", OpBmiRel, AddrModeRel, 2, 2},

	// BVC - Branch if Overflow Clear
	{"BVC", OpBvcRel, AddrModeRel, 2, 2},

	// BVS - Branch if Overflow Set
	{"BVS", OpBvsRel, AddrModeRel, 2, 2},

	// BCC - Branch if Carry Clear
	{"BCC", OpBccRel, AddrModeRel, 2, 2},

	// BCS - Branch if Carry Set
	{"BCS", OpBcsRel, AddrModeRel, 2, 2},

	// BNE - Branch if Not Equal
	{"BNE", OpBneRel, AddrModeRel, 2, 2},

	// BEQ - Branch if Equal
	{"BEQ", OpBeqRel, AddrModeRel, 2, 2},

	// CLC - Clear Carry Flag
	{"CLC", OpClcImp, AddrModeImp, 1, 2},

	// SEC - Set Carry Flag
	{"SEC", OpSecImp, AddrModeImp, 1, 2},

	// CLI - Clear Interrupt Disable
	{"CLI", OpCliImp, AddrModeImp, 1, 2},

	// SEI - Set Interrupt Disable
	{"SEI", OpSeiImp, AddrModeImp, 1, 2},

	// CLV - Clear Overflow Flag
	{"CLV", OpClvImp, AddrModeImp, 1, 2},

	// CLD - Clear Decimal Mode
	{"CLD", OpCldImp, AddrModeImp, 1, 2},

	// SED - Set Decimal Flag
	{"SED", OpSedImp, AddrModeImp, 1, 2},
}
