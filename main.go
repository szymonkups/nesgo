package main

import (
	"fmt"
	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/ui"
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

func main() {
	//sdl.Main(run)
	//

	run()
}

func run() {
	// Create two main buses:
	// 1. CPU bus where RAM, ppu and cartridge are connected and used by CPU,
	// 2. ppu bus where cartridge is connected and used by ppu.
	cpuBus := core.NewCPUBus()
	ppuBus := core.NewPPUBus()

	// Crate main system components: RAM, VRAM, ppu and cartridge.
	ram := new(core.Ram)
	ppu := core.NewPPU(ppuBus)
	crt := new(core.Cartridge)
	// TODO: think about better separation of vRam and crt
	vRam := core.NewVRam(crt)
	controller := new(core.Controller)

	// Connect devices to CPU bus.
	cpuBus.ConnectDevice(crt) // This must be first to allow grab any address and map it as it want s.
	cpuBus.ConnectDevice(ram)
	cpuBus.ConnectDevice(ppu)
	cpuBus.ConnectDevice(controller)

	// Connect devices to PPU bus.
	ppuBus.ConnectDevice(crt) // This must be first to allow grab any address and map it as it wants.
	ppuBus.ConnectDevice(vRam)

	err := crt.LoadFile("/home/szymon/Downloads/nes/dk.nes")

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

	cycles := 0
	running := true
	stepMode := false
	//paletteId := uint8(0)

	tick := func() {
		ppu.Clock()
		if cycles%3 == 0 {
			cpu.Clock()
		}
		cycles++

		if ppu.NMI {
			ppu.NMI = false
			cpu.ScheduleNMI()
		}
	}

	screen := make([]uint8, 256*240*4)
	ppu.SetDrawMethod(func(x, y int16, pixel *core.PPUColor) {
		if x >= 256 || y >= 240 {
			return
		}

		offset := (256 * 4 * int32(y)) + int32(x)*4
		screen[offset+0] = pixel.B
		screen[offset+1] = pixel.G
		screen[offset+2] = pixel.R
		screen[offset+3] = 0xFF
	})

	for running {
		startTime := sdl.GetTicks()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyboardEvent:
				if t.GetType() == sdl.KEYDOWN {
					switch t.Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false
					//
					//case sdl.K_RETURN:
					//	messages <- "step"
					//
					case sdl.K_RETURN:
						controller.PressButton(core.ButtonStart)
					case sdl.K_SPACE:
						controller.PressButton(core.ButtonSelect)
					case sdl.K_UP:
						controller.PressButton(core.ButtonUp)
					case sdl.K_DOWN:
						controller.PressButton(core.ButtonDown)

						//case sdl.K_p:
						//	paletteId = (paletteId + 1) % 8
						//
						//case sdl.K_r:
						//	messages <- "reset"
					}
				}

				if t.GetType() == sdl.KEYUP {
					switch t.Keysym.Sym {
					case sdl.K_SPACE:
						controller.ReleaseButton(core.ButtonSelect)
					case sdl.K_RETURN:
						controller.ReleaseButton(core.ButtonStart)
					case sdl.K_UP:
						controller.ReleaseButton(core.ButtonUp)
					case sdl.K_DOWN:
						controller.ReleaseButton(core.ButtonDown)
					}
				}
			}
		}

		if !stepMode {
			for !ppu.IsFrameComplete {
				tick()
			}

			ppu.IsFrameComplete = false

		}

		gui.DrawScreen(screen)
		diff := sdl.GetTicks() - startTime
		fmt.Println(1000/60, diff)
		//if diff < 1000/60 {
		//	sdl.Delay(1000/60 - diff)
		//}

	}
}
