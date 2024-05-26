package nettopbyprotocol

import (
	"fmt"
	"sort"

	"github.com/rivo/tview"
	uiutils "github.com/skushnerchuk/simda/internal/clientui/utils"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
	"github.com/skushnerchuk/simda/internal/utils"
)

type ViewNetConnectionsByProtocol struct {
	View *tview.Table
	cols []uiutils.Column
}

func NewNetworkConnectionsByProtocolView() *ViewNetConnectionsByProtocol {
	cols := []uiutils.Column{
		{Text: "Protocol", MaxWidth: 0},
		{Text: "Bytes", MaxWidth: 0},
		{Text: "Percent", MaxWidth: 0},
	}
	v := ViewNetConnectionsByProtocol{View: uiutils.CreateTable(cols, ""), cols: cols}
	v.View.SetBorder(false)
	v.View.SetBorders(false)
	return &v
}

func (v *ViewNetConnectionsByProtocol) SetData(data []*pb.NetTopByProtocol, enabled bool) {
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
		v.View.SetCell(i+1, 2, uiutils.CreateCell(s, 0, tview.AlignLeft))
	}
	v.View.SetFixed(1, 0)
	v.View.ScrollToBeginning()
}
