package cpu

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type memory [65536]Byte

func (m memory) Write(addr Word, data Byte) {
	m[addr] = data
}

func (m memory) Read(addr Word) Byte {
	return m[addr]
}

func (m memory) ReadWord(addr Word) Word {
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
			cpu.PrintState()
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

func TestNop(t *testing.T) {
	(&InstructionTest{
		opcode:      OpNopImp,
		wantCycles:  2,
		prepareFunc: func(t *testing.T, cpu *CPU, mem *memory) {},
		assertFunc: func(t *testing.T, cpu *CPU, mem *memory) {
			require.Equal(t, Word(0x5001), cpu.PC)
		},
	}).Exec(t)
}

func TestLda(t *testing.T) {
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
