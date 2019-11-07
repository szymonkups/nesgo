package instructions

import (
	"fmt"
	"github.com/szymonkups/nesgo/core/addressing"
	"github.com/szymonkups/nesgo/core/flags"
	"strings"
)

var instLookup = map[uint8]*Instruction{}

func init() {
	// Populate Instruction lookup for easy access to instructions by OpCode
	for _, instruction := range instructions {
		for op, _ := range instruction.AddrByOpCode {
			instLookup[op] = instruction
		}
	}
}

type InstructionDebugInfo struct {
	InstructionName string
	OpCode          uint8
	AddressingName  string
	Size            uint8
	Operand         string
	AddressInfo     string
}

func GetInstructionDebugInfo(opCode uint8, cpu CPUInterface) (*InstructionDebugInfo, error) {
	instruction, ok := GetInstructionByOpCode(opCode)

	if !ok {
		return nil, fmt.Errorf("cannot find Instruction by opcode $%02X", opCode)
	}

	addrModeId := instruction.AddrByOpCode[opCode].AddrMode
	addrMode, ok := addressing.GetAddressingById(addrModeId)

	if !ok {
		return nil, fmt.Errorf("cannot find addressing mode with id %d for opcode $%02X", addrModeId, opCode)
	}

	addr, _ := addrMode.CalculateAddress(cpu.GetPC(), cpu.GetX(), cpu.GetY(), cpu.Read)

	info := new(InstructionDebugInfo)
	info.InstructionName = instruction.Name
	info.OpCode = opCode
	info.AddressingName = addrMode.Name
	info.Size = addrMode.Size
	if addrModeId == addressing.RelativeAddressing {
		info.Operand = addrMode.Format(addr)
	} else {
		info.Operand = addrMode.Format(uint16(cpu.Read(addr)))
	}

	info.AddressInfo = addrDescription(addr)
	return info, nil
}

func addrDescription(addr uint16) string {
	desc := ""

	switch addr {
	case 0x2000:
		desc = "PPU_CTRL_REG1"
	}

	return desc
}

func ExecuteInstruction(opCode uint8, cpu CPUInterface) error {
	// Find instruction by opcode
	instruction, ok := GetInstructionByOpCode(opCode)

	if !ok {
		return fmt.Errorf("cannot find Instruction by opcode $%02X", opCode)
	}

	// Check which addressing is used by checking by opcode
	addrModeId := instruction.AddrByOpCode[opCode].AddrMode

	// Check how many cycles instruction with this addressing needs
	cycles := instruction.AddrByOpCode[opCode].Cycles

	// Get addressing mode by id stored in instruction
	addrMode, ok := addressing.GetAddressingById(addrModeId)

	if !ok {
		return fmt.Errorf("cannot find addressing mode with id %d for opcode $%02X", addrModeId, opCode)
	}

	// Calculate operand address using addressing mode
	addr, addCycleAddr := addrMode.CalculateAddress(cpu.GetPC(), cpu.GetX(), cpu.GetY(), cpu.Read)

	// Move Program Counter forward - opcode + operand
	cpu.SetPC(cpu.GetPC() + uint16(addrMode.Size))

	// Execute instruction
	addCycleInst := instruction.Handler(cpu, addr, addrModeId)

	// Add cycles needed by this instruction
	cpu.AddCycles(cycles)

	// Sometimes one cycle must be added when addressing and instruction require it
	if addCycleAddr && addCycleInst {
		cpu.AddCycles(1)
	}

	return nil
}

func GetInstructionByName(name string) (*Instruction, bool) {
	name = strings.ToUpper(name)
	for _, inst := range instructions {
		if inst.Name == name {
			return inst, true
		}
	}

	return nil, false
}

func GetInstructionByOpCode(op uint8) (*Instruction, bool) {
	inst, ok := instLookup[op]

	return inst, ok
}

