package server

import (
	"context"
	"strconv"
	"time"

	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/cpu"
	"github.com/skushnerchuk/simda/internal/disk"
	loadAvg "github.com/skushnerchuk/simda/internal/load_avg"
	"github.com/skushnerchuk/simda/internal/logger"
	"github.com/skushnerchuk/simda/internal/network"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
)

type Streamer interface {
	Stream() <-chan *pb.Snapshot
	createCollectors()
	createLoadAvgCollector()
	createCPUCollector()
	createDiskUsageCollector()
	createDiskIOCollector()
	createNetConnectionsCollector()
	createNetPackagesCollector()
}

type SnapshotStreamer struct {
	ch        chan *pb.Snapshot
	serverCtx context.Context
	clientCtx context.Context
	request   *pb.Request
	log       logger.Logger
	cfg       *config.DaemonConfig

	loadAvgData        []*loadAvg.AvgStat
	cpuAvgData         []*cpu.Data
	diskUsageData      []disk.UsageStatMap
	diskIOData         []disk.IOStatMap
	netConnectionsData []network.ConnectionsStat
	netPackagesData    []network.NetworkPacketStat

	loadAvgChannel     <-chan *loadAvg.AvgStat
	cpuChannel         <-chan *cpu.Data
	diskUsageChannel   <-chan disk.UsageStatMap
	diskIOChannel      <-chan disk.IOStatMap
	netConnChannel     <-chan network.ConnectionsStat
	netPackagesChannel <-chan network.NetworkPacketStat
}

func NewSnapshotStreamer(
	serverCtx, clientCtx context.Context, request *pb.Request, log logger.Logger, cfg *config.DaemonConfig,
) *SnapshotStreamer {
	return &SnapshotStreamer{
		ch:          make(chan *pb.Snapshot),
		serverCtx:   serverCtx,
		clientCtx:   clientCtx,
		request:     request,
		log:         log,
		cfg:         cfg,
		loadAvgData: []*loadAvg.AvgStat{},
	}
}

func (s *SnapshotStreamer) createCollectors() {
	s.createLoadAvgCollector()
	s.createCPUCollector()
	s.createDiskUsageCollector()
	s.createDiskIOCollector()
	s.createNetConnectionsCollector()
	s.createNetPackagesCollector()
}

func (s *SnapshotStreamer) bufLen() int {
	return int(s.request.Warming)
}

func (s *SnapshotStreamer) appendLoadAvgData(data *loadAvg.AvgStat) {
	if len(s.loadAvgData) < s.bufLen() {
		s.loadAvgData = append(s.loadAvgData, data)
	}
}

func (s *SnapshotStreamer) appendCPUData(data *cpu.Data) {
	if len(s.cpuAvgData) < s.bufLen() {
		s.cpuAvgData = append(s.cpuAvgData, data)
	}
}

func (s *SnapshotStreamer) appendDiskUsageData(data disk.UsageStatMap) {
	if len(s.diskUsageData) < s.bufLen() {
		s.diskUsageData = append(s.diskUsageData, data)
	}
}

func (s *SnapshotStreamer) appendDiskIOData(data disk.IOStatMap) {
	if len(s.diskIOData) < s.bufLen() {
		s.diskIOData = append(s.diskIOData, data)
	}
}

func (s *SnapshotStreamer) appendNetConnectionsData(data network.ConnectionsStat) {
	if len(s.netConnectionsData) < s.bufLen() {
		s.netConnectionsData = append(s.netConnectionsData, data)
	}
}

func (s *SnapshotStreamer) appendNetPackagesData(data network.NetworkPacketStat) {
	if len(s.netPackagesData) < s.bufLen() {
		s.netPackagesData = append(s.netPackagesData, data)
	}
}

