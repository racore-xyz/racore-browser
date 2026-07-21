# Racore Browser

Agentic browser & open web protocol.

Browse with AI agents. Publish verifiable, portable websites on IPFS. Own your identity and data with self-sovereign domain authority and peer-to-peer mesh networking.

## Architecture

```
┌──────────────────────────────────────────────┐
│  Electron Desktop (optional)                  │
│  ┌────────────────────────────────────────┐   │
│  │  Frontend (Next.js / Vinext / React)   │   │
│  │  app/lib/racore-client.ts  ──── HTTP ───┼───┤
│  └────────────────────────────────────────┘   │
│                       │ IPC bridge (packaged)  │
└───────────────────────┼──────────────────────┘
                        │
┌───────────────────────▼──────────────────────┐
│  racored (Go daemon)  127.0.0.1:47831         │
│                                               │
│  ┌─────────┐ ┌──────────┐ ┌──────────────┐   │
│  │ Vault   │ │ Gateway  │ │ Mesh Network │   │
│  │ (AES)   │ │(AI APIs) │ │ (libp2p/mcast)│   │
│  ├─────────┤ ├──────────┤ ├──────────────┤   │
│  │ IPFS    │ │ Kubo     │ │ Authority    │   │
│  │ Bridge  │ │ Manager  │ │ (Domains)    │   │
│  └─────────┘ └──────────┘ └──────────────┘   │
└──────────────────────────────────────────────┘
```

## Prerequisites

- **Node.js** `>=22.13.0`
- **Go** `>=1.21`

## Quick Start

```bash
# Install frontend dependencies
npm install

# Build the Go daemon
cd god && go build -o build/racored ./cmd/racored/ && cd ..

# Start the daemon
./god/build/racored &

# Start the frontend (dev mode)
npm run dev
```

Open http://localhost:3000 in your browser.

## Project Structure

```
racore-browser/
├── app/                    # Next.js frontend (React)
│   ├── components/         # UI components
│   ├── lib/
│   │   └── racore-client.ts   # API client for Go daemon
│   └── page.tsx            # Home page
├── desktop/                # Electron shell
│   ├── main.cjs            # Electron main process
│   └── preload.cjs         # Context bridge
├── god/                    # Go backend (daemon + CLI)
│   ├── build/              # Compiled binaries
│   ├── cmd/
│   │   ├── racored/        # Daemon entrypoint
│   │   └── racore/         # CLI tool entrypoint
│   ├── internal/
│   │   ├── server/         # HTTP server + REST API
│   │   ├── vault/          # Encrypted key storage (AES-GCM)
│   │   ├── providers/      # AI provider catalog
│   │   ├── gateway/        # AI API routing
│   │   ├── mesh/           # P2P multicast mesh
│   │   ├── transport/      # UDP multicast transport
│   │   ├── ipfs/           # IPFS content bridge
│   │   ├── kubo/           # Embedded Kubo manager
│   │   ├── authority/      # Domain identity + delegation
│   │   ├── identity/       # DID key generation
│   │   ├── peer/           # Peer store
│   │   ├── protocol/       # Protocol message signing
│   │   └── config/         # Configuration loader
│   └── pkg/api/            # Shared API types
├── dist/                   # Frontend build output
├── worker/                 # Cloudflare Worker
├── tests/                  # Protocol tests
├── package.json
└── vite.config.ts
```

## Go Backend (god/)

The `god/` directory contains a complete rewrite of the original Python backend in Go.

### Building

```bash
# Build for current platform
cd god && go build -o build/racored ./cmd/racored/
go build -o build/racore ./cmd/racore/

# Cross-compile all targets
./scripts/build.sh
```

### Running

```bash
# Start the daemon
RACORE_DATA_DIR=/path/to/data ./god/build/racored

# Use the CLI
./god/build/racore --help
```

### Testing

```bash
cd god
go test -race -count=1 ./internal/...
go vet ./...
```

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | System health, version, status |
| GET | `/v1/providers` | List AI providers |
| PUT | `/v1/providers/{id}/connect` | Store API key for provider |
| DELETE | `/v1/providers/{id}` | Remove provider connection |
| GET | `/v1/providers/{id}/health` | Check provider health |
| POST | `/v1/chat` | Send chat completion request |
| GET | `/v1/ipfs/status` | IPFS node status |
| POST | `/v1/ipfs/add` | Upload file to IPFS |
| POST | `/v1/ipfs/pin/{cid}` | Pin a CID |
| GET | `/v1/mesh/status` | Mesh node status |
| GET | `/v1/mesh/peers` | Mesh peer list |
| POST | `/v1/mesh/broadcast` | Broadcast message over mesh |
| GET | `/v1/authority/domains` | List claimed domains |
| POST | `/v1/authority/domains` | Claim a domain |
| GET | `/v1/authority/domains/{domain}/available` | Check domain availability |
| GET | `/v1/authority/domains/{domain}/releases` | List domain releases |
| POST | `/v1/authority/domains/{domain}/releases` | Publish release |
| POST | `/v1/authority/domains/{domain}/delegate` | Create delegation |
| WS | `/v1/events` | Real-time event stream |

### AI Providers

9 built-in providers: OpenAI, Anthropic, Google Gemini, OpenRouter, Kimi/Moonshot, Ollama (local), OpenCode (local), Claude Code (local), Kimi Code (local).

## Desktop App

```bash
# Development mode (frontend + Electron)
npm run desktop:dev

# Package for distribution
npm run desktop:package
```

The Electron app (`desktop/main.cjs`) starts `god/build/racored` as a child process, waits for it to become ready on port 47831, then loads the frontend UI.

## Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Start frontend dev server |
| `npm run build` | Build frontend for production |
| `npm start` | Start production frontend server |
| `npm run desktop` | Launch Electron app |
| `npm run desktop:dev` | Dev mode with live reload |
| `npm run desktop:package` | Package Electron distributable |
| `npm test` | Build + run protocol tests |
| `cd god && go test ./internal/...` | Run Go backend tests |

## License

Proprietary.
