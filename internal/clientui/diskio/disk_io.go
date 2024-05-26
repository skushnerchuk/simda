package diskio

import (
	"fmt"
	"sort"

	"github.com/rivo/tview"
	"github.com/skushnerchuk/simda/internal/clientui/theme"
	uiutils "github.com/skushnerchuk/simda/internal/clientui/utils"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
	"github.com/skushnerchuk/simda/internal/utils"
)

var (
	defaultTitle                = fmt.Sprintf("[%s::b] Disk I/O 🟢 ", theme.UnfocusedBorderColor.String())
	defaultFocusedTitle         = fmt.Sprintf("[%s::b] Disk I/O 🟢 ", theme.FocusedBorderColor.String())
	defaultDisabledTitle        = fmt.Sprintf("[%s::b] Disk I/O 🔴 ", theme.UnfocusedBorderColor.String())
	defaultFocusedDisabledTitle = fmt.Sprintf("[%s::b] Disk I/O 🔴 ", theme.FocusedBorderColor.String())
)

type ViewDiskIO struct {
	View    *tview.Table
	cols    []uiutils.Column
	focused bool
	enabled bool
}

func NewDiskIOView() *ViewDiskIO {
	cols := []uiutils.Column{
		{Text: "Device", MaxWidth: 10},
		{Text: "TPS", MaxWidth: 5},
		{Text: "Read kb/s", MaxWidth: 10},
		{Text: "Write kb/s", MaxWidth: 11},
		{Text: "Total kb/s", MaxWidth: 11},
	}
	v := ViewDiskIO{
		View: uiutils.CreateTable(cols, ""),
		cols: cols,
	}
	v.View.SetBorders(false)
	v.View.SetTitle(defaultTitle)
	v.View.SetFocusFunc(func() {
		v.focused = true
		v.SetTitle()
	})
	v.View.SetBlurFunc(func() {
		v.focused = false
		v.SetTitle()
	})
	return &v
}

func (v *ViewDiskIO) SetTitle() {
	title := defaultTitle
	focusedTitle := defaultFocusedTitle
	if !v.enabled {
		title = defaultDisabledTitle
		focusedTitle = defaultFocusedDisabledTitle
	}
	if v.focused {
		v.View.SetBorderColor(theme.FocusedBorderColor)
		v.View.SetTitle(focusedTitle)
	} else {
		v.View.SetBorderColor(theme.UnfocusedBorderColor)
		v.View.SetTitle(title)
	}
}

func (v *ViewDiskIO) SetMaxWidth(w int) {
	colCount := v.View.GetColumnCount()
	for i := 0; i < colCount; i++ {
		v.View.GetCell(0, i).SetMaxWidth(w / colCount)
	}
}

func (v *ViewDiskIO) SetData(data []*pb.DiskIO, enabled bool) {
	v.enabled = enabled
	v.View.Clear()
	v.SetTitle()

	if !enabled {
		return
	}

	sort.Slice(data, func(i, j int) bool { return data[i].WrSpeed > data[j].WrSpeed })

	for idx, column := range v.cols {
		v.View.SetCell(0, idx,
			uiutils.CreateHeaderCell(column.Text, column.MaxWidth, tview.AlignCenter),
		)
	}

	for i, d := range data {
		v.View.SetCell(i+1, 0, uiutils.CreateCell(d.Name, 8, tview.AlignCenter))

		s := fmt.Sprintf("%.2f", utils.RoundFloat(d.Tps, 2))
		v.View.SetCell(i+1, 1, uiutils.CreateCell(s, 8, tview.AlignCenter))

		s = fmt.Sprintf("%.2f", utils.RoundFloat(d.RdSpeed, 2))
		v.View.SetCell(i+1, 2, uiutils.CreateCell(s, 8, tview.AlignCenter))

		s = fmt.Sprintf("%.2f", utils.RoundFloat(d.WrSpeed, 2))
		v.View.SetCell(i+1, 3, uiutils.CreateCell(s, 8, tview.AlignCenter))

		s = fmt.Sprintf("%.2f", utils.RoundFloat(d.WrSpeed+d.RdSpeed, 2))
		v.View.SetCell(i+1, 4, uiutils.CreateCell(s, 8, tview.AlignCenter))
	}
	v.View.SetFixed(1, 0)
}
