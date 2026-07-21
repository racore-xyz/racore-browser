"use client";

import { useEffect, useState } from "react";
import { daemonRequest } from "../lib/racore-client";

type MeshStatus = {
  online: boolean;
  nodeId: string;
  nodeName: string;
  peers: number;
  roles: string[];
  uptimeSeconds: number;
  transport: string;
};
type IPFSStatus = {
  online: boolean;
  peerId?: string;
  agentVersion?: string;
  error?: string;
};
type Peer = {
  node_id: string;
  name: string;
  address: string;
  roles: string[];
  last_seen: number;
};

export function LiveNetworkView() {
  const [mesh, setMesh] = useState<MeshStatus | null>(null);
  const [ipfs, setIpfs] = useState<IPFSStatus | null>(null);
  const [peers, setPeers] = useState<Peer[]>([]);
  const [message, setMessage] = useState("Connecting to racored…");
  async function refresh() {
    try {
      const [m, i, p] = await Promise.all([
        daemonRequest<MeshStatus>("/v1/mesh/status"),
        daemonRequest<IPFSStatus>("/v1/ipfs/status"),
        daemonRequest<Peer[]>("/v1/mesh/peers"),
      ]);
      setMesh(m);
      setIpfs(i);
      setPeers(p);
      setMessage("");
    } catch {
      setMessage(
        "The live network requires Racore Desktop or the local Python daemon.",
      );
    }
  }
  useEffect(() => {
    const initial = setTimeout(() => void refresh(), 0);
    const timer = setInterval(refresh, 5000);
    return () => {
      clearTimeout(initial);
      clearInterval(timer);
    };
  }, []);
  async function pulse() {
    try {
      await daemonRequest("/v1/mesh/broadcast", {
        method: "POST",
        body: {
          type: "connector.test",
          data: { message: "Hello from Racore Desktop", at: Date.now() },
        },
      });
      setMessage("Signed test event broadcast to the local mesh.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Broadcast failed");
    }
  }
  return (
    <div className="screen live-network">
      <div className="screen-head">
        <div>
          <h1>Racore Mesh</h1>
          <p>
            Live signed presence, IPFS availability, and real-time connector
            events.
          </p>
        </div>
        <div>
          <button className="secondary" onClick={refresh}>
            ↻ Refresh
          </button>
          <button className="primary" onClick={pulse}>
            ◎ Broadcast test
          </button>
        </div>
      </div>
      {message && (
        <div className="provider-message">
          <span>◎</span>
          {message}
        </div>
      )}
      <div className="provider-summary">
        <article>
          <span>NODE STATUS</span>
          <strong>{mesh?.online ? "Online" : "Offline"}</strong>
          <small>{mesh?.nodeName || "racored unavailable"}</small>
        </article>
        <article>
          <span>DISCOVERED PEERS</span>
          <strong>{mesh?.peers ?? 0}</strong>
          <small>Signed RMP heartbeats</small>
        </article>
        <article>
          <span>IPFS / KUBO</span>
          <strong>{ipfs?.online ? "Ready" : "Not found"}</strong>
          <small>{ipfs?.agentVersion || "Start Kubo to publish"}</small>
        </article>
        <article>
          <span>UPTIME</span>
          <strong>
            {mesh ? `${Math.floor(mesh.uptimeSeconds / 60)}m` : "—"}
          </strong>
          <small>{mesh?.transport || "rmp/0.1"}</small>
        </article>
      </div>
      <div className="network-grid">
        <article className="network-map">
          <header>
            <div>
              <b>Live peer field</b>
              <small>
                {mesh?.nodeId
                  ? `Node ${mesh.nodeId.slice(0, 12)}…`
                  : "Waiting for local identity"}
              </small>
            </div>
            <span>
              <i /> Signed events
            </span>
          </header>
          <div className="map-visual live-map">
            <div className="network-orbit">
              <i />
              <i />
              <i />
              <i />
              <i />
              <span>R</span>
            </div>
            {peers.slice(0, 5).map((peer, index) => (
              <span className={`real-peer rp${index + 1}`} key={peer.node_id}>
                <i />
                {peer.name}
              </span>
            ))}
          </div>
        </article>
        <article className="protocol-health">
          <header>
            <b>Local services</b>
            <span className={`status ${mesh?.online ? "live" : "draft"}`}>
              ● {mesh?.online ? "Operational" : "Unavailable"}
            </span>
          </header>
          {[
            ["Python daemon", Boolean(mesh)],
            ["Signed mesh identity", Boolean(mesh?.nodeId)],
            ["UDP multicast discovery", Boolean(mesh?.online)],
            ["Kubo RPC", Boolean(ipfs?.online)],
          ].map((item) => (
            <div key={String(item[0])}>
              <span>{item[0]}</span>
              <b className={item[1] ? "up" : "warn-text"}>
                {item[1] ? "Ready" : "Waiting"}
              </b>
              <i>
                <em style={{ width: item[1] ? "100%" : "12%" }} />
              </i>
            </div>
          ))}
        </article>
      </div>
      <div className="table-card">
        <div className="table-title">
          <div>
            <b>Discovered nodes</b>
            <small>
              Peers expire automatically when signed heartbeats stop
            </small>
          </div>
          <code>rmp/0.1</code>
        </div>
        <div className="data-table nodes-table">
          <div className="tr head">
            <span>Node</span>
            <span>Address</span>
            <span>Roles</span>
            <span>Last seen</span>
            <span>Status</span>
          </div>
          {peers.length ? (
            peers.map((peer) => (
              <div className="tr" key={peer.node_id}>
                <span>
                  <i className="node-icon">◎</i>
                  <b>{peer.name}</b>
                </span>
                <span>
                  <code>{peer.address}</code>
                </span>
                <span>{peer.roles.join(", ")}</span>
                <span>{new Date(peer.last_seen * 1000).toLocaleTimeString()}</span>
                <span>
                  <em className="status live">● Verified</em>
                </span>
              </div>
            ))
          ) : (
            <div className="no-peers">
              <span>◎</span>
              <div>
                <b>No other Racore nodes discovered yet</b>
                <p>
                  Start racored on a second device on the same LAN to see signed
                  peer presence appear here.
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
