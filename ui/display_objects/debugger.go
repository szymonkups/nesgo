package display_objects

import (
	"fmt"
	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/ui/engine"
)

type Debugger struct {
	CPU *core.CPU
}

func (d *Debugger) Draw(e *engine.UIEngine) error {

	// DRAW HEADER WITH FPS
	e.DrawRect(0, 0, 256, 240, 0xFF, 0, 0, 0)
	e.FillRect(0, 0, 256, 8, 0xFF, 0, 0, 0)

	err := e.DrawText("NES CPU DEBUGGER", 1, 0, 0, 0, 0, 0)

	if err != nil {
		return err
	}

	err = e.DrawText(fmt.Sprintf("FPS:%3d", e.FPS), 199, 0, 0, 0, 0, 0)

	if err != nil {
		return err
	}

	// Draw registers
	reg := d.CPU.GetDebugInfo()
	drawRegister16(e, "PC", reg.PC, 2, 9)
	drawRegister8(e, "SP", reg.SP, 57+13, 9, false)
	drawRegister8(e, "A", reg.A, 97+25, 9, false)
	drawRegister8(e, "X", reg.X, 137+29, 9, false)
	drawRegister8(e, "Y", reg.Y, 177+33, 9, true)

	return nil
}

func (d *Debugger) GetChildren() []engine.Displayable {
	return nil
}

func drawRegister16(e *engine.UIEngine, name string, value uint16, x, y int32) {
	e.DrawRect(x, y, 67, 10, 0xFF, 0, 0, 0)
	e.FillRect(x+17, y, 49, 10, 0xFF, 0, 0, 0)
	e.DrawText(name, x+1, y+1, 0xFF, 0, 0, 0)
	e.DrawText(fmt.Sprintf("0x%04X", value), x+18, y+1, 0xff, 0xff, 0xff, 0xff)
}

func drawRegister8(e *engine.UIEngine, name string, value uint8, x, y int32, fix bool) {
	s := int32(0)
	z := int32(0)

	if len(name) == 1 {
		s = 8
	}

	// Stupid fix to make last register end wit 1px gap not 2px
	if fix {
		z = 1
	}

	e.DrawRect(x, y, 51-s, 10, 0xFF, 0, 0, 0)
	e.FillRect(x+17-s, y, 34+z, 10, 0xFF, 0, 0, 0)
	e.DrawText(name, x+1, y+1, 0xFF, 0, 0, 0)
	e.DrawText(fmt.Sprintf("0x%02X", value), x+18-s, y+1, 0xff, 0xff, 0xff, 0xff)
}
