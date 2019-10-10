package utils

import (
	"github.com/veandco/go-sdl2/sdl"
	"math"
)

const fpsInterval = 1000

var lastFPSTime = sdl.GetTicks()
var framesCount = uint32(0)
var enoughData = false
var fps = uint32(0)

func CalculateFPS() (uint32, bool) {
	framesCount++

	if sdl.GetTicks()-lastFPSTime > fpsInterval {
		lastFPSTime = sdl.GetTicks()
		fps = framesCount
		framesCount = 0
		enoughData = true
	}

	return uint32(math.Min(float64(fps), 999)), enoughData
}
