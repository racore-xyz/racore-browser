"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";

type View = "browser" | "runs" | "publish" | "releases" | "network" | "permissions" | "settings";
type Release = { version: string; cid: string; commit: string; time: string; status: "Live" | "Previous" | "Draft"; hosts: number };

const nav: { id: View; label: string; icon: string }[] = [
  { id: "browser", label: "Browser", icon: "⌘" },
  { id: "runs", label: "Agent runs", icon: "✦" },
  { id: "publish", label: "Publish", icon: "↑" },
  { id: "releases", label: "Releases", icon: "⑂" },
  { id: "network", label: "Network", icon: "◎" },
  { id: "permissions", label: "Permissions", icon: "◇" },
  { id: "settings", label: "Settings", icon: "⚙" },
];

const initialReleases: Release[] = [
  { version: "1.4.0", cid: "rcp1-8f3a…7c19", commit: "8ef75d1", time: "8 min ago", status: "Live", hosts: 3 },
  { version: "1.3.2", cid: "rcp1-12cb…9a42", commit: "2ab410c", time: "2 days ago", status: "Previous", hosts: 3 },
  { version: "1.3.1", cid: "rcp1-f914…63be", commit: "9cc8e21", time: "5 days ago", status: "Previous", hosts: 2 },
];

function BrandMark() {
  return <span className="brand-mark" aria-hidden="true"><i /><i /><i /><i /></span>;
}

function SideNav({ view, setView }: { view: View; setView: (v: View) => void }) {
  return (
    <aside className="side-nav">
      <button className="brand" onClick={() => setView("browser")} aria-label="Racore home"><BrandMark /><b>racore</b><span>.xyz</span></button>
      <div className="workspace-chip"><span>R</span><div><b>Racore Core</b><small>Personal workspace</small></div><em>⌄</em></div>
      <nav aria-label="Primary navigation">
        {nav.map(item => <button key={item.id} className={view === item.id ? "active" : ""} onClick={() => setView(item.id)}><span>{item.icon}</span>{item.label}{item.id === "runs" && <em>2</em>}</button>)}
      </nav>
      <div className="side-status"><div><span className="pulse" /><b>RCP node online</b></div><small>12 peers · Cairo edge</small></div>
      <button className="profile"><span>AK</span><div><b>Ahmed Khalil</b><small>Builder plan</small></div><em>•••</em></button>
    </aside>
  );
}

function TopBar({ title, theme, setTheme }: { title: string; theme: string; setTheme: (v: string) => void }) {
  return <header className="topbar"><div><small>RACORE /</small><b>{title}</b></div><div className="top-actions"><span className="network-pill"><i /> Network healthy</span><button aria-label="Search">⌕</button><button aria-label="Notifications">♢<sup>2</sup></button><button aria-label="Toggle theme" onClick={() => setTheme(theme === "dark" ? "light" : "dark")}>{theme === "dark" ? "☼" : "☾"}</button></div></header>;
}

