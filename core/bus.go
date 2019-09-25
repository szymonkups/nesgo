package core

// Bus for all the components in the system
type Bus struct {
	ram [64 * 1024]uint8
}

func (bus *Bus) read(addr uint16) uint8 {
	return bus.ram[addr]
}

func (bus *Bus) read16(addr uint16) uint16 {
	low := uint16(bus.read(addr))
	high := uint16(bus.read(addr + 1))
	return (high << 8) | low
}

func (bus *Bus) write(addr uint16, val uint8) {
	bus.ram[addr] = val
}

func (bus *Bus) write16(addr uint16, val uint16) {
	bus.ram[addr] = uint8(val & 0x00FF)
	bus.ram[addr+1] = uint8((val >> 8) & 0x00FF)
}
