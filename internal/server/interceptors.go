package server

import (
	"context"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SimdaServer) recovery(
	ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = status.Error(codes.Internal, "critical error on server")
			s.logger.Error(
				err.Error(),
				"stack", debug.Stack(),
			)
		}
	}()
	resp, err = handler(ctx, req)
	return resp, err
}
