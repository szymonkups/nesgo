package ui

import "github.com/veandco/go-sdl2/sdl"

func (ui *UI) drawText(text string, x int32, y int32, color sdl.Color) (err error) {
	surface, err := ui.font.RenderUTF8Solid(text, color)

	if err != nil {
		return err
	}

	defer surface.Free()

	tex, err := ui.renderer.CreateTextureFromSurface(surface)

	if err != nil {
		return err
	}

	dst := sdl.Rect{
		X: x,
		Y: y,
		W: surface.ClipRect.W,
		H: surface.ClipRect.H,
	}

	err = ui.renderer.Copy(tex, &surface.ClipRect, &dst)

	if err != nil {
		return err
	}

	return nil
}
