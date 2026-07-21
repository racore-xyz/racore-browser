from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import httpx


class IPFSBridge:
    def __init__(self, api_url: str, gateway_url: str):
        self.api_url = api_url.rstrip("/")
        self.gateway_url = gateway_url.rstrip("/")
        self.client = httpx.AsyncClient(timeout=120)

    async def status(self) -> dict[str, Any]:
        try:
            response = await self.client.post(f"{self.api_url}/api/v0/id", timeout=3)
            response.raise_for_status()
            data = response.json()
            return {"online": True, "peerId": data.get("ID"), "agentVersion": data.get("AgentVersion"), "addresses": data.get("Addresses", [])}
        except Exception as exc:
            return {"online": False, "error": str(exc), "api": self.api_url}

    async def add_bytes(self, name: str, content: bytes, pin: bool = True) -> dict[str, Any]:
        response = await self.client.post(f"{self.api_url}/api/v0/add", params={"pin": str(pin).lower(), "cid-version": "1", "quieter": "true"}, files={"file": (name, content)})
        response.raise_for_status()
        data = response.json()
        return {"cid": data["Hash"], "name": data.get("Name", name), "size": int(data.get("Size", len(content))), "gateway": f"{self.gateway_url}/ipfs/{data['Hash']}"}

    async def cat(self, cid: str) -> bytes:
        response = await self.client.post(f"{self.api_url}/api/v0/cat", params={"arg": cid})
        response.raise_for_status()
        return response.content

    async def pin(self, cid: str) -> dict[str, Any]:
        response = await self.client.post(f"{self.api_url}/api/v0/pin/add", params={"arg": cid})
        response.raise_for_status()
        return response.json()

    async def publish_ipns(self, cid: str, key: str = "self") -> dict[str, Any]:
        response = await self.client.post(f"{self.api_url}/api/v0/name/publish", params={"arg": f"/ipfs/{cid}", "key": key, "allow-offline": "true"})
        response.raise_for_status()
        return response.json()
