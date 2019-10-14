package main

import (
	"fmt"
	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/ui"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	// Create main system components
	bus := new(core.Bus)
	ram := new(core.Ram)
	ppu := new(core.PPU)
	crt := new(core.Cartridge)
	//
	//// Connect them to create NES architecture
	bus.ConnectDevice(ram)
	bus.ConnectDevice(ppu)
	bus.ConnectDevice(crt)

	err := crt.LoadFile("/home/szymon/Downloads/nes/baseball.nes")

	if err != nil {
		fmt.Printf("Could not load a file: %s.\n", err)
	}

	cpu := core.NewCPU(bus)

	gui := new(ui.UI)
	err = gui.Init(&cpu)

	if err != nil {
		panic(err)
	}

	defer gui.Destroy()

	cycles := 0

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
				break

			case *sdl.KeyboardEvent:
				// Quit on ESC
				if t.GetType() == sdl.KEYUP && t.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}

				// Step on enter
				//if t.GetType() == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_RETURN {
				//	cpu.Clock()
				//}

				// R key - reset
				if t.GetType() == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_r {
					cpu.Reset()
				}

				break
			}
		}

		ppu.Clock()
		if cycles%3 == 0 {
			cpu.Clock()
		}

		cycles++

		err = gui.DrawDebugger()

		if err != nil {
			panic(err)
		}

		sdl.Delay(1000 / 60)
	}
}
