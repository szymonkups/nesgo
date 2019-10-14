package core

import "fmt"

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

	// If true IRQ will be scheduled
	isIRQScheduled bool

	// If true NMI will be scheduled
	isNMIScheduled bool
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
	cpu.pc = cpu.bus.Read16(0xFFFC)

	// Assuming that resetting takes time
	cpu.cyclesLeft = 8
}

func (cpu *CPU) GetCyclesLeft() uint8 {
	return cpu.cyclesLeft
}

// Clock - execute single clock cycle
func (cpu *CPU) Clock() {
	if cpu.cyclesLeft == 0 {

		// Check for scheduled interrutpts
		if cpu.isNMIScheduled {
			cpu.handleInterrupt(0xFFFA)
			cpu.isNMIScheduled = false
			return
		}

		if cpu.isIRQScheduled {
			cpu.handleInterrupt(0xFFFE)
			cpu.isIRQScheduled = false
			return
		}

		// Read instruction
		instruction, opCode, ok := cpu.getInstruction(cpu.pc)

		// Unknown opcode - quit
		if !ok {
			// TODO: Think what to do here
			return
		}

		// Find correct addressing mode for this op code
		addrMode := instruction.opCodes[opCode].addrMode

		// Set new cycles needed for this instruction
		cpu.cyclesLeft = instruction.opCodes[opCode].cycles

		// Execute addressing mode
		address, addCycleAddr := addressingModes[addrMode](cpu)

		// Increment program counter
		cpu.pc += uint16(addressingSize[addrMode])

		// Execute instruction
		addCycleHandler := instruction.handler(cpu, address, opCode, addrMode)

		// We might need to add additional cycle
		if addCycleAddr && addCycleHandler {
			cpu.cyclesLeft++
		}
	}

	// One cycle done
	cpu.cyclesLeft--
}

func (cpu *CPU) getInstruction(addr uint16) (*instruction, uint8, bool) {
	opCode := cpu.bus.Read(addr)
	inst, ok := cpu.instLookup[opCode]

	return inst, opCode, ok
}

func (cpu *CPU) scheduleIRQ() {
	if !cpu.getFlag(iFLag) {
		cpu.isIRQScheduled = true
	}
}

func (cpu *CPU) scheduleNMI() {
	cpu.isNMIScheduled = true
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
	cpu.bus.Write(0x0100+uint16(cpu.sp), data)
	cpu.sp--
}

func (cpu *CPU) pushToStack16(d uint16) {
	cpu.pushToStack(uint8((d >> 8) & 0x00FF))
	cpu.pushToStack(uint8(d & 0x00FF))
}

func (cpu *CPU) pullFromStack() uint8 {
	cpu.sp++
	return cpu.bus.Read(0x0100 + uint16(cpu.sp))
}

func (cpu *CPU) pullFromStack16() uint16 {
	low := uint16(cpu.pullFromStack())
	high := uint16(cpu.pullFromStack())

	return (high << 8) | low
}

func (cpu *CPU) handleInterrupt(addr uint16) {
	// Push program counter to stack
	cpu.pushToStack16(cpu.pc)

	// Push status to stack
	// https://wiki.nesdev.com/w/index.php/Status_flags
	// Set 5 bit and unset 4 bit
	cpu.pushToStack(cpu.p&0b11001111 | 0b00100000)

	// Disable interrupts
	cpu.setFlag(iFLag, true)

	// Read new PC
	cpu.pc = cpu.bus.Read16(addr)

	// It takes 7 cycles
	// https://wiki.nesdev.com/w/index.php/CPU_interrupts
	cpu.cyclesLeft = 8
}

type CPUDebugInfo struct {
	PC uint16
	SP uint8
	A  uint8
	X  uint8
	Y  uint8
	P  uint8
}

func (cpu *CPU) GetDebugInfo() CPUDebugInfo {
	return CPUDebugInfo{
		PC: cpu.pc,
		SP: cpu.sp,
		A:  cpu.a,
		X:  cpu.x,
		Y:  cpu.y,
		P:  cpu.p,
	}
}

func (cpu *CPU) Disassemble(addr uint16) (string, bool) {
	opCode := cpu.bus.ReadDebug(addr)

	inst, ok := cpu.instLookup[opCode]
	addrMode := inst.opCodes[opCode].addrMode
	address, _ := addressingModes[addrMode](cpu)
	assembly := inst.name

	switch addrMode {
	case accumulatorAddressing:
		assembly += " A"
	case immediateAddressing:
		assembly += fmt.Sprintf(" #$%02X  [IMM]", cpu.bus.ReadDebug(address))
	case zeroPageAddressing:
		assembly += fmt.Sprintf(" $%02X  [ZPA]", cpu.bus.ReadDebug(address))
	case zeroPageXAddressing:
		assembly += fmt.Sprintf(" $%02X,X  [ZPX]", cpu.bus.ReadDebug(address))
	case zeroPageYAddressing:
		assembly += fmt.Sprintf(" $%02X,Y  [ZPY]", cpu.bus.ReadDebug(address))
	case relativeAddressing:
		assembly += fmt.Sprintf(" $%02X  [REL]", cpu.bus.ReadDebug(address))
	case absoluteAddressing:
		assembly += fmt.Sprintf(" $%04X  [ABS]", cpu.bus.ReadDebug16(address))
	case absoluteXAddressing:
		assembly += fmt.Sprintf(" $%04X,X  [ABX]", cpu.bus.ReadDebug16(address))
	case absoluteYAddressing:
		assembly += fmt.Sprintf(" $%04X,Y  [ABX]", cpu.bus.ReadDebug16(address))
	case indirectAddressing:
		assembly += fmt.Sprintf(" ($%04X)  [IND]", cpu.bus.ReadDebug16(address))
	case indirectXAddressing:
		assembly += fmt.Sprintf(" ($%02X,X)  [INX]", cpu.bus.ReadDebug(address))
	case indirectYAddressing:
		assembly += fmt.Sprintf(" ($%02X),Y  [INY]", cpu.bus.ReadDebug(address))
	}

	return assembly, ok
}
