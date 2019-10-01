package ui

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"os"
)

type UI struct {
	window *sdl.Window
	renderer *sdl.Renderer
}

func (ui *UI) CreateWindow() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}

	if err := ttf.Init(); err != nil {
		return err
	}

	window, err := sdl.CreateWindow("NESgo", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)

	if err != nil {
		return err
	}

	ui.window = window

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	if err != nil {
		return err
	}

	ui.renderer = renderer

	return nil
}

func (ui *UI) Draw() {
	renderer := ui.renderer

	viewportRect := renderer.GetViewport()
	//w, h := ui.window.GetSize()
	//viewportRect.X = (w - 800) / 2
	//viewportRect.Y = (h - 600) / 2
	//renderer.SetViewport(&viewportRect)

	//if w > h {
	//	scale := float32(viewportRect.H / h)
		fmt.Println(viewportRect)
	//}

	//renderer.SetScale(2,2)

	renderer.SetLogicalSize(800, 600)
	var font *ttf.Font
	var solid *sdl.Surface
	var err error

	if font, err = ttf.OpenFont("./assets/snoot-org-pixel10/px10.ttf", 14); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open font: %s\n", err)
		return
	}
	defer font.Close()

	if solid, err = font.RenderUTF8Solid("CPU REGISTERS: ", sdl.Color{R: 0, G: 0xFF, B: 0, A: 0}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to render text: %s\n", err)
		return
	}
	defer solid.Free()

	//if err = solid.Blit(nil, renderer, nil); err != nil {
	//	fmt.Fprintf(os.Stderr, "Failed to put text on window surface: %s\n", err)
	//	return
	//}

	ui.renderer.SetDrawColor(0, 0, 0, 0)
	ui.renderer.Clear()

	texture, _ := renderer.CreateTextureFromSurface(solid)
	renderer.Copy(texture, &solid.ClipRect, &solid.ClipRect)

	rect := sdl.Rect{0, 0, 800, 600}
	ui.renderer.SetDrawColor(255, 0, 0, 0)
	ui.renderer.DrawRect(&rect)

	ui.renderer.Present()
}

func (ui *UI) DestroyWindow() {
	sdl.Quit()
	ttf.Quit()
}

