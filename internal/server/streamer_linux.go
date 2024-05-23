//go:build linux

package server

import (
	"github.com/skushnerchuk/simda/internal/cpu/cpulinux"
	"github.com/skushnerchuk/simda/internal/disk/diskio"
	"github.com/skushnerchuk/simda/internal/disk/diskusage"
	loadAvg "github.com/skushnerchuk/simda/internal/load_avg"
	"github.com/skushnerchuk/simda/internal/network"
)

func (s *SnapshotStreamer) createLoadAvgCollector() {
	c := loadAvg.NewLinuxLoadAverageCollector(s.serverCtx, s.clientCtx, s.cfg, s.log)
	ch, err := c.Run()
	s.loadAvgChannel = ch
	if err != nil {
		s.log.Error("Failed to create load avg collector, metric disabled", "error", err.Error())
	}
}

func (s *SnapshotStreamer) createCPUCollector() {
	c := cpulinux.NewLinuxCPUCollector(s.serverCtx, s.clientCtx, s.cfg, s.log)
	ch, err := c.Run()
	s.cpuChannel = ch
	if err != nil {
		s.log.Error("Failed to create cpu avg collector, metric disabled", "error", err.Error())
	}
}

func (s *SnapshotStreamer) createDiskUsageCollector() {
	c := diskusage.NewLinuxDiskUsageCollector(s.serverCtx, s.clientCtx, s.cfg, s.log)
	ch, err := c.Run()
	s.diskUsageChannel = ch
	if err != nil {
		s.log.Error("Failed to create disk usage collector, metric disabled", "error", err.Error())
	}
}

func (s *SnapshotStreamer) createDiskIOCollector() {
	c := diskio.NewLinuxDiskIOCollector(s.serverCtx, s.clientCtx, s.cfg, s.log)
	ch, err := c.Run()
	s.diskIOChannel = ch
	if err != nil {
		s.log.Error("Failed to create disk i/o collector, metric disabled", "error", err.Error())
	}
}

func (s *SnapshotStreamer) createNetConnectionsCollector() {
	c := network.NewLinuxConnectionsCollector(s.serverCtx, s.clientCtx, s.cfg, s.log)
	ch, err := c.Run()
	s.netConnChannel = ch
	if err != nil {
		s.log.Error("Failed to create load network connections collector, metrics disabled", "error", err.Error())
	}
}

func (s *SnapshotStreamer) createNetPackagesCollector() {
	c := network.NewLinuxNetworkPackagesCollector(s.serverCtx, s.clientCtx, s.cfg, s.log, s.request)
	ch, err := c.Run()
	s.netPackagesChannel = ch
	if err != nil {
		s.log.Error("Failed to create net packages collector, metrics disabled", "error", err.Error())
		s.cfg.Metrics.NetTopByClients = false
		s.cfg.Metrics.NetTopByProtocol = false
	}
}
