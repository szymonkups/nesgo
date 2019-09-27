package core

type Ram struct {
	data [0x2000]uint8
}

func (ram *Ram) Read(addr uint16) (uint8, bool) {
	// Check if RAM range
	if addr <= 0x1FFF {
		//Memory in this range is just 0x0000 - 0x07FF cloned
		return ram.data[addr&0x07FF], true
	}

	return 0x00, false
}

func (ram *Ram) Write(addr uint16, data uint8) bool {
	// Check if RAM range
	if addr <= 0x1FFF {
		//Memory in this range is just 0x0000 - 0x07FF cloned
		ram.data[addr&0x07FF] = data

		return true
	}

	return false
}



