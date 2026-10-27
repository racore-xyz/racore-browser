"use client";

import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { daemonRequest } from "../lib/racore-client";
import { desktopBridge, isDesktopApp } from "../lib/desktop";
import { racoreDomainFromInput } from "../lib/racore-domains";

type Result = {
  text: string;
  model?: string;
  latencyMs?: number;
  actions?: { type: string; value: string }[];
};
type ChatMessage = {
  role: "user" | "assistant";
  content: string;
  model?: string;
  latencyMs?: number;
};
type BrowserTab = {
  id: string;
  title: string;
  query: string;
  messages: ChatMessage[];
  mediaUrl?: string;
  lastUrl?: string;
};
type PendingAction = {
  id: string;
  type: string;
  value: string;
  tabId: string;
};
type Health = {
  mesh: { online: boolean; peers: number };
  ipfs: { online: boolean };
  localAI: { ready: boolean; label: string };
};
type BrowserMode = "racore" | "web";
type NetworkDomain = {
  domain: string;
  controller: string;
  source: "local" | "mesh";
  nodeId?: string;
  releaseId?: string;
  cid?: string;
  updatedAt: number;
};
type ResolvedDomain = NetworkDomain & {
  gatewayUrl: string;
};

const suggestions = [
  "Research a topic with my local model",
  "Open racore.xyz",
  "Explain this page's privacy risks",
];

const STORAGE_KEY = "racore:browser-tabs:v1";
const MODE_STORAGE_KEY = "racore:browser-mode:v1";
const DIRECT_VIDEO_PATTERN = /\.(mp4|webm|ogv|ogg)(?:[?#].*)?$/i;

function isDirectVideoUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return (
      (url.protocol === "https:" || url.protocol === "http:") &&
      DIRECT_VIDEO_PATTERN.test(value)
    );
  } catch {
    return false;
  }
}

function normalizeDestination(value: string): string {
  return /^https?:\/\//i.test(value) ? value : `https://${value}`;
}

