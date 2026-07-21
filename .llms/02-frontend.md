# Frontend

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Framework | Next.js 16 (React Server Components) |
| Build tool | Vite 8 via Vinext 0.0.50 |
| Language | TypeScript 5.9 |
| Styling | Tailwind CSS 4 |
| Database ORM | Drizzle ORM 0.45 (optional, D1-compatible) |
| Package manager | npm |

## Directory Structure

```
app/
  page.tsx                    Main entry point (client component)
  layout.tsx                  Root layout (server component)
  lib/
    racore-client.ts          API client for Go daemon communication
  components/
    AgenticBrowserView.tsx    AI chat browser interface
    ProvidersView.tsx         AI provider management UI
    LiveNetworkView.tsx       Mesh/IPFS network status display
    SitesView.tsx             Domain authority management UI
    Onboarding.tsx            First-run setup flow
    system-controls.tsx       System controls (theme, settings)
  types/
    desktop.d.ts              TypeScript types for Electron bridge
```

## API Client (`app/lib/racore-client.ts`)

This is the sole communication layer between the frontend and the Go daemon. Every component imports from this module.

### Transport Modes

1. **Desktop (Electron)**: When `window.racoreDesktop?.api` exists, requests are proxied through the Electron IPC bridge:
   ```
   Component -> daemonRequest() -> window.racoreDesktop.api() -> ipcMain.handle('racore:api') -> HTTP to daemon
   ```

2. **Browser (dev)**: Direct HTTP fetch to `http://127.0.0.1:47831`:
   ```
   Component -> daemonRequest() -> fetch('http://127.0.0.1:47831/path')
   ```

### Shared Types

```typescript
interface ProviderInfo {
  id: string;               // e.g. "openai", "anthropic"
  name: string;             // Display name
  kind: string;             // "responses", "anthropic", "gemini", "openrouter"
  defaultModel: string;     // e.g. "gpt-5.6-terra"
  free: boolean;            // Free tier available
  local: boolean;           // Runs locally (Ollama, OpenCode, etc.)
  connected: boolean;       // API key configured
  maskedKey?: string | null; // Masked key for display
}
```

### Exported Functions

| Function | HTTP Call | Returns |
|----------|-----------|---------|
| `daemonRequest<T>(path, options)` | Generic request | `Promise<T>` |
| `checkDaemon()` | `GET /health` | `Record<string, unknown> \| null` |
| `listProviders()` | `GET /v1/providers` | `ProviderInfo[]` |
| `connectProvider(provider, apiKey)` | `PUT /v1/providers/{id}/connect` | Response object |

## Component Overview

### `page.tsx` (Home page)
- Client component (`"use client"`)
- Calls `checkDaemon()` and `window.racoreDesktop?.platform()` on mount
- Renders sub-views: System, Providers, Live Network, Sites, Agentic Browser
- Shows daemon connection status ("Running on 127.0.0.1:47831" vs "Not detected")

### `ProvidersView.tsx`
- Lists all AI providers in a grid layout
- Each card shows: name, default model, connection status, free/local badges
- Connect: input field + button, calls `connectProvider()`
- Disconnect: button calls `DELETE /v1/providers/{id}`
- Health check: calls `GET /v1/providers/{id}/health`

### `AgenticBrowserView.tsx`
- AI chat interface
- Calls `listProviders()` on mount to populate model selector
- Sends chat requests via `POST /v1/chat` with `{ provider, model, messages, system }`
- Displays streaming responses

### `LiveNetworkView.tsx`
- Shows mesh network status and IPFS node status
- Polls every 5 seconds: `GET /v1/mesh/status`, `GET /v1/mesh/peers`, `GET /v1/ipfs/status`
- Allows broadcasting test messages via `POST /v1/mesh/broadcast`

### `SitesView.tsx`
- Domain authority management
- Lists claimed domains via `GET /v1/authority/domains`
- Claim new domain via `POST /v1/authority/domains`
- Check domain availability via `GET /v1/authority/domains/{domain}/available`

### `Onboarding.tsx`
- First-run setup wizard
- Checks daemon health, lists providers, guides user through connecting first provider
- Uses `checkDaemon()`, `listProviders()`, `connectProvider()`

## Desktop Bridge Types (`app/types/desktop.d.ts`)

```typescript
interface Window {
  racoreDesktop?: {
    isDesktop: boolean;
    daemonUrl: string;
    status(): Promise<Record<string, unknown>>;
    api(request: {
      path: string;
      method?: string;
      body?: unknown;
    }): Promise<{ ok: boolean; status: number; data: unknown }>;
    platform(): Promise<{
      platform: string;
      version: string;
      packaged: boolean;
    }>;
    openBrowser(url: string): Promise<boolean>;
    openExternal(url: string): Promise<void>;
    onDaemonExit(callback: (code: number) => void): void;
  };
}
```

## Building

```bash
# Development server with hot reload
npm run dev

# Production build (outputs to dist/)
npm run build

# Production server
npm start
```

The build process uses Vinext, which runs 5 build steps internally:
1. Analyze client references
2. Analyze server references
3. Build RSC (React Server Components) environment
4. Build client environment
5. Build SSR (Server-Side Rendering) environment
