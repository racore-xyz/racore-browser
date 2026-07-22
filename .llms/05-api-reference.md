# API Reference

## Base URL

All API endpoints are served at `http://127.0.0.1:47831`. The port is configurable via the `RACORE_PORT` environment variable.

## Common Headers

| Header | Value | Notes |
|--------|-------|-------|
| `Content-Type` | `application/json` | Required for POST/PUT requests with a body |
| `Origin` | `http://localhost:3000` | Must be in the CORS whitelist for browser requests |

## CORS

Allowed origins:
- `http://localhost:3000` (Vite dev server)
- `http://127.0.0.1:3000`
- `http://[::1]:3000`
- Tauri desktop requests are proxied by Rust from trusted local application content.
- `https://racore.xyz`

Preflight requests (`OPTIONS`) respond with HTTP 204.

## Error Format

All errors follow a consistent format:
```json
{
  "detail": "human-readable error message"
}
```

HTTP status codes: 400 (bad request), 404 (not found), 500 (server error).

---

## Endpoints

### GET /health

Full system health check. Returns daemon version, mesh status, IPFS status, and provider count.

**Response 200:**
```json
{
  "ok": true,
  "version": "0.2.0",
  "mesh": {
    "online": true,
    "nodeId": "624c7cb66165fe9efed263db819be895",
    "nodeName": "Racore Desktop",
    "peers": 0,
    "roles": ["client", "cache"],
    "uptimeSeconds": 2,
    "transport": "udp-multicast+rmp/0.1"
  },
  "ipfs": {
    "online": false
  },
  "providers": 9
}
```

---

### GET /v1/providers

List all AI providers in the catalog.

**Response 200:**
```json
[
  {
    "id": "openai",
    "name": "OpenAI",
    "kind": "responses",
    "baseURL": "https://api.openai.com/v1",
    "defaultModel": "gpt-5.6-terra",
    "auth": "bearer",
    "free": false,
    "local": false,
    "connected": false,
    "maskedKey": ""
  },
  {
    "id": "ollama",
    "name": "Ollama",
    "kind": "openai",
    "baseURL": "http://127.0.0.1:11434/v1",
    "defaultModel": "qwen3:8b",
    "auth": "bearer",
    "free": true,
    "local": true,
    "connected": true,
    "maskedKey": ""
  }
]
```

**Fields:**
| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique provider identifier |
| `name` | string | Human-readable display name |
| `kind` | string | API protocol kind (`responses`, `anthropic`, `gemini`, `openrouter`) |
| `baseURL` | string | Base URL for API calls |
| `defaultModel` | string | Default model identifier |
| `auth` | string | Authentication method (`bearer`, `x-api-key`) |
| `free` | bool | Whether the provider has a free tier |
| `local` | bool | Whether the provider runs locally |
| `connected` | bool | Whether an API key has been configured |
| `maskedKey` | string | Masked API key (last 4 characters visible), empty if not connected |

---

### PUT /v1/providers/{id}/connect

Store an API key for a provider.