func (s *SnapshotStreamer) Stream() <-chan *pb.Snapshot {
	ch := make(chan *pb.Snapshot)
	ticker := time.NewTicker(500 * time.Millisecond)
	s.createCollectors()

	go func() {
		defer close(ch)
		for {
		L:
			for {
				select {
				case <-s.serverCtx.Done():
				case <-s.clientCtx.Done():
					s.log.Debug("snapshot collector stopped")
					return
				case value := <-s.loadAvgChannel:
					s.appendLoadAvgData(value)
				case value := <-s.cpuChannel:
					s.appendCPUData(value)
				case value := <-s.diskUsageChannel:
					s.appendDiskUsageData(value)
				case value := <-s.diskIOChannel:
					s.appendDiskIOData(value)
				case value := <-s.netConnChannel:
					s.appendNetConnectionsData(value)
				case value := <-s.netPackagesChannel:
					s.appendNetPackagesData(value)
				case <-ticker.C:
					if !s.warmingInProgress() {
						break L
					}
				}
			}
			ch <- s.createSnapshot()
			s.log.Debug("Snapshot sent to client")
			s.shiftBuffers()
		}
	}()
	return ch
}

func (s *SnapshotStreamer) calculateLoadAvg() *pb.LoadAverage {
	if !s.cfg.Metrics.LoadAvg {
		return nil
	}

	result := &pb.LoadAverage{
		One:     0,
		Five:    0,
		Fifteen: 0,
	}

	for _, stat := range s.loadAvgData {
		result.One += stat.Load1
		result.Five += stat.Load5
		result.Fifteen += stat.Load15
	}

	result.One /= float64(len(s.loadAvgData))
	result.Five /= float64(len(s.loadAvgData))
	result.Fifteen /= float64(len(s.loadAvgData))

	return result
}

func (s *SnapshotStreamer) calculateCPUAvg() *pb.CpuAverage {
	if !s.cfg.Metrics.CPUAvg {
		return nil
	}
	result := &pb.CpuAverage{
		User:   0,
		Idle:   0,
		System: 0,
	}

	for _, stat := range s.cpuAvgData {
		result.Idle += stat.Idle
		result.System += stat.System
		result.User += stat.User
	}

	result.User /= float64(len(s.cpuAvgData))
	result.Idle /= float64(len(s.cpuAvgData))
	result.System /= float64(len(s.cpuAvgData))

	return result
}

func (s *SnapshotStreamer) calculateDiskUsageAvg() []*pb.DiskUsage {
	if !s.cfg.Metrics.DiskUsage {
		return nil
	}
	avgData := make(disk.UsageStatMap)

	if len(s.diskUsageData) == 0 {
		return nil
	}

	for k, v := range s.diskUsageData[0] {
		avgData[k] = &disk.UsageStat{Mountpoint: v.Mountpoint, Device: v.Device}
	}

	for _, item := range s.diskUsageData {
		for k, v := range item {
			avgData[k].UsagePercent += v.UsagePercent
			avgData[k].Usage += v.Usage
			avgData[k].INodeAvailablePercent += v.INodeAvailablePercent
			avgData[k].INodeCount += v.INodeCount
		}
	}

	for k, v := range avgData {
		avgData[k].UsagePercent = v.UsagePercent / float64(len(s.diskUsageData))
		avgData[k].Usage = v.Usage / float64(len(s.diskUsageData))
		avgData[k].INodeAvailablePercent = v.INodeAvailablePercent / float64(len(s.diskUsageData))
		avgData[k].INodeCount = v.INodeCount / float64(len(s.diskUsageData))
	}

	result := make([]*pb.DiskUsage, 0)

	for _, v := range avgData {
		result = append(result, &pb.DiskUsage{
			Device:                v.Device,
			MountPoint:            v.Mountpoint,
			UsagePercent:          v.UsagePercent,
			Usage:                 v.Usage,
			InodeAvailablePercent: v.INodeAvailablePercent,
			InodeCount:            v.INodeCount,
		})
	}

	return result
}

