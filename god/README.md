# god - Racore Go Daemon

Complete rewrite of the Racore Python daemon in Go. Single static binary, no runtime dependencies.

## Binaries

| Binary | Description |
|--------|-------------|
| `racored` | Main daemon: HTTP API, WebSocket, P2P mesh, IPFS, domain authority, AI provider gateway |
| `racore` | CLI tool: status, domains, claim, publish, releases |

## Architecture

```
racored (single binary)
  ├── HTTP API :47831 (REST + WebSocket)
  ├── P2P Mesh :47777 (UDP multicast + unicast, Ed25519 signed envelopes)
  │   ├── Presence heartbeat & peer discovery
  │   ├── Bootstrap peers for cross-LAN connectivity
  │   ├── Ping/pong latency measurement
  │   ├── Graceful leave (goodbye messages)
  │   └── Unicast direct messaging (SendTo)
  ├── Kubo IPFS manager (subprocess)
  └── AI provider gateway (OpenAI, Anthropic, Gemini, Ollama, etc.)
```

## Build

```bash
go build -o build/racored ./cmd/racored/
go build -o build/racore ./cmd/racore/
./scripts/build.sh  # cross-compile all platforms
```

## Quick Start

```bash
# Start the daemon
./build/racored

# Check status (in another terminal)
./build/racore status

# Claim a domain
./build/racore claim example.com

# Publish a web build
./build/racore publish --domain example.com --build ./dist --version 1.0.0
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Full system health |
| GET | `/v1/providers` | AI provider catalog |
| PUT | `/v1/providers/{id}/connect` | Store API key |
| DELETE | `/v1/providers/{id}` | Remove API key |
| GET | `/v1/providers/{id}/health` | Provider connectivity |
| POST | `/v1/chat` | Chat completion |
| GET | `/v1/ipfs/status` | IPFS node status |
| POST | `/v1/ipfs/add` | Upload file to IPFS |
| POST | `/v1/ipfs/pin/{cid}` | Pin CID |
| GET | `/v1/mesh/status` | P2P mesh status |
| GET | `/v1/mesh/peers` | Peer list |
| POST | `/v1/mesh/broadcast` | Broadcast message |
| GET | `/v1/authority/domains` | List domains |
| POST | `/v1/authority/domains` | Claim domain |
| GET | `/v1/authority/domains/{d}/available` | Check availability |
| GET | `/v1/authority/domains/{d}/releases` | List releases |
| POST | `/v1/authority/domains/{d}/releases` | Publish release |
| POST | `/v1/authority/domains/{d}/delegate` | Delegate authority |
| WS | `/v1/events` | Real-time event stream |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `RACORE_DATA_DIR` | `~/.local/share/racore` | Data directory |
| `RACORE_PORT` | `47831` | HTTP API port |
| `RACORE_NODE_NAME` | `Racore Desktop` | Mesh node name visible to peers |
| `RACORE_MESH_PORT` | `47777` | Mesh UDP port |
| `RACORE_MESH_GROUP` | `239.255.77.77` | Multicast group address |
| `RACORE_MESH_HEARTBEAT_SEC` | `5` | Heartbeat interval in seconds |
| `RACORE_BOOTSTRAP_PEERS` | - | Comma-separated `host:port` list for cross-LAN discovery |
| `RACORE_USE_GOD` | - | Legacy compatibility flag; the Tauri package always bundles the Go daemon |

> Mesh is configured via `settings.json` in the data directory. Environment variables override file values.

## Migrating from Python

Single binary. Zero Python dependency. No PyInstaller. No virtualenv.
