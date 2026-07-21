# Build and Test

---

## Prerequisites

- Node.js >= 22.13.0
- Go >= 1.21
- npm >= 10

## Development Setup

```bash
# Clone and install frontend dependencies
git clone <repo>
cd racore-browser
npm install

# Build the Go daemon for current platform
cd god && go build -o build/racored ./cmd/racored/ && cd ..
go build -o build/racore ./cmd/racore/  # optional CLI
```

## Running in Development Mode

### Option 1: Browser only (no desktop)

Terminal 1 -- Start the Go daemon:
```bash
RACORE_DATA_DIR=/tmp/racored-dev ./god/build/racored
```

Terminal 2 -- Start the frontend dev server:
```bash
npm run dev
```

Open http://localhost:3000 in a browser. The frontend communicates with the daemon via direct HTTP fetch to `http://127.0.0.1:47831`.

### Option 2: Electron desktop (dev mode)

```bash
npm run desktop:dev
```

This runs both the frontend dev server and the Electron app concurrently. The Electron window loads from `http://localhost:3000` and proxies API requests through the IPC bridge.

## Frontend Build

### Production Build

```bash
npm run build
```

Output: `dist/` (client, server, RSC assets)

The build uses Vinext (Vite 8) and runs 5 steps:
1. Analyze client references
2. Analyze server references
3. Build RSC (React Server Components) environment
4. Build client environment
5. Build SSR (Server-Side Rendering) environment

### Production Server

```bash
npm start
```

Serves the built frontend on port 3000 using the Vinext SSR runtime.

## Go Backend Build

### Current Platform

```bash
cd god
go build -o build/racored ./cmd/racored/
go build -o build/racore ./cmd/racore/
```

### All Platforms (cross-compile)

```bash
cd god && ./scripts/build.sh
```

This produces binaries for:
- `racored-linux-amd64`, `racored-linux-arm64`
- `racored-darwin-amd64`, `racored-darwin-arm64`
- `racored-windows-amd64.exe`

And the same set for the `racore` CLI binary.

Build flags: `-ldflags="-s -w"` (strips debug info, reduces binary size by ~40%).

## Go Backend Testing

```bash
cd god

# Run all internal package tests with race detection
go test -race -count=1 ./internal/...

# Run tests for a specific package
go test -race -count=1 ./internal/server/...
go test -race -count=1 ./internal/mesh/...

# Run vet (static analysis)
go vet ./...

# Run tests with coverage
go test -race -cover -count=1 ./internal/...

# Run a specific test function
go test -race -run TestTwoNodeDiscovery ./internal/mesh/
```

### 12 Test Packages

All 12 internal packages have test coverage:

| Package | Files | Focus |
|---------|-------|-------|
| `authority` | Domain identity, claims, DNS records |
| `config` | Configuration loading |
| `identity` | Ed25519 key generation |
| `ipfs` | IPFS bridge operations |
| `kubo` | Kubo subprocess manager |
| `mesh` | P2P node, heartbeat, broadcast, discovery |
| `peer` | Thread-safe peer store |
| `protocol` | Envelope signing and verification |
| `providers` | Provider catalog and health checks |
| `server` | HTTP routes, handlers, CORS, WebSocket |
| `transport` | UDP multicast transport lifecycle |
| `vault` | AES-GCM encryption and key masking |

## Electron Packaging

```bash
# Package for Windows (NSIS installer)
npm run desktop:package

# This runs:
#   1. npm run build (frontend)
#   2. electron-builder --win nsis
```

The packaged output goes to `desktop-dist/`.

### electron-builder Configuration

Source: `package.json` `"build"` key

```
Output:     desktop-dist/
App ID:     xyz.racore.browser
Files:      desktop/**/*, package.json
Extra resources:
  - dist/        -> ui/          (built frontend)
  - god/build/   -> racored/     (Go binaries)
  - desktop/runtime/kubo/ -> kubo/ (IPFS binary)
Target:     NSIS (Windows)
```

## Protocol Tests

```bash
npm run protocol:test
```

Runs node-based tests from `tests/protocol.test.mjs`.

```bash
npm test
```

Runs the full build + protocol test suite.

## CI Considerations

- The mesh integration tests (`TestTwoNodeDiscovery`, `TestBroadcastReceive`) require multicast networking and will skip when unavailable (e.g., CI environments without multicast support)
- The transport tests use port 0 (OS-assigned) and clean up via `syscall.Shutdown` + `Close()`
- Race detection (`-race`) is essential for the concurrent mesh and transport code
- All Go tests use `-count=1` to disable test caching

## Linting

### Frontend

```bash
npm run lint
```

Uses ESLint 9 with `eslint-config-next`.

### Go

```bash
cd god && go vet ./...
```

No external linter is configured for the Go code.
