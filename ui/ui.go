package ui

import (
	"fmt"
	"github.com/szymonkups/nesgo/ui/colors"
	"github.com/szymonkups/nesgo/ui/utils"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"math"
)

type UI struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	font     *ttf.Font
}

func (ui *UI) CreateWindow() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}

	if err := ttf.Init(); err != nil {
		return err
	}

	mode, err := sdl.GetDesktopDisplayMode(0)

	if err != nil {
		return err
	}

	// Calculate starting window size - slightly smaller than screen size
	w := math.Floor(float64(mode.W-150) / 256)
	h := math.Floor(float64(mode.H-150) / 240)
	scale := int32(math.Min(w, h))

	// Create window
	window, err := sdl.CreateWindow("NESgo", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		256*scale, 240*scale, sdl.WINDOW_SHOWN)

	if err != nil {
		return err
	}

	ui.window = window

	// Create font
	if ui.font, err = ttf.OpenFont("./assets/silkscreen/slkscr.ttf", 8); err != nil {
		return err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	if err != nil {
		return err
	}

	err = renderer.SetLogicalSize(256, 240)

	if err != nil {
		return err
	}

	ui.renderer = renderer

	return nil
}

func (ui *UI) Draw() error {
	renderer := ui.renderer

	renderer.Clear()

	fps, enoughData := utils.CalculateFPS()

	ui.renderer.SetDrawColor(0, 0, 0, 0)
	ui.renderer.Clear()

	rect := sdl.Rect{0, 0, 256, 240}
	ui.renderer.SetDrawColor(colors.HeaderBg.R, colors.HeaderBg.G, colors.HeaderBg.B, colors.HeaderBg.A)
	ui.renderer.DrawRect(&rect)

	title := sdl.Rect{0, 0, 256, 9}
	ui.renderer.FillRect(&title)

	ui.drawText("NES CPU debugger", 1, 0, colors.HeaderText)

	fpsText := ""

	if enoughData {
		fpsText = fmt.Sprintf("FPS: %d", fps)
	} else {
		fpsText = "FPS: --"
	}

	ui.drawText(fpsText, 220, 0, colors.HeaderText)

	renderer.Present()
	return nil
}

func (ui *UI) DestroyWindow() {
	ui.font.Close()
	sdl.Quit()
	ttf.Quit()
}
