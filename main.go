package main

import (
	"fmt"
	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/ui"
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

func main() {
	// Create two main buses:
	// 1. CPU bus where RAM, ppu and cartridge are connected and used by CPU,
	// 2. ppu bus where cartridge is connected and used by ppu.
	cpuBus := core.NewCPUBus()
	ppuBus := core.NewPPUBus()

	// Crate main system components: RAM, VRAM, ppu and cartridge.
	ram := new(core.Ram)
	vRam := new(core.VRam)
	ppu := core.NewPPU(ppuBus)
	crt := new(core.Cartridge)

	// Connect devices to CPU bus.
	cpuBus.ConnectDevice(crt) // This must be first to allow grab any address and map it as it wants.
	cpuBus.ConnectDevice(ram)
	cpuBus.ConnectDevice(ppu)

	// Connect devices to PPU bus.
	ppuBus.ConnectDevice(crt) // This must be first to allow grab any address and map it as it wants.
	ppuBus.ConnectDevice(vRam)

	err := crt.LoadFile("/home/szymon/Downloads/nes/nestest.nes")

	if err != nil {
		fmt.Printf("Could not load a file: %s.\n", err)
		os.Exit(1)
	}

	cpu := core.NewCPU(cpuBus)

	gui := new(ui.UI)
	err = gui.Init(cpu, ppu, crt)

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
				if t.GetType() == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_RETURN {
					ppu.Clock()
					if cycles%3 == 0 {
						cpu.Clock()
					}

					cycles++
				}

				// R key - reset
				if t.GetType() == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_r {
					cpu.Reset()
				}

				break
			}
		}

		//for cpu.GetCyclesLeft()*3 > 0 {
		ppu.Clock()
		if cycles%3 == 0 {
			cpu.Clock()
		}
		cycles++
		//}

		if ppu.NMI {
			ppu.NMI = false
			cpu.ScheduleNMI()
		}

		err = gui.DrawDebugger()

		if err != nil {
			panic(err)
		}

		//sdl.Delay(1000 / 10)
	}
}
