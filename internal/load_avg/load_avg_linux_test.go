//go:build linux

package loadavg

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/logger"
	"github.com/stretchr/testify/require"
)

var log = logger.NewSLogger(os.Stdout, "DEBUG")

func createConfig() *config.DaemonConfig {
	return &config.DaemonConfig{
		System:  config.SystemPoints{Proc: "/proc"},
		Metrics: config.Metrics{CPUAvg: true},
	}
}

func TestLoadAvgStat(t *testing.T) {
	log.Disable()

	t.Run("load avg: Get() ok", func(t *testing.T) {
		cfg := createConfig()
		v := NewLinuxLoadAverageCollector(context.TODO(), context.TODO(), cfg, log)
		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(
			&LinuxLoadAverageCollector{}, "readLoadAvgFromFile", func() ([]string, error) {
				return []string{"2.12", "2.30", "2.37", "3/2420", "137390"}, nil
			})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.Get()
		require.NoError(t, err)
		require.Equal(t, 2.12, val.Load1)
		require.Equal(t, 2.30, val.Load5)
		require.Equal(t, 2.37, val.Load15)
	})

	t.Run("load avg: Get() error", func(t *testing.T) {
		cfg := createConfig()
		v := NewLinuxLoadAverageCollector(context.TODO(), context.TODO(), cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(
			&LinuxLoadAverageCollector{}, "readLoadAvgFromFile", func() ([]string, error) {
				return nil, fmt.Errorf("error")
			})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.Get()
		require.Error(t, err)
		require.Nil(t, val)
	})

	t.Run("load avg: Run() error", func(t *testing.T) {
		cfg := createConfig()
		v := NewLinuxLoadAverageCollector(context.TODO(), context.TODO(), cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyMethod(&LinuxLoadAverageCollector{}, "Get", func() (*AvgStat, error) {
			return nil, fmt.Errorf("error")
		})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, ch)
		require.Error(t, err)
		require.False(t, cfg.Metrics.LoadAvg)
	})

	t.Run("load avg: metric disabled", func(t *testing.T) {
		getCalled := 0

		cfg := createConfig()
		cfg.Metrics.LoadAvg = false
		patches := gomonkey.NewPatches()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		v := NewLinuxLoadAverageCollector(ctx, ctx, cfg, log)
		defer cancel()

		patches.ApplyMethod(&LinuxLoadAverageCollector{}, "Get", func() (*AvgStat, error) {
			getCalled++
			return &AvgStat{}, nil
		})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, err)
		require.NotNil(t, ch)

		val := <-ch
		require.Nil(t, val)

		require.Equal(t, 1, getCalled)
		require.False(t, cfg.Metrics.LoadAvg)
	})

	t.Run("load avg: metric enabled", func(t *testing.T) {
		getCalled := 0

		cfg := createConfig()
		cfg.Metrics.LoadAvg = true
		patches := gomonkey.NewPatches()
		ctx, cancel := context.WithCancel(context.Background())
		v := NewLinuxLoadAverageCollector(ctx, ctx, cfg, log)
		defer cancel()

		patches.ApplyMethod(&LinuxLoadAverageCollector{}, "Get", func() (*AvgStat, error) {
			getCalled++
			return &AvgStat{}, nil
		})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, err)
		require.NotNil(t, ch)

		val := <-ch
		require.NotNil(t, val)
		cancel()

		require.Equal(t, 2, getCalled)
		require.True(t, cfg.Metrics.LoadAvg)
	})

	t.Run("load avg: delayed error", func(t *testing.T) {
		getCalled := 0

		cfg := createConfig()
		cfg.Metrics.LoadAvg = true
		patches := gomonkey.NewPatches()
		ctx, cancel := context.WithCancel(context.Background())
		v := NewLinuxLoadAverageCollector(ctx, ctx, cfg, log)
		defer cancel()

		patches.ApplyMethod(&LinuxLoadAverageCollector{}, "Get", func() (*AvgStat, error) {
			getCalled++
			if getCalled > 1 {
				return nil, fmt.Errorf("error")
			}
			return &AvgStat{}, nil
		})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, err)
		require.NotNil(t, ch)

		val := <-ch
		require.Nil(t, val)
		cancel()

		require.Equal(t, 2, getCalled)
		require.False(t, cfg.Metrics.LoadAvg)
	})
}
