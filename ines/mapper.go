package ines

import "fmt"

type Mapper interface {
	ReadPRG(cart *Cartridge, addr uint16) byte
	WritePRG(cart *Cartridge, addr uint16, data byte)
	ReadCHR(cart *Cartridge, addr uint16) byte
	WriteCHR(cart *Cartridge, addr uint16, data byte)
}

func newMapper(id uint8) (Mapper, error) {
	switch id {
	case 0:
		return NewMapper0(), nil
	default:
		return nil, fmt.Errorf("unsupported mapper: %d", id)
	}
}
