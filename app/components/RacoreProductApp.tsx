"use client";

import { useEffect, useState } from "react";
import { AgenticBrowserView } from "./AgenticBrowserView";
import { LiveNetworkView } from "./LiveNetworkView";
import { Onboarding } from "./Onboarding";
import { ProvidersView } from "./ProvidersView";
import { SitesView } from "./SitesView";
import { daemonRequest } from "../lib/racore-client";
import { desktopBridge, isDesktopApp } from "../lib/desktop";

type View = "browser" | "sites" | "providers" | "network" | "system";
type WorkspaceSpace = {
  id: string;
  label: string;
  icon: string;
};

const navigation: { id: View; icon: string; label: string }[] = [
  { id: "browser", icon: "⌂", label: "Browser" },
  { id: "sites", icon: "◇", label: "Sites" },
  { id: "providers", icon: "✦", label: "AI" },
  { id: "network", icon: "◎", label: "Network" },
  { id: "system", icon: "⚙", label: "System" },
];
const DEFAULT_SPACES: WorkspaceSpace[] = [
  { id: "work", label: "Work", icon: "▣" },
  { id: "play", label: "Play", icon: "◉" },
  { id: "build", label: "Build", icon: "◆" },
];
const SPACES_STORAGE_KEY = "racore:workspace-spaces:v1";

function SystemView() {
  const [health, setHealth] = useState<Record<string, unknown> | null>(null);
  const [platform, setPlatform] = useState<{ packaged: boolean; version: string } | null>(null);

  useEffect(() => {
    const timer = setTimeout(async () => {
      try {
        setHealth(await daemonRequest<Record<string, unknown>>("/health"));
      } catch {
        setHealth(null);
      }
      if (isDesktopApp()) setPlatform(await desktopBridge.platform());
    }, 0);
    return () => clearTimeout(timer);
  }, []);

  const mesh = health?.mesh as { online?: boolean } | undefined;
  const ipfs = health?.ipfs as { online?: boolean; agentVersion?: string } | undefined;

  return (
    <div className="screen system-real">
      <div className="screen-head">
        <div><h1>System</h1><p>Only components detected on this device are shown as ready.</p></div>
      </div>
      <div className="component-stack">
        <article><i className="ready">✓</i><div><b>Racore Browser</b><p>Tauri desktop shell and local React interface</p></div><span>{platform?.packaged ? `Installed · v${platform.version}` : "Web preview"}</span></article>
        <article><i className={health ? "ready" : "waiting"}>{health ? "✓" : "○"}</i><div><b>Go agent service</b><p>Provider gateway, encrypted vault, authority, approvals, and mesh</p></div><span>{health ? "Running on 127.0.0.1:47831" : "Not detected"}</span></article>
        <article><i className={ipfs?.online ? "ready" : "waiting"}>{ipfs?.online ? "✓" : "○"}</i><div><b>IPFS Kubo</b><p>Bundled content-addressed storage node</p></div><span>{ipfs?.online ? ipfs.agentVersion : "Not detected"}</span></article>
        <article><i className={platform?.packaged ? "ready" : "waiting"}>{platform?.packaged ? "✓" : "○"}</i><div><b>Racore CLI</b><p>Framework build publishing tool included with the desktop app</p></div><span>{platform?.packaged ? "Bundled native sidecar" : "Available in desktop package"}</span></article>
        <article><i className={mesh?.online ? "ready" : "waiting"}>{mesh?.online ? "✓" : "○"}</i><div><b>Racoon Mesh</b><p>Signed RMP discovery and live connector events</p></div><span>{mesh?.online ? "Online" : "Not detected"}</span></article>
      </div>
      <section className="safety-card"><span>HUMAN-IN-THE-LOOP</span><h2>External side effects require explicit approval.</h2><p>Racore uses the user&apos;s configured network path and respects site policies. It does not bypass CAPTCHAs, access controls, rate limits, or bans.</p></section>
    </div>
  );
}

