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

func TestNestestROM(t *testing.T) {
	cart, err := ines.Load("nestest.nes")
	require.NoError(t, err, "failed to open nestest rom")

	c := cpu.New()
	mem := &Memory{rom: cart}

	c.Reset(mem)
	c.PC = 0xC000
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

	// The result for the official instruction is stored in 0x0002.
	if code := mem.Read(0x0002); code != 0x00 {
		t.Fatalf("official instruction failed (0x%02X)", code)
	}

	// The result for the unofficial instruction is stored in 0x0003.
	if code := mem.Read(0x0003); code != 0x00 {
		t.Fatalf("unofficial instruction failed (0x%02X)", code)
	}
}
