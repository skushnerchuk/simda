package loadavg

import (
	"fmt"
	"math"

	"github.com/rivo/tview"
	"github.com/skushnerchuk/simda/internal/clientui/theme"
	"github.com/skushnerchuk/simda/internal/clientui/utils"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
)

var (
	defaultTitle         = fmt.Sprintf("[%s::b] 🟢Load avg:[white::-] ", theme.LabelColor.String())
	defaultDisabledTitle = fmt.Sprintf("[%s::b] 🔴Load avg:[white::-] disabled", theme.LabelColor.String())
	defaultUnknownTitle  = fmt.Sprintf("[%s::b] 🔴Load avg:[white::-] unknown", theme.LabelColor.String())
)

type ViewLoadAvg struct {
	View    *tview.TextView
	enabled bool
}

func NewLoadAvgView() *ViewLoadAvg {
	v := ViewLoadAvg{
		View: tview.NewTextView(),
	}
	v.View.SetBorder(false)
	v.View.SetDynamicColors(true)
	v.View.SetTextAlign(tview.AlignRight)
	utils.Str(v.View, defaultUnknownTitle)
	return &v
}

func (v *ViewLoadAvg) SetData(data *pb.LoadAverage, enabled bool) {
	v.enabled = enabled
	if !enabled {
		v.View.SetText(defaultDisabledTitle)
		return
	}

	if data != nil && !math.IsNaN(data.One) && !math.IsNaN(data.Five) && !math.IsNaN(data.Fifteen) {
		s := defaultTitle + "%.2f %.2f %.2f"
		s = fmt.Sprintf(s, data.One, data.Five, data.Fifteen)
		v.View.SetText(s)
	}
}
