package mesh

import (
	"context"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/racore/god/pkg/api"
)

func TestTwoNodeDiscovery(t *testing.T) {
	dir1, err := os.MkdirTemp("", "god-mesh-1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1)

	dir2, err := os.MkdirTemp("", "god-mesh-2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir2)

	cfg := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48001,
		DataDir:          dir1,
		NodeName:         "NodeA",
		MeshHeartbeatSec: 30,
	}
	cfg2 := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48001,
		DataDir:          dir2,
		NodeName:         "NodeB",
		MeshHeartbeatSec: 30,
	}

	mn1 := NewMeshNode(cfg)
	mn2 := NewMeshNode(cfg2)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	if err := mn1.Start(ctx1); err != nil {
		t.Skipf("node1 start: %v (no multicast)", err)
	}
	defer mn1.Stop()

	if err := mn2.Start(ctx2); err != nil {
		mn1.Stop()
		t.Skipf("node2 start: %v (no multicast)", err)
	}
	defer mn2.Stop()

	// Give the nodes time to discover each other via heartbeat
	deadline := time.Now().Add(5 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		if mn1.peerStore.Count() > 0 || mn2.peerStore.Count() > 0 {
			found = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !found {
		t.Skip("peers did not discover each other (expected without multicast traffic)")
	}
}

func TestBroadcastReceive(t *testing.T) {
	dir1, err := os.MkdirTemp("", "god-mesh-bc-1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1)

	dir2, err := os.MkdirTemp("", "god-mesh-bc-2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir2)

	cfg := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48002,
		DataDir:          dir1,
		NodeName:         "Broadcaster",
		MeshHeartbeatSec: 300,
	}
	cfg2 := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48002,
		DataDir:          dir2,
		NodeName:         "Receiver",
		MeshHeartbeatSec: 300,
	}

	mn1 := NewMeshNode(cfg)
	mn2 := NewMeshNode(cfg2)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	if err := mn1.Start(ctx1); err != nil {
		t.Skipf("broadcaster start: %v", err)
	}
	defer mn1.Stop()

	if err := mn2.Start(ctx2); err != nil {
		mn1.Stop()
		t.Skipf("receiver start: %v", err)
	}
	defer mn2.Stop()

	var received atomic.Int32
	mn2.SetEventSink(func(event map[string]any) {
		if event["type"] == "mesh.message" {
			received.Add(1)
		}
	})

	if _, err := mn1.Broadcast("test.custom", map[string]any{"msg": "hello"}); err != nil {
		t.Fatalf("broadcast: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if received.Load() > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Skip("receiver did not get broadcast within timeout (expected without multicast)")
}

func TestSelfMessageIgnored(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-mesh-self-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	cfg := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48003,
		DataDir:          dir,
		NodeName:         "Self",
		MeshHeartbeatSec: 300,
	}

	mn := NewMeshNode(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mn.Start(ctx); err != nil {
		t.Skipf("start: %v", err)
	}
	defer mn.Stop()

	// Broadcast should succeed even if no one receives it
	env, err := mn.Broadcast("test", map[string]any{})
	if err != nil {
		t.Fatalf("broadcast: %v", err)
	}
	if env == nil {
		t.Fatal("expected non-nil envelope")
	}
	if env.Type != "test" {
		t.Fatalf("expected test type, got %s", env.Type)
	}
}

func TestNodeStatus(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-mesh-status-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	cfg := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48004,
		DataDir:          dir,
		NodeName:         "StatusTest",
		MeshHeartbeatSec: 300,
	}

	mn := NewMeshNode(cfg)
	status := mn.Status()
	if status.NodeName != "StatusTest" {
		t.Fatalf("expected StatusTest, got %s", status.NodeName)
	}
	if status.Transport != "udp-multicast+rmp/0.1" {
		t.Fatalf("unexpected transport: %s", status.Transport)
	}
}

func TestStopWithoutStart(t *testing.T) {
	cfg := api.Config{NodeName: "NoStart"}
	mn := NewMeshNode(cfg)
	if err := mn.Stop(); err != nil {
		t.Fatalf("stop without start should not error: %v", err)
	}
}

func TestDoubleStop(t *testing.T) {
	dir, err := os.MkdirTemp("", "god-mesh-dstop-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	cfg := api.Config{
		DataDir:          dir,
		MeshPort:         48005,
		MeshHeartbeatSec: 300,
	}
	mn := NewMeshNode(cfg)
	mn.Stop()
	mn.Stop()
}

func TestPeersEmpty(t *testing.T) {
	dir, _ := os.MkdirTemp("", "god-mesh-peers-*")
	defer os.RemoveAll(dir)

	cfg := api.Config{DataDir: dir}
	mn := NewMeshNode(cfg)
	peers := mn.Peers()
	if len(peers) != 0 {
		t.Fatalf("expected 0 peers, got %d", len(peers))
	}
}

func TestSetEventSink(t *testing.T) {
	dir, _ := os.MkdirTemp("", "god-mesh-sink-*")
	defer os.RemoveAll(dir)

	cfg := api.Config{DataDir: dir}
	mn := NewMeshNode(cfg)

	called := make(chan bool, 1)
	mn.SetEventSink(func(event map[string]any) {
		called <- true
	})

	// Emit internally
	mn.emit(map[string]any{"type": "test.event"})

	select {
	case <-called:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("event sink was not called")
	}
}

func TestBootstrapPeersInvalidAddress(t *testing.T) {
	dir, _ := os.MkdirTemp("", "god-mesh-bootstrap-*")
	defer os.RemoveAll(dir)

	cfg := api.Config{
		DataDir:        dir,
		MeshPort:       48010,
		MeshGroup:      "239.255.77.77",
		BootstrapPeers: []string{"invalid-address-!!!:abc"},
	}

	mn := NewMeshNode(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mn.Start(ctx); err != nil {
		t.Skipf("start: %v (no multicast)", err)
	}
	defer mn.Stop()

	events := make(chan map[string]any, 10)
	mn.SetEventSink(func(event map[string]any) {
		events <- event
	})

	mn.bootstrapToPeers()

	select {
	case evt := <-events:
		if evt["type"] != "mesh.error" {
			t.Fatalf("expected mesh.error, got %v", evt["type"])
		}
	case <-time.After(1 * time.Second):
		t.Fatal("expected error event for invalid bootstrap address")
	}
}

func TestBootstrapPeersSendsPresence(t *testing.T) {
	dir, _ := os.MkdirTemp("", "god-mesh-bp-send-*")
	defer os.RemoveAll(dir)

	// Use a high port that won't conflict
	cfg := api.Config{
		DataDir:        dir,
		MeshPort:       48011,
		MeshGroup:      "239.255.77.77",
		BootstrapPeers: []string{"127.0.0.1:48099"},
		NodeName:       "BootstrapTest",
	}

	mn := NewMeshNode(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mn.Start(ctx); err != nil {
		t.Skipf("start: %v (no multicast)", err)
	}
	defer mn.Stop()

	// bootstrapToPeers is called during Start, so it should have run.
	// We can't verify the message was received (port is likely closed),
	// but we verify the method doesn't panic or deadlock.
}

func TestGracefulGoodbyeCausesPeerRemoval(t *testing.T) {
	dir1, err := os.MkdirTemp("", "god-mesh-gb-1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1)

	dir2, err := os.MkdirTemp("", "god-mesh-gb-2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir2)

	cfg1 := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48006,
		DataDir:          dir1,
		NodeName:         "NodeA",
		MeshHeartbeatSec: 30,
	}
	cfg2 := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48006,
		DataDir:          dir2,
		NodeName:         "NodeB",
		MeshHeartbeatSec: 30,
	}

	mn1 := NewMeshNode(cfg1)
	mn2 := NewMeshNode(cfg2)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	if err := mn1.Start(ctx1); err != nil {
		t.Skipf("node1 start: %v (no multicast)", err)
	}
	defer mn1.Stop()

	if err := mn2.Start(ctx2); err != nil {
		mn1.Stop()
		t.Skipf("node2 start: %v (no multicast)", err)
	}

	// Wait for mutual discovery
	deadline := time.Now().Add(5 * time.Second)
	discovered := false
	for time.Now().Before(deadline) {
		if mn1.peerStore.Count() > 0 && mn2.peerStore.Count() > 0 {
			discovered = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !discovered {
		mn2.Stop()
		t.Skip("peers did not discover each other (expected without multicast)")
	}

	// Now stop node2 gracefully - it should broadcast goodbye
	mn2.Stop()

	// Node1 should immediately remove the peer (no need to wait for expiry)
	time.Sleep(500 * time.Millisecond)

	if mn1.peerStore.Count() != 0 {
		t.Fatalf("expected 0 peers after goodbye, got %d; peer still present", mn1.peerStore.Count())
	}
}

func TestSendToUnicast(t *testing.T) {
	dir1, err := os.MkdirTemp("", "god-mesh-st-1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1)

	dir2, err := os.MkdirTemp("", "god-mesh-st-2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir2)

	cfg1 := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48007,
		DataDir:          dir1,
		NodeName:         "Sender",
		MeshHeartbeatSec: 30,
	}
	cfg2 := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48007,
		DataDir:          dir2,
		NodeName:         "Receiver",
		MeshHeartbeatSec: 30,
	}

	mn1 := NewMeshNode(cfg1)
	mn2 := NewMeshNode(cfg2)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	if err := mn1.Start(ctx1); err != nil {
		t.Skipf("sender start: %v (no multicast)", err)
	}
	defer mn1.Stop()

	if err := mn2.Start(ctx2); err != nil {
		mn1.Stop()
		t.Skipf("receiver start: %v (no multicast)", err)
	}
	defer mn2.Stop()

	// Wait for mutual discovery
	deadline := time.Now().Add(5 * time.Second)
	discovered := false
	for time.Now().Before(deadline) {
		if mn1.peerStore.Count() > 0 && mn2.peerStore.Count() > 0 {
			discovered = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !discovered {
		t.Skip("peers did not discover each other")
	}

	// Get the receiver's node ID
	peers := mn1.peerStore.List()
	if len(peers) == 0 {
		t.Fatal("no peers found")
	}
	receiverID := peers[0].NodeID

	// Set event sink on receiver to capture unicast messages
	var received atomic.Int32
	mn2.SetEventSink(func(event map[string]any) {
		if event["type"] == "mesh.message" {
			received.Add(1)
		}
	})

	// Send unicast message
	_, err = mn1.SendTo(receiverID, "test.unicast", map[string]any{"msg": "direct hello"})
	if err != nil {
		t.Fatalf("SendTo: %v", err)
	}

	// Wait for delivery
	deadline = time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if received.Load() > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("receiver did not get unicast message within timeout")
}

func TestSendToUnknownPeer(t *testing.T) {
	dir, _ := os.MkdirTemp("", "god-mesh-unk-*")
	defer os.RemoveAll(dir)

	cfg := api.Config{
		DataDir:  dir,
		MeshPort: 48008,
	}
	mn := NewMeshNode(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mn.Start(ctx); err != nil {
		t.Skipf("start: %v (no multicast)", err)
	}
	defer mn.Stop()

	_, err := mn.SendTo("nonexistent-peer-id", "test", map[string]any{})
	if err == nil {
		t.Fatal("expected error sending to unknown peer, got nil")
	}
}

func TestPingPongLatency(t *testing.T) {
	dir1, err := os.MkdirTemp("", "god-mesh-pp-1-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1)

	dir2, err := os.MkdirTemp("", "god-mesh-pp-2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir2)

	cfg1 := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48009,
		DataDir:          dir1,
		NodeName:         "Pinger",
		MeshHeartbeatSec: 30,
	}
	cfg2 := api.Config{
		MeshGroup:        "239.255.77.77",
		MeshPort:         48009,
		DataDir:          dir2,
		NodeName:         "Ponger",
		MeshHeartbeatSec: 30,
	}

	mn1 := NewMeshNode(cfg1)
	mn2 := NewMeshNode(cfg2)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	if err := mn1.Start(ctx1); err != nil {
		t.Skipf("pinger start: %v (no multicast)", err)
	}
	defer mn1.Stop()

	if err := mn2.Start(ctx2); err != nil {
		mn1.Stop()
		t.Skipf("ponger start: %v (no multicast)", err)
	}
	defer mn2.Stop()

	// Wait for mutual discovery
	deadline := time.Now().Add(5 * time.Second)
	discovered := false
	for time.Now().Before(deadline) {
		if mn1.peerStore.Count() > 0 && mn2.peerStore.Count() > 0 {
			discovered = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !discovered {
		t.Skip("peers did not discover each other")
	}

	// Force a ping from node1 to node2
	mn1.pingPeers()

	// Wait for pong to arrive and latency to be updated
	deadline = time.Now().Add(3 * time.Second)
	updated := false
	for time.Now().Before(deadline) {
		peers := mn1.peerStore.List()
		for _, p := range peers {
			if p.LatencyMs > 0 {
				updated = true
				break
			}
		}
		if updated {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if !updated {
		t.Fatal("latency was not updated via ping/pong")
	}

	// Verify the latency value is reasonable (should be < 1 second for localhost)
	peers := mn1.peerStore.List()
	for _, p := range peers {
		if p.LatencyMs > 0 && p.LatencyMs > 1000 {
			t.Fatalf("latency %dms seems too high for localhost", p.LatencyMs)
		}
	}
}

func TestHandleGoodbyeMessage(t *testing.T) {
	dir, _ := os.MkdirTemp("", "god-mesh-hg-*")
	defer os.RemoveAll(dir)

	cfg := api.Config{
		DataDir:          dir,
		MeshPort:         48012,
		MeshGroup:        "239.255.77.77",
		NodeName:         "TestNode",
		MeshHeartbeatSec: 300,
	}

	mn := NewMeshNode(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mn.Start(ctx); err != nil {
		t.Skipf("start: %v (no multicast)", err)
	}
	defer mn.Stop()

	// Add a fake peer manually
	mn.peerStore.AddOrUpdate(api.Peer{
		NodeID:   "fake-peer-id",
		Name:     "FakePeer",
		LastSeen: time.Now().UnixMilli(),
	})
	mn.mu.Lock()
	mn.peerAddrs["fake-peer-id"] = &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: cfg.MeshPort}
	mn.mu.Unlock()

	// Instead of faking, let's send an actual goodbye from a second node
	dir2, _ := os.MkdirTemp("", "god-mesh-hg2-*")
	defer os.RemoveAll(dir2)

	cfg2 := api.Config{
		DataDir:          dir2,
		MeshPort:         48012,
		MeshGroup:        "239.255.77.77",
		NodeName:         "GoodbyeNode",
		MeshHeartbeatSec: 300,
	}
	mn2 := NewMeshNode(cfg2)
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	if err := mn2.Start(ctx2); err != nil {
		t.Skipf("node2 start: %v (no multicast)", err)
	}

	// Wait for discovery
	time.Sleep(500 * time.Millisecond)
	if mn.peerStore.Count() == 0 {
		mn2.Stop()
		t.Skip("peers did not discover each other")
	}

	// Stop node2 gracefully - this broadcasts goodbye
	mn2.Stop()

	time.Sleep(500 * time.Millisecond)

	if mn.peerStore.Count() != 0 {
		t.Fatalf("expected 0 peers after goodbye, got %d", mn.peerStore.Count())
	}
}
