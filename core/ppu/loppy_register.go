package ppu

type LoopyRegister struct {
	value uint16
}

func (l *LoopyRegister) GetCoarseX() uint8 {
	return uint8(l.value & 0b0000000000011111)
}

func (l *LoopyRegister) SetCoarseX(v uint8) {
	l.value = (l.value & 0b1111111111100000) | uint16(v&0b00011111)
}

func (l *LoopyRegister) GetCoarseY() uint8 {
	return uint8((l.value & 0b0000001111100000) >> 5)
}

func (l *LoopyRegister) SetCoarseY(v uint8) {
	l.value = (l.value & 0b1111110000011111) | (uint16(v&0b00011111) << 5)
}

func (l *LoopyRegister) GetNameTableX() uint8 {
	return uint8((l.value & 0b0000010000000000) >> 10)
}

func (l *LoopyRegister) SetNameTableX(v uint8) {
	l.value = (l.value & 0b1111101111111111) | (uint16(v&0b00000001) << 10)
}

func (l *LoopyRegister) GetNameTableY() uint8 {
	return uint8((l.value & 0b0000100000000000) >> 11)
}

func (l *LoopyRegister) SetNameTableY(v uint8) {
	l.value = (l.value & 0b1111011111111111) | (uint16(v&0b00000001) << 11)
}

func (l *LoopyRegister) GetFineY() uint8 {
	return uint8((l.value & 0b0111000000000000) >> 12)
}

func (l *LoopyRegister) SetFineY(v uint8) {
	l.value = (l.value & 0b1000111111111111) | (uint16(v&0b00000111) << 12)
}

func (l *LoopyRegister) Read() uint16 {
	return l.value
}

func (l *LoopyRegister) Write(value uint16) {
	l.value = value
}

func (l *LoopyRegister) Increment(mode bool) {
	if mode {
		l.value += 32
	} else {
		l.value += 1
	}
}
