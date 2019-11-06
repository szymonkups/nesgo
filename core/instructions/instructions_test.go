package instructions_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/szymonkups/nesgo/core/flags"
	"github.com/szymonkups/nesgo/core/instructions"
	"testing"
)

func TestGetInstructionByName1(t *testing.T) {
	a := assert.New(t)

	i, ok := instructions.GetInstructionByName("ADC")
	a.True(ok, "GetInstructionByName should return true if exists")
	a.NotNil(i, "GetInstructionByName - instruction should be not nil")
}

func TestGetInstructionByName2(t *testing.T) {
	a := assert.New(t)

	i, ok := instructions.GetInstructionByName("adc")
	a.True(ok, "GetInstructionByName should return true if exists - lowercase")
	a.NotNil(i, "GetInstructionByName - instruction should be not nil - lowercase")
}

func TestGetInstructionByName3(t *testing.T) {
	a := assert.New(t)

	i, ok := instructions.GetInstructionByName("xyz")
	a.False(ok, "GetInstructionByName should return false if instruction not exists")
	a.Nil(i, "GetInstructionByName - instruction should be nil when not exists")
}

func TestGetInstructionByOpCode1(t *testing.T) {
	a := assert.New(t)

	i, ok := instructions.GetInstructionByOpCode(0x69)
	a.True(ok, "GetInstructionByOpCode should return true if exists")
	a.NotNil(i, "GetInstructionByOpCode - instruction should be not nil")
}

func TestGetInstructionByOpCode2(t *testing.T) {
	a := assert.New(t)

	i, ok := instructions.GetInstructionByOpCode(0x02)
	a.False(ok, "GetInstructionByOpCode should return false if not exists")
	a.Nil(i, "GetInstructionByOpCode - instruction should be nil")
}

func TestADC1(t *testing.T) {
	a := assert.New(t)
	p := new(flags.Flags)
	p.Set(flags.C, false)
	adc, _ := instructions.GetInstructionByName("ADC")
	addr := uint16(0xABCD)

	cpu := createMockCPU(a, 0, 0, 0x10, 0, 0, p, []ioOp{{
		kind:    "read",
		address: addr,
		data:    0x30,
	}})

	adc.Handler(cpu, addr, 0x69)
	a.Equal(uint8(0x40), cpu.a, "Sum should be calculated correctly")
	assertFlags(a, cpu.p, false, false, false, false, false, false)
}

func TestADC2(t *testing.T) {
	a := assert.New(t)
	p := new(flags.Flags)
	p.Set(flags.C, false)
	adc, _ := instructions.GetInstructionByName("ADC")
	addr := uint16(0xABCD)

	cpu := createMockCPU(a, 0, 0, 0xFF, 0, 0, p, []ioOp{{
		kind:    "read",
		address: addr,
		data:    0x30,
	}})

	adc.Handler(cpu, addr, 0x69)
	a.Equal(uint8(0x2f), cpu.a, "Sum should be calculated correctly")
	assertFlags(a, cpu.p, true, false, false, false, false, false)
}

func TestADC3(t *testing.T) {
	a := assert.New(t)
	p := new(flags.Flags)
	p.Set(flags.C, false)
	adc, _ := instructions.GetInstructionByName("ADC")
	addr := uint16(0xABCD)

	cpu := createMockCPU(a, 0, 0, 0xfd, 0, 0, p, []ioOp{{
		kind:    "read",
		address: addr,
		data:    0x3,
	}})

	adc.Handler(cpu, addr, 0x69)
	a.Equal(uint8(0), cpu.a, "Sum should be calculated correctly")
	assertFlags(a, cpu.p, true, true, false, false, false, false)
}

func TestADC4(t *testing.T) {
	a := assert.New(t)
	p := new(flags.Flags)
	p.Set(flags.C, false)
	adc, _ := instructions.GetInstructionByName("ADC")
	addr := uint16(0xABCD)

	cpu := createMockCPU(a, 0, 0, 0x7f, 0, 0, p, []ioOp{{
		kind:    "read",
		address: addr,
		data:    0x1,
	}})

	adc.Handler(cpu, addr, 0x69)
	a.Equal(uint8(0x80), cpu.a, "Sum should be calculated correctly")
	assertFlags(a, cpu.p, false, false, false, false, true, true)
}

func TestADC5(t *testing.T) {
	a := assert.New(t)
	p := new(flags.Flags)
	p.Set(flags.C, true)
	adc, _ := instructions.GetInstructionByName("ADC")
	addr := uint16(0xABCD)

	cpu := createMockCPU(a, 0, 0, 0x00, 0, 0, p, []ioOp{{
		kind:    "read",
		address: addr,
		data:    0x00,
	}})

	adc.Handler(cpu, addr, 0x69)
	a.Equal(uint8(0x1), cpu.a, "Sum should be calculated correctly")
	assertFlags(a, cpu.p, false, false, false, false, false, false)
}

