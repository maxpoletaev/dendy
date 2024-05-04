package ppu

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
)

func (p *PPU) SaveState(w *binario.Writer) error {
	return errors.Join(
		w.WriteBool(p.PendingNMI),
		w.WriteBool(p.FrameComplete),
		w.WriteBool(p.ScanlineComplete),
		w.WriteUint8(p.ctrl),
		w.WriteUint8(p.mask),
		w.WriteUint8(p.status),
		w.WriteUint8(p.oamAddr),
		w.WriteByteSlice(p.oamData[:]),
		w.WriteUint16(uint16(p.vramAddr)),
		w.WriteUint16(uint16(p.tmpAddr)),
		w.WriteUint8(p.vramBuffer),
		w.WriteBool(p.addrLatch),
		w.WriteUint8(p.fineX),
		w.WriteByteSlice(p.nameTable[0][:]),
		w.WriteByteSlice(p.nameTable[1][:]),
		w.WriteByteSlice(p.paletteTable[:]),
		w.WriteUint64(uint64(p.cycle)),
		w.WriteUint64(uint64(p.scanline)),
		w.WriteBool(p.oddFrame),
	)
}

func (p *PPU) LoadState(r *binario.Reader) error {
	var (
		currAddr uint16
		tmpAddr  uint16
		cycle    uint64
		scanline uint64
	)

	err := errors.Join(
		r.ReadBoolTo(&p.PendingNMI),
		r.ReadBoolTo(&p.FrameComplete),
		r.ReadBoolTo(&p.ScanlineComplete),
		r.ReadUint8To(&p.ctrl),
		r.ReadUint8To(&p.mask),
		r.ReadUint8To(&p.status),
		r.ReadUint8To(&p.oamAddr),
		r.ReadByteSliceTo(p.oamData[:]),
		r.ReadUint16To(&currAddr),
		r.ReadUint16To(&tmpAddr),
		r.ReadUint8To(&p.vramBuffer),
		r.ReadBoolTo(&p.addrLatch),
		r.ReadUint8To(&p.fineX),
		r.ReadByteSliceTo(p.nameTable[0][:]),
		r.ReadByteSliceTo(p.nameTable[1][:]),
		r.ReadByteSliceTo(p.paletteTable[:]),
		r.ReadUint64To(&cycle),
		r.ReadUint64To(&scanline),
		r.ReadBoolTo(&p.oddFrame),
	)

	p.vramAddr = vramAddr(currAddr)
	p.tmpAddr = vramAddr(tmpAddr)
	p.scanline = int(scanline)
	p.cycle = int(cycle)

	return err
}
