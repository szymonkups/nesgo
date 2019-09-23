package core

// Instruction set
// ADC AND ASL BCC BCS BEQ BIT BMI BNE BPL BRK BVC BVS CLC
// CLD CLI CLV CMP CPX CPY DEC DEX DEY EOR INC INX INY JMP
// JSR LDA LDX LDY LSR NOP ORA PHA PHP PLA PLP ROL ROR RTI
// RTS SBC SEC SED SEI STA STX STY TAX TAY TSX TXA TXS TYA
var instructions = []*instruction{
	&adc, &and, &asl, &bcc, &bcs, &beq, &bit, &bmi, &bne, &bpl, &brk, &bvc, &bvs, &clc,
	&cld, &cli, &cmp, &cpx, &cpy, &clv, &inx, &iny,
}

// Instruction - describes instruction
type instruction struct {
	// Human readable name - for debugging
	name string

	// Op codes op code per addressing mode
	opCodes opCodesMap

	// Instruction handler
	handler instructionHandler
}

type opCodesMap map[uint8]struct {
	addrMode int
	cycles   uint8
}

// Actual instruction code. It should return true if there is a potential to
// add additional clock cycle - this will be checked together with addressing
// mode result. If both return true it will mean that one cycle needs to be
// added to instruction cycles.
// We pass CPU instance, absolute address calculated by correct addresing mode,
// actual op code (as same instruction can have different op codes depending on
// addressing mode) and addresing mode itself.
type instructionHandler func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool

// *****************************************************************************
// Instructions
// *****************************************************************************

// ADC - Add with carry
var adc = instruction{
	name: "ADC",
	opCodes: opCodesMap{
		0x69: {immediateAddressing, 2},
		0x65: {zeroPageAddressing, 3},
		0x75: {zeroPageXAddressing, 4},
		0x6D: {absoluteAddressing, 4},
		0x7D: {absoluteXAddressing, 4},
		0x79: {absoluteYAddressing, 4},
		0x61: {indirectXAddressing, 6},
		0x71: {indirectYAddressing, 5},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		var addCarry uint16 = 0
		data := cpu.bus.read(addr)

		if cpu.getFlag(cFlag) {
			addCarry = 1
		}

		acc16 := uint16(cpu.a)
		data16 := uint16(data)
		var res uint16 = acc16 + data16 + addCarry

		cpu.setFlag(cFlag, res > 0xFF)
		cpu.setFlag(zFLag, (res&0x00FF) == 0)
		cpu.setFlag(nFLag, res&0x80 != 0)

		// http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
		overflow := (^(acc16 ^ data16) & (acc16 ^ res)) & 0x0080
		cpu.setFlag(vFLag, overflow != 0)

		cpu.a = uint8(res & 0x00FF)

		return true
	},
}

// ASL - Arithmetic shift left
var asl = instruction{
	name: "ASL",
	opCodes: opCodesMap{
		0x0A: {accumulatorAddressing, 2},
		0x06: {zeroPageAddressing, 5},
		0x16: {zeroPageXAddressing, 6},
		0x0E: {absoluteAddressing, 6},
		0x1E: {absoluteXAddressing, 7},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.read(addr)
		var res16 uint16 = uint16(data) << 1
		cpu.setFlag(cFlag, (res16&0xFF00) > 0)
		cpu.setFlag(zFLag, (res16&0x00FF) == 0x00)
		cpu.setFlag(nFLag, res16&0x80 != 0)

		res := uint8(res16 & 0x00FF)

		if addrMode == accumulatorAddressing {
			cpu.a = res
		} else {
			cpu.bus.write(addr, res)
		}

		return false
	},
}

// AND - Bitwise logic AND
var and = instruction{
	name: "AND",
	opCodes: opCodesMap{
		0x29: {immediateAddressing, 2},
		0x25: {zeroPageAddressing, 3},
		0x35: {zeroPageXAddressing, 4},
		0x2D: {absoluteAddressing, 4},
		0x3D: {absoluteXAddressing, 4},
		0x39: {absoluteYAddressing, 4},
		0x21: {indirectXAddressing, 6},
		0x31: {indirectYAddressing, 4},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.read(addr)
		cpu.a = cpu.a & data

		cpu.setFlag(zFLag, cpu.a == 0x00)
		cpu.setFlag(nFLag, cpu.a&0b10000000 != 0)

		return true
	},
}

