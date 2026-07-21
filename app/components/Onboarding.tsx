"use client";

import Image from "next/image";
import { useEffect, useMemo, useState } from "react";
import {
  checkDaemon,
  connectProvider,
  listProviders,
  ProviderInfo,
} from "../lib/racore-client";

const steps = ["Welcome", "Privacy", "AI", "Network", "Ready"];

export function Onboarding({ onFinish }: { onFinish: () => void }) {
  const [step, setStep] = useState(0);
  const [privacy, setPrivacy] = useState("hybrid");
  const [nodeMode, setNodeMode] = useState(true);
  const [providers, setProviders] = useState<ProviderInfo[]>([]);
  const [provider, setProvider] = useState("openrouter");
  const [key, setKey] = useState("");
  const [message, setMessage] = useState("");
  const [daemon, setDaemon] = useState(false);

  useEffect(() => {
    checkDaemon().then((value) => {
      setDaemon(Boolean(value));
      if (value)
        listProviders()
          .then(setProviders)
          .catch(() => {});
    });
  }, []);
  const connected = useMemo(
    () => providers.filter((item) => item.connected),
    [providers],
  );

  async function saveProvider() {
    if (!key.trim()) return;
    setMessage("Connecting securely…");
    try {
      await connectProvider(provider, key.trim());
      setProviders(await listProviders());
      setKey("");
      setMessage("Connected. The key is encrypted in your local vault.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Connection failed");
    }
  }

  function finish() {
    localStorage.setItem(
      "racore:onboarded",
      JSON.stringify({ privacy, nodeMode, at: Date.now() }),
    );
    onFinish();
  }

  return (
    <div className="onboarding">
      <aside className="onboarding-brand">
        <div className="logo-frame">
          <Image src="/brand/racore-logo.png" alt="Racore.xyz" width={265} height={66} priority />
        </div>
        <div className="onboarding-copy">
          <span>THE AGENTIC WEB</span>
          <h1>
            Browse with intelligence.
            <br />
            Publish with proof.
          </h1>
          <p>
            One private workspace for research, real web navigation, AI agents,
            signed releases, and a network you can help keep alive.
          </p>
        </div>
        <div className="onboarding-proof">
          <div>
            <b>Local-first</b>
            <span>Your keys stay on this device</span>
          </div>
          <div>
            <b>Verifiable</b>
            <span>Every release has a content proof</span>
          </div>
          <div>
            <b>Portable</b>
            <span>Move between AI and hosting providers</span>
          </div>
        </div>
      </aside>
      <main className="onboarding-main">
        <header>
          <div className="onboarding-steps">
            {steps.map((name, index) => (
              <span
                key={name}
                className={
                  index === step ? "active" : index < step ? "done" : ""
                }
              >
                <i>{index < step ? "✓" : index + 1}</i>
                {name}
              </span>
            ))}
          </div>
          <button onClick={finish}>Skip setup</button>
        </header>
        <div className="onboarding-stage">
          {step === 0 && (
            <section className="welcome-step">
              <span className="step-kicker">WELCOME TO RACORE</span>
              <h2>Your browser can now work beside you.</h2>
              <p>
                Ask a question, hand off a workflow, or open any site. Racore
                shows what the agent is doing and asks before anything leaves
                your control.
              </p>
              <div className="welcome-demo">
                <div className="demo-command">
                  <span>✦</span>
                  <b>
                    Research the best decentralized hosting setup for my app
                  </b>
                  <i>↑</i>
                </div>
                <div className="demo-flow">
                  <span>
                    <i>1</i>Discover sources
                  </span>
                  <em>→</em>
                  <span>
                    <i>2</i>Compare options
                  </span>
                  <em>→</em>
                  <span>
                    <i>3</i>Publish a signed brief
                  </span>
                </div>
              </div>
            </section>
          )}
          {step === 1 && (
            <section>
              <span className="step-kicker">PRIVACY MODE</span>
              <h2>Choose where intelligence runs.</h2>
              <p>You can change this per workspace or task later.</p>
              <div className="choice-grid privacy-choices">
                {[
                  [
                    "local",
                    "Local only",
                    "Ollama and local CLI agents. Nothing is sent to cloud AI.",
                    "▣",
                  ],
                  [
                    "hybrid",
                    "Hybrid · Recommended",
                    "Local models for private work, cloud models when quality matters.",
                    "◇",
                  ],
                  [
                    "cloud",
                    "Cloud first",
                    "Use connected providers with automatic quality and cost routing.",
                    "☁",
                  ],
                ].map((item) => (
                  <button
                    key={item[0]}
                    className={privacy === item[0] ? "selected" : ""}
                    onClick={() => setPrivacy(item[0])}
                  >
                    <i>{item[3]}</i>
                    <b>{item[1]}</b>
                    <span>{item[2]}</span>
                    <em>{privacy === item[0] ? "●" : "○"}</em>
                  </button>
                ))}
              </div>
            </section>
          )}
          {step === 2 && (
            <section>
              <span className="step-kicker">AI PROVIDERS</span>
              <h2>Connect your first intelligence provider.</h2>
              <p>
                API keys are encrypted by the Python daemon and never stored in
                the website.
              </p>
              {daemon ? (
                <>
                  <div className="provider-picks">
                    {(providers.length
                      ? providers
                      : ([
                          { id: "openrouter", name: "OpenRouter" },
                          { id: "openai", name: "OpenAI" },
                          { id: "anthropic", name: "Anthropic" },
                          { id: "gemini", name: "Gemini" },
                          { id: "kimi", name: "Kimi" },
                          { id: "ollama", name: "Ollama" },
                        ] as ProviderInfo[])
                    )
                      .filter(
                        (p) =>
                          !["opencode", "claude-code", "kimi-code"].includes(
                            p.id,
                          ),
                      )
                      .map((p) => (
                        <button
                          key={p.id}
                          className={provider === p.id ? "selected" : ""}
                          onClick={() => setProvider(p.id)}
                        >
                          <span>{p.name.slice(0, 2).toUpperCase()}</span>
                          <b>{p.name}</b>
                          <em>
                            {p.connected
                              ? "CONNECTED"
                              : p.local
                                ? "LOCAL"
                                : "CONNECT"}
                          </em>
                        </button>
                      ))}
                  </div>
                  {provider !== "ollama" && (
                    <div className="key-connect">
                      <label>
                        {providers.find((p) => p.id === provider)?.name ||
                          provider}{" "}
                        API key
                        <input
                          type="password"
                          value={key}
                          onChange={(event) => setKey(event.target.value)}
                          placeholder="Paste your key — it stays on this device"
                        />
                      </label>
                      <button onClick={saveProvider} disabled={!key.trim()}>
                        Connect securely
                      </button>
                    </div>
                  )}
                  <p className="connect-message">
                    {message ||
                      `${connected.length} provider${connected.length === 1 ? "" : "s"} ready`}
                  </p>
                </>
              ) : (
                <div className="daemon-needed">
                  <span>◌</span>
                  <div>
                    <b>Open this setup in Racore Desktop</b>
                    <p>
                      The local Python service is required to encrypt
                      credentials and connect providers. You can continue
                      exploring the web preview without entering a key.
                    </p>
                  </div>
                </div>
              )}
            </section>
          )}
          {step === 3 && (
            <section>
              <span className="step-kicker">RACORE NETWORK</span>
              <h2>Help keep the web available.</h2>
              <p>
                Your desktop can join the Racore Mesh while the application
                runs. Background operation is always your choice.
              </p>
              <div className="network-onboard">
                <div className="network-orbit">
                  <i />
                  <i />
                  <i />
                  <i />
                  <i />
                  <span>R</span>
                </div>
                <div className="network-details">
                  <label>
                    <span>
                      <b>Join Racore Mesh</b>
                      <small>
                        Discover peers and exchange signed availability events.
                      </small>
                    </span>
                    <button
                      className={`toggle ${nodeMode ? "on" : ""}`}
                      onClick={() => setNodeMode(!nodeMode)}
                    >
                      <i />
                    </button>
                  </label>
                  <label>
                    <span>
                      <b>Use local IPFS</b>
                      <small>
                        Connect to Kubo on ports 5001 and 8080 when available.
                      </small>
                    </span>
                    <em>AUTO-DETECT</em>
                  </label>
                  <label>
                    <span>
                      <b>Storage allowance</b>
                      <small>Maximum cache Racore may use.</small>
                    </span>
                    <select defaultValue="5">
                      <option value="1">1 GB</option>
                      <option value="5">5 GB</option>
                      <option value="20">20 GB</option>
                    </select>
                  </label>
                </div>
              </div>
            </section>
          )}
          {step === 4 && (
            <section className="ready-step">
              <div className="ready-mark">✓</div>
              <span className="step-kicker">RACORE IS READY</span>
              <h2>Your agentic workspace is online.</h2>
              <p>
                Start with a question, open a URL, or give Racore a multi-step
                task. You stay in control of every external action.
              </p>
              <div className="ready-summary">
                <span>
                  <b>
                    {privacy === "hybrid"
                      ? "Hybrid"
                      : privacy === "local"
                        ? "Local"
                        : "Cloud"}
                  </b>{" "}
                  intelligence
                </span>
                <span>
                  <b>{connected.length || (privacy === "local" ? 1 : 0)}</b> AI
                  providers
                </span>
                <span>
                  <b>{nodeMode ? "On" : "Off"}</b> mesh node
                </span>
              </div>
            </section>
          )}
        </div>
        <footer>
          <button
            className="back"
            onClick={() => setStep((value) => Math.max(0, value - 1))}
            disabled={step === 0}
          >
            Back
          </button>
          <span>
            Step {step + 1} of {steps.length}
          </span>
          <button
            className="next"
            onClick={() =>
              step === steps.length - 1
                ? finish()
                : setStep((value) => value + 1)
            }
          >
            {step === steps.length - 1 ? "Enter Racore" : "Continue"} →
          </button>
        </footer>
      </main>
    </div>
  );
}
