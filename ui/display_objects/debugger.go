package display_objects

import (
	"github.com/szymonkups/nesgo/ui"
	"github.com/szymonkups/nesgo/ui/colors"
	"github.com/szymonkups/nesgo/ui/utils"
	"github.com/veandco/go-sdl2/sdl"
)

var Debugger = utils.DisplayObject{
	Children: []*utils.DisplayObject{},
	Draw: func(ctx *utils.DrawingContext) {
		fps, enoughData := utils.CalculateFPS()

		rect := sdl.Rect{0, 0, 256, 240}
		ctx.Renderer.SetDrawColor(colors.HeaderBg.R, colors.HeaderBg.G, colors.HeaderBg.B, colors.HeaderBg.A)
		ctx.Renderer.DrawRect(&rect)

		title := sdl.Rect{0, 0, 256, 9}
		ctx.Renderer.FillRect(&title)
		ui.drawText("NES CPU debugger", 1, 0, colors.HeaderText)
	},
}
