package nettabs

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/skushnerchuk/simda/internal/clientui/theme"
	"github.com/skushnerchuk/simda/internal/clientui/utils"
)

func createTab(index, title string, enabled bool) string {
	// ["%s"][white]%s[black][""]
	s := "🟢"
	if !enabled {
		s = "🔴"
	}
	return fmt.Sprintf(`["%s"][%s] %s%s [%s][""]`,
		index,
		theme.TabBackgroundColor,
		s,
		title,
		theme.TabColor,
	)
}

type ViewNetTabs struct {
	View                *tview.TextView
	tabConnection       string
	tabState            string
	tabTopByProtocol    string
	tabTopByConnections string
}

func NewNetworkTabsView(pages *tview.Pages) *ViewNetTabs {
	v := ViewNetTabs{
		View: tview.NewTextView(),
	}

	v.View.SetTextAlign(tview.AlignLeft).
		SetWrap(false).
		SetRegions(true).
		SetDynamicColors(true)
	v.View.SetBackgroundColor(tcell.ColorBlack)
	v.View.SetBorder(false)

	v.tabConnection = createTab("0", "Connections", true)
	v.tabState = createTab("1", "States", true)
	v.tabTopByProtocol = createTab("2", "Top by protocols", true)
	v.tabTopByConnections = createTab("3", "Top by connections", true)

	utils.Str(v.View, v.tabConnection)
	utils.Str(v.View, v.tabState)
	utils.Str(v.View, v.tabTopByProtocol)
	utils.Str(v.View, v.tabTopByConnections)

	v.View.SetHighlightedFunc(func(added, _, _ []string) {
		if len(added) > 0 {
			pages.SwitchToPage("page-" + added[0])
		} else {
			v.View.Highlight("0")
		}
	})
	v.View.Highlight("0")
	return &v
}

func (v *ViewNetTabs) Update(
	connectionsEnabled, statesEnabled, topByProtocolEnabled, topByConnectionsEnabled bool,
) {
	v.tabConnection = createTab("0", "Connections", connectionsEnabled)
	v.tabState = createTab("1", "States", statesEnabled)
	v.tabTopByProtocol = createTab("2", "Top by protocols", topByProtocolEnabled)
	v.tabTopByConnections = createTab("3", "Top by connections", topByConnectionsEnabled)

	v.View.Clear()

	utils.Str(v.View, v.tabConnection)
	utils.Str(v.View, v.tabState)
	utils.Str(v.View, v.tabTopByProtocol)
	utils.Str(v.View, v.tabTopByConnections)
}
