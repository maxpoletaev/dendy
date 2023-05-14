package cpu

var (
	instructions map[uint8]instrInfo
)

func init() {
	instructions = make(map[uint8]instrInfo)

	for _, instr := range instrTable {
		instructions[instr.opcode] = instr
	}
}

const (
	NOP_Imp  uint8 = 0xEA
	LDA_Imm  uint8 = 0xA9
	LDA_Zp   uint8 = 0xA5
	LDA_ZpX  uint8 = 0xB5
	LDA_Abs  uint8 = 0xAD
	LDA_AbsX uint8 = 0xBD
	LDA_AbsY uint8 = 0xB9
	LDA_IndX uint8 = 0xA1
	LDA_IndY uint8 = 0xB1
	STA_Zp   uint8 = 0x85
	STA_ZpX  uint8 = 0x95
	STA_Abs  uint8 = 0x8D
	STA_AbsX uint8 = 0x9D
	STA_Ind  uint8 = 0x99
	STA_IndX uint8 = 0x81
	STA_IndY uint8 = 0x91
	LDX_Imm  uint8 = 0xA2
	LDX_Zp   uint8 = 0xA6
	LDX_ZpY  uint8 = 0xB6
	LDX_Abs  uint8 = 0xAE
	LDX_AbsY uint8 = 0xBE
	STX_Zp   uint8 = 0x86
	STX_ZpY  uint8 = 0x96
	STX_Abs  uint8 = 0x8E
	LDY_Imm  uint8 = 0xA0
	LDY_Zp   uint8 = 0xA4
	LDY_ZpX  uint8 = 0xB4
	LDY_Abs  uint8 = 0xAC
	LDY_AbsX uint8 = 0xBC
	STY_Zp   uint8 = 0x84
	STY_ZpX  uint8 = 0x94
	STY_Abs  uint8 = 0x8C
	TAX_Imp  uint8 = 0xAA
	TAY_Imp  uint8 = 0xA8
	TXA_Imp  uint8 = 0x8A
	TYA_Imp  uint8 = 0x98
	TSX_Imp  uint8 = 0xBA
	TXS_Imp  uint8 = 0x9A
	PHA_Imp  uint8 = 0x48
	PLA_Imp  uint8 = 0x68
	PHP_Imp  uint8 = 0x08
	PLP_Imp  uint8 = 0x28
	ACC_Imm  uint8 = 0x69
	ADC_Zp   uint8 = 0x65
	ADC_ZpX  uint8 = 0x75
	ADC_Abs  uint8 = 0x6D
	ADC_AbsX uint8 = 0x7D
	ABC_AbsY uint8 = 0x79
	ADS_IndX uint8 = 0x61
	ADC_IndY uint8 = 0x71
	SBC_Imm  uint8 = 0xE9
	SBC_ZP   uint8 = 0xE5
	SBC_ZpX  uint8 = 0xF5
	SBC_Abs  uint8 = 0xED
	SBC_AbsX uint8 = 0xFD
	SBC_AbsY uint8 = 0xF9
	SBC_IndX uint8 = 0xE1
	SBC_IndY uint8 = 0xF1
	CMP_Imm  uint8 = 0xC9
	CMP_Zp   uint8 = 0xC5
	CMP_ZpX  uint8 = 0xD5
	CMP_Abs  uint8 = 0xCD
	CMP_AbsX uint8 = 0xDD
	CMP_AbsY uint8 = 0xD9
	CMP_IndX uint8 = 0xC1
	CMP_IndY uint8 = 0xD1
	CPX_Imm  uint8 = 0xE0
	CPX_Zp   uint8 = 0xE4
	CPX_Abs  uint8 = 0xEC
	CPY_Imm  uint8 = 0xC0
	CPY_Zp   uint8 = 0xC4
	CPY_Abs  uint8 = 0xCC
	AND_Imm  uint8 = 0x29
	AND_Zp   uint8 = 0x25
	AND_ZpX  uint8 = 0x35
	AND_Abs  uint8 = 0x2D
	AND_AbsX uint8 = 0x3D
	AND_AbsY uint8 = 0x39
	AND_IndX uint8 = 0x21
	AND_IndY uint8 = 0x31
	EOR_Imm  uint8 = 0x49
	EOR_Zp   uint8 = 0x45
	EOR_ZpX  uint8 = 0x55
	EOR_Abs  uint8 = 0x4D
	EOR_AbsX uint8 = 0x5D
	EOR_AbsY uint8 = 0x59
	EOR_IndX uint8 = 0x41
	EOR_IndY uint8 = 0x51
	ORA_Imm  uint8 = 0x09
	ORA_Zp   uint8 = 0x05
	ORA_ZpX  uint8 = 0x15
	ORA_Abs  uint8 = 0x0D
	ORA_AbsX uint8 = 0x1D
	ORA_AbsY uint8 = 0x19
	ORA_IndX uint8 = 0x01
	ORA_IndY uint8 = 0x11
	BIT_Zp   uint8 = 0x24
	BIT_Abs  uint8 = 0x2C
	ASL_Acc  uint8 = 0x0A
	ASL_Zp   uint8 = 0x06
	ASL_ZpX  uint8 = 0x16
	ASL_Abs  uint8 = 0x0E
	ASL_AbsX uint8 = 0x1E
	LSR_Acc  uint8 = 0x4A
	LSR_Zp   uint8 = 0x46
	LSR_ZpX  uint8 = 0x56
	LSR_Abs  uint8 = 0x4E
	LSR_AbsX uint8 = 0x5E
	ROL_Acc  uint8 = 0x2A
	ROL_Zp   uint8 = 0x26
	ROL_ZpX  uint8 = 0x36
	ROL_Abs  uint8 = 0x2E
	ROL_AbsX uint8 = 0x3E
	ROR_Acc  uint8 = 0x6A
	ROR_Zp   uint8 = 0x66
	ROR_ZpX  uint8 = 0x76
	ROR_Abs  uint8 = 0x6E
	ROR_AbsX uint8 = 0x7E
	INC_Zp   uint8 = 0xE6
	INC_ZpX  uint8 = 0xF6
	INC_Abs  uint8 = 0xEE
	INC_AbsX uint8 = 0xFE
	DEC_Zp   uint8 = 0xC6
	DEC_ZpX  uint8 = 0xD6
	DEC_Abs  uint8 = 0xCE
	DEC_AbsX uint8 = 0xDE
	INX_Imp  uint8 = 0xE8
	INY_Imp  uint8 = 0xC8
	DEX_Imp  uint8 = 0xCA
	DEY_Imp  uint8 = 0x88
	BCC_Rel  uint8 = 0x90
	BCS_Rel  uint8 = 0xB0
	BEQ_Rel  uint8 = 0xF0
	BMI_Rel  uint8 = 0x30
	BNE_Rel  uint8 = 0xD0
	BPL_Rel  uint8 = 0x10
	BVC_Rel  uint8 = 0x50
	BVS_Rel  uint8 = 0x70
	BRK_Imp  uint8 = 0x00
	CLC_Imp  uint8 = 0x18
	CLD_Imp  uint8 = 0xD8
	CLI_Imp  uint8 = 0x58
	CLV_Imp  uint8 = 0xB8
	SEC_Imp  uint8 = 0x38
	SED_Imp  uint8 = 0xF8
	SEI_Imp  uint8 = 0x78
	JMP_Ind  uint8 = 0x6C
	JMP_Abs  uint8 = 0x4C
	RTI_Imp  uint8 = 0x40
	JSR_Abs  uint8 = 0x20
	RTS_Imp  uint8 = 0x60
)

