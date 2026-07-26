"use client";

import Image from "next/image";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { daemonRequest } from "../lib/racore-client";
import { desktopBridge, isDesktopApp } from "../lib/desktop";

type Result = {
  text: string;
  model?: string;
  latencyMs?: number;
  actions?: { type: string; value: string }[];
};
type Health = {
  mesh: { online: boolean; peers: number };
  ipfs: { online: boolean };
};

const suggestions = [
  "Research a topic with my local model",
  "Open racore.xyz",
  "Explain this page's privacy risks",
];

export function AgenticBrowserView() {
  const [query, setQuery] = useState("");
  const [health, setHealth] = useState<Health | null>(null);
  const [result, setResult] = useState<Result | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [events, setEvents] = useState<string[]>([]);
  const isUrl = useMemo(
    () => /^(https?:\/\/)?([\w-]+\.)+[a-z]{2,}(\/.*)?$/i.test(query.trim()),
    [query],
  );

  useEffect(() => {
    const initialize = setTimeout(async () => {
      try {
        const state = await daemonRequest<Health>("/health");
        setHealth(state);
      } catch {
        setHealth(null);
      }
    }, 0);
    return () => clearTimeout(initialize);
  }, []);

  async function submit(event?: FormEvent) {
    event?.preventDefault();
    const input = query.trim();
    if (!input) return;
    if (isUrl) {
      if (isDesktopApp()) await desktopBridge.openBrowser(input);
      else
        window.open(
          /^https?:/.test(input) ? input : `https://${input}`,
          "_blank",
          "noopener,noreferrer",
        );
      return;
    }
    setLoading(true);
    setResult(null);
    setError("");
    setEvents(["Request accepted by the local Racore daemon"]);
    try {
      const response = await daemonRequest<Result>("/v1/chat", {
        method: "POST",
        body: {
          messages: [{ role: "user", content: input }],
          system:
            "You are Racore, a concise agentic browser assistant. Never claim a web action occurred unless a tool confirmed it. Ask for approval before external side effects.",
        },
      });
      setResult(response);
      setEvents((items) => [
        ...items,
        `Verified local response from ${response.model || "Hammer 2.0 0.5B"}`,
      ]);
      for (const action of response.actions || []) {
        const destination =
          action.type === "open_url"
            ? (/^https?:\/\//i.test(action.value) ? action.value : `https://${action.value}`)
            : action.type === "search_web"
              ? `https://duckduckgo.com/?q=${encodeURIComponent(action.value)}`
              : "";
        if (destination) {
          if (isDesktopApp()) await desktopBridge.openBrowser(destination);
          else window.open(destination, "_blank", "noopener,noreferrer");
        }
      }
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "No live AI route is connected.",
      );
      setEvents((items) => [
        ...items,
        "Stopped without fabricating a response",
      ]);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="real-browser">
      <div className="real-tabs">
        <button className="active">
          <span>◆</span> New tab <i>×</i>
        </button>
        <button aria-label="New tab">＋</button>
        <div />
      </div>
      <div className="real-toolbar">
        <button aria-label="Back">‹</button>
        <button aria-label="Forward">›</button>
        <button aria-label="Reload">↻</button>
        <form onSubmit={submit}>
          <span>◆</span>
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search, ask Racore, or enter an address"
          />
          <button aria-label="Go">↗</button>
        </form>
        <button aria-label="Menu">⋮</button>
      </div>
      <main className="real-page">
        {!result && !loading && !error ? (
          <section className="simple-home">
            <Image
              src="/brand/racore-logo.png"
              alt="Racore.xyz"
              width={420}
              height={105}
              priority
            />
            <h1>Browse. Ask. Act with approval.</h1>
            <form onSubmit={submit} className="simple-command">
              <textarea
                rows={3}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Ask a question or enter a website…"
              />
              <footer>
                <span className="local-model-chip">✦ Hammer 0.5B · on device</span>
                <button>↑</button>
              </footer>
            </form>
            <div className="simple-suggestions">
              {suggestions.map((item) => (
                <button key={item} onClick={() => setQuery(item)}>
                  {item}
                </button>
              ))}
            </div>
            <div className="live-core-status">
              <span className={health?.mesh.online ? "up" : "down"}>
                ● Mesh{" "}
                {health?.mesh.online
                  ? `online · ${health.mesh.peers} peers`
                  : "offline"}
              </span>
              <span className={health?.ipfs.online ? "up" : "down"}>
                ● IPFS {health?.ipfs.online ? "ready" : "offline"}
              </span>
              <span>Human approval for external actions</span>
            </div>
          </section>
        ) : (
          <section className="simple-answer">
            <button
              className="back-home"
              onClick={() => {
                setResult(null);
                setError("");
                setEvents([]);
              }}
            >
              ← New task
            </button>
            <h1>{query}</h1>
            <article>
              <header>
                <span>✦</span>
                <div>
                  <b>Racore</b>
                  <small>
                    {result?.model || "Hammer 2.0 0.5B"}
                    {result?.latencyMs ? ` · ${result.latencyMs}ms` : ""}
                  </small>
                </div>
              </header>
              {loading && (
                <p className="working">
                  Planning locally on this device…
                </p>
              )}
              {result?.text
                .split("\n")
                .map((line, index) =>
                  line.trim() ? <p key={index}>{line}</p> : null,
                )}
              {error && (
                <div className="real-error">
                  <b>No response was fabricated.</b>
                  <p>{error}</p>
                  <button
                    onClick={() =>
                      window.dispatchEvent(
                        new CustomEvent("racore:open-local-ai"),
                      )
                    }
                  >
                    View local AI status
                  </button>
                </div>
              )}
            </article>
            <div className="verified-events">
              {events.map((item) => (
                <span key={item}>✓ {item}</span>
              ))}
            </div>
          </section>
        )}
      </main>
    </div>
  );
}
