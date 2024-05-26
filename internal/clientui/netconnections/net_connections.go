package netconnections

import (
	"github.com/rivo/tview"
	uiutils "github.com/skushnerchuk/simda/internal/clientui/utils"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
)

const (
	colUserWidth    = 12
	colProcessWidth = 50
)

type ViewNetConnections struct {
	View    *tview.Table
	cols    []uiutils.Column
	enabled bool
}

func NewNetworkConnectionsView() *ViewNetConnections {
	cols := []uiutils.Column{
		{Text: "Protocol", MaxWidth: 0, Align: tview.AlignLeft},
		{Text: "Process", MaxWidth: colProcessWidth, Align: tview.AlignLeft},
		{Text: "User", MaxWidth: colUserWidth, Align: tview.AlignLeft},
		{Text: "Source", MaxWidth: 0, Align: tview.AlignLeft},
		{Text: "Destination", MaxWidth: 0, Align: tview.AlignLeft},
	}
	v := ViewNetConnections{View: uiutils.CreateTable(cols, ""), cols: cols}
	v.View.SetBorder(false)
	v.View.SetBorders(false)

	return &v
}

func (v *ViewNetConnections) SetData(data []*pb.NetConnection, enabled bool) {
	v.enabled = enabled
	v.View.Clear()

	if !enabled {
		return
	}

	for idx, column := range v.cols {
		v.View.SetCell(0, idx,
			uiutils.CreateHeaderCell(column.Text, column.MaxWidth, tview.AlignLeft),
		)
	}

	for i, d := range data {
		v.View.SetCell(i+1, 0, uiutils.CreateCell(d.Protocol, 0, tview.AlignLeft))
		if d.Process != nil {
			v.View.SetCell(i+1, 1, uiutils.CreateCell(d.Process.CmdLine, colProcessWidth, tview.AlignLeft))
		}
		v.View.SetCell(i+1, 2, uiutils.CreateCell(d.User, colUserWidth, tview.AlignLeft))
		v.View.SetCell(i+1, 3, uiutils.CreateCell(uiutils.AddrToString(d.LocalAddr), 0, tview.AlignLeft))
		v.View.SetCell(i+1, 4, uiutils.CreateCell(uiutils.AddrToString(d.ForeignAddr), 0, tview.AlignLeft))
	}
	v.View.SetFixed(1, 0)
	v.View.ScrollToBeginning()
}
