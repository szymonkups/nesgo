package core

// Bus for all the components in the system
type Bus struct {
	ram [0x2000]uint8
}

func (bus *Bus) read(addr uint16) uint8 {

	// Read from ram
	if addr >= 0x0000 && addr <= 0x1FFF {
		// Memory in this range is just 0x0000 - 0x07FF cloned
		return bus.ram[addr&0x07FF]
	}

	// Read from PPU
	if addr >= 0x2000 && addr <= 0x3FFF {
		// TODO: read from ppu registers
		// it's 7 bytes repeated
		// return ppu.read(addr & 0x0007)
		return 0x00
	}

	// Read from APU && I/O registers
	if addr >= 0x4000 && addr <= 0x4017 {
		// TODO: handle APU and I/O
		return 0x00
	}

	// APU and I/O functionality that is normally disabled.
	// https://wiki.nesdev.com/w/index.php/CPU_memory_map
	if addr >= 0x4018 && addr <= 0x401f {
		return 0x00
	}

	// Cartridge space
	if addr >= 0x4020 {
		// TODO: handle Cartridge
		return 0x00
	}

	// TODO: we should never get here - handle gracefully when out of range access is performed
	return 0x00
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
