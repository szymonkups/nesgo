package instructions

import (
	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/core/addressing"
)

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
	Name string

	// Op codes op code per addressing mode
	OpCodes opCodesMap

	// Instruction handler
	Handler instructionHandler
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
type instructionHandler func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool

// *****************************************************************************
// Instructions
// *****************************************************************************

// ADC - Add with carry
var adc = instruction{
	Name: "ADC",
	OpCodes: opCodesMap{
		0x69: {addressing.ImmediateAddressing, 2},
		0x65: {addressing.ZeroPageAddressing, 3},
		0x75: {addressing.ZeroPageXAddressing, 4},
		0x6D: {addressing.AbsoluteAddressing, 4},
		0x7D: {addressing.AbsoluteXAddressing, 4},
		0x79: {addressing.AbsoluteYAddressing, 4},
		0x61: {addressing.IndirectXAddressing, 6},
		0x71: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		acc := cpu.a
		data := cpu.bus.Read(addr)
		carry := uint8(0)

		if cpu.getFlag(core.cFlag) {
			carry = 1
		}

		cpu.a = acc + data + carry

		cpu.setFlag(core.cFlag, int(acc)+int(data)+int(carry) > 0xFF)
		cpu.setFlag(core.vFLag, (acc^data)&0x80 == 0 && (acc^cpu.a)&0x80 != 0)
		cpu.setFlag(core.zFLag, cpu.a == 0)
		cpu.setFlag(core.nFLag, cpu.a&0x80 != 0)

		return true
	},
}

// SBC - Subtract with carry
var sbc = instruction{
	Name: "SBC",
	OpCodes: opCodesMap{
		0xE9: {addressing.ImmediateAddressing, 2},
		0xE5: {addressing.ZeroPageAddressing, 3},
		0xF5: {addressing.ZeroPageXAddressing, 4},
		0xED: {addressing.AbsoluteAddressing, 4},
		0xFD: {addressing.AbsoluteXAddressing, 4},
		0xF9: {addressing.AbsoluteYAddressing, 4},
		0xE1: {addressing.IndirectXAddressing, 6},
		0xF1: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		acc := cpu.a
		data := cpu.bus.Read(addr)
		carry := uint8(0)

		if cpu.getFlag(core.cFlag) {
			carry = 1
		}

		cpu.a = acc - data - (1 - carry)

		cpu.setFlag(core.cFlag, int(acc)-int(data)-int(1-carry) >= 0)
		cpu.setFlag(core.vFLag, (acc^data)&0x80 != 0 && (acc^cpu.a)&0x80 != 0)
		cpu.setFlag(core.zFLag, cpu.a == 0)
		cpu.setFlag(core.nFLag, cpu.a&0x80 != 0)
		return true
	},
}

// ASL - Arithmetic shift left
var asl = instruction{
	Name: "ASL",
	OpCodes: opCodesMap{
		0x0A: {addressing.AccumulatorAddressing, 2},
		0x06: {addressing.ZeroPageAddressing, 5},
		0x16: {addressing.ZeroPageXAddressing, 6},
		0x0E: {addressing.AbsoluteAddressing, 6},
		0x1E: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.Read(addr)
		var res16 = uint16(data) << 1
		cpu.setFlag(core.cFlag, (res16&0xFF00) > 0)
		cpu.setFlag(core.zFLag, (res16&0x00FF) == 0x00)
		cpu.setFlag(core.nFLag, res16&0x80 != 0)

		res := uint8(res16 & 0x00FF)

		if addrMode == addressing.AccumulatorAddressing {
			cpu.a = res
		} else {
			cpu.bus.Write(addr, res)
		}

		return false
	},
}

// AND - Bitwise logic AND
var and = instruction{
	Name: "AND",
	OpCodes: opCodesMap{
		0x29: {addressing.ImmediateAddressing, 2},
		0x25: {addressing.ZeroPageAddressing, 3},
		0x35: {addressing.ZeroPageXAddressing, 4},
		0x2D: {addressing.AbsoluteAddressing, 4},
		0x3D: {addressing.AbsoluteXAddressing, 4},
		0x39: {addressing.AbsoluteYAddressing, 4},
		0x21: {addressing.IndirectXAddressing, 6},
		0x31: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.Read(addr)
		cpu.a = cpu.a & data

		cpu.setFlag(core.zFLag, cpu.a == 0x00)
		cpu.setFlag(core.nFLag, cpu.a&0b10000000 != 0)

		return true
	},
}

