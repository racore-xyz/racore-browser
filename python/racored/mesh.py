from __future__ import annotations

import asyncio
import base64
import hashlib
import json
import socket
import time
from dataclasses import dataclass, asdict
from pathlib import Path
from typing import Any, Awaitable, Callable

from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PrivateKey, Ed25519PublicKey

from .config import Settings, data_directory


def b64(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).decode().rstrip("=")


def unb64(value: str) -> bytes:
    return base64.urlsafe_b64decode(value + "=" * (-len(value) % 4))


@dataclass
class Peer:
    node_id: str
    name: str
    address: str
    public_key: str
    roles: list[str]
    last_seen: float
    latency_ms: int = 0


class NodeIdentity:
    def __init__(self, root: Path | None = None):
        self.root = root or data_directory()
        self.root.mkdir(parents=True, exist_ok=True)
        path = self.root / "mesh-identity.pem"
        if path.exists():
            self.private = serialization.load_pem_private_key(path.read_bytes(), password=None)
        else:
            self.private = Ed25519PrivateKey.generate()
            path.write_bytes(self.private.private_bytes(serialization.Encoding.PEM, serialization.PrivateFormat.PKCS8, serialization.NoEncryption()))
        self.public_bytes = self.private.public_key().public_bytes(serialization.Encoding.Raw, serialization.PublicFormat.Raw)
        self.node_id = hashlib.sha256(self.public_bytes).hexdigest()[:32]

    def sign(self, payload: bytes) -> str:
        return b64(self.private.sign(payload))


class MeshNode:
    def __init__(self, settings: Settings, event_sink: Callable[[dict[str, Any]], Awaitable[None]] | None = None):
        self.settings = settings
        self.identity = NodeIdentity()
        self.peers: dict[str, Peer] = {}
        self.started_at = time.time()
        self.event_sink = event_sink
        self.running = False
        self.transport: asyncio.DatagramTransport | None = None
        self.tasks: list[asyncio.Task] = []

    def status(self) -> dict[str, Any]:
        self._expire_peers()
        return {"online": self.running, "nodeId": self.identity.node_id, "nodeName": self.settings.node_name, "peers": len(self.peers), "roles": ["client", "cache"], "uptimeSeconds": round(time.time() - self.started_at), "transport": "udp-multicast+rmp/0.1"}

    async def start(self) -> None:
        if self.running or not self.settings.mesh_enabled:
            return
        self.running = True
        loop = asyncio.get_running_loop()
        try:
            udp_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM, socket.IPPROTO_UDP)
            udp_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            udp_socket.bind(("", self.settings.mesh_port))
            membership = socket.inet_aton(self.settings.mesh_multicast_group) + socket.inet_aton("0.0.0.0")
            udp_socket.setsockopt(socket.IPPROTO_IP, socket.IP_ADD_MEMBERSHIP, membership)
            udp_socket.setsockopt(socket.IPPROTO_IP, socket.IP_MULTICAST_TTL, 2)
            udp_socket.setblocking(False)
            transport, _ = await loop.create_datagram_endpoint(lambda: MeshDatagram(self), sock=udp_socket)
            self.transport = transport
            self.tasks.append(asyncio.create_task(self._heartbeat_loop()))
        except Exception as exc:
            self.running = False
            await self._emit({"type": "mesh.error", "message": str(exc)})

    async def stop(self) -> None:
        self.running = False
        for task in self.tasks:
            task.cancel()
        self.tasks.clear()
        if self.transport:
            self.transport.close()
            self.transport = None

    def envelope(self, message_type: str, data: dict[str, Any]) -> dict[str, Any]:
        body = {"protocol": "rmp/0.1", "type": message_type, "nodeId": self.identity.node_id, "name": self.settings.node_name, "publicKey": b64(self.identity.public_bytes), "roles": ["client", "cache"], "timestamp": int(time.time() * 1000), "data": data}
        canonical = json.dumps(body, sort_keys=True, separators=(",", ":")).encode()
        return {**body, "signature": self.identity.sign(canonical)}

    async def broadcast(self, message_type: str, data: dict[str, Any]) -> dict[str, Any]:
        message = self.envelope(message_type, data)
        if self.transport:
            self.transport.sendto(json.dumps(message).encode(), (self.settings.mesh_multicast_group, self.settings.mesh_port))
        await self._emit({"type": "mesh.broadcast", "messageType": message_type, "data": data})
        return message

    async def receive(self, raw: bytes, address: tuple[str, int]) -> None:
        try:
            message = json.loads(raw)
            if message.get("nodeId") == self.identity.node_id or message.get("protocol") != "rmp/0.1":
                return
            signature = message.pop("signature")
            canonical = json.dumps(message, sort_keys=True, separators=(",", ":")).encode()
            public_key = Ed25519PublicKey.from_public_bytes(unb64(message["publicKey"]))
            public_key.verify(unb64(signature), canonical)
            expected_id = hashlib.sha256(unb64(message["publicKey"])).hexdigest()[:32]
            if expected_id != message["nodeId"]:
                return
            now = time.time()
            peer = Peer(message["nodeId"], message.get("name", "Racore node"), address[0], message["publicKey"], message.get("roles", []), now)
            is_new = peer.node_id not in self.peers
            self.peers[peer.node_id] = peer
            if is_new:
                await self._emit({"type": "mesh.peer.joined", "peer": asdict(peer)})
            await self._emit({"type": "mesh.message", "messageType": message.get("type"), "nodeId": message["nodeId"], "data": message.get("data", {})})
        except Exception:
            return

    async def _heartbeat_loop(self) -> None:
        while self.running:
            await self.broadcast("presence", {"apiPort": self.settings.port, "ipfs": self.settings.ipfs_api})
            self._expire_peers()
            await asyncio.sleep(self.settings.mesh_heartbeat_seconds)

    def _expire_peers(self) -> None:
        cutoff = time.time() - max(20, self.settings.mesh_heartbeat_seconds * 4)
        for node_id in [node_id for node_id, peer in self.peers.items() if peer.last_seen < cutoff]:
            self.peers.pop(node_id, None)

    async def _emit(self, event: dict[str, Any]) -> None:
        if self.event_sink:
            await self.event_sink({**event, "timestamp": int(time.time() * 1000)})


class MeshDatagram(asyncio.DatagramProtocol):
    def __init__(self, node: MeshNode):
        self.node = node

    def datagram_received(self, data: bytes, addr: tuple[str, int]) -> None:
        asyncio.create_task(self.node.receive(data, addr))
