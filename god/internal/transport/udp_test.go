package transport

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNewUDPTransport(t *testing.T) {
	tport, err := NewUDPTransport("239.255.77.77", 47777)
	if err != nil {
		t.Fatalf("NewUDPTransport: %v", err)
	}
	if tport == nil {
		t.Fatal("expected non-nil transport")
	}
	if tport.group.Port != 47777 {
		t.Fatalf("expected port 47777, got %d", tport.group.Port)
	}
}

func TestInvalidGroup(t *testing.T) {
	_, err := NewUDPTransport("not-a-group", 0)
	if err == nil {
		t.Fatal("expected error for invalid group")
	}
}

func TestSendToNotStarted(t *testing.T) {
	tport, err := NewUDPTransport("239.255.77.77", 47999)
	if err != nil {
		t.Fatal(err)
	}
	err = tport.SendTo([]byte("hello"), &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9999})
	if err == nil {
		t.Fatal("expected error when sending on non-started transport")
	}
}

func TestBroadcastNotStarted(t *testing.T) {
	tport, err := NewUDPTransport("239.255.77.77", 47999)
	if err != nil {
		t.Fatal(err)
	}
	err = tport.Broadcast([]byte("hello"))
	if err == nil {
		t.Fatal("expected error when broadcasting on non-started transport")
	}
}

func TestCloseNotStarted(t *testing.T) {
	tport, err := NewUDPTransport("239.255.77.77", 47999)
	if err != nil {
		t.Fatal(err)
	}
	err = tport.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestBroadcastToGroup(t *testing.T) {
	tport, err := NewUDPTransport("239.255.77.77", 47999)
	if err != nil {
		t.Fatal(err)
	}

	if tport.group.IP.String() != "239.255.77.77" {
		t.Fatalf("expected 239.255.77.77, got %s", tport.group.IP)
	}
	if tport.group.Port != 47999 {
		t.Fatalf("expected 47999, got %d", tport.group.Port)
	}
}

func TestPortZero(t *testing.T) {
	tport, err := NewUDPTransport("239.255.77.77", 0)
	if err != nil {
		t.Fatal(err)
	}
	if tport.group.Port != 0 {
		t.Fatalf("expected port 0, got %d", tport.group.Port)
	}
}

func TestBufferPool(t *testing.T) {
	tport, _ := NewUDPTransport("239.255.77.77", 0)

	bufPtr := tport.readBuf.Get().(*[]byte)
	buf := *bufPtr
	if len(buf) != 65535 {
		t.Fatalf("expected 65535 buffer, got %d", len(buf))
	}
	tport.readBuf.Put(bufPtr)
}

func TestDoubleClose(t *testing.T) {
	tport, _ := NewUDPTransport("239.255.77.77", 0)
	tport.Close()
	tport.Close()
}

func TestStoppedReadLoop(t *testing.T) {
	tport, err := NewUDPTransport("239.255.77.77", 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := tport.Start(); err != nil {
		t.Skipf("Start failed (no multicast): %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	recvChan := make(chan Message, 16)
	tport.StartReadLoop(ctx, recvChan)

	// Allow read loop to start
	time.Sleep(50 * time.Millisecond)

	// Cancel context - this should stop the read loop
	cancel()

	// Then close - should be safe
	if err := tport.Close(); err != nil {
		t.Fatalf("Close after cancel: %v", err)
	}
}
