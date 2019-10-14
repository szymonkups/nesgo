package core

type PPU struct {
	scanLine        int16
	cycle           int16
	isFrameComplete bool
}

func (_ *PPU) Read(addr uint16) (uint8, bool) {
	if addr >= 0x2000 && addr <= 0x3FFF {
		// TODO: read from ppu registers
		// it's 7 bytes repeated
		// return ppu.read(addr & 0x0007)
		return 0x00, true
	}

	return 0x00, false
}

func (ppu *PPU) ReadDebug(addr uint16) (uint8, bool) {
	return ppu.Read(addr)
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

func (ppu *PPU) Clock() {
	ppu.cycle++

	if ppu.cycle >= 341 {
		ppu.cycle = 0
		ppu.scanLine++

		if ppu.scanLine >= 261 {
			ppu.scanLine = -1
			ppu.isFrameComplete = true
		}
	}
}

type PPUColor struct {
	R uint8
	G uint8
	B uint8
}

var palette = [0x40]*PPUColor{
	{84, 84, 84},
	{0, 30, 116},
	{8, 16, 144},
	{48, 0, 136},
	{68, 0, 100},
	{92, 0, 48},
	{84, 4, 0},
	{60, 24, 0},
	{32, 42, 0},
	{8, 58, 0},
	{0, 64, 0},
	{0, 60, 0},
	{0, 50, 60},
	{0, 0, 0},
	{0, 0, 0},
	{0, 0, 0},

	{152, 150, 152},
	{8, 76, 196},
	{48, 50, 236},
	{92, 30, 228},
	{136, 20, 176},
	{160, 20, 100},
	{152, 34, 32},
	{120, 60, 0},
	{84, 90, 0},
	{40, 114, 0},
	{8, 124, 0},
	{0, 118, 40},
	{0, 102, 120},
	{0, 0, 0},
	{0, 0, 0},
	{0, 0, 0},

	{236, 238, 236},
	{76, 154, 236},
	{120, 124, 236},
	{176, 98, 236},
	{228, 84, 236},
	{236, 88, 180},
	{236, 106, 100},
	{212, 136, 32},
	{160, 170, 0},
	{116, 196, 0},
	{76, 208, 32},
	{56, 204, 108},
	{56, 180, 204},
	{60, 60, 60},
	{0, 0, 0},
	{0, 0, 0},

	{236, 238, 236},
	{168, 204, 236},
	{188, 188, 236},
	{212, 178, 236},
	{236, 174, 236},
	{236, 174, 212},
	{236, 180, 176},
	{228, 196, 144},
	{204, 210, 120},
	{180, 222, 120},
	{168, 226, 144},
	{152, 226, 180},
	{160, 214, 228},
	{160, 162, 160},
	{0, 0, 0},
	{0, 0, 0},
}
