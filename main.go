package main

import (
	"github.com/szymonkups/nesgo/core"
)

func main() {
	bus := new(core.Bus)
	bus.ConnectDevice(new(core.Ram))
	bus.ConnectDevice(new(core.PPU))



	//cpu := core.NewCPU(cpuBus)
	//
	//cpu.Clock()
}