type CPUInterface interface {
	// Program Counter
	GetPC() uint16
	SetPC(uint16) uint16

	// Stack Pointer
	GetSP() uint8
	SetSP(uint8) uint8

	// Accumulator
	GetA() uint8
	SetA(uint8) uint8

	// X register
	GetX() uint8
	SetX(uint8) uint8

	// Y register
	GetY() uint8
	SetY(uint8) uint8

	// Processor status flags
	GetStatusFlags() *flags.Flags

	// Get/write data
	Read(uint16) uint8
	Read16(uint16) uint16
	Write(uint16, uint8)
	Write16(uint16, uint16)

	// Stack
	PushToStack(uint8)
	PushToStack16(uint16)
	PullFromStack() uint8
	PullFromStack16() uint16

	// Add Cycles
	AddCycles(uint8)
}

// Instruction set
// ADC AND ASL BCC BCS BEQ BIT BMI BNE BPL BRK BVC BVS CLC
// CLD CLI CLV CMP CPX CPY DEC DEX DEY EOR INC INX INY JMP
// JSR LDA LDX LDY LSR NOP ORA PHA PHP PLA PLP ROL ROR RTI
// RTS SBC SEC SED SEI STA STX STY TAX TAY TSX TXA TXS TYA
var instructions = []*Instruction{
	&adc, &and, &asl, &bcc, &bcs, &beq, &bit, &bmi, &bne, &bpl, &brk, &bvc, &bvs, &clc,
	&cld, &cli, &clv, &cmp, &cpx, &cpy, &dec, &dex, &dey, &eor, &inc, &inx, &iny, &jmp,
	&jsr, &lda, &ldx, &ldy, &lsr, &nop, &ora, &pha, &php, &pla, &plp, &rol, &ror, &rti,
	&rts, &sbc, &sec, &sed, &sei, &sta, &stx, &sty, &tax, &tay, &tsx, &txa, &txs, &tya,
}

// Instruction - describes Instruction
type Instruction struct {
	// Human readable name - for debugging
	Name string

	// Op codes op code per addressing mode
	AddrByOpCode opCodesMap

	// Instruction handler
	Handler instructionHandler
}

type opCodesMap map[uint8]struct {
	AddrMode int
	Cycles   uint8
}

// Actual Instruction code. It should return true if there is a potential to
// add additional clock cycle - this will be checked together with addressing
// mode result. If both return true it will mean that one cycle needs to be
// added to Instruction Cycles.
// We pass CPU instance, absolute address calculated by correct addressing mode,
// actual op code (as same Instruction can have different op codes depending on
// addressing mode) and addressing mode itself.
type instructionHandler func(cpuState CPUInterface, addr uint16, addrMode int) bool

// *****************************************************************************
// Instructions
// *****************************************************************************

// ADC - Add with carry
var adc = Instruction{
	Name: "ADC",
	AddrByOpCode: opCodesMap{
		0x69: {addressing.ImmediateAddressing, 2},
		0x65: {addressing.ZeroPageAddressing, 3},
		0x75: {addressing.ZeroPageXAddressing, 4},
		0x6D: {addressing.AbsoluteAddressing, 4},
		0x7D: {addressing.AbsoluteXAddressing, 4},
		0x79: {addressing.AbsoluteYAddressing, 4},
		0x61: {addressing.IndirectXAddressing, 6},
		0x71: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		acc := cpu.GetA()
		data := cpu.Read(addr)
		carry := uint8(0)
		flg := cpu.GetStatusFlags()

		if flg.Get(flags.C) {
			carry = 1
		}

		newAcc := cpu.SetA(acc + data + carry)
		flg.Set(flags.C, int(acc)+int(data)+int(carry) > 0xFF)
		flg.Set(flags.V, (acc^data)&0x80 == 0 && (acc^newAcc)&0x80 != 0)
		flg.SetZN(newAcc)

		return true
	},
}

