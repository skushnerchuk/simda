package tests

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint: revive
	. "github.com/onsi/gomega"    //nolint: revive
	"github.com/skushnerchuk/simda/internal/config"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	warm    uint32 = 1
	receive uint32 = 1
)

var (
	cfg          config.DaemonConfig
	clientCtx    context.Context
	clientCancel context.CancelFunc
	request      pb.Request
	client       pb.SimdaClient
	streamer     pb.Simda_StreamSnapshotsClient
)

var _ = BeforeSuite(func() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "../configs/simda_linux.yml"
	}
	err := config.Load(configPath, &cfg)
	Expect(err).ShouldNot(HaveOccurred())

	clientCtx, clientCancel = context.WithCancel(context.Background())
	grpcAddr := cfg.Host + ":" + cfg.Port
	credentials := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(grpcAddr, credentials)
	Expect(err).ToNot(HaveOccurred())
	Expect(conn).ToNot(BeNil())
	client = pb.NewSimdaClient(conn)
	request = pb.Request{Period: receive, Warming: warm}
	streamer, err = client.StreamSnapshots(clientCtx, &request)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(streamer).ToNot(BeNil())
})

var _ = AfterSuite(func() {
	clientCancel()
})

func restoreDaemonConfig() {
	viper.Set("metrics.cpu_avg", true)
	viper.Set("metrics.disk_io", true)
	viper.Set("metrics.disk_usage", true)
	viper.Set("metrics.load_avg", true)
	viper.Set("metrics.net_connections", true)
	viper.Set("metrics.net_connections_states", true)
	viper.Set("metrics.net_top_by_connection", true)
	viper.Set("metrics.net_top_by_protocol", true)
	_ = viper.WriteConfig()
}

var _ = Describe("Common", func() {
	AfterEach(func() {
		restoreDaemonConfig()
	})

	BeforeEach(func() {
		restoreDaemonConfig()
	})
	It("check warming & receiving time", func() {
		startWarming := time.Now()

		snapshot, err := streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		warmingTime := time.Since(startWarming)

		startReceiving := time.Now()
		snapshot, err = streamer.Recv()
		Expect(snapshot).ToNot(BeNil())
		receivingTime := time.Since(startReceiving)

		Expect(warmingTime.Seconds()).Should(
			BeNumerically(">=", time.Duration(warm)-500*time.Millisecond),
		)
		Expect(warmingTime.Seconds()).Should(
			BeNumerically("<", time.Duration(warm)+500*time.Millisecond),
		)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(receivingTime.Seconds()).Should(
			BeNumerically(">=", time.Duration(receive)-500*time.Millisecond),
		)
		Expect(receivingTime.Seconds()).Should(
			BeNumerically("<", time.Duration(receive)+500*time.Millisecond),
		)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("check all metrics", func() {
		snapshot, err := streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())

		Expect(snapshot.Metrics.LoadAvg).Should(BeTrue())
		Expect(snapshot.Metrics.CpuAvg).Should(BeTrue())
		Expect(snapshot.Metrics.DiskIO).Should(BeTrue())
		Expect(snapshot.Metrics.DiskUsage).Should(BeTrue())
		Expect(snapshot.Metrics.NetConnections).Should(BeTrue())
		Expect(snapshot.Metrics.NetConnectionStates).Should(BeTrue())
		Expect(snapshot.Metrics.NetTopByProtocol).Should(BeTrue())
		Expect(snapshot.Metrics.NetTopByConnection).Should(BeTrue())

		Expect(snapshot.LoadAvg).ToNot(BeNil())
		Expect(snapshot.CpuAvg).ToNot(BeNil())
		Expect(snapshot.DiskIO).ToNot(BeNil())
		Expect(snapshot.DiskUsage).ToNot(BeNil())
		Expect(snapshot.NetConnections).ToNot(BeNil())
		Expect(snapshot.NetConnectionsStates).ToNot(BeNil())
		Expect(snapshot.NetTopByProtocol).ToNot(BeNil())
		Expect(snapshot.NetTopByConnection).ToNot(BeNil())
	})
})

var _ = Describe("load avg", func() {
	var (
		err      error
		snapshot *pb.Snapshot
	)

	AfterEach(func() {
		restoreDaemonConfig()
	})

	BeforeEach(func() {
		restoreDaemonConfig()
	})

	It("check runtime values", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.LoadAvg).ToNot(BeNil())
		Expect(snapshot.LoadAvg.One).Should(BeNumerically(">", 0))
		Expect(snapshot.LoadAvg.Five).Should(BeNumerically(">", 0))
		Expect(snapshot.LoadAvg.Fifteen).Should(BeNumerically(">", 0))
	})

	It("check runtime on/off", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.LoadAvg).ToNot(BeNil())

		viper.Set("metrics.load_avg", false)
		_ = viper.WriteConfig()

		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.LoadAvg).To(BeNil())
	})
})

