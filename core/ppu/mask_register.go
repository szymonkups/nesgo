package ppu

type MaskRegister struct {
	GreyScale       bool
	ShowBgLeft      bool
	ShowSpritesLeft bool
	ShowBg          bool
	ShowSprites     bool
	EmphasizeRed    bool
	EmphasizeGreen  bool
	EmphasizeBlue   bool

	byteRepresentation uint8
}

func (mask *MaskRegister) Write(value uint8) {
	mask.byteRepresentation = value

	mask.GreyScale = value&0b00000001 > 0
	mask.ShowBgLeft = value&0b00000010 > 0
	mask.ShowSprites = value&0b00000100 > 0
	mask.ShowBg = value&0b00001000 > 0
	mask.ShowSprites = value&0b00010000 > 0
	mask.EmphasizeRed = value&0b00100000 > 0
	mask.EmphasizeGreen = value&0b01000000 > 0
	mask.EmphasizeBlue = value&0b10000000 > 0
}

func (mask *MaskRegister) Read() uint8 {
	return mask.byteRepresentation
}
