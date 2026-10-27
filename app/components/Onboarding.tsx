"use client";

import { useEffect, useState } from "react";
import { checkDaemon } from "../lib/racore-client";

const steps = ["Welcome", "Privacy", "AI", "Network", "Ready"];

type Health = {
  mesh?: { online?: boolean; peers?: number };
  ipfs?: { online?: boolean };
};

export function Onboarding({ onFinish }: { onFinish: () => void }) {
  const [step, setStep] = useState(0);
  const [health, setHealth] = useState<Health | null>(null);

  useEffect(() => {
    let active = true;
    const refresh = () =>
      checkDaemon().then((value) => {
        if (active) setHealth((value as unknown as Health) ?? null);
      });
    void refresh();
    const timer = window.setInterval(refresh, 4000);
    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, []);

  const daemon = health !== null;
  const meshOnline = Boolean(health?.mesh?.online);
  const ipfsOnline = Boolean(health?.ipfs?.online);

  function finish() {
    localStorage.setItem("racore:onboarded", JSON.stringify({ at: Date.now() }));
    onFinish();
  }

  return (
    <div className="onboarding">
      <aside className="onboarding-brand">
        <div className="logo-frame">
          <div className="terminal-wordmark compact" aria-label="Racore.xyz" role="img">
            <span>RACORE</span>
            <b>.XYZ</b>
            <i>▮</i>
          </div>
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
            <span>AI inference stays on this device</span>
          </div>
          <div>
            <b>Verifiable</b>
            <span>Every release has a content proof</span>
          </div>
          <div>
            <b>Portable</b>
            <span>No model account or API key required</span>
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
              <h2>Intelligence runs on your device.</h2>
              <p>Racore does not send prompts to a cloud model.</p>
              <div className="choice-grid privacy-choices">
                <button className="selected">
                  <i>▣</i>
                  <b>Local only</b>
                  <span>Hammer 2.0 0.5B runs through the bundled llama.cpp engine.</span>
                  <em>●</em>
                </button>
              </div>
            </section>
          )}
          {step === 2 && (
            <section>
              <span className="step-kicker">LOCAL AI</span>
              <h2>Your browser planner is already included.</h2>
              <p>
                No API key, account, subscription, or third-party model route is
                required.
              </p>
              <div className="daemon-needed">
                <span>{daemon ? "✓" : "◌"}</span>
                <div>
                  <b>{daemon ? "Local AI service detected" : "Open this setup in Racore Desktop"}</b>
                  <p>
                    {daemon
                      ? "Hammer is managed in the background with a CPU-first, low-resource configuration."
                      : "The packaged desktop app includes both the model and its inference runtime."}
                  </p>
                </div>
              </div>
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
                      <b>Racore Mesh</b>
                      <small>
                        Signed UDP discovery on 239.255.77.77:47777 while the
                        desktop app runs.
                      </small>
                    </span>
                    <em className={meshOnline ? "network-live" : ""}>
                      {meshOnline
                        ? `● ONLINE · ${health?.mesh?.peers ?? 0} PEERS`
                        : "○ OFFLINE"}
                    </em>
                  </label>
                  <label>
                    <span>
                      <b>Local IPFS (Kubo)</b>
                      <small>
                        Managed automatically by racored. Content API on 127.0.0.1:5001,
                        gateway on 127.0.0.1:8180.
                      </small>
                    </span>
                    <em className={ipfsOnline ? "network-live" : ""}>
                      {ipfsOnline ? "● READY" : "○ NOT DETECTED"}
                    </em>
                  </label>
                  <label>
                    <span>
                      <b>Background operation</b>
                      <small>
                        Mesh, IPFS, and the local model stop when Racore Desktop
                        quits. Nothing runs as a system service.
                      </small>
                    </span>
                    <em>SESSION ONLY</em>
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
                    Local
                  </b>{" "}
                  intelligence
                </span>
                <span>
                  <b>1</b> bundled AI model
                </span>
                <span>
                  <b>{meshOnline ? "On" : "Off"}</b> mesh node
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
