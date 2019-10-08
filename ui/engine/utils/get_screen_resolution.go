package utils

import "github.com/veandco/go-sdl2/sdl"

func GetScreenResolution() (w int32, h int32, err error) {
	mode, err := sdl.GetDesktopDisplayMode(0)

	if err != nil {
		return 0, 0, err
	}

	return mode.W, mode.H, nil
}
