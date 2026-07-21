package mesh

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/racore/god/internal/identity"
	"github.com/racore/god/internal/peer"
	"github.com/racore/god/internal/protocol"
	"github.com/racore/god/internal/transport"
	"github.com/racore/god/pkg/api"
)

type EventHandler func(event map[string]any)

type MeshNode struct {
	identity  *identity.NodeIdentity
	peerStore *peer.PeerStore
	transport *transport.UDPTransport
	config    api.Config
	startedAt time.Time
	running   bool
	mu        sync.RWMutex
	cancel    context.CancelFunc
	eventSink EventHandler
	roles     []string
}

func NewMeshNode(cfg api.Config) *MeshNode {
	return &MeshNode{
		config:    cfg,
		peerStore: peer.NewPeerStore(),
		roles:     []string{"client", "cache"},
	}
}

func (m *MeshNode) Identity() *identity.NodeIdentity {
	return m.identity
}

func (m *MeshNode) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	ident, err := identity.NewNodeIdentity(m.config.DataDir)
	if err != nil {
		return fmt.Errorf("identity: %w", err)
	}

	tport, err := transport.NewUDPTransport(m.config.MeshGroup, m.config.MeshPort)
	if err != nil {
		return fmt.Errorf("transport: %w", err)
	}

	if err := tport.Start(); err != nil {
		return fmt.Errorf("transport start: %w", err)
	}

	m.mu.Lock()
	m.identity = ident
	m.transport = tport
	meshCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.startedAt = time.Now()
	m.running = true
	m.mu.Unlock()

	m.peerStore.OnJoined(func(p api.Peer) {
		m.emit(map[string]any{
			"type": "mesh.peer.joined",
			"peer": p,
		})
	})

	m.peerStore.OnLeft(func(p api.Peer) {
		m.emit(map[string]any{
			"type": "mesh.peer.left",
			"peer": p,
		})
	})

	recvChan := make(chan transport.Message, 256)
	tport.StartReadLoop(meshCtx, recvChan)

	go m.networkLoop(meshCtx, recvChan)
	go m.heartbeatLoop(meshCtx)

	return nil
}

func (m *MeshNode) networkLoop(ctx context.Context, recvChan <-chan transport.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-recvChan:
			if !ok {
				return
			}
			m.handleMessage(msg)
		}
	}
}

func (m *MeshNode) handleMessage(msg transport.Message) {
	env, err := protocol.VerifyEnvelope(msg.Data)
	if err != nil {
		return
	}

	m.mu.RLock()
	selfID := m.identity.NodeID()
	m.mu.RUnlock()

	if env.NodeID == selfID {
		return
	}

	now := time.Now().UnixMilli()
	p := api.Peer{
		NodeID:    env.NodeID,
		Name:      env.Name,
		Address:   msg.Addr.IP.String(),
		PublicKey: env.PublicKey,
		Roles:     env.Roles,
		LastSeen:  now,
	}

	isNew := m.peerStore.AddOrUpdate(p)
	if isNew {
		m.peerStore.FireJoined(p)
	}

	m.emit(map[string]any{
		"type":        "mesh.message",
		"messageType": env.Type,
		"nodeId":      env.NodeID,
		"data":        env.Data,
	})
}

func (m *MeshNode) heartbeatLoop(ctx context.Context) {
	interval := time.Duration(	m.config.MeshHeartbeatSec) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.broadcastHeartbeat()
			m.expirePeers()
		}
	}
}

func (m *MeshNode) broadcastHeartbeat() {
	m.mu.RLock()
	name := m.config.NodeName
	roles := m.roles
	m.mu.RUnlock()

	data := map[string]any{
		"apiPort": m.config.Port,
	}
	env, err := protocol.NewEnvelope("presence", data, m.identity, name, roles)
	if err != nil {
		return
	}
	raw := protocol.EnvelopeToMessage(env)
	if err := m.transport.Broadcast(raw); err != nil {
		m.emit(map[string]any{"type": "mesh.error", "message": err.Error()})
	}
}

func (m *MeshNode) expirePeers() {
	maxAge := 20 * time.Second
	hb := 	m.config.MeshHeartbeatSec
	if hb > 0 {
		alt := time.Duration(hb*4) * time.Second
		if alt > maxAge {
			maxAge = alt
		}
	}
	expired := m.peerStore.Expire(maxAge)
	for _, p := range expired {
		m.peerStore.FireLeft(p)
	}
}

func (m *MeshNode) Broadcast(msgType string, data map[string]any) (*api.Envelope, error) {
	m.mu.RLock()
	name := m.config.NodeName
	roles := m.roles
	ident := m.identity
	tport := m.transport
	m.mu.RUnlock()

	if ident == nil {
		return nil, fmt.Errorf("mesh not started")
	}

	env, err := protocol.NewEnvelope(msgType, data, ident, name, roles)
	if err != nil {
		return nil, err
	}
	raw := protocol.EnvelopeToMessage(env)
	if err := tport.Broadcast(raw); err != nil {
		return nil, err
	}
	m.emit(map[string]any{
		"type":        "mesh.broadcast",
		"messageType": msgType,
		"data":        data,
	})
	return env, nil
}

func (m *MeshNode) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return nil
	}
	m.running = false
	if m.cancel != nil {
		m.cancel()
	}
	if m.transport != nil {
		return m.transport.Close()
	}
	return nil
}

func (m *MeshNode) Status() api.NodeStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	uptime := int64(time.Since(m.startedAt).Seconds())
	nodeID := ""
	if m.identity != nil {
		nodeID = m.identity.NodeID()
	}
	return api.NodeStatus{
		Online:    m.running,
		NodeID:    nodeID,
		NodeName:  m.config.NodeName,
		PeerCount: m.peerStore.Count(),
		Roles:     m.roles,
		UptimeSec: uptime,
		Transport: "udp-multicast+rmp/0.1",
	}
}

func (m *MeshNode) Peers() []api.Peer {
	return m.peerStore.List()
}

func (m *MeshNode) emit(event map[string]any) {
	m.mu.RLock()
	sink := m.eventSink
	m.mu.RUnlock()
	if sink != nil {
		event["timestamp"] = time.Now().UnixMilli()
		sink(event)
	}
}

func (m *MeshNode) SetEventSink(sink EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventSink = sink
}
