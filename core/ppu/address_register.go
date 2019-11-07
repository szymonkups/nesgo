package ppu

type AddressRegister struct {
	address uint16
	latch   bool
}

func (addr *AddressRegister) Write(value uint8) {
	if !addr.latch {
		addr.address = (uint16(value&0x3F) << 8) | (addr.address & 0x00FF)
		addr.latch = true
	} else {
		addr.address = (addr.address & 0xFF00) | uint16(value)
		addr.latch = false
	}
}

func (addr *AddressRegister) GetAddress() uint16 {
	return addr.address
}

func (addr *AddressRegister) ResetLatch() {
	addr.latch = false
}

func (addr *AddressRegister) Increment() {
	addr.address++
}