function createTab(): BrowserTab {
  return {
    id:
      typeof crypto !== "undefined" && "randomUUID" in crypto
        ? crypto.randomUUID()
        : `tab-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    title: "New tab",
    query: "",
    messages: [],
  };
}

function actionLabel(action: PendingAction): string {
  if (action.type === "open_url") return `Open ${action.value}`;
  if (action.type === "search_web") return `Search the web for “${action.value}”`;
  if (action.type === "resolve_racore_domain")
    return `Resolve Racore domain ${action.value}`;
  return `${action.type}: ${action.value}`;
}

export function AgenticBrowserView() {
  const [tabs, setTabs] = useState<BrowserTab[]>([
    { id: "initial-tab", title: "New tab", query: "", messages: [] },
  ]);
  const [activeTabId, setActiveTabId] = useState("initial-tab");
  const [health, setHealth] = useState<Health | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [events, setEvents] = useState<string[]>([]);
  const [storageReady, setStorageReady] = useState(false);
  const [mode, setMode] = useState<BrowserMode>("racore");
  const [networkDomains, setNetworkDomains] = useState<NetworkDomain[]>([]);
  const [pending, setPending] = useState<PendingAction[]>([]);
  const [menuOpen, setMenuOpen] = useState(false);
  const [videoKey, setVideoKey] = useState(0);
  const menuRef = useRef<HTMLDivElement>(null);
  const activeTab = tabs.find((tab) => tab.id === activeTabId) ?? tabs[0];
  const query = activeTab?.query ?? "";
  const messages = activeTab?.messages ?? [];
  const mediaUrl = activeTab?.mediaUrl;
  const pendingForTab = pending.filter((item) => item.tabId === activeTabId);
  const isUrl = useMemo(
    () => /^(https?:\/\/)?([\w-]+\.)+[a-z]{2,}(\/.*)?$/i.test(query.trim()),
    [query],
  );

  useEffect(() => {
    const initialize = setTimeout(async () => {
      try {
        const state = await daemonRequest<Health>("/health");
        setHealth(state);
        try {
          setNetworkDomains(
            await daemonRequest<NetworkDomain[]>(
              "/v1/authority/network-domains",
            ),
          );
        } catch {
          setNetworkDomains([]);
        }
      } catch {
        setHealth(null);
      }
    }, 0);
    return () => clearTimeout(initialize);
  }, []);

  useEffect(() => {
    const restore = setTimeout(() => {
      try {
        const saved = localStorage.getItem(STORAGE_KEY);
        const savedMode = localStorage.getItem(MODE_STORAGE_KEY);
        if (savedMode === "racore" || savedMode === "web") {
          setMode(savedMode);
        }
        if (!saved) return;
        const parsed = JSON.parse(saved) as {
          tabs?: BrowserTab[];
          activeTabId?: string;
        };
        const restored = parsed.tabs?.filter(
          (tab) =>
            typeof tab.id === "string" &&
            typeof tab.title === "string" &&
            typeof tab.query === "string" &&
            Array.isArray(tab.messages) &&
            (tab.mediaUrl === undefined || isDirectVideoUrl(tab.mediaUrl)) &&
            (tab.lastUrl === undefined || typeof tab.lastUrl === "string"),
        );
        if (!restored?.length) return;
        setTabs(restored);
        setActiveTabId(
          restored.some((tab) => tab.id === parsed.activeTabId)
            ? parsed.activeTabId!
            : restored[0].id,
        );
      } catch {
        localStorage.removeItem(STORAGE_KEY);
      } finally {
        setStorageReady(true);
      }
    }, 0);
    return () => clearTimeout(restore);
  }, []);

  useEffect(() => {
    if (!storageReady) return;
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({ tabs, activeTabId }),
    );
    localStorage.setItem(MODE_STORAGE_KEY, mode);
  }, [tabs, activeTabId, mode, storageReady]);

  useEffect(() => {
    if (!menuOpen) return;
    const close = (event: MouseEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) setMenuOpen(false);
    };
    window.addEventListener("mousedown", close);
    return () => window.removeEventListener("mousedown", close);
  }, [menuOpen]);

  function updateActiveTab(update: (tab: BrowserTab) => BrowserTab) {
    setTabs((items) =>
      items.map((tab) => (tab.id === activeTabId ? update(tab) : tab)),
    );
  }

  function setQuery(value: string) {
    updateActiveTab((tab) => ({ ...tab, query: value }));
  }

  function addTab() {
    const tab = createTab();
    setTabs((items) => [...items, tab]);
    setActiveTabId(tab.id);
    setError("");
    setEvents([]);
    setMenuOpen(false);
  }

  function closeTab(id: string) {
    setPending((items) => items.filter((item) => item.tabId !== id));
    if (tabs.length === 1) {
      const replacement = createTab();
      setTabs([replacement]);
      setActiveTabId(replacement.id);
    } else {
      const index = tabs.findIndex((tab) => tab.id === id);
      const remaining = tabs.filter((tab) => tab.id !== id);
      setTabs(remaining);
      if (id === activeTabId) {
        setActiveTabId(remaining[Math.max(0, index - 1)].id);
      }
    }
    setError("");
    setEvents([]);
  }

  async function openDestination(destination: string): Promise<void> {
    updateActiveTab((tab) => ({ ...tab, lastUrl: destination }));
    if (isDesktopApp()) {
      await desktopBridge.openBrowser(destination);
      return;
    }
    window.open(destination, "_blank", "noopener,noreferrer");
  }

  async function submit(event?: FormEvent) {
    event?.preventDefault();
    const input = query.trim();
    if (!input) return;
    const racoreDomain = racoreDomainFromInput(input);
    if (racoreDomain) {
      setLoading(true);
      setError("");
      setEvents([`Resolving ${racoreDomain} on the active Racore Mesh`]);
      try {
        const resolved = await daemonRequest<ResolvedDomain>(
          `/v1/authority/resolve/${encodeURIComponent(racoreDomain)}`,
        );
        updateActiveTab((tab) => ({
          ...tab,
          title: racoreDomain,
          query: "",
          messages: [],
          mediaUrl: undefined,
        }));
        setMode("racore");
        setEvents([
          `Resolved ${racoreDomain}`,
          `${resolved.source === "local" ? "Local authority" : "Mesh peer"} · ${resolved.cid}`,
        ]);
        await openDestination(resolved.gatewayUrl);
      } catch (cause) {
        setError(
          cause instanceof Error ? cause.message : "Racore domain resolution failed.",
        );
        setEvents(["Resolution stopped without an unverified fallback"]);
      } finally {
        setLoading(false);
      }
      return;
    }
    if (isUrl) {
      if (mode === "racore") {
        setError(
          "This is a public-web address. Switch to Normal Browsing to open it.",
        );
        return;
      }
      const normalized = normalizeDestination(input);
      if (isDirectVideoUrl(normalized)) {
        const videoUrl = new URL(normalized);
        const filename =
          decodeURIComponent(videoUrl.pathname.split("/").pop() || "") ||
          "Video";
        updateActiveTab((tab) => ({
          ...tab,
          title: filename.slice(0, 42),
          query: "",
          messages: [],
          mediaUrl: normalized,
          lastUrl: normalized,
        }));
        setError("");
        setEvents(["Direct video opened in Racore's private player"]);
        return;
      }
      await openDestination(normalized);
      return;
    }
    if (mode === "web") {
      await openDestination(
        `https://duckduckgo.com/?q=${encodeURIComponent(input)}`,
      );
      return;
    }
    setLoading(true);
    setError("");
    setEvents(["Request accepted by the local Racore daemon"]);
    const requestMessages = [
      ...messages.map(({ role, content }) => ({ role, content })),
      { role: "user" as const, content: input },
    ];
    updateActiveTab((tab) => ({
      ...tab,
      title: tab.messages.length ? tab.title : input.slice(0, 42),
      query: "",
      mediaUrl: undefined,
      messages: [...tab.messages, { role: "user", content: input }],
    }));
    try {
      const response = await daemonRequest<Result>("/v1/chat", {
        method: "POST",
        body: {
          messages: requestMessages,
          system:
            "You are Racore, a concise agentic browser assistant. Never claim a web action occurred unless a tool confirmed it. Ask for approval before external side effects.",
        },
      });
      updateActiveTab((tab) => ({
        ...tab,
        messages: [
          ...tab.messages,
          {
            role: "assistant",
            content: response.text,
            model: response.model || "Hammer 2.0 0.5B",
            latencyMs: response.latencyMs,
          },
        ],
      }));
      setEvents((items) => [
        ...items,
        `Verified local response from ${response.model || "Hammer 2.0 0.5B"}`,
      ]);
      const queued = (response.actions || [])
        .filter((action) => action?.type && action?.value)
        .map((action, index) => ({
          id: `${Date.now()}-${index}-${action.type}`,
          type: action.type,
          value: action.value,
          tabId: activeTabId,
        }));
      if (queued.length) {
        setPending((items) => [...items, ...queued]);
        setEvents((items) => [
          ...items,
          `${queued.length} action${queued.length === 1 ? "" : "s"} waiting for your approval`,
        ]);
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

  async function approve(action: PendingAction) {
    setPending((items) => items.filter((item) => item.id !== action.id));
    try {
      let destination = "";
      if (action.type === "open_url") {
        destination = normalizeDestination(action.value);
      } else if (action.type === "search_web") {
        destination = `https://duckduckgo.com/?q=${encodeURIComponent(action.value)}`;
      } else if (action.type === "resolve_racore_domain") {
        const resolved = await daemonRequest<ResolvedDomain>(
          `/v1/authority/resolve/${encodeURIComponent(action.value)}`,
        );
        destination = resolved.gatewayUrl;
      }
      if (!destination) {
        setError(`Unsupported action type: ${action.type}`);
        return;
      }
      await openDestination(destination);
      setEvents((items) => [...items, `Approved · ${action.type}: ${action.value}`]);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Action failed");
    }
  }

  function deny(action: PendingAction) {
    setPending((items) => items.filter((item) => item.id !== action.id));
    setEvents((items) => [...items, `Denied · ${action.type}: ${action.value}`]);
  }

  async function control(action: "back" | "forward" | "reload") {
    if (action === "reload" && mediaUrl) {
      setVideoKey((value) => value + 1);
      return;
    }
    if (isDesktopApp()) {
      try {
        await desktopBridge.browserAction(action);
        return;
      } catch (cause) {
        setError(
          cause instanceof Error ? cause.message : "No browser window is open.",
        );
        return;
      }
    }
    setError("Browser controls are available in Racore Desktop.");
  }

  async function copyAddress() {
    const address = activeTab?.lastUrl || mediaUrl;
    setMenuOpen(false);
    if (!address) return;
    try {
      await navigator.clipboard.writeText(address);
      setEvents((items) => [...items, `Copied ${address}`]);
    } catch {
      setError("Clipboard access was denied.");
    }
  }

  return (
    <div className="real-browser">
      <div className="browser-mode-switch" aria-label="Browser mode">
        <button
          className={mode === "racore" ? "active" : ""}
          onClick={() => {
            setMode("racore");
            setError("");
          }}
        >
          <span>◆</span>
          <b>Racore Mode</b>
          <small>AI + mesh domains</small>
        </button>
        <button
          className={mode === "web" ? "active" : ""}
          onClick={() => {
            setMode("web");
            setError("");
          }}
        >
          <span>◎</span>
          <b>Normal Browsing</b>
          <small>Public web + search</small>
        </button>
      </div>
      <div className="real-tabs">
        {tabs.map((tab) => (
          <button
            className={tab.id === activeTabId ? "active" : ""}
            key={tab.id}
            onClick={() => {
              setActiveTabId(tab.id);
              setError("");
              setEvents([]);
            }}
          >
            <span>◆</span> {tab.title}
            <i
              role="button"
              aria-label={`Close ${tab.title}`}
              onClick={(event) => {
                event.stopPropagation();
                closeTab(tab.id);
              }}
            >
              ×
            </i>
          </button>
        ))}
        <button aria-label="New tab" onClick={addTab}>＋</button>
        <div />
      </div>
      <div className="real-toolbar">
        <button aria-label="Back" onClick={() => void control("back")}>‹</button>
        <button aria-label="Forward" onClick={() => void control("forward")}>›</button>
        <button aria-label="Reload" onClick={() => void control("reload")}>↻</button>
        <form onSubmit={submit}>
          <span>◆</span>
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={
              mode === "racore"
                ? "Ask Racore or enter a .racore / .rac / .core / .ra domain"
                : "Search the web or enter an address"
            }
          />
          <button aria-label="Go">↗</button>
        </form>
        <div className="toolbar-menu" ref={menuRef}>
          <button
            aria-label="Menu"
            aria-expanded={menuOpen}
            onClick={() => setMenuOpen((value) => !value)}
          >
            ⋮
            {pendingForTab.length > 0 && (
              <em className="toolbar-badge">{pendingForTab.length}</em>
            )}
          </button>
          {menuOpen && (
            <div className="toolbar-popover" role="menu">
              <button role="menuitem" onClick={addTab}>New tab</button>
              <button role="menuitem" onClick={() => closeTab(activeTabId)}>
                Close tab
              </button>
              <button
                role="menuitem"
                onClick={() => {
                  setMenuOpen(false);
                  void control("reload");
                }}
              >
                Reload
              </button>
              <button
                role="menuitem"
                onClick={() => void copyAddress()}
                disabled={!activeTab?.lastUrl && !mediaUrl}
              >
                Copy address
              </button>
              <button
                role="menuitem"
                onClick={async () => {
                  const address = activeTab?.lastUrl || mediaUrl;
                  setMenuOpen(false);
                  if (address && isDesktopApp())
                    await desktopBridge.openExternal(address);
                }}
                disabled={!activeTab?.lastUrl && !mediaUrl}
              >
                Open in system browser
              </button>
            </div>
          )}
        </div>
      </div>
      {pendingForTab.length > 0 && (
        <section className="approval-dock" aria-live="polite">
          {pendingForTab.map((action) => (
            <article className="approval-card" key={action.id}>
              <span className="approval-icon">!</span>
              <div>
                <small>HUMAN APPROVAL REQUIRED</small>
                <b>{actionLabel(action)}</b>
                <p>
                  The local planner requested an external side effect. Nothing
                  opens until you approve it on this device.
                </p>
                <div className="risk">
                  <span>⚠</span>
                  <div>
                    <b>{action.type}</b>
                    <small>Runs through your normal browser profile</small>
                  </div>
                </div>
              </div>
              <footer>
                <button
                  className="secondary"
                  onClick={() => deny(action)}
                >
                  Deny
                </button>
                <button
                  className="primary"
                  onClick={() => void approve(action)}
                >
                  Approve & open
                </button>
              </footer>
            </article>
          ))}
        </section>
      )}
      <main className="real-page">
        {mediaUrl ? (
          <section className="browser-video-page">
            <header>
              <button
                className="back-home"
                onClick={() =>
                  updateActiveTab((tab) => ({
                    ...tab,
                    title: "New tab",
                    mediaUrl: undefined,
                  }))
                }
              >
                ← New task
              </button>
              <div>
                <h1>{activeTab.title}</h1>
                <p>{new URL(mediaUrl).hostname}</p>
              </div>
            </header>
            <div className="browser-video-stage">
              <video
                key={videoKey}
                src={mediaUrl}
                controls
                playsInline
                preload="metadata"
              >
                Your system webview cannot play this video format.
              </video>
            </div>
            <div className="verified-events">
              {events.map((item) => (
                <span key={item}>✓ {item}</span>
              ))}
            </div>
          </section>
        ) : !messages.length && !loading && !error ? (
          <section className="simple-home">
            <div className="terminal-wordmark" aria-label="Racore.xyz" role="img">
              <span>RACORE</span>
              <b>.XYZ</b>
              <i>▮</i>
            </div>
            <h1>Browse. Ask. Act with approval.</h1>
            <p className="browser-mode-copy">
              {mode === "racore"
                ? "Private AI workspace and signed sites discovered across your Racore Mesh."
                : "A familiar browser profile for public websites, searches, media, and logins."}
            </p>
            <form onSubmit={submit} className="simple-command">
              <textarea
                rows={3}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder={
                  mode === "racore"
                    ? "Ask a question or enter a website (.racore / .rac / .core / .ra)…"
                    : "Search or enter a public website…"
                }
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
            {mode === "racore" && (
              <section className="network-domain-shelf">
                <header>
                  <div>
                    <b>Ready on your network</b>
                    <small>
                      .racore · .rac · .core · .ra
                    </small>
                  </div>
                  <span>{networkDomains.filter((item) => item.cid).length} live</span>
                </header>
                <div>
                  {networkDomains.filter((item) => item.cid).length ? (
                    networkDomains
                      .filter((item) => item.cid)
                      .slice(0, 8)
                      .map((item) => (
                        <button
                          key={item.domain}
                          onClick={() => setQuery(item.domain)}
                        >
                          <i>{item.source === "local" ? "◆" : "◎"}</i>
                          <span>
                            <b>{item.domain}</b>
                            <small>
                              {item.source === "local" ? "This device" : "Mesh peer"}
                            </small>
                          </span>
                        </button>
                      ))
                  ) : (
                    <p>
                      No published Racore domains have been discovered yet.
                    </p>
                  )}
                </div>
              </section>
            )}
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
                updateActiveTab((tab) => ({
                  ...tab,
                  title: "New tab",
                  query: "",
                  messages: [],
                  mediaUrl: undefined,
                }));
                setError("");
                setEvents([]);
              }}
            >
              ← New task
            </button>
            <h1>{activeTab.title}</h1>
            {messages.map((message, messageIndex) => (
              <article
                className={`conversation-message ${message.role}`}
                key={`${message.role}-${messageIndex}`}
              >
                <header>
                  <span>{message.role === "assistant" ? "✦" : "◆"}</span>
                  <div>
                    <b>{message.role === "assistant" ? "Racore" : "You"}</b>
                    {message.role === "assistant" && (
                      <small>
                        {message.model || "Hammer 2.0 0.5B"}
                        {message.latencyMs ? ` · ${message.latencyMs}ms` : ""}
                      </small>
                    )}
                  </div>
                </header>
                {message.content
                  .split("\n")
                  .map((line, index) =>
                    line.trim() ? <p key={index}>{line}</p> : null,
                  )}
              </article>
            ))}
            {loading && (
              <article>
                <p className="working">
                  Planning locally on this device…
                </p>
              </article>
            )}
            {error && (
              <article>
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
              </article>
            )}
            <div className="verified-events">
              {events.map((item) => (
                <span key={item}>✓ {item}</span>
              ))}
            </div>
            <form onSubmit={submit} className="simple-command follow-up-command">
              <textarea
                rows={2}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Continue the conversation or enter a website…"
              />
              <footer>
                <span className="local-model-chip">✦ Hammer 0.5B · on device</span>
                <button disabled={loading}>↑</button>
              </footer>
            </form>
          </section>
        )}
      </main>
    </div>
  );
}
