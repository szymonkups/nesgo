package ui

import (
	"github.com/szymonkups/nesgo/ui/display_objects"
	"github.com/szymonkups/nesgo/ui/engine"
	"github.com/szymonkups/nesgo/ui/engine/utils"
	"math"
)

type UI struct {
	engine *engine.UIEngine

	// Display objects
	debugger *engine.DisplayObject
}

const (
	windowWidth  = 256
	windowHeight = 240
)

func (ui *UI) Init() error {
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

	w := math.Floor(float64(screenW-150) / windowWidth)
	h := math.Floor(float64(screenH-150) / windowHeight)
	scale := int32(math.Min(w, h))

	err = ui.engine.CreateWindow(windowWidth*scale, windowHeight*scale, windowWidth, windowHeight)

	if err != nil {
		return err
	}

	// Initialize all display objects
	ui.debugger = &display_objects.Debugger
	ui.engine.Children = append(ui.engine.Children, ui.debugger)

	return nil
}

func (ui *UI) Destroy() {
	ui.engine.Destroy()
}

func (ui *UI) Draw() error {
	// Clear screen.
	err := ui.engine.ClearScreen(0, 0, 0, 0)

	if err != nil {
		return err
	}

	ui.engine.Render()
	return nil
}
