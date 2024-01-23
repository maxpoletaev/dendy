package ines

import (
	"fmt"

	"github.com/maxpoletaev/dendy/internal/binario"
)

type Cartridge interface {
	// Reset resets the cartridge to its initial state.
	Reset()
	// ScanlineTick performs a scanline tick used by some mappers.
	ScanlineTick()
	// PendingIRQ returns true if the cartridge has an IRQ pending.
	PendingIRQ() bool
	// MirrorMode returns the cartridge's mirroring mode.
	MirrorMode() MirrorMode
	// ReadPRG handles CPU reads from PRG ROM (0x8000-0xFFFF).
	ReadPRG(addr uint16) byte
	// WritePRG handles CPU writes to PRG ROM (0x8000-0xFFFF).
	WritePRG(addr uint16, data byte)
	// ReadCHR handles PPU reads from CHR ROM (0x0000-0x1FFF).
	ReadCHR(addr uint16) byte
	// WriteCHR handles PPU writes to CHR ROM (0x0000-0x1FFF).
	WriteCHR(addr uint16, data byte)
	// SaveState saves the cartridge state to the given writer.
	SaveState(w *binario.Writer) error
	// LoadState restores the cartridge state from the given reader.
	LoadState(r *binario.Reader) error
}

func NewCartridge(rom *ROM) (Cartridge, error) {
	switch rom.MapperID {
	case 0:
		return NewMapper0(rom), nil
	case 1:
		return NewMapper1(rom), nil
	case 2:
		return NewMapper2(rom), nil
	case 3:
		return NewMapper3(rom), nil
	case 4:
		return NewMapper4(rom), nil
	case 7:
		return NewMapper7(rom), nil
	default:
		return nil, fmt.Errorf("unsupported mapper: %d", rom.MapperID)
	}
}
