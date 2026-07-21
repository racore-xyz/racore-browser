package peer

import (
	"sync"
	"time"

	"github.com/racore/god/pkg/api"
)

type EventHandler func(api.Peer)

type PeerStore struct {
	mu       sync.RWMutex
	peers    map[string]*api.Peer
	onJoined EventHandler
	onLeft   EventHandler
}

func NewPeerStore() *PeerStore {
	return &PeerStore{
		peers: make(map[string]*api.Peer),
	}
}

func (ps *PeerStore) OnJoined(h EventHandler) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.onJoined = h
}

func (ps *PeerStore) OnLeft(h EventHandler) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.onLeft = h
}

func (ps *PeerStore) AddOrUpdate(p api.Peer) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	existing, exists := ps.peers[p.NodeID]
	if exists {
		existing.Name = p.Name
		existing.Address = p.Address
		existing.PublicKey = p.PublicKey
		existing.Roles = p.Roles
		existing.LastSeen = p.LastSeen
	} else {
		cp := p
		ps.peers[p.NodeID] = &cp
	}
	return !exists
}

func (ps *PeerStore) Get(nodeID string) (api.Peer, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	p, ok := ps.peers[nodeID]
	if !ok {
		return api.Peer{}, false
	}
	return *p, true
}

func (ps *PeerStore) List() []api.Peer {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	out := make([]api.Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		out = append(out, *p)
	}
	return out
}

func (ps *PeerStore) Count() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return len(ps.peers)
}

func (ps *PeerStore) Expire(maxAge time.Duration) []api.Peer {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	cutoff := time.Now().UnixMilli() - maxAge.Milliseconds()
	var expired []api.Peer
	for id, p := range ps.peers {
		if p.LastSeen < cutoff {
			delete(ps.peers, id)
			expired = append(expired, *p)
		}
	}
	return expired
}

func (ps *PeerStore) FireJoined(p api.Peer) {
	ps.mu.RLock()
	h := ps.onJoined
	ps.mu.RUnlock()
	if h != nil {
		h(p)
	}
}

func (ps *PeerStore) FireLeft(p api.Peer) {
	ps.mu.RLock()
	h := ps.onLeft
	ps.mu.RUnlock()
	if h != nil {
		h(p)
	}
}

func (ps *PeerStore) PeersMap() map[string]api.Peer {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	out := make(map[string]api.Peer, len(ps.peers))
	for k, v := range ps.peers {
		out[k] = *v
	}
	return out
}
