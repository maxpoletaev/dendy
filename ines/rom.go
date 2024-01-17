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
		prgBanks   = int(header[4])
		chrBanks   = int(header[5])
		hasTrainer = header[6]&(1<<2) != 0
		hasBattery = header[6]&(1<<1) != 0
		mirrorMode = header[6] & (1 << 0)
	)

	// Skip trainer if present.
	if hasTrainer {
		if _, err = file.Seek(512, io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("failed to skip trainer: %w", err)
		}
	}

	// Read PRG-ROM.
	prgData := make([]uint8, prgBanks*16384)
	if _, err = reader.Read(prgData); err != nil {
		return nil, fmt.Errorf("failed to read PRG ROM: %w", err)
	}

	// Read CHR-ROM.
	chrData := make([]uint8, chrBanks*8192)
	if _, err = reader.Read(chrData); err != nil {
		return nil, fmt.Errorf("failed to read chr ROM: %w", err)
	}

	var chrRAM bool
	if len(chrData) == 0 {
		// No CHR-ROM, so allocate 8KB of CHR-RAM.
		chrData = make([]uint8, 8192)
		chrRAM = true
	}

	return &ROM{
		PRG:        prgData,
		CHR:        chrData,
		MapperID:   mapperID,
		Battery:    hasBattery,
		MirrorMode: mirrorMode,
		PRGBanks:   prgBanks,
		CHRBanks:   chrBanks,
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
