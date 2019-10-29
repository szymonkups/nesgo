package addressing_test

import (
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

	for i := range cases {
		_, ok := addressing.GetAddressingById(cases[i])

		if !ok {
			t.Errorf("could not find addressing with id %d", cases[i])
		}

		_, ok = addressing.GetAddressingById(999)

		if ok {
			t.Errorf("method GetAddressingById should return 'false' for incorrect addressing id")
		}

	}

}

func TestAccumulatorAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.AccumulatorAddressing)

	if addr.Name != "ACC" {
		t.Errorf("Wrong accumulator addressing name")
	}

	if addr.Size != 1 {
		t.Errorf("Wrong accumulator addressing size")
	}

	if addr.Format(0xFFFF) != "A" {
		t.Errorf("Wrong accumulator addressing formatting")
	}

	calculated, add := addr.CalculateAddress(0, 0, 0, dummyRead)

	if calculated != 0 {
		t.Errorf("Accumulator addressing should always return 0")
	}

	if add {
		t.Errorf("Accumulator addressing should not add cycles")
	}
}

func TestImpliedAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.ImpliedAddressing)

	if addr.Name != "IMP" {
		t.Errorf("Wrong implied addressing name")
	}

	if addr.Size != 1 {
		t.Errorf("Wrong implied addressing formatting")
	}

	if addr.Format(0xFFFF) != "" {
		t.Errorf("Wrong implied addressing name")
	}

	calculated, add := addr.CalculateAddress(0, 0, 0, dummyRead)

	if calculated != 0 {
		t.Errorf("Implied addressing should always return 0")
	}

	if add {
		t.Errorf("Implied addressing should not add cycles")
	}
}

func TestImmediateAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.ImmediateAddressing)

	if addr.Name != "IMM" {
		t.Errorf("Wrong immediate addressing name")
	}

	if addr.Size != 2 {
		t.Errorf("Wrong immediate addressing size")
	}

	formatted := addr.Format(0x20)
	if formatted != "#$20" {
		t.Errorf("Wrong immediate addressing formatting: %s", formatted)
	}

	calculated, add := addr.CalculateAddress(0x44, 0, 0, dummyRead)

	if calculated != 0x45 {
		t.Errorf("Immediate addressing should return next address to opcode's")
	}

	if add {
		t.Errorf("Immediate addressing should not add cycles")
	}
}

func TestZeroPageAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.ZeroPageAddressing)

	if addr.Name != "ZPA" {
		t.Errorf("Wrong zero page addressing name")
	}

	if addr.Size != 2 {
		t.Errorf("Wrong zero page addressing size")
	}

	formatted := addr.Format(0x20)
	if formatted != "$20" {
		t.Errorf("Wrong zero page addressing formatting: %s", formatted)
	}

	calculated, add := addr.CalculateAddress(0x44, 0, 0,
		createMockReadFunction(t, "zero page addressing", []uint8{0x40}, []uint16{0x44 + 1}))

	if calculated != 0x40 {
		t.Errorf("Zero page addressing should return address read from zero page")
	}

	if add {
		t.Errorf("Zero page addressing should not add cycles")
	}
}

func TestZeroPageXAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.ZeroPageXAddressing)

	if addr.Name != "ZPX" {
		t.Errorf("Wrong zero page X addressing name")
	}

	if addr.Size != 2 {
		t.Errorf("Wrong zero page X addressing size")
	}

	formatted := addr.Format(0x20)
	if formatted != "$20,X" {
		t.Errorf("Wrong zero page X addressing formatting: %s", formatted)
	}

	calculated, add := addr.CalculateAddress(0x44, 0x10, 0,
		createMockReadFunction(t, "zero page X", []uint8{0x20}, []uint16{0x44 + 1}))

	if calculated != 0x20+0x10 {
		t.Errorf("Zero page X addressing should add X to the addres from memory")
	}

	if add {
		t.Errorf("Zero page X addressing should not add cycles")
	}

	calculated, add = addr.CalculateAddress(0x44, 0x10, 0, createMockReadFunction(t, "zero page X", []uint8{0xFF}, []uint16{0x44 + 1}))

	if calculated != 0x0F {
		t.Errorf("Zero page X addressing should wrap around 0xFF %x", calculated)
	}

	if add {
		t.Errorf("Zero page X addressing should not add cycles when wrapping around")
	}
}

func TestZeroPageYAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.ZeroPageYAddressing)

	if addr.Name != "ZPY" {
		t.Errorf("Wrong zero page Y addressing name")
	}

	if addr.Size != 2 {
		t.Errorf("Wrong zero page Y addressing size")
	}

	formatted := addr.Format(0x20)
	if formatted != "$20,Y" {
		t.Errorf("Wrong zero page Y addressing formatting: %s", formatted)
	}

	calculated, add := addr.CalculateAddress(0x44, 0, 0x10,
		createMockReadFunction(t, "zero page Y", []uint8{0x20}, []uint16{0x44 + 1}))

	if calculated != 0x20+0x10 {
		t.Errorf("Zero page Y addressing should add Y to the addres from memory")
	}

	if add {
		t.Errorf("Zero page Y addressing should not add cycles")
	}

	calculated, add = addr.CalculateAddress(0x44, 0, 0x10,
		createMockReadFunction(t, "zero page Y", []uint8{0xFF}, []uint16{0x44 + 1}))

	if calculated != 0x0F {
		t.Errorf("Zero page Y addressing should wrap around 0xFF %x", calculated)
	}

	if add {
		t.Errorf("Zero page Y addressing should not add cycles when wrapping around")
	}
}

func TestRelativeAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.RelativeAddressing)

	if addr.Name != "REL" {
		t.Errorf("Wrong relative addressing name")
	}

	if addr.Size != 2 {
		t.Errorf("Wrong relative addressing size")
	}

	formatted := addr.Format(0x4)
	if formatted != "$04" {
		t.Errorf("Wrong relative addressing formatting: %s", formatted)
	}

	calculated, add := addr.CalculateAddress(0x44, 0, 0x10,
		createMockReadFunction(t, "relative addressing", []uint8{0x20}, []uint16{0x44 + 1}))

	if calculated != 0x44+0x2+0x20 {
		t.Errorf("Relative addressing should add positive number to PC")
	}

	if add {
		t.Errorf("Relative addressing should not add cycles")
	}

	calculated, _ = addr.CalculateAddress(0x44, 0, 0x10,
		// 0b11111100 => 0b00000011 + 1 => 3+1 => 4 => -4 on NES
		createMockReadFunction(t, "relative addressing", []uint8{0b11111100}, []uint16{0x44 + 1}))

	if calculated != 0x44+0x2-0x4 {
		t.Errorf("Relative addressing should subtract negative number from PC, returned 0x%04X", calculated)
	}

	calculated, _ = addr.CalculateAddress(0, 0, 0x10,
		// 0b11111100 => 0b00000011 + 1 => 3+1 => 4 => -4 on NES
		createMockReadFunction(t, "relative addressing", []uint8{0b11111100}, []uint16{0 + 1}))

	if calculated != 0xFFFE {
		t.Errorf("Relative addressing should wrap around on 0xFFFF when subtracting from PC, returned 0x%04X", calculated)
	}

	calculated, _ = addr.CalculateAddress(0xFFFE, 0, 0x10,
		createMockReadFunction(t, "relative addressing", []uint8{0x4}, []uint16{0xFFFE + 1}))

	if calculated != 0x4 {
		t.Errorf("Relative addressing should wrap around on 0xFFFF when adding to PC, returned 0x%04X", calculated)
	}
}

func TestAbsoluteAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.AbsoluteAddressing)

	if addr.Name != "ABS" {
		t.Errorf("Wrong absolute addressing name")
	}

	if addr.Size != 3 {
		t.Errorf("Wrong absolute addressing size")
	}

	formatted := addr.Format(0xFA)
	if formatted != "$00FA" {
		t.Errorf("Wrong absolute addressing formatting: %s", formatted)
	}

	calculated, add := addr.CalculateAddress(0x44, 0, 0x10,
		createMockReadFunction(t, "relative addressing", []uint8{0xFF, 0xCC}, []uint16{0x45, 0x46}))

	if calculated != 0xCCFF {
		t.Errorf("Absolute addressing should return correct value, got %04X", calculated)
	}

	if add {
		t.Errorf("Absolute addressing should not add cycles")
	}
}

func TestAbsoluteXAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.AbsoluteXAddressing)

	if addr.Name != "ABX" {
		t.Errorf("Wrong absolute X addressing name")
	}

	if addr.Size != 3 {
		t.Errorf("Wrong absolute X addressing size")
	}

	formatted := addr.Format(0xFA)
	if formatted != "$00FA,X" {
		t.Errorf("Wrong absolute X addressing formatting: %s", formatted)
	}

	calculated, add := addr.CalculateAddress(0x44, 0x10, 0,
		createMockReadFunction(t, "relative addressing", []uint8{0xAA, 0xBB}, []uint16{0x45, 0x46}))

	if calculated != 0xBBAA+0x10 {
		t.Errorf("Absolute addressing X should return correct value, got %04X", calculated)
	}

	if add {
		t.Errorf("Absolute addressing X should not add cycles when not crossing pages")
	}

	calculated, add = addr.CalculateAddress(0x44, 0x10, 0,
		createMockReadFunction(t, "relative addressing", []uint8{0xFF, 0xAA}, []uint16{0x45, 0x46}))

	if calculated != 0xAAFF+0x10 {
		t.Errorf("Absolute addressing X should return correct value when crossing pages, got %04X", calculated)
	}

	if !add {
		t.Errorf("Absolute addressing X should add cycles when crossing pages")
	}
}

func TestAbsoluteYAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.AbsoluteYAddressing)

	if addr.Name != "ABY" {
		t.Errorf("Wrong absolute Y addressing name")
	}

	if addr.Size != 3 {
		t.Errorf("Wrong absolute Y addressing size")
	}

	formatted := addr.Format(0xFA)
	if formatted != "$00FA,Y" {
		t.Errorf("Wrong absolute Y addressing formatting: %s", formatted)
	}

	calculated, add := addr.CalculateAddress(0x44, 0, 0x10,
		createMockReadFunction(t, "relative addressing", []uint8{0xAA, 0xBB}, []uint16{0x45, 0x46}))

	if calculated != 0xBBAA+0x10 {
		t.Errorf("Absolute addressing Y should return correct value, got %04X", calculated)
	}

	if add {
		t.Errorf("Absolute addressing Y should not add cycles when not crossing pages")
	}

	calculated, add = addr.CalculateAddress(0x44, 0, 0x10,
		createMockReadFunction(t, "relative addressing", []uint8{0xFF, 0xAA}, []uint16{0x45, 0x46}))

	if calculated != 0xAAFF+0x10 {
		t.Errorf("Absolute addressing Y should return correct value when crossing pages, got %04X", calculated)
	}

	if !add {
		t.Errorf("Absolute addressing Y should add cycles when crossing pages")
	}
}

func TestIndirectAddressing(t *testing.T) {
	addr, _ := addressing.GetAddressingById(addressing.IndirectAddressing)

	if addr.Name != "IND" {
		t.Errorf("Wrong indirect addressing name")
	}

	if addr.Size != 3 {
		t.Errorf("Wrong indirect addressing size")
	}

	formatted := addr.Format(0xFA)
	if formatted != "($00FA)" {
		t.Errorf("Wrong indirect addressing formatting: %s", formatted)
	}

	calculated, add := addr.CalculateAddress(0x44, 0, 0,
		createMockReadFunction(t, "indirect addressing", []uint8{0xAA, 0xBB, 0xFF, 0x21}, []uint16{0x45, 0x46, 0xBBAA, 0xBBAB}))

	if calculated != 0x21FF {
		t.Errorf("Indirect addressing should return correct value, got $%04X", calculated)
	}

	if add {
		t.Errorf("Indirect addresing should always return false")
	}

}

// Creates mock read function to be used with CalculateAddress. Name will be used in the error reports to indicate
// which of the methods failed. Data is a slice of bytes that will be returned in each call of the read function.
// Addresses is a slice of addresses that should be used by addressing method in each call.
func createMockReadFunction(t *testing.T, name string, data []uint8, addresses []uint16) addressing.ReadFunction {
	i := 0

	return func(addr uint16) uint8 {
		if i < len(data) && i < len(addresses) {
			if addresses[i] != addr {
				t.Errorf("Wrong address read by %s. Expected: $%04X, got: $%04X", name, addresses[i], addr)
			}

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
