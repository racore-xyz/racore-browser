# Types

This document defines all core types shared between the Go backend and TypeScript frontend.

---

## Go Types (`god/pkg/api/types.go`)

### Config

```go
type Config struct {
    Host              string   `json:"host"`
    Port              int      `json:"port"`
    NodeName          string   `json:"nodeName"`
    MeshEnabled       bool     `json:"meshEnabled"`
    MeshGroup         string   `json:"meshGroup"`
    MeshPort          int      `json:"meshPort"`
    MeshAPIPort       int      `json:"meshAPIPort"`
    MeshHeartbeatSec  int      `json:"meshHeartbeatSec"`
    IPFSAPI           string   `json:"ipfsAPI"`
    IPFSGateway       string   `json:"ipfsGateway"`
    BootstrapPeers    []string `json:"bootstrapPeers"`
    DataDir           string   `json:"dataDir"`
}
```

### Provider

```go
type Provider struct {
    ID           string `json:"id"`
    Name         string `json:"name"`
    Kind         string `json:"kind"`
    BaseURL      string `json:"baseURL"`
    DefaultModel string `json:"defaultModel"`
    Auth         string `json:"auth"`
    Free         bool   `json:"free"`
    Local        bool   `json:"local"`
    Connected    bool   `json:"connected"`
    MaskedKey    string `json:"maskedKey"`
}
```

### ChatRequest

```go
type ChatRequest struct {
    Provider string              `json:"provider"`
    Model    string              `json:"model,omitempty"`
    Messages []map[string]string `json:"messages"`
    System   string              `json:"system,omitempty"`
}
```

### ChatResponse

```go
type ChatResponse struct {
    Text      string         `json:"text"`
    Usage     map[string]int `json:"usage,omitempty"`
    Provider  string         `json:"provider"`
    Model     string         `json:"model"`
    LatencyMs int64          `json:"latencyMs"`
    RawID     string         `json:"rawId,omitempty"`
}
```

### HealthResponse

```go
type HealthResponse struct {
    OK        bool   `json:"ok"`
    Version   string `json:"version"`
    Mesh      any    `json:"mesh,omitempty"`
    IPFS      any    `json:"ipfs,omitempty"`
    Providers int    `json:"providers,omitempty"`
}
```

### DomainRecord

```go
type DomainRecord struct {
    Domain     string `json:"domain"`
    Controller string `json:"controller"`
    CreatedAt  string `json:"createdAt"`
    Status     string `json:"status"`
}
```

### ReleaseManifest

```go
type ReleaseManifest struct {
    Protocol    string `json:"protocol"`
    Domain      string `json:"domain"`
    Version     string `json:"version"`
    CID         string `json:"cid"`
    ContentRoot string `json:"contentRoot"`
    Entrypoint  string `json:"entrypoint"`
    Files       []File `json:"files,omitempty"`
    Size        int64  `json:"size,omitempty"`
    CreatedAt   string `json:"createdAt"`
    Parent      string `json:"parent,omitempty"`
    Signature   string `json:"signature,omitempty"`
}
```

### DelegateGrant

```go
type DelegateGrant struct {
    PublicKey    string   `json:"publicKey"`
    Capabilities []string `json:"capabilities"`
    ExpiresAt    string   `json:"expiresAt"`
    CreatedAt    string   `json:"createdAt"`
}
```

---

## Mesh Types (`god/pkg/api/gomesh_types.go`)

### Envelope

```go
type Envelope struct {
    Sender    string `json:"sender"`
    Recipient string `json:"recipient,omitempty"`
    Type      string `json:"type"`
    Payload   []byte `json:"payload"`
    Timestamp int64  `json:"timestamp"`
    Signature []byte `json:"signature"`
}
```

### Heartbeat

```go
type Heartbeat struct {
    NodeID      string   `json:"nodeId"`
    NodeName    string   `json:"nodeName"`
    Address     string   `json:"address"`
    Roles       []string `json:"roles"`
    Uptime      int64    `json:"uptime"`
    Timestamp   int64    `json:"timestamp"`
}
```

### PeerInfo

```go
type PeerInfo struct {
    NodeID   string `json:"nodeId"`
    NodeName string `json:"nodeName"`
    Address  string `json:"address"`
    Roles    string `json:"roles,omitempty"`
    LastSeen int64  `json:"lastSeen"`
}
```

---

## TypeScript desktop types (`app/lib/desktop.ts`)

### Desktop bridge

```typescript
type DaemonApiRequest = {
  path: string;
  method?: string;
  body?: unknown;
};

type DaemonApiResponse = {
  ok: boolean;
  status: number;
  data: unknown;
};

type DaemonExitPayload = {
  code: number | null;
  success: boolean;
}
```

### Racore Client (`app/lib/racore-client.ts`)

```typescript
interface ProviderInfo {
  id: string;
  name: string;
  kind: string;
  defaultModel: string;
  free: boolean;
  local: boolean;
  connected: boolean;
  maskedKey?: string | null;
}
```

---

## Wire Format

### Mesh Message Wire Format

A mesh message on the UDP wire is:

```
[JSON-encoded Envelope]
```

The envelope payload contains a JSON-encoded inner message (e.g., Heartbeat). The entire envelope is signed with the sender's Ed25519 private key.

### Event Wire Format (WebSocket)

```json
{
  "type": "event.name",
  "timestamp": "2026-07-21T13:00:00Z",
  "data": { ... }
}
```

### Error Wire Format

```json
{
  "detail": "error description"
}
```

### Multipart Upload Wire Format

For `POST /v1/ipfs/add`, the request body is `multipart/form-data` with a single file field. The response is JSON.
