package cpu

// CPU represents 6502 processor
type CPU struct {
	// Program Counter
	pc uint16

	// Stack Pointer
	sp uint8

	// Accumulator
	a uint8

	// X register
	x uint8

	// Y register
	y uint8

	// Processor status flags
	p uint8

	// Additional cycles to perform
	cycles uint8
}

type flag uint8

const (
	// Carry
	c flag = 0
	// Zero
	z flag = 1
	// Interrupt disable
	i flag = 2
	// Deciman mode
	d flag = 3
	// Break command
	b flag = 4
	// Unused flag, in processor's manual called "expansion bit", always set to 1
	u flag = 5
	// Overflow flag
	v flag = 6
	// Negative flag
	n flag = 7
)

// Addressing mode is a function that returns true if there is a potantial that
// it might require additional clock cycle
type addressingMode func() bool

// Instruction handler is a function that returns true if there is a potantial that
// it might require additional clock cycle
type instructionHandler func(*CPU) bool

type instruction struct {
	// Human readable name - for debugging
	name string

	// Op code - to have this info inside the struct too
	opCode uint8

	// Number of cycles instruction takes
	noCycles uint8

	// Addresing mode
	addrMode addressingMode

	// Istruction handler
	handler instructionHandler
}

func dummyAddressing() bool {
	return false
}

func dummyHandler(c *CPU) bool {
	return false
}

var instructions = map[uint8]instruction{
	iny.opCode: iny,
	inx.opCode: inx,
}

func (cpu *CPU) setFlag(fn flag, v bool) {
	var flag uint8 = 1 << fn

	if v {
		cpu.p |= flag
	} else {
		cpu.p &= ^flag
	}
}

func (cpu *CPU) clock() {
	if cpu.cycles == 0 {
		// 1. Read opcode
		opCode := uint8(0x00)
		instruction, ok := instructions[opCode]

		// Unknown opcode - quit
		if !ok {
			return
		}

		// 2. Set unused flag to 1
		cpu.setFlag(u, true)

		// 3. Increment Program Counter
		cpu.pc++

		// 4. Execute instruction
		cpu.cycles = instruction.noCycles

		// We might need to add additional cycle
		if instruction.addrMode() && instruction.handler(cpu) {
			cpu.cycles++
		}

		// 5. Set back unused flat to 1
		cpu.setFlag(u, true)
	}

	cpu.cycles--
}
