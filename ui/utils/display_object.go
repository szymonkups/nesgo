package utils

import "github.com/veandco/go-sdl2/sdl"

type DrawingContext struct {
	Renderer sdl.Renderer
}

type DisplayObject struct {
	Children []*DisplayObject
	Draw     func(ctx *DrawingContext)
}