func (s *SnapshotStreamer) calculateDiskIOAvg() []*pb.DiskIO {
	if !s.cfg.Metrics.DiskIO {
		return nil
	}
	avgData := make(map[string]*disk.IOStat)

	if len(s.diskIOData) == 0 {
		return nil
	}

	for k, v := range s.diskIOData[0] {
		avgData[k] = &disk.IOStat{Name: v.Name}
	}

	for _, item := range s.diskIOData {
		for k, v := range item {
			avgData[k].Tps += v.Tps
			avgData[k].RdSpeed += v.RdSpeed
			avgData[k].WrSpeed += v.WrSpeed
		}
	}

	for k, v := range avgData {
		avgData[k].Tps = v.Tps / float64(len(s.diskIOData))
		avgData[k].RdSpeed = v.RdSpeed / float64(len(s.diskIOData))
		avgData[k].WrSpeed = v.WrSpeed / float64(len(s.diskIOData))
	}

	result := make([]*pb.DiskIO, 0)

	for _, v := range avgData {
		result = append(result, &pb.DiskIO{
			Name:    v.Name,
			Tps:     v.Tps,
			RdSpeed: v.RdSpeed,
			WrSpeed: v.WrSpeed,
		})
	}

	return result
}

func (s *SnapshotStreamer) calculateNetworkConnectionsAvg() []*pb.NetConnection {
	if !s.cfg.Metrics.NetConnections {
		return nil
	}
	avgData := make(map[string]*network.Connection)

	for _, item := range s.netConnectionsData {
		for _, v := range item {
			v := v
			avgData[v.SocketID] = &v
		}
	}

	result := make([]*pb.NetConnection, 0)

	for _, v := range avgData {
		item := &pb.NetConnection{
			Protocol: v.Protocol,
			User:     v.User,
			State:    v.State,
			UserId:   v.UserID,
		}
		if v.Process != nil {
			item.Process = &pb.Process{Pid: uint32(v.Process.Pid), CmdLine: v.Process.CmdLine}
		}
		if v.LocalAddress != nil {
			item.LocalAddr = &pb.SockAddr{Ip: v.LocalAddress.IP.String(), Port: uint32(v.LocalAddress.Port)}
		}
		if v.ForeignAddress != nil {
			item.ForeignAddr = &pb.SockAddr{Ip: v.ForeignAddress.IP.String(), Port: uint32(v.ForeignAddress.Port)}
		}
		result = append(result, item)
	}

	return result
}

func (s *SnapshotStreamer) calculateNetworkConnectionsStatesAvg() []*pb.NetConnectionStates {
	if !s.cfg.Metrics.NetConnectionsStates {
		return nil
	}
	avgData := make(map[string]*network.Connection)

	for _, item := range s.netConnectionsData {
		for _, v := range item {
			v := v
			avgData[v.SocketID] = &v
		}
	}

	states := make(map[string]uint32)

	for _, v := range avgData {
		states[v.State]++
	}
	result := make([]*pb.NetConnectionStates, 0)
	for k, v := range states {
		result = append(result, &pb.NetConnectionStates{State: k, Count: v})
	}
	return result
}

func (s *SnapshotStreamer) CalcProtocolStat() []*pb.NetTopByProtocol {
	if !s.cfg.Metrics.NetTopByProtocol {
		return nil
	}

	protocols := make(map[string]uint64)
	totalBytes := uint64(0)
	for _, elem := range s.netPackagesData {
		for _, p := range elem {
			totalBytes += p.PayloadSize
			protocols[p.Protocol] += p.PayloadSize
		}
	}

	result := make([]*pb.NetTopByProtocol, 0)

	for protocol, bytes := range protocols {
		result = append(result, &pb.NetTopByProtocol{
			Protocol: protocol,
			Bytes:    bytes,
			Percent:  (float64(bytes) / float64(totalBytes)) * 100.0,
		})
	}
	return result
}

