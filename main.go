package main

import (
	"fmt"
	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/ui"
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"sync"
)

//var runningMutex sync.Mutex

//func run() int {
//
//	//defer gui.Destroy()
//
//	cycles := 0
//	running := true
//
//	for running {
//		sdl.Do(func() {
//			stop := func() {
//				runningMutex.Lock()
//				running = false
//				runningMutex.Unlock()
//			}
//
//		})
//
//		ppu.Clock()
//		if cycles%3 == 0 {
//			cpu.Clock()
//		}
//		cycles++
//
//		if ppu.NMI {
//			ppu.NMI = false
//			cpu.ScheduleNMI()
//		}
//
//		if err != nil {
//			panic(err)
//		}
//
//		go func() {
//			err = gui.DrawDebugger()
//
//			if err != nil {
//				panic(err)
//			}
//
//			sdl.Delay(1000 / 60)
//		}()
//
//	}
//
//	return 0
//}

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

	err := crt.LoadFile("/home/szymon/Downloads/nes/smb.nes")

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
	mux := new(sync.Mutex)
	var wg sync.WaitGroup
	wg.Add(1)
	go cpuLoop(messages, &wg, cpu, ppu)
	sdlLoop(messages, mux, gui)
	wg.Wait()
}

func cpuLoop(messages chan string, wg *sync.WaitGroup, cpu *core.CPU, ppu *core.PPU) {
	cycles := 0
	running := true

	for running {
		select {
		case msg := <-messages:
			if msg == "quit" {
				running = false
			}

		default:
		}

		ppu.Clock()
		if cycles%3 == 0 {
			cpu.Clock()
		}
		cycles++
	}

	wg.Done()
}

func sdlLoop(messages chan string, mux *sync.Mutex, ui *ui.UI) {
	running := true

	for running {

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyboardEvent:
				if t.GetType() == sdl.KEYUP && t.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}
			}
		}

		ui.DrawDebugger()
		sdl.Delay(1000 / 200)
	}

	messages <- "quit"
}
