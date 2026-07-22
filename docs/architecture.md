# Racore Browser architecture

Racore separates presentation, native desktop authority, and network/service authority.

```text
React components
  ├─ hosted: fetch http://127.0.0.1:47831
  └─ desktop: @tauri-apps/api invoke
                    │
                    ▼
              typed Rust commands
              ├─ request allowlist
              ├─ URL validation
              ├─ webview windows
              └─ owned sidecar state
                    │
                    ▼
             racored Go daemon
```

The Rust layer is intentionally a narrow broker. It does not duplicate provider, vault, authority, IPFS, or mesh logic. `racored` retains those responsibilities and the established REST/JSON schemas.

The main webview receives the capability in `src-tauri/capabilities/main.json`. Remote HTTP(S) browsing windows use labels outside that capability, a separate persistent data directory, restricted navigation, and denied popups.

At startup Rust checks daemon health, reuses an external daemon when present, or starts the bundled sidecar and waits up to 15 seconds. Managed state stores only an app-owned child handle; shutdown cannot kill an independently started daemon.
