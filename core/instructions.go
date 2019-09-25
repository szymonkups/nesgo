package core

// Instruction set
// ADC AND ASL BCC BCS BEQ BIT BMI BNE BPL BRK BVC BVS CLC
// CLD CLI CLV CMP CPX CPY DEC DEX DEY EOR INC INX INY JMP
// JSR LDA LDX LDY LSR NOP ORA PHA PHP PLA PLP ROL ROR RTI
// RTS SBC SEC SED SEI STA STX STY TAX TAY TSX TXA TXS TYA
var instructions = []*instruction{
	&adc, &and, &asl, &bcc, &bcs, &beq, &bit, &bmi, &bne, &bpl, &brk, &bvc, &bvs, &clc,
	&cld, &cli, &clv, &cmp, &cpx, &cpy, &dec, &dex, &dey, &eor, &inc, &inx, &iny, &jmp,
	&jsr, &lda, &ldx, &ldy, &lsr, &nop, &ora, &pha, &php, &pla, &plp, &rol, &ror, &rti,
	&rts, &sbc, &sec, &sed, &sei, &sta, &stx, &sty, &tax, &tay, &tsx, &txa, &txs, &tya,
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
// We pass CPU instance, absolute address calculated by correct addressing mode,
// actual op code (as same instruction can have different op codes depending on
// addressing mode) and addressing mode itself.
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
		acc := cpu.a
		data := cpu.bus.read(addr)
		carry := uint8(0)

		if cpu.getFlag(cFlag) {
			carry = 1
		}

		cpu.a = acc + data + carry

		cpu.setFlag(cFlag, int(acc)+int(data)+int(carry) > 0xFF)
		cpu.setFlag(vFLag, (acc^data)&0x80 == 0 && (acc^cpu.a)&0x80 != 0)
		cpu.setFlag(zFLag, cpu.a == 0)
		cpu.setFlag(nFLag, cpu.a&0x80 != 0)

		return true
	},
}

