package engine

import (
	"github.com/szymonkups/nesgo/ui/engine/utils"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type UIEngine struct {
	font     *ttf.Font
	window   *sdl.Window
	renderer *sdl.Renderer
	FPS      uint32
}

func (ui *UIEngine) Init() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}

	if err := ttf.Init(); err != nil {
		return err
	}

	return nil
}

func (ui *UIEngine) CreateWindow(w int32, h int32, logicalW int32, logicalH int32) error {
	// Create window
	window, err := sdl.CreateWindow("NESgo", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, w, h, sdl.WINDOW_SHOWN)

	if err != nil {
		return err
	}

	ui.window = window

	// Create font
	if ui.font, err = ttf.OpenFont("./assets/kongtext/kongtext.ttf", 8); err != nil {
		return err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	if err != nil {
		return err
	}

	err = renderer.SetLogicalSize(logicalW, logicalH)

	if err != nil {
		return err
	}

	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	ui.renderer = renderer

	return nil
}

func (ui *UIEngine) ClearScreen(r, g, b, a uint8) error {
	err := ui.renderer.SetDrawColor(r, g, b, a)

	if err != nil {
		return err
	}

	err = ui.renderer.Clear()

	if err != nil {
		return err
	}

	return nil
}

func (ui *UIEngine) Render(root Displayable) {
	fps, isData := utils.CalculateFPS()

	if isData {
		ui.FPS = fps
	} else {
		ui.FPS = 0
	}

	root.Draw(ui)
	ui.renderChildren(root.GetChildren())
	ui.renderer.Present()
}

func (ui *UIEngine) DrawRect(x, y, w, h int32, r, g, b, a uint8) {
	rect := sdl.Rect{X: x, Y: y, W: w, H: h}
	ui.renderer.SetDrawColor(r, g, b, a)
	ui.renderer.DrawRect(&rect)
}

func (ui *UIEngine) FillRect(x, y, w, h int32, r, g, b, a uint8) {
	rect := sdl.Rect{X: x, Y: y, W: w, H: h}
	ui.renderer.SetDrawColor(r, g, b, a)
	ui.renderer.FillRect(&rect)
}

func (ui *UIEngine) DrawPixel(x, y int32, r, g, b, a uint8) {
	ui.renderer.SetDrawColor(r, g, b, a)
	ui.renderer.DrawPoint(x, y)
}

func (ui *UIEngine) DrawText(text string, x int32, y int32, r, g, b, a uint8) (err error) {
	surface, err := ui.font.RenderUTF8Blended(text, sdl.Color{
		R: r,
		G: g,
		B: b,
		A: a,
	})

	if err != nil {
		return err
	}

	defer surface.Free()

	tex, err := ui.renderer.CreateTextureFromSurface(surface)

	if err != nil {
		return err
	}

	defer tex.Destroy()

	tex.SetAlphaMod(a)

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

func (ui *UIEngine) Destroy() {
	ui.font.Close()
	sdl.Quit()
	ttf.Quit()
}

func (ui *UIEngine) renderChildren(children []Displayable) {
	if children == nil {
		return
	}

	for _, item := range children {
		item.Draw(ui)
	}

	for _, item := range children {
		ui.renderChildren(item.GetChildren())
	}
}