// SBC - Subtract with carry
var sbc = Instruction{
	Name: "SBC",
	AddrByOpCode: opCodesMap{
		0xE9: {addressing.ImmediateAddressing, 2},
		0xE5: {addressing.ZeroPageAddressing, 3},
		0xF5: {addressing.ZeroPageXAddressing, 4},
		0xED: {addressing.AbsoluteAddressing, 4},
		0xFD: {addressing.AbsoluteXAddressing, 4},
		0xF9: {addressing.AbsoluteYAddressing, 4},
		0xE1: {addressing.IndirectXAddressing, 6},
		0xF1: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		acc := cpu.GetA()
		data := cpu.Read(addr)
		carry := uint8(0)
		flg := cpu.GetStatusFlags()

		if flg.Get(flags.C) {
			carry = 1
		}

		newAcc := cpu.SetA(acc - data - (1 - carry))

		flg.Set(flags.C, int(acc)-int(data)-int(1-carry) >= 0)
		flg.Set(flags.V, (acc^data)&0x80 != 0 && (acc^newAcc)&0x80 != 0)
		flg.SetZN(newAcc)
		return true
	},
}

// ASL - Arithmetic shift left
var asl = Instruction{
	Name: "ASL",
	AddrByOpCode: opCodesMap{
		0x0A: {addressing.AccumulatorAddressing, 2},
		0x06: {addressing.ZeroPageAddressing, 5},
		0x16: {addressing.ZeroPageXAddressing, 6},
		0x0E: {addressing.AbsoluteAddressing, 6},
		0x1E: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		data := cpu.Read(addr)
		var res16 = uint16(data) << 1
		flg := cpu.GetStatusFlags()

		flg.Set(flags.C, (res16&0xFF00) > 0)
		flg.Set(flags.Z, (res16&0x00FF) == 0x00)
		flg.Set(flags.N, res16&0x80 != 0)

		res := uint8(res16 & 0x00FF)

		if addrMode == addressing.AccumulatorAddressing {
			cpu.SetA(res)
		} else {
			cpu.Write(addr, res)
		}

		return false
	},
}

// AND - Bitwise logic AND
var and = Instruction{
	Name: "AND",
	AddrByOpCode: opCodesMap{
		0x29: {addressing.ImmediateAddressing, 2},
		0x25: {addressing.ZeroPageAddressing, 3},
		0x35: {addressing.ZeroPageXAddressing, 4},
		0x2D: {addressing.AbsoluteAddressing, 4},
		0x3D: {addressing.AbsoluteXAddressing, 4},
		0x39: {addressing.AbsoluteYAddressing, 4},
		0x21: {addressing.IndirectXAddressing, 6},
		0x31: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		data := cpu.Read(addr)
		newAcc := cpu.SetA(cpu.GetA() & data)
		cpu.GetStatusFlags().SetZN(newAcc)

		return true
	},
}

