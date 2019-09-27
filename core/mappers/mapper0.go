package mappers

type Mapper0 struct {
	prgRomSize uint8
	chrRomSize uint8
	prgMem     []uint8
	chrMem     []uint8
}

func (mpr *Mapper0) Initialize(prgRomSize uint8, chrRomSize uint8, prgMem []uint8, chrMem []uint8) {
	mpr.prgRomSize = prgRomSize
	mpr.chrRomSize = chrRomSize
	mpr.prgMem = prgMem
	mpr.chrMem = chrMem
}

func (mpr *Mapper0) Read(addr uint16) (uint8, bool) {
	if addr >= 0x8000 {
		return mpr.prgMem[mpr.getMappedAddress(addr)], true
	}
	return 0, false
}

func (mpr *Mapper0) Write(addr uint16, data uint8) bool {
	if addr >= 0x8000 {
		mpr.prgMem[mpr.getMappedAddress(addr)] = data

		return true
	}

	return false
}

func (mpr *Mapper0) getMappedAddress(addr uint16) uint16 {
	if mpr.prgRomSize > 1 {
		// 32KB
		return addr & 0x7FFF
	}

	// 16KB
	return addr & 0x3FFF
}

func (mpr *Mapper0) PPURead(addr uint16) (uint8, bool) {
	if addr <= 0x1FFF {
		return mpr.chrMem[addr], true
	}

	return 0x00, false
}

func (mpr *Mapper0) PPUWrite(addr uint16, data uint8) bool {
	if addr <= 0x1FFF {
		if mpr.chrRomSize == 0 {
			mpr.chrMem[addr] = data

			return true
		}
	}

	return false
}
