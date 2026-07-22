# Build and test

## Prerequisites

- Node.js 22.13+
- npm 10+
- Rust 1.77.2+
- Go 1.21+
- platform Tauri prerequisites and a matching Kubo binary

## Hosted application

```powershell
npm run dev
npm run build
```

The hosted output remains under `dist/` and retains the Vinext/Cloudflare workflow.

## Desktop application

```powershell
npm run tauri:prepare
npm run desktop:ui
npm run desktop:dev
npm run desktop:package
npm run tauri:verify
```

`tauri:prepare` uses `RACORE_GO`, `RACORE_KUBO_BINARY`, and `RACORE_TAURI_TARGET` when supplied. Generated sidecars/resources are ignored; source and lockfiles are committed.

## Tests

```powershell
npm test
npm run lint
npx tsc --noEmit

cargo fmt --manifest-path src-tauri/Cargo.toml --check
cargo clippy --manifest-path src-tauri/Cargo.toml --all-targets -- -D warnings
cargo test --manifest-path src-tauri/Cargo.toml

cd god
go test -race -count=1 ./internal/...
go vet ./...
```

The JavaScript suite covers protocol behavior, hosted rendering, migration contracts, Tauri configuration, bridge payloads/events, and forbidden bundle artifacts. Rust tests cover request/path validation and browser URL normalization. Go retains twelve internal package suites.

Race detection may require a platform C toolchain. When it is unavailable, run `go test -count=1 ./internal/...` and record the limitation.
