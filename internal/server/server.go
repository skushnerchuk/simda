package server

import (
	"context"
	"net"

	"github.com/bufbuild/protovalidate-go"
	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/logger"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
	"google.golang.org/grpc"
)

type SimdaServer struct {
	address   string
	server    *grpc.Server
	logger    logger.Logger
	serverCtx context.Context
	pb.UnimplementedSimdaServer
	cfg       *config.DaemonConfig
	validator *protovalidate.Validator
}

func NewSimdaServer(c *config.DaemonConfig, l logger.Logger) SimdaServer {
	v, _ := protovalidate.New()
	return SimdaServer{
		address:   c.Host + ":" + c.Port,
		logger:    l,
		cfg:       c,
		validator: v,
	}
}

func (s *SimdaServer) Start(ctx context.Context) error {
	s.serverCtx = ctx
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer(grpc.ChainUnaryInterceptor(s.recovery))
	pb.RegisterSimdaServer(s.server, s)

	go func() {
		err = s.server.Serve(listener)
		if err != nil {
			s.logger.Error(err.Error())
			return
		}
	}()
	s.logger.Info("server started", "grpc", s.address)
	return nil
}

func (s *SimdaServer) Stop() {
	s.server.GracefulStop()
}
