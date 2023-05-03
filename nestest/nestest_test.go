//go:build testrom

package nestest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/maxpoletaev/dendy/cpu"
	"github.com/maxpoletaev/dendy/ines"
)

type Memory struct {
	rom ines.Cartridge
	ram [2048]byte
}

func (r *Memory) Read(addr uint16) byte {
	if addr <= 0x1FFF {
		return r.ram[addr%2048]
	} else if addr >= 0x4200 {
		return r.rom.ReadPRG(addr)
	}

	return 0
}

func (r *Memory) Write(addr uint16, value byte) {
	if addr <= 0x1FFF {
		r.ram[addr%2048] = value
	} else if addr >= 0x4200 {
		r.rom.WritePRG(addr, value)
	}
}

var failCodes = map[byte]string{
	// branch tests
	0x01: "BCS failed to branch",
	0x02: "BCS branched when it shouldn't have",
	0x03: "BCC branched when it shouldn't have",
	0x04: "BCC failed to branch",
	0x05: "BEQ failed to branch",
	0x06: "BEQ branched when it shouldn't have",
	0x07: "BNE failed to branch",
	0x08: "BNE branched when it shouldn't have",
	0x09: "BVS failed to branch",
	0x0A: "BVC branched when it shouldn't have",
	0x0B: "BVC failed to branch",
	0x0C: "BVS branched when it shouldn't have",
	0x0D: "BPL failed to branch",
	0x0E: "BPL branched when it shouldn't have",
	0x0F: "BMI failed to branch",
	0x10: "BMI branched when it shouldn't have",

	// flag tests
	0x11: "PHP/flags failure (bits set)",
	0x12: "PHP/flags failure (bits clear)",
	0x13: "PHP/flags failure (misc bit states)",
	0x14: "PLP/flags failure (misc bit states)",
	0x15: "PLP/flags failure (misc bit states)",
	0x16: "PHA/PLA failure (PLA didn't affect Z and N properly)",
	0x17: "PHA/PLA failure (PLA didn't affect Z and N properly)",

	// immediate instruction tests
	0x18: "ORA # failure",
	0x19: "ORA # failure",
	0x1A: "AND # failure",
	0x1B: "AND # failure",
	0x1C: "EOR # failure",
	0x1D: "EOR # failure",
	0x1E: "ADC # failure (overflow/carry problems)",
	0x1F: "ADC # failure (decimal mode was turned on)",
	0x20: "ADC # failure",
	0x21: "ADC # failure",
	0x22: "ADC # failure",
	0x23: "LDA # failure (didn't set N and Z correctly)",
	0x24: "LDA # failure (didn't set N and Z correctly)",
	0x25: "CMP # failure (messed up flags)",
	0x26: "CMP # failure (messed up flags)",
	0x27: "CMP # failure (messed up flags)",
	0x28: "CMP # failure (messed up flags)",
	0x29: "CMP # failure (messed up flags)",
	0x2A: "CMP # failure (messed up flags)",
	0x2B: "CPY # failure (messed up flags)",
	0x2C: "CPY # failure (messed up flags)",
	0x2D: "CPY # failure (messed up flags)",
	0x2E: "CPY # failure (messed up flags)",
	0x2F: "CPY # failure (messed up flags)",
	0x30: "CPY # failure (messed up flags)",
	0x31: "CPY # failure (messed up flags)",
	0x32: "CPX # failure (messed up flags)",
	0x33: "CPX # failure (messed up flags)",
	0x34: "CPX # failure (messed up flags)",
	0x35: "CPX # failure (messed up flags)",
	0x36: "CPX # failure (messed up flags)",
	0x37: "CPX # failure (messed up flags)",
	0x38: "CPX # failure (messed up flags)",
	0x39: "LDX # failure (didn't set N and Z correctly)",
	0x3A: "LDX # failure (didn't set N and Z correctly)",
	0x3B: "LDY # failure (didn't set N and Z correctly)",
	0x3C: "LDY # failure (didn't set N and Z correctly)",
	0x3D: "compare(s) stored the result in a register (whoops!)",
	0x71: "SBC # failure",
	0x72: "SBC # failure",
	0x73: "SBC # failure",
	0x74: "SBC # failure",
	0x75: "SBC # failure",

	// implied instruction tests
	0x3E: "INX/DEX/INY/DEY did something bad",
	0x3F: "INY/DEY messed up overflow or carry",
	0x40: "INX/DEX messed up overflow or carry",
	0x41: "TAY did something bad (changed wrong regs, messed up flags)",
	0x42: "TAX did something bad (changed wrong regs, messed up flags)",
	0x43: "TYA did something bad (changed wrong regs, messed up flags)",
	0x44: "TXA did something bad (changed wrong regs, messed up flags)",
	0x45: "TXS didn't set flags right, or TSX touched flags and it shouldn't have",

	// stack tests
	0x46: "wrong data popped, or data not in right location on stack",
	0x47: "JSR didn't work as expected",
	0x48: "RTS/JSR shouldn't have affected flags",
	0x49: "RTI/RTS didn't work right when return addys/data were manually pushed",

	// accumulator tests
	0x4A: "LSR A failed",
	0x4B: "ASL A failed",
	0x4C: "ROR A failed",
	0x4D: "ROL A failed",

	// (indirect,x) tests
	0x58: "LDA didn't load the data it expected to load",
	0x59: "STA didn't store the data where it was supposed to",
	0x5A: "ORA failure",
	0x5B: "ORA failure",
	0x5C: "AND failure",
	0x5D: "AND failure",
	0x5E: "EOR failure",
	0x5F: "EOR failure",
	0x60: "ADC failure",
	0x61: "ADC failure",
	0x62: "ADC failure",
	0x63: "ADC failure",
	0x64: "ADC failure",
	0x65: "CMP failure",
	0x66: "CMP failure",
	0x67: "CMP failure",
	0x68: "CMP failure",
	0x69: "CMP failure",
	0x6A: "CMP failure",
	0x6B: "CMP failure",
	0x6C: "SBC failure",
	0x6D: "SBC failure",
	0x6E: "SBC failure",
	0x6F: "SBC failure",
	0x70: "SBC failure",

	// zero page tests
	0x76: "LDA didn't set the flags properly",
	0x77: "STA affected flags it shouldn't",
	0x78: "LDY didn't set the flags properly",
	0x79: "STY affected flags it shouldn't",
	0x7A: "LDX didn't set the flags properly",
	0x7B: "STX affected flags it shouldn't",
	0x7C: "BIT failure",
	0x7D: "BIT failure",
	0x7E: "ORA failure",
	0x7F: "ORA failure",
	0x80: "AND failure",
	0x81: "AND failure",
	0x82: "EOR failure",
	0x83: "EOR failure",
	0x84: "ADC failure",
	0x85: "ADC failure",
	0x86: "ADC failure",
	0x87: "ADC failure",
	0x88: "ADC failure",
	0x89: "CMP failure",
	0x8A: "CMP failure",
	0x8B: "CMP failure",
	0x8C: "CMP failure",
	0x8D: "CMP failure",
	0x8E: "CMP failure",
	0x8F: "CMP failure",
	0x90: "SBC failure",
	0x91: "SBC failure",
	0x92: "SBC failure",
	0x93: "SBC failure",
	0x94: "SBC failure",
	0x95: "CPX failure",
	0x96: "CPX failure",
	0x97: "CPX failure",
	0x98: "CPX failure",
	0x99: "CPX failure",
	0x9A: "CPX failure",
	0x9B: "CPX failure",
	0x9C: "CPY failure",
	0x9D: "CPY failure",
	0x9E: "CPY failure",
	0x9F: "CPY failure",
	0xA0: "CPY failure",
	0xA1: "CPY failure",
	0xA2: "CPY failure",
	0xA3: "LSR failure",
	0xA4: "LSR failure",
	0xA5: "ASL failure",
	0xA6: "ASL failure",
	0xA7: "ROL failure",
	0xA8: "ROL failure",
	0xA9: "ROR failure",
	0xAA: "ROR failure",
	0xAB: "INC failure",
	0xAC: "INC failure",
	0xAD: "DEC failure",
	0xAE: "DEC failure",
	0xAF: "DEC failure",

	// absolute tests
	0xB0: "LDA didn't set the flags properly",
	0xB1: "STA affected flags it shouldn't",
	0xB2: "LDY didn't set the flags properly",
	0xB3: "STY affected flags it shouldn't",
	0xB4: "LDX didn't set the flags properly",
	0xB5: "STX affected flags it shouldn't",
	0xB6: "BIT failure",
	0xB7: "BIT failure",
	0xB8: "ORA failure",
	0xB9: "ORA failure",
	0xBA: "AND failure",
	0xBB: "AND failure",
	0xBC: "EOR failure",
	0xBD: "EOR failure",
	0xBE: "ADC failure",
	0xBF: "ADC failure",
	0xC0: "ADC failure",
	0xC1: "ADC failure",
	0xC2: "ADC failure",
	0xC3: "CMP failure",
	0xC4: "CMP failure",
	0xC5: "CMP failure",
	0xC6: "CMP failure",
	0xC7: "CMP failure",
	0xC8: "CMP failure",
	0xC9: "CMP failure",
	0xCA: "SBC failure",
	0xCB: "SBC failure",
	0xCC: "SBC failure",
	0xCD: "SBC failure",
	0xCE: "SBC failure",
	0xCF: "CPX failure",
	0xD0: "CPX failure",
	0xD1: "CPX failure",
	0xD2: "CPX failure",
	0xD3: "CPX failure",
	0xD4: "CPX failure",
	0xD5: "CPX failure",
	0xD6: "CPY failure",
	0xD7: "CPY failure",
	0xD8: "CPY failure",
	0xD9: "CPY failure",
	0xDA: "CPY failure",
	0xDB: "CPY failure",
	0xDC: "CPY failure",
	0xDD: "LSR failure",
	0xDE: "LSR failure",
	0xDF: "ASL failure",
	0xE0: "ASL failure",
	0xE1: "ROR failure",
	0xE2: "ROR failure",
	0xE3: "ROL failure",
	0xE4: "ROL failure",
	0xE5: "INC failure",
	0xE6: "INC failure",
	0xE7: "DEC failure",
	0xE8: "DEC failure",
	0xE9: "DEC failure",

	// (indirect),y tests
	0xEA: "LDA failed to load correct data",
	0xEB: "Read location should have wrapped around to 0x0000",
	0xEC: "Should have wrapped zeropage address",
	0xED: "ORA failure",
	0xEE: "ORA failure",
	0xEF: "AND failure",
	0xF0: "AND failure",
	0xF1: "EOR failure",
	0xF2: "EOR failure",
	0xF3: "ADC failure",
	0xF4: "ADC failure",
	0xF5: "ADC failure",
	0xF6: "ADC failure",
	0xF7: "ADC failure",
	0xF8: "CMP failure",
	0xF9: "CMP failure",
	0xFA: "CMP failure",
	0xFB: "CMP failure",
	0xFC: "CMP failure",
	0xFD: "CMP failure",
	0xFE: "CMP failure",
}

func TestNestestROM(t *testing.T) {
	cart, err := ines.Load("nestest.nes")
	require.NoError(t, err, "failed to open nestest rom")

	c := cpu.New()
	mem := &Memory{rom: cart}

	c.Reset(mem)
	c.PC = 0xC000
	c.Cycles = 6 // Set the cycles to 6 to match the nestest's good.log.
	c.EnableDisasm = true
	c.AllowIllegal = true

	for {
		if c.Tick(mem) {
			// Nestest ends at 0xC66E.
			if c.PC == 0xC66E {
				break
			}
		}
	}

	// The result for the official instruction is stored in 0x0002. If it's not 0x00,
	// then the test failed, and we should decode the failure code.
	if code := mem.Read(0x0002); code != 0x00 {
		reason, ok := failCodes[code]
		if !ok {
			t.Fatalf("unknown failure code: 0x%02X", code)
		}

		t.Fatalf("%s (0x%02X)", reason, code)
	}
}
