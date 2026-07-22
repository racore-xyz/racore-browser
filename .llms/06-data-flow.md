# Data Flow

This document describes the key data flows through the system, from user action to persistence and back.

---

## AI Chat Flow

```
User types message in AgenticBrowserView
       |
       v
AgenticBrowserView.tsx
  - Calls daemonRequest("/v1/chat", { method: "POST", body: chatRequest })
       |
       v
racore-client.ts
  - Desktop path: desktopBridge.api({ path, method, body })
  - Browser path: fetch("http://127.0.0.1:47831/v1/chat", { method, body })
       |
       v
Tauri Rust daemon_request command (if desktop)
  - Validates the method/path against the Racore API allowlist
  - Forwards to the fixed loopback daemon origin
       |
       v
server.go (racored, port 47831)
  - corsMiddleware checks Origin
  - routes to chatHandler
       |
       v
handlers.go chatHandler()
  - Parses ChatRequest from body
  - Broadcasts { type: "agent.started" } via Hub (WebSocket)
  - Calls s.gateway.Chat(ctx, providerID, req)
       |
       v
gateway/gateway.go
  - Looks up provider config (baseURL, API key from Vault)
  - Routes to appropriate provider adapter
  - Calls provider API via HTTPS
       |
       v
External Provider API (OpenAI, Anthropic, etc.)
  - Returns completion response
       |
       v
gateway/gateway.go
  - Parses provider-specific response
  - Converts to unified ChatResponse format
       |
       v
handlers.go chatHandler()
  - Broadcasts { type: "agent.completed", latencyMs } via Hub
  - Writes ChatResponse as JSON response
       |
       v
racore-client.ts
  - Parses JSON response
  - Returns typed data to component
       |
       v
AgenticBrowserView.tsx
  - Displays response text to user
```

## Provider Connection Flow

```
User enters API key in ProvidersView
       |
       v
ProvidersView.tsx
  - Calls connectProvider(providerId, apiKey)
       |
       v
racore-client.ts
  - PUT /v1/providers/{id}/connect { api_key: "sk-..." }
       |
       v
handlers.go providerByIDHandler (PUT /connect)
  - Validates key length (3-4096 chars)
  - Stores key: s.vault.Set(providerId, apiKey)
  - Checks health: s.gateway.Health(ctx, providerId)
  - Broadcasts { type: "provider.connected" } via Hub
  - Returns { connected: true, health, maskedKey }
       |
       v
vault/vault.go
  - Encrypts key with AES-256-GCM (key derived via PBKDF2 from machine seed)
  - Writes encrypted data to vault.json on disk
  - Returns masked version (sk-...6789)
       |
       v
UI shows "Connected" status with masked key
```

## Mesh Network Flow

```
racored starts
       |
       v
server.go Start()
  - Creates MeshNode with config
  - Calls mesh.Start(ctx)
       |
       v
mesh/mesh.go Start()
  - Creates UDPTransport: NewUDPTransport(group, port)
  - Calls transport.Start() (binds to multicast socket, joins group)
  - Starts readLoop goroutine (listens for incoming mesh messages)
  - Starts heartbeat goroutine (periodic broadcast)
       |
       v
Heartbeat (every N seconds):
  - Creates signed Heartbeat message (nodeId, nodeName, timestamp)
  - Encodes as envelope, signs with Ed25519 key
  - Calls transport.Broadcast(data)
  - UDP multicast to 239.255.77.77:47777
       |
       v
All peers on the same multicast group receive the heartbeat
       |
       v
transport/udp.go readLoop
  - Receives UDP datagram
  - Sends Message{Data, Addr} to recvChan
       |
       v
mesh/mesh.go handleMessage()
  - Parses envelope from raw bytes
  - Verifies Ed25519 signature
  - Routes by type:
    - "heartbeat" -> peerStore.Add(peer) -> emit("mesh.peer.joined")
    - "domain.claim" -> authority.ObserveClaim() -> emit("domain.claim")
    - other -> emit("mesh.message")
       |
       v
Event Sink (registered in server.go New())
  - domain.claim -> authority.ObserveClaim(domain, controller, nodeId)
  - All events -> hub.Broadcast(event) (WebSocket to UI)
```

## Domain Claim Flow

```
User enters domain in SitesView
       |
       v
SitesView.tsx
  - Calls daemonRequest("/v1/authority/domains", { method: "POST", body: { domain } })
       |
       v
handlers.go domainsHandler (POST)
  - Creates Ed25519 keypair via identity.Generate()
  - Constructs DomainRecord { domain, controller: did:key, status: "pending" }
  - Stores in authority store
  - Creates DNS txt record data for verification
  - Broadcasts { type: "domain.claim" } over mesh
  - Returns domain record with DNS instructions
       |
       v
User configures DNS TXT record at their domain registrar:
  _racore-key.example.com TXT "did:key:z6Mk..."
       |
       v
Mesh peers receive the domain.claim broadcast
  -> ObserveClaim(domain, controller, nodeId)
  -> Records the claim in memory for availability checks
```

## IPFS Content Flow

```
User publishes a release
       |
       v
handlers.go domainByIDHandler (POST /releases)
  - Creates ReleaseManifest { protocol, domain, version, cid, ... }
  - Stores release in authority store
  - Broadcasts { type: "release.available" } over mesh
  - Returns release confirmation
       |
       v
If uploading a file (POST /v1/ipfs/add):
  - Parses multipart form
  - Reads file bytes
  - Calls ipfs.Bridge.AddBytes(name, data)
  - Bridge calls Kubo RPC API (HTTP)
  - Kubo adds content to IPFS, returns CID
  - Broadcasts { type: "release.available", cid } over mesh
  - Returns { cid, size, name }
```

## Error Flow

```
Any API request
       |
       v
server.go corsMiddleware
  - If OPTIONS: return 204 (CORS preflight handled immediately)
  - If disallowed Origin: no Access-Control-Allow-Origin header set
  - Browser blocks the response -> CORS error in console
       |
       v
handler (e.g., connectProvider with empty key)
  - Validation fails
  - Calls writeError(w, 400, "invalid key length")
       |
       v
server.go writeError()
  - Writes {"detail": "invalid key length"}
  - Sets Content-Type: application/json
  - Sets HTTP 400 status
       |
       v
racore-client.ts
  - response.ok is false
  - Parses response JSON
  - Throws Error(error.detail || "Racore daemon returned 400")
       |
       v
Component catches error
  - Shows error message in UI (toast, inline error, etc.)
```

## Data Persistence

```
racored runtime directory ($RACORE_DATA_DIR)
  |
  |- vault.json           Encrypted API keys (AES-256-GCM)
  |   Format: {"openai": "<base64 ciphertext>"}
  |
  |- config.json          User configuration
  |   Format: {"nodeName": "...", "meshPort": 47777}
  |
  |- ipfs/                IPFS repo (if Kubo is running)
  |
  |- identity.key         Ed25519 private key for mesh identity
  |   Format: raw 64-byte seed

(In-memory, not persisted)
  - Peer store (mesh peers)
  - Authority domain records
  - WebSocket Hub connections
```
