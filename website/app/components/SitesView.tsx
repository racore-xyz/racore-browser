"use client";

import { useEffect, useState } from "react";
import { daemonRequest } from "../lib/racore-client";

type Release = {
  version: string;
  cid: string;
  releaseId: string;
  createdAt: number;
  files: number;
  size: number;
};
type Domain = {
  domain: string;
  controller: string;
  createdAt: number;
  status: string;
  current: string | null;
  releases: Release[];
};

export function SitesView() {
  const [domains, setDomains] = useState<Domain[]>([]);
  const [domain, setDomain] = useState("");
  const [message, setMessage] = useState("Connecting to the local authority…");
  const [busy, setBusy] = useState(false);

  async function refresh() {
    try {
      setDomains(await daemonRequest<Domain[]>("/v1/authority/domains"));
      setMessage("");
    } catch {
      setMessage("Open Racore Desktop to manage real domains and releases.");
    }
  }
  useEffect(() => {
    const timer = setTimeout(() => void refresh(), 0);
    return () => clearTimeout(timer);
  }, []);

  async function claim() {
    const normalized = domain.toLowerCase().trim();
    if (!normalized) return;
    setBusy(true);
    try {
      const availability = await daemonRequest<{ available: boolean }>(
        `/v1/authority/domains/${normalized}/available`,
      );
      if (!availability.available)
        throw new Error(
          "This domain is already claimed on the active known mesh.",
        );
      await daemonRequest("/v1/authority/domains", {
        method: "POST",
        body: { domain: normalized },
      });
      setDomain("");
      setMessage(`Claimed ${normalized} with a new Ed25519 authority.`);
      await refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Claim failed");
    } finally {
      setBusy(false);
    }
  }

  const command =
    "racore publish --domain app.example.com --build ./dist --version 1.0.0";
  return (
    <div className="screen sites-real">
      <div className="screen-head">
        <div>
          <h1>Sites</h1>
          <p>Claim a domain and publish signed framework builds to IPFS.</p>
        </div>
        <button className="secondary" onClick={refresh}>
          ↻ Refresh
        </button>
      </div>
      {message && (
        <div className="provider-message">
          <span>◆</span>
          {message}
        </div>
      )}
      <section className="domain-claim-card">
        <div>
          <span>DOMAIN AUTHORITY</span>
          <h2>Claim an unused domain</h2>
          <p>
            Availability is checked against this device and claims observed from
            active Racore Mesh peers.
          </p>
        </div>
        <form
          onSubmit={(event) => {
            event.preventDefault();
            void claim();
          }}
        >
          <input
            value={domain}
            onChange={(event) => setDomain(event.target.value)}
            placeholder="app.example.com"
          />
          <button disabled={busy}>
            {busy ? "Checking…" : "Check & claim"}
          </button>
        </form>
      </section>
      <section className="cli-card">
        <header>
          <div>
            <span>INCLUDED COMMAND LINE</span>
            <h2>Publish any static JavaScript build</h2>
          </div>
          <b>racore.exe</b>
        </header>
        <code>{command}</code>
        <p>
          Works with Vite, React, Vue, Svelte, Angular, Next static export, or
          any build folder containing index.html. The CLI hashes every file,
          creates a compressed bundle, uploads it to the bundled Kubo node, and
          asks the authority to sign the release.
        </p>
      </section>
      <div className="real-domain-list">
        {domains.length ? (
          domains.map((item) => (
            <article key={item.domain}>
              <header>
                <div>
                  <i>◆</i>
                  <span>
                    <b>{item.domain}</b>
                    <small>{item.controller.slice(0, 28)}…</small>
                  </span>
                </div>
                <em className="status live">● {item.status}</em>
              </header>
              <div className="domain-facts">
                <span>
                  <small>CURRENT RELEASE</small>
                  <b>
                    {item.current
                      ? item.current.slice(0, 22) + "…"
                      : "No release yet"}
                  </b>
                </span>
                <span>
                  <small>VERSIONS</small>
                  <b>{item.releases.length}</b>
                </span>
                <span>
                  <small>CLAIMED</small>
                  <b>{new Date(item.createdAt * 1000).toLocaleDateString()}</b>
                </span>
              </div>
              {item.releases.length ? (
                <div className="release-real-list">
                  {item.releases.map((release) => (
                    <div key={release.releaseId}>
                      <b>v{release.version}</b>
                      <code>{release.cid}</code>
                      <span>
                        {release.files} files ·{" "}
                        {(release.size / 1024).toFixed(1)} KB
                      </span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="empty-real">
                  Publish your first build with the included CLI. No placeholder
                  releases are shown.
                </p>
              )}
            </article>
          ))
        ) : (
          <div className="empty-state-real">
            <span>◇</span>
            <h2>No domains on this device</h2>
            <p>
              Claim one above. Racore will never invent sample domains or
              releases.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
