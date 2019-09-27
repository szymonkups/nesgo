package main

import (
	"fmt"
	"github.com/szymonkups/nesgo/core"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
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

	if err := ttf.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize TTF: %s\n", err)
		return
	}

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	var font *ttf.Font
	var surface *sdl.Surface
	var solid *sdl.Surface

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

	if surface, err = window.GetSurface(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get window surface: %s\n", err)
		return
	}

	if err = solid.Blit(nil, surface, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to put text on window surface: %s\n", err)
		return
	}

	// Show the pixels for a while
	window.UpdateSurface()

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
