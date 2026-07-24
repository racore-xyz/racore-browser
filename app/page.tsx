"use client";

import { useEffect, useRef, useState } from "react";

const providers = ["OpenAI", "Anthropic", "Gemini", "OpenRouter", "Kimi", "Ollama", "OpenCode", "Claude Code", "Kimi Code"];

const architecture = [
  { n: "01", title: "Agentic browser", body: "A Chromium desktop workspace where AI can research, navigate, and act—while every meaningful side effect waits for you." },
  { n: "02", title: "Local control plane", body: "The Go daemon keeps provider keys, approvals, identity, publishing, and peer connections on your machine." },
  { n: "03", title: "Open web protocol", body: "Signed releases, portable domains, content-addressed storage, and a peer mesh make sites verifiable beyond one platform." },
];

const flow = [
  ["01", "ASK", "Give Racore a goal in natural language."],
  ["02", "PLAN", "The agent chooses tools and the right provider."],
  ["03", "APPROVE", "You confirm actions that change the outside world."],
  ["04", "VERIFY", "Racore returns evidence and signed outputs."],
];

function Arrow() {
  return <span aria-hidden="true">↗</span>;
}

export default function Home() {
  const [menuOpen, setMenuOpen] = useState(false);
  const progressRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let frame = 0;
    const updateProgress = () => {
      frame = 0;
      const max = document.documentElement.scrollHeight - window.innerHeight;
      const value = max > 0 ? window.scrollY / max : 0;
      if (progressRef.current) {
        progressRef.current.style.transform = `scaleX(${value})`;
      }
    };
    const onScroll = () => {
      if (frame === 0) {
        frame = window.requestAnimationFrame(updateProgress);
      }
    };
    updateProgress();
    window.addEventListener("scroll", onScroll, { passive: true });

    const observer = new IntersectionObserver(
      (entries) => entries.forEach((entry) => entry.isIntersecting && entry.target.classList.add("is-visible")),
      { threshold: 0.12 },
    );
    document.querySelectorAll("[data-reveal]").forEach((element) => observer.observe(element));
    return () => {
      window.removeEventListener("scroll", onScroll);
      if (frame !== 0) window.cancelAnimationFrame(frame);
      observer.disconnect();
    };
  }, []);

  return (
    <main
      className="landing"
      onPointerMove={(event) => {
        const x = event.clientX / window.innerWidth - 0.5;
        const y = event.clientY / window.innerHeight - 0.5;
        event.currentTarget.style.setProperty("--pointer-x", `${x * 10}deg`);
        event.currentTarget.style.setProperty("--pointer-y", `${y * -8}deg`);
      }}
    >
      <div className="scroll-progress" ref={progressRef} style={{ transform: "scaleX(0)" }} />
      <header className="landing-nav">
        <a className="landing-logo" href="#top" aria-label="Racore home">
          <img src="/brand/racore-logo.png" alt="Racore.xyz" width="202" height="52" />
        </a>
        <nav className={menuOpen ? "nav-links is-open" : "nav-links"} aria-label="Main navigation">
          <a href="#system" onClick={() => setMenuOpen(false)}>System</a>
          <a href="#protocol" onClick={() => setMenuOpen(false)}>Protocol</a>
          <a href="#providers" onClick={() => setMenuOpen(false)}>Providers</a>
          <a href="#safety" onClick={() => setMenuOpen(false)}>Safety</a>
        </nav>
        <a className="nav-launch" href="/browser">Launch browser <Arrow /></a>
        <button className="menu-toggle" onClick={() => setMenuOpen((open) => !open)} aria-expanded={menuOpen} aria-label="Toggle navigation">
          <span /><span />
        </button>
      </header>

      <section className="hero" id="top">
        <div className="hero-meta"><span>RACORE / 0.1</span><span>AGENTIC BROWSER + OPEN WEB PROTOCOL</span></div>
        <div className="hero-visual" aria-hidden="true">
          <div className="hero-orbit"><i /><i /><i /></div>
          <img className="hero-core-image" src="/generated/racore-core.png" alt="" />
          <span className="hero-image-label">RACORE INTELLIGENCE CORE / LIVE</span>
        </div>
        <div className="hero-copy" data-reveal>
          <p className="eyebrow">Browse with agents. Publish without lock-in.</p>
          <h1>The browser<br />built for <em>agency.</em></h1>
          <div className="hero-bottom">
            <p>Racore brings AI, identity, publishing, and a peer-to-peer web into one local-first workspace—without surrendering control.</p>
            <div className="hero-actions">
              <a className="button button-dark" href="/browser">Enter Racore <Arrow /></a>
              <a className="text-link" href="#system">Explore the system <span>↓</span></a>
            </div>
          </div>
        </div>
        <div className="hero-status"><span>SCROLL TO DISCOVER</span><span className="status-live"><i /> LOCAL-FIRST / ONLINE</span></div>
      </section>

      <section className="statement section-grid" id="system">
        <div className="section-kicker"><span>01</span><p>THE RACORE SYSTEM</p></div>
        <div className="statement-copy" data-reveal>
          <p className="eyebrow">One interface. Three layers.</p>
          <h2>From browsing the web<br />to <em>owning</em> the outcome.</h2>
        </div>
        <div className="architecture-grid">
          {architecture.map((item) => (
            <article key={item.n} data-reveal>
              <span>{item.n}</span>
              <div className="architecture-icon" aria-hidden="true"><i /><i /></div>
              <h3>{item.title}</h3>
              <p>{item.body}</p>
              <a href={item.n === "01" ? "/browser" : "#protocol"} aria-label={`Learn about ${item.title}`}><Arrow /></a>
            </article>
          ))}
        </div>
      </section>

      <section className="network-section" id="protocol">
        <div className="network-top"><span>02 / OPEN WEB INFRASTRUCTURE</span><span>(LIVE SYSTEM MODEL)</span></div>
        <div className="network-title" data-reveal>
          <p className="eyebrow">Your work should outlive a platform</p>
          <h2>A web with<br /><em>memory & proof.</em></h2>
        </div>
        <div className="network-stage" aria-label="Racore decentralized network illustration">
          <div className="network-rings"><i /><i /><i /></div>
          <div className="network-core-wrap"><img className="network-core-image" src="/generated/racore-core.png" alt="" /></div>
          <span className="node node-a">SIGNED RELEASES</span>
          <span className="node node-b">IPFS / KUBO</span>
          <span className="node node-c">DOMAIN AUTHORITY</span>
          <span className="node node-d">RACOON MESH</span>
        </div>
        <div className="network-foot">
          <p>Every release can be content-addressed, signed by its owner, and shared across a discovery mesh. Verification travels with the work.</p>
          <a className="button button-light" href="#architecture">See architecture <Arrow /></a>
        </div>
      </section>

      <section className="metrics section-grid" id="safety">
        <div className="section-kicker"><span>03</span><p>CONTROL BY DESIGN</p></div>
        <div className="metrics-heading" data-reveal><p className="eyebrow">Automation with boundaries</p><h2>AI that knows<br />when to <em>ask.</em></h2></div>
        <div className="metric-grid">
          <article data-reveal><span>01</span><strong>100<sup>%</sup></strong><h3>Explicit approval</h3><p>External side effects stop at a human checkpoint before execution.</p></article>
          <article data-reveal><span>02</span><strong>127.0.0.1</strong><h3>Local control plane</h3><p>The daemon exposes its API on the loopback interface, close to your data.</p></article>
          <article data-reveal><span>03</span><strong>AES<sup>256</sup></strong><h3>Encrypted secrets</h3><p>Provider credentials are protected at rest in the local vault.</p></article>
          <article data-reveal><span>04</span><strong>0</strong><h3>Policy bypasses</h3><p>No CAPTCHA evasion, access-control bypass, rate-limit avoidance, or ban circumvention.</p></article>
        </div>
      </section>

      <section className="providers section-grid" id="providers">
        <div className="section-kicker"><span>04</span><p>PROVIDER FREEDOM</p></div>
        <div className="provider-heading" data-reveal>
          <p className="eyebrow">Bring the intelligence you trust</p>
          <h2>Nine providers.<br />One <em>gateway.</em></h2>
          <p>Connect cloud models, local runtimes, or coding agents. Racore routes requests through a consistent interface while keys remain under your control.</p>
        </div>
        <div className="provider-stack" aria-label="Supported AI providers">
          {providers.map((provider, index) => <div key={provider} style={{ "--i": index } as React.CSSProperties}><span>{String(index + 1).padStart(2, "0")}</span><b>{provider}</b><i>{index < 5 ? "CLOUD" : "LOCAL"}</i></div>)}
        </div>
      </section>

      <section className="workflow section-grid">
        <div className="section-kicker"><span>05</span><p>HOW IT MOVES</p></div>
        <div className="workflow-title" data-reveal><p className="eyebrow">A visible chain of intent</p><h2>From a goal<br />to a <em>verified result.</em></h2></div>
        <div className="flow-grid">
          {flow.map(([n, title, body]) => <article key={n} data-reveal><span>{n}</span><div className="flow-glyph">{title.slice(0, 1)}</div><h3>{title}</h3><p>{body}</p></article>)}
        </div>
      </section>

      <section className="architecture-map" id="architecture">
        <div className="map-header"><span>06 / REPOSITORY ARCHITECTURE</span><span>WINDOWS · WEB · PEER NETWORK</span></div>
        <div className="map-title" data-reveal><p className="eyebrow">Built as interoperable layers</p><h2>One product.<br />Clear <em>responsibilities.</em></h2></div>
        <div className="system-map" data-reveal>
          <div className="map-column"><span>INTERFACE</span><article><b>React + Next.js</b><p>Agentic browser, sites, providers, network, and system views.</p></article><article><b>Electron / Tauri</b><p>Desktop packaging and a secure bridge to the local runtime.</p></article></div>
          <div className="map-spine"><i /><i /><i /><i /><i /></div>
          <div className="map-column map-column-right"><span>CONTROL PLANE</span><article><b>racored / Go</b><p>REST + WebSocket API, approvals, gateway, vault, and orchestration.</p></article><article><b>Protocol + storage</b><p>DID identity, signed messages, domain authority, mesh, and IPFS.</p></article></div>
        </div>
      </section>

      <section className="final-cta">
        <div className="cta-noise" aria-hidden="true" />
        <div data-reveal>
          <p className="eyebrow">The web can be useful and yours</p>
          <h2>Browse boldly.<br /><em>Keep control.</em></h2>
          <a className="button button-light" href="/browser">Launch Racore <Arrow /></a>
        </div>
        <span className="cta-code">R / 2026 / OPEN AGENTIC WEB</span>
      </section>

      <footer className="landing-footer">
        <a className="footer-logo" href="#top"><img src="/brand/racore-logo.png" alt="Racore.xyz" width="190" height="48" /></a>
        <p>Agentic browser &amp; open web protocol.<br />Designed for human agency.</p>
        <div><a href="#system">System</a><a href="#protocol">Protocol</a><a href="#providers">Providers</a><a href="/browser">Launch</a></div>
        <span>© 2026 RACORE.XYZ</span>
      </footer>
    </main>
  );
}
