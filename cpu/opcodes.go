package cpu

var (
	Instructions map[uint8]Instruction
)

func init() {
	Instructions = make(map[uint8]Instruction)
	for _, instr := range instructionTable {
		Instructions[instr.Opcode] = instr
	}
}

type Instruction struct {
	ID       InstrID
	Name     string
	Opcode   uint8
	AddrMode AddrMode
	Size     int
	Cycles   int
}

const (
	NOP InstrID = iota
	LDA
	STA
	LDX
	STX
	LDY
	STY
	TAX
	TAY
	TXA
	TYA
	TSX
	TXS
	PHA
	PLA
	PHP
	PLP
	ADC
	SBC
	CMP
	CPX
	CPY
	AND
	EOR
	ORA
	BIT
	ASL
	LSR
	ROL
	ROR
	INC
	DEC
	INX
	INY
	DEX
	DEY
	BCC
	BCS
	BEQ
	BMI
	BNE
	BPL
	BVC
	BVS
	BRK
	CLC
	CLD
	CLI
	CLV
	SEC
	SED
	SEI
	JMP
	RTI
	JSR
	RTS

	XSBC
	XDCP
	XLAX
	XNOP
	XRLA
	XRRA
	XSLO
	XSRE
	XSAX
	XISB
)

var instructionTable = []Instruction{
	// Official 6502 instructions.
	{NOP, "NOP", 0xEA, AddrModeImp, 1, 2},
	{BRK, "BRK", 0x00, AddrModeImp, 2, 7},
	{LDA, "LDA", 0xA9, AddrModeImm, 2, 2},
	{LDA, "LDA", 0xA5, AddrModeZp, 2, 3},
	{LDA, "LDA", 0xB5, AddrModeZpX, 2, 4},
	{LDA, "LDA", 0xAD, AddrModeAbs, 3, 4},
	{LDA, "LDA", 0xBD, AddrModeAbsX, 3, 4},
	{LDA, "LDA", 0xB9, AddrModeAbsY, 3, 4},
	{LDA, "LDA", 0xA1, AddrModeIndX, 2, 6},
	{LDA, "LDA", 0xB1, AddrModeIndY, 2, 5},
	{STA, "STA", 0x85, AddrModeZp, 2, 3},
	{STA, "STA", 0x95, AddrModeZpX, 2, 4},
	{STA, "STA", 0x8D, AddrModeAbs, 3, 4},
	{STA, "STA", 0x9D, AddrModeAbsX, 3, 5},
	{STA, "STA", 0x99, AddrModeAbsY, 3, 5},
	{STA, "STA", 0x81, AddrModeIndX, 2, 6},
	{STA, "STA", 0x91, AddrModeIndY, 2, 6},
	{LDX, "LDX", 0xA2, AddrModeImm, 2, 2},
	{LDX, "LDX", 0xA6, AddrModeZp, 2, 3},
	{LDX, "LDX", 0xB6, AddrModeZpY, 2, 4},
	{LDX, "LDX", 0xAE, AddrModeAbs, 3, 4},
	{LDX, "LDX", 0xBE, AddrModeAbsY, 3, 4},
	{STX, "STX", 0x86, AddrModeZp, 2, 3},
	{STX, "STX", 0x96, AddrModeZpY, 2, 4},
	{STX, "STX", 0x8E, AddrModeAbs, 3, 4},
	{LDY, "LDY", 0xA0, AddrModeImm, 2, 2},
	{LDY, "LDY", 0xA4, AddrModeZp, 2, 3},
	{LDY, "LDY", 0xB4, AddrModeZpX, 2, 4},
	{LDY, "LDY", 0xAC, AddrModeAbs, 3, 4},
	{LDY, "LDY", 0xBC, AddrModeAbsX, 3, 4},
	{STY, "STY", 0x84, AddrModeZp, 2, 3},
	{STY, "STY", 0x94, AddrModeZpX, 2, 4},
	{STY, "STY", 0x8C, AddrModeAbs, 3, 4},
	{TAX, "TAX", 0xAA, AddrModeImp, 1, 2},
	{TAY, "TAY", 0xA8, AddrModeImp, 1, 2},
	{TXA, "TXA", 0x8A, AddrModeImp, 1, 2},
	{TYA, "TYA", 0x98, AddrModeImp, 1, 2},
	{TSX, "TSX", 0xBA, AddrModeImp, 1, 2},
	{TXS, "TXS", 0x9A, AddrModeImp, 1, 2},
	{PHA, "PHA", 0x48, AddrModeImp, 1, 3},
	{PHP, "PHP", 0x08, AddrModeImp, 1, 3},
	{PLA, "PLA", 0x68, AddrModeImp, 1, 4},
	{PLP, "PLP", 0x28, AddrModeImp, 1, 4},
	{ADC, "ADC", 0x69, AddrModeImm, 2, 2},
	{ADC, "ADC", 0x65, AddrModeZp, 2, 3},
	{ADC, "ADC", 0x75, AddrModeZpX, 2, 4},
	{ADC, "ADC", 0x6D, AddrModeAbs, 3, 4},
	{ADC, "ADC", 0x7D, AddrModeAbsX, 3, 4},
	{ADC, "ADC", 0x79, AddrModeAbsY, 3, 4},
	{ADC, "ADC", 0x61, AddrModeIndX, 2, 6},
	{ADC, "ADC", 0x71, AddrModeIndY, 2, 5},
	{SBC, "SBC", 0xE9, AddrModeImm, 2, 2},
	{SBC, "SBC", 0xE5, AddrModeZp, 2, 3},
	{SBC, "SBC", 0xF5, AddrModeZpX, 2, 4},
	{SBC, "SBC", 0xED, AddrModeAbs, 3, 4},
	{SBC, "SBC", 0xFD, AddrModeAbsX, 3, 4},
	{SBC, "SBC", 0xF9, AddrModeAbsY, 3, 4},
	{SBC, "SBC", 0xE1, AddrModeIndX, 2, 6},
	{SBC, "SBC", 0xF1, AddrModeIndY, 2, 5},
	{AND, "AND", 0x29, AddrModeImm, 2, 2},
	{AND, "AND", 0x25, AddrModeZp, 2, 3},
	{AND, "AND", 0x35, AddrModeZpX, 2, 4},
	{AND, "AND", 0x2D, AddrModeAbs, 3, 4},
	{AND, "AND", 0x3D, AddrModeAbsX, 3, 4},
	{AND, "AND", 0x39, AddrModeAbsY, 3, 4},
	{AND, "AND", 0x21, AddrModeIndX, 2, 6},
	{AND, "AND", 0x31, AddrModeIndY, 2, 5},
	{ORA, "ORA", 0x09, AddrModeImm, 2, 2},
	{ORA, "ORA", 0x05, AddrModeZp, 2, 3},
	{ORA, "ORA", 0x15, AddrModeZpX, 2, 4},
	{ORA, "ORA", 0x0D, AddrModeAbs, 3, 4},
	{ORA, "ORA", 0x1D, AddrModeAbsX, 3, 4},
	{ORA, "ORA", 0x19, AddrModeAbsY, 3, 4},
	{ORA, "ORA", 0x01, AddrModeIndX, 2, 6},
	{ORA, "ORA", 0x11, AddrModeIndY, 2, 5},
	{EOR, "EOR", 0x49, AddrModeImm, 2, 2},
	{EOR, "EOR", 0x45, AddrModeZp, 2, 3},
	{EOR, "EOR", 0x55, AddrModeZpX, 2, 4},
	{EOR, "EOR", 0x4D, AddrModeAbs, 3, 4},
	{EOR, "EOR", 0x5D, AddrModeAbsX, 3, 4},
	{EOR, "EOR", 0x59, AddrModeAbsY, 3, 4},
	{EOR, "EOR", 0x41, AddrModeIndX, 2, 6},
	{EOR, "EOR", 0x51, AddrModeIndY, 2, 5},
	{CMP, "CMP", 0xC9, AddrModeImm, 2, 2},
	{CMP, "CMP", 0xC5, AddrModeZp, 2, 3},
	{CMP, "CMP", 0xD5, AddrModeZpX, 2, 4},
	{CMP, "CMP", 0xCD, AddrModeAbs, 3, 4},
	{CMP, "CMP", 0xDD, AddrModeAbsX, 3, 4},
	{CMP, "CMP", 0xD9, AddrModeAbsY, 3, 4},
	{CMP, "CMP", 0xC1, AddrModeIndX, 2, 6},
	{CMP, "CMP", 0xD1, AddrModeIndY, 2, 5},
	{CPX, "CPX", 0xE0, AddrModeImm, 2, 2},
	{CPX, "CPX", 0xE4, AddrModeZp, 2, 3},
	{CPX, "CPX", 0xEC, AddrModeAbs, 3, 4},
	{CPY, "CPY", 0xC0, AddrModeImm, 2, 2},
	{CPY, "CPY", 0xC4, AddrModeZp, 2, 3},
	{CPY, "CPY", 0xCC, AddrModeAbs, 3, 4},
	{BIT, "BIT", 0x24, AddrModeZp, 2, 3},
	{BIT, "BIT", 0x2C, AddrModeAbs, 3, 4},
	{ASL, "ASL", 0x0A, AddrModeAcc, 1, 2},
	{ASL, "ASL", 0x06, AddrModeZp, 2, 5},
	{ASL, "ASL", 0x16, AddrModeZpX, 2, 6},
	{ASL, "ASL", 0x0E, AddrModeAbs, 3, 6},
	{ASL, "ASL", 0x1E, AddrModeAbsX, 3, 7},
	{LSR, "LSR", 0x4A, AddrModeAcc, 1, 2},
	{LSR, "LSR", 0x46, AddrModeZp, 2, 5},
	{LSR, "LSR", 0x56, AddrModeZpX, 2, 6},
	{LSR, "LSR", 0x4E, AddrModeAbs, 3, 6},
	{LSR, "LSR", 0x5E, AddrModeAbsX, 3, 7},
	{ROL, "ROL", 0x2A, AddrModeAcc, 1, 2},
	{ROL, "ROL", 0x26, AddrModeZp, 2, 5},
	{ROL, "ROL", 0x36, AddrModeZpX, 2, 6},
	{ROL, "ROL", 0x2E, AddrModeAbs, 3, 6},
	{ROL, "ROL", 0x3E, AddrModeAbsX, 3, 7},
	{ROR, "ROR", 0x6A, AddrModeAcc, 1, 2},
	{ROR, "ROR", 0x66, AddrModeZp, 2, 5},
	{ROR, "ROR", 0x76, AddrModeZpX, 2, 6},
	{ROR, "ROR", 0x6E, AddrModeAbs, 3, 6},
	{ROR, "ROR", 0x7E, AddrModeAbsX, 3, 7},
	{INC, "INC", 0xE6, AddrModeZp, 2, 5},
	{INC, "INC", 0xF6, AddrModeZpX, 2, 6},
	{INC, "INC", 0xEE, AddrModeAbs, 3, 6},
	{INC, "INC", 0xFE, AddrModeAbsX, 3, 7},
	{INX, "INX", 0xE8, AddrModeImp, 1, 2},
	{INY, "INY", 0xC8, AddrModeImp, 1, 2},
	{DEX, "DEX", 0xCA, AddrModeImp, 1, 2},
	{DEY, "DEY", 0x88, AddrModeImp, 1, 2},
	{DEC, "DEC", 0xC6, AddrModeZp, 2, 5},
	{DEC, "DEC", 0xD6, AddrModeZpX, 2, 6},
	{DEC, "DEC", 0xCE, AddrModeAbs, 3, 6},
	{DEC, "DEC", 0xDE, AddrModeAbsX, 3, 7},
	{JMP, "JMP", 0x4C, AddrModeAbs, 3, 3},
	{JMP, "JMP", 0x6C, AddrModeInd, 3, 5},
	{JSR, "JSR", 0x20, AddrModeAbs, 3, 6},
	{RTS, "RTS", 0x60, AddrModeImp, 1, 6},
	{RTI, "RTI", 0x40, AddrModeImp, 1, 6},
	{BPL, "BPL", 0x10, AddrModeRel, 2, 2},
	{BMI, "BMI", 0x30, AddrModeRel, 2, 2},
	{BVC, "BVC", 0x50, AddrModeRel, 2, 2},
	{BVS, "BVS", 0x70, AddrModeRel, 2, 2},
	{BCC, "BCC", 0x90, AddrModeRel, 2, 2},
	{BCS, "BCS", 0xB0, AddrModeRel, 2, 2},
	{BNE, "BNE", 0xD0, AddrModeRel, 2, 2},
	{BEQ, "BEQ", 0xF0, AddrModeRel, 2, 2},
	{SEC, "SEC", 0x38, AddrModeImp, 1, 2},
	{SEI, "SEI", 0x78, AddrModeImp, 1, 2},
	{CLC, "CLC", 0x18, AddrModeImp, 1, 2},
	{SED, "SED", 0xF8, AddrModeImp, 1, 2},
	{CLI, "CLI", 0x58, AddrModeImp, 1, 2},
	{CLV, "CLV", 0xB8, AddrModeImp, 1, 2},
	{CLD, "CLD", 0xD8, AddrModeImp, 1, 2},

	// The unofficial opcodes, used by some games and nestest.
	// https://www.masswerk.at/nowgobang/2021/6502-illegal-opcodes
	{XSBC, "*SBC", 0xEB, AddrModeImm, 2, 2},
	{XDCP, "*DCP", 0xC7, AddrModeZp, 2, 5},
	{XDCP, "*DCP", 0xD7, AddrModeZpX, 2, 6},
	{XDCP, "*DCP", 0xCF, AddrModeAbs, 3, 6},
	{XDCP, "*DCP", 0xDF, AddrModeAbsX, 3, 7},
	{XDCP, "*DCP", 0xDB, AddrModeAbsY, 3, 7},
	{XDCP, "*DCP", 0xC3, AddrModeIndX, 2, 8},
	{XDCP, "*DCP", 0xD3, AddrModeIndY, 2, 8},
	{XLAX, "*LAX", 0xA7, AddrModeZp, 2, 3},
	{XLAX, "*LAX", 0xB7, AddrModeZpY, 2, 4},
	{XLAX, "*LAX", 0xAF, AddrModeAbs, 3, 4},
	{XLAX, "*LAX", 0xBF, AddrModeAbsY, 3, 4}, // +1 if page crossed
	{XLAX, "*LAX", 0xA3, AddrModeIndX, 2, 6},
	{XLAX, "*LAX", 0xB3, AddrModeIndY, 2, 5}, // +1 if page crossed
	{XNOP, "*NOP", 0x1A, AddrModeImp, 1, 2},
	{XNOP, "*NOP", 0x3A, AddrModeImp, 1, 2},
	{XNOP, "*NOP", 0x5A, AddrModeImp, 1, 2},
	{XNOP, "*NOP", 0x7A, AddrModeImp, 1, 2},
	{XNOP, "*NOP", 0xDA, AddrModeImp, 1, 2},
	{XNOP, "*NOP", 0xFA, AddrModeImp, 1, 2},
	{XNOP, "*NOP", 0x80, AddrModeImm, 2, 2},
	{XNOP, "*NOP", 0x82, AddrModeImm, 2, 2},
	{XNOP, "*NOP", 0x89, AddrModeImm, 2, 2},
	{XNOP, "*NOP", 0xC2, AddrModeImm, 2, 2},
	{XNOP, "*NOP", 0xE2, AddrModeImm, 2, 2},
	{XNOP, "*NOP", 0x04, AddrModeZp, 2, 3},
	{XNOP, "*NOP", 0x44, AddrModeZp, 2, 3},
	{XNOP, "*NOP", 0x64, AddrModeZp, 2, 3},
	{XNOP, "*NOP", 0x14, AddrModeZpX, 2, 4},
	{XNOP, "*NOP", 0x34, AddrModeZpX, 2, 4},
	{XNOP, "*NOP", 0x54, AddrModeZpX, 2, 4},
	{XNOP, "*NOP", 0x74, AddrModeZpX, 2, 4},
	{XNOP, "*NOP", 0xD4, AddrModeZpX, 2, 4},
	{XNOP, "*NOP", 0xF4, AddrModeZpX, 2, 4},
	{XNOP, "*NOP", 0x0C, AddrModeAbs, 3, 4},
	{XNOP, "*NOP", 0x1C, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{XNOP, "*NOP", 0x3C, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{XNOP, "*NOP", 0x5C, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{XNOP, "*NOP", 0x7C, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{XNOP, "*NOP", 0xDC, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{XNOP, "*NOP", 0xFC, AddrModeAbsX, 3, 4}, // +1 if page crossed
	{XRLA, "*RLA", 0x27, AddrModeZp, 2, 5},
	{XRLA, "*RLA", 0x37, AddrModeZpX, 2, 6},
	{XRLA, "*RLA", 0x2F, AddrModeAbs, 3, 6},
	{XRLA, "*RLA", 0x3F, AddrModeAbsX, 3, 7},
	{XRLA, "*RLA", 0x3B, AddrModeAbsY, 3, 7},
	{XRLA, "*RLA", 0x23, AddrModeIndX, 2, 8},
	{XRLA, "*RLA", 0x33, AddrModeIndY, 2, 8},
	{XRRA, "*RRA", 0x67, AddrModeZp, 2, 5},
	{XRRA, "*RRA", 0x77, AddrModeZpX, 2, 6},
	{XRRA, "*RRA", 0x6F, AddrModeAbs, 3, 6},
	{XRRA, "*RRA", 0x7F, AddrModeAbsX, 3, 7},
	{XRRA, "*RRA", 0x7B, AddrModeAbsY, 3, 7},
	{XRRA, "*RRA", 0x63, AddrModeIndX, 2, 8},
	{XRRA, "*RRA", 0x73, AddrModeIndY, 2, 8},
	{XSAX, "*SAX", 0x87, AddrModeZp, 2, 3},
	{XSAX, "*SAX", 0x97, AddrModeZpY, 2, 4},
	{XSAX, "*SAX", 0x8F, AddrModeAbs, 3, 4},
	{XSAX, "*SAX", 0x83, AddrModeIndX, 2, 6},
	{XSLO, "*SLO", 0x07, AddrModeZp, 2, 5},
	{XSLO, "*SLO", 0x17, AddrModeZpX, 2, 6},
	{XSLO, "*SLO", 0x0F, AddrModeAbs, 3, 6},
	{XSLO, "*SLO", 0x1F, AddrModeAbsX, 3, 7},
	{XSLO, "*SLO", 0x1B, AddrModeAbsY, 3, 7},
	{XSLO, "*SLO", 0x03, AddrModeIndX, 2, 8},
	{XSLO, "*SLO", 0x13, AddrModeIndY, 2, 8},
	{XSRE, "*SRE", 0x47, AddrModeZp, 2, 5},
	{XSRE, "*SRE", 0x57, AddrModeZpX, 2, 6},
	{XSRE, "*SRE", 0x4F, AddrModeAbs, 3, 6},
	{XSRE, "*SRE", 0x5F, AddrModeAbsX, 3, 7},
	{XSRE, "*SRE", 0x5B, AddrModeAbsY, 3, 7},
	{XSRE, "*SRE", 0x43, AddrModeIndX, 2, 8},
	{XSRE, "*SRE", 0x53, AddrModeIndY, 2, 8},
	{XISB, "*ISB", 0xE7, AddrModeZp, 2, 5},
	{XISB, "*ISB", 0xF7, AddrModeZpX, 2, 6},
	{XISB, "*ISB", 0xEF, AddrModeAbs, 3, 6},
	{XISB, "*ISB", 0xFF, AddrModeAbsX, 3, 7},
	{XISB, "*ISB", 0xFB, AddrModeAbsY, 3, 7},
	{XISB, "*ISB", 0xE3, AddrModeIndX, 2, 8},
	{XISB, "*ISB", 0xF3, AddrModeIndY, 2, 8},

	// The rest of the illegal opcodes are not emulated, but are
	// included here for disassembly just in case.
	{XNOP, "???", 0x0B, AddrModeImm, 2, 2},
	{XNOP, "???", 0x2B, AddrModeImm, 2, 2},
	{XNOP, "???", 0x4B, AddrModeImm, 2, 2},
	{XNOP, "???", 0x6B, AddrModeImm, 2, 2},
	{XNOP, "???", 0x8B, AddrModeImm, 2, 2},
	{XNOP, "???", 0xBB, AddrModeImm, 2, 2},
	{XNOP, "???", 0xAB, AddrModeImm, 2, 2},
	{XNOP, "???", 0xCB, AddrModeImm, 2, 2},
	{XNOP, "???", 0x9F, AddrModeAbsX, 3, 5},
	{XNOP, "???", 0x93, AddrModeIndY, 2, 6},
	{XNOP, "???", 0x9E, AddrModeAbsY, 3, 5},
	{XNOP, "???", 0x9C, AddrModeAbs, 3, 5},
	{XNOP, "???", 0x9B, AddrModeImm, 2, 2},
}
