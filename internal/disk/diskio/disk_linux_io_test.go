//go:build linux

package diskio

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/disk"
	"github.com/skushnerchuk/simda/internal/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

var (
	cfg = config.DaemonConfig{
		System:  config.SystemPoints{Proc: "/proc", Sys: "/sys"},
		Metrics: config.Metrics{DiskIO: true},
	}
	log = logger.NewSLogger(os.Stdout, "DEBUG")
)

func TestDiskIOStat(t *testing.T) { //nolint:funlen
	log.Disable()
	defer goleak.VerifyNone(t)

	t.Run("disk i/o: get uptime", func(t *testing.T) {
		v := NewLinuxDiskIOCollector(context.TODO(), context.TODO(), &cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(
			&LinuxDiskIOCollector{}, "readUptimeFile", func(_ string) ([]string, error) {
				return []string{"38716.67 581428.07"}, nil
			})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.getUptime()
		require.NoError(t, err)
		require.Equal(t, uint64(3871667), val.Uptime)
		require.Equal(t, uint64(58142807), val.IdleTime)
	})

	t.Run("disk i/o: IOCounters() ok", func(t *testing.T) {
		v := NewLinuxDiskIOCollector(context.TODO(), context.TODO(), &cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(
			&LinuxDiskIOCollector{}, "readStatFile", func(_ string) ([]string, error) {
				return []string{
					"868236 185895 87551823 476939 827290 693876 40173900 1215704 0 1280087 2128882 0 0 0 0 75276 436239",
				}, nil
			})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.IOCounters("stub")
		require.NoError(t, err)
		require.NotNil(t, val)

		require.Equal(t, uint64(868236), val.RdIos)
		require.Equal(t, uint64(827290), val.WrIos)
		require.Equal(t, uint64(0), val.DcIos)
		require.Equal(t, uint64(87551823), val.RdSectors)
		require.Equal(t, uint64(40173900), val.WrSectors)
	})

	t.Run("disk i/o: IOCounters() error read stat file", func(t *testing.T) {
		v := NewLinuxDiskIOCollector(context.TODO(), context.TODO(), &cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(
			&LinuxDiskIOCollector{}, "readStatFile", func(_ string) ([]string, error) {
				return nil, fmt.Errorf("error")
			})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.IOCounters("stub")
		require.Error(t, err)
		require.Nil(t, val)
	})

	t.Run("disk i/o: IOCounters() stat file incorrect data", func(t *testing.T) {
		v := NewLinuxDiskIOCollector(context.TODO(), context.TODO(), &cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(
			&LinuxDiskIOCollector{}, "readStatFile", func(_ string) ([]string, error) {
				return []string{"868236 185895 87551823"}, nil
			})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.IOCounters("stub")
		require.Error(t, err)
		require.Nil(t, val)
	})

	t.Run("disk i/o: GetDiskIOStat() ok", func(t *testing.T) {
		v := NewLinuxDiskIOCollector(context.TODO(), context.TODO(), &cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyMethodFunc(
			&LinuxDiskIOCollector{}, "IOCounters", func(_ string) (*IOCountersStatExp, error) {
				return &IOCountersStatExp{
					Tps:     1.1,
					RdSpeed: 2.2,
					WrSpeed: 3.3,
				}, nil
			})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.GetDiskIOStat("stub")
		require.NoError(t, err)
		require.NotNil(t, val)
		require.Equal(t, "stub", val.Name)
		require.Equal(t, 1.1, val.Tps)
		require.Equal(t, 2.2, val.RdSpeed)
		require.Equal(t, 3.3, val.WrSpeed)
	})

	t.Run("disk i/o: GetDiskIOStat() error", func(t *testing.T) {
		v := NewLinuxDiskIOCollector(context.TODO(), context.TODO(), &cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyMethodFunc(
			&LinuxDiskIOCollector{}, "IOCounters", func(_ string) (*IOCountersStatExp, error) {
				return nil, fmt.Errorf("error")
			})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.GetDiskIOStat("stub")
		require.Error(t, err)
		require.Nil(t, val)
	})

	t.Run("disk i/o: Run() error", func(t *testing.T) {
		cfg.Metrics.DiskIO = true
		v := NewLinuxDiskIOCollector(context.TODO(), context.TODO(), &cfg, log)
		patches := gomonkey.NewPatches()
		patches.ApplyFunc(disk.GetDevices, func() ([]string, error) {
			return nil, fmt.Errorf("error")
		})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, ch)
		require.Error(t, err)
		require.False(t, cfg.Metrics.DiskIO)
	})

	t.Run("disk i/o: metric disabled", func(t *testing.T) {
		called := 0

		cfg.Metrics.DiskIO = false
		patches := gomonkey.NewPatches()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		v := NewLinuxDiskIOCollector(ctx, ctx, &cfg, log)
		defer cancel()

		patches.ApplyMethodFunc(
			&LinuxDiskIOCollector{}, "GetDiskIOStat", func(_ string) (*disk.IOStat, error) {
				called++
				return &disk.IOStat{}, nil
			})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, err)
		require.NotNil(t, ch)

		val := <-ch
		require.Nil(t, val)

		require.Equal(t, 0, called)
		require.False(t, cfg.Metrics.DiskIO)
	})

	t.Run("disk i/o: metric enabled", func(t *testing.T) {
		called := 0

		cfg.Metrics.DiskIO = true
		patches := gomonkey.NewPatches()
		ctx, cancel := context.WithCancel(context.Background())
		v := NewLinuxDiskIOCollector(ctx, ctx, &cfg, log)
		defer cancel()

		patches.ApplyFunc(disk.GetDevices, func() ([]string, error) {
			return []string{"stub"}, nil
		})
		patches.ApplyMethodFunc(
			&LinuxDiskIOCollector{}, "GetDiskIOStat", func(_ string) (*disk.IOStat, error) {
				called++
				return &disk.IOStat{}, nil
			})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, err)
		require.NotNil(t, ch)

		val := <-ch
		require.NotNil(t, val)
		cancel()

		require.Equal(t, 1, called)
		require.True(t, cfg.Metrics.DiskIO)
	})
}
