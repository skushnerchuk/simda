package connection

import (
	"fmt"

	"github.com/rivo/tview"
)

var (
	caption = "Server: %s (warm: %d sec, recv: %d sec), status: %s"
	paused  = "[red::b]paused[::-]"
	running = "[green::b]running[::-]"
)

type ViewConnections struct {
	View    *tview.TextView
	server  string
	warm    int
	receive int
}

func NewConnectionView(server string, warm, receive int) *ViewConnections {
	v := ViewConnections{
		View:    tview.NewTextView(),
		server:  server,
		warm:    warm,
		receive: receive,
	}
	v.View = tview.NewTextView()
	v.View.SetBorder(false)
	v.View.SetTextAlign(tview.AlignRight)
	v.View.SetDynamicColors(true)
	v.Resume()
	return &v
}

func (v *ViewConnections) Pause() {
	v.View.SetText(fmt.Sprintf(caption, v.server, v.warm, v.receive, paused))
}

func (v *ViewConnections) Resume() {
	v.View.SetText(fmt.Sprintf(caption, v.server, v.warm, v.receive, running))
}
