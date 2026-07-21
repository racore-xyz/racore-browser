# Architecture

## Overview

Racore Browser is a three-tier application:

1. **Frontend** -- React/Next.js SPA rendered via Vinext (Vite-based React Server Components framework)
2. **Daemon** -- Go binary (`racored`) that serves a REST/WebSocket API on `127.0.0.1:47831`
3. **Desktop Shell** -- Electron wrapper that spawns the Go daemon and loads the frontend

All components run locally on the user's machine. There is no cloud backend. The daemon handles AI provider API calls, IPFS content storage, P2P mesh networking, and domain authority.

## Component Diagram

```
+------------------------------------------------------------------+
|  Electron Desktop (optional)                                      |
|  +------------------------------------------------------------+  |
|  |  Frontend (Next.js / Vinext / React SPA)                   |  |
|  |  app/lib/racore-client.ts  -- HTTP --> daemon (port 47831)|  |
|  |  app/page.tsx             main entry point                |  |
|  |  app/components/          UI components                   |  |
|  +------------------------------------------------------------+  |
|                    | IPC bridge (when packaged)                  |
+--------------------+---------------------------------------------+
                     |
                     | HTTP (localhost:47831)
                     |
+--------------------v---------------------------------------------+
|  racored (Go daemon)  -- single static binary, no runtime deps    |
|                                                                    |
|  +-----------+  +--------------+  +----------------------------+  |
|  | Server    |  | Gateway      |  | Mesh Network               |  |
|  | (HTTP API)|  | (AI routing) |  | (UDP multicast, Ed25519)   |  |
|  +-----------+  +--------------+  +----------------------------+  |
|  +-----------+  +--------------+  +----------------------------+  |
|  | Vault     |  | IPFS Bridge  |  | Authority (Domains)        |  |
|  | (AES-GCM) |  | (content)    |  | (DID + delegation)         |  |
|  +-----------+  +--------------+  +----------------------------+  |
+------------------------------------------------------------------+
```

## Process Model

- **Development**: Frontend runs via `npm run dev` (Vite dev server on port 3000); daemon runs as a separate process started manually or via the Electron dev mode (`npm run desktop:dev`)
- **Production**: Electron app spawns `god/build/racored` as a child process, waits for `/health` to return 200, then loads the pre-built UI from `dist/`
- **Packaged**: Both the daemon binary and the built UI are bundled inside the Electron distributable via `electron-builder`

## Communication Patterns

| Pattern | Transport | Direction |
|---------|-----------|-----------|
| REST API | HTTP (`127.0.0.1:47831`) | Frontend -> Daemon |
| Real-time events | WebSocket (`/v1/events`) | Daemon -> Frontend |
| Desktop IPC | Electron `ipcMain.handle` | Frontend -> Electron -> Daemon |
| AI provider calls | HTTPS (to external APIs) | Daemon -> OpenAI/Anthropic/etc. |
| P2P mesh | UDP multicast (`239.255.77.77:47777`) | Daemon <-> Daemon |
| IPFS | HTTP (to local Kubo daemon) | Daemon -> Kubo (subprocess) |

## Key Design Decisions

- **Single Go binary for backend**: Zero runtime dependencies. No Python, no virtualenv, no PyInstaller. Cross-compiled for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.
- **Local-first**: All data stays on the user's machine. API keys are encrypted at rest (AES-GCM). No user accounts, no cloud sync.
- **P2P mesh for discovery**: Nodes discover each other via UDP multicast on the LAN. Messages are signed with Ed25519 keys. No central server required.
- **Standard library HTTP**: The Go server uses `net/http` with `http.ServeMux` -- no third-party router. CORS is handled by a middleware wrapper.
- **Shared API types**: Go structs in `pkg/api/types.go` define all request/response shapes. The frontend mirrors these in TypeScript interfaces in `app/lib/racore-client.ts`.