// BCC - Branch if carry is clear
var bcc = Instruction{
	Name: "BCC",
	AddrByOpCode: opCodesMap{
		0x90: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		flg := cpu.GetStatusFlags()
		if !flg.Get(flags.C) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BCS - Branch if carry is set
var bcs = Instruction{
	Name: "BCS",
	AddrByOpCode: opCodesMap{
		0xB0: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		flg := cpu.GetStatusFlags()
		if flg.Get(flags.C) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BEQ - Branch if equal - zero flag is set
var beq = Instruction{
	Name: "BEQ",
	AddrByOpCode: opCodesMap{
		0xF0: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		flg := cpu.GetStatusFlags()

		if flg.Get(flags.Z) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BMI - Branch if minus - negative flag is set
var bmi = Instruction{
	Name: "BMI",
	AddrByOpCode: opCodesMap{
		0x30: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		flg := cpu.GetStatusFlags()
		if flg.Get(flags.N) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BNE - Branch if no equal - zero flag is not set
var bne = Instruction{
	Name: "BNE",
	AddrByOpCode: opCodesMap{
		0xD0: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		flg := cpu.GetStatusFlags()
		if !flg.Get(flags.Z) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BPL - Branch if positive - negative flag is not set
var bpl = Instruction{
	Name: "BPL",
	AddrByOpCode: opCodesMap{
		0x10: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		flg := cpu.GetStatusFlags()
		if !flg.Get(flags.N) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BVC - Branch if overflow clear
var bvc = Instruction{
	Name: "BVC",
	AddrByOpCode: opCodesMap{
		0x50: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		flg := cpu.GetStatusFlags()
		if !flg.Get(flags.V) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

// BVS - Branch if overflow is set
var bvs = Instruction{
	Name: "BVS",
	AddrByOpCode: opCodesMap{
		0x70: {addressing.RelativeAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		flg := cpu.GetStatusFlags()

		if flg.Get(flags.V) {
			branchHandler(cpu, addr)
		}

		return false
	},
}

func branchHandler(cpu CPUInterface, addr uint16) {
	cpu.AddCycles(1)

	// Page might be crossed - add one cycle if that happens
	if (addr & 0xFF00) != (cpu.GetPC() & 0xFF00) {
		cpu.AddCycles(1)
	}

	cpu.SetPC(addr)
}

// BIT - bit test
var bit = Instruction{
	Name: "BIT",
	AddrByOpCode: opCodesMap{
		0x24: {addressing.ZeroPageAddressing, 3},
		0x2C: {addressing.AbsoluteAddressing, 4},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		data := cpu.Read(addr)
		tmp := data & cpu.GetA()
		flg := cpu.GetStatusFlags()

		flg.Set(flags.Z, tmp == 0x00)
		flg.Set(flags.N, data&0b010000000 != 0)
		flg.Set(flags.V, data&0b001000000 != 0)

		return false
	},
}

// BRK - Force interrupt
var brk = Instruction{
	Name: "BRK",
	AddrByOpCode: opCodesMap{
		0x00: {addressing.ImpliedAddressing, 7},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		// Padding byte
		// http://nesdev.com/the%20%27B%27%20flag%20&%20BRK%20instruction.txt
		cpu.SetPC(cpu.GetPC() + 1)

		flg := cpu.GetStatusFlags()

		// Set interrupt flag
		flg.Set(flags.I, true)

		// Push Program Counter to stack
		cpu.PushToStack16(cpu.GetPC())

		// Push processor status flags to stack
		// https://wiki.nesdev.com/w/index.php/Status_flags - set 4 and 5 bit before pushing
		cpu.PushToStack(flg.GetByte() | 0b00110000)

		// Get data from 0xFFFE and 0xFFFF and set PC
		cpu.SetPC(cpu.Read16(0xFFFE))

		return false
	},
}

// INX - Increment X register by one
var inx = Instruction{
	Name: "INX",
	AddrByOpCode: opCodesMap{
		0xE8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newX := cpu.SetX(cpu.GetX() + 1)
		cpu.GetStatusFlags().SetZN(newX)

		return false
	},
}

// INY - Increment Y register by one
var iny = Instruction{
	Name: "INY",
	AddrByOpCode: opCodesMap{
		0xC8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newY := cpu.SetY(cpu.GetY() + 1)
		cpu.GetStatusFlags().SetZN(newY)

		return false
	},
}

// INC - Increment memory location
var inc = Instruction{
	Name: "INC",
	AddrByOpCode: opCodesMap{
		0xE6: {addressing.ZeroPageAddressing, 5},
		0xF6: {addressing.ZeroPageXAddressing, 6},
		0xEE: {addressing.AbsoluteAddressing, 6},
		0xFE: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		data := cpu.Read(addr)
		data++
		cpu.Write(addr, data)
		cpu.GetStatusFlags().SetZN(data)

		return false
	},
}

// CLC - Clear carry flag
var clc = Instruction{
	Name: "CLC",
	AddrByOpCode: opCodesMap{
		0x18: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.GetStatusFlags().Set(flags.C, false)

		return false
	},
}

// CLD - Clear decimal flag
var cld = Instruction{
	Name: "CLD",
	AddrByOpCode: opCodesMap{
		0xD8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.GetStatusFlags().Set(flags.D, false)

		return false
	},
}

// CLI - Clear interrupt flag (disable interrupts)
var cli = Instruction{
	Name: "CLI",
	AddrByOpCode: opCodesMap{
		0x58: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.GetStatusFlags().Set(flags.I, false)

		return false
	},
}

// CLV - Clear overflow flag
var clv = Instruction{
	Name: "CLV",
	AddrByOpCode: opCodesMap{
		0xB8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.GetStatusFlags().Set(flags.V, false)

		return false
	},
}

// CMP - Compare accumulator
var cmp = Instruction{
	Name: "CMP",
	AddrByOpCode: opCodesMap{
		0xC9: {addressing.ImmediateAddressing, 2},
		0xC5: {addressing.ZeroPageAddressing, 3},
		0xD5: {addressing.ZeroPageXAddressing, 4},
		0xCD: {addressing.AbsoluteAddressing, 4},
		0xDD: {addressing.AbsoluteXAddressing, 4},
		0xD9: {addressing.AbsoluteYAddressing, 4},
		0xC1: {addressing.IndirectXAddressing, 6},
		0xD1: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		compareHandler(cpu, cpu.GetA(), addr)

		return true
	},
}

// CPX - Compare X register
var cpx = Instruction{
	Name: "CPX",
	AddrByOpCode: opCodesMap{
		0xE0: {addressing.ImmediateAddressing, 2},
		0xE4: {addressing.ZeroPageAddressing, 3},
		0xEC: {addressing.AbsoluteAddressing, 4},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		compareHandler(cpu, cpu.GetX(), addr)

		return false
	},
}

// CPY - Compare Y register
var cpy = Instruction{
	Name: "CPY",
	AddrByOpCode: opCodesMap{
		0xC0: {addressing.ImmediateAddressing, 2},
		0xC4: {addressing.ZeroPageAddressing, 3},
		0xCC: {addressing.AbsoluteAddressing, 4},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		compareHandler(cpu, cpu.GetY(), addr)

		return false
	},
}

func compareHandler(cpu CPUInterface, a byte, addr uint16) {
	data := cpu.Read(addr)
	diff := a - data

	flg := cpu.GetStatusFlags()
	flg.Set(flags.C, a >= data)
	flg.SetZN(diff)
}

// DEC - decrement memory
var dec = Instruction{
	Name: "DEC",
	AddrByOpCode: opCodesMap{
		0xC6: {addressing.ZeroPageAddressing, 5},
		0xD6: {addressing.ZeroPageXAddressing, 6},
		0xCE: {addressing.AbsoluteAddressing, 6},
		0xDE: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		data := cpu.Read(addr)
		data--
		cpu.Write(addr, data)
		cpu.GetStatusFlags().SetZN(data)

		return false
	},
}

// DEX - decrement X register
var dex = Instruction{
	Name: "DEX",
	AddrByOpCode: opCodesMap{
		0xCA: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newX := cpu.SetX(cpu.GetX() - 1)
		cpu.GetStatusFlags().SetZN(newX)

		return false
	},
}

// DEY - decrement Y register
var dey = Instruction{
	Name: "DEY",
	AddrByOpCode: opCodesMap{
		0x88: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newY := cpu.SetY(cpu.GetY() - 1)

		flg := cpu.GetStatusFlags()
		flg.Set(flags.Z, newY == 0)
		flg.Set(flags.N, newY&0x80 != 0)

		return false
	},
}

// EOR - Exclusive OR
var eor = Instruction{
	Name: "EOR",
	AddrByOpCode: opCodesMap{
		0x49: {addressing.ImmediateAddressing, 2},
		0x45: {addressing.ZeroPageAddressing, 3},
		0x55: {addressing.ZeroPageXAddressing, 4},
		0x4D: {addressing.AbsoluteAddressing, 4},
		0x5D: {addressing.AbsoluteXAddressing, 4},
		0x59: {addressing.AbsoluteYAddressing, 4},
		0x41: {addressing.IndirectXAddressing, 6},
		0x51: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		data := cpu.Read(addr)
		newA := cpu.SetA(cpu.GetA() ^ data)
		cpu.GetStatusFlags().SetZN(newA)

		return true
	},
}

// PHA - Push accumulator to stack
var pha = Instruction{
	Name: "PHA",
	AddrByOpCode: opCodesMap{
		0x48: {addressing.ImpliedAddressing, 3},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.PushToStack(cpu.GetA())
		return false
	},
}

// PHP - Push processor status to stack
var php = Instruction{
	Name: "PHP",
	AddrByOpCode: opCodesMap{
		0x08: {addressing.ImpliedAddressing, 3},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		// Push processor status flags to stack
		// https://wiki.nesdev.com/w/index.php/Status_flags - set 4 and 5 bit before pushing
		cpu.PushToStack(cpu.GetStatusFlags().GetByte() | 0b00110000)
		return false
	},
}

// PLA = Pull accumulator from stack
var pla = Instruction{
	Name: "PLA",
	AddrByOpCode: opCodesMap{
		0x68: {addressing.ImpliedAddressing, 4},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newA := cpu.SetA(cpu.PullFromStack())
		cpu.GetStatusFlags().SetZN(newA)

		return false
	},
}

// PLP = Pull processor status from stack
var plp = Instruction{
	Name: "PLP",
	AddrByOpCode: opCodesMap{
		0x28: {addressing.ImpliedAddressing, 4},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		// https://wiki.nesdev.com/w/index.php/Status_flags - ignore 4 and 5 bit - make sure 5 is set in p register
		cpu.GetStatusFlags().SetByte(cpu.PullFromStack()&0b11001111 | 0b00100000)

		return false
	},
}

// JMP = Jump to address
var jmp = Instruction{
	// TODO: check from http://obelisk.me.uk/6502/reference.html#JMP
	// An original 6502 has does not correctly fetch the target address if the indirect vector falls on a page
	// boundary (e.g. $xxFF where xx is any value from $00 to $FF). In this case fetches the LSB from $xxFF as
	// expected but takes the MSB from $xx00. This is fixed in some later chips like the 65SC02 so for compatibility
	// always ensure the indirect vector is not at the end of the page.

	Name: "JMP",
	AddrByOpCode: opCodesMap{
		0x4C: {addressing.AbsoluteAddressing, 3},
		0x6C: {addressing.IndirectAddressing, 5},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.SetPC(addr)
		return false
	},
}

// JSR - Jump to subroutine
var jsr = Instruction{
	Name: "JSR",
	AddrByOpCode: opCodesMap{
		0x20: {addressing.AbsoluteAddressing, 6},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.PushToStack16(cpu.GetPC() - 1)
		cpu.SetPC(addr)

		return false
	},
}

// LDA - Load accumulator
var lda = Instruction{
	Name: "LDA",
	AddrByOpCode: opCodesMap{
		0xA9: {addressing.ImmediateAddressing, 2},
		0xA5: {addressing.ZeroPageAddressing, 3},
		0xB5: {addressing.ZeroPageXAddressing, 4},
		0xAD: {addressing.AbsoluteAddressing, 4},
		0xBD: {addressing.AbsoluteXAddressing, 4},
		0xB9: {addressing.AbsoluteYAddressing, 4},
		0xA1: {addressing.IndirectXAddressing, 6},
		0xB1: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newA := cpu.SetA(cpu.Read(addr))
		cpu.GetStatusFlags().SetZN(newA)

		return true
	},
}

// LDX - Load X register
var ldx = Instruction{
	Name: "LDX",
	AddrByOpCode: opCodesMap{
		0xA2: {addressing.ImmediateAddressing, 2},
		0xA6: {addressing.ZeroPageAddressing, 3},
		0xB6: {addressing.ZeroPageYAddressing, 4},
		0xAE: {addressing.AbsoluteAddressing, 4},
		0xBE: {addressing.AbsoluteYAddressing, 4},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newX := cpu.SetX(cpu.Read(addr))
		cpu.GetStatusFlags().SetZN(newX)

		return true
	},
}

// LDY - Load Y register
var ldy = Instruction{
	Name: "LDY",
	AddrByOpCode: opCodesMap{
		0xA0: {addressing.ImmediateAddressing, 2},
		0xA4: {addressing.ZeroPageAddressing, 3},
		0xB4: {addressing.ZeroPageXAddressing, 4},
		0xAC: {addressing.AbsoluteAddressing, 4},
		0xBC: {addressing.AbsoluteXAddressing, 4},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newY := cpu.SetY(cpu.Read(addr))
		cpu.GetStatusFlags().SetZN(newY)

		return true
	},
}

// LSR - Logical shift right
var lsr = Instruction{
	Name: "LSR",
	AddrByOpCode: opCodesMap{
		0x4A: {addressing.AccumulatorAddressing, 2},
		0x46: {addressing.ZeroPageAddressing, 5},
		0x56: {addressing.ZeroPageXAddressing, 6},
		0x4E: {addressing.AbsoluteAddressing, 6},
		0x5E: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		var data uint8

		// Get correct data
		if addrMode == addressing.AccumulatorAddressing {
			data = cpu.GetA()
		} else {
			data = cpu.Read(addr)
		}

		flg := cpu.GetStatusFlags()
		flg.Set(flags.C, data&0x01 != 0x00)
		data = data >> 1
		flg.SetZN(data)

		// Save result to correct place
		if addrMode == addressing.AccumulatorAddressing {
			cpu.SetA(data)
		} else {
			cpu.Write(addr, data)
		}

		return false
	},
}

// NOP - No operation
// TODO: there is a lot of unofficial op codes that does NOP, add them
var nop = Instruction{
	Name: "NOP",
	AddrByOpCode: opCodesMap{
		0xEA: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		return false
	},
}

// ORA - Logical inclusive OR
var ora = Instruction{
	Name: "ORA",
	AddrByOpCode: opCodesMap{
		0x09: {addressing.ImmediateAddressing, 2},
		0x05: {addressing.ZeroPageAddressing, 3},
		0x15: {addressing.ZeroPageXAddressing, 4},
		0x0D: {addressing.AbsoluteAddressing, 4},
		0x1D: {addressing.AbsoluteXAddressing, 4},
		0x19: {addressing.AbsoluteYAddressing, 4},
		0x01: {addressing.IndirectXAddressing, 6},
		0x11: {addressing.IndirectYAddressing, 5},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newA := cpu.SetA(cpu.GetA() | cpu.Read(addr))
		cpu.GetStatusFlags().SetZN(newA)

		return true
	},
}

// ROL - Rotate left
var rol = Instruction{
	Name: "ROL",
	AddrByOpCode: opCodesMap{
		0x2A: {addressing.AccumulatorAddressing, 2},
		0x26: {addressing.ZeroPageAddressing, 5},
		0x36: {addressing.ZeroPageXAddressing, 6},
		0x2E: {addressing.AbsoluteAddressing, 6},
		0x3E: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		var data uint8

		if addrMode == addressing.AccumulatorAddressing {
			data = cpu.GetA()
		} else {
			data = cpu.Read(addr)
		}

		flg := cpu.GetStatusFlags()
		tmpC := flg.Get(flags.C)
		flg.Set(flags.C, data&0b10000000 != 0)
		data = data << 1

		if tmpC {
			data |= 0b00000001
		}

		flg.SetZN(data)

		if addrMode == addressing.AccumulatorAddressing {
			cpu.SetA(data)
		} else {
			cpu.Write(addr, data)
		}

		return false
	},
}

// ROR - Rotate right
var ror = Instruction{
	Name: "ROR",
	AddrByOpCode: opCodesMap{
		0x6A: {addressing.AccumulatorAddressing, 2},
		0x66: {addressing.ZeroPageAddressing, 5},
		0x76: {addressing.ZeroPageXAddressing, 6},
		0x6E: {addressing.AbsoluteAddressing, 6},
		0x7E: {addressing.AbsoluteXAddressing, 7},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		var data uint8

		if addrMode == addressing.AccumulatorAddressing {
			data = cpu.GetA()
		} else {
			data = cpu.Read(addr)
		}

		flg := cpu.GetStatusFlags()
		tmpC := flg.Get(flags.C)
		flg.Set(flags.C, data&0b00000001 != 0)
		data = data >> 1

		if tmpC {
			data |= 0b10000000
		}

		flg.SetZN(data)

		if addrMode == addressing.AccumulatorAddressing {
			cpu.SetA(data)
		} else {
			cpu.Write(addr, data)
		}

		return false
	},
}

// RTI - Return from interrupt
var rti = Instruction{
	Name: "RTI",
	AddrByOpCode: opCodesMap{
		0x40: {addressing.ImpliedAddressing, 6},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		// https://wiki.nesdev.com/w/index.php/Status_flags - ignore 4 and 5 bit - make sure 5 is set in p register
		cpu.GetStatusFlags().SetByte(cpu.PullFromStack()&0b11001111 | 0b00100000)
		cpu.SetPC(cpu.PullFromStack16())

		return false
	},
}

// RTS - Return from subroutine
var rts = Instruction{
	Name: "RTS",
	AddrByOpCode: opCodesMap{
		0x60: {addressing.ImpliedAddressing, 6},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.SetPC(cpu.PullFromStack16() + 1)

		return false
	},
}

// SEC - Set carry flag
var sec = Instruction{
	Name: "SEC",
	AddrByOpCode: opCodesMap{
		0x38: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.GetStatusFlags().Set(flags.C, true)

		return false
	},
}

// SED - Set decimal flag
var sed = Instruction{
	Name: "SED",
	AddrByOpCode: opCodesMap{
		0xF8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.GetStatusFlags().Set(flags.D, true)

		return false
	},
}

// SEI - Set interrupt disable flag
var sei = Instruction{
	Name: "SEI",
	AddrByOpCode: opCodesMap{
		0x78: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.GetStatusFlags().Set(flags.I, true)

		return false
	},
}

// STA - Store accumulator
var sta = Instruction{
	Name: "STA",
	AddrByOpCode: opCodesMap{
		0x85: {addressing.ZeroPageAddressing, 3},
		0x95: {addressing.ZeroPageXAddressing, 4},
		0x8D: {addressing.AbsoluteAddressing, 4},
		0x9D: {addressing.AbsoluteXAddressing, 5},
		0x99: {addressing.AbsoluteYAddressing, 5},
		0x81: {addressing.IndirectXAddressing, 6},
		0x91: {addressing.IndirectYAddressing, 6},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.Write(addr, cpu.GetA())

		return false
	},
}

// STX - Store X register
var stx = Instruction{
	Name: "STX",
	AddrByOpCode: opCodesMap{
		0x86: {addressing.ZeroPageAddressing, 3},
		0x96: {addressing.ZeroPageYAddressing, 4},
		0x8E: {addressing.AbsoluteAddressing, 4},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.Write(addr, cpu.GetX())

		return false
	},
}

// STY - Store Y register
var sty = Instruction{
	Name: "STY",
	AddrByOpCode: opCodesMap{
		0x84: {addressing.ZeroPageAddressing, 3},
		0x94: {addressing.ZeroPageXAddressing, 4},
		0x8C: {addressing.AbsoluteAddressing, 4},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.Write(addr, cpu.GetY())

		return false
	},
}

// TAX - Transfer accumulator to X
var tax = Instruction{
	Name: "TAX",
	AddrByOpCode: opCodesMap{
		0xAA: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newX := cpu.SetX(cpu.GetA())
		cpu.GetStatusFlags().SetZN(newX)

		return false
	},
}

// TAY - Transfer accumulator to Y
var tay = Instruction{
	Name: "TAY",
	AddrByOpCode: opCodesMap{
		0xA8: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newY := cpu.SetY(cpu.GetA())
		cpu.GetStatusFlags().SetZN(newY)

		return false
	},
}

// TSX - Transfer stack pointer to X
var tsx = Instruction{
	Name: "TSX",
	AddrByOpCode: opCodesMap{
		0xBA: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newX := cpu.SetX(cpu.GetSP())
		cpu.GetStatusFlags().SetZN(newX)

		return false
	},
}

// TXA - Transfer X to accumulator
var txa = Instruction{
	Name: "TXA",
	AddrByOpCode: opCodesMap{
		0x8A: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newA := cpu.SetA(cpu.GetX())
		cpu.GetStatusFlags().SetZN(newA)

		return false
	},
}

// TXS - Transfer X to stack pointer
var txs = Instruction{
	Name: "TXS",
	AddrByOpCode: opCodesMap{
		0x9A: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		cpu.SetSP(cpu.GetX())

		return false
	},
}

// TYA - Transfer Y to accumulator
var tya = Instruction{
	Name: "TYA",
	AddrByOpCode: opCodesMap{
		0x98: {addressing.ImpliedAddressing, 2},
	},
	Handler: func(cpu CPUInterface, addr uint16, addrMode int) bool {
		newA := cpu.SetA(cpu.GetY())
		cpu.GetStatusFlags().SetZN(newA)

		return false
	},
}
