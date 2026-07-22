# Racore Browser

Racore is a local-first agentic browser and open-web protocol application. The desktop product uses React in a Tauri v2 webview, a Rust shell, and the existing Go `racored` daemon.

## Architecture

```text
Tauri v2 desktop
  React static UI (dist-desktop)
       │ typed invoke/events
       ▼
  Rust commands and managed process state
       │ fixed-origin HTTP
       ▼
racored Go daemon (127.0.0.1:47831)
  providers · encrypted vault · authority · IPFS/Kubo · Racore mesh
```

The hosted React application keeps its Vinext/Cloudflare build. Both modes share the components and `daemonRequest(path, { method, body })` schema:

- Hosted mode fetches the loopback daemon directly.
- Desktop mode uses `app/lib/desktop.ts` and typed Rust commands.

## Prerequisites

- Node.js 22.13 or newer
- Rust 1.77.2 or newer with platform Tauri prerequisites
- Go 1.21 or newer
- A platform-matching Kubo executable

On Windows, place Kubo at `desktop/runtime/kubo/ipfs.exe` or set `RACORE_KUBO_BINARY`. Set `RACORE_GO` when `go` is not on `PATH`.

## Development

```powershell
npm install

# Hosted React development
npm run dev

# Native desktop development (builds Go sidecars first)
npm run desktop:dev
```

The Tauri command starts or reuses `racored` on `127.0.0.1:47831`. A daemon started outside the app is never terminated by the app.

## Build and test

```powershell
npm test
npm run lint
cargo test --manifest-path src-tauri/Cargo.toml
cargo clippy --manifest-path src-tauri/Cargo.toml --all-targets -- -D warnings

cd god
go test -race -count=1 ./internal/...
cd ..

npm run desktop:package
npm run tauri:verify
```

The desktop package is written below `src-tauri/target/release/bundle/`. Tauri bundles architecture-suffixed `racored` and `racore` sidecars plus Kubo; it does not bundle a Node.js runtime or native Node modules.

## Project structure

```text
app/                    React UI and typed desktop/daemon clients
desktop-ui/             Static Vite entry for Tauri
src-tauri/              Rust backend, capabilities, config, icons
scripts/                Sidecar preparation and bundle verification
god/                    Go daemon and CLI
protocol/               RCP protocol JavaScript tools
worker/                 Cloudflare Worker entry
tests/                  JavaScript/TypeScript contract tests
docs/                   Architecture and migration documentation
```

See [Tauri desktop development](docs/tauri-development.md), [architecture](docs/architecture.md), and [the migration audit](docs/tauri-migration-audit.md).
