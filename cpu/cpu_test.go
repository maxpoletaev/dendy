package cpu

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type memory [65536]Byte

func (m *memory) Write(addr Word, data Byte) {
	m[addr] = data
}

func (m *memory) Read(addr Word) Byte {
	return m[addr]
}

func (m *memory) ReadWord(addr Word) Word {
	return Word(m[addr]) | Word(m[addr+1])<<8
}

type InstructionTest struct {
	opcode      Byte
	operand     Word
	wantCycles  int
	prepareFunc func(*testing.T, *CPU, *memory)
	assertFunc  func(*testing.T, *CPU, *memory)
}

func (i *InstructionTest) Exec(t *testing.T) {
	t.Helper()

	var (
		mem = new(memory)
		cpu = NewCPU()
	)

	cpu.Reset(mem)

	mem[0xFFFC] = 0x00
	mem[0xFFFD] = 0x50

	mem[0x5000] = i.opcode
	mem[0x5001] = Byte(i.operand)
	mem[0x5002] = Byte(i.operand >> 8)

	t.Cleanup(func() {
		if t.Failed() {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("0x%04X\t\t", 0x5000))
			b.WriteString(disassemble(mem, 0x5000))
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
		opcode:      OpNopImp,
		wantCycles:  2,
		prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {},
		assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			require.Equal(t, Word(0x5001), cpu.PC)
		},
	}).Exec(t)
}

func TestLDA(t *testing.T) {
	tests := map[string]InstructionTest{
		"Imm": {
			opcode:      OpLdaImm,
			operand:     0x42,
			wantCycles:  2,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Byte(0x42), cpu.A)
				require.Equal(t, Word(0x5002), cpu.PC)
			},
		},
		"Zp": {
			opcode:     OpLdaZp,
			operand:    0x42,
			wantCycles: 3,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x42] = 0x42
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5002), cpu.PC)
				require.Equal(t, Byte(0x42), cpu.A)
			},
		},
		"ZpX": {
			opcode:     OpLdaZpX,
			operand:    0x41,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x42] = 0x42
				cpu.X = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5002), cpu.PC)
				require.Equal(t, Byte(0x42), cpu.A)
			},
		},
		"Abs": {
			opcode:     OpLdaAbs,
			operand:    0x4242,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x4242] = 0x42
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5003), cpu.PC)
				require.Equal(t, Byte(0x42), cpu.A)
			},
		},
		"AbsX": {
			opcode:     OpLdaAbsX,
			operand:    0x4241,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x4242] = 0x42
				cpu.X = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5003), cpu.PC)
				require.Equal(t, Byte(0x42), cpu.A)
			},
		},
		"AbsY": {
			opcode:     OpLdaAbsY,
			operand:    0x4241,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x4242] = 0x42
				cpu.Y = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5003), cpu.PC)
				require.Equal(t, Byte(0x42), cpu.A)
			},
		},
		"IndX": {
			opcode:     OpLdaIndX,
			operand:    0xA0,
			wantCycles: 6,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x01FF] = 0x42
				mem[0x00A2] = 0xFF
				mem[0x00A3] = 0x01
				cpu.X = 0x02
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5002), cpu.PC)
				require.Equal(t, Byte(0x42), cpu.A)
			},
		},
		"IndY": {
			opcode:     OpLdaIndY,
			operand:    0xA0,
			wantCycles: 6,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				mem[0x0100] = 0x42
				mem[0x00A0] = 0xFF
				mem[0x00A1] = 0x00
				cpu.Y = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5002), cpu.PC)
				require.Equal(t, Byte(0x42), cpu.A)
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
			opcode:     OpStaZp,
			operand:    0x42,
			wantCycles: 3,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				cpu.A = 0x42
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5002), cpu.PC)
				require.Equal(t, Byte(0x42), mem[0x42])
			},
		},
		"ZpX": {
			opcode:     OpStaZpX,
			operand:    0x41,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				cpu.A = 0x42
				cpu.X = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5002), cpu.PC)
				require.Equal(t, Byte(0x42), mem[0x42])
			},
		},
		"Abs": {
			opcode:     OpStaAbs,
			operand:    0x4242,
			wantCycles: 4,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				cpu.A = 0x42
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5003), cpu.PC)
				require.Equal(t, Byte(0x42), mem[0x4242])
			},
		},
		"AbsX": {
			opcode:     OpStaAbsX,
			operand:    0x4241,
			wantCycles: 5,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				cpu.A = 0x42
				cpu.X = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5003), cpu.PC)
				require.Equal(t, Byte(0x42), mem[0x4242])
			},
		},
		"AbsY": {
			opcode:     OpStaAbsY,
			operand:    0x4241,
			wantCycles: 5,
			prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				cpu.A = 0x42
				cpu.Y = 0x01
			},
			assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
				require.Equal(t, Word(0x5003), cpu.PC)
				require.Equal(t, Byte(0x42), mem[0x4242])
			},
		},
	}

	for name, test := range tests {
		t.Run(name, test.Exec)
	}
}

func TestTAX(t *testing.T) {
	(&InstructionTest{
		opcode:     OpTaxImp,
		wantCycles: 2,
		prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			cpu.A = 0x42
		},
		assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			require.Equal(t, Word(0x5001), cpu.PC)
			require.Equal(t, Byte(0x42), cpu.X)
		},
	}).Exec(t)
}

func TestTXA(t *testing.T) {
	(&InstructionTest{
		opcode:     OpTxaImp,
		wantCycles: 2,
		prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			cpu.X = 0x42
		},
		assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			require.Equal(t, Word(0x5001), cpu.PC)
			require.Equal(t, Byte(0x42), cpu.A)
		},
	}).Exec(t)
}

func TestJSR(t *testing.T) {
	(&InstructionTest{
		opcode:      OpJsrAbs,
		operand:     0x4242,
		wantCycles:  6,
		prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {},
		assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			require.Equal(t, Word(0x4242), cpu.PC)
			require.Equal(t, Byte(0xFD), cpu.SP)
			require.Equal(t, Byte(0x50), mem[0x01FF])
			require.Equal(t, Byte(0x02), mem[0x01FE])
		},
	}).Exec(t)
}
