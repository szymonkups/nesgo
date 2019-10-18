package core

type VRam struct {
	patternTable [0x2000]uint8
	palette      [0x20]uint8
}

func (vRam *VRam) Read(_ string, addr uint16, _ bool) (uint8, bool) {
	// Wrap address on 0x4000
	addr %= 0x4000

	// Pattern table (if not handled by cartridge)
	// TODO: this is probably not needed as always provided by cartridges (I hope!)
	if addr <= 0x1FFF {
		return vRam.patternTable[addr], true
	}

	// Sprite palette.
	if addr >= 0x3F00 {
		addr = (addr & 0x00FF) % 0x20

		// Todo implement sprite/bg palette mirroring of bg color
		return vRam.palette[addr], true
	}

	return 0x00, false
}

func (vRam *VRam) Write(_ string, addr uint16, data uint8, _ bool) bool {

	return false
}
