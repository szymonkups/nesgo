package main

import (
	"fmt"
	"github.com/szymonkups/nesgo/core"
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

func main() {
	// Create main system components
	bus := new(core.Bus)
	ram := new(core.Ram)
	ppu := new(core.PPU)
	crt := new(core.Cartridge)

	// Connect them to create NES architecture
	bus.ConnectDevice(ram)
	bus.ConnectDevice(ppu)
	bus.ConnectDevice(crt)

	err := crt.LoadFile("/home/szymon/Downloads/nes/baseball.nes")

	if err != nil {
		fmt.Printf("Could not load a file: %s.\n", err)
	}

	//cpu := core.NewCPU(cpuBus)
	//
	//cpu.Clock()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	if err != nil {
		panic(err)
	}

	var renderer *sdl.Renderer
	var points []sdl.Point
	var rect sdl.Rect
	var rects []sdl.Rect

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return
	}

	defer renderer.Destroy()

	renderer.Clear()
	renderer.SetDrawColor(255, 255, 255, 255)
	renderer.DrawPoint(150, 300)

	renderer.SetDrawColor(0, 0, 255, 255)
	renderer.DrawLine(0, 0, 200, 200)

	points = []sdl.Point{{0, 0}, {100, 300}, {100, 300}, {200, 0}}
	renderer.SetDrawColor(255, 255, 0, 255)
	renderer.DrawLines(points)

	rect = sdl.Rect{300, 0, 200, 200}
	renderer.SetDrawColor(255, 0, 0, 255)
	renderer.DrawRect(&rect)

	rects = []sdl.Rect{{400, 400, 100, 100}, {550, 350, 200, 200}}
	renderer.SetDrawColor(0, 255, 255, 255)
	renderer.DrawRects(rects)

	rect = sdl.Rect{250, 250, 200, 200}
	renderer.SetDrawColor(0, 255, 0, 255)
	renderer.FillRect(&rect)

	rects = []sdl.Rect{{500, 300, 100, 100}, {200, 300, 200, 200}}
	renderer.SetDrawColor(255, 0, 255, 255)
	renderer.FillRects(rects)

	renderer.Present()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
				break

			case *sdl.KeyboardEvent:
				// Quit on ESC
				if t.GetType() == sdl.KEYUP && t.Keysym.Sym == 27 {
					running = false
				}

				break
			}
		}
	}
}
