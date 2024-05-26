package cpuavg

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/skushnerchuk/simda/internal/clientui/theme"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
)

var (
	defaultTitle         = fmt.Sprintf("[%s::b]🟢CPU:[white::-] ", theme.LabelColor.String())
	defaultDisabledTitle = fmt.Sprintf("[%s::b]🔴CPU:[white::-] disabled", theme.LabelColor.String())
	defaultUnknownTitle  = fmt.Sprintf("[%s::b]🔴CPU:[white::-] unknown", theme.LabelColor.String())
)

type ViewCPUAvg struct {
	View    *tview.TextView
	enabled bool
}

func NewCPUAvgView() *ViewCPUAvg {
	v := ViewCPUAvg{
		View: tview.NewTextView(),
	}
	v.View.SetBorder(false)
	v.View.SetDynamicColors(true)
	v.View.SetTextAlign(tview.AlignLeft)
	_, _ = fmt.Fprint(v.View, defaultUnknownTitle)
	return &v
}

func (v *ViewCPUAvg) SetData(data *pb.CpuAverage, enabled bool) {
	v.enabled = enabled
	if !enabled {
		v.View.SetText(defaultDisabledTitle)
		return
	}

	s := defaultUnknownTitle
	if data != nil {
		s = defaultTitle + "[orange::-]sys[white] %.2f [orange]usr[white] %.2f [orange]idl[white] %.2f"
		s = fmt.Sprintf(s, data.System, data.User, data.Idle)
	}
	v.View.SetText(s)
}
