package netstates

import (
	"fmt"
	"sort"

	"github.com/rivo/tview"
	"github.com/skushnerchuk/simda/internal/clientui/utils"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
)

type NetworkConnectionsStatesView struct {
	View *tview.Table
	cols []utils.Column
}

func ViewNetConnectionsStates() *NetworkConnectionsStatesView {
	cols := []utils.Column{
		{Text: "State", Align: 0},
		{Text: "Count", Align: 0},
	}
	v := NetworkConnectionsStatesView{
		View: utils.CreateTable(cols, ""),
		cols: cols,
	}
	v.View.SetBorder(false)
	v.View.SetBorders(false)
	return &v
}

func (v *NetworkConnectionsStatesView) SetData(data []*pb.NetConnectionStates, enabled bool) {
	v.View.Clear()

	if !enabled {
		return
	}

	sort.Slice(data, func(i, j int) bool { return data[i].State < data[j].State })

	for idx, column := range v.cols {
		v.View.SetCell(0, idx, utils.CreateHeaderCell(column.Text, column.MaxWidth, tview.AlignLeft))
	}

	for i, d := range data {
		v.View.SetCell(i+1, 0, utils.CreateCell(d.State, 0, tview.AlignLeft))
		v.View.SetCell(i+1, 1, utils.CreateCell(fmt.Sprintf("%d", d.Count), 0, tview.AlignLeft))
	}
	v.View.SetFixed(1, 0)
	v.View.ScrollToBeginning()
}
