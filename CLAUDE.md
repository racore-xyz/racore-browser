# Racore Browser Project Guide

This file provides context for AI agents working on the Racore Browser project. Read this first to understand the project structure and conventions.

## Quick Summary

Racore Browser is a desktop application with three layers:
1. **Frontend**: React/Next.js SPA built with Vinext (Vite-based RSC framework)
2. **Backend**: Go daemon (`racored`) serving REST/WebSocket API on port 47831
3. **Desktop**: Electron shell that spawns the Go binary and loads the UI

## Directory Structure

```
app/                  Next.js frontend (React, TypeScript, Tailwind)
god/                  Go backend daemon + CLI
desktop/              Electron shell (main.cjs, preload.cjs)
dist/                 Frontend production build
worker/               Cloudflare Worker
tests/                Protocol tests
.llms/                Detailed project documentation (8 files)
```

## Detailed Documentation by Section

For comprehensive details on each part of the project, refer to these files:

| Topic | File |
|-------|------|
| System architecture | `.llms/01-architecture.md` |
| Frontend (React/Next.js) | `.llms/02-frontend.md` |
| Go backend (god/) | `.llms/03-backend.md` |
| Electron desktop | `.llms/04-desktop.md` |
| REST API reference | `.llms/05-api-reference.md` |
| Data flow diagrams | `.llms/06-data-flow.md` |
| Shared types (Go + TS) | `.llms/07-types.md` |
| Build and test guide | `.llms/08-build-and-test.md` |

## Key Conventions

- **Go backend**: Module `github.com/racore/god`, packages under `god/internal/`, API types in `god/pkg/api/`
- **Error responses**: Always `{"detail": "message"}`, never `{"error": "message"}`
- **API client**: Single entry point `app/lib/racore-client.ts` -- all frontend-to-daemon communication goes through `daemonRequest()`
- **Config**: Environment variables prefixed `RACORE_` (see `.llms/03-backend.md` for full list)
- **Testing**: Go tests use `-race -count=1`; frontend tests use `node --test`
- **CORS**: Whitelist-based in `god/internal/server/server.go`, origins include `localhost:3000` and `127.0.0.1:47832`

## Common Tasks

### Build and run everything
```bash
cd god && go build -o build/racored ./cmd/racored/ && cd ..
RACORE_DATA_DIR=/tmp/racored-dev ./god/build/racored &
npm run dev
```

### Run Go tests
```bash
cd god && go test -race -count=1 ./internal/...
```

### Build frontend
```bash
npm run build
```

### Package Electron app
```bash
npm run desktop:package
```

## Important Implementation Details

1. The Go transport layer (`god/internal/transport/`) uses `syscall.Shutdown(SHUT_RD)` in `Close()` to unblock pending reads -- do not change this pattern
2. Mesh tests (`TestTwoNodeDiscovery`, `TestBroadcastReceive`) skip when multicast is unavailable; they use port 0 on failure
3. The frontend has two API transport modes: direct HTTP fetch (browser dev) and Electron IPC bridge (desktop packaged) -- both converge in `daemonRequest()`
4. API keys are encrypted at rest using AES-256-GCM with a PBKDF2-derived key in `god/internal/vault/`
5. The CORS middleware list is fixed -- any new dev origins must be added to `corsMiddleware()` in `god/internal/server/server.go`

## Questions to Ask Before Making Changes

1. Does this change affect the Go-Frontend API contract? If so, update BOTH `god/pkg/api/types.go` and `app/lib/racore-client.ts`
2. Does this change affect error responses? Ensure `{"detail": "..."}` format is used
3. Does this change add a new API endpoint? Register it in `god/internal/server/server.go`, add handler in `handlers.go`, update `.llms/05-api-reference.md`
4. Does this change affect the Electron shell? Test both `npm run desktop` and `npm run desktop:dev`
