package cpu

const (
	// NOP - No Operation
	OpNopImp Byte = 0xEA

	// RTI - Return from Interrupt
	OpRtiImp Byte = 0x40

	// LDA - Load Accumulator
	OpLdaImm  Byte = 0xA9
	OpLdaZp   Byte = 0xA5
	OpLdaZpX  Byte = 0xB5
	OpLdaAbs  Byte = 0xAD
	OpLdaAbsX Byte = 0xBD
	OpLdaAbsY Byte = 0xB9
	OpLdaIndX Byte = 0xA1
	OpLdaIndY Byte = 0xB1

	// STA - Store Accumulator
	OpStaZp   Byte = 0x85
	OpStaZpX  Byte = 0x95
	OpStaAbs  Byte = 0x8D
	OpStaAbsX Byte = 0x9D
	OpStaAbsY Byte = 0x99
	OpStaIndX Byte = 0x81
	OpStaIndY Byte = 0x91

	// LDX - Load X Register
	OpLdxImm  Byte = 0xA2
	OpLdxZp   Byte = 0xA6
	OpLdxZpY  Byte = 0xB6
	OpLdxAbs  Byte = 0xAE
	OpLdxAbsY Byte = 0xBE

	// STX - Store X Register
	OpStxZp  Byte = 0x86
	OpStxZpY Byte = 0x96
	OpStxAbs Byte = 0x8E

	// LDY - Load Y Register
	OpLdyImm  Byte = 0xA0
	OpLdyZp   Byte = 0xA4
	OpLdyZpX  Byte = 0xB4
	OpLdyAbs  Byte = 0xAC
	OpLdyAbsX Byte = 0xBC

	// STY - Store Y Register
	OpStyZp  Byte = 0x84
	OpStyZpX Byte = 0x94
	OpStyAbs Byte = 0x8C

	// TAX - Transfer Accumulator to X
	OpTaxImp Byte = 0xAA

	// TAY - Transfer Accumulator to Y
	OpTayImp Byte = 0xA8

	// TXA - Transfer X to Accumulator
	OpTxaImp Byte = 0x8A

	// TYA - Transfer Y to Accumulator
	OpTyaImp Byte = 0x98

	// TSX - Transfer Stack Pointer to X
	OpTsxImp Byte = 0xBA

	// TXS - Transfer X to Stack Pointer
	OpTxsImp Byte = 0x9A

	// PHA - Push Accumulator
	OpPhaImp Byte = 0x48

	// PLA - Pull Accumulator
	OpPlaImp Byte = 0x68

	// PHP - Push Processor Status
	OpPhpImp Byte = 0x08

	// PLP - Pull Processor Status
	OpPlpImp Byte = 0x28

	// ADC - Add with Carry
	OpAdcImm  Byte = 0x69
	OpAdcZp   Byte = 0x65
	OpAdcZpX  Byte = 0x75
	OpAdcAbs  Byte = 0x6D
	OpAdcAbsX Byte = 0x7D
	OpAdcAbsY Byte = 0x79
	OpAdcIndX Byte = 0x61
	OpAdcIndY Byte = 0x71

	// SBC - Subtract with Carry
	OpSbcImm  Byte = 0xE9
	OpSbcZp   Byte = 0xE5
	OpSbcZpX  Byte = 0xF5
	OpSbcAbs  Byte = 0xED
	OpSbcAbsX Byte = 0xFD
	OpSbcAbsY Byte = 0xF9
	OpSbcIndX Byte = 0xE1
	OpSbcIndY Byte = 0xF1

	// CMP - Compare Accumulator
	OpCmpImm  Byte = 0xC9
	OpCmpZp   Byte = 0xC5
	OpCmpZpX  Byte = 0xD5
	OpCmpAbs  Byte = 0xCD
	OpCmpAbsX Byte = 0xDD
	OpCmpAbsY Byte = 0xD9
	OpCmpIndX Byte = 0xC1
	OpCmpIndY Byte = 0xD1

	// CPX - Compare X Register
	OpCpxImm Byte = 0xE0
	OpCpxZp  Byte = 0xE4
	OpCpxAbs Byte = 0xEC

	// CPY - Compare Y Register
	OpCpyImm Byte = 0xC0
	OpCpyZp  Byte = 0xC4
	OpCpyAbs Byte = 0xCC

	// AND - Logical AND
	OpAndImm  Byte = 0x29
	OpAndZp   Byte = 0x25
	OpAndZpX  Byte = 0x35
	OpAndAbs  Byte = 0x2D
	OpAndAbsX Byte = 0x3D
	OpAndAbsY Byte = 0x39
	OpAndIndX Byte = 0x21
	OpAndIndY Byte = 0x31

	// EOR - Exclusive OR
	OpEorImm  Byte = 0x49
	OpEorZp   Byte = 0x45
	OpEorZpX  Byte = 0x55
	OpEorAbs  Byte = 0x4D
	OpEorAbsX Byte = 0x5D
	OpEorAbsY Byte = 0x59
	OpEorIndX Byte = 0x41
	OpEorIndY Byte = 0x51

	// ORA - Logical Inclusive OR
	OpOraImm  Byte = 0x09
	OpOraZp   Byte = 0x05
	OpOraZpX  Byte = 0x15
	OpOraAbs  Byte = 0x0D
	OpOraAbsX Byte = 0x1D
	OpOraAbsY Byte = 0x19
	OpOraIndX Byte = 0x01
	OpOraIndY Byte = 0x11

	// BIT - Bit Test
	OpBitZp  Byte = 0x24
	OpBitAbs Byte = 0x2C

	// ASL - Arithmetic Shift Left
	OpAslAcc  Byte = 0x0A
	OpAslZp   Byte = 0x06
	OpAslZpX  Byte = 0x16
	OpAslAbs  Byte = 0x0E
	OpAslAbsX Byte = 0x1E

	// LSR - Logical Shift Right
	OpLsrAcc  Byte = 0x4A
	OpLsrZp   Byte = 0x46
	OpLsrZpX  Byte = 0x56
	OpLsrAbs  Byte = 0x4E
	OpLsrAbsX Byte = 0x5E

	// ROL - Rotate Left
	OpRolAcc  Byte = 0x2A
	OpRolZp   Byte = 0x26
	OpRolZpX  Byte = 0x36
	OpRolAbs  Byte = 0x2E
	OpRolAbsX Byte = 0x3E

	// ROR - Rotate Right
	OpRorAcc  Byte = 0x6A
	OpRorZp   Byte = 0x66
	OpRorZpX  Byte = 0x76
	OpRorAbs  Byte = 0x6E
	OpRorAbsX Byte = 0x7E

	// INC - Increment Memory
	OpIncZp   Byte = 0xE6
	OpIncZpX  Byte = 0xF6
	OpIncAbs  Byte = 0xEE
	OpIncAbsX Byte = 0xFE

	// DEC - Decrement Memory
	OpDecZp   Byte = 0xC6
	OpDecZpX  Byte = 0xD6
	OpDecAbs  Byte = 0xCE
	OpDecAbsX Byte = 0xDE

	// INX - Increment X Register
	OpInxImp Byte = 0xE8

	// INY - Increment Y Register
	OpInyImp Byte = 0xC8

	// DEX - Decrement X Register
	OpDexImp Byte = 0xCA

	// DEY - Decrement Y Register
	OpDeyImp Byte = 0x88

	// BCC - Branch on Carry Clear
	OpBccRel Byte = 0x90

	// BCS - Branch on Carry Set
	OpBcsRel Byte = 0xB0

	// BEQ - Branch on Result Zero
	OpBeqRel Byte = 0xF0

	// BMI - Branch on Result Minus
	OpBmiRel Byte = 0x30

	// BNE - Branch on Result not Zero
	OpBneRel Byte = 0xD0

	// BPL - Branch on Result Plus
	OpBplRel Byte = 0x10

	// BVC - Branch on Overflow Clear
	OpBvcRel Byte = 0x50

	// BVS - Branch on Overflow Set
	OpBvsRel Byte = 0x70

	// BRK - Force Break
	OpBrkImp Byte = 0x00

	// CLC - Clear Carry Flag
	OpClcImp Byte = 0x18

	// CLD - Clear Decimal Mode
	OpCldImp Byte = 0xD8

	// CLI - Clear Interrupt Disable
	OpCliImp Byte = 0x58

	// CLV - Clear Overflow Flag
	OpClvImp Byte = 0xB8

	// SEC - Set Carry Flag
	OpSecImp Byte = 0x38

	// SED - Set Decimal Mode
	OpSedImp Byte = 0xF8

	// SEI - Set Interrupt Disable
	OpSeiImp Byte = 0x78

	// JMP - Jump to Memory Location
	OpJmpInd Byte = 0x6C
	OpJmpAbs Byte = 0x4C

	// JSR - Jump to Subroutine
	OpJsrAbs Byte = 0x20

	// RTS - Return from Subroutine
	OpRtsImp Byte = 0x60
)
