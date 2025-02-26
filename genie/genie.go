package genie

import (
	"errors"
	"log"
	"strings"

	"github.com/maxpoletaev/dendy/ines"
	"github.com/maxpoletaev/dendy/internal/binario"
)

var (
	_ ines.Cartridge = (*GameGenie)(nil)
)

type override struct {
	mode byte
	cmp  byte
	data byte
}

type GameGenie struct {
	prgOverride []override
	cart        ines.Cartridge
}

func New(cart ines.Cartridge) *GameGenie {
	return &GameGenie{
		prgOverride: make([]override, 0x8000),
		cart:        cart,
	}
}

func (gg *GameGenie) ApplyCode(code string) error {
	sb := []byte(strings.ToUpper(code))

	switch len(sb) {
	case 6:
		addr, ov := decode6(sb)
		gg.prgOverride[addr-0x8000] = ov
		log.Printf("[INFO] code applied: %s (addr=0x%04X data=0x%02X)", code, addr, ov.data)
	case 8:
		addr, ov := decode8(sb)
		gg.prgOverride[addr-0x8000] = ov
		log.Printf("[INFO] code applied: %s (addr=0x%04X data=0x%02X cmp=0x%02X)", code, addr, ov.cmp, ov.data)
	default:
		return errors.New("invalid code length")
	}

	return nil
}

func (gg *GameGenie) Reset() {
	gg.cart.Reset()
}

func (gg *GameGenie) ScanlineTick() {
	gg.cart.ScanlineTick()
}

func (gg *GameGenie) PendingIRQ() bool {
	return gg.cart.PendingIRQ()
}

func (gg *GameGenie) MirrorMode() ines.MirrorMode {
	return gg.cart.MirrorMode()
}

func (gg *GameGenie) ReadPRG(addr uint16) byte {
	realVal := gg.cart.ReadPRG(addr)
	ov := gg.prgOverride[addr-0x8000]

	switch ov.mode {
	case 6:
		return ov.data
	case 8:
		if ov.cmp == realVal {
			return ov.data
		}
	default:
		return realVal
	}

	return realVal
}

func (gg *GameGenie) WritePRG(addr uint16, data byte) {
	gg.cart.WritePRG(addr, data)
}

func (gg *GameGenie) ReadCHR(addr uint16) byte {
	return gg.cart.ReadCHR(addr)
}

func (gg *GameGenie) WriteCHR(addr uint16, data byte) {
	gg.cart.WriteCHR(addr, data)
}

func (gg *GameGenie) SaveState(w *binario.Writer) error {
	return gg.cart.SaveState(w)
}

func (gg *GameGenie) LoadState(r *binario.Reader) error {
	return gg.cart.LoadState(r)
}
