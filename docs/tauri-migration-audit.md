# Electron to Tauri v2 migration audit

Status: Step 1 complete on 2026-07-22. This document is the migration contract for the remaining implementation steps.

## Agreed boundaries

- Keep React and the existing Racore UI.
- Preserve the frontend-to-daemon HTTP workflow and JSON schemas. The Go daemon remains the authority for providers, vault data, IPFS, domains, and mesh operations.
- Do not modify `god/internal/mesh/**` during this migration.
- Remove Electron from the shipped application and use the current stable Tauri v2 ecosystem.
- Do not ship Node.js, Electron, `.node` native bindings, or `node-gyp` output. Native modules used only by the frontend build toolchain are development inputs and must never be bundled into the desktop application.
- Add or update documentation and automated tests in every implementation step, then create one Git commit for that step.

## Current architecture

The application has three runtime layers:

1. React UI built by Vinext/Vite.
2. `racored`, a Go HTTP/WebSocket daemon on `127.0.0.1:47831`.
3. Electron, which starts the daemon, hosts the production UI on `127.0.0.1:47832`, and provides renderer IPC.

The hosted web build also targets Cloudflare through `.openai/hosting.json`. The Tauri work must preserve that workflow. Desktop production assets therefore need a separate static React build target; the existing Electron production server imports `dist/server/index.js`, which is a Vinext server worker and cannot run inside Tauri without retaining a Node-compatible server runtime.

## Electron and Node.js inventory

### Desktop runtime imports

`desktop/main.cjs` imports:

| Import | Current responsibility | Tauri/Rust replacement |
| --- | --- | --- |
| `electron.app` | application lifecycle, version, resource path | Tauri `App`, `AppHandle`, package metadata, path resolver |
| `electron.BrowserWindow` | main and child webview windows | configured main window plus `WebviewWindowBuilder` |
| `electron.ipcMain` | five request handlers | typed `#[tauri::command]` functions |
| `electron.shell` | open URL in the system browser | a validated Rust command using the official opener/shell integration |
| `electron.session` | renderer permission filtering | Tauri capabilities, CSP, navigation policy, and webview permission policy |
| `child_process.spawn` | start `racored` | supervised Tauri sidecar process held in managed state |
| `path`, `fs` | resource resolution and static file reads | Tauri path APIs and `std::path`/`std::fs` |
| `http` | daemon readiness checks and packaged UI server | `reqwest` readiness checks; static Tauri `frontendDist` removes the UI server |
| `url.pathToFileURL` | dynamic import of the server worker | removed with the desktop static build |

`desktop/preload.cjs` imports `contextBridge` and `ipcRenderer`. The preload layer is removed completely and replaced by a typed TypeScript adapter over `@tauri-apps/api/core`.

### IPC contract

| Electron channel/event | Payload | Result | Current UI usage | Migration |
| --- | --- | --- | --- | --- |
| `racore:daemon-status` | none | daemon health JSON or `{ ok: false, error }` | bridge-only, not called by components | `daemon_status` command; retain during compatibility transition |
| `racore:api` | `{ path, method?, body? }` | `{ ok, status, data }` | all daemon API calls | typed `daemon_request` command preserving this envelope |
| `racore:open-browser` | URL string | boolean | `AgenticBrowserView` | validated `open_browser` command and isolated remote webview |
| `racore:open-external` | URL string | no value | bridge-only, currently unused | validated `open_external` command |
| `racore:platform` | none | `{ platform, version, packaged }` | `app/page.tsx` | `platform_info` command |
| `racore:daemon-exit` | numeric exit code | event | bridge-only, currently unused | `racore://daemon-exit` Tauri event with a typed payload |

The preload bridge exposes `status`, `api`, `platform`, `openBrowser`, `openExternal`, and `onDaemonExit`; only `api`, `platform`, and `openBrowser` are currently called by React components.

### Daemon routes currently requested by React

The compatibility adapter must continue accepting the existing method/path/body workflow for:

