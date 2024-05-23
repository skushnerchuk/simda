//go:build darwin

package loadavg

import (
	"context"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/logger"
)

type DarwinLoadAverageCollector struct {
	serverCtx context.Context
	clientCtx context.Context
	cfg       *config.DaemonConfig
	l         logger.Logger
}

func NewDarwinLoadAverageCollector(
	serverCtx, clientCtx context.Context, cfg *config.DaemonConfig, l logger.Logger,
) *DarwinLoadAverageCollector {
	return &DarwinLoadAverageCollector{
		serverCtx: serverCtx,
		clientCtx: clientCtx,
		cfg:       cfg,
		l:         l,
	}
}

func (l *DarwinLoadAverageCollector) Run() (<-chan *AvgStat, error) {
	_, err := l.Get()
	if err != nil {
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
				l.l.Info("load average", "load1", stat.Load1, "load5", stat.Load5, "load15", stat.Load15)
				ch <- stat
			}
		}
	}()
	return ch, nil
}

func (l *DarwinLoadAverageCollector) Get() (*AvgStat, error) {
	type loadavg struct {
		load  [3]uint32
		scale int
	}
	b, err := unix.SysctlRaw("vm.loadavg")
	if err != nil {
		return nil, err
	}
	load := *(*loadavg)(unsafe.Pointer((&b[0])))
	scale := float64(load.scale)
	ret := &AvgStat{
		Load1:  float64(load.load[0]) / scale,
		Load5:  float64(load.load[1]) / scale,
		Load15: float64(load.load[2]) / scale,
	}

	return ret, nil
}
