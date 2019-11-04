package core

import (
	"github.com/szymonkups/nesgo/core/flags"
	"github.com/szymonkups/nesgo/core/instructions"
)

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
	p *flags.Flags

	// Additional cycles to perform on current instruction
	cyclesLeft uint8

	// Data bus to which CPU is connected
	bus *bus

	// If true IRQ will be scheduled
	isIRQScheduled bool

	// If true NMI will be scheduled
	isNMIScheduled bool
}

func NewCPU(bus *bus) *CPU {
	cpu := new(CPU)
	cpu.bus = bus
	cpu.p = new(flags.Flags)

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
	cpu.p.SetByte(0b00100000)

	// Stack pointer is initialized to address found under 0xFFFC
	// Where start address is stored
	cpu.pc = cpu.bus.Read16(0xFFFC)

	// Assuming that resetting takes time
	cpu.cyclesLeft = 8
}

func (cpu *CPU) GetCyclesLeft() uint8 {
	return cpu.cyclesLeft
}

func (cpu *CPU) GetPC() uint16 {
	return cpu.pc
}

func (cpu *CPU) SetPC(a uint16) uint16 {
	cpu.pc = a
	return cpu.pc
}

func (cpu *CPU) SetSP(a uint8) uint8 {
	cpu.sp = a
	return cpu.sp
}

func (cpu *CPU) GetSP() uint8 {
	return cpu.sp
}

func (cpu *CPU) SetA(a uint8) uint8 {
	cpu.a = a
	return cpu.a
}

func (cpu *CPU) GetA() uint8 {
	return cpu.a
}

func (cpu *CPU) SetX(a uint8) uint8 {
	cpu.x = a
	return cpu.x
}

func (cpu *CPU) GetX() uint8 {
	return cpu.x
}

func (cpu *CPU) SetY(a uint8) uint8 {
	cpu.y = a
	return cpu.y
}

func (cpu *CPU) GetY() uint8 {
	return cpu.y
}

func (cpu *CPU) GetStatusFlags() *flags.Flags {
	return cpu.p
}

func (cpu *CPU) Read(addr uint16) uint8 {
	return cpu.bus.Read(addr)
}

func (cpu *CPU) Read16(addr uint16) uint16 {
	return cpu.bus.Read16(addr)
}

func (cpu *CPU) Write(addr uint16, data uint8) {
	cpu.bus.Write(addr, data)
}

func (cpu *CPU) Write16(addr uint16, data uint16) {
	cpu.bus.Write16(addr, data)
}

func (cpu *CPU) PushToStack(data uint8) {
	cpu.bus.Write(0x0100+uint16(cpu.sp), data)
	cpu.sp--
}

func (cpu *CPU) PushToStack16(d uint16) {
	cpu.PushToStack(uint8((d >> 8) & 0x00FF))
	cpu.PushToStack(uint8(d & 0x00FF))
}

func (cpu *CPU) PullFromStack() uint8 {
	cpu.sp++
	return cpu.bus.Read(0x0100 + uint16(cpu.sp))
}

func (cpu *CPU) PullFromStack16() uint16 {
	low := uint16(cpu.PullFromStack())
	high := uint16(cpu.PullFromStack())

	return (high << 8) | low
}

func (cpu *CPU) AddCycles(c uint8) {
	cpu.cyclesLeft += c
}

// Clock - execute single clock cycle
func (cpu *CPU) Clock() {
	if cpu.cyclesLeft == 0 {

		// Check for scheduled interrupts
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

		// Read opcode
		opCode := cpu.bus.Read(cpu.pc)
		err := instructions.ExecuteInstruction(opCode, cpu)

		if err != nil {
			// TODO: what to do now?
			panic(err)
		}
	}

	// One cycle done
	cpu.cyclesLeft--
}

func (cpu *CPU) scheduleIRQ() {
	if !cpu.p.Get(flags.I) {
		cpu.isIRQScheduled = true
	}
}

func (cpu *CPU) ScheduleNMI() {
	cpu.isNMIScheduled = true
}

func (cpu *CPU) handleInterrupt(addr uint16) {
	// Push program counter to stack
	cpu.PushToStack16(cpu.pc)

	// Push status to stack
	// https://wiki.nesdev.com/w/index.php/Status_flags
	// Set 5 bit and unset 4 bit
	cpu.PushToStack(cpu.p.GetByte()&0b11001111 | 0b00100000)

	// Disable interrupts
	cpu.p.Set(flags.I, true)

	// Get new PC
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
		P:  cpu.p.GetByte(),
	}
}

func (cpu *CPU) Disassemble(addr uint16, state instructions.CPUState) (*instructions.InstructionDebugInfo, error) {
	opCode := cpu.bus.ReadDebug(addr)
	info, err := instructions.GetInstructionDebugInfo(opCode, state)

	return info, err

	// TODO: handle errors here
	//address, _ := addrMode.CalculateAddress(addr, cpu.x, cpu.y, func(addr uint16) uint8 {
	//	return cpu.bus.ReadDebug(addr)
	//})
	//info.Operand = addrMode.Format(address)
	//info.AddressingName = addrMode.Name
	//info.Size = addrMode.Size
}

func (cpu *CPU) Clone() CPU {
	cpu2 := *cpu
	return cpu2
}
