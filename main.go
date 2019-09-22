package main

import (
	"github.com/szymonkups/nesgo/core"
)

func main() {
	bus := new(core.Bus)
	cpu := core.NewCPU(bus)

	cpu.Clock()
}
