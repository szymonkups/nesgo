package ui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"math"
)

type UI struct {
	window *sdl.Window
	renderer *sdl.Renderer
	font *ttf.Font
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
		256 * scale, 240 * scale, sdl.WINDOW_SHOWN )

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

	//viewportRect := renderer.GetViewport()
	//w, h := ui.window.GetSize()
	//viewportRect.X = (w - 800) / 2
	//viewportRect.Y = (h - 600) / 2
	//renderer.SetViewport(&viewportRect)

	//if w > h {
	//	scale := float32(viewportRect.H / h)
	//}

	//renderer.SetScale(2,2)

	var solid *sdl.Surface
	var err error

	ui.renderer.SetDrawColor(0, 0, 0, 0)
	ui.renderer.Clear()

	rect := sdl.Rect{0, 0, 256, 240}
	ui.renderer.SetDrawColor(255, 0, 0, 0)
	ui.renderer.DrawRect(&rect)

	title := sdl.Rect{0,0,256, 10}
	ui.renderer.FillRect(&title)

	if solid, err = ui.font.RenderUTF8Solid("CPU DEBUGGER", sdl.Color{R: 0, G: 0, B: 0, A: 0}); err != nil {
		return err
	}

	texture, _ := renderer.CreateTextureFromSurface(solid)
	renderer.Copy(texture, &solid.ClipRect, &solid.ClipRect)

	defer solid.Free()

	renderer.Present()
	return nil
}

func (ui *UI) DestroyWindow() {
	ui.font.Close()
	sdl.Quit()
	ttf.Quit()
}

