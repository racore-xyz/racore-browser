# Tauri desktop development

Racore now has a Tauri v2 scaffold alongside the Electron shell during the staged migration. React components remain under `app/` and are consumed by two build pipelines:

- `npm run build` keeps the existing Vinext/Cloudflare hosted workflow.
- `npm run desktop:ui` creates static React assets in `dist-desktop/` for the Tauri webview.

The desktop entry is `desktop-ui/main.tsx`. Its Vite configuration aliases `next/image` to a web-standard image component so the same React UI can run without a Next.js runtime. Public assets continue to come from `public/`.

## Commands

```powershell
npm run desktop:ui
npm run tauri:dev
npm run tauri:build
```

`src-tauri/tauri.conf.json` retains the Electron main-window dimensions and uses a production content security policy. `src-tauri/capabilities/main.json` applies only to the trusted `main` window. Remote browsing windows will use different labels and will not inherit this capability when they are implemented.

The scaffold intentionally does not expose filesystem, operating-system, or shell plugin APIs to React. Desktop-native operations are implemented as application commands in the next migration step.

## Backend commands and sidecars

The Rust backend is split across `src-tauri/src/commands.rs`, `daemon.rs`, `state.rs`, `windows.rs`, `models.rs`, and `error.rs`. It exposes five typed commands that replace the five Electron handlers while retaining the existing daemon response envelope. Daemon requests are restricted to the React application's current route/method combinations and always use `http://127.0.0.1:47831`.

`npm run tauri:prepare` builds `racored` and `racore` from the unchanged Go source for the current Rust target triple and copies the matching Kubo executable into generated Tauri resource directories. Set `RACORE_GO` when the Go executable is not on `PATH`, `RACORE_KUBO_BINARY` for a platform-specific Kubo binary, or `RACORE_TAURI_TARGET` for an explicit cross-compilation target.

At startup, Tauri reuses a healthy daemon already listening on the fixed loopback port. Otherwise it starts the bundled `racored` sidecar, passes only the trusted bundled Kubo path, waits up to 15 seconds, and retains the child handle in managed state. Only an owned child is killed during shutdown. Process termination emits `racore://daemon-exit`.

Browser commands accept credential-free HTTP(S) URLs only. In-app browser windows use labels outside the main capability, a dedicated persistent webview data directory, an HTTP(S)-only navigation policy, and denied popup creation.

The Go UDP transport now routes `SO_REUSEADDR` and read-side shutdown through build-tagged Unix and Windows syscall adapters. This is a platform compilation fix only: packet formats, discovery, heartbeat, peer handling, and all files under `god/internal/mesh/` remain unchanged.

## React desktop adapter

`app/lib/desktop.ts` is the single typed desktop boundary. It maps React calls to `daemon_status`, `daemon_request`, `platform_info`, `open_browser`, and `open_external` with `@tauri-apps/api/core`, and maps daemon termination to the `racore://daemon-exit` event with `@tauri-apps/api/event`. Event subscriptions return Tauri's unsubscribe function so components can release listeners during unmount.

`app/lib/racore-client.ts` keeps the established `daemonRequest(path, { method, body })` workflow. In the Tauri static build it calls the Rust proxy; in the hosted Vinext build it continues to fetch the loopback daemon directly. HTTP success/error envelopes and provider request schemas are unchanged.

The desktop entry installs a temporary `window.racoreDesktop` compatibility facade backed entirely by Tauri. This allows older UI surfaces to coexist during the migration without a preload script; new and migrated code imports the typed adapter directly.
