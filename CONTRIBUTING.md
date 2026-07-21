# Contributing to Racore Browser

## Getting Started

1. Read `CLAUDE.md` for a project overview
2. Read the documentation in `.llms/` for detailed section-by-section guides
3. Ensure you have Go >= 1.21 and Node.js >= 22.13.0 installed
4. Run through the full build and test suite once before making changes:

```bash
# Build the Go backend
cd god && go build -o build/racored ./cmd/racored/ && go build -o build/racore ./cmd/racore/ && cd ..

# Run all Go tests with race detection
cd god && go test -race -count=1 ./internal/... && cd ..

# Run Go static analysis
cd god && go vet ./... && cd ..

# Install frontend dependencies (if not done yet)
npm install

# Build the frontend
npm run build

# Run protocol tests
npm run protocol:test

# Run full test suite
npm test
```

## Code of Conduct

- Be respectful and constructive
- Assume good intent
- Focus on what is best for the project

## Workflow

1. Fork the repository or create a feature branch
2. Make focused, atomic commits
3. Run all tests before submitting
4. Submit a pull request with a clear description

## Commit Messages

Use conventional commits:

```
type(scope): concise description

body (optional, wrapped at 72 characters)
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `style`, `perf`

Scopes: `frontend`, `daemon`, `desktop`, `api`, `mesh`, `vault`, `transport`, `ipfs`, `authority`, `config`, `deps`

Examples:
```
feat(mesh): add peer discovery via heartbeat broadcast

fix(transport): unblock read loop on Close using syscall.Shutdown

docs(api): document all REST endpoints in .llms/05-api-reference.md
```

## Pull Requests

- Title must match the conventional commit format
- Description must explain what and why, not how
- Link to any related issues
- Include screenshots for UI changes
- Keep PRs focused on a single concern

## Before Submitting

You must build the entire project and run all tests before submitting. Follow these steps in order:

### 1. Build Everything

```bash
# Go backend -- must compile without errors
cd god && go build -o build/racored ./cmd/racored/ && go build -o build/racore ./cmd/racore/ && cd ..

# Frontend -- must build without errors
npm run build
```

### 2. Run All Tests

```bash
# Go tests (all 12 packages with race detection)
cd god && go test -race -count=1 ./internal/... && cd ..

# Go static analysis
cd god && go vet ./... && cd ..

# Protocol tests
npm run protocol:test

# Full test suite (build + protocol)
npm test
```

### 3. Verify Checklist

#### All Changes

### Go Changes (`god/`)

- [ ] New code has tests with `-race` enabled
- [ ] All tests pass: `go test -race -count=1 ./internal/...`
- [ ] Test coverage has not regressed
- [ ] New API endpoints are registered in `god/internal/server/server.go`
- [ ] Error responses use `{"detail": "..."}` format (not `{"error": "..."}`)

### Frontend Changes (`app/`)

- [ ] TypeScript compiles without errors
- [ ] Frontend build succeeds: `npm run build`
- [ ] New API calls go through `app/lib/racore-client.ts` `daemonRequest()`
- [ ] TypeScript interfaces match their Go counterparts in `god/pkg/api/types.go`

### Desktop Changes (`desktop/`)

- [ ] Tested with `npm run desktop` (packaged mode)
- [ ] Tested with `npm run desktop:dev` (development mode)
- [ ] IPC handler changes are reflected in `app/types/desktop.d.ts`
- [ ] Daemon discovery logic in `daemonCommand()` is updated if binary paths change

### Documentation Changes

- [ ] `.llms/` files are updated to reflect any changes in architecture, API, or types
- [ ] `CLAUDE.md` is updated if top-level project conventions change
- [ ] API changes are documented in `.llms/05-api-reference.md` with request/response examples

## Architecture Decisions

### Go Daemon

- Single static binary, zero runtime dependencies
- Standard library HTTP router (`net/http` with `http.ServeMux`)
- AES-256-GCM for secret storage (key derived via PBKDF2)
- UDP multicast for P2P mesh transport
- Ed25519 for message signing and identity

### Frontend

- All daemon communication goes through `app/lib/racore-client.ts`
- Two transport modes: direct HTTP (browser) and Electron IPC (desktop)
- No state management library: component-local state + props
- Tailwind CSS for styling, no CSS modules or styled components

## Testing Guidelines

### Go Tests

- Run with `-race` to detect data races
- Use `-count=1` to disable test caching
- Mesh tests that require multicast should skip gracefully when unavailable
- Mock external dependencies (network, file system) where possible
- Test both success and error paths

### Frontend Tests

- Protocol tests live in `tests/` and run via `node --test`
- Component tests can use the testing framework of your choice

## Adding a New API Endpoint

1. Define request/response types in `god/pkg/api/types.go` (if new)
2. Implement the handler in `god/internal/server/handlers.go`
3. Register the route in `god/internal/server/server.go` (`mux.HandleFunc`)
4. Add the frontend function in `app/lib/racore-client.ts`
5. Document in `.llms/05-api-reference.md`
6. Update Go types documentation in `.llms/07-types.md`

## Adding a New AI Provider

1. Add provider metadata in `god/internal/providers/providers.go` (add to `catalog`)
2. Implement the API adapter in `god/internal/gateway/` if the API format differs from OpenAI
3. Test with `go test -race -count=1 ./internal/providers/...`

## Questions

Open an issue for:
- Feature requests
- Bug reports
- Architecture discussions
- Documentation improvements
