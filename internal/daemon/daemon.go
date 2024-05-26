package daemon

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/logger"
	"github.com/skushnerchuk/simda/internal/server"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
)

type Daemon struct {
	pb.UnimplementedSimdaServer
	cfg    *config.DaemonConfig
	logger logger.Logger
	server server.SimdaServer
}

func NewDaemon(cfg *config.DaemonConfig) (*Daemon, error) {
	l := logger.NewSLogger(os.Stdout, cfg.LogLevel)

	return &Daemon{
		logger: l,
		cfg:    cfg,
		server: server.NewSimdaServer(cfg, l),
	}, nil
}

func (d *Daemon) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	var err error
	defer cancel()

	go func() {
		err = d.server.Start(ctx)
		if err != nil {
			d.logger.Error("failed to start server", "error", err.Error())
			cancel()
			return
		}
	}()

	<-ctx.Done()
	if err == nil {
		d.server.Stop()
	}
	return err
}
