//go:build linux

package diskusage

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

var log = logger.NewSLogger(os.Stdout, "DEBUG")

func createConfig() *config.DaemonConfig {
	return &config.DaemonConfig{
		System:  config.SystemPoints{Proc: "/proc", Sys: "/sys", ProcMountInfo: ""},
		Metrics: config.Metrics{DiskUsage: true},
	}
}

func TestDiskUsageStat(t *testing.T) {
	log.Disable()
	defer goleak.VerifyNone(t)

	t.Run("disk usage: get filesystems", func(t *testing.T) {
		cfg := createConfig()
		v := NewLinuxDiskUsageCollector(context.TODO(), context.TODO(), cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(
			&LinuxDiskUsageCollector{}, "readFileSystems", func() ([]string, error) {
				return []string{"ext3", "nodev pstore", "btrfs"}, nil
			})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.getFileSystems()
		require.NoError(t, err)
		require.Len(t, val, 2)
	})

	t.Run("disk usage: Run() error", func(t *testing.T) {
		cfg := createConfig()
		v := NewLinuxDiskUsageCollector(context.TODO(), context.TODO(), cfg, log)
		patches := gomonkey.NewPatches()
		patches.ApplyMethod(
			&LinuxDiskUsageCollector{}, "Partitions", func() ([]string, error) {
				return nil, fmt.Errorf("error")
			})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, ch)
		require.Error(t, err)
		require.False(t, cfg.Metrics.DiskUsage)
	})

	t.Run("disk usage: metric disabled", func(t *testing.T) {
		called := 0
		cfg := createConfig()
		cfg.Metrics.DiskUsage = false
		patches := gomonkey.NewPatches()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		v := NewLinuxDiskUsageCollector(ctx, ctx, cfg, log)
		defer cancel()

		patches.ApplyMethodFunc(
			&LinuxDiskUsageCollector{}, "GetDiskUsageStat", func(_, _ string) (*disk.UsageStat, error) {
				called++
				return &disk.UsageStat{}, nil
			})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, err)
		require.NotNil(t, ch)

		val := <-ch
		require.Nil(t, val)

		require.Equal(t, 0, called)
		require.False(t, cfg.Metrics.DiskUsage)
	})

	t.Run("disk usage: metric enabled", func(t *testing.T) {
		called := 0
		cfg := createConfig()
		patches := gomonkey.NewPatches()
		ctx, cancel := context.WithCancel(context.Background())
		v := NewLinuxDiskUsageCollector(ctx, ctx, cfg, log)
		defer cancel()

		patches.ApplyFunc(disk.GetDevices, func() ([]string, error) {
			return []string{"stub"}, nil
		})
		patches.ApplyMethod(
			&LinuxDiskUsageCollector{}, "Partitions", func() ([]PartitionStat, error) {
				return []PartitionStat{{
					Device:     "stub",
					Mountpoint: "stub",
					FsType:     "ext4",
					Opts:       nil,
				}}, nil
			})
		patches.ApplyMethodFunc(
			&LinuxDiskUsageCollector{}, "GetDiskUsageStat", func(_, _ string) (*disk.UsageStat, error) {
				called++
				return &disk.UsageStat{}, nil
			})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, err)
		require.NotNil(t, ch)

		val := <-ch
		require.NotNil(t, val)
		cancel()

		require.Equal(t, 1, called)
		require.True(t, cfg.Metrics.DiskUsage)
	})
}
