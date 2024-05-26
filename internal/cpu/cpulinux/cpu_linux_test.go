//go:build linux

package cpulinux

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/cpu"
	"github.com/skushnerchuk/simda/internal/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

var log = logger.NewSLogger(os.Stdout, "DEBUG")

func createConfig() *config.DaemonConfig {
	return &config.DaemonConfig{
		System:  config.SystemPoints{Proc: "/proc"},
		Metrics: config.Metrics{CPUAvg: false},
	}
}

func TestCPUStat(t *testing.T) {
	log.Disable()

	t.Run("cpu: check stat parser", func(t *testing.T) {
		cfg := createConfig()
		v := NewLinuxCPUCollector(context.TODO(), context.TODO(), cfg, log)

		err := v.parseStatLine("cpu  1826207 68727 673820 42671281 86015 158628 47813 0 0 0")
		require.Nil(t, err)

		err = v.parseStatLine("cpu  hello 68727 673820 42671281 86015 158628 47813 0 0 0")
		require.Error(t, err)

		err = v.parseStatLine("cpu  1826207 68727 673820 42671281 86015 158628")
		require.Error(t, err)
	})

	t.Run("cpu: Get() error", func(t *testing.T) {
		cfg := createConfig()
		v := NewLinuxCPUCollector(context.TODO(), context.TODO(), cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(&LinuxCPUCollector{}, "parseStatLine", func(_ string) error {
			return fmt.Errorf("error")
		})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.Get()
		require.Error(t, err)
		require.Nil(t, val)
	})

	t.Run("cpu: Get() ok", func(t *testing.T) {
		cfg := createConfig()
		v := NewLinuxCPUCollector(context.TODO(), context.TODO(), cfg, log)

		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(&LinuxCPUCollector{}, "parseStatLine", func(_ string) error {
			v.UserInPercent = 1.0
			v.SystemInPercent = 2.0
			v.IdleInPercent = 3.0
			return nil
		})
		t.Cleanup(func() { patches.Reset() })

		val, err := v.Get()
		require.Nil(t, err)
		require.Equal(t, val.User, 1.0)
		require.Equal(t, val.System, 2.0)
		require.Equal(t, val.Idle, 3.0)
	})
}

func TestCPUWithMocks(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("cpu: Get() error", func(t *testing.T) {
		cfg := createConfig()
		cfg.Metrics.CPUAvg = true
		v := NewLinuxCPUCollector(context.TODO(), context.TODO(), cfg, log)
		patches := gomonkey.NewPatches()
		patches.ApplyMethod(&LinuxCPUCollector{}, "Get", func() (*cpu.Data, error) {
			return nil, fmt.Errorf("error")
		})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, ch)
		require.Error(t, err)
		require.False(t, cfg.Metrics.CPUAvg)
	})

	t.Run("cpu: metric disabled", func(t *testing.T) {
		getCalled := 0

		cfg := createConfig()
		cfg.Metrics.CPUAvg = false
		patches := gomonkey.NewPatches()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		v := NewLinuxCPUCollector(ctx, ctx, cfg, log)
		defer cancel()

		patches.ApplyMethod(&LinuxCPUCollector{}, "Get", func() (*cpu.Data, error) {
			getCalled++
			return &cpu.Data{}, nil
		})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, err)
		require.NotNil(t, ch)

		val := <-ch
		require.Nil(t, val)

		require.Equal(t, 1, getCalled)
		require.False(t, cfg.Metrics.CPUAvg)
	})

	t.Run("cpu: metric enabled", func(t *testing.T) {
		getCalled := 0

		cfg := createConfig()
		cfg.Metrics.CPUAvg = true
		patches := gomonkey.NewPatches()
		ctx, cancel := context.WithCancel(context.Background())
		v := NewLinuxCPUCollector(ctx, ctx, cfg, log)
		defer cancel()

		patches.ApplyMethod(&LinuxCPUCollector{}, "Get", func() (*cpu.Data, error) {
			getCalled++
			return &cpu.Data{}, nil
		})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, err)
		require.NotNil(t, ch)

		val := <-ch
		require.NotNil(t, val)
		cancel()

		require.Equal(t, 2, getCalled)
		require.True(t, cfg.Metrics.CPUAvg)
	})

	t.Run("cpu: delayed error", func(t *testing.T) {
		getCalled := 0

		cfg := createConfig()
		cfg.Metrics.CPUAvg = true
		patches := gomonkey.NewPatches()
		ctx, cancel := context.WithCancel(context.Background())
		v := NewLinuxCPUCollector(ctx, ctx, cfg, log)
		defer cancel()

		patches.ApplyMethod(&LinuxCPUCollector{}, "Get", func() (*cpu.Data, error) {
			getCalled++
			if getCalled > 1 {
				return nil, fmt.Errorf("error")
			}
			return &cpu.Data{}, nil
		})
		t.Cleanup(func() { patches.Reset() })

		ch, err := v.Run()
		require.Nil(t, err)
		require.NotNil(t, ch)

		val := <-ch
		require.Nil(t, val)
		cancel()

		require.Equal(t, 2, getCalled)
		require.False(t, cfg.Metrics.CPUAvg)
	})
}