export function RacoreProductApp() {
  const [view, setView] = useState<View>("browser");
  const [onboarded, setOnboarded] = useState<boolean | null>(null);
  const [spaces, setSpaces] = useState<WorkspaceSpace[]>(DEFAULT_SPACES);
  const [activeSpaceId, setActiveSpaceId] = useState("work");
  const [spacesReady, setSpacesReady] = useState(false);
  const [clock, setClock] = useState("");

  useEffect(() => {
    const initialize = setTimeout(() => {
      const force = new URLSearchParams(window.location.search).get("onboarding") === "1";
      setOnboarded(!force && Boolean(localStorage.getItem("racore:onboarded")));
      try {
        const saved = JSON.parse(
          localStorage.getItem(SPACES_STORAGE_KEY) || "null",
        ) as { spaces?: WorkspaceSpace[]; activeSpaceId?: string } | null;
        const restored = saved?.spaces?.filter(
          (space) =>
            typeof space.id === "string" &&
            typeof space.label === "string" &&
            typeof space.icon === "string",
        );
        if (restored?.length) {
          setSpaces(restored);
          setActiveSpaceId(
            restored.some((space) => space.id === saved?.activeSpaceId)
              ? saved!.activeSpaceId!
              : restored[0].id,
          );
        }
      } catch {
        localStorage.removeItem(SPACES_STORAGE_KEY);
      } finally {
        setSpacesReady(true);
      }
    }, 0);
    const openProviders = () => setView("providers");
    window.addEventListener("racore:open-providers", openProviders);
    return () => {
      clearTimeout(initialize);
      window.removeEventListener("racore:open-providers", openProviders);
    };
  }, []);

  useEffect(() => {
    if (!spacesReady) return;
    localStorage.setItem(
      SPACES_STORAGE_KEY,
      JSON.stringify({ spaces, activeSpaceId }),
    );
  }, [spaces, activeSpaceId, spacesReady]);

  useEffect(() => {
    function updateClock() {
      setClock(
        new Intl.DateTimeFormat(undefined, {
          hour: "2-digit",
          minute: "2-digit",
        }).format(new Date()),
      );
    }
    const initial = setTimeout(updateClock, 0);
    const timer = setInterval(updateClock, 30_000);
    return () => {
      clearTimeout(initial);
      clearInterval(timer);
    };
  }, []);

  function addSpace() {
    const number = spaces.length + 1;
    const id =
      typeof crypto !== "undefined" && "randomUUID" in crypto
        ? crypto.randomUUID()
        : `space-${Date.now()}`;
    const space = { id, label: `Space ${number}`, icon: "◌" };
    setSpaces((items) => [...items, space]);
    setActiveSpaceId(space.id);
    setView("browser");
  }

  if (onboarded === null) return <main className="racore-loading"><span className="brand-mark"><i /><i /><i /><i /></span></main>;
  if (!onboarded) return <Onboarding onFinish={() => { localStorage.setItem("racore:onboarded", "1"); setOnboarded(true); }} />;

  const activeSpace =
    spaces.find((space) => space.id === activeSpaceId) ?? spaces[0];

  return (
    <main className="browser-app workspace-os">
      <header className="workspace-menubar">
        <div className="workspace-brand">
          <span className="window-lights"><i /><i /><i /></span>
          <b><i>◆</i> RACORE</b>
        </div>
        <nav className="workspace-spaces" aria-label="Spaces">
          {spaces.map((space) => (
            <button
              key={space.id}
              className={space.id === activeSpaceId ? "active" : ""}
              onClick={() => {
                setActiveSpaceId(space.id);
                setView("browser");
              }}
              title={`${space.label} tab cluster`}
            >
              <i>{space.icon}</i>
              <span>{space.label}</span>
              <small>
                {space.id === activeSpaceId ? "Active cluster" : "Space"}
              </small>
            </button>
          ))}
          <button className="add-workspace-space" onClick={addSpace} title="Add space">
            ＋
          </button>
        </nav>
        <div className="workspace-system">
          <span><i /> Mesh</span>
          <b>{clock}</b>
          <button onClick={() => setView("system")} aria-label="Workspace settings">⚙</button>
          <button onClick={() => setOnboarded(false)} aria-label="Open onboarding">?</button>
        </div>
      </header>

      <section className="workspace-stage">
        <div className="workspace-glow glow-one" />
        <div className="workspace-glow glow-two" />
        <section className="workspace-shell">
          <header className="workspace-window-title">
            <div>
              <i>{activeSpace.icon}</i>
              <span>
                <b>{activeSpace.label}</b>
                <small>{navigation.find((item) => item.id === view)?.label}</small>
              </span>
            </div>
            <span>Racore Workspace</span>
          </header>
          <div className="workspace-view">
            {view === "browser" && (
              <AgenticBrowserView
                spaceId={activeSpace.id}
                spaceLabel={activeSpace.label}
              />
            )}
            {view === "sites" && <SitesView />}
            {view === "providers" && <ProvidersView />}
            {view === "network" && <LiveNetworkView />}
            {view === "system" && <SystemView />}
          </div>
        </section>
      </section>

      <nav className="workspace-dock" aria-label="Workspace apps">
        <span className="dock-racore">R</span>
        {navigation.map((item) => (
          <button
            key={item.id}
            className={view === item.id ? "active" : ""}
            onClick={() => setView(item.id)}
            title={item.label}
          >
            <i>{item.icon}</i>
            <small>{item.label}</small>
          </button>
        ))}
        <i className="dock-divider" />
        <button onClick={addSpace} title="New space">
          <i>＋</i>
          <small>Space</small>
        </button>
      </nav>
    </main>
  );
}