**Request:**
```json
{
  "api_key": "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

**Constraints:** `api_key` must be 3-4096 characters.

**Response 200:**
```json
{
  "connected": true,
  "provider": "openai",
  "health": {
    "ok": true
  },
  "maskedKey": "sk-...6789"
}
```

**Response 400:**
```json
{
  "detail": "invalid key length"
}
```

---

### DELETE /v1/providers/{id}

Remove a provider's API key.

**Response 200:**
```json
{
  "connected": false,
  "provider": "openai"
}
```

---

### GET /v1/providers/{id}/health

Check connectivity for a specific provider.

**Response 200 (connected):**
```json
{
  "ok": true
}
```

**Response 200 (not connected):**
```json
{
  "ok": false,
  "error": "not connected"
}
```

**Response 404:**
```json
{
  "detail": "provider not found"
}
```

---

### POST /v1/chat

Send a chat completion request to an AI provider.

**Request:**
```json
{
  "provider": "openai",
  "model": "gpt-5.6-terra",
  "messages": [
    { "role": "user", "content": "Hello" }
  ],
  "system": "You are a helpful assistant."
}
```

**Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `provider` | string | yes | Provider ID |
| `model` | string | no | Model override (uses default if omitted) |
| `messages` | array | yes | Message history `[{role, content}]` |
| `system` | string | no | System prompt prepended to messages |

**Response 200:**
```json
{
  "text": "Hello! How can I help you today?",
  "usage": {
    "prompt_tokens": 12,
    "completion_tokens": 9,
    "total_tokens": 21
  },
  "provider": "openai",
  "model": "gpt-5.6-terra",
  "latencyMs": 1234,
  "rawId": "chatcmpl-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

**Response 400:**
```json
{
  "detail": "provider not configured"
}
```

---

### GET /v1/ipfs/status

Check IPFS node status.

**Response 200:**
```json
{
  "online": false
}
```

---

### POST /v1/ipfs/add

Upload a file to IPFS via multipart form.

**Request:** `multipart/form-data` with a single file field.

**Response 200:**
```json
{
  "cid": "Qm...",
  "size": 12345,
  "name": "filename.ext"
}
```

---

### POST /v1/ipfs/pin/{cid}

Pin a CID to the local IPFS node.

**Response 200:**
```json
{
  "pinned": true,
  "cid": "Qm..."
}
```

---

### GET /v1/mesh/status

Get the P2P mesh node status.

**Response 200:**
```json
{
  "online": true,
  "nodeId": "624c7cb66165fe9efed263db819be895",
  "nodeName": "Racore Desktop",
  "peers": 0,
  "roles": ["client", "cache"],
  "uptimeSeconds": 42,
  "transport": "udp-multicast+rmp/0.1"
}
```

---

### GET /v1/mesh/peers

List connected mesh peers.

**Response 200:**
```json
[]
```

Each peer object:
```json
{
  "nodeId": "abc...",
  "nodeName": "Other Node",
  "address": "192.168.1.5:47777",
  "lastSeen": "2026-07-21T13:00:00Z"
}
```

---

### POST /v1/mesh/broadcast

Broadcast a message to all mesh peers.

**Request:**
```json
{
  "type": "custom.event",
  "data": {
    "key": "value"
  }
}
```

**Response 200:**
```json
{
  "sent": true,
  "type": "custom.event"
}
```

---

### GET /v1/authority/domains

List all claimed domains.

**Response 200:**
```json
[]
```

Each domain record:
```json
{
  "domain": "example.com",
  "controller": "did:key:z6Mk...",
  "createdAt": "2026-07-21T12:00:00Z",
  "status": "active"
}
```

---

### POST /v1/authority/domains

Claim a new domain.

**Request:**
```json
{
  "domain": "example.com"
}
```

**Response 200:**
```json
{
  "domain": "example.com",
  "controller": "did:key:z6Mk...",
  "createdAt": "2026-07-21T12:00:00Z",
  "status": "active",
  "dnsRecord": {
    "type": "TXT",
    "name": "_racore-key.example.com",
    "value": "did:key:z6Mk..."
  }
}
```

---

### GET /v1/authority/domains/{domain}/available

Check domain availability on the mesh.

**Response 200:**
```json
{
  "domain": "example.com",
  "available": true
}
```

---

### GET /v1/authority/domains/{domain}/releases

List all releases for a domain.

**Response 200:**
```json
[]
```

---

### POST /v1/authority/domains/{domain}/releases

Publish a new release for a domain.

**Request:**
```json
{
  "version": "1.0.0",
  "cid": "Qm...",
  "contentRoot": "/",
  "entrypoint": "index.html",
  "protocol": "web"
}
```

**Response 200:**
```json
{
  "published": true,
  "domain": "example.com",
  "version": "1.0.0",
  "cid": "Qm..."
}
```

---

### POST /v1/authority/domains/{domain}/delegate

Create a delegation grant for a domain.

**Request:**
```json
{
  "publicKey": "did:key:z6Mk...",
  "capabilities": ["publish", "transfer"],
  "expiresAt": "2027-01-01T00:00:00Z"
}
```

**Response 200:**
```json
{
  "delegated": true,
  "domain": "example.com",
  "grant": {
    "publicKey": "did:key:z6Mk...",
    "capabilities": ["publish", "transfer"],
    "expiresAt": "2027-01-01T00:00:00Z",
    "createdAt": "2026-07-21T12:00:00Z"
  }
}
```

---

### WebSocket /v1/events

Real-time event stream. Connect via:

```
ws://127.0.0.1:47831/v1/events
```

**Event format:**
```json
{
  "type": "event.type",
  "timestamp": "2026-07-21T13:00:00Z",
  "data": { ... }
}
```

**Event types:**
| Type | Description |
|------|-------------|
| `agent.started` | AI agent began processing |
| `agent.completed` | AI agent finished processing |
| `provider.connected` | Provider API key was configured |
| `mesh.peer.joined` | New mesh peer discovered |
| `mesh.peer.left` | Mesh peer disconnected |
| `mesh.message` | Message received from a mesh peer |
| `domain.claim` | Domain claim observed on mesh |
| `release.available` | New release published on mesh |
