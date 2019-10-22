package display_objects

import (
	"fmt"
	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/ui/engine"
)

type Debugger struct {
	CPU *core.CPU
	PPU *core.PPU
	CRT *core.Cartridge
}

func (d *Debugger) Draw(e *engine.UIEngine) error {

	// DRAW HEADER WITH FPS
	e.DrawRect(0, 0, 256*2, 240*2, 0xFF, 0, 0, 0xFF)
	e.FillRect(0, 0, 256*2, 8, 0xFF, 0, 0, 0xFF)
	err := e.DrawText("NES CPU AND PPU DEBUGGER", 1, 0, 0, 0, 0, 0xFF)

	if err != nil {
		return err
	}

	err = e.DrawText(fmt.Sprintf("FPS:%3d", e.FPS), 455, 0, 0, 0, 0, 0xFF)

	if err != nil {
		return err
	}
	//
	// Draw registers
	reg := d.CPU.GetDebugInfo()
	drawRegister16(e, "PC", reg.PC, 2, 9)
	drawRegister8(e, "SP", reg.SP, 63, 9)
	drawRegister8(e, "A", reg.A, 108, 9)
	drawRegister8(e, "X", reg.X, 145, 9)
	drawRegister8(e, "Y", reg.Y, 182, 9)
	drawFlags8(e, "NV--DIZC", reg.P, 219, 9)

	//// Draw current memory range
	drawAssembly(d.CPU, e, 2, 21, reg.PC)

	chr := d.CRT.GetCHRMem()

	for y := 0; y < 20; y++ {
		for x := 0; x < 16; x++ {
			d.drawSinglePattern(e, x+(y*16), chr, 10+(int32(x)*8), 40+(int32(y)*8))
		}
	}

	return nil
}

func (d *Debugger) drawSinglePattern(e *engine.UIEngine, n int, chr []uint8, x, y int32) {
	offset := int32(n * 16)

	for i := 0; i < 8; i++ {
		first := chr[offset+int32(i)]
		second := chr[offset+int32(i)+8]

		for j := 0; j < 8; j++ {
			b1 := getBit(first, 7-uint8(j))
			b2 := getBit(second, 7-uint8(j))

			clr := (b2 << 1) | b1
			if clr != 0 {
				//d.PPU.GetColorFromPalette(0, clr)
				color := d.PPU.GetColorFromPalette(0, clr)

				e.DrawPixel(x+int32(j), y+int32(i), color.R, color.G, color.B, 0xFF)
			}

		}

	}
}

func getBit(i uint8, n uint8) uint8 {
	return (i >> n) & 1
}

func (d *Debugger) GetChildren() []engine.Displayable {
	return nil
}

func drawRegister16(e *engine.UIEngine, name string, value uint16, x, y int32) {
	e.DrawRect(x, y, 60, 10, 0xFF, 0, 0, 0xFF)
	e.DrawText(name, x+1, y+1, 0xFF, 0, 0, 0xAA)
	e.DrawText(fmt.Sprintf("$%04X", value), x+19, y+1, 0xff, 0xff, 0xff, 0xff)
}

func drawRegister8(e *engine.UIEngine, name string, value uint8, x, y int32) {
	s := int32(0)

	if len(name) == 1 {
		s = 8
	}
	e.DrawRect(x, y, 44-s, 10, 0xFF, 0, 0, 0xFF)
	e.DrawText(name, x+1, y+1, 0xFF, 0, 0, 0xAA)
	e.DrawText(fmt.Sprintf("$%02X", value), x+19-s, y+1, 0xff, 0xff, 0xff, 0xff)
}

func drawFlags8(e *engine.UIEngine, registers string, value uint8, x, y int32) {
	e.DrawRect(x, y, 75, 10, 0xFF, 0, 0, 0xFF)

	ty := y + 1
	for i, l := range registers {
		tx := x + 6 + (int32(i) * 8)
		text := string(l)
		isByteActive := value&(1<<(7-i)) > 0

		if isByteActive {
			e.DrawText(text, tx, ty, 0xFF, 0xFF, 0xFF, 0xFF)
		} else {
			e.DrawText(text, tx, ty, 0xFF, 0, 0, 0x66)
		}

	}
}

func drawAssembly(cpu *core.CPU, e *engine.UIEngine, x, y int32, pc uint16) {
	addr := int32(pc)
	assembly, ok := cpu.Disassemble(pc)

	e.FillRect(x, y, 292, 8, 0xFF, 0, 0, 0xFF)

	if !ok {
		e.DrawText(fmt.Sprintf("$%04X   #!UNKNOWN OPCODE!#", addr), x, y, 0xFF, 0xFF, 0xFF, 0xFF)
	} else {
		e.DrawText(fmt.Sprintf("$%04X   %s", addr, assembly), x, y, 0xFF, 0xFF, 0xFF, 0xFF)
	}
}
