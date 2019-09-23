package core

// Addressing mode functions returns data for an instruction
// Also give information if there is potential to add cycle to the method
type addressingMode func(*CPU) (uint8, bool)

// Accumulator addressing - accumulator is used as argument
func accumulatorAddressing(cpu *CPU) (uint8, bool) {
	return cpu.a, false
}

// Implied addressing there is no data involved, return 0 as it is not used
func impliedAddr(_ *CPU) (uint8, bool) {
	return 0, false
}

// Immediate addressing means that data is next byte after the op code
// Read the value and increment the Program Counter
func immediateAddr(cpu *CPU) (uint8, bool) {
	data := cpu.bus.read(cpu.pc)
	cpu.pc++

	return data, false
}

// Zero page addressing means that next byte after op code is an offset
// inside zero page.
// Memory is represented as uint16 - two bytes: higer one is page, lower one is
// offset. Zero page addressing means that only one byte can be used to
// get value from first memory page.
func zeroPageAddr(cpu *CPU) (uint8, bool) {
	var offset uint8 = cpu.bus.read(cpu.pc)
	cpu.pc++

	data := cpu.bus.read(uint16(offset))
	return data, false
}

// Same as zero page addressing but offset is calculated as sum of readed data
// and current X register. If it exceeds byte - it wraps around so no page
// crossing is allowed
func zeroPageXAddr(cpu *CPU) (uint8, bool) {
	var offset uint8 = cpu.bus.read(cpu.pc) + cpu.x
	cpu.pc++

	data := cpu.bus.read(uint16(offset))
	return data, false
}

// Same as zero page addressing but offset is calculated as sum of readed data
// and current Y register. If it exceeds byte - it wraps around so no page
// crossing is allowed
func zeroPageYAddr(cpu *CPU) (uint8, bool) {
	var offset uint8 = cpu.bus.read(cpu.pc) + cpu.y
	cpu.pc++

	data := cpu.bus.read(uint16(offset))
	return data, false
}

// Relative addressing - next byte after op code is an relative (-128 ro 127)
// number that is added to the program counter.
// We need to read next byte and if most significant bit is set - w substract it
// from PC, otherwise we add it.
func relativeAddr(cpu *CPU) (uint8, bool) {
	uOffset := cpu.bus.read(cpu.pc)
	cpu.pc++

	return uOffset, false
}

// Absolute addressing - next two bytes represents lower and higher bytes
// of the absolute address.
func absoluteAddr(cpu *CPU) (uint8, bool) {
	low := cpu.bus.read(cpu.pc)
	cpu.pc++
	high := cpu.bus.read(cpu.pc)
	cpu.pc++

	var addr uint16 = (uint16(high) << 8) | uint16(low)
	data := cpu.bus.read(addr)

	return data, false
}

// Same as absolute addresing but with adding X register to the result
func absoluteXAddr(cpu *CPU) (uint8, bool) {
	low := cpu.bus.read(cpu.pc)
	cpu.pc++
	high := cpu.bus.read(cpu.pc)
	cpu.pc++

	high16 := (uint16(high) << 8)
	var addr uint16 = (high16 | uint16(low)) + uint16(cpu.x)
	data := cpu.bus.read(addr)

	// Check if after adding X register we crossed a page
	pageCrossed := (addr & 0xFF00) != high16

	return data, pageCrossed
}

// Same as absolute addresing but with adding X register to the result
func absoluteYAddr(cpu *CPU) (uint8, bool) {
	low := cpu.bus.read(cpu.pc)
	cpu.pc++
	high := cpu.bus.read(cpu.pc)
	cpu.pc++

	high16 := (uint16(high) << 8)
	var addr uint16 = (high16 | uint16(low)) + uint16(cpu.y)
	data := cpu.bus.read(addr)

	// Check if after adding Y register we crossed a page
	pageCrossed := (addr & 0xFF00) != high16

	return data, pageCrossed
}

// Indirect addressing - something like pointers to data. There are two bytes
// after the op code determining when actual address to the data is stored.
// First we read two bytes using program counter to calculate pointer address
// and then from that place we read two bytes to get the actual data address.
func indirectAddr(cpu *CPU) (uint8, bool) {
	low := cpu.bus.read(cpu.pc)
	cpu.pc++
	high := cpu.bus.read(cpu.pc)
	cpu.pc++

	var pointerAddr uint16 = (uint16(high) << 8) | uint16(low)

	// There is a hardware but when reading from xxFF address. By adding 1 to
	// it (to read more significant byte of data address) we should cross pages
	// but instead it is fetched from xx00 :)
	dataLow := cpu.bus.read(pointerAddr)
	var dataHigh uint8

	if low == 0xFF {
		// Simulate bug
		dataHigh = cpu.bus.read(pointerAddr & 0xFF00)
	} else {
		dataHigh = cpu.bus.read(pointerAddr + 1)
	}

	var dataAddr uint16 = (uint16(dataHigh) << 8) | uint16(dataLow)
	data := cpu.bus.read(dataAddr)

	return data, false
}

// Indirect X addresing - page zero addressing where we need to add X register
// to next byte after op code (without carry, ex. 200 + 66 = 10) to obtain low
// byte of the actual data address.
// High byte is next to it.
func indirectXAddr(cpu *CPU) (uint8, bool) {
	arg := cpu.bus.read(cpu.pc)
	cpu.pc++

	low := arg + cpu.x
	high := low + 1

	var dataAddr uint16 = (uint16(high) << 8) | uint16(low)
	data := cpu.bus.read(dataAddr)

	return data, false
}

// Indirect Y addressing - different than indirect X addressing.
// Next byte after op code points to is an offset in the zero page from where
// we reads two bytes to compose base addres. We add Y register to that address
// and this is where data is stored. We need to check if page is crossed after
// adding Y register.
func indirectYAddr(cpu *CPU) (uint8, bool) {
	zeroLow := cpu.bus.read(cpu.pc)
	cpu.pc++

	low := cpu.bus.read(uint16(zeroLow))
	high := cpu.bus.read(uint16(zeroLow + 1))
	high16 := (uint16(high) << 8)

	var dataAddr uint16 = (high16 | uint16(low)) + uint16(cpu.y)
	data := cpu.bus.read(dataAddr)

	// Check if after adding Y register we crossed a page
	pageCrossed := (dataAddr & 0xFF00) != high16

	return data, pageCrossed
}
