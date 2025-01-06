//go:build testrom

package nestest

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/maxpoletaev/dendy/cpu"
	"github.com/maxpoletaev/dendy/disasm"
	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/internal/loglevel"
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

func disableLogger(t *testing.T) {
	t.Helper()

	log.SetOutput(loglevel.New(os.Stderr, loglevel.LevelNone))

	t.Cleanup(func() { log.SetOutput(os.Stderr) })
}

func TestNestestROM(t *testing.T) {
	disableLogger(t)

	rom, err := ines.NewFromFile("nestest.nes")
	if err != nil {
		log.Printf("[ERROR] failed to open rom file: %s", err)
		os.Exit(1)
	}

	cart, err := ines.NewCartridge(rom)
	if err != nil {
		t.Fatal(fmt.Errorf("failed to load nestest rom: %w", err))
	}

	c := cpu.New()
	mem := &Memory{rom: cart}
	writer := bufio.NewWriter(os.Stdout)

	c.Reset(mem)
	c.PC = 0xC000

	for {
		if c.Tick(mem) {
			line := disasm.DebugStep(mem, c) + "\n"
			if _, err := writer.WriteString(line); err != nil {
				t.Fatal(fmt.Errorf("failed to write disasm line: %w", err))
			}

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
