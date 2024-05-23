package clientui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/skushnerchuk/simda/internal/clientui/connection"
	"github.com/skushnerchuk/simda/internal/clientui/cpuavg"
	"github.com/skushnerchuk/simda/internal/clientui/diskio"
	"github.com/skushnerchuk/simda/internal/clientui/diskusage"
	"github.com/skushnerchuk/simda/internal/clientui/loadavg"
	"github.com/skushnerchuk/simda/internal/clientui/netconnections"
	"github.com/skushnerchuk/simda/internal/clientui/netstates"
	"github.com/skushnerchuk/simda/internal/clientui/nettabs"
	"github.com/skushnerchuk/simda/internal/clientui/nettopbyconnection"
	"github.com/skushnerchuk/simda/internal/clientui/nettopbyprotocol"
	"github.com/skushnerchuk/simda/internal/clientui/statusbar"
	"github.com/skushnerchuk/simda/internal/clientui/theme"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
)

const logo = `  ___  ____  __  __  ____    __    
 / __)(_  _)(  \/  )(  _ \  /__\   
 \__ \ _)(_  )    (  )(_) )/(__)\  
 (___/(____)(_/\/\_)(____/(__)(__)`

type ViewMainWindow struct {
	data                  *pb.Snapshot
	logoHeight            int
	logoWidth             int
	View                  *tview.Flex
	logo                  *tview.TextView
	connectionView        *connection.ViewConnections
	header                *tview.Flex
	diskUsageView         *diskusage.ViewDiskUsage
	diskIOView            *diskio.ViewDiskIO
	netConnView           *netconnections.ViewNetConnections
	netConnStatesView     *netstates.NetworkConnectionsStatesView
	netConnByProtocolView *nettopbyprotocol.ViewNetConnectionsByProtocol
	netConnByClientView   *nettopbyconnection.ViewNetConnectionsByClient
	loadAvgView           *loadavg.ViewLoadAvg
	cpuAvgView            *cpuavg.ViewCPUAvg
	netTabsView           *nettabs.ViewNetTabs
	refreshPaused         bool
	warm                  int
	receive               int
	server                string
	port                  string
}

func (w *ViewMainWindow) createLogoView() {
	lines := strings.Split(logo, "\n")
	w.logoWidth = 0
	w.logoHeight = len(lines)
	for _, line := range lines {
		if len(line) > w.logoWidth {
			w.logoWidth = len(line)
		}
	}
	w.logo = tview.NewTextView().SetTextColor(tcell.ColorGreen).SetText(logo)
}

func (w *ViewMainWindow) createConnectionView() {
	s := fmt.Sprintf("%s:%s", w.server, w.port)
	w.connectionView = connection.NewConnectionView(s, w.warm, w.receive)
	w.connectionView.View.SetBorderPadding(1, 0, 0, 1)
}