function BrowserView() {
  const [url, setUrl] = useState("console.racore.xyz");
  const [task, setTask] = useState("Compare decentralized hosting options and build a recommendation");
  const [running, setRunning] = useState(false);
  const [step, setStep] = useState(2);
  const [approval, setApproval] = useState(true);
  const [tab, setTab] = useState("Live");

  function runTask(e: FormEvent) {
    e.preventDefault();
    if (!task.trim()) return;
    setRunning(true); setStep(1); setApproval(false);
    window.setTimeout(() => setStep(2), 900);
    window.setTimeout(() => setApproval(true), 1700);
  }

  return <div className="browser-view">
    <div className="tab-strip"><button className="tab active"><BrandMark /> Racore Console <span>×</span></button><button className="new-tab">＋</button><span className="window-space" /><button>—</button><button>□</button><button>×</button></div>
    <div className="omnibar"><button>‹</button><button>›</button><button>↻</button><form onSubmit={e => { e.preventDefault(); setUrl(url.replace(/^https?:\/\//, "")); }}><span className="verified">◆</span><input value={url} onChange={e => setUrl(e.target.value)} aria-label="Address or command" /><span className="protocol-badge">RCP VERIFIED</span><button type="button">☆</button></form><button>⋮</button></div>
    <div className="browser-content">
      <section className="webpage">
        <header className="site-header"><div className="site-brand"><BrandMark /><b>Racore</b></div><nav><button className="active">Overview</button><button>Domains</button><button>Analytics</button><button>Docs</button></nav><button className="primary small">New project</button></header>
        <div className="page-body">
          <div className="page-title"><div><span className="eyebrow">PRODUCTION</span><h1>Infrastructure overview</h1><p>Verified delivery across every host and edge.</p></div><button className="secondary">Last 24 hours⌄</button></div>
          <div className="metrics"><article><span>REQUESTS</span><strong>24.8M</strong><small className="up">↗ 18.2%</small></article><article><span>VERIFIED LOADS</span><strong>99.998%</strong><small className="up">↗ 0.12%</small></article><article><span>EDGE LATENCY</span><strong>38<small>ms</small></strong><small className="up">↘ 4.8%</small></article><article><span>ACTIVE PEERS</span><strong>1,284</strong><small>+42 today</small></article></div>
          <article className="chart-card"><div className="card-head"><div><b>Verified requests</b><small>Content delivered with a valid release signature</small></div><div className="legend"><i /> HTTPS <i /> RCP peer</div></div><div className="chart"><div className="y-labels"><span>30k</span><span>20k</span><span>10k</span><span>0</span></div><div className="chart-grid"><div /><div /><div /><div /><div className="area-chart" /></div></div><div className="x-labels"><span>00:00</span><span>04:00</span><span>08:00</span><span>12:00</span><span>16:00</span><span>20:00</span><span>Now</span></div></article>
          <div className="bottom-grid"><article className="host-card"><div className="card-head"><div><b>Delivery network</b><small>Traffic by source</small></div><button>View network →</button></div>{[["Racore Edge","54.2%","#d9ff43"],["IPFS peers","28.7%","#8fa8ff"],["HTTPS origin","17.1%","#566172"]].map(x => <div className="host-row" key={x[0]}><span><i style={{background:x[2]}} />{x[0]}</span><div><i style={{width:x[1], background:x[2]}} /></div><b>{x[1]}</b></div>)}</article><article className="release-mini"><div className="card-head"><div><b>Current release</b><small>racore.xyz</small></div><span className="status live">● LIVE</span></div><strong>v1.4.0</strong><code>rcp1-8f3a9d2e…7c19</code><footer><span>3/3 hosts healthy</span><button>Inspect release →</button></footer></article></div>
        </div>
      </section>
      <aside className="agent-panel">
        <header><div><span className="agent-orb">✦</span><div><b>Racore Agent</b><small><i /> Active in this tab</small></div></div><button>⋯</button><button>×</button></header>
        <div className="agent-tabs">{["Live","Plan","Sources"].map(x => <button key={x} className={tab === x ? "active" : ""} onClick={() => setTab(x)}>{x}{x === "Sources" && <em>6</em>}</button>)}</div>
        <div className="agent-scroll">
          {tab === "Live" && <><div className="user-message"><small>YOU · JUST NOW</small><p>{task}</p></div><div className="agent-message"><small>RACORE AGENT</small><p>I’ll compare reliability, portability, verification, and operating cost, then prepare a recommendation you can publish.</p></div>
          <div className="plan-card"><div className={step >= 1 ? "done" : ""}><i>{step > 1 ? "✓" : "1"}</i><span><b>Map requirements</b><small>Availability, ownership, latency, cost</small></span></div><div className={step === 2 ? "current" : step > 2 ? "done" : ""}><i>{step > 2 ? "✓" : "2"}</i><span><b>Compare providers</b><small>{step === 2 ? "Reading 6 sources…" : "IPFS, object storage, edge networks"}</small></span></div><div><i>3</i><span><b>Model deployment options</b><small>Hybrid and peer-first configurations</small></span></div><div><i>4</i><span><b>Write recommendation</b><small>Decision matrix with implementation plan</small></span></div></div>
          <div className="tool-event"><span>⌕</span><div><b>Reading provider documentation</b><small>docs.ipfs.tech · developers.cloudflare.com</small></div><em>•••</em></div>
          {approval && <div className="approval-card"><span className="approval-icon">↑</span><div><small>APPROVAL REQUIRED</small><b>Publish the recommendation?</b><p>The agent will create a draft in <strong>Racore / Research</strong>. Nothing will be made public.</p><div className="risk"><span>◇</span><div><b>Writes 1 workspace document</b><small>Reversible · no external message</small></div></div><footer><button onClick={() => setApproval(false)} className="secondary">Deny</button><button onClick={() => {setApproval(false); setStep(4); setRunning(false)}} className="primary">Approve once</button></footer></div></div>}</>}
          {tab === "Plan" && <div className="empty-state"><span>⑂</span><h3>Four-step plan</h3><p>The agent checks each step against your active permissions before it acts.</p></div>}
          {tab === "Sources" && <div className="source-list">{["IPFS Content Addressing","DNSLink Custom Domains","Sigstore Verification","Cloudflare Edge Network","Racore Protocol 0.1","Deployment Cost Model"].map((s,i)=><article key={s}><span>{i+1}</span><div><b>{s}</b><small>{i < 4 ? "External source" : "Workspace document"}</small></div><em>↗</em></article>)}</div>}
        </div>
        <form className="agent-input" onSubmit={runTask}><textarea value={task} onChange={e => setTask(e.target.value)} aria-label="Ask Racore" rows={2} /><div><span><button type="button">＋</button><button type="button">@</button></span><button className="send" disabled={running} aria-label="Run task">{running ? "■" : "↑"}</button></div></form>
        <footer className="agent-footer"><span><i /> Auto-read</span><span>Approval for external actions</span></footer>
      </aside>
    </div>
  </div>;
}

function Stat({ label, value, note, tone }: { label: string; value: string; note: string; tone?: string }) { return <article className="stat"><span>{label}</span><strong>{value}</strong><small className={tone}>{note}</small></article> }

function RunsView() {
  const runs = [
    ["Compare decentralized hosting options", "Running", "6 sources · 4m 18s", "Browser, workspace.read"],
    ["Verify production release and mirrors", "Waiting approval", "3 hosts · 42s", "network.read, site.verify"],
    ["Summarize protocol security review", "Completed", "12 sources · 8m 04s", "browser.read"],
    ["Prepare weekly infrastructure report", "Completed", "4 documents · 2m 31s", "workspace.write"],
  ];
  return <Screen title="Agent runs" subtitle="Monitor, approve, and audit delegated work." actions={<button className="primary">✦ New agent run</button>}><div className="stats-row"><Stat label="ACTIVE RUNS" value="2" note="1 needs approval" tone="warn"/><Stat label="COMPLETED THIS WEEK" value="48" note="↗ 14% vs last week" tone="up"/><Stat label="TIME SAVED" value="11.4h" note="Across 6 workflows"/><Stat label="APPROVAL RATE" value="92%" note="3 denied safely"/></div><div className="table-card"><div className="table-tools"><div className="segment"><button className="active">All runs</button><button>Active 2</button><button>Completed</button></div><div><button className="secondary">⌕ Search</button><button className="secondary">Filters</button></div></div><div className="data-table"><div className="tr head"><span>Task</span><span>Status</span><span>Activity</span><span>Capabilities</span><span /></div>{runs.map((r,i)=><div className="tr" key={r[0]}><span className="task-name"><i className={i<2?"agent-orb small-orb":"run-done"}>{i<2?"✦":"✓"}</i><span><b>{r[0]}</b><small>Run #{1048-i} · Racore Agent</small></span></span><span><em className={`status ${r[1].toLowerCase().replace(" ", "-")}`}>● {r[1]}</em></span><span>{r[2]}</span><span><code>{r[3]}</code></span><span><button>•••</button></span></div>)}</div></div></Screen>
}

function PublishView({ addRelease }: { addRelease: () => void }) {
  const [stage, setStage] = useState(1); const [publishing, setPublishing] = useState(false);
  function publish() { setPublishing(true); window.setTimeout(() => { setPublishing(false); setStage(4); addRelease(); }, 1400) }
  return <Screen title="Publish" subtitle="Build once. Verify everywhere." actions={<button className="secondary">Read protocol docs ↗</button>}><div className="publish-layout"><div className="publish-main"><div className="stepper">{["Source","Build","Deploy","Verify"].map((x,i)=><div className={stage > i+1 ? "done" : stage===i+1?"active":""} key={x}><i>{stage > i+1?"✓":i+1}</i><span>{x}</span></div>)}</div>{stage===1&&<div className="form-card"><span className="eyebrow">STEP 1 OF 4</span><h2>Choose source</h2><p>Connect a Git repository or publish the current workspace.</p><label>Repository<div className="select-box"><span>⑂</span><div><b>racore-labs/racore.xyz</b><small>main · updated 8 minutes ago</small></div><em>⌄</em></div></label><div className="two-col"><label>Branch<input defaultValue="main" /></label><label>Root directory<input defaultValue="/" /></label></div><label>Build command<input defaultValue="npm run build" /></label><footer><span>Last build passed · 1m 12s</span><button className="primary" onClick={()=>setStage(2)}>Continue to build →</button></footer></div>}{stage===2&&<div className="form-card"><span className="eyebrow">STEP 2 OF 4</span><h2>Review deterministic build</h2><p>Racore generated an immutable bundle from commit <code>8ef75d1</code>.</p><div className="build-output"><header><span><i /> Build successful</span><small>1m 12s</small></header><div><span>Files</span><b>148</b></div><div><span>Bundle size</span><b>4.8 MB</b></div><div><span>Content root</span><code>rcp1-8f3a9d2e…7c19</code></div><div><span>Reproducible</span><b className="up">✓ Verified</b></div></div><footer><button className="secondary" onClick={()=>setStage(1)}>Back</button><button className="primary" onClick={()=>setStage(3)}>Choose deployment →</button></footer></div>}{stage===3&&<div className="form-card"><span className="eyebrow">STEP 3 OF 4</span><h2>Deploy signed release</h2><p>The same verified bundle will be published to all selected providers.</p><div className="deploy-targets">{[["Racore Edge","Global edge network","Selected"],["IPFS","Decentralized peer network","Selected"],["HTTPS origin","racore.xyz origin server","Selected"]].map(x=><label key={x[0]}><input type="checkbox" defaultChecked/><span className="target-icon">◎</span><div><b>{x[0]}</b><small>{x[1]}</small></div><em>{x[2]}</em></label>)}</div><div className="signer"><span>◇</span><div><b>Signed as Racore Labs</b><small>did:key:z6Mkracore…9f2a · Ed25519</small></div><em>VERIFIED</em></div><footer><button className="secondary" onClick={()=>setStage(2)}>Back</button><button className="primary" onClick={publish}>{publishing?"Publishing…":"Sign & publish"}</button></footer></div>}{stage===4&&<div className="form-card success-card"><div className="success-mark">✓</div><span className="eyebrow">RELEASE VERIFIED</span><h2>v1.4.1 is live</h2><p>All three providers returned the expected content root. The custom domain now resolves to this release.</p><div className="release-proof"><code>rcp1-a92f31c8…41de</code><span>3/3 healthy</span></div><footer><button className="secondary" onClick={()=>setStage(1)}>Publish another</button><button className="primary">Inspect release →</button></footer></div>}</div><aside className="publish-side"><h3>Release preview</h3><div><span>Domain</span><b>racore.xyz</b></div><div><span>Next version</span><b>1.4.1</b></div><div><span>Environment</span><em className="status live">● Production</em></div><div><span>Source commit</span><code>8ef75d1</code></div><div><span>Policy</span><b>3-host quorum</b></div><hr/><h4>Preflight checks</h4>{["Build is deterministic","Publisher key available","DNS authorization valid","All providers reachable"].map(x=><p key={x}><span>✓</span>{x}</p>)}</aside></div></Screen>
}

function ReleasesView({ releases, rollback }: { releases: Release[]; rollback: (v:string)=>void }) {
  const [selected,setSelected]=useState(releases[0]);
  useEffect(()=>setSelected(releases[0]),[releases]);
  return <Screen title="Releases" subtitle="Every deployment is immutable, signed, and reversible." actions={<button className="primary">↑ Publish release</button>}><div className="release-layout"><div className="release-list"><header><div><b>racore.xyz</b><small>Production history</small></div><button>Filters ⌄</button></header>{releases.map((r,i)=><button className={selected.version===r.version?"selected":""} onClick={()=>setSelected(r)} key={r.version}><span className={`timeline-dot ${i===0?"live":""}`} /><div><span><b>v{r.version}</b><em className={`status ${r.status.toLowerCase()}`}>● {r.status}</em></span><code>{r.cid}</code><small>{r.time} · {r.commit}</small></div><em>›</em></button>)}</div><article className="release-detail"><header><div><span className="eyebrow">SIGNED RELEASE</span><h2>v{selected.version}</h2><p>Published {selected.time} from commit <code>{selected.commit}</code></p></div><em className={`status ${selected.status.toLowerCase()}`}>● {selected.status}</em></header><div className="proof-grid"><div><span>CONTENT ROOT</span><code>{selected.cid}</code></div><div><span>PUBLISHER</span><b>Racore Labs</b><small>did:key:z6Mkracore…9f2a</small></div><div><span>SIGNATURE</span><b className="up">✓ Valid Ed25519</b><small>Transparency entry #28491</small></div><div><span>SOURCE</span><b>racore-labs/racore.xyz</b><code>{selected.commit}</code></div></div><h3>Deployment verification</h3>{[["Racore Edge","edge://global/"],["IPFS","ipfs://bafy…"],["HTTPS origin","https://racore.xyz/"]].map((h,i)=><div className="deployment-row" key={h[0]}><span className="target-icon">{i===1?"⬡":"◎"}</span><div><b>{h[0]}</b><code>{h[1]}</code></div><span className="latency">{38+i*17}ms</span><em className="status live">✓ Verified</em></div>)}<footer><button className="secondary">Compare files</button>{selected.status!=="Live"&&<button className="danger" onClick={()=>rollback(selected.version)}>Roll back to this release</button>}</footer></article></div></Screen>
}

function NetworkView() {
  const nodes=[["cairo-eg-01","Cairo, Egypt","12ms","84%"],["fra-de-03","Frankfurt, Germany","48ms","71%"],["sin-sg-02","Singapore","126ms","66%"],["nyc-us-05","New York, USA","142ms","79%"]];
  return <Screen title="Network" subtitle="Content availability across Racore peers and public transports." actions={<button className="primary">＋ Connect node</button>}><div className="stats-row"><Stat label="CONNECTED PEERS" value="1,284" note="↗ 42 today" tone="up"/><Stat label="GLOBAL AVAILABILITY" value="99.998%" note="30-day window"/><Stat label="MEDIAN LATENCY" value="38ms" note="↘ 4.8%" tone="up"/><Stat label="VERIFIED STORAGE" value="14.2 TB" note="Across 38 regions"/></div><div className="network-grid"><article className="network-map"><header><div><b>Peer distribution</b><small>Live verified connections</small></div><span><i/> 38 regions</span></header><div className="map-visual"><div className="map-lines"/><i className="node n1"/><i className="node n2"/><i className="node n3"/><i className="node n4"/><i className="node n5"/><i className="node n6"/><span className="map-label l1">Cairo <b>128</b></span><span className="map-label l2">Frankfurt <b>214</b></span><span className="map-label l3">Singapore <b>176</b></span></div></article><article className="protocol-health"><header><b>Protocol health</b><span className="status live">● Operational</span></header>{[["Domain resolution","99.999%"],["Manifest retrieval","99.998%"],["Signature verification","100%"],["Content availability","99.996%"]].map(x=><div key={x[0]}><span>{x[0]}</span><b>{x[1]}</b><i><em style={{width:x[1]}}/></i></div>)}</article></div><div className="table-card"><div className="table-title"><div><b>Your gateway nodes</b><small>Health and routing priority</small></div><button className="secondary">Configure routing</button></div><div className="data-table nodes-table"><div className="tr head"><span>Node</span><span>Region</span><span>Latency</span><span>Load</span><span>Status</span></div>{nodes.map(n=><div className="tr" key={n[0]}><span><i className="node-icon">◎</i><b>{n[0]}</b></span><span>{n[1]}</span><span><code>{n[2]}</code></span><span><div className="load"><i style={{width:n[3]}}/></div>{n[3]}</span><span><em className="status live">● Healthy</em></span></div>)}</div></div></Screen>
}

function PermissionsView() {
  const [grants,setGrants]=useState([true,true,false,true]);
  const rows=[["Browser Reader","Racore Agent","browser.read, browser.navigate","All sites","Automatic"],["Workspace Writer","Racore Agent","workspace.write","Racore / Research","Ask once"],["Release Publisher","Deploy Agent","site.publish, identity.use","racore.xyz","Every time"],["File Export","Research Agent","files.write:selected","Downloads","Ask once"]];
  return <Screen title="Permissions" subtitle="Control exactly what every agent can see and do." actions={<button className="primary">＋ Create policy</button>}><div className="permission-banner"><span>◇</span><div><b>External effects always remain visible</b><p>Messages, purchases, publishing, identity use, and destructive actions require approval unless you explicitly create a narrower policy.</p></div><button>Review defaults</button></div><div className="table-card"><div className="table-title"><div><b>Active capability grants</b><small>4 policies across 3 agents</small></div><div className="segment"><button className="active">Policies</button><button>Agents</button><button>Sites</button></div></div><div className="data-table permissions-table"><div className="tr head"><span>Policy</span><span>Agent</span><span>Capabilities</span><span>Scope</span><span>Approval</span><span>Enabled</span></div>{rows.map((r,i)=><div className="tr" key={r[0]}><span><b>{r[0]}</b><small>Updated {i+1}d ago</small></span><span><i className="agent-orb small-orb">✦</i>{r[1]}</span><span><code>{r[2]}</code></span><span>{r[3]}</span><span><em className={r[4]==="Every time"?"warn-text":""}>{r[4]}</em></span><span><button className={`toggle ${grants[i]?"on":""}`} onClick={()=>setGrants(g=>g.map((x,j)=>i===j?!x:x))}><i/></button></span></div>)}</div></div><div className="audit-callout"><div><span>⑂</span><div><b>Append-only permission audit</b><p>Every grant, denial, expiration, and use is linked in the local audit chain.</p></div></div><button className="secondary">Open audit log →</button></div></Screen>
}

function SettingsView({ theme,setTheme }:{theme:string;setTheme:(x:string)=>void}) {
  const [auto,setAuto]=useState(true);
  return <Screen title="Settings" subtitle="Browser, agent, identity, and protocol preferences."><div className="settings-layout"><nav><button className="active">General</button><button>Agent defaults</button><button>Identity & keys</button><button>Protocol</button><button>Privacy</button><button>Advanced</button></nav><div className="settings-card"><section><div><h3>Appearance</h3><p>Choose how Racore looks on this device.</p></div><div className="theme-options"><button className={theme==="dark"?"selected":""} onClick={()=>setTheme("dark")}><span className="theme-preview dark-preview"><i/><i/><i/></span><b>Dark</b></button><button className={theme==="light"?"selected":""} onClick={()=>setTheme("light")}><span className="theme-preview light-preview"><i/><i/><i/></span><b>Light</b></button></div></section><section><div><h3>Default agent behavior</h3><p>These defaults apply unless a workspace policy overrides them.</p></div><div className="setting-row"><div><b>Allow read-only actions</b><small>Read pages and navigate during an active task.</small></div><button className={`toggle ${auto?"on":""}`} onClick={()=>setAuto(!auto)}><i/></button></div><div className="setting-row"><div><b>External effect approval</b><small>Publishing, messages, uploads, purchases, and identity use.</small></div><select defaultValue="always"><option value="always">Always ask</option><option>Use policies</option></select></div></section><section><div><h3>Racore Protocol</h3><p>Local node and resolution preferences.</p></div><div className="setting-row"><div><b>Local RCP node</b><small>Accept verified blocks and help serve cached releases.</small></div><span className="status live">● Online</span></div><div className="setting-row"><div><b>Preferred resolution</b><small>Order used when multiple transports are available.</small></div><select defaultValue="fast"><option value="fast">Fastest verified</option><option>Peer first</option><option>HTTPS first</option></select></div></section></div></div></Screen>
}

function Screen({ title, subtitle, actions, children }: { title:string; subtitle:string; actions?:React.ReactNode; children:React.ReactNode }) { return <div className="screen"><div className="screen-head"><div><h1>{title}</h1><p>{subtitle}</p></div><div>{actions}</div></div>{children}</div> }

export default function Home() {
  const [view,setView]=useState<View>("browser");
  const [theme,setTheme]=useState("dark");
  const [releases,setReleases]=useState(initialReleases);
  const title=useMemo(()=>nav.find(x=>x.id===view)?.label ?? "Browser",[view]);
  function addRelease(){if(releases[0]?.version==="1.4.1")return;setReleases(r=>[{version:"1.4.1",cid:"rcp1-a92f…41de",commit:"8ef75d1",time:"just now",status:"Live",hosts:3},...r.map(x=>({...x,status:"Previous" as const}))])}
  function rollback(version:string){setReleases(r=>r.map(x=>({...x,status:x.version===version?"Live" as const:"Previous" as const})).sort((a,b)=>a.status==="Live"?-1:b.status==="Live"?1:0))}
  return <main className={`app theme-${theme}`}><SideNav view={view} setView={setView}/><div className="main-shell">{view!=="browser"&&<TopBar title={title} theme={theme} setTheme={setTheme}/>}<div className="main-content">{view==="browser"&&<BrowserView/>}{view==="runs"&&<RunsView/>}{view==="publish"&&<PublishView addRelease={addRelease}/>} {view==="releases"&&<ReleasesView releases={releases} rollback={rollback}/>} {view==="network"&&<NetworkView/>}{view==="permissions"&&<PermissionsView/>}{view==="settings"&&<SettingsView theme={theme} setTheme={setTheme}/>}</div></div></main>;
}
