"use client";

import { useEffect, useState } from "react";
import {
  connectProvider,
  listProviders,
  ProviderInfo,
  daemonRequest,
} from "../lib/racore-client";

const accents: Record<string, string> = {
  openai: "OA",
  anthropic: "AI",
  gemini: "G",
  openrouter: "OR",
  kimi: "K",
  ollama: "OL",
  opencode: "OC",
  "claude-code": "CC",
  "kimi-code": "KC",
};

export function ProvidersView() {
  const [providers, setProviders] = useState<ProviderInfo[]>([]);
  const [selected, setSelected] = useState<ProviderInfo | null>(null);
  const [key, setKey] = useState("");
  const [message, setMessage] = useState("Loading local gateway…");
  const [busy, setBusy] = useState(false);
  async function refresh() {
    try {
      setProviders(await listProviders());
      setMessage("");
    } catch {
      setMessage(
        "Open Racore Desktop or start racored to manage live providers.",
      );
    }
  }
  useEffect(() => {
    const initial = setTimeout(() => void refresh(), 0);
    return () => clearTimeout(initial);
  }, []);
  async function connect() {
    if (!selected || !key.trim()) return;
    setBusy(true);
    try {
      await connectProvider(selected.id, key);
      setSelected(null);
      setKey("");
      await refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Connection failed");
    } finally {
      setBusy(false);
    }
  }
  async function disconnect(id: string) {
    await daemonRequest(`/v1/providers/${id}`, { method: "DELETE" });
    await refresh();
  }
  return (
    <div className="screen providers-screen">
      <div className="screen-head">
        <div>
          <h1>AI providers</h1>
          <p>One gateway for local, free, paid, and routed intelligence.</p>
        </div>
        <div>
          <button className="secondary" onClick={refresh}>
            ↻ Refresh health
          </button>
        </div>
      </div>
      <div className="provider-summary">
        <article>
          <span>CONNECTED ROUTES</span>
          <strong>{providers.filter((p) => p.connected).length}</strong>
          <small>Ready on this device</small>
        </article>
        <article>
          <span>DEFAULT POLICY</span>
          <strong>Hybrid</strong>
          <small>Private tasks stay local</small>
        </article>
        <article>
          <span>MONTHLY BUDGET</span>
          <strong>$50</strong>
          <small>Hard stop enabled</small>
        </article>
        <article>
          <span>FALLBACK</span>
          <strong>On</strong>
          <small>Explain every route</small>
        </article>
      </div>
      {message && (
        <div className="provider-message">
          <span>◇</span>
          {message}
        </div>
      )}
      <div className="provider-grid">
        {(providers.length
          ? providers
          : Object.keys(accents).map((id) => ({
              id,
              name: id,
              kind: "",
              default_model: "",
              free: false,
              local: false,
              connected: false,
            }))
        ).map((provider) => (
          <article key={provider.id}>
            <header>
              <span>{accents[provider.id] || provider.name.slice(0, 2)}</span>
              <div>
                <b>{provider.name}</b>
                <small>{provider.local ? "LOCAL / CLI" : "CLOUD API"}</small>
              </div>
              <em className={provider.connected ? "connected" : ""}>
                ● {provider.connected ? "CONNECTED" : "OFFLINE"}
              </em>
            </header>
            <div className="provider-model">
              <span>Default model</span>
              <code>{provider.default_model || "Detect after connection"}</code>
            </div>
            <div className="provider-tags">
              <span>{provider.free ? "Free route" : "Usage billed"}</span>
              <span>{provider.local ? "Device private" : "Encrypted key"}</span>
              <span>Tool capable</span>
            </div>
            <footer>
              {provider.connected && !provider.local ? (
                <button
                  className="secondary"
                  onClick={() => disconnect(provider.id)}
                >
                  Disconnect
                </button>
              ) : (
                <span />
              )}
              <button
                className="primary"
                onClick={() =>
                  provider.local
                    ? daemonRequest(`/v1/providers/${provider.id}/health`).then(
                        () => refresh(),
                      )
                    : setSelected(provider)
                }
              >
                {provider.connected ? "Test connection" : "Connect"}
              </button>
            </footer>
          </article>
        ))}
      </div>
      <section className="routing-policy">
        <div>
          <span>⑂</span>
          <div>
            <b>Racore Smart Route</b>
            <p>
              Choose the best available model by privacy, tool support, latency,
              and budget—then show the decision in the agent trace.
            </p>
          </div>
        </div>
        <button className="toggle on">
          <i />
        </button>
      </section>
      {selected && (
        <div className="provider-modal">
          <div>
            <header>
              <div>
                <span>{accents[selected.id]}</span>
                <div>
                  <b>Connect {selected.name}</b>
                  <small>Stored only in your encrypted local vault</small>
                </div>
              </div>
              <button onClick={() => setSelected(null)}>×</button>
            </header>
            <label>
              API key
              <input
                type="password"
                autoFocus
                value={key}
                onChange={(event) => setKey(event.target.value)}
                placeholder="Paste your provider API key"
              />
            </label>
            <div className="provider-security">
              <span>◇</span>
              <p>
                The browser sends this directly to the local Python daemon. It
                is encrypted before being written to disk.
              </p>
            </div>
            <footer>
              <button className="secondary" onClick={() => setSelected(null)}>
                Cancel
              </button>
              <button
                className="primary"
                disabled={busy || !key.trim()}
                onClick={connect}
              >
                {busy ? "Connecting…" : "Connect securely"}
              </button>
            </footer>
          </div>
        </div>
      )}
    </div>
  );
}
