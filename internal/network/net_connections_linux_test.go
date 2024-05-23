//go:build linux

package network

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

var (
	cfg = config.DaemonConfig{
		System: config.SystemPoints{
			Proc: "/proc",
			TCP:  "/proc/net/tcp",
			TCP6: "/proc/net/tcp6",
			UDP:  "/proc/net/udp",
			UDP6: "/proc/net/udp6",
		},
		Metrics: config.Metrics{NetConnections: true, NetConnectionsStates: true},
	}
	log = logger.NewSLogger(os.Stdout, "DEBUG")
)

func TestNetworkConnectionsStat(t *testing.T) {
	log.Disable()

	t.Run("address parser (ok)", func(t *testing.T) {
		addr, err := parseAddr("7C01A8C0:9F52")
		require.NoError(t, err)
		require.NotNil(t, addr)
		require.Equal(t, "192.168.1.124", addr.IP.String())
		require.Equal(t, uint16(40786), addr.Port)
	})

	t.Run("address parser (fail)", func(t *testing.T) {
		addr, err := parseAddr("C01A8C0:9F5")
		require.Error(t, err)
		require.Nil(t, addr)
	})

	t.Run("tcp connections", func(t *testing.T) {
		ctx := context.Background()

		l := NewLinuxConnectionsCollector(ctx, ctx, &cfg, log)

		stat, err := l.TCPSocks()
		require.NoError(t, err)
		require.Greater(t, len(stat), 0)

		stat, err = l.TCP6Socks()
		require.NoError(t, err)
		require.Greater(t, len(stat), 0)
	})

	t.Run("udp connections", func(t *testing.T) {
		ctx := context.Background()

		l := NewLinuxConnectionsCollector(ctx, ctx, &cfg, log)

		stat, err := l.UDPSocks()
		require.NoError(t, err)
		require.Greater(t, len(stat), 0)

		stat, err = l.UDP6Socks()
		require.NoError(t, err)
		require.Greater(t, len(stat), 0)
	})

	t.Run("all connections", func(t *testing.T) {
		ctx := context.Background()

		l := NewLinuxConnectionsCollector(ctx, ctx, &cfg, log)

		stat, err := l.GetConnection()
		require.NoError(t, err)
		require.Greater(t, len(stat), 0)
	})

	t.Run("check network connections collector runner", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		cfg.Metrics.NetConnectionsStates = true
		cfg.Metrics.NetConnections = true

		ctx, cancel := context.WithCancel(context.Background())
		v := NewLinuxConnectionsCollector(ctx, ctx, &cfg, log)
		ch, err := v.Run()

		require.Nil(t, err)
		value := <-ch
		cancel()
		require.Greater(t, len(value), 0)
	})

	t.Run("check network connections collector runner (disabled)", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		cfg.Metrics.NetConnectionsStates = false
		cfg.Metrics.NetConnections = false

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		v := NewLinuxConnectionsCollector(ctx, ctx, &cfg, log)
		ch, err := v.Run()

		require.Nil(t, err)
		value := <-ch
		cancel()
		require.Nil(t, value)
	})
}
