package ppu

// https://wiki.nesdev.com/w/index.php/PPU_registers
// TODO: maybe make all fields private and use getters?
type ControlRegister struct {
	value uint8
}

func (ctrl *ControlRegister) Write(value uint8) {
	ctrl.value = value
	//
	//ctrl.NMIEnable = value&0b10000000 > 0
	//ctrl.MasterSlaveSelect = value&0b01000000 > 0
	//
	//ctrl.SpriteWidth = 8
	//if value&0b00100000 == 0 {
	//	ctrl.SpriteHeight = 8
	//} else {
	//	ctrl.SpriteHeight = 16
	//}
	//
	//if value&0b00010000 == 0 {
	//	ctrl.BgPatternTableAddress = 0x0000
	//} else {
	//	ctrl.BgPatternTableAddress = 0x1000
	//}
	//
	//if value&0b00001000 == 0 {
	//	ctrl.SpritePatternTableAddress = 0x0000
	//} else {
	//	ctrl.SpritePatternTableAddress = 0x1000
	//}
	//
	//ctrl.IncrementMode = value&0b00000100 != 0
	//
	//switch value & 0b00000011 {
	//case 0:
	//	ctrl.NameTableAddress = 0x2000
	//case 1:
	//	ctrl.NameTableAddress = 0x2400
	//case 2:
	//	ctrl.NameTableAddress = 0x2800
	//case 3:
	//	ctrl.NameTableAddress = 0x2C00
	//}
}

func (ctrl *ControlRegister) GetNMIEnable() bool {
	return ctrl.value&0b10000000 > 0
}

func (ctrl *ControlRegister) GetIncrementMode() bool {
	return ctrl.value&0b00000100 != 0
}

func (ctrl *ControlRegister) GetNameTableX() uint8 {
	return ctrl.value & 0b00000001
}

func (ctrl *ControlRegister) GetNameTableY() uint8 {
	return (ctrl.value & 0b00000010) >> 1
}

func (ctrl *ControlRegister) Read() uint8 {
	return ctrl.value
}
