package main

import (
	"fmt"
	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/ui"
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"sync"
)

func main() {
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

	// Connect devices to CPU bus.
	cpuBus.ConnectDevice(crt) // This must be first to allow grab any address and map it as it want s.
	cpuBus.ConnectDevice(ram)
	cpuBus.ConnectDevice(ppu)

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

	messages := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go cpuLoop(messages, &wg, cpu, ppu)
	sdlLoop(messages, gui)
	wg.Wait()
}

var mutex = &sync.Mutex{}
var screen [500][500]core.PPUColor = [500][500]core.PPUColor{}

func cpuLoop(messages chan string, wg *sync.WaitGroup, cpu *core.CPU, ppu *core.PPU) {
	cycles := 0
	running := true
	stepMode := true

	ppu.SetDrawMethod(func(x, y int16, pixel *core.PPUColor) {
		if x == -1 || y == -1 {
			return
		}

		mutex.Lock()
		screen[x][y] = *pixel
		mutex.Unlock()
	})

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

	for running {
		select {
		case msg := <-messages:
			if msg == "quit" {
				running = false
			}

			if msg == "step toggle" {
				stepMode = !stepMode
			}

			if msg == "step" {
				for {
					tick()
					if cpu.GetCyclesLeft() == 0 && cycles%3 == 0 {
						break
					}
				}
			}

			if msg == "reset" {
				cpu.Reset()
				cycles = 0
			}

		default:
		}

		if !stepMode {
			tick()
		}

		//time.Sleep(1000 / 30)
	}

	wg.Done()
}

func sdlLoop(messages chan string, ui *ui.UI) {
	running := true
	paletteId := uint8(0)

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyboardEvent:
				if t.GetType() == sdl.KEYDOWN {
					switch t.Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false

					case sdl.K_RETURN:
						messages <- "step"

					case sdl.K_SPACE:
						messages <- "step toggle"

					case sdl.K_p:
						paletteId = (paletteId + 1) % 8

					case sdl.K_r:
						messages <- "reset"
					}
				}
			}
		}

		mutex.Lock()
		ui.DrawScreen(screen)
		mutex.Unlock()
		sdl.Delay(1000 / 60)
	}

	messages <- "quit"
}
