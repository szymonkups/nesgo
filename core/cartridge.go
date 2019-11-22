package core

import (
	"encoding/binary"
	"fmt"
	"github.com/szymonkups/nesgo/core/mappers"
	"io"
	"os"
)

const (
	MirroringVertical = iota
	MirroringHorizontal
)

type Cartridge struct {
	mapper    mappers.Mapper
	chrMem    []uint8
	mirroring uint8
}

var allMappers = map[uint8]mappers.Mapper{
	0x00: &mappers.Mapper0{},
}

func (crt *Cartridge) GetMirroring() uint8 {
	return crt.mirroring
}

func (crt *Cartridge) Read(busId string, addr uint16, debug bool) (uint8, bool) {
	if crt.mapper != nil {
		// TODO: don't pass reading writing to the mapper - just map the address
		// TODO: let's decide it when more mappers will be in place
		return crt.mapper.Read(busId, addr, debug)
	}

	return 0x00, false
}

func (crt *Cartridge) Write(busId string, addr uint16, data uint8, debug bool) bool {
	if crt.mapper != nil {
		return crt.mapper.Write(busId, addr, data, debug)
	}

	return false
}

// https://wiki.nesdev.com/w/index.php/INES
type fileHeader struct {
	Name        [4]uint8
	PrgRomBanks uint8
	ChrRomBanks uint8
	Flags6      uint8
	Flags7      uint8
	Flags8      uint8
	Flags9      uint8
	Flags10     uint8
	Unused      [5]uint8
}

func (crt *Cartridge) LoadFile(fileName string) error {
	// TODO: extract file parsing to some utility function
	f, err := os.Open(fileName)

	if err != nil {
		return err
	}

	defer f.Close()

	header := fileHeader{}
	err = binary.Read(f, binary.BigEndian, &header)

	if err != nil {
		return err
	}

	// If trainer data is present - skip it
	if header.Flags6&0b00000100 != 0 {
		_, err := f.Seek(512, io.SeekCurrent)

		if err != nil {
			return err
		}
	}

	// Mapper number is spread across flags 6 and 7
	mapperNumber := ((header.Flags7 >> 4) << 4) | (header.Flags6 >> 4)

	// Mirroring
	crt.mirroring = MirroringHorizontal
	if header.Flags6&0b00000001 > 0 {
		crt.mirroring = MirroringVertical
	}

	// Load PRG ROM data
	prgMem := make([]uint8, int(header.PrgRomBanks)*0x4000)
	_, err = f.Read(prgMem)

	if err != nil {
		return err
	}

	// Load CHR ROM data
	crt.chrMem = make([]uint8, int(header.ChrRomBanks)*0x2000)
	_, err = f.Read(crt.chrMem)

	if err != nil {
		return err
	}

	mapper, ok := allMappers[mapperNumber]

	if !ok {
		return fmt.Errorf("mapper 0x%X not supported, yet", mapperNumber)
	}

	mapper.Initialize(header.PrgRomBanks, header.ChrRomBanks, prgMem, crt.chrMem)
	crt.mapper = mapper

	return nil
}

func (crt *Cartridge) GetCHRMem() []uint8 {
	return crt.chrMem
}
