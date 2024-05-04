package ines

import (
	"fmt"

	"github.com/maxpoletaev/dendy/internal/binario"
)

type MapperID uint8

const (
	MapperID0 MapperID = 0
	MapperID1 MapperID = 1
	MapperID2 MapperID = 2
	MapperID3 MapperID = 3
	MapperID4 MapperID = 4
	MapperID7 MapperID = 7
)

// StaticCartridge is a devirtualized cartridge type that uses static dispatch to
// call the correct mapper implementation. This is an experimental approach to
// improve performance by avoiding dynamic dispatch in the hot path of the
// emulator.
type StaticCartridge struct {
	mapperID MapperID
	mapper   Cartridge
}

func NewStaticCartridge(rom *ROM) (*StaticCartridge, error) {
	c := &StaticCartridge{
		mapperID: MapperID(rom.MapperID),
	}

	switch c.mapperID {
	case MapperID0:
		c.mapper = NewMapper0(rom)
	case MapperID1:
		c.mapper = NewMapper1(rom)
	case MapperID2:
		c.mapper = NewMapper2(rom)
	case MapperID3:
		c.mapper = NewMapper3(rom)
	case MapperID4:
		c.mapper = NewMapper4(rom)
	case MapperID7:
		c.mapper = NewMapper7(rom)
	default:
		return nil, fmt.Errorf("unsupported mapper: %d", rom.MapperID)
	}

	return c, nil
}

func (c *StaticCartridge) Reset() {
	switch c.mapperID {
	case MapperID0:
		c.mapper.(*Mapper0).Reset()
	case MapperID1:
		c.mapper.(*Mapper1).Reset()
	case MapperID2:
		c.mapper.(*Mapper2).Reset()
	case MapperID3:
		c.mapper.(*Mapper3).Reset()
	case MapperID4:
		c.mapper.(*Mapper4).Reset()
	case MapperID7:
		c.mapper.(*Mapper7).Reset()
	default:
		panic("unreachable")
	}
}

func (c *StaticCartridge) ScanlineTick() {
	switch c.mapperID {
	case MapperID0:
		c.mapper.(*Mapper0).ScanlineTick()
	case MapperID1:
		c.mapper.(*Mapper1).ScanlineTick()
	case MapperID2:
		c.mapper.(*Mapper2).ScanlineTick()
	case MapperID3:
		c.mapper.(*Mapper3).ScanlineTick()
	case MapperID4:
		c.mapper.(*Mapper4).ScanlineTick()
	case MapperID7:
		c.mapper.(*Mapper7).ScanlineTick()
	default:
		panic("unreachable")
	}
}

func (c *StaticCartridge) PendingIRQ() bool {
	switch c.mapperID {
	case MapperID0:
		return c.mapper.(*Mapper0).PendingIRQ()
	case MapperID1:
		return c.mapper.(*Mapper1).PendingIRQ()
	case MapperID2:
		return c.mapper.(*Mapper2).PendingIRQ()
	case MapperID3:
		return c.mapper.(*Mapper3).PendingIRQ()
	case MapperID4:
		return c.mapper.(*Mapper4).PendingIRQ()
	case MapperID7:
		return c.mapper.(*Mapper7).PendingIRQ()
	default:
		panic("unreachable")
	}
}

func (c *StaticCartridge) MirrorMode() MirrorMode {
	switch c.mapperID {
	case MapperID0:
		return c.mapper.(*Mapper0).MirrorMode()
	case MapperID1:
		return c.mapper.(*Mapper1).MirrorMode()
	case MapperID2:
		return c.mapper.(*Mapper2).MirrorMode()
	case MapperID3:
		return c.mapper.(*Mapper3).MirrorMode()
	case MapperID4:
		return c.mapper.(*Mapper4).MirrorMode()
	case MapperID7:
		return c.mapper.(*Mapper7).MirrorMode()
	default:
		panic("unreachable")
	}
}

