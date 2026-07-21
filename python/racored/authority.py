from __future__ import annotations

import json
import hashlib
import time
from pathlib import Path
from typing import Any

from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PrivateKey

from .config import data_directory
from .mesh import b64


class DomainAuthority:
    def __init__(self, root: Path | None = None):
        self.root = root or data_directory()
        self.path = self.root / "authorities.json"
        self.root.mkdir(parents=True, exist_ok=True)

    def _read(self) -> dict[str, Any]:
        state = json.loads(self.path.read_text(encoding="utf-8")) if self.path.exists() else {"domains": {}}
        state.setdefault("domains", {})
        state.setdefault("observed", {})
        return state

    def _write(self, data: dict[str, Any]) -> None:
        self.path.write_text(json.dumps(data, indent=2), encoding="utf-8")

    def create(self, domain: str) -> dict[str, Any]:
        state = self._read()
        if not self.available(domain):
            raise ValueError(f"Domain {domain} is already claimed on the known Racore network")
        private = Ed25519PrivateKey.generate()
        public = private.public_key().public_bytes(serialization.Encoding.Raw, serialization.PublicFormat.Raw)
        private_path = self.root / f"authority-{domain.replace('.', '-')}.pem"
        private_path.write_bytes(private.private_bytes(serialization.Encoding.PEM, serialization.PrivateFormat.PKCS8, serialization.NoEncryption()))
        key_id = f"did:key:{b64(public)}"
        record = {"domain": domain, "controller": key_id, "createdAt": int(time.time()), "delegates": [], "status": "active", "releases": [], "current": None}
        state["domains"][domain] = record; self._write(state)
        return {**record, "dnsRecords": [f'_racore-key.{domain} TXT "ed25519={b64(public)}"', f'_racore.{domain} TXT "rcp=0.1 manifest=<release-cid>"']}

    def list(self) -> list[dict[str, Any]]:
        return list(self._read()["domains"].values())

    def available(self, domain: str) -> bool:
        state = self._read()
        return domain not in state["domains"] and domain not in state["observed"]

    def observe_claim(self, domain: str, controller: str, node_id: str) -> None:
        state = self._read()
        if domain not in state["domains"]:
            state["observed"][domain] = {"domain": domain, "controller": controller, "nodeId": node_id, "lastSeen": int(time.time())}
            self._write(state)

    def publish_release(self, domain: str, release: dict[str, Any]) -> dict[str, Any]:
        state = self._read()
        if domain not in state["domains"]:
            raise KeyError(domain)
        private_path = self.root / f"authority-{domain.replace('.', '-')}.pem"
        private = serialization.load_pem_private_key(private_path.read_bytes(), password=None)
        manifest = {
            "protocol": "rcp/0.2",
            "domain": domain,
            "version": release["version"],
            "cid": release["cid"],
            "contentRoot": release["contentRoot"],
            "entrypoint": release.get("entrypoint", "index.html"),
            "files": release["files"],
            "size": release["size"],
            "createdAt": int(time.time()),
            "parent": state["domains"][domain].get("current"),
        }
        canonical = json.dumps(manifest, sort_keys=True, separators=(",", ":")).encode()
        signature = b64(private.sign(canonical))
        release_id = "rcp2-" + hashlib.sha256(canonical + signature.encode()).hexdigest()
        signed = {**manifest, "releaseId": release_id, "signature": signature}
        record = state["domains"][domain]
        record["current"] = release_id
        record.setdefault("releases", []).insert(0, signed)
        self._write(state)
        return signed

    def releases(self, domain: str) -> list[dict[str, Any]]:
        state = self._read()
        if domain not in state["domains"]:
            raise KeyError(domain)
        return state["domains"][domain].get("releases", [])

    def delegate(self, domain: str, public_key: str, capabilities: list[str], expires_at: int) -> dict[str, Any]:
        state = self._read()
        if domain not in state["domains"]:
            raise KeyError(domain)
        grant = {"publicKey": public_key, "capabilities": capabilities, "expiresAt": expires_at, "createdAt": int(time.time())}
        state["domains"][domain]["delegates"].append(grant)
        self._write(state)
        return grant
