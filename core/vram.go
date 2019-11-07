package core

import "fmt"

type VRam struct {
	patternTable [0x2000]uint8
	palette      [0x20]uint8
	nametable    [0x1000]uint8
}

var paletteMappings = map[uint16]uint16{
	0x3F10: 0x3F00,
	0x3F14: 0x3F04,
	0x3F18: 0x3F08,
	0x3F1C: 0x3F0C,
}

func (vRam *VRam) Read(_ string, addr uint16, _ bool) (uint8, bool) {
	// Wrap address on 0x4000
	addr %= 0x4000

	// Pattern table (if not handled by cartridge)
	// TODO: this is probably not needed as always provided by cartridges (I hope!)
	if addr <= 0x1FFF {
		return vRam.patternTable[addr], true
	}

	if addr >= 0x2000 && addr < 0x3F00 {
		addr = (addr - 0x2000) % 0x1000

		return vRam.nametable[addr], true
	}

	// Sprite palette.
	if addr >= 0x3F00 {
		addr = (addr & 0x00FF) % 0x20

		// https://wiki.nesdev.com/w/index.php/PPU_palettes
		// Addresses $3F10/$3F14/$3F18/$3F1C are mirrors of $3F00/$3F04/$3F08/$3F0C.
		// Note that this goes for writing as well as reading. A symptom of not having implemented this correctly
		// in an emulator is the sky being black in Super Mario Bros., which writes the backdrop color through $3F10.
		mapping, ok := paletteMappings[addr]
		if ok {
			addr = mapping
		}

		return vRam.palette[addr], true
	}

	return 0x00, false
}

func (vRam *VRam) Write(_ string, addr uint16, data uint8, _ bool) bool {
	// Wrap address on 0x4000
	addr %= 0x4000

	// Pattern table (if not handled by cartridge)
	// TODO: this is probably not needed as always provided by cartridges (I hope!)
	if addr <= 0x1FFF {
		vRam.patternTable[addr] = data
		return true
	}

	if addr >= 0x2000 && addr < 0x3F00 {
		addr = (addr - 0x2000) % 0x1000
		vRam.nametable[addr] = data

		return true
	}

	// Sprite palette.
	if addr >= 0x3F00 {
		fmt.Println("writing to palette mem")
		addr = (addr & 0x00FF) % 0x20

		// https://wiki.nesdev.com/w/index.php/PPU_palettes
		// Addresses $3F10/$3F14/$3F18/$3F1C are mirrors of $3F00/$3F04/$3F08/$3F0C.
		// Note that this goes for writing as well as reading. A symptom of not having implemented this correctly
		// in an emulator is the sky being black in Super Mario Bros., which writes the backdrop color through $3F10.
		mapping, ok := paletteMappings[addr]
		if ok {
			addr = mapping
		}

		vRam.palette[addr] = data

		return true
	}

	return false
}
