package ui

import (
	"math"

	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/ui/display_objects"
	"github.com/szymonkups/nesgo/ui/engine"
	"github.com/szymonkups/nesgo/ui/engine/utils"
)

type UI struct {
	engine *engine.UIEngine

	// Display objects
	debugger *display_objects.Debugger
}

const (
	windowWidth  = 256 * 2
	windowHeight = 240 * 2
)

func (ui *UI) Init(cpu *core.CPU, ppu *core.PPU, crt *core.Cartridge) error {
	ui.engine = new(engine.UIEngine)
	err := ui.engine.Init()
	if err != nil {
		return err
	}

	// Calculate starting window size - slightly smaller than screen size
	screenW, screenH, err := utils.GetScreenResolution()

	if err != nil {
		return err
	}

	w := float64(screenW-150) / windowWidth
	h := float64(screenH-150) / windowHeight
	scale := math.Min(w, h)

	err = ui.engine.CreateWindow(int32(scale*float64(windowWidth)), int32(scale*float64(windowHeight)), windowWidth, windowHeight)

	if err != nil {
		return err
	}

	// Initialize all display objects
	ui.debugger = &display_objects.Debugger{CPU: cpu, PPU: ppu, CRT: crt}

	return nil
}

func (ui *UI) Destroy() {
	ui.engine.Destroy()

}

func (ui *UI) DrawDebugger(paletteId uint8) error {
	//Clear screen.
	err := ui.engine.ClearScreen(0, 0, 0, 0)

	if err != nil {
		return err
	}

	ui.debugger.SetPaletteId(paletteId)
	ui.engine.DrawRect(0, 0, 200, 200, 255, 0, 0, 255)
	ui.engine.Render(ui.debugger)
	return nil
}

func (ui *UI) DrawScreen(screen []byte) error {
	//Clear screen.
	err := ui.engine.ClearScreen(0, 0, 0, 0)

	if err != nil {
		return err
	}

	ui.engine.SetScreenPixels(screen)
	ui.engine.Render(ui.debugger)
	ui.engine.Present()
	return nil
}
