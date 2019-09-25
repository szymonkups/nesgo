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
	cFlag flag = 0
	// Zero
	zFLag flag = 1
	// Interrupt disable
	iFLag flag = 2
	// Decimal mode
	dFLag flag = 3
	// Overflow flag
	vFLag flag = 6
	// Negative flag
	nFLag flag = 7
)

// NewCPU performs cpu initialization
// TODO: use single lookup for all cpu instances, use init() method for package
// to initialize it
func NewCPU(bus *Bus) CPU {
	cpu := CPU{}
	cpu.bus = bus

	cpu.instLookup = map[uint8]*instruction{}
	for _, inst := range instructions {
		for opCode := range inst.opCodes {
			cpu.instLookup[opCode] = inst
		}
	}

	cpu.Reset()

	return cpu
}

// Reset - resets cpu to known state
func (cpu *CPU) Reset() {
	// Reset registers
	cpu.a = 0
	cpu.x = 0
	cpu.y = 0
	cpu.sp = 0xFD
	cpu.p = 0b00100000

	// Stack pointer is initialized to address found under 0xFFFC
	// Where start address is stored
	addr := uint16(0xFFFC)
	low := uint16(cpu.bus.read(addr))
	high := uint16(cpu.bus.read(addr + 1))
	cpu.pc = (high << 8) | low

	// Assuming that resetting takes time
	cpu.cyclesLeft = 8
}

func (cpu *CPU) setFlag(f flag, v bool) {
	var flag uint8 = 1 << f

	if v {
		cpu.p |= flag
	} else {
		cpu.p &= ^flag
	}
}

func (cpu *CPU) getFlag(f flag) bool {
	var flag uint8 = 1 << f

	return (cpu.p & flag) != 0
}

func (cpu *CPU) pushToStack(data uint8) {
	cpu.bus.write(0x0100+uint16(cpu.sp), data)
	cpu.sp--
}

func (cpu *CPU) pushToStack16(d uint16) {
	cpu.pushToStack(uint8((d >> 8) & 0x00FF))
	cpu.pushToStack(uint8(d & 0x00FF))
}

func (cpu *CPU) pullFromStack() uint8 {
	cpu.sp++
	return cpu.bus.read(0x0100 + uint16(cpu.sp))
}

func (cpu *CPU) pullFromStack16() uint16 {
	low := uint16(cpu.pullFromStack())
	high := uint16(cpu.pullFromStack())

	return (high << 8) | low
}

// Clock - execute single clock cycle
func (cpu *CPU) Clock() {
	if cpu.cyclesLeft == 0 {
		// Read opcode
		opCode := cpu.bus.read(cpu.pc)
		instruction, ok := cpu.instLookup[opCode]

		// Unknown opcode - quit
		if !ok {
			// TODO: Think what to do here
			return
		}

		// Increment Program Counter
		cpu.pc++

		// Execute instruction
		addrMode := instruction.opCodes[opCode].addrMode
		cpu.cyclesLeft = instruction.opCodes[opCode].cycles
		address, addCycleAddr := addressingModes[addrMode](cpu)
		addCycleHandler := instruction.handler(cpu, address, opCode, addrMode)

		// We might need to add additional cycle
		if addCycleAddr && addCycleHandler {
			cpu.cyclesLeft++
		}
	}

	// One cycle done
	cpu.cyclesLeft--
}
