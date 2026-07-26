"use client";

import Image from "next/image";
import { useEffect, useState } from "react";
import { checkDaemon } from "../lib/racore-client";

const steps = ["Welcome", "Privacy", "AI", "Network", "Ready"];

export function Onboarding({ onFinish }: { onFinish: () => void }) {
  const [step, setStep] = useState(0);
  const [nodeMode, setNodeMode] = useState(true);
  const [daemon, setDaemon] = useState(false);

  useEffect(() => {
    checkDaemon().then((value) => setDaemon(Boolean(value)));
  }, []);

  function finish() {
    localStorage.setItem(
      "racore:onboarded",
      JSON.stringify({ intelligence: "local", nodeMode, at: Date.now() }),
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
                  <p>{daemon ? "Hammer is managed in the background with a CPU-first, low-resource configuration." : "The packaged desktop app includes both the model and its inference runtime."}</p>
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
                    Local
                  </b>{" "}
                  intelligence
                </span>
                <span>
                  <b>1</b> bundled AI model
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
