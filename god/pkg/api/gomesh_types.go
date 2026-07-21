package api

type Peer struct {
	NodeID    string   `json:"nodeId"`
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	PublicKey string   `json:"publicKey"`
	Roles     []string `json:"roles"`
	LastSeen  int64    `json:"lastSeen"`
	LatencyMs int      `json:"latencyMs"`
}

type NodeStatus struct {
	Online       bool     `json:"online"`
	NodeID       string   `json:"nodeId"`
	NodeName     string   `json:"nodeName"`
	PeerCount    int      `json:"peers"`
	Roles        []string `json:"roles"`
	UptimeSec    int64    `json:"uptimeSeconds"`
	Transport    string   `json:"transport"`
}

type Envelope struct {
	Protocol  string         `json:"protocol"`
	Type      string         `json:"type"`
	NodeID    string         `json:"nodeId"`
	Name      string         `json:"name"`
	PublicKey string         `json:"publicKey"`
	Roles     []string       `json:"roles"`
	Timestamp int64          `json:"timestamp"`
	Data      map[string]any `json:"data"`
	Signature string         `json:"signature"`
}

type WSEvent struct {
	Type        string `json:"type"`
	Timestamp   int64  `json:"timestamp"`
}

type PeerEvent struct {
	Type  string `json:"type"`
	Peer  Peer   `json:"peer,omitempty"`
}

type MessageEvent struct {
	Type        string `json:"type"`
	MessageType string `json:"messageType,omitempty"`
	NodeID      string `json:"nodeId,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

type ErrorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
