package engine

import (
	"fmt"
	"github.com/szymonkups/nesgo/ui/engine/utils"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"strings"
)

type UIEngine struct {
	window       *sdl.Window
	renderer     *sdl.Renderer
	fontTextures map[string]*sdl.Texture
	FPS          uint32
	screen       *sdl.Texture
}

func (ui *UIEngine) Init() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}

	if err := ttf.Init(); err != nil {
		return err
	}

	ui.fontTextures = map[string]*sdl.Texture{}

	return nil
}

func (ui *UIEngine) CreateWindow(w int32, h int32, logicalW int32, logicalH int32) error {
	// Create window
	window, err := sdl.CreateWindow("NESgo", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, w, h, sdl.WINDOW_SHOWN)

	if err != nil {
		return err
	}

	ui.window = window

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	if err != nil {
		return err
	}

	err = renderer.SetLogicalSize(logicalW, logicalH)

	if err != nil {
		return err
	}

	err = renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	if err != nil {
		return err
	}

	ui.screen, err = renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, 320, 240)
	if err != nil {
		return err
	}

	ui.renderer = renderer

	return nil
}

const glyphs = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&:-,.{}[]()_<>"

func (ui *UIEngine) createFontTexture(r, g, b, a uint8) (*sdl.Texture, error) {
	// Create font
	font, err := ttf.OpenFont("./assets/kongtext/kongtext.ttf", 8)

	if err != nil {
		return nil, err
	}

	surface, err := font.RenderUTF8Blended(glyphs, sdl.Color{
		R: r,
		G: g,
		B: b,
		A: a,
	})

	if err != nil {
		return nil, err
	}

	defer surface.Free()

	tex, err := ui.renderer.CreateTextureFromSurface(surface)

	if err != nil {
		return nil, err
	}

	font.Close()

	return tex, nil
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
func (ui *UIEngine) SetScreenPixels(data []byte) {
	updateRect := &sdl.Rect{
		X: 0,
		Y: 0,
		W: 256,
		H: 240,
	}

	ui.screen.Update(updateRect, data, 256*4)
}

func (ui *UIEngine) Render(root Displayable) {
	fps, isData := utils.CalculateFPS()

	if isData {
		ui.FPS = fps
	} else {
		ui.FPS = 0
	}

	src := &sdl.Rect{
		X: 0,
		Y: 0,
		W: 320,
		H: 240,
	}

	dst := &sdl.Rect{
		X: 0,
		Y: 0,
		W: 320 * 2,
		H: 240 * 2,
	}

	ui.renderer.Copy(ui.screen, src, dst)

	root.Draw(ui)
	ui.renderChildren(root.GetChildren())

	ui.renderer.Present()
}

func (ui *UIEngine) Present() {
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

var letterSrcRect = sdl.Rect{
	X: 0,
	Y: 0,
	W: 8,
	H: 9,
}

var letterDstRect = sdl.Rect{
	X: 0,
	Y: 0,
	W: 8,
	H: 9,
}

func (ui *UIEngine) DrawText(text string, x int32, y int32, r, g, b, a uint8) (err error) {
	// Load font texture if present, otherwise create new one
	fontId := fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
	tex, ok := ui.fontTextures[fontId]

	if !ok {
		tex, err = ui.createFontTexture(r, g, b, a)

		if err != nil {
			return err
		}

		ui.fontTextures[fontId] = tex
	}

	for i, l := range text {
		index := strings.IndexByte(glyphs, byte(l))
		if index == -1 {
			continue
		}

		letterSrcRect.X = int32(index * 8)
		letterDstRect.X = x + int32(i)*8
		letterDstRect.Y = y

		ui.renderer.Copy(tex, &letterSrcRect, &letterDstRect)
	}

	return nil
}

func (ui *UIEngine) Destroy() {
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
