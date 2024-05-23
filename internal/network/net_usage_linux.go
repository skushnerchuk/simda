//go:build linux

package network

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/logger"
	pb "github.com/skushnerchuk/simda/internal/server/gen"
)

type LinuxNetworkPackagesCollector struct {
	serverCtx context.Context
	clientCtx context.Context
	cfg       *config.DaemonConfig
	l         logger.Logger
	r         *pb.Request
}

func NewLinuxNetworkPackagesCollector(
	serverCtx, clientCtx context.Context, cfg *config.DaemonConfig, l logger.Logger, r *pb.Request,
) *LinuxNetworkPackagesCollector {
	return &LinuxNetworkPackagesCollector{
		serverCtx: serverCtx,
		clientCtx: clientCtx,
		cfg:       cfg,
		l:         l,
		r:         r,
	}
}

func (l *LinuxNetworkPackagesCollector) Run() (<-chan NetworkPacketStat, error) {
	handle, err := pcap.OpenLive(l.cfg.System.Interface, 65535, false, 100*time.Millisecond)
	if err != nil {
		l.cfg.Metrics.NetTopByClients = false
		l.cfg.Metrics.NetTopByProtocol = false
		return nil, err
	}
	ch := make(chan NetworkPacketStat)
	sendTicker := time.NewTicker(time.Second)

	go func() {
		defer close(ch)
		defer handle.Close()

		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		stat := make(NetworkPacketStat, 0)
		for {
			select {
			case <-l.serverCtx.Done():
			case <-l.clientCtx.Done():
				l.l.Debug("network packages collector stopped")
				return
			case <-sendTicker.C:
				ch <- stat
				stat = stat[:0]
			default:
				if !l.cfg.Metrics.NetTopByClients && !l.cfg.Metrics.NetTopByProtocol {
					continue
				}
				np, err := packetSource.NextPacket()
				if err == nil && np != nil {
					p := getPacket(np)
					if p != nil {
						stat = append(stat, *p)
					}
				} else {
					if errors.Is(err, pcap.NextErrorTimeoutExpired) {
						packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
					} else {
						l.l.Error("connections package collector error", "error", err.Error())
						l.cfg.Metrics.NetTopByClients = false
						l.cfg.Metrics.NetTopByProtocol = false
						return
					}
				}
			}
		}
	}()
	return ch, nil
}

func getPacket(packet gopacket.Packet) *PacketInfo {
	p := NetworkLayer(packet)
	if p == nil {
		p = TransportLayer(packet)
	}
	return p
}

func NetworkLayer(packet gopacket.Packet) *PacketInfo { //nolint:revive
	var info PacketInfo
	arpLayer := packet.Layer(layers.LayerTypeARP)
	if arpLayer != nil {
		arp := arpLayer.(*layers.ARP)
		info.SourceIP = net.IP(arp.SourceProtAddress).String()
		info.DestinationIP = net.IP(arp.DstProtAddress).String()
		info.Protocol = "ARP"
		md := packet.Metadata()
		if md != nil {
			info.Timestamp = packet.Metadata().Timestamp
		}
		info.PayloadSize = uint64(len(arp.Payload)) + uint64(len(arp.Contents))
		return &info
	}
	return nil
}

func getIP(packet gopacket.Packet) *layers.IPv4 {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		return ipLayer.(*layers.IPv4)
	}
	return nil
}

func TransportLayer(packet gopacket.Packet) *PacketInfo {
	var info PacketInfo

	if packet.Metadata() != nil {
		info.PayloadSize = uint64(packet.Metadata().CaptureLength)
		info.Timestamp = packet.Metadata().Timestamp
	}

	// IPv6 не поддерживаем для упрощения
	ip := getIP(packet)
	if ip == nil {
		return nil
	}

	info.Protocol = ip.Protocol.String()
	info.SourceIP = ip.SrcIP.String()
	info.DestinationIP = ip.DstIP.String()

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp := tcpLayer.(*layers.TCP)
		info.SourcePort = tcp.SrcPort.String()
		info.DestinationPort = tcp.DstPort.String()
	}

	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		udp := udpLayer.(*layers.UDP)
		info.SourcePort = udp.SrcPort.String()
		info.DestinationPort = udp.DstPort.String()
	}

	return &info
}

func (l *LinuxNetworkPackagesCollector) CalcProtocolStat(stat []NetworkPacketStat) []*pb.NetTopByProtocol {
	protocols := make(map[string]uint64)
	totalBytes := uint64(0)
	for _, elem := range stat {
		for _, p := range elem {
			totalBytes += p.PayloadSize
			protocols[p.Protocol] += p.PayloadSize
		}
	}

	result := make([]*pb.NetTopByProtocol, 0)

	for protocol, bytes := range protocols {
		result = append(result, &pb.NetTopByProtocol{
			Protocol: protocol,
			Bytes:    bytes,
			Percent:  (float64(bytes) / float64(totalBytes)) * 100.0,
		})
	}
	return result
}

func (l *LinuxNetworkPackagesCollector) CalcProtocolConnectionStat(
	stat []NetworkPacketStat,
) []*pb.NetTopByConnection {
	connections := make(map[string][]PacketInfo)
	for _, elem := range stat {
		for _, p := range elem {
			if _, ok := connections[p.ConnectionID()]; ok {
				connections[p.ConnectionID()] = append(connections[p.ConnectionID()], p)
			} else {
				connections[p.ConnectionID()] = make([]PacketInfo, 0)
				connections[p.ConnectionID()] = append(connections[p.ConnectionID()], p)
			}
		}
	}

	result := make([]*pb.NetTopByConnection, 0)

	for _, packets := range connections {
		sourceAddr := &pb.SockAddr{Ip: packets[0].SourceIP}
		port, _ := strconv.Atoi(packets[0].SourcePort)
		sourceAddr.Port = uint32(port)

		destinationAddr := &pb.SockAddr{Ip: packets[0].DestinationIP}
		port, _ = strconv.Atoi(packets[0].DestinationPort)
		destinationAddr.Port = uint32(port)

		bytes := uint64(0)
		for _, packet := range packets {
			bytes += packet.PayloadSize
		}
		percent := 0.0
		if bytes > 0 {
			percent = (float64(l.r.Warming) / float64(bytes)) * 100.0
		}

		result = append(result, &pb.NetTopByConnection{
			Protocol:        packets[0].Protocol,
			Bytes:           bytes,
			Percent:         percent,
			SourceAddr:      sourceAddr,
			DestinationAddr: destinationAddr,
		})
	}
	return result
}
