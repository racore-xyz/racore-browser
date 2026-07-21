"use client";

import Image from "next/image";
import { useEffect, useState } from "react";
import { AgenticBrowserView } from "./components/AgenticBrowserView";
import { LiveNetworkView } from "./components/LiveNetworkView";
import { Onboarding } from "./components/Onboarding";
import { ProvidersView } from "./components/ProvidersView";
import { SitesView } from "./components/SitesView";
import { daemonRequest } from "./lib/racore-client";

type View = "browser" | "sites" | "providers" | "network" | "system";
const navigation: { id: View; icon: string; label: string }[] = [
  { id: "browser", icon: "⌕", label: "Browser" },
  { id: "sites", icon: "◇", label: "Sites" },
  { id: "providers", icon: "✦", label: "AI" },
  { id: "network", icon: "◎", label: "Network" },
  { id: "system", icon: "⚙", label: "System" },
];

function SystemView() {
  const [health, setHealth] = useState<Record<string, unknown> | null>(null);
  const [platform, setPlatform] = useState<{
    packaged: boolean;
    version: string;
  } | null>(null);
  useEffect(() => {
    const timer = setTimeout(async () => {
      try {
        setHealth(await daemonRequest<Record<string, unknown>>("/health"));
      } catch {
        setHealth(null);
      }
      if (window.racoreDesktop)
        setPlatform(await window.racoreDesktop.platform());
    }, 0);
    return () => clearTimeout(timer);
  }, []);
  const mesh = health?.mesh as { online?: boolean } | undefined;
  const ipfs = health?.ipfs as
    | { online?: boolean; agentVersion?: string }
    | undefined;
  return (
    <div className="screen system-real">
      <div className="screen-head">
        <div>
          <h1>System</h1>
          <p>Only components detected on this device are shown as ready.</p>
        </div>
      </div>
      <div className="component-stack">
        <article>
          <i className="ready">✓</i>
          <div>
            <b>Racore Browser</b>
            <p>Chromium desktop shell and local interface</p>
          </div>
          <span>
            {platform?.packaged
              ? `Installed · v${platform.version}`
              : "Web preview"}
          </span>
        </article>
        <article>
          <i className={health ? "ready" : "waiting"}>{health ? "✓" : "○"}</i>
          <div>
            <b>Python agent service</b>
            <p>
              Provider gateway, encrypted vault, authority, approvals, and mesh
            </p>
          </div>
          <span>{health ? "Running on 127.0.0.1:47831" : "Not detected"}</span>
        </article>
        <article>
          <i className={ipfs?.online ? "ready" : "waiting"}>
            {ipfs?.online ? "✓" : "○"}
          </i>
          <div>
            <b>IPFS Kubo</b>
            <p>Bundled content-addressed storage node</p>
          </div>
          <span>{ipfs?.online ? ipfs.agentVersion : "Not detected"}</span>
        </article>
        <article>
          <i className={platform?.packaged ? "ready" : "waiting"}>
            {platform?.packaged ? "✓" : "○"}
          </i>
          <div>
            <b>Racore CLI</b>
            <p>Framework build publishing tool included with the desktop app</p>
          </div>
          <span>
            {platform?.packaged
              ? "resources/racored/racore.exe"
              : "Available in desktop package"}
          </span>
        </article>
        <article>
          <i className={mesh?.online ? "ready" : "waiting"}>
            {mesh?.online ? "✓" : "○"}
          </i>
          <div>
            <b>Racoon Mesh</b>
            <p>Signed RMP discovery and live connector events</p>
          </div>
          <span>{mesh?.online ? "Online" : "Not detected"}</span>
        </article>
      </div>
      <section className="safety-card">
        <span>HUMAN-IN-THE-LOOP</span>
        <h2>External side effects require explicit approval.</h2>
        <p>
          Racore uses the user&apos;s configured network path and respects site
          policies. It does not bypass CAPTCHAs, access controls, rate limits,
          or bans.
        </p>
      </section>
    </div>
  );
}

export default function Home() {
  const [view, setView] = useState<View>("browser");
  const [onboarded, setOnboarded] = useState<boolean | null>(null);
  useEffect(() => {
    const initialize = setTimeout(() => {
      const force =
        new URLSearchParams(window.location.search).get("onboarding") === "1";
      setOnboarded(!force && Boolean(localStorage.getItem("racore:onboarded")));
    }, 0);
    const openProviders = () => setView("providers");
    window.addEventListener("racore:open-providers", openProviders);
    return () => {
      clearTimeout(initialize);
      window.removeEventListener("racore:open-providers", openProviders);
    };
  }, []);
  if (onboarded === null)
    return (
      <main className="racore-loading">
        <span className="brand-mark">
          <i />
          <i />
          <i />
          <i />
        </span>
      </main>
    );
  if (!onboarded)
    return (
      <Onboarding
        onFinish={() => {
          localStorage.setItem("racore:onboarded", "1");
          setOnboarded(true);
        }}
      />
    );
  return (
    <main className="browser-app">
      <aside className="utility-rail">
        <div className="rail-logo">
          <Image
            src="/brand/racore-logo.png"
            alt="Racore"
            width={170}
            height={43}
          />
        </div>
        <nav>
          {navigation.map((item) => (
            <button
              key={item.id}
              className={view === item.id ? "active" : ""}
              onClick={() => setView(item.id)}
              title={item.label}
            >
              <i>{item.icon}</i>
              <span>{item.label}</span>
            </button>
          ))}
        </nav>
        <button
          className="reopen-onboarding"
          onClick={() => setOnboarded(false)}
          title="Open onboarding"
        >
          ?
        </button>
      </aside>
      <section className="workspace-shell">
        {view === "browser" && <AgenticBrowserView />}
        {view === "sites" && <SitesView />}
        {view === "providers" && <ProvidersView />}
        {view === "network" && <LiveNetworkView />}
        {view === "system" && <SystemView />}
      </section>
    </main>
  );
}
