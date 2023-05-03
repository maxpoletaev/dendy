package ines

import "fmt"

type Cartridge interface {
	Reset()
	MirrorMode() MirrorMode
	ReadPRG(addr uint16) byte
	WritePRG(addr uint16, data byte)
	ReadCHR(addr uint16) byte
	WriteCHR(addr uint16, data byte)
}

func Load(path string) (Cartridge, error) {
	rom, err := loadROM(path)
	if err != nil {
		return nil, err
	}

	switch rom.MapperID {
	case 0:
		return NewMapper0(rom), nil
	case 1:
		return NewMapper1(rom), nil
	case 2:
		return NewMapper2(rom), nil
	default:
		return nil, fmt.Errorf("unsupported mapper: %d", rom.MapperID)
	}
}
