//go:build linux

package network

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/logger"
	"github.com/skushnerchuk/simda/internal/utils"
)

const (
	ipv4StrLen = 8
	ipv6StrLen = 32
	sockPrefix = "socket:["
)

type ConnectionsState uint8

const (
	Established ConnectionsState = 0x01 //nolint:staticcheck
	SynSent                      = 0x02
	SynRecv                      = 0x03
	FinWait1                     = 0x04
	FinWait2                     = 0x05
	TimeWait                     = 0x06
	Close                        = 0x07
	CloseWait                    = 0x08
	LastAck                      = 0x09
	Listen                       = 0x0a
	Closing                      = 0x0b
	NewSynRecv                   = 0x0c
)

func (c ConnectionsState) String() string {
	switch c {
	case Established:
		return "ESTABLISHED"
	case SynSent:
		return "SYN_SENT"
	case SynRecv:
		return "SYN_RECV"
	case FinWait1:
		return "FIN_WAIT1"
	case FinWait2:
		return "FIN_WAIT2"
	case TimeWait:
		return "TIME_WAIT"
	case Close:
		return "CLOSE"
	case CloseWait:
		return "CLOSE_WAIT"
	case LastAck:
		return "LAST_ACK"
	case Listen:
		return "LISTEN"
	case Closing:
		return "CLOSING"
	case NewSynRecv:
		return "NEW_SYN_RECV"
	default:
		return "UNKNOWN"
	}
}

func parseAddr(s string) (*SockAddr, error) {
	fields := strings.Split(s, ":")
	if len(fields) < 2 {
		return nil, fmt.Errorf("netstat: not enough fields: %v", s)
	}
	var ip net.IP
	var err error
	switch len(fields[0]) {
	case ipv4StrLen:
		ip, err = utils.ParseIPv4(fields[0])
	case ipv6StrLen:
		ip, err = utils.ParseIPv6(fields[0])
	default:
		err = fmt.Errorf("netstat: bad formatted string: %v", fields[0])
	}
	if err != nil {
		return nil, err
	}
	v, err := strconv.ParseUint(fields[1], 16, 16)
	if err != nil {
		return nil, err
	}
	return &SockAddr{IP: ip, Port: uint16(v)}, nil
}

func parseConnectionsFile(r io.Reader, protocol string) ([]Connection, error) {
	br := bufio.NewScanner(r)
	tab := make([]Connection, 0, 4)

	// Пропускаем заголовок
	br.Scan()

	// Бежим по всем открытым соединениям
	for br.Scan() {
		var entry Connection
		entry.Protocol = protocol
		line := br.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		if len(fields) < 12 {
			return nil, fmt.Errorf("netstat: not enough fields: %v, %v", len(fields), fields)
		}
		addr, err := parseAddr(fields[1])
		if err != nil {
			return nil, err
		}
		entry.LocalAddress = addr
		addr, err = parseAddr(fields[2])
		if err != nil {
			return nil, err
		}
		entry.ForeignAddress = addr
		u, err := strconv.ParseUint(fields[3], 16, 8)
		if err != nil {
			return nil, err
		}
		entry.State = ConnectionsState(u).String()
		userID := fields[7]
		u, err = strconv.ParseUint(userID, 10, 32)
		if err != nil {
			return nil, err
		}
		entry.UserID = uint32(u)
		entry.User = utils.GetUsernameByID(userID)
		entry.SocketID = fields[9]
		tab = append(tab, entry)
	}
	return tab, br.Err()
}

type runningProcess struct {
	rootDir     string
	pid         int
	connections []Connection
	p           *Process
}