// SBC - Subtract with carry
var sbc = instruction{
	name: "SBC",
	opCodes: opCodesMap{
		0xE9: {immediateAddressing, 2},
		0xE5: {zeroPageAddressing, 3},
		0xF5: {zeroPageXAddressing, 4},
		0xED: {absoluteAddressing, 4},
		0xFD: {absoluteXAddressing, 4},
		0xF9: {absoluteYAddressing, 4},
		0xE1: {indirectXAddressing, 6},
		0xF1: {indirectYAddressing, 5},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		acc := cpu.a
		data := cpu.bus.read(addr)
		carry := uint8(0)

		if cpu.getFlag(cFlag) {
			carry = 1
		}

		cpu.a = acc - data - (1 - carry)

		cpu.setFlag(cFlag, int(acc)-int(data)-int(1-carry) >= 0)
		cpu.setFlag(vFLag, (acc^data)&0x80 != 0 && (acc^cpu.a)&0x80 != 0)
		cpu.setFlag(zFLag, cpu.a == 0)
		cpu.setFlag(nFLag, cpu.a&0x80 != 0)
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
		var res16 = uint16(data) << 1
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

		// Push Program Counter to stack
		cpu.pushToStack16(cpu.pc)

		// Push processor status flags to stack
		// https://wiki.nesdev.com/w/index.php/Status_flags - set 4 and 5 bit before pushing
		cpu.pushToStack(cpu.p | 0b00110000)

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

// INC - Increment memory location
var inc = instruction{
	name: "INC",
	opCodes: opCodesMap{
		0xE6: {zeroPageAddressing, 5},
		0xF6: {zeroPageXAddressing, 6},
		0xEE: {absoluteAddressing, 6},
		0xFE: {absoluteXAddressing, 7},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.read(addr)
		data++
		cpu.bus.write(addr, data)

		cpu.setFlag(zFLag, data == 0x00)
		cpu.setFlag(nFLag, data&0b10000000 != 0)

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

// DEC - decrement memory
var dec = instruction{
	name: "DEC",
	opCodes: opCodesMap{
		0xC6: {zeroPageAddressing, 5},
		0xD6: {zeroPageXAddressing, 6},
		0xCE: {absoluteAddressing, 6},
		0xDE: {absoluteXAddressing, 7},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.read(addr)
		data--
		cpu.bus.write(addr, data)

		cpu.setFlag(zFLag, data == 0)
		cpu.setFlag(nFLag, data&0x80 != 0)

		return false
	},
}

// DEX - decrement X register
var dex = instruction{
	name: "DEX",
	opCodes: opCodesMap{
		0xCA: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.x--

		cpu.setFlag(zFLag, cpu.x == 0)
		cpu.setFlag(nFLag, cpu.x&0x80 != 0)

		return false
	},
}

// DEY - decrement Y register
var dey = instruction{
	name: "DEY",
	opCodes: opCodesMap{
		0x88: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.y--

		cpu.setFlag(zFLag, cpu.y == 0)
		cpu.setFlag(nFLag, cpu.y&0x80 != 0)

		return false
	},
}

// EOR - Exclusive OR
var eor = instruction{
	name: "EOR",
	opCodes: opCodesMap{
		0x49: {immediateAddressing, 2},
		0x45: {zeroPageAddressing, 3},
		0x55: {zeroPageXAddressing, 4},
		0x4D: {absoluteAddressing, 4},
		0x5D: {absoluteXAddressing, 4},
		0x59: {absoluteYAddressing, 4},
		0x41: {indirectXAddressing, 6},
		0x51: {indirectYAddressing, 5},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.read(addr)
		cpu.a = cpu.a ^ data

		cpu.setFlag(zFLag, cpu.a == 0x00)
		cpu.setFlag(nFLag, cpu.a&0x80 != 0x00)

		return true
	},
}

// PHA - Push accumulator to stack
var pha = instruction{
	name: "PHA",
	opCodes: opCodesMap{
		0x48: {impliedAddressing, 3},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.pushToStack(cpu.a)
		return false
	},
}

// PHP - Push processor status to stack
var php = instruction{
	name: "PHP",
	opCodes: opCodesMap{
		0x08: {impliedAddressing, 3},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		// Push processor status flags to stack
		// https://wiki.nesdev.com/w/index.php/Status_flags - set 4 and 5 bit before pushing
		cpu.pushToStack(cpu.p | 0b00110000)
		return false
	},
}

// PLA = Pull accumulator from stack
var pla = instruction{
	name: "PLA",
	opCodes: opCodesMap{
		0x68: {impliedAddressing, 4},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.a = cpu.pullFromStack()
		cpu.setFlag(zFLag, cpu.a == 0x00)
		cpu.setFlag(nFLag, cpu.a&0x80 != 0)

		return false
	},
}

// PLP = Pull processor status from stack
var plp = instruction{
	name: "PLP",
	opCodes: opCodesMap{
		0x28: {impliedAddressing, 4},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		// https://wiki.nesdev.com/w/index.php/Status_flags - ignore 4 and 5 bit - make sure 5 is set in p register
		cpu.p = cpu.pullFromStack()&0b11001111 | 0b00100000

		return false
	},
}

// JMP = Jump to address
var jmp = instruction{
	// TODO: check from http://obelisk.me.uk/6502/reference.html#JMP
	// An original 6502 has does not correctly fetch the target address if the indirect vector falls on a page
	// boundary (e.g. $xxFF where xx is any value from $00 to $FF). In this case fetches the LSB from $xxFF as
	// expected but takes the MSB from $xx00. This is fixed in some later chips like the 65SC02 so for compatibility
	// always ensure the indirect vector is not at the end of the page.

	name: "JMP",
	opCodes: opCodesMap{
		0x4C: {absoluteAddressing, 3},
		0x6C: {indirectAddressing, 5},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.pc = addr
		return false
	},
}

// JSR - Jump to subroutine
var jsr = instruction{
	name: "JSR",
	opCodes: opCodesMap{
		0x20: {absoluteAddressing, 6},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.pc--
		cpu.pushToStack16(cpu.pc)
		cpu.pc = addr

		return false
	},
}

// LDA - Load accumulator
var lda = instruction{
	name: "LDA",
	opCodes: opCodesMap{
		0xA9: {immediateAddressing, 2},
		0xA5: {zeroPageAddressing, 3},
		0xB5: {zeroPageXAddressing, 4},
		0xAD: {absoluteAddressing, 4},
		0xBD: {absoluteXAddressing, 4},
		0xB9: {absoluteYAddressing, 4},
		0xA1: {indirectXAddressing, 6},
		0xB1: {indirectYAddressing, 5},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.a = cpu.bus.read(addr)
		cpu.setFlag(zFLag, cpu.a == 0x00)
		cpu.setFlag(nFLag, cpu.a&0x80 != 0x00)
		return true
	},
}

// LDX - Load X register
var ldx = instruction{
	name: "LDX",
	opCodes: opCodesMap{
		0xA2: {immediateAddressing, 2},
		0xA6: {zeroPageAddressing, 3},
		0xB6: {zeroPageYAddressing, 4},
		0xAE: {absoluteAddressing, 4},
		0xBE: {absoluteYAddressing, 4},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.x = cpu.bus.read(addr)
		cpu.setFlag(zFLag, cpu.x == 0x00)
		cpu.setFlag(nFLag, cpu.x&0x80 != 0x00)
		return true
	},
}

// LDY - Load Y register
var ldy = instruction{
	name: "LDY",
	opCodes: opCodesMap{
		0xA0: {immediateAddressing, 2},
		0xA4: {zeroPageAddressing, 3},
		0xB4: {zeroPageXAddressing, 4},
		0xAC: {absoluteAddressing, 4},
		0xBC: {absoluteXAddressing, 4},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.y = cpu.bus.read(addr)
		cpu.setFlag(zFLag, cpu.y == 0x00)
		cpu.setFlag(nFLag, cpu.y&0x80 != 0x00)
		return true
	},
}

// LSR - Logical shift right
var lsr = instruction{
	name: "LSR",
	opCodes: opCodesMap{
		0x4A: {accumulatorAddressing, 2},
		0x46: {zeroPageAddressing, 5},
		0x56: {zeroPageXAddressing, 6},
		0x4E: {absoluteAddressing, 6},
		0x5E: {absoluteXAddressing, 7},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		var data uint8

		// Get correct data
		if addrMode == accumulatorAddressing {
			data = cpu.a
		} else {
			data = cpu.bus.read(addr)
		}

		cpu.setFlag(cFlag, data&0x01 != 0x00)
		data = data >> 1
		cpu.setFlag(zFLag, data == 0x00)
		cpu.setFlag(nFLag, data&0x80 != 0x00)

		// Save result to correct place
		if addrMode == accumulatorAddressing {
			cpu.a = data
		} else {
			cpu.bus.write(addr, data)
		}

		return false
	},
}

// NOP - No operation
// TODO: there is a lot of unofficial op codes that does NOP, add them
var nop = instruction{
	name: "NOP",
	opCodes: opCodesMap{
		0xEA: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		return false
	},
}

// ORA - Logical inclusive OR
var ora = instruction{
	name: "ORA",
	opCodes: opCodesMap{
		0x09: {immediateAddressing, 2},
		0x05: {zeroPageAddressing, 3},
		0x15: {zeroPageXAddressing, 4},
		0x0D: {absoluteAddressing, 4},
		0x1D: {absoluteXAddressing, 4},
		0x19: {absoluteYAddressing, 4},
		0x01: {indirectXAddressing, 6},
		0x11: {indirectYAddressing, 5},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.a |= cpu.bus.read(addr)
		cpu.setFlag(zFLag, cpu.a == 0x00)
		cpu.setFlag(nFLag, cpu.a&0x80 != 0x00)

		return true
	},
}

// ROL - Rotate left
var rol = instruction{
	name: "ROL",
	opCodes: opCodesMap{
		0x2A: {accumulatorAddressing, 2},
		0x26: {zeroPageAddressing, 5},
		0x36: {zeroPageXAddressing, 6},
		0x2E: {absoluteAddressing, 6},
		0x3E: {absoluteXAddressing, 7},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		var data uint8

		if addrMode == accumulatorAddressing {
			data = cpu.a
		} else {
			data = cpu.bus.read(addr)
		}

		tmpC := cpu.getFlag(cFlag)
		cpu.setFlag(cFlag, data&0b10000000 != 0)
		data = data << 1

		if tmpC {
			data |= 0b00000001
		}

		cpu.setFlag(zFLag, data == 0x00)
		cpu.setFlag(nFLag, data&0x80 != 0x00)

		if addrMode == accumulatorAddressing {
			cpu.a = data
		} else {
			cpu.bus.write(addr, data)
		}

		return false
	},
}

// ROR - Rotate right
var ror = instruction{
	name: "ROR",
	opCodes: opCodesMap{
		0x6A: {accumulatorAddressing, 2},
		0x66: {zeroPageAddressing, 5},
		0x76: {zeroPageXAddressing, 6},
		0x6E: {absoluteAddressing, 6},
		0x7E: {absoluteXAddressing, 7},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		var data uint8

		if addrMode == accumulatorAddressing {
			data = cpu.a
		} else {
			data = cpu.bus.read(addr)
		}

		tmpC := cpu.getFlag(cFlag)
		cpu.setFlag(cFlag, data&0b00000001 != 0)
		data = data >> 1

		if tmpC {
			data |= 0b10000000
		}

		cpu.setFlag(zFLag, data == 0x00)
		cpu.setFlag(nFLag, data&0x80 != 0x00)

		if addrMode == accumulatorAddressing {
			cpu.a = data
		} else {
			cpu.bus.write(addr, data)
		}

		return false
	},
}

// RTI - Return from interrupt
var rti = instruction{
	name: "RTI",
	opCodes: opCodesMap{
		0x40: {impliedAddressing, 6},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		// https://wiki.nesdev.com/w/index.php/Status_flags - ignore 4 and 5 bit - make sure 5 is set in p register
		cpu.p = cpu.pullFromStack()&0b11001111 | 0b00100000
		cpu.pc = cpu.pullFromStack16()

		return false
	},
}

// RTS - Return from subroutine
var rts = instruction{
	name: "RTS",
	opCodes: opCodesMap{
		0x60: {impliedAddressing, 6},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.pc = cpu.pullFromStack16() + 1

		return false
	},
}

// SEC - Set carry flag
var sec = instruction{
	name: "SEC",
	opCodes: opCodesMap{
		0x38: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(cFlag, true)

		return false
	},
}

// SED - Set decimal flag
var sed = instruction{
	name: "SED",
	opCodes: opCodesMap{
		0xF8: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(dFLag, true)

		return false
	},
}

// SEI - Set interrupt disable flag
var sei = instruction{
	name: "SEI",
	opCodes: opCodesMap{
		0x78: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(iFLag, true)

		return false
	},
}

// STA - Store accumulator
var sta = instruction{
	name: "STA",
	opCodes: opCodesMap{
		0x85: {zeroPageAddressing, 3},
		0x95: {zeroPageXAddressing, 4},
		0x8D: {absoluteAddressing, 4},
		0x9D: {absoluteXAddressing, 5},
		0x99: {absoluteYAddressing, 5},
		0x81: {indirectXAddressing, 6},
		0x91: {indirectYAddressing, 6},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.bus.write(addr, cpu.a)

		return false
	},
}

// STX - Store X register
var stx = instruction{
	name: "STX",
	opCodes: opCodesMap{
		0x86: {zeroPageAddressing, 3},
		0x96: {zeroPageYAddressing, 4},
		0x8E: {absoluteAddressing, 4},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.bus.write(addr, cpu.x)

		return false
	},
}

// STY - Store Y register
var sty = instruction{
	name: "STY",
	opCodes: opCodesMap{
		0x84: {zeroPageAddressing, 3},
		0x94: {zeroPageXAddressing, 4},
		0x8C: {absoluteAddressing, 4},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.bus.write(addr, cpu.y)

		return false
	},
}

// TAX - Transfer accumulator to X
var tax = instruction{
	name: "TAX",
	opCodes: opCodesMap{
		0xAA: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.x = cpu.a
		cpu.setFlag(zFLag, cpu.x == 0x00)
		cpu.setFlag(nFLag, cpu.x&0x80 != 0x00)

		return false
	},
}

// TAY - Transfer accumulator to Y
var tay = instruction{
	name: "TAY",
	opCodes: opCodesMap{
		0xA8: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.y = cpu.a
		cpu.setFlag(zFLag, cpu.y == 0x00)
		cpu.setFlag(nFLag, cpu.y&0x80 != 0x00)

		return false
	},
}

// TSX - Transfer stack pointer to X
var tsx = instruction{
	name: "TSX",
	opCodes: opCodesMap{
		0xBA: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.x = cpu.sp
		cpu.setFlag(zFLag, cpu.x == 0x00)
		cpu.setFlag(nFLag, cpu.x&0x80 != 0x00)

		return false
	},
}

// TXA - Transfer X to accumulator
var txa = instruction{
	name: "TXA",
	opCodes: opCodesMap{
		0x8A: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.a = cpu.x
		cpu.setFlag(zFLag, cpu.a == 0x00)
		cpu.setFlag(nFLag, cpu.a&0x80 != 0x00)

		return false
	},
}

// TXS - Transfer X to stack pointer
var txs = instruction{
	name: "TXS",
	opCodes: opCodesMap{
		0x9A: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.sp = cpu.x

		return false
	},
}

// TYA - Transfer Y to accumulator
var tya = instruction{
	name: "TYA",
	opCodes: opCodesMap{
		0x98: {impliedAddressing, 2},
	},
	handler: func(cpu *CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.a = cpu.y
		cpu.setFlag(zFLag, cpu.a == 0x00)
		cpu.setFlag(nFLag, cpu.a&0x80 != 0x00)

		return false
	},
}
