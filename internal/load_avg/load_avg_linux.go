//go:build linux

package loadavg

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/logger"
)

type LinuxLoadAverageCollector struct {
	serverCtx context.Context
	clientCtx context.Context
	cfg       *config.DaemonConfig
	l         logger.Logger
}

func NewLinuxLoadAverageCollector(
	serverCtx, clientCtx context.Context, cfg *config.DaemonConfig, l logger.Logger,
) *LinuxLoadAverageCollector {
	return &LinuxLoadAverageCollector{
		serverCtx: serverCtx,
		clientCtx: clientCtx,
		cfg:       cfg,
		l:         l,
	}
}

func (l *LinuxLoadAverageCollector) Run() (<-chan *AvgStat, error) {
	if _, err := l.Get(); err != nil {
		l.l.Error("load average collector error", "error", err.Error())
		l.cfg.Metrics.LoadAvg = false
		return nil, err
	}
	ch := make(chan *AvgStat)
	ticker := time.NewTicker(time.Second)

	go func() {
		defer close(ch)
		for {
			select {
			case <-l.serverCtx.Done():
			case <-l.clientCtx.Done():
				l.l.Debug("load average collector stopped")
				return
			case <-ticker.C:
				if !l.cfg.Metrics.LoadAvg {
					continue
				}
				stat, err := l.Get()
				if err != nil {
					l.l.Error("load average collector error", "error", err.Error())
					l.cfg.Metrics.LoadAvg = false
					return
				}
				ch <- stat
			}
		}
	}()
	return ch, nil
}

func (l *LinuxLoadAverageCollector) Get() (*AvgStat, error) {
	values, err := l.readLoadAvgFromFile()
	if err != nil {
		return nil, err
	}

	load1, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return nil, err
	}
	load5, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		return nil, err
	}
	load15, err := strconv.ParseFloat(values[2], 64)
	if err != nil {
		return nil, err
	}

	ret := &AvgStat{
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}

	return ret, nil
}

func (l *LinuxLoadAverageCollector) readLoadAvgFromFile() ([]string, error) {
	loadavgFilename := filepath.Join(l.cfg.System.Proc, "loadavg")
	line, err := os.ReadFile(loadavgFilename)
	if err != nil {
		return nil, err
	}

	values := strings.Fields(string(line))
	return values, nil
}
