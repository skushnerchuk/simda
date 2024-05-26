//go:build linux

package diskio

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/disk"
	"github.com/skushnerchuk/simda/internal/logger"
	"github.com/skushnerchuk/simda/internal/utils"
)

type IOCountersStatExp struct {
	RdSectors uint64
	WrSectors uint64
	DcSectors uint64
	RdIos     uint64
	RdMerges  uint64
	WrIos     uint64
	WrMerges  uint64
	DcIos     uint64
	DcMerges  uint64
	FlIos     uint64
	RdTicks   uint
	WrTicks   uint
	DcTicks   uint
	FlTicks   uint
	IosPgr    uint
	TotTicks  uint
	RqTicks   uint
	Tps       float64
	RdSpeed   float64
	WrSpeed   float64
}

type Uptime struct {
	Uptime   uint64
	IdleTime uint64
}

func (l *LinuxDiskIOCollector) readStatFile(devName string) ([]string, error) {
	statFile := fmt.Sprintf("%s/block/%s/stat", l.cfg.System.Sys, devName)
	lines, err := utils.ReadLines(statFile)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func (l *LinuxDiskIOCollector) IOCounters(devName string) (*IOCountersStatExp, error) {
	lines, err := l.readStatFile(devName)
	if err != nil {
		return nil, err
	}
	ret := IOCountersStatExp{}

	fields := strings.Fields(lines[0])

	if len(fields) >= 11 { //nolint:gocritic
		ret.RdIos, _ = strconv.ParseUint(fields[0], 10, 64)
		ret.RdMerges, _ = strconv.ParseUint(fields[1], 10, 64)
		ret.RdSectors, _ = strconv.ParseUint(fields[2], 10, 64)
		v, _ := strconv.ParseUint(fields[3], 10, 64)
		ret.RdTicks = uint(v)
		ret.WrIos, _ = strconv.ParseUint(fields[4], 10, 64)
		ret.WrMerges, _ = strconv.ParseUint(fields[5], 10, 64)
		ret.WrSectors, _ = strconv.ParseUint(fields[6], 10, 64)
		v, _ = strconv.ParseUint(fields[7], 10, 64)
		ret.WrTicks = uint(v)
		v, _ = strconv.ParseUint(fields[8], 10, 64)
		ret.IosPgr = uint(v)
		v, _ = strconv.ParseUint(fields[9], 10, 64)
		ret.TotTicks = uint(v)
		v, _ = strconv.ParseUint(fields[10], 10, 64)
		ret.RqTicks = uint(v)

		if len(fields) >= 15 {
			ret.DcIos, _ = strconv.ParseUint(fields[11], 10, 64)
			ret.DcMerges, _ = strconv.ParseUint(fields[12], 10, 64)
			ret.DcSectors, _ = strconv.ParseUint(fields[13], 10, 64)
			v, _ = strconv.ParseUint(fields[14], 10, 64)
			ret.DcTicks = uint(v)
		}
		if len(fields) >= 17 {
			ret.FlIos, _ = strconv.ParseUint(fields[15], 10, 64)
			v, _ = strconv.ParseUint(fields[16], 10, 64)
			ret.FlTicks = uint(v)
		}
	} else if len(fields) == 4 {
		ret.RdIos, _ = strconv.ParseUint(fields[0], 10, 64)
		ret.RdSectors, _ = strconv.ParseUint(fields[1], 10, 64)
		ret.WrIos, _ = strconv.ParseUint(fields[2], 10, 64)
		ret.WrSectors, _ = strconv.ParseUint(fields[3], 10, 64)
	} else {
		return nil, fmt.Errorf("unexpected number of fields: %d", len(fields))
	}
	u, _ := l.getUptime()

	val := float64(ret.RdIos + ret.WrIos + ret.DcIos)
	ret.Tps = val / float64(u.Uptime) * 100.0

	ret.RdSpeed = float64(ret.RdSectors/2) / float64(u.Uptime) * 100.0
	ret.WrSpeed = float64(ret.WrSectors/2) / float64(u.Uptime) * 100.0

	return &ret, nil
}

func (l *LinuxDiskIOCollector) GetDiskIOStat(device string) (*disk.IOStat, error) {
	stat, err := l.IOCounters(device)
	if err != nil {
		return nil, err
	}
	return &disk.IOStat{
		Name:    device,
		Tps:     stat.Tps,
		RdSpeed: stat.RdSpeed,
		WrSpeed: stat.WrSpeed,
	}, nil
}

func (l *LinuxDiskIOCollector) readUptimeFile() ([]string, error) {
	uptimeFile := filepath.Join(l.cfg.System.Proc, path.Join("uptime"))
	lines, err := utils.ReadLines(uptimeFile)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func (l *LinuxDiskIOCollector) getUptime() (*Uptime, error) {
	lines, err := l.readUptimeFile()
	if err != nil {
		return nil, err
	}
	ret := Uptime{}

	fields := strings.Fields(lines[0])
	if len(fields) != 2 {
		return &ret, nil
	}

	fields[0] = strings.ReplaceAll(fields[0], ".", "")
	fields[1] = strings.ReplaceAll(fields[1], ".", "")

	f, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return nil, err
	}
	ret.Uptime = f

	f, err = strconv.ParseUint(fields[1], 10, 64)
	ret.IdleTime = f
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

type LinuxDiskIOCollector struct {
	serverCtx context.Context
	clientCtx context.Context
	cfg       *config.DaemonConfig
	l         logger.Logger
}

func NewLinuxDiskIOCollector(
	serverCtx, clientCtx context.Context, cfg *config.DaemonConfig, l logger.Logger,
) *LinuxDiskIOCollector {
	return &LinuxDiskIOCollector{
		serverCtx: serverCtx,
		clientCtx: clientCtx,
		cfg:       cfg,
		l:         l,
	}
}

func (l *LinuxDiskIOCollector) Run() (<-chan disk.IOStatMap, error) {
	devices, err := disk.GetDevices(l.cfg.System.Sys)
	if err != nil {
		l.cfg.Metrics.DiskIO = false
		return nil, err
	}
	ch := make(chan disk.IOStatMap)
	ticker := time.NewTicker(time.Second)

	go func() {
		defer close(ch)
		for {
			select {
			case <-l.serverCtx.Done():
			case <-l.clientCtx.Done():
				l.l.Debug("disk i/o collector stopped")
				return
			case <-ticker.C:
				if !l.cfg.Metrics.DiskIO {
					continue
				}
				message := make(map[string]*disk.IOStat, len(devices))
				for _, device := range devices {
					stat, err := l.GetDiskIOStat(device)
					if err != nil {
						l.l.Error("failed to get disk i/o stat", "error", err.Error(), "device", device)
						continue
					}
					message[device] = stat
				}
				ch <- message
			}
		}
	}()
	return ch, nil
}
