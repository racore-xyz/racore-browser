import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

async function render() {
  const workerUrl = new URL("../dist/server/index.js", import.meta.url);
  workerUrl.searchParams.set("test", `${process.pid}-${Date.now()}`);
  const { default: worker } = await import(workerUrl.href);

  return worker.fetch(
    new Request("http://localhost/", { headers: { accept: "text/html" } }),
    {
      ASSETS: { fetch: async () => new Response("Not found", { status: 404 }) },
    },
    { waitUntil() {}, passThroughOnException() {} },
  );
}

test("server-renders Racore metadata and branded loading state", async () => {
  const response = await render();
  assert.equal(response.status, 200);
  assert.match(response.headers.get("content-type") ?? "", /^text\/html\b/i);

  const html = await response.text();
  assert.match(
    html,
    /<title>Racore — Agentic Browser &amp; Open Web Protocol<\/title>/i,
  );
  assert.match(html, /og-racore-(?:v2|editorial)\.png/);
  assert.match(
    html,
    /Racore (?:agentic browser and decentralized AI network|— the browser built for agency)/,
  );
  assert.match(html, /class="(?:racore-loading|landing)"/);
});

test("ships the agentic browser, onboarding, providers, and live mesh views", async () => {
  const [page, onboarding, browser, providers, network, layout] =
    await Promise.all([
      readFile(new URL("../app/page.tsx", import.meta.url), "utf8"),
      readFile(
        new URL("../app/components/Onboarding.tsx", import.meta.url),
        "utf8",
      ),
      readFile(
        new URL("../app/components/AgenticBrowserView.tsx", import.meta.url),
        "utf8",
      ),
      readFile(
        new URL("../app/components/ProvidersView.tsx", import.meta.url),
        "utf8",
      ),
      readFile(
        new URL("../app/components/LiveNetworkView.tsx", import.meta.url),
        "utf8",
      ),
      readFile(new URL("../app/layout.tsx", import.meta.url), "utf8"),
    ]);

  assert.match(page, /<Onboarding|className="landing"/);
  assert.match(onboarding, /AI PROVIDERS/);
  assert.match(browser, /Ask a question(?:, describe a task, or paste a URL| or enter a website)/);
  assert.match(providers, /AI providers/);
  assert.match(network, /Racore Mesh/);
  assert.match(layout, /og-racore-(?:v2|editorial)\.png/);
});
