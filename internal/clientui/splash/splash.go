package splash

import (
	"fmt"

	"github.com/rivo/tview"
)

type ViewSplash struct {
	View     *tview.Flex
	textView *tview.TextView
}

func (w *ViewSplash) Update(elapsedTime int) {
	message := fmt.Sprintf("[orange]Warming in progress, estimated %d seconds\n[white]Press Ctrl-Q to exit", elapsedTime)
	w.textView.SetText(message)
}

func NewWarmingWindow(elapsedTime int) *ViewSplash {
	v := ViewSplash{}
	v.View = tview.NewFlex()
	v.textView = tview.NewTextView()
	v.textView.SetTextAlign(tview.AlignCenter)
	v.textView.SetDynamicColors(true)

	v.View = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 0, 2, false).
		AddItem(v.textView, 2, 0, false).
		AddItem(tview.NewBox(), 0, 2, false)
	v.Update(elapsedTime)
	return &v
}