type ioOp struct {
	kind    string
	address uint16
	data    uint8
}

type mockCPU struct {
	pc         uint16
	sp         uint8
	a          uint8
	x          uint8
	y          uint8
	p          *flags.Flags
	cyclesLeft uint8

	// Debug data
	ioOpIndex int
	ioOps     []ioOp
	asrt      *assert.Assertions
}

func assertFlags(a *assert.Assertions, p *flags.Flags, c, z, i, d, v, n bool) {
	a.Equal(c, p.Get(flags.C), "Wrong state of carry flag")
	a.Equal(z, p.Get(flags.Z), "Wrong state of zero flag")
	a.Equal(i, p.Get(flags.I), "Wrong state of interrupt flag")
	a.Equal(d, p.Get(flags.D), "Wrong state of decimal flag")
	a.Equal(v, p.Get(flags.V), "Wrong state of overflow flag")
	a.Equal(n, p.Get(flags.N), "Wrong state of negative flag")
}

func createMockCPU(asrt *assert.Assertions, pc uint16, sp, a, x, y uint8, p *flags.Flags, ioOps []ioOp) *mockCPU {
	cpu := new(mockCPU)
	cpu.asrt = asrt
	cpu.pc = pc
	cpu.sp = sp
	cpu.a = a
	cpu.x = x
	cpu.y = y
	cpu.p = p
	cpu.ioOps = ioOps

	return cpu
}

func (cpu *mockCPU) GetPC() uint16 {
	return cpu.pc
}

func (cpu *mockCPU) SetPC(a uint16) uint16 {
	cpu.pc = a
	return cpu.pc
}

func (cpu *mockCPU) SetSP(a uint8) uint8 {
	cpu.sp = a
	return cpu.sp
}

func (cpu *mockCPU) GetSP() uint8 {
	return cpu.sp
}

func (cpu *mockCPU) SetA(a uint8) uint8 {
	cpu.a = a
	return cpu.a
}

func (cpu *mockCPU) GetA() uint8 {
	return cpu.a
}

func (cpu *mockCPU) SetX(a uint8) uint8 {
	cpu.x = a
	return cpu.x
}

func (cpu *mockCPU) GetX() uint8 {
	return cpu.x
}

func (cpu *mockCPU) SetY(a uint8) uint8 {
	cpu.y = a
	return cpu.y
}

func (cpu *mockCPU) GetY() uint8 {
	return cpu.y
}

func (cpu *mockCPU) GetStatusFlags() *flags.Flags {
	return cpu.p
}

func (cpu *mockCPU) Read(addr uint16) uint8 {
	if !cpu.asrt.False(cpu.ioOpIndex >= len(cpu.ioOps), "Instruction wanted to read but test data went out of bounds") {
		return 0
	}

	op := cpu.ioOps[cpu.ioOpIndex]
	cpu.ioOpIndex++

	cpu.asrt.Equal(op.kind, "read", "Performing unexpected read operation")
	cpu.asrt.Equal(op.address, addr, "Performing read from unexpected address")

	return op.data
}

func (cpu *mockCPU) Read16(addr uint16) uint16 {
	low := uint16(cpu.Read(addr))
	high := uint16(cpu.Read(addr + 1))

	return (high << 8) | low
}

func (cpu *mockCPU) Write(addr uint16, data uint8) {
	op := cpu.ioOps[cpu.ioOpIndex]
	cpu.ioOpIndex++

	cpu.asrt.Equal(op.kind, "write", "Performing unexpected write operation")
	cpu.asrt.Equal(op.address, addr, "Performing write to unexpected address")
}

func (cpu *mockCPU) Write16(addr uint16, val uint16) {
	cpu.Write(addr, uint8(val&0x00FF))
	cpu.Write(addr+1, uint8((val>>8)&0x00FF))
}

func (cpu *mockCPU) PushToStack(data uint8) {
	cpu.Write(0x0100+uint16(cpu.sp), data)
	cpu.sp--
}

func (cpu *mockCPU) PushToStack16(d uint16) {
	cpu.PushToStack(uint8((d >> 8) & 0x00FF))
	cpu.PushToStack(uint8(d & 0x00FF))
}

func (cpu *mockCPU) PullFromStack() uint8 {
	cpu.sp++
	return cpu.Read(0x0100 + uint16(cpu.sp))
}

func (cpu *mockCPU) PullFromStack16() uint16 {
	low := uint16(cpu.PullFromStack())
	high := uint16(cpu.PullFromStack())

	return (high << 8) | low
}

func (cpu *mockCPU) AddCycles(c uint8) {
	cpu.cyclesLeft += c
}
