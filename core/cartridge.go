package core

import (
	"encoding/binary"
	"fmt"
	"github.com/szymonkups/nesgo/core/mappers"
	"os"
)

type Cartridge struct {
	mapper mappers.Mapper
}

var allMappers = map[uint8]mappers.Mapper{
	0x00: &mappers.Mapper0{},
}

func (crt *Cartridge) Read(addr uint16) (uint8, bool) {
	if crt.mapper != nil {
		return crt.mapper.Read(addr)
	}

	return 0x00, false
}

func (crt *Cartridge) ReadDebug(addr uint16) (uint8, bool) {
	return crt.Read(addr)
}

func (crt *Cartridge) Write(addr uint16, data uint8) bool {
	if crt.mapper != nil {
		return crt.mapper.Write(addr, data)
	}

	return false
}

func (crt *Cartridge) ppuRead(addr uint16) (uint8, bool) {
	if crt.mapper != nil {
		return crt.mapper.PPURead(addr)
	}

	return 0x00, false
}

func (crt *Cartridge) ppuWrite(addr uint16, data uint8) bool {
	if crt.mapper != nil {
		return crt.mapper.PPUWrite(addr, data)
	}

	return false
}

// https://wiki.nesdev.com/w/index.php/INES
type fileHeader struct {
	Name       [4]uint8
	PrgRomSize uint8
	ChrRomSize uint8
	Flags6     uint8
	Flags7     uint8
	Flags8     uint8
	Flags9     uint8
	Flags10    uint8
	Unused     [5]uint8
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
		_, err := f.Seek(512, os.SEEK_CUR)

		if err != nil {
			return err
		}
	}

	// Mapper number is spread across flags 6 and 7
	mapperNumber := ((header.Flags7 >> 4) << 4) | (header.Flags6 >> 4)

	// Load PRG ROM data
	prgMem := make([]uint8, int(header.PrgRomSize)*16384)
	_, err = f.Read(prgMem)

	if err != nil {
		return err
	}

	// Load CHR ROM data
	chrMem := make([]uint8, int(header.ChrRomSize)*8192)
	_, err = f.Read(chrMem)

	if err != nil {
		return err
	}

	mapper, ok := allMappers[mapperNumber]

	if !ok {
		return fmt.Errorf("mapper 0x%X not supported, yet", mapperNumber)
	}

	mapper.Initialize(header.PrgRomSize, header.ChrRomSize, prgMem, chrMem)
	crt.mapper = mapper

	return nil
}
