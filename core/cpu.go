package core

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

	// Additional cycles to perform on current instruction
	cyclesLeft uint8

	// Instructions lookup table
	instLookup map[uint8]*instruction

	// Data bus to which CPU is connected
	bus *Bus
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

// NewCPU performs cpu initialization
func NewCPU(bus *Bus) CPU {
	cpu := CPU{}
	cpu.bus = bus

	cpu.instLookup = map[uint8]*instruction{}
	for _, inst := range instructions {
		for opCode := range inst.opCodes {
			cpu.instLookup[opCode] = inst
		}
	}

	return cpu
}

func (cpu *CPU) setFlag(fn flag, v bool) {
	var flag uint8 = 1 << fn

	if v {
		cpu.p |= flag
	} else {
		cpu.p &= ^flag
	}
}

// Clock - execute single clock cycle
func (cpu *CPU) Clock() {
	if cpu.cyclesLeft == 0 {
		// 1. Read opcode
		opCode := cpu.bus.read(cpu.pc)
		instruction, ok := cpu.instLookup[opCode]

		// Unknown opcode - quit
		if !ok {
			// TODO: Think what to do here
			return
		}

		// 2. Set unused flag to 1
		cpu.setFlag(u, true)

		// 3. Increment Program Counter
		cpu.pc++

		// 4. Execute instruction
		cpu.cyclesLeft = instruction.opCodes[opCode].cycles
		data, addCycleAddr := instruction.opCodes[opCode].addrMode(cpu)
		addCycleHandler := instruction.handler(cpu, data)

		// We might need to add additional cycle
		if addCycleAddr && addCycleHandler {
			cpu.cyclesLeft++
		}

		// 5. Set back unused flat to 1
		cpu.setFlag(u, true)
	}

	// One cycle done
	cpu.cyclesLeft--
}
