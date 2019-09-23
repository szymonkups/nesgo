package core

// Instruction set
// ADC AND ASL BCC BCS BEQ BIT BMI BNE BPL BRK BVC BVS CLC
// CLD CLI CLV CMP CPX CPY DEC DEX DEY EOR INC INX INY JMP
// JSR LDA LDX LDY LSR NOP ORA PHA PHP PLA PLP ROL ROR RTI
// RTS SBC SEC SED SEI STA STX STY TAX TAY TSX TXA TXS TYA
var instructions = []*instruction{
	&and, &bcc, &bcs, &beq, &bmi, &bne, &bpl, &bvc, &bvs, &clc,
	&cld, &cli, &clv, &inx, &iny,
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
	addrMode addressingMode
	cycles   uint8
}

// Actual instruction code. It should return true if there is a potential to
// add additional clock cycle - this will be checked together with addressing
// mode result. If both return true it will mean that one cycle needs to be
// added to instruction cycles.
// We pass CPU instance, data fetched by correct addresing mode, actual op code
// (as same instruction can have different op codes depending on addressing mode)
// and addresing mode itself.
type instructionHandler func(cpu *CPU, data uint8, opCode uint8, addrMode addressingMode) bool

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
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
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
		0x90: {relativeAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		if !cpu.getFlag(cFlag) {
			branchHandler(cpu, data)
		}

		return false
	},
}

// BCS - Branch if carry is set
var bcs = instruction{
	name: "BCS",
	opCodes: opCodesMap{
		0xB0: {relativeAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		if cpu.getFlag(cFlag) {
			branchHandler(cpu, data)
		}

		return false
	},
}

// BEQ - Branch if equal - zero flag is set
var beq = instruction{
	name: "BEQ",
	opCodes: opCodesMap{
		0xF0: {relativeAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		if cpu.getFlag(zFLag) {
			branchHandler(cpu, data)
		}

		return false
	},
}

// BMI - Branch if minus - negative flag is set
var bmi = instruction{
	name: "BMI",
	opCodes: opCodesMap{
		0x30: {relativeAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		if cpu.getFlag(nFLag) {
			branchHandler(cpu, data)
		}

		return false
	},
}

// BNE - Branch if no equal - zero flag is not set
var bne = instruction{
	name: "BNE",
	opCodes: opCodesMap{
		0xD0: {relativeAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		if !cpu.getFlag(zFLag) {
			branchHandler(cpu, data)
		}

		return false
	},
}

// BPL - Branch if positive - negative flag is not set
var bpl = instruction{
	name: "BPL",
	opCodes: opCodesMap{
		0x10: {relativeAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		if !cpu.getFlag(nFLag) {
			branchHandler(cpu, data)
		}

		return false
	},
}

// BVC - Branch if overflow clear
var bvc = instruction{
	name: "BVC",
	opCodes: opCodesMap{
		0x50: {relativeAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		if !cpu.getFlag(vFLag) {
			branchHandler(cpu, data)
		}

		return false
	},
}

// BVS - Branch if overflow is set
var bvs = instruction{
	name: "BVS",
	opCodes: opCodesMap{
		0x70: {relativeAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		if cpu.getFlag(vFLag) {
			branchHandler(cpu, data)
		}

		return false
	},
}

func branchHandler(cpu *CPU, data uint8) {
	cpu.cyclesLeft++

	abs := uint16(toAbs(data))
	tmp := cpu.pc
	if isNegative(data) {
		tmp -= abs
	} else {
		tmp += abs
	}

	// Page might be crossed - add one cycle if that happens
	if (tmp & 0xFF00) != (cpu.pc & 0xFF00) {
		cpu.cyclesLeft++
	}

	cpu.pc = tmp
}

// INX - Increment X register by one
var inx = instruction{
	name: "INX",
	opCodes: opCodesMap{
		0xE8: {impliedAddr, 2},
	},
	handler: func(cpu *CPU, _ uint8, op uint8, addrMod addressingMode) bool {
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
		0xC8: {impliedAddr, 2},
	},
	handler: func(cpu *CPU, _ uint8, op uint8, addrMod addressingMode) bool {
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
		0x18: {impliedAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		cpu.setFlag(cFlag, false)

		return false
	},
}

// CLD - Clear decimal flag
var cld = instruction{
	name: "CLD",
	opCodes: opCodesMap{
		0xD8: {impliedAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		cpu.setFlag(dFLag, false)

		return false
	},
}

// CLI - Clear interrupt flag (disable interrupts)
var cli = instruction{
	name: "CLI",
	opCodes: opCodesMap{
		0x58: {impliedAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		cpu.setFlag(iFLag, false)

		return false
	},
}

// CLV - Clear overflow flag
var clv = instruction{
	name: "CLV",
	opCodes: opCodesMap{
		0xB8: {impliedAddr, 2},
	},
	handler: func(cpu *CPU, data uint8, op uint8, addrMod addressingMode) bool {
		cpu.setFlag(vFLag, false)

		return false
	},
}