- `GET /health`
- `GET /v1/providers`
- `PUT /v1/providers/{provider}/connect`
- `DELETE /v1/providers/{provider}`
- `GET /v1/providers/{provider}/health`
- `POST /v1/chat`
- `GET /v1/mesh/status`
- `GET /v1/mesh/peers`
- `POST /v1/mesh/broadcast`
- `GET /v1/ipfs/status`
- `GET /v1/authority/domains`
- `GET /v1/authority/domains/{domain}/available`
- `POST /v1/authority/domains`

The Rust proxy must reject traversal, fragments, absolute URLs, credentials, and methods outside `GET`, `POST`, `PUT`, and `DELETE`. It must always target the fixed loopback daemon origin and preserve non-2xx response bodies in the existing response envelope.

### Window and lifecycle configuration

The Electron main window is `1480x940`, has minimum size `980x680`, dark background `#090c10`, title `Racore Browser`, hidden menu bar, sandboxing, context isolation, Node integration disabled, and spellcheck enabled. It loads `RACORE_UI_URL`, the packaged local server, or `http://localhost:3000`, with `desktop=1` appended.

External HTTP(S) navigation opens a `1280x820` child window with persistent storage partition `persist:racore-web`. Popups are denied and redirected into that child. The Tauri replacement must validate HTTP(S) URLs, keep remote webviews outside the main window capability, give them an isolated persistent data directory, and deny navigation to local/file/custom schemes.

On startup Electron starts the packaged UI server, starts or reuses `racored`, polls `/health` for up to 15 seconds at 350 ms intervals, then opens the main window. A non-zero daemon exit is emitted to the renderer. On quit, Electron kills its child daemon and closes the UI server. Tauri will remove the UI server and supervise only a daemon process that it started; it must not terminate an independently running daemon discovered during the health check.

### Bundled executables and native artifacts

- `god/build/racored` and `god/build/racore` are current Go build outputs.
- `desktop/runtime/kubo/ipfs.exe` is an 88,728,064-byte ignored runtime asset.
- `desktop/runtime/python/racored.exe` is an ignored legacy Python executable and is not referenced by the current Electron packaging configuration.
- Installed build dependencies currently contain native `.node` files for Sharp, Next SWC, Rolldown, Tailwind Oxide, and the resolver. A previous ignored `desktop-dist` Electron package also contains Sharp and Next SWC native files. These are build/stale-output artifacts, not application source, and the Tauri bundle verification test must prove that none are shipped.
- No application source imports Electron `remote`, `desktopCapturer`, tray, menu, native theme, or global shortcut APIs. There is no current tray, application-menu, or global-shortcut behavior to migrate.

## Security findings

- The existing renderer is hardened for Electron, but the generic `racore:api` proxy accepts every loopback path. Tauri will preserve the wire envelope while restricting the path grammar and fixed origin.
- Remote browser content must never receive the main window's Tauri capability. Capabilities will be bound by window label and limited to local app content.
- Frontend code does not need direct filesystem, OS, process, or unrestricted shell plugin permissions. Sensitive work stays in Rust commands.
- The daemon and Kubo executable paths are resolved by trusted backend code. The frontend must not be allowed to supply an executable or arguments.
- The main webview requires a production CSP. Network connections should be limited to the application asset origin and the fixed loopback daemon only where direct browser development needs it.
- Existing Go CORS configuration already permits the Electron UI origin. A Tauri-specific origin change, if required by the final asset protocol, belongs in `god/internal/server/**`, not `god/internal/mesh/**`, and must retain the current CORS tests.

## Heavy dependency disposition

- Electron and electron-builder are removed after the Tauri build is functional.
- The packaged Node HTTP/Vinext SSR runtime is removed from the desktop path by generating static React assets for `frontendDist`.
- `racored` remains Go rather than being rewritten in Rust. Rewriting its provider, vault, IPFS, authority, and mesh subsystems would change the established workflow and schema and is outside this shell migration.
- Kubo remains an external executable managed by `racored`; Tauri passes its trusted bundled path through `RACORE_KUBO_PATH`.
- Rust owns process supervision, health polling, loopback HTTP proxying, URL validation, platform metadata, and window lifecycle.

