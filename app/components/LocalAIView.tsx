"use client";

import { useEffect, useState } from "react";
import {
  localAIStatus,
  LocalAIStatus as LocalAIStatusType,
} from "../lib/racore-client";

export function LocalAIView() {
  const [status, setStatus] = useState<LocalAIStatusType | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    const refresh = () =>
      localAIStatus()
        .then((value) => {
          setStatus(value);
          setError("");
        })
        .catch((cause) => {
          setStatus(null);
          setError(cause instanceof Error ? cause.message : "Local AI unavailable");
        });
    void refresh();
    const timer = window.setInterval(refresh, 5000);
    return () => window.clearInterval(timer);
  }, []);

  return (
    <div className="screen providers-screen">
      <div className="screen-head">
        <div>
          <h1>Local AI</h1>
          <p>One bundled browser planner. No API keys and no cloud model routes.</p>
        </div>
        <span className={status?.ready ? "connected" : ""}>
          ● {status?.ready ? "READY" : "STARTING"}
        </span>
      </div>
      <div className="provider-summary">
        <article>
          <span>MODEL</span>
          <strong>{status?.label || "Hammer 2.0 0.5B"}</strong>
        </article>
        <article>
          <span>FOOTPRINT</span>
          <strong>0.5B · Q8_0</strong>
        </article>
        <article>
          <span>PRIVACY</span>
          <strong>100% local</strong>
        </article>
      </div>
      {(error || status?.error) && (
        <div className="provider-message">{error || status?.error}</div>
      )}
      <div className="provider-grid">
        <article>
          <header>
            <span>H2</span>
            <div>
              <b>{status?.model || "MadeAgents/Hammer2.0-0.5b"}</b>
              <small>ON-DEVICE BROWSER PLANNER</small>
            </div>
            <em className={status?.ready ? "connected" : ""}>
              ● {status?.state?.toUpperCase() || "LOADING"}
            </em>
          </header>
          <div className="provider-model">
            <span>Inference engine</span>
            <code>{status?.engine || "llama.cpp · CPU first"}</code>
          </div>
          <div className="provider-tags">
            <span>No network inference</span>
            <span>No credentials</span>
            <span>CC-BY-4.0</span>
          </div>
        </article>
      </div>
    </div>
  );
}