func (s *SnapshotStreamer) CalcProtocolConnectionStat() []*pb.NetTopByConnection {
	if !s.cfg.Metrics.NetTopByClients {
		return nil
	}
	connections := make(map[string][]network.PacketInfo)
	for _, elem := range s.netPackagesData {
		for _, p := range elem {
			if _, ok := connections[p.ConnectionID()]; ok {
				connections[p.ConnectionID()] = append(connections[p.ConnectionID()], p)
			} else {
				connections[p.ConnectionID()] = make([]network.PacketInfo, 0)
				connections[p.ConnectionID()] = append(connections[p.ConnectionID()], p)
			}
		}
	}

	result := make([]*pb.NetTopByConnection, 0)

	for _, packets := range connections {
		sourceAddr := &pb.SockAddr{Ip: packets[0].SourceIP}
		port, _ := strconv.Atoi(packets[0].SourcePort)
		sourceAddr.Port = uint32(port)

		destinationAddr := &pb.SockAddr{Ip: packets[0].DestinationIP}
		port, _ = strconv.Atoi(packets[0].DestinationPort)
		destinationAddr.Port = uint32(port)

		bytes := uint64(0)
		for _, packet := range packets {
			bytes += packet.PayloadSize
		}
		percent := 0.0
		if bytes > 0 {
			percent = (float64(s.request.Warming) / float64(bytes)) * 100.0
		}

		result = append(result, &pb.NetTopByConnection{
			Protocol:        packets[0].Protocol,
			Bytes:           bytes,
			Percent:         percent,
			SourceAddr:      sourceAddr,
			DestinationAddr: destinationAddr,
		})
	}
	return result
}

func (s *SnapshotStreamer) warmingInProgress() bool {
	bufLen := int(s.request.Warming)

	if (s.cfg.Metrics.LoadAvg && len(s.loadAvgData) < bufLen) &&
		(s.cfg.Metrics.CPUAvg && len(s.cpuAvgData) < bufLen) &&
		(s.cfg.Metrics.DiskUsage && len(s.diskUsageData) < bufLen) &&
		(s.cfg.Metrics.DiskIO && len(s.diskIOData) < bufLen) &&
		((s.cfg.Metrics.NetConnections || s.cfg.Metrics.NetConnectionsStates) && len(s.netConnectionsData) < bufLen) &&
		((s.cfg.Metrics.NetTopByClients || s.cfg.Metrics.NetTopByProtocol) && len(s.netPackagesData) < bufLen) {
		return true
	}

	return false
}

func (s *SnapshotStreamer) shiftBuffers() {
	p := int(s.request.Period)

	if len(s.loadAvgData) >= p {
		s.loadAvgData = s.loadAvgData[p:]
	}
	if len(s.cpuAvgData) >= p {
		s.cpuAvgData = s.cpuAvgData[p:]
	}
	if len(s.diskUsageData) >= p {
		s.diskUsageData = s.diskUsageData[p:]
	}
	if len(s.diskIOData) >= p {
		s.diskIOData = s.diskIOData[p:]
	}
	if len(s.netConnectionsData) >= p {
		s.netConnectionsData = s.netConnectionsData[p:]
	}
	if len(s.netPackagesData) >= p {
		s.netPackagesData = s.netPackagesData[p:]
	}
}

func (s *SnapshotStreamer) createSnapshot() *pb.Snapshot {
	snapshot := &pb.Snapshot{}
	snapshot.Metrics = &pb.EnabledMetrics{
		LoadAvg:             s.cfg.Metrics.LoadAvg,
		CpuAvg:              s.cfg.Metrics.CPUAvg,
		DiskIO:              s.cfg.Metrics.DiskIO,
		DiskUsage:           s.cfg.Metrics.DiskUsage,
		NetConnections:      s.cfg.Metrics.NetConnections,
		NetConnectionStates: s.cfg.Metrics.NetConnectionsStates,
		NetTopByProtocol:    s.cfg.Metrics.NetTopByProtocol,
		NetTopByConnection:  s.cfg.Metrics.NetTopByClients,
	}
	snapshot.LoadAvg = s.calculateLoadAvg()
	snapshot.CpuAvg = s.calculateCPUAvg()
	snapshot.DiskUsage = s.calculateDiskUsageAvg()
	snapshot.DiskIO = s.calculateDiskIOAvg()
	snapshot.NetConnections = s.calculateNetworkConnectionsAvg()
	snapshot.NetConnectionsStates = s.calculateNetworkConnectionsStatesAvg()
	snapshot.NetTopByProtocol = s.CalcProtocolStat()
	snapshot.NetTopByConnection = s.CalcProtocolConnectionStat()
	return snapshot
}