func (c *StaticCartridge) ReadPRG(addr uint16) byte {
	switch c.mapperID {
	case MapperID0:
		return c.mapper.(*Mapper0).ReadPRG(addr)
	case MapperID1:
		return c.mapper.(*Mapper1).ReadPRG(addr)
	case MapperID2:
		return c.mapper.(*Mapper2).ReadPRG(addr)
	case MapperID3:
		return c.mapper.(*Mapper3).ReadPRG(addr)
	case MapperID4:
		return c.mapper.(*Mapper4).ReadPRG(addr)
	case MapperID7:
		return c.mapper.(*Mapper7).ReadPRG(addr)
	default:
		panic("unreachable")
	}
}

func (c *StaticCartridge) WritePRG(addr uint16, data byte) {
	switch c.mapperID {
	case MapperID0:
		c.mapper.(*Mapper0).WritePRG(addr, data)
	case MapperID1:
		c.mapper.(*Mapper1).WritePRG(addr, data)
	case MapperID2:
		c.mapper.(*Mapper2).WritePRG(addr, data)
	case MapperID3:
		c.mapper.(*Mapper3).WritePRG(addr, data)
	case MapperID4:
		c.mapper.(*Mapper4).WritePRG(addr, data)
	case MapperID7:
		c.mapper.(*Mapper7).WritePRG(addr, data)
	default:
		panic("unreachable")
	}
}

func (c *StaticCartridge) ReadCHR(addr uint16) byte {
	switch c.mapperID {
	case MapperID0:
		return c.mapper.(*Mapper0).ReadCHR(addr)
	case MapperID1:
		return c.mapper.(*Mapper1).ReadCHR(addr)
	case MapperID2:
		return c.mapper.(*Mapper2).ReadCHR(addr)
	case MapperID3:
		return c.mapper.(*Mapper3).ReadCHR(addr)
	case MapperID4:
		return c.mapper.(*Mapper4).ReadCHR(addr)
	case MapperID7:
		return c.mapper.(*Mapper7).ReadCHR(addr)
	default:
		panic("unreachable")
	}
}

func (c *StaticCartridge) WriteCHR(addr uint16, data byte) {
	switch c.mapperID {
	case MapperID0:
		c.mapper.(*Mapper0).WriteCHR(addr, data)
	case MapperID1:
		c.mapper.(*Mapper1).WriteCHR(addr, data)
	case MapperID2:
		c.mapper.(*Mapper2).WriteCHR(addr, data)
	case MapperID3:
		c.mapper.(*Mapper3).WriteCHR(addr, data)
	case MapperID4:
		c.mapper.(*Mapper4).WriteCHR(addr, data)
	case MapperID7:
		c.mapper.(*Mapper7).WriteCHR(addr, data)
	default:
		panic("unreachable")
	}
}

func (c *StaticCartridge) SaveState(w *binario.Writer) error {
	switch c.mapperID {
	case MapperID0:
		return c.mapper.(*Mapper0).SaveState(w)
	case MapperID1:
		return c.mapper.(*Mapper1).SaveState(w)
	case MapperID2:
		return c.mapper.(*Mapper2).SaveState(w)
	case MapperID3:
		return c.mapper.(*Mapper3).SaveState(w)
	case MapperID4:
		return c.mapper.(*Mapper4).SaveState(w)
	case MapperID7:
		return c.mapper.(*Mapper7).SaveState(w)
	default:
		panic("unreachable")
	}
}

func (c *StaticCartridge) LoadState(r *binario.Reader) error {
	switch c.mapperID {
	case MapperID0:
		return c.mapper.(*Mapper0).LoadState(r)
	case MapperID1:
		return c.mapper.(*Mapper1).LoadState(r)
	case MapperID2:
		return c.mapper.(*Mapper2).LoadState(r)
	case MapperID3:
		return c.mapper.(*Mapper3).LoadState(r)
	case MapperID4:
		return c.mapper.(*Mapper4).LoadState(r)
	case MapperID7:
		return c.mapper.(*Mapper7).LoadState(r)
	default:
		panic("unreachable")
	}
}
