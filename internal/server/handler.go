package server

import (
	"time"

	pb "github.com/skushnerchuk/simda/internal/server/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func (s *SimdaServer) StreamSnapshots(r *pb.Request, srv pb.Simda_StreamSnapshotsServer) error {
	s.logger.Debug(
		"Client connected",
		"warming_uptime", r.Warming,
		"period", r.Period,
	)

	err := s.validator.Validate(r)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = s.serveClient(r, srv)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			s.logger.Error("Error serving client", "error", e.Message(), "code", e.Code())
		} else {
			s.logger.Error("Error serving client", "error", err.Error())
		}
		return err
	}
	return nil
}

func (s *SimdaServer) serveClient(r *pb.Request, srv pb.Simda_StreamSnapshotsServer) error {
	p, _ := peer.FromContext(srv.Context())

	la := s.streamSnapshot(r, srv)

	for {
		select {
		case <-s.serverCtx.Done():
			s.logger.Debug("Server stopped, shutdown worker", "client", p.Addr.String())
			return nil
		case <-srv.Context().Done():
			s.logger.Debug("Client disconnected, shutdown worker", "client", p.Addr.String())
			return nil
		case snapshot, ok := <-la:
			if ok {
				err := srv.Send(snapshot)
				if err != nil {
					return err
				}
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (s *SimdaServer) streamSnapshot(r *pb.Request, srv pb.Simda_StreamSnapshotsServer) <-chan *pb.Snapshot {
	streamer := NewSnapshotStreamer(s.serverCtx, srv.Context(), r, s.logger, s.cfg)
	return streamer.Stream()
}
