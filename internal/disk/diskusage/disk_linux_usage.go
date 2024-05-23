//go:build linux

package diskusage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/disk"
	"github.com/skushnerchuk/simda/internal/logger"
	"github.com/skushnerchuk/simda/internal/utils"
	"golang.org/x/sys/unix"
)

type LinuxUsageStat struct {
	Path              string  `json:"path"`
	FsType            string  `json:"fstype"`
	Total             uint64  `json:"total"`
	Free              uint64  `json:"free"`
	Used              uint64  `json:"used"`
	UsedPercent       float64 `json:"usedPercent"`
	InodesTotal       uint64  `json:"inodesTotal"`
	InodesUsed        uint64  `json:"inodesUsed"`
	InodesFree        uint64  `json:"inodesFree"`
	InodesUsedPercent float64 `json:"inodesUsedPercent"`
}

type PartitionStat struct {
	Device     string   `json:"device"`
	Mountpoint string   `json:"mountpoint"`
	FsType     string   `json:"fstype"`
	Opts       []string `json:"opts"`
}

type LinuxDiskUsageCollector struct {
	serverCtx context.Context
	clientCtx context.Context
	cfg       *config.DaemonConfig
	l         logger.Logger
}

func NewLinuxDiskUsageCollector(
	serverCtx, clientCtx context.Context, cfg *config.DaemonConfig, l logger.Logger,
) *LinuxDiskUsageCollector {
	return &LinuxDiskUsageCollector{
		serverCtx: serverCtx,
		clientCtx: clientCtx,
		cfg:       cfg,
		l:         l,
	}
}

func (l *LinuxDiskUsageCollector) readMountFile(root string) (
	lines []string, useMounts bool, filename string, err error,
) {
	filename = path.Join(root, "mountinfo")
	lines, err = utils.ReadLines(filename)
	if err != nil {
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) {
			return
		}
		useMounts = true
		filename = path.Join(root, "mounts")
		lines, err = utils.ReadLines(filename)
		if err != nil {
			return
		}
		return
	}
	return
}

func (l *LinuxDiskUsageCollector) readFileSystems() ([]string, error) {
	filename := filepath.Join(l.cfg.System.Proc, "filesystems")
	lines, err := utils.ReadLines(filename)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func (l *LinuxDiskUsageCollector) getFileSystems() ([]string, error) {
	lines, err := l.readFileSystems()
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0)
	for _, line := range lines {
		if !strings.HasPrefix(line, "nodev") {
			ret = append(ret, strings.TrimSpace(line))
			continue
		}
		t := strings.Split(line, "\t")
		if len(t) != 2 || t[1] != "zfs" {
			continue
		}
		ret = append(ret, strings.TrimSpace(t[1]))
	}

	return ret, nil
}

func (l *LinuxDiskUsageCollector) GetDiskUsageStat(device, mountpoint string) (*disk.UsageStat, error) {
	stat := unix.Statfs_t{}
	err := unix.Statfs(mountpoint, &stat)
	if err != nil {
		return nil, err
	}
	blockSize := stat.Bsize
	free := stat.Bavail * uint64(blockSize)
	inodesTotal := stat.Files
	inodesFree := stat.Ffree

	ret := &disk.UsageStat{
		Device:     device,
		Mountpoint: disk.UnescapeFstab(mountpoint),
		Type:       disk.GetFsType(stat),
		INodeCount: float64(inodesTotal),
	}

	ret.Usage = float64((stat.Blocks - stat.Bfree) * uint64(blockSize))

	if (ret.Usage + float64(free)) == 0 {
		ret.UsagePercent = 0
	} else {
		ret.UsagePercent = (ret.Usage / (ret.Usage + float64(free))) * 100.0
	}

	if inodesTotal < inodesFree {
		return ret, nil
	}

	inodesUsed := inodesTotal - inodesFree

	if inodesTotal == 0 {
		ret.INodeAvailablePercent = 0
	} else {
		ret.INodeAvailablePercent = 100 - (float64(inodesUsed)/float64(inodesTotal))*100.0
	}

	return ret, nil
}

