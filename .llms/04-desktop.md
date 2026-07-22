# Tauri desktop

## Layout

```text
desktop-ui/                  static React entry and Vite configuration
src-tauri/
  capabilities/main.json    trusted main-window capability
  src/commands.rs            typed invoke commands
  src/daemon.rs              fixed-origin HTTP client and route allowlist
  src/state.rs               daemon process ownership and lifecycle events
  src/windows.rs             URL validation and isolated remote webviews
  src/models.rs              serde request/response types
  tauri.conf.json            shared application and bundle configuration
  tauri.*.conf.json          platform-specific Kubo resources
```

The main window is 1480x940 with a 980x680 minimum and dark background. Static files come from `dist-desktop`; no local UI server runs in production.

Rust exposes five commands matching the established desktop workflow. `daemon_request` accepts only current Racore route/method pairs, rejects traversal and origin escape, limits responses to 8 MiB, and always uses `127.0.0.1:47831`.

`scripts/prepare-tauri-sidecars.mjs` cross-builds `racored` and `racore` for the Rust target triple and copies the matching Kubo executable. Tauri supervises only the daemon process it starts.

Remote browsing accepts credential-free HTTP(S) URLs, uses an isolated persistent data directory, denies popups, and rejects navigation to other schemes.

Build with `npm run desktop:package`; verify the release tree with `npm run tauri:verify`.
