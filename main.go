package main

import (
	"fmt"
	//"github.com/pkg/profile"
	"github.com/szymonkups/nesgo/core"
	"github.com/szymonkups/nesgo/ui"
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

const (
	screenFPS           uint32 = 60
	screenTicksPerFrame        = 1000 / screenFPS
)

//func main() {
//	p := profile.Start()
//	// os.Exit(..) must run AFTER sdl.Main(..) below; so keep track of exit
//	// status manually outside the closure passed into sdl.Main(..) below
//	var exitcode int
//	sdl.Main(func() {
//		exitcode = run()
//	})
//
//	p.Stop()
//	// os.Exit(..) must run here! If run in sdl.Main(..) above, it will cause
//	// premature quitting of sdl.Main(..) function; resource cleaning deferred
//	// calls/closing of channels may never run
//	os.Exit(exitcode)
//}

func main() {
	// CPU profiling by default

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

	var gui *ui.UI
	gui = new(ui.UI)
	err = gui.Init(cpu, ppu, crt)

	if err != nil {
		panic(err)
	}

	cycles := 0
	running := true
	stepMode := false
	fpsTimer := new(SDLTimer)
	capTimer := new(SDLTimer)
	countedFrames := uint32(0)
	fpsTimer.Start()

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
		// Timer for FPS cap
		capTimer.Start()

		//Calculate and correct FPS
		avgFPS := float32(countedFrames) / (float32(fpsTimer.GetTicks()) / 1000)

		if avgFPS > 2000000 {
			avgFPS = 0
		}

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
		fmt.Println("FPS: ", avgFPS)
		countedFrames++

		//If frame finished early
		frameTicks := capTimer.GetTicks()
		if frameTicks < screenTicksPerFrame {
			//Wait remaining time
			sdl.Delay(screenTicksPerFrame - frameTicks)
		}
	}
}

// Kudos to https://lazyfoo.net/tutorials/SDL/23_advanced_timers/index.php
type SDLTimer struct {
	startTicks  uint32
	pausedTicks uint32
	paused      bool
	started     bool
}

func (t *SDLTimer) Start() {
	t.started = true
	t.paused = false
	t.startTicks = sdl.GetTicks()
	t.pausedTicks = 0
}

func (t *SDLTimer) Stop() {
	t.started = false
	t.paused = false
	t.startTicks = 0
	t.pausedTicks = 0
}

func (t *SDLTimer) Pause() {
	if t.started && !t.paused {
		t.paused = true
		t.pausedTicks = sdl.GetTicks() - t.startTicks
		t.startTicks = 0
	}
}

func (t *SDLTimer) Unpause() {
	if t.started && t.paused {
		t.paused = false
		t.startTicks = sdl.GetTicks() - t.pausedTicks
		t.pausedTicks = 0
	}
}

func (t *SDLTimer) GetTicks() uint32 {
	time := uint32(0)

	if t.started {
		if t.paused {
			time = t.pausedTicks
		} else {
			time = sdl.GetTicks() - t.startTicks
		}
	}

	return time
}

func (t *SDLTimer) IsStarted() bool {
	return t.started
}

func (t *SDLTimer) IsPaused() bool {
	return t.paused && t.started
}
