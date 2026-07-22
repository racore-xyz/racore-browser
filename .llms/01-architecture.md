# Architecture

Racore Browser is a three-tier application:

1. React UI: hosted Vinext/Cloudflare output and static Tauri output.
2. Tauri v2 shell: Rust commands, window policy, URL validation, and sidecar ownership.
3. Go daemon: REST/WebSocket API, provider gateway, vault, authority, Kubo/IPFS, and mesh.

Desktop calls flow through `app/lib/desktop.ts` to Rust. Hosted calls flow directly to `127.0.0.1:47831`. Both converge in `app/lib/racore-client.ts` and preserve the same method/path/body and JSON response schema.

The desktop application bundles static React assets, `racored`, `racore`, and Kubo. It contains no Node.js runtime or native Node binding. Rust reuses a healthy external daemon, otherwise starts and owns the bundled daemon, emits `racore://daemon-exit` on termination, and stops only its owned child.

Security boundaries are window-label based. Only trusted local application content uses the `main` label/capability. Remote HTTP(S) webviews are isolated and cannot navigate to local or custom schemes.
