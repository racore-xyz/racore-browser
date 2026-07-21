---
name: racore-project
description: Full project context for Racore Browser -- a three-tier desktop app (Go daemon + Next.js frontend + Electron shell). Guides agents to the right code paths and documentation.
---

# Racore Browser Project

## Overview

Three-tier desktop application:

| Layer | Tech | Entry Point |
|-------|------|-------------|
| Frontend | Next.js 16 / Vinext / React / TypeScript / Tailwind | `app/` |
| Backend | Go daemon (`racored`) | `god/` |
| Desktop | Electron 37 | `desktop/main.cjs` |

The Go daemon serves a REST API on `127.0.0.1:47831`. The frontend communicates with it via `app/lib/racore-client.ts` (direct HTTP or Electron IPC bridge).

## Documentation Files

Consult these before modifying any part of the project:

| Area | File |
|------|------|
| System architecture and component relationships | `.llms/01-architecture.md` |
| Frontend component tree, API client, types | `.llms/02-frontend.md` |
| Go backend packages, module structure, env vars | `.llms/03-backend.md` |
| Electron shell, IPC handlers, packaging | `.llms/04-desktop.md` |
| Complete REST API with request/response examples | `.llms/05-api-reference.md` |
| Data flow for chat, mesh, domains, IPFS, errors | `.llms/06-data-flow.md` |
| All shared Go and TypeScript types | `.llms/07-types.md` |
| Build commands, test patterns, CI notes | `.llms/08-build-and-test.md` |

## Development Commands

```bash
# Build Go daemon
cd god && go build -o build/racored ./cmd/racored/

# Start daemon
RACORE_DATA_DIR=/tmp/racored-dev ./god/build/racored

# Start frontend
npm run dev

# Run Go tests
cd god && go test -race -count=1 ./internal/...

# Run all project tests
npm test
```

## Coding Conventions

- **Error format**: Always `{"detail": "error message"}`, never `{"error": "..."}` (the server's `writeError()` function handles this)
- **API client**: All frontend-to-daemon calls go through `app/lib/racore-client.ts` `daemonRequest()` -- never use raw `fetch()` directly
- **Config**: Environment variables prefixed `RACORE_`, loaded in `god/internal/config/`
- **JSON field naming**: camelCase throughout (Go JSON tags and TypeScript interfaces agree)
- **Go imports**: Standard lib first, then third-party (`golang.org/x/...`, `github.com/gorilla/...`)
- **Daemon shutdown**: `transport.Close()` must call `syscall.Shutdown(SHUT_RD)` before `pc.Close()` to unblock read goroutines

## Safety Checks

1. Before adding an API endpoint: Register in `server.go`, implement in `handlers.go`, add to `.llms/05-api-reference.md`
2. Before changing a Go type: Check `app/lib/racore-client.ts` for matching TypeScript type
3. Before changing an error response: Verify the key is `"detail"` (not `"error"`)
4. Before modifying CORS: Add the new origin to `corsMiddleware()` in `server.go`
5. Before editing transport code: Ensure `Close()` still unblocks pending reads via shutdown
6. Before changing the Electron startup: Test both `npm run desktop` and `npm run desktop:dev`
