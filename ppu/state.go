package ppu

import (
	"encoding/gob"
	"errors"
	"fmt"
)

func (p *PPU) Save(enc *gob.Encoder) error {
	err := errors.Join(
		enc.Encode(p.RequestNMI),
		enc.Encode(p.FrameComplete),
		enc.Encode(p.ScanlineComplete),
		enc.Encode(p.ctrl),
		enc.Encode(p.mask),
		enc.Encode(p.status),
		enc.Encode(p.oamAddr),
		enc.Encode(p.oamData),
		enc.Encode(p.vramAddr),
		enc.Encode(p.tmpAddr),
		enc.Encode(p.vramBuffer),
		enc.Encode(p.addrLatch),
		enc.Encode(p.fineX),
		enc.Encode(p.nameTable),
		enc.Encode(p.paletteTable),
		enc.Encode(p.cycle),
		enc.Encode(p.scanline),
		enc.Encode(p.oddFrame),
	)

	if err != nil {
		return fmt.Errorf("failed to encode PPU state: %w", err)
	}

	return nil
}

func (p *PPU) Load(dec *gob.Decoder) error {
	err := errors.Join(
		dec.Decode(&p.RequestNMI),
		dec.Decode(&p.FrameComplete),
		dec.Decode(&p.ScanlineComplete),
		dec.Decode(&p.ctrl),
		dec.Decode(&p.mask),
		dec.Decode(&p.status),
		dec.Decode(&p.oamAddr),
		dec.Decode(&p.oamData),
		dec.Decode(&p.vramAddr),
		dec.Decode(&p.tmpAddr),
		dec.Decode(&p.vramBuffer),
		dec.Decode(&p.addrLatch),
		dec.Decode(&p.fineX),
		dec.Decode(&p.nameTable),
		dec.Decode(&p.paletteTable),
		dec.Decode(&p.cycle),
		dec.Decode(&p.scanline),
		dec.Decode(&p.oddFrame),
	)

	if err != nil {
		return fmt.Errorf("failed to decode PPU state: %w", err)
	}

	return nil
}