var instrTable = []instrInfo{
	// Official 6502 instructions.
	{"NOP", NOP_Imp, AddrModeImp, 1, 2},
	{"BRK", BRK_Imp, AddrModeImp, 1, 7},
	{"LDA", LDA_Imm, AddrModeImm, 2, 2},
	{"LDA", LDA_Zp, AddrModeZp, 2, 3},
	{"LDA", LDA_ZpX, AddrModeZpX, 2, 4},
	{"LDA", LDA_Abs, AddrModeAbs, 3, 4},
	{"LDA", LDA_AbsX, AddrModeAbsX, 3, 4},
	{"LDA", LDA_AbsY, AddrModeAbsY, 3, 4},
	{"LDA", LDA_IndX, AddrModeIndX, 2, 6},
	{"LDA", LDA_IndY, AddrModeIndY, 2, 5},
	{"STA", STA_Zp, AddrModeZp, 2, 3},
	{"STA", STA_ZpX, AddrModeZpX, 2, 4},
	{"STA", STA_Abs, AddrModeAbs, 3, 4},
	{"STA", STA_AbsX, AddrModeAbsX, 3, 5},
	{"STA", STA_Ind, AddrModeAbsY, 3, 5},
	{"STA", STA_IndX, AddrModeIndX, 2, 6},
	{"STA", STA_IndY, AddrModeIndY, 2, 6},
	{"LDX", LDX_Imm, AddrModeImm, 2, 2},
	{"LDX", LDX_Zp, AddrModeZp, 2, 3},
	{"LDX", LDX_ZpY, AddrModeZpY, 2, 4},
	{"LDX", LDX_Abs, AddrModeAbs, 3, 4},
	{"LDX", LDX_AbsY, AddrModeAbsY, 3, 4},
	{"STX", STX_Zp, AddrModeZp, 2, 3},
	{"STX", STX_ZpY, AddrModeZpY, 2, 4},
	{"STX", STX_Abs, AddrModeAbs, 3, 4},
	{"LDY", LDY_Imm, AddrModeImm, 2, 2},
	{"LDY", LDY_Zp, AddrModeZp, 2, 3},
	{"LDY", LDY_ZpX, AddrModeZpX, 2, 4},
	{"LDY", LDY_Abs, AddrModeAbs, 3, 4},
	{"LDY", LDY_AbsX, AddrModeAbsX, 3, 4},
	{"STY", STY_Zp, AddrModeZp, 2, 3},
	{"STY", STY_ZpX, AddrModeZpX, 2, 4},
	{"STY", STY_Abs, AddrModeAbs, 3, 4},
	{"TAX", TAX_Imp, AddrModeImp, 1, 2},
	{"TAY", TAY_Imp, AddrModeImp, 1, 2},
	{"TXA", TXA_Imp, AddrModeImp, 1, 2},
	{"TYA", TYA_Imp, AddrModeImp, 1, 2},
	{"TSX", TSX_Imp, AddrModeImp, 1, 2},
	{"TXS", TXS_Imp, AddrModeImp, 1, 2},
	{"PHA", PHA_Imp, AddrModeImp, 1, 3},
	{"PHP", PHP_Imp, AddrModeImp, 1, 3},
	{"PLA", PLA_Imp, AddrModeImp, 1, 4},
	{"PLP", PLP_Imp, AddrModeImp, 1, 4},
	{"ADC", ACC_Imm, AddrModeImm, 2, 2},
	{"ADC", ADC_Zp, AddrModeZp, 2, 3},
	{"ADC", ADC_ZpX, AddrModeZpX, 2, 4},
	{"ADC", ADC_Abs, AddrModeAbs, 3, 4},
	{"ADC", ADC_AbsX, AddrModeAbsX, 3, 4},
	{"ADC", ABC_AbsY, AddrModeAbsY, 3, 4},
	{"ADC", ADS_IndX, AddrModeIndX, 2, 6},
	{"ADC", ADC_IndY, AddrModeIndY, 2, 5},
	{"SBC", SBC_Imm, AddrModeImm, 2, 2},
	{"SBC", SBC_ZP, AddrModeZp, 2, 3},
	{"SBC", SBC_ZpX, AddrModeZpX, 2, 4},
	{"SBC", SBC_Abs, AddrModeAbs, 3, 4},
	{"SBC", SBC_AbsX, AddrModeAbsX, 3, 4},
	{"SBC", SBC_AbsY, AddrModeAbsY, 3, 4},
	{"SBC", SBC_IndX, AddrModeIndX, 2, 6},
	{"SBC", SBC_IndY, AddrModeIndY, 2, 5},
	{"AND", AND_Imm, AddrModeImm, 2, 2},
	{"AND", AND_Zp, AddrModeZp, 2, 3},
	{"AND", AND_ZpX, AddrModeZpX, 2, 4},
	{"AND", AND_Abs, AddrModeAbs, 3, 4},
	{"AND", AND_AbsX, AddrModeAbsX, 3, 4},
	{"AND", AND_AbsY, AddrModeAbsY, 3, 4},
	{"AND", AND_IndX, AddrModeIndX, 2, 6},
	{"AND", AND_IndY, AddrModeIndY, 2, 5},
	{"ORA", ORA_Imm, AddrModeImm, 2, 2},
	{"ORA", ORA_Zp, AddrModeZp, 2, 3},
	{"ORA", ORA_ZpX, AddrModeZpX, 2, 4},
	{"ORA", ORA_Abs, AddrModeAbs, 3, 4},
	{"ORA", ORA_AbsX, AddrModeAbsX, 3, 4},
	{"ORA", ORA_AbsY, AddrModeAbsY, 3, 4},
	{"ORA", ORA_IndX, AddrModeIndX, 2, 6},
	{"ORA", ORA_IndY, AddrModeIndY, 2, 5},
	{"EOR", EOR_Imm, AddrModeImm, 2, 2},
	{"EOR", EOR_Zp, AddrModeZp, 2, 3},
	{"EOR", EOR_ZpX, AddrModeZpX, 2, 4},
	{"EOR", EOR_Abs, AddrModeAbs, 3, 4},
	{"EOR", EOR_AbsX, AddrModeAbsX, 3, 4},
	{"EOR", EOR_AbsY, AddrModeAbsY, 3, 4},
	{"EOR", EOR_IndX, AddrModeIndX, 2, 6},
	{"EOR", EOR_IndY, AddrModeIndY, 2, 5},
	{"CMP", CMP_Imm, AddrModeImm, 2, 2},
	{"CMP", CMP_Zp, AddrModeZp, 2, 3},
	{"CMP", CMP_ZpX, AddrModeZpX, 2, 4},
	{"CMP", CMP_Abs, AddrModeAbs, 3, 4},
	{"CMP", CMP_AbsX, AddrModeAbsX, 3, 4},
	{"CMP", CMP_AbsY, AddrModeAbsY, 3, 4},
	{"CMP", CMP_IndX, AddrModeIndX, 2, 6},
	{"CMP", CMP_IndY, AddrModeIndY, 2, 5},
	{"CPX", CPX_Imm, AddrModeImm, 2, 2},
	{"CPX", CPX_Zp, AddrModeZp, 2, 3},
	{"CPX", CPX_Abs, AddrModeAbs, 3, 4},
	{"CPY", CPY_Imm, AddrModeImm, 2, 2},
	{"CPY", CPY_Zp, AddrModeZp, 2, 3},
	{"CPY", CPY_Abs, AddrModeAbs, 3, 4},
	{"BIT", BIT_Zp, AddrModeZp, 2, 3},
	{"BIT", BIT_Abs, AddrModeAbs, 3, 4},
	{"ASL", ASL_Acc, AddrModeAcc, 1, 2},
	{"ASL", ASL_Zp, AddrModeZp, 2, 5},
	{"ASL", ASL_ZpX, AddrModeZpX, 2, 6},
	{"ASL", ASL_Abs, AddrModeAbs, 3, 6},
	{"ASL", ASL_AbsX, AddrModeAbsX, 3, 7},
	{"LSR", LSR_Acc, AddrModeAcc, 1, 2},
	{"LSR", LSR_Zp, AddrModeZp, 2, 5},
	{"LSR", LSR_ZpX, AddrModeZpX, 2, 6},
	{"LSR", LSR_Abs, AddrModeAbs, 3, 6},
	{"LSR", LSR_AbsX, AddrModeAbsX, 3, 7},
	{"ROL", ROL_Acc, AddrModeAcc, 1, 2},
	{"ROL", ROL_Zp, AddrModeZp, 2, 5},
	{"ROL", ROL_ZpX, AddrModeZpX, 2, 6},
	{"ROL", ROL_Abs, AddrModeAbs, 3, 6},
	{"ROL", ROL_AbsX, AddrModeAbsX, 3, 7},
	{"ROR", ROR_Acc, AddrModeAcc, 1, 2},
	{"ROR", ROR_Zp, AddrModeZp, 2, 5},
	{"ROR", ROR_ZpX, AddrModeZpX, 2, 6},
	{"ROR", ROR_Abs, AddrModeAbs, 3, 6},
	{"ROR", ROR_AbsX, AddrModeAbsX, 3, 7},
	{"INC", INC_Zp, AddrModeZp, 2, 5},
	{"INC", INC_ZpX, AddrModeZpX, 2, 6},
	{"INC", INC_Abs, AddrModeAbs, 3, 6},
	{"INC", INC_AbsX, AddrModeAbsX, 3, 7},
	{"INX", INX_Imp, AddrModeImp, 1, 2},
	{"INY", INY_Imp, AddrModeImp, 1, 2},
	{"DEX", DEX_Imp, AddrModeImp, 1, 2},
	{"DEY", DEY_Imp, AddrModeImp, 1, 2},
	{"DEC", DEC_Zp, AddrModeZp, 2, 5},
	{"DEC", DEC_ZpX, AddrModeZpX, 2, 6},
	{"DEC", DEC_Abs, AddrModeAbs, 3, 6},
	{"DEC", DEC_AbsX, AddrModeAbsX, 3, 7},
	{"JMP", JMP_Abs, AddrModeAbs, 3, 3},
	{"JMP", JMP_Ind, AddrModeInd, 3, 5},
	{"JSR", JSR_Abs, AddrModeAbs, 3, 6},
	{"RTS", RTS_Imp, AddrModeImp, 1, 6},
	{"RTI", RTI_Imp, AddrModeImp, 1, 6},
	{"BPL", BPL_Rel, AddrModeRel, 2, 2},
	{"BMI", BMI_Rel, AddrModeRel, 2, 2},
	{"BVC", BVC_Rel, AddrModeRel, 2, 2},
	{"BVS", BVS_Rel, AddrModeRel, 2, 2},
	{"BCC", BCC_Rel, AddrModeRel, 2, 2},
	{"BCS", BCS_Rel, AddrModeRel, 2, 2},
	{"BNE", BNE_Rel, AddrModeRel, 2, 2},
	{"BEQ", BEQ_Rel, AddrModeRel, 2, 2},
	{"SEC", SEC_Imp, AddrModeImp, 1, 2},
	{"SEI", SEI_Imp, AddrModeImp, 1, 2},
	{"CLC", CLC_Imp, AddrModeImp, 1, 2},
	{"SED", SED_Imp, AddrModeImp, 1, 2},
	{"CLI", CLI_Imp, AddrModeImp, 1, 2},
	{"CLV", CLV_Imp, AddrModeImp, 1, 2},
	{"CLD", CLD_Imp, AddrModeImp, 1, 2},

	// The unofficial opcodes, used by some games and nestest.
	// https://www.masswerk.at/nowgobang/2021/6502-illegal-opcodes
	{"*SBC", 0xEB, AddrModeImm, 2, 2},
	{"*DCP", 0xC7, AddrModeZp, 2, 5},
	{"*DCP", 0xD7, AddrModeZpX, 2, 6},
	{"*DCP", 0xCF, AddrModeAbs, 3, 6},
	{"*DCP", 0xDF, AddrModeAbsX, 3, 7},
	{"*DCP", 0xDB, AddrModeAbsY, 3, 7},
	{"*DCP", 0xC3, AddrModeIndX, 2, 8},
	{"*DCP", 0xD3, AddrModeIndY, 2, 8},
	{"*LAX", 0xA7, AddrModeZp, 2, 3},
	{"*LAX", 0xB7, AddrModeZpY, 2, 4},
	{"*LAX", 0xAF, AddrModeAbs, 3, 4},
	{"*LAX", 0xBF, AddrModeAbsY, 3, 4}, // +1 if page crossed
	{"*LAX", 0xA3, AddrModeIndX, 2, 6},
	{"*LAX", 0xB3, AddrModeIndY, 2, 5}, // +1 if page crossed
	{"*NOP", 0x1A, AddrModeImp, 1, 2},
	{"*NOP", 0x3A, AddrModeImp, 1, 2},
	{"*NOP", 0x5A, AddrModeImp, 1, 2},
	{"*NOP", 0x7A, AddrModeImp, 1, 2},
	{"*NOP", 0xDA, AddrModeImp, 1, 2},
	{"*NOP", 0xFA, AddrModeImp, 1, 2},
	{"*NOP", 0x80, AddrModeImm, 2, 2},
	{"*NOP", 0x82, AddrModeImm, 2, 2},
	{"*NOP", 0x89, AddrModeImm, 2, 2},
	{"*NOP", 0xC2, AddrModeImm, 2, 2},
	{"*NOP", 0xE2, AddrModeImm, 2, 2},
	{"*NOP", 0x04, AddrModeZp, 2, 3},
	{"*NOP", 0x44, AddrModeZp, 2, 3},
	{"*NOP", 0x64, AddrModeZp, 2, 3},
	{"*NOP", 0x14, AddrModeZpX, 2, 4},
	{"*NOP", 0x34, AddrModeZpX, 2, 4},
	{"*NOP", 0x54, AddrModeZpX, 2, 4},
	{"*NOP", 0x74, AddrModeZpX, 2, 4},
	{"*NOP", 0xD4, AddrModeZpX, 2, 4},
	{"*NOP", 0xF4, AddrModeZpX, 2, 4},
	{"*NOP", 0x0C, AddrModeAbs, 3, 4},
	{"*NOP", 0x1C, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{"*NOP", 0x3C, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{"*NOP", 0x5C, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{"*NOP", 0x7C, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{"*NOP", 0xDC, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{"*NOP", 0xFC, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{"*RLA", 0x27, AddrModeZp, 2, 5},
	{"*RLA", 0x37, AddrModeZpX, 2, 6},
	{"*RLA", 0x2F, AddrModeAbs, 3, 6},
	{"*RLA", 0x3F, AddrModeAbsX, 3, 7},
	{"*RLA", 0x3B, AddrModeAbsY, 3, 7},
	{"*RLA", 0x23, AddrModeIndX, 2, 8},
	{"*RLA", 0x33, AddrModeIndY, 2, 8},
	{"*RRA", 0x67, AddrModeZp, 2, 5},
	{"*RRA", 0x77, AddrModeZpX, 2, 6},
	{"*RRA", 0x6F, AddrModeAbs, 3, 6},
	{"*RRA", 0x7F, AddrModeAbsX, 3, 7},
	{"*RRA", 0x7B, AddrModeAbsY, 3, 7},
	{"*RRA", 0x63, AddrModeIndX, 2, 8},
	{"*RRA", 0x73, AddrModeIndY, 2, 8},
	{"*SAX", 0x87, AddrModeZp, 2, 3},
	{"*SAX", 0x97, AddrModeZpY, 2, 4},
	{"*SAX", 0x8F, AddrModeAbs, 3, 4},
	{"*SAX", 0x83, AddrModeIndX, 2, 6},
	{"*SLO", 0x07, AddrModeZp, 2, 5},
	{"*SLO", 0x17, AddrModeZpX, 2, 6},
	{"*SLO", 0x0F, AddrModeAbs, 3, 6},
	{"*SLO", 0x1F, AddrModeAbsX, 3, 7},
	{"*SLO", 0x1B, AddrModeAbsY, 3, 7},
	{"*SLO", 0x03, AddrModeIndX, 2, 8},
	{"*SLO", 0x13, AddrModeIndY, 2, 8},
	{"*SRE", 0x47, AddrModeZp, 2, 5},
	{"*SRE", 0x57, AddrModeZpX, 2, 6},
	{"*SRE", 0x4F, AddrModeAbs, 3, 6},
	{"*SRE", 0x5F, AddrModeAbsX, 3, 7},
	{"*SRE", 0x5B, AddrModeAbsY, 3, 7},
	{"*SRE", 0x43, AddrModeIndX, 2, 8},
	{"*SRE", 0x53, AddrModeIndY, 2, 8},
	{"*ISB", 0xE7, AddrModeZp, 2, 5},
	{"*ISB", 0xF7, AddrModeZpX, 2, 6},
	{"*ISB", 0xEF, AddrModeAbs, 3, 6},
	{"*ISB", 0xFF, AddrModeAbsX, 3, 7},
	{"*ISB", 0xFB, AddrModeAbsY, 3, 7},
	{"*ISB", 0xE3, AddrModeIndX, 2, 8},
	{"*ISB", 0xF3, AddrModeIndY, 2, 8},

	// The rest of the illegal opcodes are not emulated, but are
	// included here for disassembly just in case.
	{"???", 0x0B, AddrModeImm, 2, 2},
	{"???", 0x2B, AddrModeImm, 2, 2},
	{"???", 0x4B, AddrModeImm, 2, 2},
	{"???", 0x6B, AddrModeImm, 2, 2},
	{"???", 0x8B, AddrModeImm, 2, 2},
	{"???", 0xBB, AddrModeImm, 2, 2},
	{"???", 0xAB, AddrModeImm, 2, 2},
	{"???", 0xCB, AddrModeImm, 2, 2},
	{"???", 0x9F, AddrModeAbsX, 3, 5},
	{"???", 0x93, AddrModeIndY, 2, 6},
	{"???", 0x9E, AddrModeAbsY, 3, 5},
	{"???", 0x9C, AddrModeAbs, 3, 5},
	{"???", 0x9B, AddrModeImm, 2, 2},
}
