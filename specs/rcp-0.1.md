# Racore Protocol 0.1

Status: implementable draft  
Media type: `application/vnd.racore.manifest+json`  
Canonical signed form: deterministic CBOR in production; sorted-key canonical JSON in the reference implementation

## Purpose

RCP binds a human-controlled domain to a signed immutable web release that can be retrieved from several independent transports. It separates mutable naming from immutable content and makes every domain update independently verifiable and reversible.

## Objects

### Content block

A content block is addressed by the SHA-256 digest of its bytes. The reference identifier is `rcp1-<base64url-sha256>`. Production transport adapters may map the same digest to a CIDv1 with the raw or DAG-CBOR codec.

### Bundle index

The bundle index is an ordered list of `{ path, size, hash }` records. Paths use `/`, are relative, contain no `..` segment, and sort by Unicode code point. The bundle root is the content identifier of its canonical index.

### Release manifest

Required unsigned fields:

```json
{
  "protocol": "rcp/0.1",
  "domain": "example.com",
  "version": "1.0.0",
  "root": "rcp1-...",
  "parent": "rcp1-...",
  "commit": "abc123",
  "createdAt": "2026-07-21T12:00:00.000Z",
  "entrypoint": "/index.html",
  "mirrors": ["https://origin.example.com/releases/1.0.0/"]
}
```

The signed object adds an Ed25519 public key and signature. Its `manifestId` is calculated after the signature is attached and before `manifestId` itself is attached.

## Domain authorization

The domain owner publishes:

```dns
_racore.example.com TXT "rcp=0.1 manifest=rcp1-..."
```

Optional key pinning:

```dns
_racore-key.example.com TXT "ed25519=<base64url-spki-digest>"
```

Resolvers must reject records whose domain does not match the manifest domain. When key pinning is present, the publisher key must match. DNSSEC validation is recommended; without it the UI must disclose that domain authorization depends on ordinary DNS transport security.

## Resolution algorithm

1. Normalize the requested domain using IDNA and remove a trailing dot.
2. Query `_racore.<domain>` TXT records.
3. Select the highest supported RCP version with a valid manifest identifier.
4. Fetch the manifest concurrently from local cache, declared gateways, HTTPS mirrors, and peer routing.
5. Require byte-identical canonical objects from at least the configured trust quorum.
6. Recalculate the manifest identifier.
7. Verify the Ed25519 signature and optional pinned publisher key.
8. Confirm the manifest domain equals the requested domain.
9. Fetch the bundle index, recalculate its root, then fetch content blocks.
10. Verify every block before making it visible to a renderer.
11. Load under an origin derived from the requested domain, never a shared gateway origin.
12. Fall back to ordinary HTTPS when no valid RCP record exists.

## Transports

Transport adapters implement `has(id)`, `get(id)`, `put(bytes)`, and optional `announce(id)` operations.

- **Local cache:** mandatory and always checked first.
- **HTTPS:** mandatory compatibility transport with range requests and immutable caching.
- **IPFS:** recommended decentralized transport using CIDv1 mappings.
- **Racore peer:** optional libp2p transport with provider discovery.
- **Object storage:** publishing adapter for S3-compatible stores; retrieval occurs over HTTPS.

Transport responses are never trusted without content verification.

## Version history and rollback

Every manifest points to its parent. A registry maintains the currently selected manifest and an ordered history. Rollback changes only the mutable registry pointer after verifying the target release; it never mutates or rebuilds old content.

Implementations should publish release statements to an append-only transparency service. Equivocation detection compares observed statements for the same domain and sequence.

## Browser origin model

- RCP content for `example.com` receives the same logical origin only after domain authorization succeeds.
- Raw content identifiers and shared gateways receive isolated opaque or subdomain origins.
- Service workers are scoped to the verified logical origin and release policy.
- Active content is subject to standard CSP, permissions policy, mixed-content, and sandbox enforcement.
- A release update must not silently expand browser or agent capabilities.

## Agent capability model

Agent tools are brokered separately from web content. Grants are scoped by agent, task, origin, capability, resource, and expiration. External effects—publishing, messages, uploads, purchases, identity use, deletion, and shell execution—require an approval record unless an explicit narrower policy exists.

## Failure behavior

- Invalid signature: block RCP load and offer HTTPS fallback.
- Root mismatch: discard the source, penalize the transport, try another source.
- Domain mismatch: block without fallback to the manifest's declared domain.
- Missing block: retry other transports; never render partial active content.
- Unsupported protocol: ignore the record and offer HTTPS fallback.
- Stale but valid release: allow cached offline load with a visible stale-state warning.

## Test vectors

The executable tests in `tests/protocol.test.mjs` cover canonical ordering, two-release publication, content change detection, signature verification, rollback, and manifest tampering. Additional interoperability vectors will freeze canonical CBOR bytes before RCP 0.1 is marked stable.
