from __future__ import annotations

import asyncio
import os
from pathlib import Path

from racored.config import Settings
from racored.mesh import MeshNode


async def run() -> None:
    root = Path(os.getenv("RACORE_SMOKE_DIR", "work/mesh-smoke")).resolve()
    settings_a = Settings(node_name="Racore Alpha", port=47901, mesh_port=47779, mesh_heartbeat_seconds=1)
    settings_b = Settings(node_name="Racore Beta", port=47902, mesh_port=47779, mesh_heartbeat_seconds=1)

    os.environ["RACORE_DATA_DIR"] = str(root / "alpha")
    alpha = MeshNode(settings_a)
    os.environ["RACORE_DATA_DIR"] = str(root / "beta")
    beta = MeshNode(settings_b)

    await alpha.start()
    await beta.start()
    try:
        await alpha.broadcast("connector.test", {"source": "mesh-smoke"})
        await beta.broadcast("connector.test", {"source": "mesh-smoke"})
        for _ in range(12):
            if alpha.peers and beta.peers:
                break
            await asyncio.sleep(0.5)
        if not alpha.peers or not beta.peers:
            raise RuntimeError(f"mesh discovery failed: alpha={len(alpha.peers)} beta={len(beta.peers)}")
        print(f"RMP mesh OK: {alpha.identity.node_id} <-> {beta.identity.node_id}")
    finally:
        await alpha.stop()
        await beta.stop()


if __name__ == "__main__":
    asyncio.run(run())
