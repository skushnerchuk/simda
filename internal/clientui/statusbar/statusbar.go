package statusbar

import (
	"github.com/rivo/tview"
	"github.com/skushnerchuk/simda/internal/clientui/utils"
)

func CreateStatusbar() *tview.TextView {
	widget := tview.NewTextView()
	widget.SetBorder(false)
	widget.SetBorderPadding(0, 0, 1, 1)
	widget.SetDynamicColors(true)
	widget.SetTextAlign(tview.AlignLeft)
	utils.Str(widget, `[orange]Ctrl+Q[white] Exit `)
	utils.Str(widget, `[orange]Ctrl+P[white] Pause `)
	utils.Str(widget, `[orange]Ctrl+R[white] Resume `)
	return widget
}
