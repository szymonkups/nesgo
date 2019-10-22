package ppu

// https://wiki.nesdev.com/w/index.php/PPU_registers#PPUSTATUS
type StatusRegister struct {
	byteRepresentation uint8
}

func (status *StatusRegister) Write(value uint8) {
	status.byteRepresentation = value
}

func (status *StatusRegister) GetVBlank() bool {
	return status.byteRepresentation&0b10000000 > 0
}

func (status *StatusRegister) SetVBlank(value bool) {
	if value {
		status.byteRepresentation |= 0b10000000
	} else {
		status.byteRepresentation &= 0b01111111
	}
}

func (status *StatusRegister) GetSpriteZeroHit() bool {
	return status.byteRepresentation&0b01000000 > 0
}

func (status *StatusRegister) GetSpriteOverflow() bool {
	return status.byteRepresentation&0b00100000 > 0
}

func (status *StatusRegister) Read() uint8 {
	return status.byteRepresentation
}
