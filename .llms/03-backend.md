# Go Backend (god/)

## Overview

The `god/` directory contains a complete Go rewrite of the original Python backend. It produces two static binaries with zero runtime dependencies.

| Binary | Module | Description |
|--------|--------|-------------|
| `build/racored` | `cmd/racored/main.go` | Main daemon: HTTP API, WebSocket, P2P mesh, IPFS bridge, Kubo manager |
| `build/racore` | `cmd/racore/main.go` | CLI tool: status queries, domain management, content publishing |

## Directory Structure

```
god/
  go.mod                    Module: github.com/racore/god
  go.sum
  scripts/
    build.sh                Cross-compilation script (5 platforms)
  cmd/
    racored/
      main.go               Daemon entrypoint
    racore/
      main.go               CLI entrypoint
  internal/
    server/                 HTTP server, routes, handlers, WebSocket hub, CORS middleware
    vault/                  Encrypted key-value store (AES-GCM, PBKDF2-derived key)
    providers/              AI provider catalog (9 providers), health checks
    gateway/                AI API routing and request forwarding
    mesh/                   P2P mesh node: heartbeat, broadcast, peer discovery, event sink
    transport/              UDP multicast transport with IPv4 control messages
    ipfs/                   IPFS bridge: add, pin, gateway URL resolution
    kubo/                   Embedded Kubo (IPFS) subprocess manager
    authority/              Domain identity: DID keys, DNS records, delegation, release management
    identity/               Ed25519 key generation, DID key encoding
    peer/                   Peer store (in-memory, thread-safe)
    protocol/               Message envelope, signing, serialization
    config/                 Configuration loader (environment + JSON file)
  pkg/api/
    types.go                Shared API types (Config, Provider, ChatRequest, etc.)
    gomesh_types.go         Mesh-specific types (envelope, heartbeat, peer info)
```

## Module Details

### Server (`internal/server/`)
- **File**: `server.go`, `handlers.go`, `hub.go`
- **Port**: 47831 (configurable via `RACORE_PORT`)
- **Router**: `http.ServeMux` with 19 registered routes
- **CORS**: Whitelist-based middleware allowing `localhost:3000`, `127.0.0.1:3000`, `127.0.0.1:47832`, `racore.xyz`
- **WebSocket**: `/v1/events` endpoint for real-time event broadcasting to connected clients
- **Error format**: `{"detail": "<message>"}` (consistent across all error responses)

### Vault (`internal/vault/`)
- **Purpose**: Secure storage for AI provider API keys
- **Encryption**: AES-256-GCM with a key derived from a machine-local seed via PBKDF2
- **Storage**: JSON file on disk (`vault.json` in data directory)
- **Masking**: Keys displayed with only last 4 characters visible (e.g., `sk-...6789`)

### Providers (`internal/providers/`)
- **Catalog**: 9 built-in providers:
  - Cloud APIs: openai, anthropic, gemini, openrouter, kimi
  - Local/CLI: ollama, opencode, claude-code, kimi-code
- **Health checks**: Each provider exposes a `/health`-style endpoint check
- **Authentication**: Bearer token (OpenAI-style) or x-api-key header (Anthropic-style)

### Gateway (`internal/gateway/`)
- **Purpose**: Routes AI chat completion requests to the appropriate provider
- **Format conversion**: Translates between internal `ChatRequest` format and each provider's native API format
- **Error handling**: Returns structured error responses on provider failure

### Mesh (`internal/mesh/`)
- **Transport**: UDP multicast on configurable group (default `239.255.77.77`) and port (default `47777`)
- **Message signing**: Ed25519 signatures on all mesh messages
- **Heartbeat**: Periodic broadcast announcing node presence (configurable interval)
- **Peer discovery**: Passive via heartbeat messages; peers tracked in thread-safe `peer.Store`
- **Event sink**: Plug-in callback for processing incoming mesh events (domain claims, releases)

### Transport (`internal/transport/`)
- **Protocol**: UDP multicast with IPv4 control message support
- **Buffer pool**: `sync.Pool` for zero-allocation reads
- **Lifecycle**: `Start()` -> `StartReadLoop()` -> `Close()` with proper cleanup via `syscall.Shutdown`

### IPFS Bridge (`internal/ipfs/`)
- **Purpose**: Content-addressed storage interface
- **Methods**: `AddBytes(name, data)`, `Pin(cid)`, `GatewayURL(cid)`
- **Connection**: HTTP client to local Kubo RPC API

### Kubo Manager (`internal/kubo/`)
- **Purpose**: Manages the embedded Kubo IPFS node as a child process
- **Lifecycle**: `Start()` spawns the process, `Stop()` sends SIGTERM
- **Binary path**: Configured via `RACORE_KUBO_PATH` env var

### Authority (`internal/authority/`)
- **Purpose**: Domain identity and delegation management
- **DID keys**: Ed25519 keypair encoded as `did:key:<base64url>`
- **DNS records**: TXT record format for DNS-based verification
- **Delegation**: Grant capabilities to other keys with expiration
- **Releases**: Versioned content manifests pointing to IPFS CIDs

### Identity (`internal/identity/`)
- **Purpose**: Cryptographic key generation
- **Key type**: Ed25519 (via `golang.org/x/crypto/ed25519`)
- **Encoding**: `did:key` format using base64url (no padding)

### Protocol (`internal/protocol/`)
- **Purpose**: Mesh message envelope and signing
- **Envelope fields**: Sender, recipient, type, payload, timestamp, signature
- **Serialization**: JSON with deterministic key ordering for signature verification

### Peer (`internal/peer/`)
- **Purpose**: In-memory thread-safe peer store
- **Interface**: `Add()`, `Remove()`, `Get()`, `List()`, `Count()`

## Key Go Dependencies

```
module github.com/racore/god

go 1.21

require (
    github.com/gorilla/websocket v1.5.3   // WebSocket support
    golang.org/x/net v0.30.0              // IPv4 multicast control messages
    golang.org/x/crypto v0.28.0           // Ed25519, PBKDF2
    golang.org/x/sys v0.26.0              // Unix syscalls (SO_REUSEADDR, shutdown)
)
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `RACORE_DATA_DIR` | `~/.local/share/racore` | Data directory for vault, config, IPFS |
| `RACORE_PORT` | `47831` | HTTP API server port |
| `RACORE_HOST` | `127.0.0.1` | HTTP API bind address |
| `RACORE_NODE_NAME` | `Racore Desktop` | Display name on the mesh |
| `RACORE_MESH_ENABLED` | `true` | Enable P2P mesh networking |
| `RACORE_MESH_GROUP` | `239.255.77.77` | Multicast group address |
| `RACORE_MESH_PORT` | `47777` | Multicast port |
| `RACORE_MESH_HEARTBEAT` | `30` | Heartbeat interval (seconds) |
| `RACORE_KUBO_PATH` | (auto-detect) | Path to Kubo/IPFS binary |
| `RACORE_IPFS_API` | `/ip4/127.0.0.1/tcp/5001` | Kubo RPC API multiaddr |
| `RACORE_IPFS_GATEWAY` | `http://127.0.0.1:8080` | IPFS gateway URL |

## Building

```bash
# Current platform only
cd god
go build -o build/racored ./cmd/racored/
go build -o build/racore ./cmd/racore/

# All platforms (uses scripts/build.sh)
./scripts/build.sh

# Cross-compilation targets
# - linux/amd64, linux/arm64
# - darwin/amd64, darwin/arm64
# - windows/amd64
```

## Testing

```bash
cd god
go test -race -count=1 ./internal/...
go vet ./...
```

12 internal packages, all with race detection enabled.