func NewMainView(server, port string, warm, receive int) *ViewMainWindow {
	v := ViewMainWindow{
		server:  server,
		port:    port,
		warm:    warm,
		receive: receive,
	}
	v.createLogoView()
	v.createConnectionView()

	v.loadAvgView = loadavg.NewLoadAvgView()
	v.loadAvgView.View.SetBorderPadding(0, 0, 0, 1)
	v.cpuAvgView = cpuavg.NewCPUAvgView()

	avgBox := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(v.loadAvgView.View, 0, 1, false).
		AddItem(v.cpuAvgView.View, 35, 0, false)
	avgBox.SetBorder(false)

	infoBox := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(avgBox, 0, 1, false).
		AddItem(v.connectionView.View, 0, 1, false)

	v.header = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(v.logo, v.logoWidth, 0, false).
		AddItem(infoBox, 0, 2, false)

	v.diskUsageView = diskusage.NewDiskUsageView()
	v.diskUsageView.View.SetBorderPadding(1, 1, 1, 1)

	v.diskIOView = diskio.NewDiskIOView()
	v.diskIOView.View.SetBorderPadding(1, 1, 1, 1)

	diskMetrics := tview.NewFlex()
	diskMetrics.SetBorder(false)
	diskMetrics.SetDirection(tview.FlexRow)
	diskMetrics.AddItem(v.diskUsageView.View, 0, 1, false)
	diskMetrics.AddItem(v.diskIOView.View, 0, 1, false)

	_, _, w, _ := diskMetrics.GetRect() //nolint:dogsled
	v.diskIOView.SetMaxWidth(w)
	v.diskUsageView.SetMaxWidth(w)

	v.netConnView = netconnections.NewNetworkConnectionsView()
	v.netConnStatesView = netstates.ViewNetConnectionsStates()
	v.netConnByProtocolView = nettopbyprotocol.NewNetworkConnectionsByProtocolView()
	v.netConnByClientView = nettopbyconnection.NewNetworkConnectionsByClientView()

	pages := tview.NewPages().
		AddPage("page-0", v.netConnView.View, true, true).
		AddPage("page-1", v.netConnStatesView.View, true, false).
		AddPage("page-2", v.netConnByProtocolView.View, true, false).
		AddPage("page-3", v.netConnByClientView.View, true, false)

	v.netTabsView = nettabs.NewNetworkTabsView(pages)

	networkMetrics := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(v.netTabsView.View, 2, 0, false).
		AddItem(pages, 0, 1, false)
	networkMetrics.SetBorder(true)

	v.netTabsView.View.SetFocusFunc(func() {
		networkMetrics.SetBorderColor(theme.FocusedBorderColor)
	})
	v.netTabsView.View.SetBlurFunc(func() {
		networkMetrics.SetBorderColor(theme.UnfocusedBorderColor)
	})

	v.netConnView.View.SetFocusFunc(func() {
		networkMetrics.SetBorderColor(theme.FocusedBorderColor)
	})
	v.netConnView.View.SetBlurFunc(func() {
		networkMetrics.SetBorderColor(theme.UnfocusedBorderColor)
	})

	v.netConnStatesView.View.SetFocusFunc(func() {
		networkMetrics.SetBorderColor(theme.FocusedBorderColor)
	})
	v.netConnStatesView.View.SetBlurFunc(func() {
		networkMetrics.SetBorderColor(theme.UnfocusedBorderColor)
	})

	v.netConnByProtocolView.View.SetFocusFunc(func() {
		networkMetrics.SetBorderColor(theme.FocusedBorderColor)
	})
	v.netConnByProtocolView.View.SetBlurFunc(func() {
		networkMetrics.SetBorderColor(theme.UnfocusedBorderColor)
	})

	v.netConnByClientView.View.SetFocusFunc(func() {
		networkMetrics.SetBorderColor(theme.FocusedBorderColor)
	})
	v.netConnByClientView.View.SetBlurFunc(func() {
		networkMetrics.SetBorderColor(theme.UnfocusedBorderColor)
	})

	mainWindow := tview.NewFlex()
	mainWindow.
		AddItem(
			tview.NewFlex().
				AddItem(diskMetrics, 0, 2, false).
				AddItem(networkMetrics, 0, 4, false),
			0, 1, true,
		).
		SetDirection(tview.FlexRow).
		AddItem(statusbar.CreateStatusbar(), 1, 0, false).
		SetBorderPadding(0, 0, 1, 0)

	mainWindow.SetBorder(false)
	v.View = tview.NewFlex().
		AddItem(v.header, v.logoHeight, 0, false).
		SetDirection(tview.FlexRow).
		AddItem(mainWindow, 0, 1, false)

	v.View.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() { //nolint:exhaustive
		case tcell.KeyCtrlP:
			v.refreshPaused = true
			v.connectionView.Pause()
		case tcell.KeyCtrlR:
			v.refreshPaused = false
			v.connectionView.Resume()
		default:
			return event
		}
		return event
	})

	return &v
}

func (w *ViewMainWindow) SetData(data *pb.Snapshot) {
	if w.refreshPaused {
		return
	}
	w.data = data
	w.loadAvgView.SetData(data.LoadAvg, data.Metrics.LoadAvg)
	w.cpuAvgView.SetData(data.CpuAvg, data.Metrics.CpuAvg)
	w.diskIOView.SetData(data.DiskIO, data.Metrics.DiskIO)
	w.diskUsageView.SetData(data.DiskUsage, data.Metrics.DiskUsage)
	w.netConnByProtocolView.SetData(data.NetTopByProtocol, data.Metrics.NetTopByProtocol)
	w.netConnByClientView.SetData(data.NetTopByConnection, data.Metrics.NetTopByConnection)
	w.netConnStatesView.SetData(data.NetConnectionsStates, data.Metrics.NetConnectionStates)
	w.netConnView.SetData(data.NetConnections, data.Metrics.NetConnections)
	w.netTabsView.Update(
		data.Metrics.NetConnections,
		data.Metrics.NetConnectionStates,
		data.Metrics.NetTopByProtocol,
		data.Metrics.NetTopByConnection,
	)
}
