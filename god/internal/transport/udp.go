package transport

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/ipv4"
)

type Message struct {
	Data []byte
	Addr *net.UDPAddr
}

type UDPTransport struct {
	group   *net.UDPAddr
	conn    *net.UDPConn
	pc      *ipv4.PacketConn
	closed  atomic.Bool
	started atomic.Bool
	readBuf sync.Pool
	mu      sync.Mutex
}

func NewUDPTransport(group string, port int) (*UDPTransport, error) {
	addr := net.JoinHostPort(group, fmt.Sprintf("%d", port))
	gaddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("resolve group %s: %w", addr, err)
	}
	return &UDPTransport{
		group: gaddr,
		readBuf: sync.Pool{
			New: func() any {
				b := make([]byte, 65535)
				return &b
			},
		},
	}, nil
}

func (t *UDPTransport) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.started.Load() {
		return nil
	}

	addr := &net.UDPAddr{Port: t.group.Port}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("listen udp %d: %w", t.group.Port, err)
	}
	t.conn = conn

	file, err := conn.File()
	if err != nil {
		conn.Close()
		return fmt.Errorf("conn file: %w", err)
	}
	defer file.Close()

	if err := syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		conn.Close()
		return fmt.Errorf("setsockopt reuseaddr: %w", err)
	}

	pc := ipv4.NewPacketConn(conn)
	t.pc = pc

	if err := pc.JoinGroup(nil, &net.UDPAddr{IP: t.group.IP}); err != nil {
		pc.Close()
		conn.Close()
		return fmt.Errorf("join group %s: %w", t.group.IP, err)
	}

	if err := pc.SetMulticastTTL(2); err != nil {
		pc.Close()
		conn.Close()
		return fmt.Errorf("set multicast ttl: %w", err)
	}

	if err := pc.SetControlMessage(ipv4.FlagTTL|ipv4.FlagDst, true); err != nil {
		pc.Close()
		conn.Close()
		return fmt.Errorf("set control message: %w", err)
	}

	t.started.Store(true)
	return nil
}

func (t *UDPTransport) StartReadLoop(ctx context.Context, recvChan chan<- Message) {
	go t.readLoop(ctx, recvChan)
}

func (t *UDPTransport) readLoop(ctx context.Context, recvChan chan<- Message) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if t.closed.Load() {
			return
		}
		bufPtr := t.readBuf.Get().(*[]byte)
		buf := *bufPtr
		t.mu.Lock()
		pc := t.pc
		t.mu.Unlock()
		if pc == nil {
			t.readBuf.Put(bufPtr)
			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}
		n, _, addr, err := pc.ReadFrom(buf)
		if err != nil || n == 0 || addr == nil {
			t.readBuf.Put(bufPtr)
			if t.closed.Load() {
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}
		udpAddr, ok := addr.(*net.UDPAddr)
		if !ok {
			t.readBuf.Put(bufPtr)
			continue
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		t.readBuf.Put(bufPtr)
		select {
		case recvChan <- Message{Data: data, Addr: udpAddr}:
		case <-ctx.Done():
			return
		}
	}
}

func (t *UDPTransport) SendTo(data []byte, addr *net.UDPAddr) error {
	t.mu.Lock()
	conn := t.conn
	t.mu.Unlock()
	if t.closed.Load() || conn == nil {
		return net.ErrClosed
	}
	_, err := conn.WriteTo(data, addr)
	if err != nil {
		return fmt.Errorf("sendto %s: %w", addr, err)
	}
	return nil
}

func (t *UDPTransport) Broadcast(data []byte) error {
	t.mu.Lock()
	pc := t.pc
	t.mu.Unlock()
	if t.closed.Load() || pc == nil {
		return net.ErrClosed
	}
	_, err := pc.WriteTo(data, nil, t.group)
	return err
}

func (t *UDPTransport) Close() error {
	t.closed.Store(true)
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn != nil {
		if raw, err := t.conn.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				syscall.Shutdown(int(fd), syscall.SHUT_RD)
			})
		}
	}
	var err error
	if t.pc != nil {
		err = t.pc.Close()
		t.pc = nil
	}
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
	return err
}
