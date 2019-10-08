package display_objects

import (
	"fmt"
	"github.com/szymonkups/nesgo/ui/colors"
	"github.com/szymonkups/nesgo/ui/engine"
)

var Debugger = engine.DisplayObject{
	Children: nil,
	Draw: func(e *engine.UIEngine) {
		e.DrawRect(0, 0, 256, 240, 0xFF, 0, 0, 0)
		e.FillRect(0, 0, 256, 9, 0xFF, 0, 0, 0)

		e.DrawText("NES CPU debugger", 1, 0, colors.HeaderText)
		e.DrawText(fmt.Sprintf("FPS: %d", e.FPS), 210, 0, colors.HeaderText)
	},
}
