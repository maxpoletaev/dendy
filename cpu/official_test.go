package cpu

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type memory [65536]uint8

func (m *memory) Write(addr uint16, data uint8) {
	m[addr] = data
}

func (m *memory) Read(addr uint16) uint8 {
	return m[addr]
}

type InstructionTest struct {
	opcode      uint8
	operand     uint16
	wantCycles  int
	prepareFunc func(*testing.T, *CPU, *memory)
	assertFunc  func(*testing.T, *CPU, *memory)
}

func (i *InstructionTest) Exec(t *testing.T) {
	t.Helper()

	var (
		mem = new(memory)
		cpu = New()
	)

	cpu.Reset(mem)

	mem[0xFFFC] = 0x00
	mem[0xFFFD] = 0x50

	mem[0x5000] = i.opcode
	mem[0x5001] = uint8(i.operand)
	mem[0x5002] = uint8(i.operand >> 8)

	t.Cleanup(func() {
		if t.Failed() {
			var b strings.Builder
			cpu.PC = 0x5000

			b.WriteString(fmt.Sprintf("0x%04X\t\t", 0x5000))
			b.WriteString(disassemble(cpu, mem))
			b.WriteString("\t\t")
			b.WriteString(fmt.Sprintf(" A:%02X", cpu.A))
			b.WriteString(fmt.Sprintf(" X:%02X", cpu.X))
			b.WriteString(fmt.Sprintf(" Y:%02X", cpu.Y))
			b.WriteString(fmt.Sprintf(" P:%08b", cpu.P))
			b.WriteString(fmt.Sprintf(" SP:%02X", cpu.SP))
			fmt.Println(b.String())
		}
	})

	cpu.Reset(mem)
	i.prepareFunc(t, cpu, mem)

	var (
		cycles   int
		lastTick bool
	)

	for {
		lastTick = cpu.Tick(mem)
		cycles++

		if lastTick {
			break
		}
	}

	require.Equal(t, i.wantCycles, cycles)
	i.assertFunc(t, cpu, mem)
}

func TestNOP(t *testing.T) {
	(&InstructionTest{
		opcode:      NOP_Imp,
		wantCycles:  2,
		prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {},
		assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			require.Equal(t, uint16(0x5001), cpu.PC)
		},
	}).Exec(t)
}

func TestLDA(t *testing.T) {
	tests := map[string]InstructionTest{
		"Imm": {
			opcode:      LDA_Imm,
			operand:     0x42,
			wantCycles:  2,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint8(0x42), cpu.A)
				require.Equal(t, uint16(0x5002), cpu.PC)
			},
		},
		"Zp": {
			opcode:     LDA_Zp,
			operand:    0x42,
			wantCycles: 3,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x42] = 0x42
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5002), cpu.PC)
				require.Equal(t, uint8(0x42), cpu.A)
			},
		},
		"ZpX": {
			opcode:     LDA_ZpX,
			operand:    0x41,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x42] = 0x42
				cpu.X = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5002), cpu.PC)
				require.Equal(t, uint8(0x42), cpu.A)
			},
		},
		"Abs": {
			opcode:     LDA_Abs,
			operand:    0x4242,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x4242] = 0x42
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5003), cpu.PC)
				require.Equal(t, uint8(0x42), cpu.A)
			},
		},
		"AbsX": {
			opcode:     LDA_AbsX,
			operand:    0x4241,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x4242] = 0x42
				cpu.X = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5003), cpu.PC)
				require.Equal(t, uint8(0x42), cpu.A)
			},
		},
		"AbsY": {
			opcode:     LDA_AbsY,
			operand:    0x4241,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x4242] = 0x42
				cpu.Y = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5003), cpu.PC)
				require.Equal(t, uint8(0x42), cpu.A)
			},
		},
		"IndX": {
			opcode:     LDA_IndX,
			operand:    0xA0,
			wantCycles: 6,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x01FF] = 0x42
				mem[0x00A2] = 0xFF
				mem[0x00A3] = 0x01
				cpu.X = 0x02
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5002), cpu.PC)
				require.Equal(t, uint8(0x42), cpu.A)
			},
		},
		"IndX_Overflow": {
			opcode:     LDA_IndX,
			operand:    0xA0,
			wantCycles: 6,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0xA0] = 0xFF
				mem[0xA1] = 0xAA
				mem[0xAAFF] = 0x42

				cpu.X = 0x01
				cpu.A = 0x00
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5002), cpu.PC)
				require.Equal(t, uint8(0x42), cpu.A)
			},
		},
		"IndY": {
			opcode:     LDA_IndY,
			operand:    0xA0,
			wantCycles: 6,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x0100] = 0x42
				mem[0x00A0] = 0xFF
				mem[0x00A1] = 0x00
				cpu.Y = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5002), cpu.PC)
				require.Equal(t, uint8(0x42), cpu.A)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, test.Exec)
	}
}

func TestSTA(t *testing.T) {
	tests := map[string]InstructionTest{
		"Zp": {
			opcode:     STA_Zp,
			operand:    0x42,
			wantCycles: 3,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				cpu.A = 0x42
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5002), cpu.PC)
				require.Equal(t, uint8(0x42), mem[0x42])
			},
		},
		"ZpX": {
			opcode:     STA_ZpX,
			operand:    0x41,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				cpu.A = 0x42
				cpu.X = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5002), cpu.PC)
				require.Equal(t, uint8(0x42), mem[0x42])
			},
		},
		"Abs": {
			opcode:     STA_Abs,
			operand:    0x4242,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				cpu.A = 0x42
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5003), cpu.PC)
				require.Equal(t, uint8(0x42), mem[0x4242])
			},
		},
		"AbsX": {
			opcode:     STA_AbsX,
			operand:    0x4241,
			wantCycles: 5,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				cpu.A = 0x42
				cpu.X = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5003), cpu.PC)
				require.Equal(t, uint8(0x42), mem[0x4242])
			},
		},
		"AbsY": {
			opcode:     STA_Ind,
			operand:    0x4241,
			wantCycles: 5,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				cpu.A = 0x42
				cpu.Y = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, uint16(0x5003), cpu.PC)
				require.Equal(t, uint8(0x42), mem[0x4242])
			},
		},
	}

	for name, test := range tests {
		t.Run(name, test.Exec)
	}
}

func TestTAX(t *testing.T) {
	(&InstructionTest{
		opcode:     TAX_Imp,
		wantCycles: 2,
		prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			cpu.A = 0x42
		},
		assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			require.Equal(t, uint16(0x5001), cpu.PC)
			require.Equal(t, uint8(0x42), cpu.X)
		},
	}).Exec(t)
}

func TestTXA(t *testing.T) {
	(&InstructionTest{
		opcode:     TXA_Imp,
		wantCycles: 2,
		prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			cpu.X = 0x42
		},
		assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			require.Equal(t, uint16(0x5001), cpu.PC)
			require.Equal(t, uint8(0x42), cpu.A)
		},
	}).Exec(t)
}

func TestJSR(t *testing.T) {
	(&InstructionTest{
		opcode:      JSR_Abs,
		operand:     0x4242,
		wantCycles:  6,
		prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {},
		assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			require.Equal(t, uint16(0x4242), cpu.PC)
			require.Equal(t, uint8(0xFD), cpu.SP)
			require.Equal(t, uint8(0x50), mem[0x01FF])
			require.Equal(t, uint8(0x02), mem[0x01FE])
		},
	}).Exec(t)
}
