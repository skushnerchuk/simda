package nettopbyconnection

import (
	"fmt"
	"sort"

	"github.com/rivo/tview"
	uiutils "github.com/skushnerchuk/simda/internal/clientui/utils"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
	"github.com/skushnerchuk/simda/internal/utils"
)

type ViewNetConnectionsByClient struct {
	View *tview.Table
	cols []uiutils.Column
}

func NewNetworkConnectionsByClientView() *ViewNetConnectionsByClient {
	cols := []uiutils.Column{
		{Text: "Protocol", MaxWidth: 0},
		{Text: "Bytes", MaxWidth: 0},
		{Text: "Percent", MaxWidth: 0},
		{Text: "Source", MaxWidth: 0},
		{Text: "Destination", MaxWidth: 0},
	}
	v := ViewNetConnectionsByClient{View: uiutils.CreateTable(cols, " Top by connection "), cols: cols}
	v.View.SetBorder(false)
	v.View.SetBorders(false)
	return &v
}

func (v *ViewNetConnectionsByClient) SetData(data []*pb.NetTopByConnection, enabled bool) {
	v.View.Clear()

	if !enabled {
		return
	}

	sort.Slice(data, func(i, j int) bool { return data[i].Percent > data[j].Percent })

	for idx, column := range v.cols {
		v.View.SetCell(0, idx, uiutils.CreateHeaderCell(column.Text, column.MaxWidth, tview.AlignLeft))
	}

	for i, d := range data {
		v.View.SetCell(i+1, 0, uiutils.CreateCell(d.Protocol, 0, tview.AlignLeft))

		s := fmt.Sprintf("%d", d.Bytes)
		v.View.SetCell(i+1, 1, uiutils.CreateCell(s, 0, tview.AlignLeft))

		s = fmt.Sprintf("%.2f", utils.RoundFloat(d.Percent, 2))
		v.View.SetCell(i+1, 2, uiutils.CreateCell(s, 0, tview.AlignCenter))

		v.View.SetCell(i+1, 3, uiutils.CreateCell(uiutils.AddrToString(d.SourceAddr), 0, tview.AlignLeft))
		v.View.SetCell(i+1, 4, uiutils.CreateCell(uiutils.AddrToString(d.DestinationAddr), 0, tview.AlignLeft))
	}
	v.View.SetFixed(1, 0)
	v.View.ScrollToBeginning()
}
