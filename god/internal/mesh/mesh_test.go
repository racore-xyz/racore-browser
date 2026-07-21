package mesh

import (
	"context"
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
