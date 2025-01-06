package ines

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
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

var mapperNames = map[uint8]string{
	0: "NROM",
	1: "SxROM",
	2: "UxROM",
	3: "CNROM",
	4: "TxROM",
	7: "AxROM",
}

type ROM struct {
	MirrorMode MirrorMode
	MapperID   uint8
	Battery    bool
	PRGBanks   int
	CHRBanks   int
	PRG        []byte
	CHR        []byte
	CRC32      uint32
	chrRAM     bool
}

func NewFromBuffer(buf []byte) (*ROM, error) {
	return newROM(bytes.NewReader(buf))
}

func NewFromFile(filename string) (*ROM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	return newROM(file)
}

func newROM(file io.ReadSeeker) (*ROM, error) {
	// Read header.
	header := make([]uint8, 16)
	_, err := file.Read(header)
	if err != nil {
		return nil, err
	}

	// Check header signature.
	if header[0] != 'N' || header[1] != 'E' || header[2] != 'S' || header[3] != 0x1A {
		return nil, errors.New("invalid ROM file")
	}

	var (
		mapperID   = ((header[6] >> 4) & 0x0F) | (header[7] & 0xF0)
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

	// CRC32 of CHR+PRG
	hasher := crc32.NewIEEE()
	romReader := io.TeeReader(file, hasher)

	// Read PRG-ROM.
	prgData := make([]uint8, prgBanks*16384)
	if _, err = romReader.Read(prgData); err != nil {
		return nil, fmt.Errorf("failed to read PRG ROM: %w", err)
	}

	// Read CHR-ROM.
	chrData := make([]uint8, chrBanks*8192)
	if _, err = romReader.Read(chrData); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read chr ROM: %w", err)
	}

	var chrRAM bool
	if len(chrData) == 0 {
		// No CHR-ROM, so allocate 8KB of CHR-RAM.
		chrData = make([]uint8, 8192)
		chrRAM = true
	}

	log.Printf("[INFO] ROM info:")
	log.Printf("[INFO]   > mapper ID:  %d (%s)", mapperID, mapperNames[mapperID])
	log.Printf("[INFO]   > PRG banks:  %d (%d KB)", prgBanks, prgBanks*16)
	log.Printf("[INFO]   > CHR banks:  %d (%d KB)", chrBanks, chrBanks*8)
	log.Printf("[INFO]   > CRC32:      %08X", hasher.Sum32())

	return &ROM{
		PRG:        prgData,
		CHR:        chrData,
		MapperID:   mapperID,
		Battery:    hasBattery,
		MirrorMode: mirrorMode,
		PRGBanks:   prgBanks,
		CHRBanks:   chrBanks,
		chrRAM:     chrRAM,
		CRC32:      hasher.Sum32(),
	}, nil
}

func (r *ROM) SaveState(w *binario.Writer) error {
	if err := w.WriteUint32(r.CRC32); err != nil {
		return err
	}

	if r.chrRAM {
		if err := w.WriteByteSlice(r.CHR); err != nil {
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

	if hash != r.CRC32 {
		return ErrSavedStateMismatch
	}

	if r.chrRAM {
		if err = reader.ReadByteSliceTo(r.CHR); err != nil {
			return err
		}
	}

	return nil
}
