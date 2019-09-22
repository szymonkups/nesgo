package core

// Bus for all the components in the system
type Bus struct {
	ram [64 * 1024]uint8
}

func (bus *Bus) read(addr uint16) uint8 {
	return bus.ram[addr]
}

func (bus *Bus) write(addr uint16, val uint8) {
	bus.ram[addr] = val
}
