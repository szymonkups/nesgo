package display_objects

import (
	"bytes"
	"fmt"
	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/ui/engine"
	"text/tabwriter"
)

type Debugger struct {
	CPU *core.CPU
	PPU *core.PPU
	CRT *core.Cartridge

	paletteId uint8
}

func (d *Debugger) SetPaletteId(newId uint8) {
	d.paletteId = newId
}
func (d *Debugger) Draw(e *engine.UIEngine) error {

	// DRAW HEADER WITH FPS
	//e.DrawRect(0, 0, 256*2, 240*2, 0xFF, 0, 0, 0xFF)
	e.FillRect(0, 0, 256*2, 8, 0xFF, 0, 0, 0xFF)
	err := e.DrawText("NES CPU AND PPU DEBUGGER", 1, 0, 0, 0, 0, 0xFF)

	if err != nil {
		return err
	}

	err = e.DrawText(fmt.Sprintf("FPS:%3d", e.FPS), 455, 0, 0, 0, 0, 0xFF)

	if err != nil {
		return err
	}

	// Draw registers
	reg := d.CPU.GetDebugInfo()
	drawRegister16(e, "PC", reg.PC, 2, 9)
	drawRegister8(e, "SP", reg.SP, 63, 9)
	drawRegister8(e, "A", reg.A, 108, 9)
	drawRegister8(e, "X", reg.X, 145, 9)
	drawRegister8(e, "Y", reg.Y, 182, 9)
	drawFlags8(e, "NV--DIZC", reg.P, 219, 9)

	// Draw current memory range
	d.drawAssembly(e, 2, 21, reg.PC)
	//
	// Draw palettes
	d.drawPalettes(e, 0, 130)

	d.PPU.DrawPatternTable(0, func(x, y uint16, pixel uint8) {
		color := d.PPU.GetColorFromPalette(d.paletteId, pixel)
		if pixel == 0 {
			return
		}
		e.DrawPixel(100+int32(x), 132+int32(y), color.R, color.G, color.B, 0xFF)
	})

	d.PPU.DrawPatternTable(1, func(x, y uint16, pixel uint8) {
		color := d.PPU.GetColorFromPalette(d.paletteId, pixel)
		if pixel == 0 {
			return
		}
		e.DrawPixel(230+int32(x), 132+int32(y), color.R, color.G, color.B, 0xFF)
	})

	return nil
}

func (d *Debugger) drawPalettes(e *engine.UIEngine, x, y int32) {
	e.FillRect(x, y, 98, 9, 0xFF, 0, 0, 0xFF)
	e.DrawRect(x, y, 360, 132, 0xFF, 0, 0, 0xFF)
	e.DrawText("PPU", x+1, y+1, 0, 0, 0, 0xFF)

	e.DrawText("BG COLOR", x+1, y+10, 0xFF, 0, 0, 0xFF)
	cl := d.PPU.GetUniversalBGColor()
	e.FillRect(x+90, y+10, 8, 8, cl.R, cl.G, cl.B, 0xFF)

	d.drawPalette(e, "BG #0", 0, x, y)
	d.drawPalette(e, "BG #1", 1, x, y+10)
	d.drawPalette(e, "BG #2", 2, x, y+20)
	d.drawPalette(e, "BG #3", 3, x, y+30)

	d.drawPalette(e, "SP #0", 4, x, y+40)
	d.drawPalette(e, "SP #1", 5, x, y+50)
	d.drawPalette(e, "SP #2", 6, x, y+60)
	d.drawPalette(e, "SP #3", 7, x, y+70)

	e.DrawText(">", x+50, y+20+(int32(d.paletteId)*10), 0xFF, 0xFF, 0xFF, 0xFF)
}

func (d *Debugger) drawPalette(e *engine.UIEngine, name string, index uint8, x, y int32) {
	e.DrawText(name, x+1, y+20, 0xFF, 0, 0, 0xFF)

	cl := d.PPU.GetColorFromPalette(index, 4)
	e.FillRect(x+90, y+20, 8, 8, cl.R, cl.G, cl.B, 0xFF)
	cl = d.PPU.GetColorFromPalette(index, 3)
	e.FillRect(x+80, y+20, 8, 8, cl.R, cl.G, cl.B, 0xFF)
	cl = d.PPU.GetColorFromPalette(index, 2)
	e.FillRect(x+70, y+20, 8, 8, cl.R, cl.G, cl.B, 0xFF)
	cl = d.PPU.GetColorFromPalette(index, 1)
	e.FillRect(x+60, y+20, 8, 8, cl.R, cl.G, cl.B, 0xFF)
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

func (d *Debugger) drawAssembly(e *engine.UIEngine, x, y int32, startAddr uint16) {
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 31, 1, 0, ' ', 0)
	e.FillRect(x, y, 292, 8, 0xFF, 0, 0, 0xFF)
	addr := startAddr

	clone := d.CPU.Clone()

	for i := 0; i < 10; i++ {
		info, err := d.CPU.Disassemble(addr, &clone)

		if err != nil {
			// TODO: error handling
			continue
		}

		clone.SetPC(clone.GetPC() + uint16(info.Size))

		buf.Truncate(0)
		fmt.Fprintf(w, "$%04X %s %s\t{%s}", addr, info.InstructionName, info.Operand, info.AddressingName)
		w.Flush()

		e.DrawText(buf.String(), x, y+(10*int32(i)), 0xFF, 0xFF, 0xFF, 0xFF)

		addr += uint16(info.Size)
	}
}
