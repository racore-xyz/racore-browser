package peer

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/racore/god/pkg/api"
)

func TestAddAndGet(t *testing.T) {
	ps := NewPeerStore()
	p := api.Peer{
		NodeID:    "abc123",
		Name:      "Test",
		Address:   "10.0.0.1",
		PublicKey: "key",
		Roles:     []string{"client"},
		LastSeen:  time.Now().UnixMilli(),
	}

	isNew := ps.AddOrUpdate(p)
	if !isNew {
		t.Fatal("expected new peer")
	}

	got, ok := ps.Get("abc123")
	if !ok {
		t.Fatal("peer not found")
	}
	if got.Name != "Test" {
		t.Fatalf("expected Test, got %s", got.Name)
	}
}

func TestUpdatePeer(t *testing.T) {
	ps := NewPeerStore()
	p1 := api.Peer{NodeID: "x", Name: "First", LastSeen: 1000}
	p2 := api.Peer{NodeID: "x", Name: "Updated", LastSeen: 2000}

	ps.AddOrUpdate(p1)
	isNew := ps.AddOrUpdate(p2)
	if isNew {
		t.Fatal("expected update, not new")
	}

	got, _ := ps.Get("x")
	if got.Name != "Updated" {
		t.Fatalf("expected Updated, got %s", got.Name)
	}
	if got.LastSeen != 2000 {
		t.Fatalf("expected 2000, got %d", got.LastSeen)
	}
}

func TestListAndCount(t *testing.T) {
	ps := NewPeerStore()
	ps.AddOrUpdate(api.Peer{NodeID: "a", LastSeen: 1})
	ps.AddOrUpdate(api.Peer{NodeID: "b", LastSeen: 2})
	ps.AddOrUpdate(api.Peer{NodeID: "c", LastSeen: 3})

	if ps.Count() != 3 {
		t.Fatalf("expected 3, got %d", ps.Count())
	}

	list := ps.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 in list, got %d", len(list))
	}
}

func TestExpire(t *testing.T) {
	ps := NewPeerStore()
	now := time.Now().UnixMilli()
	ps.AddOrUpdate(api.Peer{NodeID: "fresh", LastSeen: now})
	ps.AddOrUpdate(api.Peer{NodeID: "stale", LastSeen: now - 30000})

	expired := ps.Expire(10 * time.Second)
	if len(expired) != 1 {
		t.Fatalf("expected 1 expired, got %d", len(expired))
	}
	if expired[0].NodeID != "stale" {
		t.Fatalf("expected stale, got %s", expired[0].NodeID)
	}
	if ps.Count() != 1 {
		t.Fatalf("expected 1 remaining, got %d", ps.Count())
	}
}

func TestOnJoined(t *testing.T) {
	ps := NewPeerStore()
	var joined int32
	ps.OnJoined(func(p api.Peer) {
		atomic.AddInt32(&joined, 1)
	})

	ps.AddOrUpdate(api.Peer{NodeID: "a", LastSeen: 1})
	ps.FireJoined(api.Peer{NodeID: "a", LastSeen: 1})

	if atomic.LoadInt32(&joined) != 1 {
		t.Fatalf("expected 1 join event, got %d", joined)
	}
}

func TestOnLeft(t *testing.T) {
	ps := NewPeerStore()
	var leftCount int32
	ps.OnLeft(func(p api.Peer) {
		atomic.AddInt32(&leftCount, 1)
	})

	ps.FireLeft(api.Peer{NodeID: "gone", LastSeen: 1})

	if atomic.LoadInt32(&leftCount) != 1 {
		t.Fatalf("expected 1 left event, got %d", leftCount)
	}
}
