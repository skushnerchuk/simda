//go:build darwin

package server

import loadAvg "github.com/skushnerchuk/simda/internal/load_avg"

func (s *SnapshotStreamer) createLoadAvgCollector() {
	c := loadAvg.NewDarwinLoadAverageCollector(s.serverCtx, s.clientCtx, s.cfg, s.log)
	ch, err := c.Run()
	s.loadAvgChannel = ch
	if err != nil {
		s.log.Error("Failed to create load avg collector, metric disabled", "error", err.Error())
	}
}

func (s *SnapshotStreamer) createCPUCollector() {
	s.cfg.Metrics.CPUAvg = false
}

func (s *SnapshotStreamer) createDiskUsageCollector() {
	s.cfg.Metrics.DiskUsage = false
}

func (s *SnapshotStreamer) createDiskIOCollector() {
	s.cfg.Metrics.DiskIO = false
}

func (s *SnapshotStreamer) createNetConnectionsCollector() {
	s.cfg.Metrics.NetConnections = false
	s.cfg.Metrics.NetConnectionsStates = false
}

func (s *SnapshotStreamer) createNetPackagesCollector() {
	s.cfg.Metrics.NetTopByProtocol = false
	s.cfg.Metrics.NetTopByClients = false
}
