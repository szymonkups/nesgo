package core

const (
	accumulatorAddressing = iota
	impliedAddressing
	immediateAddressing
	zeroPageAddressing
	zeroPageXAddressing
	zeroPageYAddressing
	relativeAddressing
	absoluteAddressing
	absoluteXAddressing
	absoluteYAddressing
	indirectAddressing
	indirectXAddressing
	indirectYAddressing
)

// Addressing mode functions returns data for an instruction
// Also give information if there is potential to add cycle to the method
type addressingMode func(*CPU) (address uint16, addCycle bool)

var addressingModes = map[int]addressingMode{
	// Accumulator addressing - accumulator is used as argument, return 0 as
	// no addressing is used
	accumulatorAddressing: func(cpu *CPU) (uint16, bool) {
		return 0, false
	},

	// Implied addressing means there is addressing involved,
	// return 0 as it is not used
	impliedAddressing: func(_ *CPU) (uint16, bool) {
		return 0, false
	},

	// Immediate addressing means that data is next byte after the op code
	// Read the value and increment the Program Counter
	immediateAddressing: func(cpu *CPU) (uint16, bool) {
		// Address is current PC location
		address := cpu.pc

		// Move program counter forward.
		cpu.pc++

		return address, false
	},

	// Zero page addressing means that next byte after op code is an offset
	// inside zero page.
	// Memory is represented as uint16 - two bytes: higher one is page, lower one is
	// offset. Zero page addressing means that only one byte can be used to
	// get value from first memory page.
	zeroPageAddressing: func(cpu *CPU) (uint16, bool) {
		var offset uint8 = cpu.bus.Read(cpu.pc)
		cpu.pc++

		return uint16(offset), false
	},

	// Same as zero page addressing but offset is calculated as sum of read data
	// and current X register. If it exceeds byte - it wraps around so no page
	// crossing is allowed
	zeroPageXAddressing: func(cpu *CPU) (uint16, bool) {
		var offset uint8 = cpu.bus.Read(cpu.pc) + cpu.x
		cpu.pc++

		return uint16(offset), false
	},

	// Same as zero page addressing but offset is calculated as sum of read data
	// and current Y register. If it exceeds byte - it wraps around so no page
	// crossing is allowed
	zeroPageYAddressing: func(cpu *CPU) (uint16, bool) {
		var offset uint8 = cpu.bus.Read(cpu.pc) + cpu.y
		cpu.pc++

		return uint16(offset), false
	},

	// Relative addressing - next byte after op code is an relative (-128 to 127)
	// number that is added to the program counter.
	// We need to read next byte and
	relativeAddressing: func(cpu *CPU) (uint16, bool) {
		uOffset := cpu.bus.Read(cpu.pc)
		cpu.pc++

		// Calculate absolute position based on PC
		addr := cpu.pc + uint16(uOffset)
		if uOffset >= 0x80 {
			addr -= 0x100
		}

		return addr, false
	},

	// Absolute addressing - next two bytes represents lower and higher bytes
	// of the absolute address.
	absoluteAddressing: func(cpu *CPU) (uint16, bool) {
		low := cpu.bus.Read(cpu.pc)
		cpu.pc++
		high := cpu.bus.Read(cpu.pc)
		cpu.pc++

		var addr uint16 = (uint16(high) << 8) | uint16(low)

		return addr, false
	},

	// Same as absolute addresing but with adding X register to the result
	absoluteXAddressing: func(cpu *CPU) (uint16, bool) {
		low := cpu.bus.Read(cpu.pc)
		cpu.pc++
		high := cpu.bus.Read(cpu.pc)
		cpu.pc++

		high16 := (uint16(high) << 8)
		var addr uint16 = (high16 | uint16(low)) + uint16(cpu.x)

		// Check if after adding X register we crossed a page
		pageCrossed := (addr & 0xFF00) != high16

		return addr, pageCrossed
	},

	// Same as absolute addresing but with adding X register to the result
	absoluteYAddressing: func(cpu *CPU) (uint16, bool) {
		low := cpu.bus.Read(cpu.pc)
		cpu.pc++
		high := cpu.bus.Read(cpu.pc)
		cpu.pc++

		high16 := (uint16(high) << 8)
		var addr uint16 = (high16 | uint16(low)) + uint16(cpu.y)

		// Check if after adding Y register we crossed a page
		pageCrossed := (addr & 0xFF00) != high16

		return addr, pageCrossed
	},

	// Indirect addressing - something like pointers to data. There are two bytes
	// after the op code determining when actual address to the data is stored.
	// First we read two bytes using program counter to calculate pointer address
	// and then from that place we read two bytes to get the actual data address.
	indirectAddressing: func(cpu *CPU) (uint16, bool) {
		low := cpu.bus.Read(cpu.pc)
		cpu.pc++
		high := cpu.bus.Read(cpu.pc)
		cpu.pc++

		var pointerAddr uint16 = (uint16(high) << 8) | uint16(low)

		// There is a hardware bug when reading from xxFF address. By adding 1 to
		// it (to read more significant byte of data address) we should cross
		// pages but instead it is fetched from xx00 :)
		dataLow := cpu.bus.Read(pointerAddr)
		var dataHigh uint8

		if low == 0xFF {
			// Simulate bug
			dataHigh = cpu.bus.Read(pointerAddr & 0xFF00)
		} else {
			dataHigh = cpu.bus.Read(pointerAddr + 1)
		}

		var addr uint16 = (uint16(dataHigh) << 8) | uint16(dataLow)

		return addr, false
	},

	// Indirect X addresing - page zero addressing where we need to add X register
	// to next byte after op code (without carry, ex. 200 + 66 = 10) to obtain low
	// byte of the actual data address.
	// High byte is next to it.
	indirectXAddressing: func(cpu *CPU) (uint16, bool) {
		arg := cpu.bus.Read(cpu.pc)
		cpu.pc++

		low := arg + cpu.x
		high := low + 1

		var addr uint16 = (uint16(high) << 8) | uint16(low)

		return addr, false
	},

	// Indirect Y addressing - different than indirect X addressing.
	// Next byte after op code points to is an offset in the zero page from where
	// we reads two bytes to compose base addres. We add Y register to that address
	// and this is where data is stored. We need to check if page is crossed after
	// adding Y register.
	indirectYAddressing: func(cpu *CPU) (uint16, bool) {
		zeroLow := cpu.bus.Read(cpu.pc)
		cpu.pc++

		low := cpu.bus.Read(uint16(zeroLow))
		high := cpu.bus.Read(uint16(zeroLow + 1))
		high16 := (uint16(high) << 8)

		var addr uint16 = (high16 | uint16(low)) + uint16(cpu.y)

		// Check if after adding Y register we crossed a page
		pageCrossed := (addr & 0xFF00) != high16

		return addr, pageCrossed
	},
}