func (l *LinuxDiskUsageCollector) Partitions(all bool) ([]PartitionStat, error) { //nolint:gocognit
	root := filepath.Join(l.cfg.System.Proc, path.Join("1"))
	hpmPath := l.cfg.System.ProcMountInfo
	if hpmPath != "" {
		root = filepath.Dir(hpmPath)
	}

	lines, useMounts, filename, err := l.readMountFile(root)
	if err != nil {
		if hpmPath != "" { // don't fallback with HOST_PROC_MOUNTINFO
			return nil, err
		}
		selfPath := filepath.Join(l.cfg.System.Proc, path.Join("self"))
		lines, useMounts, filename, err = l.readMountFile(selfPath)
		if err != nil {
			return nil, err
		}
	}

	fs, err := l.getFileSystems()
	if err != nil && !all {
		return nil, err
	}

	ret := make([]PartitionStat, 0, len(lines))

	for _, line := range lines {
		var d PartitionStat
		if useMounts { //nolint:nestif
			fields := strings.Fields(line)

			d = PartitionStat{
				Device:     fields[0],
				Mountpoint: disk.UnescapeFstab(fields[1]),
				FsType:     fields[2],
				Opts:       strings.Fields(fields[3]),
			}

			if !all {
				if d.Device == "none" || !utils.StringsHas(fs, d.FsType) {
					continue
				}
			}
		} else {
			parts := strings.Split(line, " - ")
			if len(parts) != 2 {
				return nil, fmt.Errorf("found invalid mountinfo line in file %s: %s ", filename, line)
			}

			fields := strings.Fields(parts[0])
			blockDeviceID := fields[2]
			mountPoint := fields[4]
			mountOpts := strings.Split(fields[5], ",")

			if rootDir := fields[3]; rootDir != "" && rootDir != "/" {
				mountOpts = append(mountOpts, "bind")
			}

			fields = strings.Fields(parts[1])
			fsType := fields[0]
			device := fields[1]

			d = PartitionStat{
				Device:     device,
				Mountpoint: disk.UnescapeFstab(mountPoint),
				FsType:     fsType,
				Opts:       mountOpts,
			}

			if !all {
				if d.Device == "none" || !utils.StringsHas(fs, d.FsType) {
					continue
				}
			}

			if strings.HasPrefix(d.Device, "/dev/mapper/") {
				devPath, err := filepath.EvalSymlinks(filepath.Join(l.cfg.System.Dev, strings.Replace(d.Device, "/dev", "", 1)))
				if err == nil {
					d.Device = devPath
				}
			}

			if d.Device == "/dev/root" {
				devPath, err := os.Readlink(filepath.Join(l.cfg.System.Sys, "/dev/block/"+blockDeviceID))
				if err == nil {
					d.Device = strings.Replace(d.Device, "root", filepath.Base(devPath), 1)
				}
			}
		}
		ret = append(ret, d)
	}

	return ret, nil
}

func (l *LinuxDiskUsageCollector) Run() (<-chan disk.UsageStatMap, error) {
	devices, err := l.Partitions(false)
	if err != nil {
		l.cfg.Metrics.DiskUsage = false
		return nil, err
	}
	ch := make(chan disk.UsageStatMap)
	ticker := time.NewTicker(time.Second)

	go func() {
		defer close(ch)
		for {
			select {
			case <-l.serverCtx.Done():
			case <-l.clientCtx.Done():
				l.l.Debug("disk usage collector stopped")
				return
			case <-ticker.C:
				if !l.cfg.Metrics.DiskUsage {
					continue
				}
				message := make(disk.UsageStatMap)
				for _, device := range devices {
					stat, err := l.GetDiskUsageStat(device.Device, device.Mountpoint)
					if err != nil {
						l.l.Error(
							"failed to get disk usage stat", "error", err.Error(), "path", device.Mountpoint,
						)
						l.cfg.Metrics.DiskUsage = false
						continue
					}
					message[device.Mountpoint] = stat
				}
				ch <- message
			}
		}
	}()
	return ch, nil
}
