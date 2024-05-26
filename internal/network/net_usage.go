package network

import (
	"fmt"
	"time"
)

type PacketInfo struct {
	SourceIP        string
	DestinationIP   string
	Protocol        string
	SourcePort      string
	DestinationPort string
	PayloadSize     uint64
	Timestamp       time.Time
}

func (p *PacketInfo) ConnectionID() string {
	return fmt.Sprintf("%s %s:%s-%s:%s", p.Protocol, p.SourceIP, p.SourcePort, p.DestinationIP, p.DestinationPort)
}

type NetUsageByProtocol struct {
	Protocol string
	Bytes    uint64
	Percent  float64
}

type NetworkPacketStat []PacketInfo //nolint:revive

type NetworkPacketCollector interface { //nolint:revive
	Run() (<-chan NetworkPacketStat, error)
}