## Migration roadmap and commit checkpoints

### Step 1 — audit and migration contract

- Record this inventory, constraints, risk decisions, route schema, and verification plan.
- Add an automated contract test for the audit.
- Commit: `docs: audit Electron to Tauri migration`.

### Step 2 — scaffold Tauri v2 and desktop React build

- Add current stable Tauri v2 Rust/JavaScript dependencies and commit `src-tauri/Cargo.lock`.
- Create `src-tauri/`, `Cargo.toml`, `build.rs`, icons, strict `tauri.conf.json`, and a main-window-only capability.
- Add a static desktop React/Vite entry that reuses the current UI while preserving the existing Vinext/Cloudflare build.
- Add configuration and static-output tests. Build both web and desktop frontend targets.
- Commit: `build: scaffold Tauri v2 desktop application`.

### Step 3 — implement and test the Rust backend shell

- Add modules for typed commands, daemon HTTP, process supervision, platform metadata, URL validation, managed state, and window creation.
- Bundle architecture-suffixed `racored`, `racore`, and Kubo sidecars/resources. Remove the Python fallback from the desktop runtime.
- Emit daemon lifecycle events without blocking the UI thread.
- Unit-test request validation, URL normalization, response decoding, process-state transitions, and platform payload serialization. Do not edit `god/internal/mesh/**`.
- Commit: `feat: implement Tauri daemon and window backend`.

### Step 4 — replace the Electron frontend bridge

- Add a typed Tauri adapter using `invoke` and `listen` while preserving `daemonRequest` call sites and API envelopes.
- Replace `window.racoreDesktop` checks and declarations with an environment-safe desktop abstraction.
- Add TypeScript tests for command names, payload mapping, error propagation, browser opening, platform information, and daemon-exit unsubscribe behavior.
- Update architecture and data-flow documentation.
- Commit: `refactor: replace Electron IPC with Tauri bridge`.

### Step 5 — remove Electron and verify distribution

- Delete `desktop/main.cjs`, `desktop/preload.cjs`, and Electron-only package configuration/dependencies/scripts.
- Add Tauri development, build, and bundle scripts and update all desktop documentation.
- Run frontend tests, lint, TypeScript validation, Rust format/lint/tests, unchanged Go tests where the Go toolchain is available, both frontend builds, and `tauri build`.
- Inspect bundle contents and assert that Electron, Node.js, `.node`, and `node-gyp` artifacts are absent. Record installer and unpacked sizes plus a repeatable idle-memory measurement procedure; compare with an Electron baseline only if a reproducible baseline is available.
- Commit: `chore: complete Electron removal and Tauri verification`.

## Verification matrix

| Area | Required evidence |
| --- | --- |
| React behavior | existing rendered/UI contract tests plus Tauri adapter unit tests |
| Rust backend | `cargo test`, `cargo fmt --check`, and Clippy with warnings denied |
| Go contract | existing server tests; no diff under `god/internal/mesh/**` |
| Hosted workflow | existing Vinext/Cloudflare production build remains successful |
| Desktop workflow | static React build, `tauri dev`, and release `tauri build` |
| Security | capability snapshot tests, URL/path validator tests, remote-window isolation |
| Packaging | automated archive/directory scan for forbidden Electron/Node/native artifacts |
| Lifecycle | daemon reuse, startup timeout, abnormal-exit event, and owned-child shutdown tests |

## Toolchain audit

The audit host has Rust `1.97.1`, Cargo `1.97.1`, Node `24.16.0`, and npm `11.13.0`. Go is not available on `PATH`, so unchanged Go tests cannot run on this host until that toolchain is installed or exposed. At audit time the registry reports `@tauri-apps/api` `2.11.1`, `@tauri-apps/cli` `2.11.4`, `@tauri-apps/plugin-shell` `2.3.5`, and `@tauri-apps/plugin-os` `2.3.2`; Step 2 will lock the then-current stable compatible versions rather than using floating ranges.
