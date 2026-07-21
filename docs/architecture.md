# Racore architecture

## Runtime boundaries

```text
Browser UI
├─ Browser core ── Chromium renderer processes
├─ Agent orchestrator
│  ├─ plan and model adapter
│  ├─ typed tool broker
│  ├─ policy and approval engine
│  └─ append-only audit chain
├─ Identity vault
└─ RCP resolver
   ├─ DNS authorization
   ├─ signature and hash verifier
   ├─ local cache
   └─ HTTPS / IPFS / peer transports
```

Web renderers, agents, secrets, and protocol networking are separate trust domains. The model receives semantic page observations and typed tool results, not raw secret material or unrestricted process access.

## Repository map

- `app/` — working Racore control surface and browser/agent experience.
- `protocol/` — dependency-free executable RCP reference implementation and CLI.
- `specs/` — wire behavior, resolution, origin, transport, and security requirements.
- `tests/` — protocol lifecycle and tamper-resistance tests.
- `examples/hello-racore/` — runnable site used to exercise publishing locally.

## Production progression

1. Keep the current web control surface as the shared UI package.
2. Add an Electron desktop host with hardened `BrowserView` partitions and a Rust broker process.
3. Replace the local JSON name registry with DNS and a signed transparency service.
4. Add CIDv1/IPFS and libp2p transport adapters behind the transport interface.
5. Freeze deterministic CBOR interoperability vectors.
6. Conduct renderer, updater, key-vault, and agent-tool security reviews before general release.
