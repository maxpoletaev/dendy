package nes

import (
	"encoding/gob"
	"errors"
)

func (b *Bus) Save(enc *gob.Encoder) error {
	err := errors.Join(
		enc.Encode(b.RAM),
		enc.Encode(b.cycles),
		b.CPU.Save(enc),
		b.PPU.Save(enc),
		b.Cart.Save(enc),
		b.Joy1.Save(enc),
		b.Joy2.Save(enc),
	)

	return err
}

func (b *Bus) Load(dec *gob.Decoder) error {
	err := errors.Join(
		dec.Decode(&b.RAM),
		dec.Decode(&b.cycles),
		b.CPU.Load(dec),
		b.PPU.Load(dec),
		b.Cart.Load(dec),
		b.Joy1.Load(dec),
		b.Joy2.Load(dec),
	)

	return err
}