func (p *runningProcess) getProcessDetail() {
	cmdLink, _ := os.Readlink(path.Join(p.rootDir, "/exe"))

	fdDir := path.Join(p.rootDir, "/fd")
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return
	}
	// Бежим по файловым дескрипторам в поисках открытых сокетов
	for _, entry := range entries {
		fd := path.Join(fdDir, entry.Name())
		// Получаем линк на файл
		socketLink, err := os.Readlink(fd)
		// Если не сокет, или не смогли - просто игнорируем
		if err != nil || !strings.HasPrefix(socketLink, sockPrefix) {
			continue
		}
		// Бежим по открытым соединениям процесса
		for i := range p.connections {
			// Если этот сокет принадлежит этому процессу....
			conn := &p.connections[i]
			ss := sockPrefix + conn.SocketID + "]"
			if ss != socketLink {
				continue
			}
			if p.p == nil {
				p.p = &Process{p.pid, cmdLink}
			}
			conn.Process = p.p
		}
	}
}

func (l *LinuxConnectionsCollector) extractProcInfo(connections ConnectionsStat) {
	baseDir := l.cfg.System.Proc
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		processRootDir := path.Join(baseDir, entry.Name())
		process := runningProcess{rootDir: processRootDir, pid: pid, connections: connections}
		process.getProcessDetail()
	}
}

func (l *LinuxConnectionsCollector) netstat(path string, protocol string) ([]Connection, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	connections, err := parseConnectionsFile(f, protocol)
	if err != nil {
		return nil, err
	}
	// Обогащаем соединения информацией о процессе, который его открыл
	l.extractProcInfo(connections)
	return connections, nil
}

func (l *LinuxConnectionsCollector) TCPSocks() (ConnectionsStat, error) {
	c, err := l.netstat(l.cfg.System.TCP, ProtocolTCP)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (l *LinuxConnectionsCollector) TCP6Socks() (ConnectionsStat, error) {
	c, err := l.netstat(l.cfg.System.TCP6, ProtocolTCP6)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (l *LinuxConnectionsCollector) UDPSocks() (ConnectionsStat, error) {
	c, err := l.netstat(l.cfg.System.UDP, ProtocolUDP)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (l *LinuxConnectionsCollector) UDP6Socks() (ConnectionsStat, error) {
	c, err := l.netstat(l.cfg.System.UDP6, ProtocolUDP6)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (l *LinuxConnectionsCollector) GetConnection() (ConnectionsStat, error) {
	connections, err := l.TCPSocks()
	if err != nil {
		return nil, err
	}

	c, err := l.TCP6Socks()
	if err != nil {
		return nil, err
	}
	connections = append(connections, c...)

	c, err = l.UDPSocks()
	if err != nil {
		return nil, err
	}
	connections = append(connections, c...)

	c, err = l.UDP6Socks()
	if err != nil {
		return nil, err
	}
	connections = append(connections, c...)

	return connections, nil
}

type LinuxConnectionsCollector struct {
	serverCtx context.Context
	clientCtx context.Context
	cfg       *config.DaemonConfig
	l         logger.Logger
}

func NewLinuxConnectionsCollector(
	serverCtx, clientCtx context.Context, cfg *config.DaemonConfig, l logger.Logger,
) *LinuxConnectionsCollector {
	return &LinuxConnectionsCollector{
		serverCtx: serverCtx,
		clientCtx: clientCtx,
		cfg:       cfg,
		l:         l,
	}
}

func (l *LinuxConnectionsCollector) Run() (<-chan ConnectionsStat, error) {
	if _, err := l.GetConnection(); err != nil {
		l.cfg.Metrics.NetConnections = false
		l.cfg.Metrics.NetConnectionsStates = false
		return nil, err
	}
	ch := make(chan ConnectionsStat)
	ticker := time.NewTicker(time.Second)

	go func() {
		defer close(ch)
		for {
			select {
			case <-l.serverCtx.Done():
			case <-l.clientCtx.Done():
				l.l.Debug("connections collector stopped")
				return
			case <-ticker.C:
				if !l.cfg.Metrics.NetConnections && !l.cfg.Metrics.NetConnectionsStates {
					continue
				}
				stat, err := l.GetConnection()
				if err != nil {
					l.l.Error("connections collector error", "error", err.Error())
					l.cfg.Metrics.NetConnections = false
					l.cfg.Metrics.NetConnectionsStates = false
					return
				}
				ch <- stat
			}
		}
	}()
	return ch, nil
}
