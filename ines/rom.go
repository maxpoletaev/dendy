package ines

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"

	"github.com/maxpoletaev/dendy/internal/binario"
)

var (
	ErrSavedStateMismatch = errors.New("saved state mismatch (probably different roms)")
)

type MirrorMode = uint8

const (
	MirrorHorizontal MirrorMode = 0
	MirrorVertical   MirrorMode = 1
	MirrorSingle0    MirrorMode = 2
	MirrorSingle1    MirrorMode = 3
)

type ROM struct {
	MirrorMode MirrorMode
	MapperID   uint8
	Battery    bool
	PRGBanks   int
	CHRBanks   int
	PRG        []byte
	CHR        []byte
	crc32      uint32
	chrRAM     bool
}

func loadROM(filename string) (*ROM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	h := crc32.NewIEEE()
	reader := io.TeeReader(file, h)

	// Read header.
	header := make([]uint8, 16)
	_, err = reader.Read(header)
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
	if _, err = reader.Read(prg); err != nil {
		return nil, fmt.Errorf("failed to read PRG ROM: %w", err)
	}

	// Read CHR-ROM.
	chr := make([]uint8, chrSize)
	if _, err = reader.Read(chr); err != nil {
		return nil, fmt.Errorf("failed to read chr ROM: %w", err)
	}

	// If CHR-ROM is empty, allocate 8KB of CHR-RAM.
	var chrRAM bool
	if chrSize == 0 {
		chr = make([]uint8, 8192)
		chrSize = 8192
		chrRAM = true
	}

	return &ROM{
		PRG:        prg,
		CHR:        chr,
		MapperID:   mapperID,
		Battery:    hasBattery,
		MirrorMode: mirrorMode,
		PRGBanks:   prgSize / 16384,
		CHRBanks:   chrSize / 8192,
		crc32:      h.Sum32(),
		chrRAM:     chrRAM,
	}, nil
}

func (r *ROM) SaveState(w *binario.Writer) error {
	if err := w.WriteUint32(r.crc32); err != nil {
		return err
	}

	if r.chrRAM {
		if err := w.WriteBytes(r.CHR); err != nil {
			return err
		}
	}

	return nil
}

func (r *ROM) LoadState(reader *binario.Reader) error {
	hash, err := reader.ReadUint32()
	if err != nil {
		return err
	}

	if hash != r.crc32 {
		return ErrSavedStateMismatch
	}

	if r.chrRAM {
		if err = reader.ReadBytesTo(r.CHR); err != nil {
			return err
		}
	}

	return nil
}
