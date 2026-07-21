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
		return nil, err
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
		return err
	}
	t.conn = conn

	file, err := conn.File()
	if err != nil {
		conn.Close()
		return err
	}
	defer file.Close()

	if err := syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		conn.Close()
		return err
	}

	pc := ipv4.NewPacketConn(conn)
	t.pc = pc

	if err := pc.JoinGroup(nil, &net.UDPAddr{IP: t.group.IP}); err != nil {
		pc.Close()
		conn.Close()
		return err
	}

	if err := pc.SetMulticastTTL(2); err != nil {
		pc.Close()
		conn.Close()
		return err
	}

	if err := pc.SetControlMessage(ipv4.FlagTTL|ipv4.FlagDst, true); err != nil {
		pc.Close()
		conn.Close()
		return err
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
		n, _, addr, err := t.pc.ReadFrom(buf)
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
	if t.closed.Load() || t.conn == nil {
		return net.ErrClosed
	}
	_, err := t.conn.WriteTo(data, addr)
	return err
}

func (t *UDPTransport) Broadcast(data []byte) error {
	if t.closed.Load() || t.pc == nil {
		return net.ErrClosed
	}
	_, err := t.pc.WriteTo(data, nil, t.group)
	return err
}

func (t *UDPTransport) Close() error {
	t.closed.Store(true)
	if t.conn != nil {
		if raw, err := t.conn.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				syscall.Shutdown(int(fd), syscall.SHUT_RD)
			})
		}
	}
	if t.pc != nil {
		return t.pc.Close()
	}
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}