// BCC - Branch if carry is clear
var bcc = instruction{
	name: "BCC",
	opCodes: opCodesMap{
		0x90: {relativeAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		if !cpu.getFlag(cFlag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BCS - Branch if carry is set
var bcs = instruction{
	name: "BCS",
	opCodes: opCodesMap{
		0xB0: {relativeAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		if cpu.getFlag(cFlag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BEQ - Branch if equal - zero flag is set
var beq = instruction{
	name: "BEQ",
	opCodes: opCodesMap{
		0xF0: {relativeAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		if cpu.getFlag(zFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BMI - Branch if minus - negative flag is set
var bmi = instruction{
	name: "BMI",
	opCodes: opCodesMap{
		0x30: {relativeAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		if cpu.getFlag(nFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BNE - Branch if no equal - zero flag is not set
var bne = instruction{
	name: "BNE",
	opCodes: opCodesMap{
		0xD0: {relativeAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		if !cpu.getFlag(zFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BPL - Branch if positive - negative flag is not set
var bpl = instruction{
	name: "BPL",
	opCodes: opCodesMap{
		0x10: {relativeAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		if !cpu.getFlag(nFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BVC - Branch if overflow clear
var bvc = instruction{
	name: "BVC",
	opCodes: opCodesMap{
		0x50: {relativeAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		if !cpu.getFlag(vFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BVS - Branch if overflow is set
var bvs = instruction{
	name: "BVS",
	opCodes: opCodesMap{
		0x70: {relativeAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		if cpu.getFlag(vFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

func branchHandler(cpu *CPU, addr uint16) {
	cpu.cyclesLeft++

	// Page might be crossed - add one cycle if that happens
	if (addr & 0xFF00) != (cpu.pc & 0xFF00) {
		cpu.cyclesLeft++
	}

	cpu.pc = addr
}

// BIT - bit test
var bit = instruction{
	name: "BIT",
	opCodes: opCodesMap{
		0x24: {zeroPageAddressing, 3},
		0x2C: {absoluteAddressing, 4},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.read(addr)
		tmp := data & cpu.a

		cpu.setFlag(zFLag, tmp == 0x00)
		cpu.setFlag(nFLag, data&0b010000000 != 0)
		cpu.setFlag(vFLag, data&0b001000000 != 0)

		return false
	},
}

// BRK - Force interrupt
var brk = instruction{
	name: "BRK",
	opCodes: opCodesMap{
		0x00: {impliedAddressing, 7},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		// Padding byte
		// http://nesdev.com/the%20%27B%27%20flag%20&%20BRK%20instruction.txt
		cpu.pc++

		// Set interrupt flag
		cpu.setFlag(iFLag, true)

		// Push Program Counter to tack
		cpu.pushToStack(uint8((cpu.pc >> 8) & 0x00FF))
		cpu.pushToStack(uint8(cpu.pc & 0x00FF))

		// Set Break flag
		cpu.setFlag(bFLag, true)

		// Push processor status flags to stack
		cpu.pushToStack(cpu.p)

		// Clear Break flag - it is only set to be stored on stack
		// TODO: just push to stack P register with Break flag set maybe?
		cpu.setFlag(bFLag, false)

		// Read data from 0xFFFE and 0xFFFF and set PC
		low := uint16(cpu.bus.read(0xFFFE))
		high := uint16(cpu.bus.read(0xFFFF))

		cpu.pc = (high << 8) | low

		return false
	},
}

// INX - Increment X register by one
var inx = instruction{
	name: "INX",
	opCodes: opCodesMap{
		0xE8: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.x++

		cpu.setFlag(zFLag, cpu.x == 0x00)
		cpu.setFlag(nFLag, cpu.x&0b10000000 != 0)

		return false
	},
}

// INY - Increment Y register by one
var iny = instruction{
	name: "INY",
	opCodes: opCodesMap{
		0xC8: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.y++

		cpu.setFlag(zFLag, cpu.y == 0x00)
		cpu.setFlag(nFLag, cpu.y&0b10000000 != 0)

		return false
	},
}

// CLC - Clear carry flag
var clc = instruction{
	name: "CLC",
	opCodes: opCodesMap{
		0x18: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(cFlag, false)

		return false
	},
}

// CLD - Clear decimal flag
var cld = instruction{
	name: "CLD",
	opCodes: opCodesMap{
		0xD8: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(dFLag, false)

		return false
	},
}

// CLI - Clear interrupt flag (disable interrupts)
var cli = instruction{
	name: "CLI",
	opCodes: opCodesMap{
		0x58: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(iFLag, false)

		return false
	},
}

// CLV - Clear overflow flag
var clv = instruction{
	name: "CLV",
	opCodes: opCodesMap{
		0xB8: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(vFLag, false)

		return false
	},
}

// CMP - Compare accumulator
var cmp = instruction{
	name: "CMP",
	opCodes: opCodesMap{
		0xC9: {immediateAddressing, 2},
		0xC5: {zeroPageAddressing, 3},
		0xD5: {zeroPageXAddressing, 4},
		0xCD: {absoluteAddressing, 4},
		0xDD: {absoluteXAddressing, 4},
		0xD9: {absoluteYAddressing, 4},
		0xC1: {indirectXAddressing, 6},
		0xD1: {indirectYAddressing, 5},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		compareHandler(cpu, cpu.a, addr)

		return true
	},
}

// CPX - Compare X register
var cpx = instruction{
	name: "CPX",
	opCodes: opCodesMap{
		0xE0: {immediateAddressing, 2},
		0xE4: {zeroPageAddressing, 3},
		0xEC: {absoluteAddressing, 4},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		compareHandler(cpu, cpu.x, addr)

		return false
	},
}

// CPY - Compare Y register
var cpy = instruction{
	name: "CPY",
	opCodes: opCodesMap{
		0xC0: {immediateAddressing, 2},
		0xC4: {zeroPageAddressing, 3},
		0xCC: {absoluteAddressing, 4},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		compareHandler(cpu, cpu.y, addr)

		return false
	},
}

func compareHandler(cpu *CPU, a byte, addr uint16) {
	data := cpu.bus.read(addr)
	diff := a - data

	cpu.setFlag(cFlag, a >= data)
	cpu.setFlag(zFLag, diff == 0)
	cpu.setFlag(nFLag, diff&0x80 != 0)
}
