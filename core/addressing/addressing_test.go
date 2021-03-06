package addressing_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/szymonkups/nesgo/core/addressing"
	"testing"
)

func TestGetAddressingById(t *testing.T) {
	cases := []int{
		addressing.AccumulatorAddressing,
		addressing.ImpliedAddressing,
		addressing.ImmediateAddressing,
		addressing.ZeroPageAddressing,
		addressing.ZeroPageXAddressing,
		addressing.ZeroPageYAddressing,
		addressing.RelativeAddressing,
		addressing.AbsoluteAddressing,
		addressing.AbsoluteXAddressing,
		addressing.AbsoluteYAddressing,
		addressing.IndirectAddressing,
		addressing.IndirectXAddressing,
		addressing.IndirectYAddressing,
	}

	a := assert.New(t)

	for i := range cases {
		_, ok := addressing.GetAddressingById(cases[i])
		a.True(ok, "GetAddressingById couldn't return addressing with id %d.", cases[i])

		_, ok = addressing.GetAddressingById(999)
		a.False(ok, "GetAddressingById should not return addressing for index that not exists", cases[i])
	}

}

func TestAccumulatorAddressing(t *testing.T) {
	a := assert.New(t)
	addr, _ := addressing.GetAddressingById(addressing.AccumulatorAddressing)

	a.Equal("ACC", addr.Name, "Wrong accumulator addressing name")
	a.Equal(uint8(1), addr.Size, "Wrong accumulator addressing size")
	a.Equal("A", addr.Format(0xFFFF), "Wrong accumulator addressing formatting")

	calculated, add := addr.CalculateAddress(0, 0, 0, dummyRead)
	a.Equal(uint16(0), calculated, "Accumulator addressing should always return 0")
	a.False(add, "Accumulator addressing should not add cycles")
}

func TestImpliedAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.ImpliedAddressing)
	a.Equal("IMP", addr.Name, "Wrong implied addressing name")
	a.Equal(uint8(1), addr.Size, "Wrong implied addressing size")
	a.Equal("", addr.Format(0xFFFF), "Wrong implied addressing formatting")

	calculated, add := addr.CalculateAddress(0, 0, 0, dummyRead)

	a.Equal(uint16(0), calculated, "Implied addressing should always return 0")
	a.False(add, "Implied addressing should not add cycles")
}

func TestImmediateAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.ImmediateAddressing)
	a.Equal("IMM", addr.Name, "Wrong immediate addressing name")
	a.Equal(uint8(2), addr.Size, "Wrong immediate addressing size")
	a.Equal("#$20", addr.Format(0x20), "Wrong immediate addressing formatting")

	pc := uint16(0x44)
	calculated, add := addr.CalculateAddress(pc, 0, 0, dummyRead)
	a.Equal(pc+1, calculated, "Immediate addressing should return address next to current PC")
	a.False(add, "Immediate addressing should not add cycles")
}

func TestZeroPageAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.ZeroPageAddressing)
	a.Equal("ZPA", addr.Name, "Wrong zero page addressing name")
	a.Equal(uint8(2), addr.Size, "Wrong zero page addressing size")
	a.Equal("$20", addr.Format(0x20))

	pc := uint16(0x44)
	calculated, add := addr.CalculateAddress(pc, 0, 0,
		createMockReadFunction(t, "zero page addressing", []uint8{0x40}, []uint16{pc + 1}))

	a.Equal(uint16(0x40), calculated, "Zero page addressing should return correct address")
	a.False(add, "Zero page addressing should not add cycles")
}

func TestZeroPageXAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.ZeroPageXAddressing)
	a.Equal("ZPX", addr.Name, "Wrong zero page X addressing name")
	a.Equal(uint8(2), addr.Size, "Wrong zero page X addressing size")
	a.Equal("$20,X", addr.Format(0x20), "Wrong zero page X addressing formatting")

	// No wrapping
	pc := uint16(0x44)
	x := uint8(0x10)
	d1 := uint8(0x20)
	mockRead := createMockReadFunction(t, "zero page X", []uint8{d1}, []uint16{pc + 1})
	calculated, add := addr.CalculateAddress(pc, x, 0, mockRead)
	a.Equal(uint16(d1+x), calculated, "Zero page X addressing should add X to the address from memory")
	a.False(add, "Zero page X addressing should not add cycles")

	// Wrapping around
	pc = uint16(0x44)
	x = uint8(0x10)
	d1 = uint8(0xFF)
	mockRead = createMockReadFunction(t, "zero page X", []uint8{d1}, []uint16{pc + 1})
	calculated, add = addr.CalculateAddress(pc, x, 0, mockRead)
	a.Equal(uint16(0x0F), calculated, "Zero page X addressing should add X to the address from memory")
	a.False(add, "Zero page X addressing should not add cycles when wrapping around")
}

func TestZeroPageYAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.ZeroPageYAddressing)
	a.Equal("ZPY", addr.Name, "Wrong zero page Y addressing name")
	a.Equal(uint8(2), addr.Size, "Wrong zero page Y addressing size")
	a.Equal("$20,Y", addr.Format(0x20), "Wrong zero page Y addressing formatting")

	// No wrapping
	pc := uint16(44)
	y := uint8(0x10)
	d1 := uint8(0x20)
	mockRead := createMockReadFunction(t, "zero page Y", []uint8{d1}, []uint16{pc + 1})
	calculated, add := addr.CalculateAddress(pc, 0, y, mockRead)
	a.Equal(uint16(d1+y), calculated, "Zero page Y addressing should add Y to the address from memory")
	a.False(add, "Zero page Y addressing should not add cycles")

	// Wrapping around
	pc = uint16(0xCA)
	y = uint8(0x10)
	d1 = uint8(0xFF)
	mockRead = createMockReadFunction(t, "zero page Y", []uint8{d1}, []uint16{pc + 1})
	calculated, add = addr.CalculateAddress(pc, 0, y, mockRead)
	a.Equal(uint16(0x0F), calculated, "Zero page Y addressing should wrap around 0xFF")
	a.False(add, "Zero page Y addressing should not add cycles when wrapping around")
}

func TestRelativeAddressing(t *testing.T) {
	a := assert.New(t)
	addr, _ := addressing.GetAddressingById(addressing.RelativeAddressing)
	a.Equal("REL", addr.Name, "Wrong relative addressing name")
	a.Equal(uint8(2), addr.Size, "Wrong relative addressing size")
	a.Equal("$04", addr.Format(0x04), "Wrong relative addressing formattin")

	// Adding
	pc := uint16(0x44)
	d1 := uint8(0x20)
	mockRead := createMockReadFunction(t, "relative addressing - adding", []uint8{d1}, []uint16{pc + 1})
	calculated, add := addr.CalculateAddress(pc, 0, 0, mockRead)
	a.Equal(pc+2+uint16(d1), calculated, "Relative addressing should add positive number to PC")
	a.False(add, "Relative addressing should not add cycles - adding")

	// Subtracting
	pc = uint16(0x44)
	// 0b11111100 => 0b00000011 + 1 => 3+1 => 4 => -4 on NES
	d1 = 0b11111100
	mockRead = createMockReadFunction(t, "relative addressing - subtracting", []uint8{d1}, []uint16{pc + 1})
	calculated, add = addr.CalculateAddress(pc, 0, 0, mockRead)
	a.Equal(pc+2-0x4, calculated, "Relative addressing should subtract negative number from PC")
	a.False(add, "Relative addressing should not add cycles - subtracting")

	// Adding - wrap
	pc = uint16(0xFFFE)
	d1 = uint8(0x4)
	mockRead = createMockReadFunction(t, "relative addressing - adding and wrapping", []uint8{d1}, []uint16{pc + 1})
	calculated, _ = addr.CalculateAddress(0xFFFE, 0, 0, mockRead)
	a.Equal(uint16(0x4), calculated, "Relative addressing should wrap around on 0xFFFF when adding to PC")
	a.False(add, "Relative addressing should not add cycles - adding and wrapping")

	// Subtracting - wrap
	pc = uint16(0)
	// 0b11111100 => 0b00000011 + 1 => 3+1 => 4 => -4 on NES
	d1 = 0b11111100
	mockRead = createMockReadFunction(t, "relative addressing - subtracting and wrapping", []uint8{d1}, []uint16{0 + 1})
	calculated, add = addr.CalculateAddress(pc, 0, 0, mockRead)
	a.Equal(pc+2-0x4, calculated, "Relative addressing should wrap around on 0xFFFF when subtracting from PC")
	a.False(add, "Relative addressing should not add cycles - subtracting and wrapping")
}

func TestAbsoluteAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.AbsoluteAddressing)
	a.Equal("ABS", addr.Name, "Wrong absolute addressing name")
	a.Equal(uint8(3), addr.Size, "Wrong absolute addressing size")
	a.Equal("$00FA", addr.Format(0xFA), "Wrong absolute addressing formatting")

	pc := uint16(0x44)
	mockRead := createMockReadFunction(t, "relative addressing", []uint8{0xFF, 0xCC}, []uint16{pc + 1, pc + 2})
	calculated, add := addr.CalculateAddress(pc, 0, 0, mockRead)
	a.Equal(uint16(0xCCFF), calculated, "Absolute addressing should return correct value")
	a.False(add, "Absolute addressing should not add cycles")
}

func TestAbsoluteXAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.AbsoluteXAddressing)
	a.Equal("ABX", addr.Name, "Wrong absolute X addressing name")
	a.Equal(uint8(3), addr.Size, "Wrong absolute X addressing size")
	a.Equal("$00FA,X", addr.Format(0xFA), "Wrong absolute X addressing formatting")

	// No page crossed
	pc := uint16(0x44)
	x := uint8(0x10)
	mockRead := createMockReadFunction(t, "relative addressing", []uint8{0xAA, 0xBB}, []uint16{pc + 1, pc + 2})
	calculated, add := addr.CalculateAddress(pc, x, 0, mockRead)
	a.Equal(uint16(0xBBAA)+uint16(x), calculated, "Absolute addressing X should return correct value")
	a.False(add, "Absolute addressing X should not add cycles when not crossing pages")

	// Page crossed
	pc = uint16(0x44)
	x = uint8(0x10)
	mockRead = createMockReadFunction(t, "relative addressing", []uint8{0xFF, 0xAA}, []uint16{pc + 1, pc + 2})
	calculated, add = addr.CalculateAddress(pc, x, 0, mockRead)
	a.Equal(uint16(0xAAFF)+uint16(x), calculated, "Absolute addressing X should return correct value when crossing pages")
	a.True(add, "Absolute addressing X should add cycles when crossing pages")
}

func TestAbsoluteYAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.AbsoluteYAddressing)
	a.Equal("ABY", addr.Name, "Wrong absolute Y addressing name")
	a.Equal(uint8(3), addr.Size, "Wrong absolute Y addressing size")
	a.Equal("$00FA,Y", addr.Format(0xFA), "Wrong absolute Y addressing formatting")

	// No page crossing
	pc := uint16(0x44)
	y := uint8(0x10)
	mockRead := createMockReadFunction(t, "relative addressing", []uint8{0xAA, 0xBB}, []uint16{pc + 1, pc + 2})
	calculated, add := addr.CalculateAddress(pc, 0, y, mockRead)
	a.Equal(uint16(0xBBAA)+uint16(y), calculated, "Absolute addressing Y should return correct value")
	a.False(add, "Absolute addressing Y should not add cycles when not crossing pages")

	// Page crossing
	pc = uint16(0x44)
	y = uint8(0x10)
	mockRead = createMockReadFunction(t, "relative addressing", []uint8{0xFF, 0xAA}, []uint16{pc + 1, pc + 2})
	calculated, add = addr.CalculateAddress(pc, 0, y, mockRead)
	a.Equal(uint16(0xAAFF)+uint16(y), calculated, "Absolute addressing Y should return correct value when crossing pages")
	a.True(add, "Absolute addressing Y should add cycles when crossing pages")
}

func TestIndirectAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.IndirectAddressing)
	a.Equal("IND", addr.Name, "Wrong indirect addressing name")
	a.Equal(uint8(3), addr.Size, "Wrong indirect addressing size")
	a.Equal("($00FA)", addr.Format(0xFA), "Wrong indirect addressing formatting")

	// Normal operation
	pc := uint16(0x44)
	mockRead := createMockReadFunction(t, "indirect addressing", []uint8{0xAA, 0xBB, 0xFF, 0x21}, []uint16{pc + 1, pc + 2, 0xBBAA, 0xBBAB})
	calculated, add := addr.CalculateAddress(pc, 0, 0, mockRead)
	a.Equal(uint16(0x21FF), calculated, "Indirect addressing should return correct value")
	a.False(add, "Indirect addressing should not add cycle")

	// Check the 6502 bug with address wrapping on LSB when reading from xxFF
	pc = uint16(0x44)
	mockRead = createMockReadFunction(t, "indirect addressing - wrap bug", []uint8{0xFF, 0xBB, 0xFF, 0x21}, []uint16{pc + 1, pc + 2, 0xBBFF, 0xBB00})
	calculated, add = addr.CalculateAddress(pc, 0, 0, mockRead)
	a.Equal(uint16(0x21FF), calculated, "Indirect addressing should wrap around LSB (6502 bug)")
	a.False(add, "Indirect addressing should not add cycle when LSB bug appears")
}

func TestIndirectXAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.IndirectXAddressing)
	a.Equal("INX", addr.Name, "Wrong indirect X addressing name")
	a.Equal(uint8(2), addr.Size, "Wrong indirect X addressing size")
	a.Equal("($FA,X)", addr.Format(0xFA), "Wrong indirect X addressing formatting")

	pc := uint16(0x44)
	x := uint8(0x12)
	d1 := uint8(0x11)
	mockRead := createMockReadFunction(t, "indirect X addressing", []uint8{d1, 0xAA, 0xBB}, []uint16{pc + 1, uint16(d1 + x), uint16(d1 + x + 1)})
	calculated, add := addr.CalculateAddress(pc, x, 0, mockRead)
	a.Equal(uint16(0xBBAA), calculated, "Indirect X addressing should return correct value")
	a.False(add, "Indirect X addressing should not add cycles")
}

func TestIndirectYAddressing(t *testing.T) {
	a := assert.New(t)

	addr, _ := addressing.GetAddressingById(addressing.IndirectYAddressing)
	a.Equal("INY", addr.Name, "Wrong indirect Y addressing name")
	a.Equal(uint8(2), addr.Size, "Wrong indirect Y addressing size")
	a.Equal("($FA),Y", addr.Format(0xFA), "Wrong indirect Y addressing formatting")

	// Page not crossed
	pc := uint16(0x44)
	y := uint8(0x12)
	d1 := uint8(0x11)
	mockRead := createMockReadFunction(t, "indirect Y addressing", []uint8{d1, 0xAA, 0xBB}, []uint16{pc + 1, uint16(d1), uint16(d1 + 1)})
	calculated, add := addr.CalculateAddress(pc, 0, y, mockRead)
	a.Equal(0xBBAA+uint16(y), calculated, "Indirect Y addressing should return correct value")
	a.False(add, "Indirect Y addressing should return false when page is not crossed")

	// Page crossed
	pc = uint16(0x44)
	y = uint8(0x12)
	d1 = uint8(0x11)
	readMock := createMockReadFunction(t, "indirect Y addressing", []uint8{d1, 0xFF, 0xBB}, []uint16{pc + 1, uint16(d1), uint16(d1 + 1)})
	calculated, add = addr.CalculateAddress(pc, 0, y, readMock)
	a.Equal(0xBBFF+uint16(y), calculated, "Indirect Y addressing should return correct value when crossing pages")
	a.True(add, "Indirect Y addressing should return true when page is crossed")
}

// Creates mock read function to be used with CalculateAddress. Name will be used in the error reports to indicate
// which of the methods failed. Data is a slice of bytes that will be returned in each call of the read function.
// Addresses is a slice of addresses that should be used by addressing method in each call.
func createMockReadFunction(t *testing.T, name string, data []uint8, addresses []uint16) addressing.ReadFunction {
	a := assert.New(t)
	i := 0

	return func(addr uint16) uint8 {
		if i < len(data) && i < len(addresses) {
			a.Equalf(addresses[i], addr, "Wrong address read by %s", name)

			ret := data[i]
			i++

			return ret
		}

		t.Errorf("Not enough data provided to test functions of %s, there were %d calls", name, i+1)
		return 0
	}
}

func dummyRead(addr uint16) uint8 {
	return 0
}
