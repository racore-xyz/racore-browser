from __future__ import annotations

import asyncio
from contextlib import asynccontextmanager
from typing import Any

from fastapi import FastAPI, File, HTTPException, UploadFile, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

from . import __version__
from .authority import DomainAuthority
from .config import Settings
from .ipfs import IPFSBridge
from .kubo import KuboManager
from .mesh import MeshNode
from .providers import ProviderGateway, PROVIDERS
from .vault import CredentialVault


settings = Settings.load()
vault = CredentialVault()
gateway = ProviderGateway(vault)
ipfs = IPFSBridge(settings.ipfs_api, settings.ipfs_gateway)
kubo = KuboManager(settings.ipfs_api, settings.ipfs_gateway)
authority = DomainAuthority()
event_clients: set[WebSocket] = set()


async def emit(event: dict[str, Any]) -> None:
    if event.get("type") == "mesh.message" and event.get("messageType") == "domain.claim":
        data = event.get("data", {})
        if data.get("domain") and data.get("controller"):
            authority.observe_claim(data["domain"], data["controller"], event.get("nodeId", "unknown"))
    stale = []
    for client in event_clients:
        try:
            await client.send_json(event)
        except Exception:
            stale.append(client)
    for client in stale:
        event_clients.discard(client)


mesh = MeshNode(settings, emit)


@asynccontextmanager
async def lifespan(_: FastAPI):
    await kubo.start()
    await mesh.start()
    yield
    await mesh.stop()
    await kubo.stop()
    await gateway.client.aclose()
    await ipfs.client.aclose()


app = FastAPI(title="Racore Daemon", version=__version__, lifespan=lifespan)
app.add_middleware(CORSMiddleware, allow_origins=["http://localhost:3000", "http://127.0.0.1:3000", "http://[::1]:3000", "http://127.0.0.1:47832", "https://racore.xyz"], allow_origin_regex=r"https://.*\.chatgpt\.site", allow_credentials=False, allow_methods=["*"], allow_headers=["*"])


class ConnectProvider(BaseModel):
    api_key: str = Field(min_length=3, max_length=4096)


class ChatRequest(BaseModel):
    provider: str
    model: str | None = None
    messages: list[dict[str, str]]
    system: str | None = "You are Racore, an accurate browsing and research agent. Cite sources when supplied and never claim an action completed unless a tool confirmed it."


class MeshBroadcast(BaseModel):
    type: str = "connector.event"
    data: dict[str, Any]


class DomainCreate(BaseModel):
    domain: str


class ReleaseCreate(BaseModel):
    version: str = Field(min_length=1, max_length=64)
    cid: str = Field(min_length=10, max_length=256)
    contentRoot: str = Field(min_length=16, max_length=256)
    entrypoint: str = "index.html"
    files: int = Field(ge=1)
    size: int = Field(ge=1)


class DelegateCreate(BaseModel):
    public_key: str
    capabilities: list[str]
    expires_at: int


@app.get("/health")
async def health():
    return {"ok": True, "version": __version__, "mesh": mesh.status(), "ipfs": await ipfs.status(), "providers": len(PROVIDERS)}


@app.get("/v1/providers")
async def providers():
    return gateway.catalog()


@app.put("/v1/providers/{provider_id}/connect")
async def connect_provider(provider_id: str, body: ConnectProvider):
    if provider_id not in PROVIDERS:
        raise HTTPException(404, "Provider not found")
    vault.set(provider_id, body.api_key)
    result = await gateway.health(provider_id)
    await emit({"type": "provider.connected", "provider": provider_id})
    return {"connected": True, "provider": provider_id, "health": result, "maskedKey": vault.masked(provider_id)}


@app.delete("/v1/providers/{provider_id}")
async def disconnect_provider(provider_id: str):
    vault.delete(provider_id)
    return {"connected": False, "provider": provider_id}


