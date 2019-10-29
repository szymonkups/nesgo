package addressing

import "fmt"

const (
	AccumulatorAddressing = iota
	ImpliedAddressing
	ImmediateAddressing
	ZeroPageAddressing
	ZeroPageXAddressing
	ZeroPageYAddressing
	RelativeAddressing
	AbsoluteAddressing
	AbsoluteXAddressing
	AbsoluteYAddressing
	IndirectAddressing
	IndirectXAddressing
	IndirectYAddressing
)

func GetAddressingById(id int) (*addressingMode, bool) {
	mode, ok := addressingModes[id]

	return mode, ok
}

type addressingMode struct {
	// Short Name of the addressing mode
	Name string

	// Size in bytes of addressing mode (including opcode)
	Size uint8

	// String Format used to present address
	Format func(address uint16) string

	// Addressing mode functions returns data for an instruction
	// Also give information if there is potential to add cycle to the method
	CalculateAddress func(pc uint16, x, y uint8, read ReadFunction) (address uint16, addCycle bool)
}

type ReadFunction func(addr uint16) uint8

var addressingModes = map[int]*addressingMode{
	AccumulatorAddressing: {
		Name:   "ACC",
		Size:   1,
		Format: func(address uint16) string { return "A" },

		// Accumulator addressing - accumulator is used as argument, return 0 as
		// no addressing is used
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			return 0, false
		},
	},

	ImpliedAddressing: {
		Name:   "IMP",
		Size:   1,
		Format: func(address uint16) string { return "" },

		// Implied addressing means there is addressing involved,
		// return 0 as it is not used
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			return 0, false
		},
	},

	ImmediateAddressing: {
		Name:   "IMM",
		Size:   2,
		Format: func(address uint16) string { return fmt.Sprintf("#$%02X", address) },

		// Immediate addressing means that data is next byte after the op code
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			address := pc + 1

			return address, false
		},
	},

	ZeroPageAddressing: {
		Name:   "ZPA",
		Size:   2,
		Format: func(address uint16) string { return fmt.Sprintf("$%02X", address) },

		// Zero page addressing means that next byte after op code is an offset
		// inside zero page.
		// Memory is represented as uint16 - two bytes: higher one is page, lower one is
		// offset. Zero page addressing means that only one byte can be used to
		// get value from first memory page.
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			var offset = read(pc + 1)

			return uint16(offset), false
		},
	},

	ZeroPageXAddressing: {
		Name:   "ZPX",
		Size:   2,
		Format: func(address uint16) string { return fmt.Sprintf("$%02X,X", address) },

		// Same as zero page addressing but offset is calculated as sum of read data
		// and current X register. If it exceeds byte - it wraps around so no page
		// crossing is allowed
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			var offset = read(pc+1) + x

			return uint16(offset), false
		},
	},

	ZeroPageYAddressing: {
		Name:   "ZPY",
		Size:   2,
		Format: func(address uint16) string { return fmt.Sprintf("$%02X,Y", address) },

		// Same as zero page addressing but offset is calculated as sum of read data
		// and current Y register. If it exceeds byte - it wraps around so no page
		// crossing is allowed
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			var offset = read(pc+1) + y

			return uint16(offset), false
		},
	},

	RelativeAddressing: {
		Name:   "REL",
		Size:   2,
		Format: func(address uint16) string { return fmt.Sprintf("$%02X", address) },

		// Relative addressing - next byte after op code is an relative (-128 to 127)
		// number that is added to the program counter.
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			uOffset := read(pc + 1)

			// Calculate absolute position based on PC
			addr := pc + 2 + uint16(uOffset)
			if uOffset >= 0x80 {
				addr -= 0x100
			}

			return addr, false
		},
	},

	AbsoluteAddressing: {
		Name:   "ABS",
		Size:   3,
		Format: func(address uint16) string { return fmt.Sprintf("$%04X", address) },

		// Absolute addressing - next two bytes represents lower and higher bytes
		// of the absolute address.
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			low := read(pc + 1)
			high := read(pc + 2)

			var addr = (uint16(high) << 8) | uint16(low)

			return addr, false
		},
	},

	AbsoluteXAddressing: {
		Name:   "ABX",
		Size:   3,
		Format: func(address uint16) string { return fmt.Sprintf("$%04X,X", address) },

		// Same as absolute addressing but with adding X register to the result
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			low := read(pc + 1)
			high := read(pc + 2)

			high16 := uint16(high) << 8
			var addr = (high16 | uint16(low)) + uint16(x)

			// Check if after adding X register we crossed a page
			pageCrossed := (addr & 0xFF00) != high16

			return addr, pageCrossed
		},
	},

	AbsoluteYAddressing: {
		Name:   "ABY",
		Size:   3,
		Format: func(address uint16) string { return fmt.Sprintf("$%04X,Y", address) },

		// Same as absolute addressing but with adding X register to the result
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			low := read(pc + 1)
			high := read(pc + 2)

			high16 := uint16(high) << 8
			var addr = (high16 | uint16(low)) + uint16(y)

			// Check if after adding Y register we crossed a page
			pageCrossed := (addr & 0xFF00) != high16

			return addr, pageCrossed
		},
	},

	IndirectAddressing: {
		Name:   "IND",
		Size:   3,
		Format: func(address uint16) string { return fmt.Sprintf("($%04X)", address) },

		// Indirect addressing - something like pointers to data. There are two bytes
		// after the op code determining when actual address to the data is stored.
		// First we read two bytes using program counter to calculate pointer address
		// and then from that place we read two bytes to get the actual data address.
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			low := read(pc + 1)
			high := read(pc + 2)

			var pointerAddr = (uint16(high) << 8) | uint16(low)

			// There is a hardware bug when reading from xxFF address. By adding 1 to
			// it (to read more significant byte of data address) we should cross
			// pages but instead it is fetched from xx00 :)
			finalLow := read(pointerAddr)
			var finalHigh uint8

			if low == 0xFF {
				// Simulate bug
				finalHigh = read(pointerAddr & 0xFF00)
			} else {
				finalHigh = read(pointerAddr + 1)
			}

			var addr = (uint16(finalHigh) << 8) | uint16(finalLow)

			return addr, false
		},
	},

	IndirectXAddressing: {
		Name:   "INX",
		Size:   2,
		Format: func(address uint16) string { return fmt.Sprintf("($%02X,X)", address) },

		// Indirect X addressing - page zero addressing where we need to add X register
		// to next byte after op code (without carry, ex. 200 + 66 = 10) to obtain low
		// byte of the actual data address.
		// High byte is next to it.
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			arg := read(pc + 1)

			low := arg + x
			high := low + 1

			var addr = (uint16(high) << 8) | uint16(low)

			return addr, false
		},
	},

	IndirectYAddressing: {
		Name:   "INY",
		Size:   2,
		Format: func(address uint16) string { return fmt.Sprintf("($%02X),Y", address) },

		// Indirect Y addressing - different than indirect X addressing.
		// Next byte after op code points to is an offset in the zero page from where
		// we reads two bytes to compose base address. We add Y register to that address
		// and this is where data is stored. We need to check if page is crossed after
		// adding Y register.
		CalculateAddress: func(pc uint16, x, y uint8, read ReadFunction) (uint16, bool) {
			zeroLow := read(pc + 1)

			low := read(uint16(zeroLow))
			high := read(uint16(zeroLow + 1))
			high16 := uint16(high) << 8

			var addr = (high16 | uint16(low)) + uint16(y)

			// Check if after adding Y register we crossed a page
			pageCrossed := (addr & 0xFF00) != high16

			return addr, pageCrossed
		},
	},
}
