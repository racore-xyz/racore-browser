from __future__ import annotations

import base64
import json
import os
from pathlib import Path

from cryptography.fernet import Fernet, InvalidToken

from .config import data_directory


class CredentialVault:
    """Machine-local encrypted provider credential storage.

    The encryption key is placed in the OS credential store when available. A
    permission-restricted local fallback is used for development environments.
    """

    SERVICE = "Racore Browser"
    ACCOUNT = "local-vault-key"

    def __init__(self, root: Path | None = None):
        self.root = root or data_directory()
        self.root.mkdir(parents=True, exist_ok=True)
        self.path = self.root / "credentials.vault"
        self.fernet = Fernet(self._load_or_create_key())

    def _load_or_create_key(self) -> bytes:
        try:
            import keyring

            value = keyring.get_password(self.SERVICE, self.ACCOUNT)
            if value:
                return value.encode()
            key = Fernet.generate_key()
            keyring.set_password(self.SERVICE, self.ACCOUNT, key.decode())
            return key
        except Exception:
            fallback = self.root / ".vault-key"
            if fallback.exists():
                return fallback.read_bytes().strip()
            key = Fernet.generate_key()
            fallback.write_bytes(key)
            try:
                os.chmod(fallback, 0o600)
            except OSError:
                pass
            return key

    def _read(self) -> dict[str, str]:
        if not self.path.exists():
            return {}
        try:
            clear = self.fernet.decrypt(self.path.read_bytes())
            return json.loads(clear.decode("utf-8"))
        except (InvalidToken, json.JSONDecodeError) as exc:
            raise RuntimeError("Racore credential vault cannot be decrypted") from exc

    def _write(self, values: dict[str, str]) -> None:
        self.path.write_bytes(self.fernet.encrypt(json.dumps(values).encode("utf-8")))

    def set(self, provider: str, secret: str) -> None:
        values = self._read()
        values[provider] = secret
        self._write(values)

    def get(self, provider: str) -> str | None:
        return self._read().get(provider)

    def delete(self, provider: str) -> None:
        values = self._read()
        values.pop(provider, None)
        self._write(values)

    def connected(self) -> list[str]:
        return sorted(self._read().keys())

    def masked(self, provider: str) -> str | None:
        secret = self.get(provider)
        if not secret:
            return None
        if len(secret) < 10:
            return "••••••••"
        return f"{secret[:4]}…{secret[-4:]}"
