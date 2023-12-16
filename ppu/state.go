package ppu

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
)

func (p *PPU) SaveState(w *binario.Writer) error {
	return errors.Join(
		w.WriteBool(p.pendingNMI),
		w.WriteBool(p.frameComplete),
		w.WriteBool(p.scanlineComplete),
		w.WriteUint8(p.ctrl),
		w.WriteUint8(p.mask),
		w.WriteUint8(p.status),
		w.WriteUint8(p.oamAddr),
		w.WriteBytes(p.oamData[:]),
		w.WriteUint16(uint16(p.vramAddr)),
		w.WriteUint16(uint16(p.tmpAddr)),
		w.WriteUint8(p.vramBuffer),
		w.WriteBool(p.addrLatch),
		w.WriteUint8(p.fineX),
		w.WriteBytes(p.nameTable[0][:]),
		w.WriteBytes(p.nameTable[1][:]),
		w.WriteBytes(p.paletteTable[:]),
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
		r.ReadBoolTo(&p.pendingNMI),
		r.ReadBoolTo(&p.frameComplete),
		r.ReadBoolTo(&p.scanlineComplete),
		r.ReadUint8To(&p.ctrl),
		r.ReadUint8To(&p.mask),
		r.ReadUint8To(&p.status),
		r.ReadUint8To(&p.oamAddr),
		r.ReadBytesTo(p.oamData[:]),
		r.ReadUint16To(&currAddr),
		r.ReadUint16To(&tmpAddr),
		r.ReadUint8To(&p.vramBuffer),
		r.ReadBoolTo(&p.addrLatch),
		r.ReadUint8To(&p.fineX),
		r.ReadBytesTo(p.nameTable[0][:]),
		r.ReadBytesTo(p.nameTable[1][:]),
		r.ReadBytesTo(p.paletteTable[:]),
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
