# Racore Protocol reference implementation

This directory contains the executable RCP 0.1 local reference implementation.

It provides deterministic file indexes, SHA-256 content roots, Ed25519 publisher identities, canonical signed release manifests, an append-only release history, verification, and safe rollback.

```bash
npm run racore -- init --domain example.com
npm run racore -- build --input ./dist
npm run racore -- publish --version 1.0.0 --commit abc123 --input ./dist
npm run racore -- releases
npm run racore -- verify
```

The `.racore/registry.json` file is the local mutable name record. A production resolver publishes its current manifest ID through `_racore.<domain>` and retrieves blocks from configured HTTPS, IPFS, or Racore peer transports.
