package core

type PPU struct {}

func (_ *PPU) Read(addr uint16) (uint8, bool) {
	if addr >= 0x2000 && addr <= 0x3FFF {
		// TODO: read from ppu registers
		// it's 7 bytes repeated
		// return ppu.read(addr & 0x0007)
		return 0x00, true
	}

	return 0x00, false
}

func (_ *PPU) Write(addr uint16, data uint8) bool {
	if addr >= 0x2000 && addr <= 0x3FFF {
		// TODO: write to ppu registers
		// it's 7 bytes repeated
		// return ppu.write(addr & 0x0007, data)
		return true
	}

	return false
}