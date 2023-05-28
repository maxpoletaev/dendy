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
	MirrorHorizontal MirrorMode = iota
	MirrorVertical
	MirrorSingleLo
	MirrorSingleHi
)

type ROM struct {
	MirrorMode MirrorMode
	MapperID   uint8
	Battery    bool
	PRGBanks   int
	CHRBanks   int
	PRG        []byte
	CHR        []byte
}

func loadROM(filename string) (*ROM, error) {
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
		hasBattery = header[6]&(1<<1) != 0
		mirrorMode = MirrorMode(header[6] & (1 << 0))
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

	// If CHR-ROM is empty, allocate 8KB of CHR-RAM.
	if chrSize == 0 {
		chr = make([]uint8, 8192)
		chrSize = 8192
	}

	return &ROM{
		PRG:        prg,
		CHR:        chr,
		MapperID:   mapperID,
		Battery:    hasBattery,
		MirrorMode: mirrorMode,
		PRGBanks:   prgSize / 16384,
		CHRBanks:   chrSize / 8192,
	}, nil
}