@app.get("/v1/providers/{provider_id}/health")
async def provider_health(provider_id: str):
    if provider_id not in PROVIDERS:
        raise HTTPException(404, "Provider not found")
    return await gateway.health(provider_id)


@app.post("/v1/chat")
async def chat(body: ChatRequest):
    try:
        await emit({"type": "agent.started", "provider": body.provider, "model": body.model})
        result = await gateway.chat(body.provider, body.messages, body.model, body.system)
        await emit({"type": "agent.completed", "provider": body.provider, "model": result["model"], "latencyMs": result["latencyMs"]})
        return result
    except PermissionError as exc:
        raise HTTPException(401, str(exc)) from exc
    except Exception as exc:
        raise HTTPException(502, str(exc)) from exc


@app.get("/v1/ipfs/status")
async def ipfs_status():
    return await ipfs.status()


@app.post("/v1/ipfs/add")
async def ipfs_add(file: UploadFile = File(...)):
    try:
        result = await ipfs.add_bytes(file.filename or "upload.bin", await file.read())
        await mesh.broadcast("release.available", {"cid": result["cid"], "name": result["name"]})
        return result
    except Exception as exc:
        raise HTTPException(503, f"Kubo is unavailable: {exc}") from exc


@app.post("/v1/ipfs/pin/{cid}")
async def ipfs_pin(cid: str):
    return await ipfs.pin(cid)


@app.get("/v1/mesh/status")
async def mesh_status():
    return mesh.status()


@app.get("/v1/mesh/peers")
async def mesh_peers():
    return [peer.__dict__ for peer in mesh.peers.values()]


@app.post("/v1/mesh/broadcast")
async def mesh_broadcast(body: MeshBroadcast):
    return await mesh.broadcast(body.type, body.data)


@app.websocket("/v1/events")
async def events(websocket: WebSocket):
    await websocket.accept()
    event_clients.add(websocket)
    await websocket.send_json({"type": "racore.ready", "version": __version__, "mesh": mesh.status()})
    try:
        while True:
            await websocket.receive_text()
    except WebSocketDisconnect:
        event_clients.discard(websocket)


@app.get("/v1/authority/domains")
async def domains():
    return authority.list()


@app.post("/v1/authority/domains")
async def create_domain(body: DomainCreate):
    domain = body.domain.lower().strip(".")
    try:
        record = authority.create(domain)
    except ValueError as exc:
        raise HTTPException(409, str(exc)) from exc
    await mesh.broadcast("domain.claim", {"domain": domain, "controller": record["controller"]})
    return record


@app.get("/v1/authority/domains/{domain}/available")
async def domain_available(domain: str):
    normalized = domain.lower().strip(".")
    return {"domain": normalized, "available": authority.available(normalized), "scope": "active-known-mesh"}


@app.get("/v1/authority/domains/{domain}/releases")
async def domain_releases(domain: str):
    try:
        return authority.releases(domain.lower().strip("."))
    except KeyError as exc:
        raise HTTPException(404, "Domain authority not found") from exc


@app.post("/v1/authority/domains/{domain}/releases")
async def create_release(domain: str, body: ReleaseCreate):
    normalized = domain.lower().strip(".")
    try:
        release = authority.publish_release(normalized, body.model_dump())
    except KeyError as exc:
        raise HTTPException(404, "Claim the domain before publishing") from exc
    await mesh.broadcast("release.available", {"domain": normalized, "releaseId": release["releaseId"], "cid": release["cid"]})
    return release


@app.post("/v1/authority/domains/{domain}/delegate")
async def delegate_domain(domain: str, body: DelegateCreate):
    try:
        return authority.delegate(domain, body.public_key, body.capabilities, body.expires_at)
    except KeyError as exc:
        raise HTTPException(404, "Domain authority not found") from exc


def main() -> None:
    import uvicorn

    uvicorn.run(app, host=settings.host, port=settings.port, reload=False, log_level="info")


if __name__ == "__main__":
    main()
