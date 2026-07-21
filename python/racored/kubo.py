from __future__ import annotations

import asyncio
import os
import shutil
from pathlib import Path

import httpx

from .config import data_directory


class KuboManager:
    """Starts the bundled IPFS Kubo node when one is not already available."""

    def __init__(self, api_url: str, gateway_url: str) -> None:
        self.api_url = api_url.rstrip("/")
        self.gateway_url = gateway_url.rstrip("/")
        self.process: asyncio.subprocess.Process | None = None
        self.executable = self._find_executable()
        self.repo = data_directory() / "ipfs"

    @staticmethod
    def _find_executable() -> str | None:
        explicit = os.getenv("RACORE_KUBO_PATH")
        candidates = [
            explicit,
            str(Path(__file__).resolve().parents[2] / "desktop" / "runtime" / "kubo" / "ipfs.exe"),
            shutil.which("ipfs"),
        ]
        return next((item for item in candidates if item and Path(item).exists()), None)

    async def _online(self) -> bool:
        try:
            async with httpx.AsyncClient(timeout=1.2) as client:
                response = await client.post(f"{self.api_url}/api/v0/id")
                return response.is_success
        except httpx.HTTPError:
            return False

    async def start(self) -> dict[str, object]:
        if await self._online():
            return {"managed": False, "online": True, "reason": "external-node"}
        if not self.executable:
            return {"managed": False, "online": False, "reason": "kubo-not-installed"}

        self.repo.mkdir(parents=True, exist_ok=True)
        env = {**os.environ, "IPFS_PATH": str(self.repo)}
        if not (self.repo / "config").exists():
            init = await asyncio.create_subprocess_exec(
                self.executable,
                "init",
                "--profile=lowpower",
                env=env,
                stdout=asyncio.subprocess.DEVNULL,
                stderr=asyncio.subprocess.DEVNULL,
            )
            if await init.wait() != 0:
                return {"managed": False, "online": False, "reason": "kubo-init-failed"}

        gateway_port = self.gateway_url.rsplit(":", 1)[-1]
        configure = await asyncio.create_subprocess_exec(
            self.executable,
            "config",
            "Addresses.Gateway",
            f"/ip4/127.0.0.1/tcp/{gateway_port}",
            env=env,
            stdout=asyncio.subprocess.DEVNULL,
            stderr=asyncio.subprocess.DEVNULL,
        )
        try:
            await asyncio.wait_for(configure.wait(), timeout=15)
        except asyncio.TimeoutError:
            configure.terminate()
            await configure.wait()
            return {"managed": False, "online": False, "reason": "kubo-config-timeout"}

        self.process = await asyncio.create_subprocess_exec(
            self.executable,
            "daemon",
            "--enable-gc",
            env=env,
            stdout=asyncio.subprocess.DEVNULL,
            stderr=asyncio.subprocess.DEVNULL,
        )
        for _ in range(40):
            if await self._online():
                return {"managed": True, "online": True, "pid": self.process.pid}
            if self.process.returncode is not None:
                break
            await asyncio.sleep(0.25)
        return {"managed": True, "online": False, "reason": "kubo-start-timeout"}

    async def stop(self) -> None:
        if self.process and self.process.returncode is None:
            self.process.terminate()
            try:
                await asyncio.wait_for(self.process.wait(), timeout=5)
            except asyncio.TimeoutError:
                self.process.kill()
                await self.process.wait()
