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

The Rust layer is intentionally a narrow broker. It does not duplicate local AI, authority, IPFS, or mesh logic. `racored` retains those responsibilities and the established REST/JSON schemas.

The main webview receives the capability in `src-tauri/capabilities/main.json`. Remote HTTP(S) browsing windows use labels outside that capability, a separate persistent data directory, restricted navigation, and denied popups.

At startup Rust checks daemon health, reuses an external daemon when present, or starts the bundled sidecar and waits up to 15 seconds. Managed state stores only an app-owned child handle; shutdown cannot kill an independently started daemon.

## Local ports

| Service | Address | Notes |
| --- | --- | --- |
| `racored` HTTP + WebSocket | `127.0.0.1:47831` | Loopback only; `/v1/events` streams runtime events. |
| Tauri preview / dev UI | `127.0.0.1:1420` | Allowed CORS origin for the webview build. |
| Hosted preview | `http://localhost:3000` | Vinext dev server; allowed for local inspection. |
| Kubo RPC | `127.0.0.1:5001` | Publishing and pinning; must be running for release uploads. |
| Kubo gateway | `127.0.0.1:8180` | Serves verified UnixFS releases to the browser. |
| Local AI | `127.0.0.1:47834` | llama.cpp `llama-server` hosting Hammer 2.0 Q8_0. |
| Mesh transport | UDP `239.255.77.77:47777` | Multicast discovery plus signed RMP heartbeats. |

On Windows the multicast socket opts into `SO_REUSEADDR` before binding so several daemons on one host can join the group. `x/net/ipv4` does not implement control-message reception on Windows, so the daemon skips `SetControlMessage` there and relies on `ReadFrom` for payload and source address.

## Approval boundary

The planner never executes a side effect directly. `/v1/chat` returns text plus zero or more proposed `actions` (`open_url`, `search_web`, `resolve_racore_domain`). The UI queues those proposals in the approval dock, and only an explicit Approve gesture resolves a Racore domain or opens a destination. Deny discards the proposal. External navigation goes through the Rust `open_browser` command, which validates the URL and opens it in an isolated webview window.

## Publishing and verification

`racore publish --domain <name> --build <dir> --version <semver>` hashes every file, uploads it to the local Kubo node as a UnixFS directory that keeps relative paths (`assets/…` stays nested), and asks the authority to sign an RCP/0.2 manifest. Manifests carry the controller DID, release ID, domain, CID, version, and file list; the authority stores the Ed25519 signature and exposes it through `/v1/authority/domains/<domain>/releases`.

Resolution (`/v1/authority/resolve/<domain>`) returns a manifest only when the signature verifies against a known controller. The UI distinguishes `verified` from unverified records and never falls back to an unverified CID. Gateways serve bytes from IPFS; verification establishes that those bytes are the ones the controller signed.

Builds must reference their assets relatively. A path gateway serves a release under `/ipfs/<cid>/`, so a root-absolute reference such as `/assets/app.js` misses the release even though the file is inside it. `racore publish` warns when HTML, CSS, or JS reference a known build file that way; set Vite's `base` to `"./"` (the desktop UI does) before publishing.

## Schema conventions

Every API payload uses `camelCase` JSON fields and millisecond Unix timestamps (`createdAt`, `updatedAt`, `lastSeen`). Domain names must end in `.racore`, `.rac`, `.core`, or `.ra`; public domains are rejected by the authority and by the Rust URL validator.

