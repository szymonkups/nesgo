package core

var instructions = []*instruction{
	&and, &inx, &iny,
}

// Instruction - describes instruction
type instruction struct {
	// Human readable name - for debugging
	name string

	// Op codes op code per adrresing mode
	opCodes opCodesMap

	// Istruction handler
	handler instructionHandler
}

type opCodesMap map[uint8]struct {
	addrMode addressingMode
	cycles   uint8
}

// Actual instruction code. It should return true if there is a potential to
// add additional clock cycle
type instructionHandler func(*CPU, uint8) bool

// *****************************************************************************
// Instructions
// *****************************************************************************

// AND - Bitwise logic AND
var and = instruction{
	name: "AND",
	opCodes: opCodesMap{
		0x29: {immediateAddr, 2},
		0x25: {zeroPageAddr, 3},
		0x35: {zeroPageXAddr, 4},
		0x2D: {absoluteAddr, 4},
		0x3D: {absoluteXAddr, 4},
		0x39: {absoluteYAddr, 4},
		0x21: {indirectXAddr, 6},
		0x31: {indirectYAddr, 4},
	},
	handler: func(cpu *CPU, data uint8) bool {
		cpu.a = cpu.a & data

		cpu.setFlag(z, cpu.a == 0x00)
		cpu.setFlag(n, cpu.a&0x80 != 0)

		return true
	},
}

// INX - Increment X register by one
var inx = instruction{
	name: "INX",
	opCodes: opCodesMap{
		0xE8: {impliedAddr, 2},
	},
	handler: func(cpu *CPU, _ uint8) bool {
		cpu.x++

		cpu.setFlag(z, cpu.x == 0x00)
		cpu.setFlag(n, cpu.x&0x80 != 0)

		return false
	},
}

// INY - Increment Y register by one
var iny = instruction{
	name: "INY",
	opCodes: opCodesMap{
		0xC8: {impliedAddr, 2},
	},
	handler: func(cpu *CPU, _ uint8) bool {
		cpu.y++

		cpu.setFlag(z, cpu.y == 0x00)
		cpu.setFlag(n, cpu.y&0x80 != 0)

		return false
	},
}