var _ = Describe("cpu", func() {
	var (
		err      error
		snapshot *pb.Snapshot
	)

	AfterEach(func() {
		restoreDaemonConfig()
	})

	BeforeEach(func() {
		restoreDaemonConfig()
	})

	It("check runtime values", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())

		Expect(snapshot.CpuAvg).ToNot(BeNil())
		Expect(snapshot.CpuAvg.System).Should(BeNumerically(">", 0))
		Expect(snapshot.CpuAvg.User).Should(BeNumerically(">", 0))
		Expect(snapshot.CpuAvg.Idle).Should(BeNumerically(">", 0))
	})

	It("check runtime on/off", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.CpuAvg).ToNot(BeNil())

		viper.Set("metrics.cpu_avg", false)
		_ = viper.WriteConfig()

		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.CpuAvg).To(BeNil())
	})
})

var _ = Describe("disk i/o", func() {
	var (
		err      error
		snapshot *pb.Snapshot
	)

	AfterEach(func() {
		restoreDaemonConfig()
	})

	BeforeEach(func() {
		restoreDaemonConfig()
	})

	It("check runtime values", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.DiskIO).ToNot(BeNil())
	})

	It("check runtime on/off", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.DiskIO).ToNot(BeNil())

		viper.Set("metrics.disk_io", false)
		_ = viper.WriteConfig()

		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.DiskIO).To(BeNil())
	})
})

var _ = Describe("disk usage", func() {
	var (
		err      error
		snapshot *pb.Snapshot
	)

	AfterEach(func() {
		restoreDaemonConfig()
	})

	BeforeEach(func() {
		restoreDaemonConfig()
	})

	It("check runtime values", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.DiskUsage).ToNot(BeNil())
	})

	It("check runtime on/off", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.DiskIO).ToNot(BeNil())

		viper.Set("metrics.disk_usage", false)
		_ = viper.WriteConfig()

		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.DiskUsage).To(BeNil())
	})
})

var _ = Describe("net connections", func() {
	var (
		err      error
		snapshot *pb.Snapshot
	)

	AfterEach(func() {
		restoreDaemonConfig()
	})

	BeforeEach(func() {
		restoreDaemonConfig()
	})

	It("check runtime values", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetConnections).ToNot(BeNil())
	})

	It("check runtime on/off", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetConnections).ToNot(BeNil())

		viper.Set("metrics.net_connections", false)
		_ = viper.WriteConfig()

		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetConnections).To(BeNil())
	})
})

var _ = Describe("net connections states", func() {
	var (
		err      error
		snapshot *pb.Snapshot
	)

	AfterEach(func() {
		restoreDaemonConfig()
	})

	BeforeEach(func() {
		restoreDaemonConfig()
	})

	It("check runtime values", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetConnectionsStates).ToNot(BeNil())
	})

	It("check runtime on/off", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetConnectionsStates).ToNot(BeNil())

		viper.Set("metrics.net_connections_states", false)
		_ = viper.WriteConfig()

		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetConnectionsStates).To(BeNil())
	})
})

var _ = Describe("net top by connection", func() {
	var (
		err      error
		snapshot *pb.Snapshot
	)

	AfterEach(func() {
		restoreDaemonConfig()
	})

	BeforeEach(func() {
		restoreDaemonConfig()
	})

	It("check runtime values", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetTopByConnection).ToNot(BeNil())
	})

	It("check runtime on/off", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetTopByConnection).ToNot(BeNil())

		viper.Set("metrics.net_top_by_connection", false)
		_ = viper.WriteConfig()

		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetTopByConnection).To(BeNil())
	})
})

var _ = Describe("net top by protocol", func() {
	var (
		err      error
		snapshot *pb.Snapshot
	)

	AfterEach(func() {
		restoreDaemonConfig()
	})

	BeforeEach(func() {
		restoreDaemonConfig()
	})

	It("check runtime values", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetTopByProtocol).ToNot(BeNil())
	})

	It("check runtime on/off", func() {
		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetTopByProtocol).ToNot(BeNil())

		viper.Set("metrics.net_top_by_protocol", false)
		_ = viper.WriteConfig()

		snapshot, err = streamer.Recv()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.NetTopByProtocol).To(BeNil())
	})
})
