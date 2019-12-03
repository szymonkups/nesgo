package core

type vRam struct {
	crt          *Cartridge
	patternTable [0x2000]uint8
	palette      [0x20]uint8
	nameTable1   [0x400]uint8
	nameTable2   [0x400]uint8
}

func NewVRam(crt *Cartridge) *vRam {
	v := new(vRam)
	v.crt = crt

	return v
}

func (vRam *vRam) Read(_ string, addr uint16, _ bool) (uint8, bool) {
	// Wrap address on 0x4000
	addr %= 0x4000

	// Pattern table (if not handled by cartridge)
	// TODO: this is probably not needed as always provided by cartridges (I hope!)
	if addr <= 0x1FFF {
		return vRam.patternTable[addr], true
	}

	if addr >= 0x2000 && addr < 0x3F00 {
		addr &= 0x0FFF

		if vRam.crt.GetMirroring() == MirroringVertical {
			if (addr <= 0x03FF) || (addr >= 0x0800 && addr <= 0x0BFF) {
				return vRam.nameTable1[addr&0x03FF], true
			}

			if (addr >= 0x0400 && addr <= 0x07FF) || (addr >= 0x0C00 && addr <= 0x0FF) {
				return vRam.nameTable2[addr&0x03FF], true
			}
		} else if vRam.crt.GetMirroring() == MirroringHorizontal {
			if (addr <= 0x03FF) || (addr >= 0x0400 && addr <= 0x07FF) {
				return vRam.nameTable1[addr&0x03FF], true
			}

			if (addr >= 0x0800 && addr <= 0x0BFF) || addr >= 0x0C00 && addr <= 0x0FFF {
				return vRam.nameTable2[addr&0x03FF], true
			}
		}
	}

	// Sprite palette.
	if addr >= 0x3F00 {
		// https://wiki.nesdev.com/w/index.php/PPU_palettes
		// Addresses $3F10/$3F14/$3F18/$3F1C are mirrors of $3F00/$3F04/$3F08/$3F0C.
		// Note that this goes for writing as well as reading. A symptom of not having implemented this correctly
		// in an emulator is the sky being black in Super Mario Bros., which writes the backdrop color through $3F10.
		addr = mapPalette(addr)
		addr = (addr & 0x00FF) % 0x20

		return vRam.palette[addr], true
	}

	return 0x00, false
}

func (vRam *vRam) Write(_ string, addr uint16, data uint8, _ bool) bool {
	// Wrap address on 0x4000
	addr %= 0x4000

	// Pattern table (if not handled by cartridge)
	// TODO: this is probably not needed as always provided by cartridges (I hope!)
	if addr <= 0x1FFF {
		vRam.patternTable[addr] = data
		return true
	}

	if addr >= 0x2000 && addr < 0x3F00 {
		addr &= 0x0FFF

		if vRam.crt.GetMirroring() == MirroringVertical {
			if (addr <= 0x03FF) || (addr >= 0x0800 && addr <= 0x0BFF) {
				vRam.nameTable1[addr&0x03FF] = data
				return true
			}

			if (addr >= 0x0400 && addr <= 0x07FF) || (addr >= 0x0C00 && addr <= 0x0FF) {
				vRam.nameTable2[addr&0x03FF] = data
				return true
			}
		} else if vRam.crt.GetMirroring() == MirroringHorizontal {
			if (addr <= 0x03FF) || (addr >= 0x0400 && addr <= 0x07FF) {
				vRam.nameTable1[addr&0x03FF] = data
				return true
			}

			if (addr >= 0x0800 && addr <= 0x0BFF) || addr >= 0x0C00 && addr <= 0x0FFF {
				vRam.nameTable2[addr&0x03FF] = data
				return true
			}
		}
	}

	// Sprite palette.
	if addr >= 0x3F00 {
		// https://wiki.nesdev.com/w/index.php/PPU_palettes
		// Addresses $3F10/$3F14/$3F18/$3F1C are mirrors of $3F00/$3F04/$3F08/$3F0C.
		// Note that this goes for writing as well as reading. A symptom of not having implemented this correctly
		// in an emulator is the sky being black in Super Mario Bros., which writes the backdrop color through $3F10.
		addr = mapPalette(addr)

		addr = (addr & 0x00FF) % 0x20
		vRam.palette[addr] = data

		return true
	}

	return false
}

func mapPalette(addr uint16) uint16 {
	switch addr {
	case 0x3F10:
		return 0x3F00
	case 0x3F14:
		return 0x3F04
	case 0x3F18:
		return 0x3F08
	case 0x3F1C:
		return 0x3F0C
	}

	return addr
}
