package console

import (
	"errors"

	"github.com/maxpoletaev/dendy/internal/binario"
)

func (b *Bus) SaveState(w *binario.Writer) error {
	err := errors.Join(
		w.WriteBytes(b.RAM[:]),
		w.WriteUint64(b.cycles),
		b.CPU.SaveState(w),
		b.PPU.SaveState(w),
		b.APU.SaveState(w),
		b.Cart.SaveState(w),
		b.Joy1.SaveState(w),
		b.Joy2.SaveState(w),
	)

	return err
}

func (b *Bus) LoadState(r *binario.Reader) error {
	err := errors.Join(
		r.ReadBytesTo(b.RAM[:]),
		r.ReadUint64To(&b.cycles),
		b.CPU.LoadState(r),
		b.PPU.LoadState(r),
		b.APU.LoadState(r),
		b.Cart.LoadState(r),
		b.Joy1.LoadState(r),
		b.Joy2.LoadState(r),
	)

	return err
}
