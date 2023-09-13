package ines

import (
	"encoding/gob"
	"fmt"
	"log"
)

type TickInfo struct {
	IRQ bool
}

type Cartridge interface {
	// Reset resets the cartridge to its initial state.
	Reset()
	// ScanlineTick performs a scanline tick used by some mappers.
	ScanlineTick() TickInfo
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
	// Save saves the cartridge state to the given encoder.
	Save(enc *gob.Encoder) error
	// Load restores the cartridge state from the given decoder.
	Load(dec *gob.Decoder) error
}

func Load(path string) (Cartridge, error) {
	rom, err := loadROM(path)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] loaded rom: mapper:%d", rom.MapperID)

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
	default:
		return nil, fmt.Errorf("unsupported mapper: %d", rom.MapperID)
	}
}
