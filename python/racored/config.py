from __future__ import annotations

import json
import os
from dataclasses import dataclass, asdict
from pathlib import Path


def data_directory() -> Path:
    override = os.getenv("RACORE_DATA_DIR")
    if override:
        return Path(override).expanduser().resolve()
    if os.name == "nt":
        return Path(os.getenv("APPDATA", Path.home())) / "Racore"
    return Path(os.getenv("XDG_DATA_HOME", Path.home() / ".local" / "share")) / "racore"


@dataclass
class Settings:
    host: str = "127.0.0.1"
    port: int = 47831
    node_name: str = "Racore Desktop"
    mesh_enabled: bool = True
    mesh_multicast_group: str = "239.255.77.77"
    mesh_port: int = 47777
    mesh_heartbeat_seconds: int = 5
    ipfs_api: str = "http://127.0.0.1:5001"
    ipfs_gateway: str = "http://127.0.0.1:8180"
    bootstrap_peers: list[str] | None = None

    @classmethod
    def load(cls) -> "Settings":
        root = data_directory()
        root.mkdir(parents=True, exist_ok=True)
        path = root / "settings.json"
        if not path.exists():
            instance = cls(bootstrap_peers=[])
            path.write_text(json.dumps(asdict(instance), indent=2), encoding="utf-8")
            return cls._from_environment(instance)
        raw = json.loads(path.read_text(encoding="utf-8"))
        if raw.get("ipfs_gateway") == "http://127.0.0.1:8080":
            raw["ipfs_gateway"] = "http://127.0.0.1:8180"
        return cls._from_environment(cls(**{**asdict(cls(bootstrap_peers=[])), **raw}))

    @staticmethod
    def _from_environment(instance: "Settings") -> "Settings":
        if os.getenv("RACORE_PORT"):
            instance.port = int(os.environ["RACORE_PORT"])
        if os.getenv("RACORE_NODE_NAME"):
            instance.node_name = os.environ["RACORE_NODE_NAME"]
        if os.getenv("RACORE_MESH_PORT"):
            instance.mesh_port = int(os.environ["RACORE_MESH_PORT"])
        return instance

    def save(self) -> None:
        root = data_directory()
        root.mkdir(parents=True, exist_ok=True)
        (root / "settings.json").write_text(json.dumps(asdict(self), indent=2), encoding="utf-8")
