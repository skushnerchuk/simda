//go:build linux

package cpulinux

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/cpu"
	"github.com/skushnerchuk/simda/internal/logger"
	"github.com/skushnerchuk/simda/internal/utils"
)

var ClocksPerSec = float64(100)

type TimesStat struct {
	CPU             string
	User            float64
	System          float64
	Idle            float64
	UserInPercent   float64
	SystemInPercent float64
	IdleInPercent   float64
	Nice            float64
	Iowait          float64
	HardIrq         float64
	Softirq         float64
	Steal           float64
	Guest           float64
	GuestNice       float64
}

type LinuxCPUCollector struct {
	serverCtx context.Context
	clientCtx context.Context
	cfg       *config.DaemonConfig
	l         logger.Logger

	CPU             string
	User            float64
	System          float64
	Idle            float64
	UserInPercent   float64
	SystemInPercent float64
	IdleInPercent   float64
	Nice            float64
	Iowait          float64
	HardIrq         float64
	Softirq         float64
	Steal           float64
	Guest           float64
	GuestNice       float64
}

func NewLinuxCPUCollector(
	serverCtx, clientCtx context.Context, cfg *config.DaemonConfig, l logger.Logger,
) *LinuxCPUCollector {
	return &LinuxCPUCollector{
		serverCtx: serverCtx,
		clientCtx: clientCtx,
		cfg:       cfg,
		l:         l,
	}
}

func (l *LinuxCPUCollector) Get() (*cpu.Data, error) {
	filename := filepath.Join(l.cfg.System.Proc, "stat")
	var lines []string
	lines, err := utils.ReadLinesOffsetN(filename, 0, 1)
	if len(lines) == 0 || err != nil {
		return nil, errors.New("no cpu stat")
	}
	err = l.parseStatLine(lines[0])
	if err != nil {
		return nil, err
	}
	return &cpu.Data{
		User:   l.UserInPercent,
		Idle:   l.IdleInPercent,
		System: l.SystemInPercent,
	}, nil
}

func (l *LinuxCPUCollector) parseStatLine(line string) error {
	fields := strings.Fields(line)

	if len(fields) < 8 {
		return errors.New("stat does not contain cpu info")
	}

	if !strings.HasPrefix(fields[0], "cpu") {
		return errors.New("not contain cpu")
	}

	// Смотрим суммарную нагрузку по всем ядрам
	cpuID := fields[0]
	user, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return err
	}
	nice, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return err
	}
	system, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return err
	}
	idle, err := strconv.ParseFloat(fields[4], 64)
	if err != nil {
		return err
	}
	iowait, err := strconv.ParseFloat(fields[5], 64)
	if err != nil {
		return err
	}
	irq, err := strconv.ParseFloat(fields[6], 64)
	if err != nil {
		return err
	}
	softirq, err := strconv.ParseFloat(fields[7], 64)
	if err != nil {
		return err
	}

	if len(fields) > 8 { // Linux >= 2.6.11
		steal, err := strconv.ParseFloat(fields[8], 64)
		if err != nil {
			return err
		}
		l.Steal = steal / ClocksPerSec
	}
	if len(fields) > 9 { // Linux >= 2.6.24
		guest, err := strconv.ParseFloat(fields[9], 64)
		if err != nil {
			return err
		}
		l.Guest = guest / ClocksPerSec
	}
	if len(fields) > 10 { // Linux >= 3.2.0
		guestNice, err := strconv.ParseFloat(fields[10], 64)
		if err != nil {
			return err
		}
		l.GuestNice = guestNice / ClocksPerSec
	}
	l.CPU = cpuID
	l.User = user / ClocksPerSec
	l.Nice = nice / ClocksPerSec
	l.System = system / ClocksPerSec
	l.Idle = idle / ClocksPerSec
	l.Iowait = iowait / ClocksPerSec
	l.HardIrq = irq / ClocksPerSec
	l.Softirq = softirq / ClocksPerSec

	jiff := l.User + l.System + l.Nice + l.System + l.Idle + l.Iowait + l.HardIrq + l.Softirq + l.Steal

	l.UserInPercent = l.User / jiff * 100
	l.SystemInPercent = l.System / jiff * 100
	l.IdleInPercent = l.Idle / jiff * 100
	return nil
}

func (l *LinuxCPUCollector) Run() (<-chan *cpu.Data, error) {
	ch := make(chan *cpu.Data)
	ticker := time.NewTicker(time.Second)
	if _, err := l.Get(); err != nil {
		l.cfg.Metrics.CPUAvg = false
		return nil, err
	}

	go func() {
		defer close(ch)
		for {
			select {
			case <-l.serverCtx.Done():
			case <-l.clientCtx.Done():
				l.l.Debug("cpu average collector stopped")
				return
			case <-ticker.C:
				if !l.cfg.Metrics.CPUAvg {
					continue
				}
				stat, err := l.Get()
				if err != nil {
					l.l.Error("cpu average collector error", "error", err.Error())
					l.cfg.Metrics.CPUAvg = false
					return
				}
				ch <- stat
			}
		}
	}()
	return ch, nil
}
