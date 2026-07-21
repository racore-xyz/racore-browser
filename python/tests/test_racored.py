from pathlib import Path

import pytest

from racored.authority import DomainAuthority
from racored.mesh import NodeIdentity
from racored.vault import CredentialVault


def test_vault_round_trip(tmp_path: Path):
    vault = CredentialVault(tmp_path)
    vault.set("openai", "sk-test-123456789")
    assert vault.get("openai") == "sk-test-123456789"
    assert vault.masked("openai") == "sk-t…6789"
    vault.delete("openai")
    assert vault.get("openai") is None


def test_mesh_identity_is_stable(tmp_path: Path):
    first = NodeIdentity(tmp_path)
    second = NodeIdentity(tmp_path)
    assert first.node_id == second.node_id
    assert first.sign(b"hello") == second.sign(b"hello")


def test_domain_authority_records(tmp_path: Path):
    authority = DomainAuthority(tmp_path)
    result = authority.create("example.com")
    assert result["domain"] == "example.com"
    assert result["controller"].startswith("did:key:")
    assert "_racore-key.example.com" in result["dnsRecords"][0]
    assert authority.available("example.com") is False
    with pytest.raises(ValueError):
        authority.create("example.com")


def test_domain_release_is_signed_and_parent_linked(tmp_path: Path):
    authority = DomainAuthority(tmp_path)
    authority.create("app.example.com")
    first = authority.publish_release("app.example.com", {"version": "1.0.0", "cid": "bafkreiaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "contentRoot": "a" * 64, "entrypoint": "index.html", "files": 2, "size": 120})
    second = authority.publish_release("app.example.com", {"version": "1.1.0", "cid": "bafkreibbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "contentRoot": "b" * 64, "entrypoint": "index.html", "files": 3, "size": 180})
    assert first["releaseId"].startswith("rcp2-")
    assert first["signature"]
    assert second["parent"] == first["releaseId"]
    assert authority.releases("app.example.com")[0]["version"] == "1.1.0"