// BCC - Branch if carry is clear
var bcc = instruction{
	Name: "BCC",
	OpCodes: opCodesMap{
		0x90: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		if !cpu.getFlag(core.cFlag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BCS - Branch if carry is set
var bcs = instruction{
	Name: "BCS",
	OpCodes: opCodesMap{
		0xB0: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		if cpu.getFlag(core.cFlag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BEQ - Branch if equal - zero flag is set
var beq = instruction{
	Name: "BEQ",
	OpCodes: opCodesMap{
		0xF0: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		if cpu.getFlag(core.zFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BMI - Branch if minus - negative flag is set
var bmi = instruction{
	Name: "BMI",
	OpCodes: opCodesMap{
		0x30: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		if cpu.getFlag(core.nFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BNE - Branch if no equal - zero flag is not set
var bne = instruction{
	Name: "BNE",
	OpCodes: opCodesMap{
		0xD0: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		if !cpu.getFlag(core.zFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BPL - Branch if positive - negative flag is not set
var bpl = instruction{
	Name: "BPL",
	OpCodes: opCodesMap{
		0x10: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		if !cpu.getFlag(core.nFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BVC - Branch if overflow clear
var bvc = instruction{
	Name: "BVC",
	OpCodes: opCodesMap{
		0x50: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		if !cpu.getFlag(core.vFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BVS - Branch if overflow is set
var bvs = instruction{
	Name: "BVS",
	OpCodes: opCodesMap{
		0x70: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		if cpu.getFlag(core.vFLag) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

func branchHandler(cpu *core.CPU, addr uint16) {
	cpu.cyclesLeft++

	// Page might be crossed - add one cycle if that happens
	if (addr & 0xFF00) != (cpu.pc & 0xFF00) {
		cpu.cyclesLeft++
	}

	cpu.pc = addr
}

// BIT - bit test
var bit = instruction{
	Name: "BIT",
	OpCodes: opCodesMap{
		0x24: {addressing.ZeroPageAddressing, 3},
		0x2C: {addressing.AbsoluteAddressing, 4},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.Read(addr)
		tmp := data & cpu.a

		cpu.setFlag(core.zFLag, tmp == 0x00)
		cpu.setFlag(core.nFLag, data&0b010000000 != 0)
		cpu.setFlag(core.vFLag, data&0b001000000 != 0)

		return false
	},
}

// BRK - Force interrupt
var brk = instruction{
	Name: "BRK",
	OpCodes: opCodesMap{
		0x00: {addressing.ImpliedAddressing, 7},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		// Padding byte
		// http://nesdev.com/the%20%27B%27%20flag%20&%20BRK%20instruction.txt
		cpu.pc++

		// Set interrupt flag
		cpu.setFlag(core.iFLag, true)

		// Push Program Counter to stack
		cpu.pushToStack16(cpu.pc)

		// Push processor status flags to stack
		// https://wiki.nesdev.com/w/index.php/Status_flags - set 4 and 5 bit before pushing
		cpu.pushToStack(cpu.p | 0b00110000)

		// Read data from 0xFFFE and 0xFFFF and set PC
		cpu.pc = cpu.bus.Read16(0xFFFE)

		return false
	},
}

// INX - Increment X register by one
var inx = instruction{
	Name: "INX",
	OpCodes: opCodesMap{
		0xE8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.x++

		cpu.setFlag(core.zFLag, cpu.x == 0x00)
		cpu.setFlag(core.nFLag, cpu.x&0b10000000 != 0)

		return false
	},
}

// INY - Increment Y register by one
var iny = instruction{
	Name: "INY",
	OpCodes: opCodesMap{
		0xC8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.y++

		cpu.setFlag(core.zFLag, cpu.y == 0x00)
		cpu.setFlag(core.nFLag, cpu.y&0b10000000 != 0)

		return false
	},
}

// INC - Increment memory location
var inc = instruction{
	Name: "INC",
	OpCodes: opCodesMap{
		0xE6: {addressing.ZeroPageAddressing, 5},
		0xF6: {addressing.ZeroPageXAddressing, 6},
		0xEE: {addressing.AbsoluteAddressing, 6},
		0xFE: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.Read(addr)
		data++
		cpu.bus.Write(addr, data)

		cpu.setFlag(core.zFLag, data == 0x00)
		cpu.setFlag(core.nFLag, data&0b10000000 != 0)

		return false
	},
}

// CLC - Clear carry flag
var clc = instruction{
	Name: "CLC",
	OpCodes: opCodesMap{
		0x18: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(core.cFlag, false)

		return false
	},
}

// CLD - Clear decimal flag
var cld = instruction{
	Name: "CLD",
	OpCodes: opCodesMap{
		0xD8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(core.dFLag, false)

		return false
	},
}

// CLI - Clear interrupt flag (disable interrupts)
var cli = instruction{
	Name: "CLI",
	OpCodes: opCodesMap{
		0x58: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(core.iFLag, false)

		return false
	},
}

// CLV - Clear overflow flag
var clv = instruction{
	Name: "CLV",
	OpCodes: opCodesMap{
		0xB8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(core.vFLag, false)

		return false
	},
}

// CMP - Compare accumulator
var cmp = instruction{
	Name: "CMP",
	OpCodes: opCodesMap{
		0xC9: {addressing.ImmediateAddressing, 2},
		0xC5: {addressing.ZeroPageAddressing, 3},
		0xD5: {addressing.ZeroPageXAddressing, 4},
		0xCD: {addressing.AbsoluteAddressing, 4},
		0xDD: {addressing.AbsoluteXAddressing, 4},
		0xD9: {addressing.AbsoluteYAddressing, 4},
		0xC1: {addressing.IndirectXAddressing, 6},
		0xD1: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		compareHandler(cpu, cpu.a, addr)

		return true
	},
}

// CPX - Compare X register
var cpx = instruction{
	Name: "CPX",
	OpCodes: opCodesMap{
		0xE0: {addressing.ImmediateAddressing, 2},
		0xE4: {addressing.ZeroPageAddressing, 3},
		0xEC: {addressing.AbsoluteAddressing, 4},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		compareHandler(cpu, cpu.x, addr)

		return false
	},
}

// CPY - Compare Y register
var cpy = instruction{
	Name: "CPY",
	OpCodes: opCodesMap{
		0xC0: {addressing.ImmediateAddressing, 2},
		0xC4: {addressing.ZeroPageAddressing, 3},
		0xCC: {addressing.AbsoluteAddressing, 4},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		compareHandler(cpu, cpu.y, addr)

		return false
	},
}

func compareHandler(cpu *core.CPU, a byte, addr uint16) {
	data := cpu.bus.Read(addr)
	diff := a - data

	cpu.setFlag(core.cFlag, a >= data)
	cpu.setFlag(core.zFLag, diff == 0)
	cpu.setFlag(core.nFLag, diff&0x80 != 0)
}

// DEC - decrement memory
var dec = instruction{
	Name: "DEC",
	OpCodes: opCodesMap{
		0xC6: {addressing.ZeroPageAddressing, 5},
		0xD6: {addressing.ZeroPageXAddressing, 6},
		0xCE: {addressing.AbsoluteAddressing, 6},
		0xDE: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.Read(addr)
		data--
		cpu.bus.Write(addr, data)

		cpu.setFlag(core.zFLag, data == 0)
		cpu.setFlag(core.nFLag, data&0x80 != 0)

		return false
	},
}

// DEX - decrement X register
var dex = instruction{
	Name: "DEX",
	OpCodes: opCodesMap{
		0xCA: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.x--

		cpu.setFlag(core.zFLag, cpu.x == 0)
		cpu.setFlag(core.nFLag, cpu.x&0x80 != 0)

		return false
	},
}

// DEY - decrement Y register
var dey = instruction{
	Name: "DEY",
	OpCodes: opCodesMap{
		0x88: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.y--

		cpu.setFlag(core.zFLag, cpu.y == 0)
		cpu.setFlag(core.nFLag, cpu.y&0x80 != 0)

		return false
	},
}

// EOR - Exclusive OR
var eor = instruction{
	Name: "EOR",
	OpCodes: opCodesMap{
		0x49: {addressing.ImmediateAddressing, 2},
		0x45: {addressing.ZeroPageAddressing, 3},
		0x55: {addressing.ZeroPageXAddressing, 4},
		0x4D: {addressing.AbsoluteAddressing, 4},
		0x5D: {addressing.AbsoluteXAddressing, 4},
		0x59: {addressing.AbsoluteYAddressing, 4},
		0x41: {addressing.IndirectXAddressing, 6},
		0x51: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		data := cpu.bus.Read(addr)
		cpu.a = cpu.a ^ data

		cpu.setFlag(core.zFLag, cpu.a == 0x00)
		cpu.setFlag(core.nFLag, cpu.a&0x80 != 0x00)

		return true
	},
}

// PHA - Push accumulator to stack
var pha = instruction{
	Name: "PHA",
	OpCodes: opCodesMap{
		0x48: {addressing.ImpliedAddressing, 3},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.pushToStack(cpu.a)
		return false
	},
}

// PHP - Push processor status to stack
var php = instruction{
	Name: "PHP",
	OpCodes: opCodesMap{
		0x08: {addressing.ImpliedAddressing, 3},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		// Push processor status flags to stack
		// https://wiki.nesdev.com/w/index.php/Status_flags - set 4 and 5 bit before pushing
		cpu.pushToStack(cpu.p | 0b00110000)
		return false
	},
}

// PLA = Pull accumulator from stack
var pla = instruction{
	Name: "PLA",
	OpCodes: opCodesMap{
		0x68: {addressing.ImpliedAddressing, 4},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.a = cpu.pullFromStack()
		cpu.setFlag(core.zFLag, cpu.a == 0x00)
		cpu.setFlag(core.nFLag, cpu.a&0x80 != 0)

		return false
	},
}

// PLP = Pull processor status from stack
var plp = instruction{
	Name: "PLP",
	OpCodes: opCodesMap{
		0x28: {addressing.ImpliedAddressing, 4},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
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

	Name: "JMP",
	OpCodes: opCodesMap{
		0x4C: {addressing.AbsoluteAddressing, 3},
		0x6C: {addressing.IndirectAddressing, 5},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.pc = addr
		return false
	},
}

// JSR - Jump to subroutine
var jsr = instruction{
	Name: "JSR",
	OpCodes: opCodesMap{
		0x20: {addressing.AbsoluteAddressing, 6},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.pc--
		cpu.pushToStack16(cpu.pc)
		cpu.pc = addr

		return false
	},
}

// LDA - Load accumulator
var lda = instruction{
	Name: "LDA",
	OpCodes: opCodesMap{
		0xA9: {addressing.ImmediateAddressing, 2},
		0xA5: {addressing.ZeroPageAddressing, 3},
		0xB5: {addressing.ZeroPageXAddressing, 4},
		0xAD: {addressing.AbsoluteAddressing, 4},
		0xBD: {addressing.AbsoluteXAddressing, 4},
		0xB9: {addressing.AbsoluteYAddressing, 4},
		0xA1: {addressing.IndirectXAddressing, 6},
		0xB1: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.a = cpu.bus.Read(addr)
		cpu.setFlag(core.zFLag, cpu.a == 0x00)
		cpu.setFlag(core.nFLag, cpu.a&0x80 != 0x00)
		return true
	},
}

// LDX - Load X register
var ldx = instruction{
	Name: "LDX",
	OpCodes: opCodesMap{
		0xA2: {addressing.ImmediateAddressing, 2},
		0xA6: {addressing.ZeroPageAddressing, 3},
		0xB6: {addressing.ZeroPageYAddressing, 4},
		0xAE: {addressing.AbsoluteAddressing, 4},
		0xBE: {addressing.AbsoluteYAddressing, 4},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.x = cpu.bus.Read(addr)
		cpu.setFlag(core.zFLag, cpu.x == 0x00)
		cpu.setFlag(core.nFLag, cpu.x&0x80 != 0x00)
		return true
	},
}

// LDY - Load Y register
var ldy = instruction{
	Name: "LDY",
	OpCodes: opCodesMap{
		0xA0: {addressing.ImmediateAddressing, 2},
		0xA4: {addressing.ZeroPageAddressing, 3},
		0xB4: {addressing.ZeroPageXAddressing, 4},
		0xAC: {addressing.AbsoluteAddressing, 4},
		0xBC: {addressing.AbsoluteXAddressing, 4},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.y = cpu.bus.Read(addr)
		cpu.setFlag(core.zFLag, cpu.y == 0x00)
		cpu.setFlag(core.nFLag, cpu.y&0x80 != 0x00)
		return true
	},
}

// LSR - Logical shift right
var lsr = instruction{
	Name: "LSR",
	OpCodes: opCodesMap{
		0x4A: {addressing.AccumulatorAddressing, 2},
		0x46: {addressing.ZeroPageAddressing, 5},
		0x56: {addressing.ZeroPageXAddressing, 6},
		0x4E: {addressing.AbsoluteAddressing, 6},
		0x5E: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		var data uint8

		// Get correct data
		if addrMode == addressing.AccumulatorAddressing {
			data = cpu.a
		} else {
			data = cpu.bus.Read(addr)
		}

		cpu.setFlag(core.cFlag, data&0x01 != 0x00)
		data = data >> 1
		cpu.setFlag(core.zFLag, data == 0x00)
		cpu.setFlag(core.nFLag, data&0x80 != 0x00)

		// Save result to correct place
		if addrMode == addressing.AccumulatorAddressing {
			cpu.a = data
		} else {
			cpu.bus.Write(addr, data)
		}

		return false
	},
}

// NOP - No operation
// TODO: there is a lot of unofficial op codes that does NOP, add them
var nop = instruction{
	Name: "NOP",
	OpCodes: opCodesMap{
		0xEA: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		return false
	},
}

// ORA - Logical inclusive OR
var ora = instruction{
	Name: "ORA",
	OpCodes: opCodesMap{
		0x09: {addressing.ImmediateAddressing, 2},
		0x05: {addressing.ZeroPageAddressing, 3},
		0x15: {addressing.ZeroPageXAddressing, 4},
		0x0D: {addressing.AbsoluteAddressing, 4},
		0x1D: {addressing.AbsoluteXAddressing, 4},
		0x19: {addressing.AbsoluteYAddressing, 4},
		0x01: {addressing.IndirectXAddressing, 6},
		0x11: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.a |= cpu.bus.Read(addr)
		cpu.setFlag(core.zFLag, cpu.a == 0x00)
		cpu.setFlag(core.nFLag, cpu.a&0x80 != 0x00)

		return true
	},
}

// ROL - Rotate left
var rol = instruction{
	Name: "ROL",
	OpCodes: opCodesMap{
		0x2A: {addressing.AccumulatorAddressing, 2},
		0x26: {addressing.ZeroPageAddressing, 5},
		0x36: {addressing.ZeroPageXAddressing, 6},
		0x2E: {addressing.AbsoluteAddressing, 6},
		0x3E: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		var data uint8

		if addrMode == addressing.AccumulatorAddressing {
			data = cpu.a
		} else {
			data = cpu.bus.Read(addr)
		}

		tmpC := cpu.getFlag(core.cFlag)
		cpu.setFlag(core.cFlag, data&0b10000000 != 0)
		data = data << 1

		if tmpC {
			data |= 0b00000001
		}

		cpu.setFlag(core.zFLag, data == 0x00)
		cpu.setFlag(core.nFLag, data&0x80 != 0x00)

		if addrMode == addressing.AccumulatorAddressing {
			cpu.a = data
		} else {
			cpu.bus.Write(addr, data)
		}

		return false
	},
}

// ROR - Rotate right
var ror = instruction{
	Name: "ROR",
	OpCodes: opCodesMap{
		0x6A: {addressing.AccumulatorAddressing, 2},
		0x66: {addressing.ZeroPageAddressing, 5},
		0x76: {addressing.ZeroPageXAddressing, 6},
		0x6E: {addressing.AbsoluteAddressing, 6},
		0x7E: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		var data uint8

		if addrMode == addressing.AccumulatorAddressing {
			data = cpu.a
		} else {
			data = cpu.bus.Read(addr)
		}

		tmpC := cpu.getFlag(core.cFlag)
		cpu.setFlag(core.cFlag, data&0b00000001 != 0)
		data = data >> 1

		if tmpC {
			data |= 0b10000000
		}

		cpu.setFlag(core.zFLag, data == 0x00)
		cpu.setFlag(core.nFLag, data&0x80 != 0x00)

		if addrMode == addressing.AccumulatorAddressing {
			cpu.a = data
		} else {
			cpu.bus.Write(addr, data)
		}

		return false
	},
}

// RTI - Return from interrupt
var rti = instruction{
	Name: "RTI",
	OpCodes: opCodesMap{
		0x40: {addressing.ImpliedAddressing, 6},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		// https://wiki.nesdev.com/w/index.php/Status_flags - ignore 4 and 5 bit - make sure 5 is set in p register
		cpu.p = cpu.pullFromStack()&0b11001111 | 0b00100000
		cpu.pc = cpu.pullFromStack16()

		return false
	},
}

// RTS - Return from subroutine
var rts = instruction{
	Name: "RTS",
	OpCodes: opCodesMap{
		0x60: {addressing.ImpliedAddressing, 6},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.pc = cpu.pullFromStack16() + 1

		return false
	},
}

// SEC - Set carry flag
var sec = instruction{
	Name: "SEC",
	OpCodes: opCodesMap{
		0x38: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(core.cFlag, true)

		return false
	},
}

// SED - Set decimal flag
var sed = instruction{
	Name: "SED",
	OpCodes: opCodesMap{
		0xF8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(core.dFLag, true)

		return false
	},
}

// SEI - Set interrupt disable flag
var sei = instruction{
	Name: "SEI",
	OpCodes: opCodesMap{
		0x78: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.setFlag(core.iFLag, true)

		return false
	},
}

// STA - Store accumulator
var sta = instruction{
	Name: "STA",
	OpCodes: opCodesMap{
		0x85: {addressing.ZeroPageAddressing, 3},
		0x95: {addressing.ZeroPageXAddressing, 4},
		0x8D: {addressing.AbsoluteAddressing, 4},
		0x9D: {addressing.AbsoluteXAddressing, 5},
		0x99: {addressing.AbsoluteYAddressing, 5},
		0x81: {addressing.IndirectXAddressing, 6},
		0x91: {addressing.IndirectYAddressing, 6},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.bus.Write(addr, cpu.a)

		return false
	},
}

// STX - Store X register
var stx = instruction{
	Name: "STX",
	OpCodes: opCodesMap{
		0x86: {addressing.ZeroPageAddressing, 3},
		0x96: {addressing.ZeroPageYAddressing, 4},
		0x8E: {addressing.AbsoluteAddressing, 4},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.bus.Write(addr, cpu.x)

		return false
	},
}

// STY - Store Y register
var sty = instruction{
	Name: "STY",
	OpCodes: opCodesMap{
		0x84: {addressing.ZeroPageAddressing, 3},
		0x94: {addressing.ZeroPageXAddressing, 4},
		0x8C: {addressing.AbsoluteAddressing, 4},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.bus.Write(addr, cpu.y)

		return false
	},
}

// TAX - Transfer accumulator to X
var tax = instruction{
	Name: "TAX",
	OpCodes: opCodesMap{
		0xAA: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.x = cpu.a
		cpu.setFlag(core.zFLag, cpu.x == 0x00)
		cpu.setFlag(core.nFLag, cpu.x&0x80 != 0x00)

		return false
	},
}

// TAY - Transfer accumulator to Y
var tay = instruction{
	Name: "TAY",
	OpCodes: opCodesMap{
		0xA8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.y = cpu.a
		cpu.setFlag(core.zFLag, cpu.y == 0x00)
		cpu.setFlag(core.nFLag, cpu.y&0x80 != 0x00)

		return false
	},
}

// TSX - Transfer stack pointer to X
var tsx = instruction{
	Name: "TSX",
	OpCodes: opCodesMap{
		0xBA: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.x = cpu.sp
		cpu.setFlag(core.zFLag, cpu.x == 0x00)
		cpu.setFlag(core.nFLag, cpu.x&0x80 != 0x00)

		return false
	},
}

// TXA - Transfer X to accumulator
var txa = instruction{
	Name: "TXA",
	OpCodes: opCodesMap{
		0x8A: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.a = cpu.x
		cpu.setFlag(core.zFLag, cpu.a == 0x00)
		cpu.setFlag(core.nFLag, cpu.a&0x80 != 0x00)

		return false
	},
}

// TXS - Transfer X to stack pointer
var txs = instruction{
	Name: "TXS",
	OpCodes: opCodesMap{
		0x9A: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.sp = cpu.x

		return false
	},
}

// TYA - Transfer Y to accumulator
var tya = instruction{
	Name: "TYA",
	OpCodes: opCodesMap{
		0x98: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu *core.CPU, addr uint16, opCode uint8, addrMode int) bool {
		cpu.a = cpu.y
		cpu.setFlag(core.zFLag, cpu.a == 0x00)
		cpu.setFlag(core.nFLag, cpu.a&0x80 != 0x00)

		return false
	},
}
