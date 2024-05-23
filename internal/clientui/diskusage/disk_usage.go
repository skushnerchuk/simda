package diskusage

import (
	"fmt"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/rivo/tview"
	"github.com/skushnerchuk/simda/internal/clientui/theme"
	uiutils "github.com/skushnerchuk/simda/internal/clientui/utils"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
	"github.com/skushnerchuk/simda/internal/utils"
)

var (
	defaultTitle                = fmt.Sprintf("[%s::b] Disk usage 🟢 ", theme.UnfocusedBorderColor.String())
	defaultFocusedTitle         = fmt.Sprintf("[%s::b] Disk usage 🟢 ", theme.FocusedBorderColor.String())
	defaultDisabledTitle        = fmt.Sprintf("[%s::b] Disk usage 🔴 ", theme.UnfocusedBorderColor.String())
	defaultFocusedDisabledTitle = fmt.Sprintf("[%s::b] Disk usage 🔴 ", theme.FocusedBorderColor.String())
)

type ViewDiskUsage struct {
	View     *tview.Table
	maxWidth int
	cols     []uiutils.Column
	focused  bool
	enabled  bool
}

func NewDiskUsageView() *ViewDiskUsage {
	cols := []uiutils.Column{
		{Text: "Device", MaxWidth: 0},
		{Text: "Mounted", MaxWidth: 10},
		{Text: "Usage", MaxWidth: 0},
		{Text: "Usage %", MaxWidth: 0},
		{Text: "Inode", MaxWidth: 0},
		{Text: "Inode %", MaxWidth: 0},
	}
	v := ViewDiskUsage{View: uiutils.CreateTable(cols, " Disk usage "), cols: cols}
	v.View.SetTitle(defaultTitle)
	v.View.SetBorders(false)
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

func (v *ViewDiskUsage) SetTitle() {
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

func (v *ViewDiskUsage) SetMaxWidth(w int) {
	v.maxWidth = w
}

func (v *ViewDiskUsage) SetData(data []*pb.DiskUsage, enabled bool) {
	v.enabled = enabled
	v.View.Clear()
	v.SetTitle()

	if !enabled {
		return
	}
	sort.Slice(data, func(i, j int) bool { return data[i].Device < data[j].Device })

	for idx, column := range v.cols {
		v.View.SetCell(0, idx, uiutils.CreateHeaderCell(column.Text, column.MaxWidth, tview.AlignLeft))
	}

	for i, d := range data {
		v.View.SetCell(i+1, 0, uiutils.CreateCell(d.Device, 0, tview.AlignLeft))
		v.View.SetCell(i+1, 1, uiutils.CreateCell(d.MountPoint, 10, tview.AlignLeft))

		v.View.SetCell(i+1, 2, uiutils.CreateCell(humanize.Bytes(uint64(d.Usage)), 0, tview.AlignLeft))

		s := fmt.Sprintf("%.2f", utils.RoundFloat(d.UsagePercent, 2))
		v.View.SetCell(i+1, 3, uiutils.CreateCell(s, 0, tview.AlignLeft))

		s = fmt.Sprintf("%d", uint64(d.InodeCount))
		v.View.SetCell(i+1, 4, uiutils.CreateCell(s, 0, tview.AlignLeft))

		s = fmt.Sprintf("%.2f", utils.RoundFloat(d.InodeAvailablePercent, 2))
		v.View.SetCell(i+1, 5, uiutils.CreateCell(s, 0, tview.AlignLeft))
	}
	v.View.SetFixed(1, 0)
}
