package mappers

type Mapper0 struct {
	prgRomBanks uint8
	chrRomBanks uint8
	prgMem      []uint8
	chrMem      []uint8
}

func (mpr *Mapper0) Initialize(prgRomSize uint8, chrRomSize uint8, prgMem []uint8, chrMem []uint8) {
	mpr.prgRomBanks = prgRomSize
	mpr.chrRomBanks = chrRomSize
	mpr.prgMem = prgMem
	mpr.chrMem = chrMem
}

func (mpr *Mapper0) Read(busId string, addr uint16, _ bool) (uint8, bool) {
	if busId == "cpu" && addr >= 0x8000 {
		return mpr.prgMem[mpr.getMappedAddress(addr)], true
	}

	if busId == "ppu" && addr <= 0x1FFF {
		return mpr.chrMem[addr], true
	}

	return 0, false
}

func (mpr *Mapper0) Write(busId string, addr uint16, data uint8, _ bool) bool {
	if busId == "cpu" && addr >= 0x8000 {
		mpr.prgMem[mpr.getMappedAddress(addr)] = data
		return true
	}

	if busId == "ppu" && addr <= 0x1FFF {
		if mpr.chrRomBanks == 0 {
			mpr.chrMem[addr] = data
			return true
		}
	}

	return false
}

func (mpr *Mapper0) getMappedAddress(addr uint16) uint16 {
	if mpr.prgRomBanks > 1 {
		// 32KB
		return addr & 0x7FFF
	}

	// 16KB
	return addr & 0x3FFF
}
