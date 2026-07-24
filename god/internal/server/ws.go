package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	clientSendBuffer = 64
	writeWait        = 10 * time.Second
	pongWait         = 60 * time.Second
	pingPeriod       = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		return originAllowed(origin)
	},
}

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
	once sync.Once
}

func (c *wsClient) close() {
	c.once.Do(func() { close(c.send) })
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*wsClient]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*wsClient]struct{}),
	}
}

func (h *Hub) add(c *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) remove(c *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		c.close()
	}
}

func (h *Hub) Broadcast(event map[string]any) {
	event["timestamp"] = time.Now().UnixMilli()

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	var stale []*wsClient
	h.mu.RLock()
	for c := range h.clients {
		select {
		case c.send <- data:
		default:
			stale = append(stale, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range stale {
		log.Printf("ws: dropping slow client")
		h.remove(c)
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case data, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}

	client := &wsClient{conn: conn, send: make(chan []byte, clientSendBuffer)}
	s.hub.add(client)

	ready, _ := json.Marshal(map[string]any{
		"type": "racore.ready", "version": s.version,
		"mesh": s.mesh.Status(),
	})
	client.send <- ready

	go func() {
		client.writePump()
		conn.Close()
	}()

	go func() {
		defer func() {
			s.hub.remove(client)
			conn.Close()
		}()
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(pongWait))
		})
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
