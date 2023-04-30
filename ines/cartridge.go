package ines

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type (
	MirrorMode byte
)

const (
	Horizontal MirrorMode = 0
	Vertical   MirrorMode = 1
)

type Cartridge struct {
	MapperID   byte
	mapper     Mapper
	Mirror     MirrorMode
	Battery    bool
	FourScreen bool

	// Main memory banks are accessed through the mapper.
	prg []byte
	chr []byte
}

func OpenROM(filename string) (*Cartridge, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	// Read header.
	header := make([]uint8, 16)
	_, err = file.Read(header)
	if err != nil {
		return nil, err
	}

	// Check header signature.
	if header[0] != 'N' || header[1] != 'E' || header[2] != 'S' || header[3] != 0x1A {
		return nil, errors.New("invalid ROM file")
	}

	var (
		mapperID   = (header[6] >> 4) | (header[7] & (1 << 4))
		prgSize    = int(header[4]) * 16384
		chrSize    = int(header[5]) * 8192
		hasTrainer = header[6]&(1<<2) != 0
	)

	// Skip trainer if present.
	if hasTrainer {
		if _, err = file.Seek(512, io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("failed to skip trainer: %w", err)
		}
	}

	// Read PRG-ROM.
	prg := make([]uint8, prgSize)
	if _, err = file.Read(prg); err != nil {
		return nil, fmt.Errorf("failed to read PRG ROM: %w", err)
	}

	// Read CHR-ROM.
	chr := make([]uint8, chrSize)
	if _, err = file.Read(chr); err != nil {
		return nil, fmt.Errorf("failed to read chr ROM: %w", err)
	}

	// Initialize the mapper.
	mapper, err := newMapper(mapperID)
	if err != nil {
		return nil, err
	}

	return &Cartridge{
		prg:        prg,
		chr:        chr,
		mapper:     mapper,
		MapperID:   mapperID,
		Battery:    header[6]&(1<<1) != 0,
		FourScreen: header[6]&(1<<3) != 0,
		Mirror:     MirrorMode(header[6] & (1 << 0)),
	}, nil
}

func (c *Cartridge) ReadPRG(addr uint16) byte {
	return c.mapper.ReadPRG(c, addr)
}

func (c *Cartridge) WritePRG(addr uint16, value byte) {
	c.mapper.WritePRG(c, addr, value)
}

func (c *Cartridge) ReadCHR(addr uint16) byte {
	return c.mapper.ReadCHR(c, addr)
}

func (c *Cartridge) WriteCHR(addr uint16, value byte) {
	c.mapper.WriteCHR(c, addr, value)
}
