package network

import (
	"encoding/json"
	"fmt"
	"net"
)

const (
	ProtocolTCP  = "tcp"
	ProtocolTCP6 = "tcp6"
	ProtocolUDP  = "udp"
	ProtocolUDP6 = "udp6"
)

type SockAddr struct {
	IP   net.IP `json:"ip"`
	Port uint16 `json:"port"`
}

func (s *SockAddr) String() string {
	return fmt.Sprintf("%v:%d", s.IP, s.Port)
}

type Process struct {
	Pid     int
	CmdLine string
}

type Connection struct {
	SocketID       string
	Protocol       string
	Process        *Process
	User           string
	LocalAddress   *SockAddr
	ForeignAddress *SockAddr
	State          string
	UserID         uint32
}

func (c *Connection) String() string {
	v, _ := json.Marshal(c)
	return string(v)
}

type ConnectionsStat []Connection

func (c *ConnectionsStat) String() string {
	v, _ := json.Marshal(c)
	return string(v)
}

type ConnectionsCollector interface {
	Run() (<-chan ConnectionsStat, error)
}
