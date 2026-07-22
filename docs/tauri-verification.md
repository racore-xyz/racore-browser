# Tauri migration verification

Verification was completed on 2026-07-22 on Windows x64. The desktop application uses React 19 with Tauri 2, a Rust command backend, and the existing Go daemon. No file under `god/internal/mesh/` was changed by the migration.

## Automated checks

The following commands passed:

```powershell
npm test
npm run lint
npx tsc --noEmit
npm run desktop:package
npm run tauri:verify
cargo fmt --manifest-path src-tauri/Cargo.toml -- --check
cargo test --manifest-path src-tauri/Cargo.toml
cargo clippy --manifest-path src-tauri/Cargo.toml --all-targets -- -D warnings
go test -count=1 ./internal/...
```

`npm run tauri:verify` inspects the release executable, installers, and bundled resources. It fails if a shipped deliverable contains Electron, Node.js, `node_modules`, a `.node` native binding, or a `node-gyp` artifact. The repository's remaining `electron-to-chromium` package is browser compatibility metadata used by the frontend toolchain; it is not the Electron runtime.

The Rust suite covers URL validation, daemon route/method authorization, and response-size parsing. The Node suite covers the typed Tauri bridge, scaffold and capability policy, sidecar preparation, and release-bundle inspection. The existing Go internal test suite continues to cover the daemon and mesh implementation without migration changes to the mesh package.

## Release artifacts

The successful local package build produced:

| Artifact | Size |
| --- | ---: |
| Tauri MSI | 59,961,344 bytes |
| Tauri NSIS installer | 49,259,494 bytes |
| Tauri application executable | 22,401,024 bytes |
| Previous Electron NSIS installer | 211,861,743 bytes |
| Previous Electron shell executable | 205,883,904 bytes |

Compared with the previous Electron output, the Tauri NSIS installer is 76.7% smaller and its application executable is 89.1% smaller. These figures are local build measurements and can vary slightly with toolchain, signing, and compression settings.

## Runtime memory sample

Each release executable was launched independently on the same machine, allowed 12 seconds to settle, and measured as the sum of the working sets of the newly created application process tree:

| Runtime | Processes | Working set |
| --- | ---: | ---: |
| Tauri | 2 | 131.4 MiB |
| Electron | 4 | 337.1 MiB |

The sampled Tauri process tree used 61.0% less working-set memory. This is a smoke-test comparison rather than a benchmark; repeat it under representative workloads for capacity planning.

## Distribution notes

The local installers are unsigned development artifacts. Production releases must be signed with the organization's Windows code-signing certificate in the release pipeline. The generated Electron distribution and obsolete Python daemon fallback were removed after comparison; Tauri packages only the Rust application, the Go daemon sidecars, and Kubo.
