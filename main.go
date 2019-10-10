package main

import (
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
	//
	//err := crt.LoadFile("/home/szymon/Downloads/nes/baseball.nes")
	//
	//if err != nil {
	//	fmt.Printf("Could not load a file: %s.\n", err)
	//}

	cpu := core.NewCPU(bus)
	//
	//cpu.Clock()
	gui := new(ui.UI)
	err := gui.Init(&cpu)

	if err != nil {
		panic(err)
	}

	defer gui.Destroy()

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

		err = gui.DrawDebugger()

		if err != nil {
			panic(err)
		}

		sdl.Delay(1000 / 60)
	}
}
