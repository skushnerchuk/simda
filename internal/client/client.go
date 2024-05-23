package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/skushnerchuk/simda/internal/logger"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SimdaClient struct {
	id            uuid.UUID
	logger        logger.Logger
	warm          uint
	receivePeriod uint
	host, port    string
	err           error
	ch            chan *pb.Snapshot
}

func NewClient(warm, receive uint, host, port string) *SimdaClient {
	return &SimdaClient{
		warm:          warm,
		receivePeriod: receive,
		host:          host,
		port:          port,
		id:            uuid.New(),
		logger:        logger.NewSLogger(os.Stdout, "DEBUG"),
		err:           nil,
	}
}

func (d *SimdaClient) Run(ctx context.Context, ch chan *pb.Snapshot, stop context.CancelFunc) error {
	d.ch = ch
	stream, err := d.ListenStream(ctx)
	if err != nil {
		d.err = err
		return err
	}
	d.receive(ctx, stream, stop)
	<-ctx.Done()
	return d.err
}

func (d *SimdaClient) ListenStream(ctx context.Context) (pb.Simda_StreamSnapshotsClient, error) {
	credentials := grpc.WithTransportCredentials(insecure.NewCredentials())
	addr := fmt.Sprintf("%s:%s", d.host, d.port)
	conn, err := grpc.DialContext(ctx, addr, credentials)
	if err != nil {
		return nil, fmt.Errorf("failed connect to server %s: %w", addr, err)
	}
	client := pb.NewSimdaClient(conn)

	go func() {
		defer conn.Close()
		<-ctx.Done()
	}()

	in := &pb.Request{
		Period:  uint32(d.receivePeriod),
		Warming: uint32(d.warm),
	}
	return client.StreamSnapshots(ctx, in)
}

func (d *SimdaClient) receive(ctx context.Context, stream pb.Simda_StreamSnapshotsClient, stop context.CancelFunc) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				snapshot, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					d.err = fmt.Errorf("server terminated")
					stop()
					return
				}
				if err != nil {
					d.err = err
					stop()
					return
				}
				d.ch <- snapshot
			}
		}
	}()
}
