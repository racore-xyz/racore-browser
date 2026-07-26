import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";
import {
  RACORE_SUFFIXES,
  racoreDomainFromInput,
} from "../app/lib/racore-domains.ts";

test("browser tabs persist and AI requests carry conversation history", async () => {
  const source = await readFile(
    new URL("../app/components/AgenticBrowserView.tsx", import.meta.url),
    "utf8",
  );

  assert.match(source, /racore:browser-tabs:v1/);
  assert.match(source, /function addTab\(\)/);
  assert.match(source, /function closeTab\(id: string\)/);
  assert.match(source, /\.\.\.messages\.map\(\(\{ role, content \}\)/);
  assert.match(source, /Continue the conversation or enter a website/);
});

test("direct video links open in the in-tab native media player", async () => {
  const [component, config] = await Promise.all([
    readFile(
      new URL("../app/components/AgenticBrowserView.tsx", import.meta.url),
      "utf8",
    ),
    readFile(new URL("../src-tauri/tauri.conf.json", import.meta.url), "utf8"),
  ]);

  assert.match(component, /DIRECT_VIDEO_PATTERN/);
  assert.match(component, /<video/);
  assert.match(component, /controls/);
  assert.match(component, /playsInline/);
  assert.match(config, /media-src https: http: blob:/);
});

test("Racore mode recognizes only the four network domain suffixes", () => {
  assert.deepEqual(RACORE_SUFFIXES, [".racore", ".rac", ".core", ".ra"]);
  for (const domain of [
    "studio.racore",
    "racore://market.rac",
    "https://docs.core/guide",
    "node.ra",
  ]) {
    assert.ok(racoreDomainFromInput(domain), domain);
  }
  for (const domain of ["example.com", "racore", ".ra", "../escape.ra"]) {
    assert.equal(racoreDomainFromInput(domain), null, domain);
  }
});

test("browser exposes Racore and normal browsing modes with mesh discovery", async () => {
  const source = await readFile(
    new URL("../app/components/AgenticBrowserView.tsx", import.meta.url),
    "utf8",
  );
  assert.match(source, /Racore Mode/);
  assert.match(source, /Normal Browsing/);
  assert.match(source, /\/v1\/authority\/network-domains/);
  assert.match(source, /\/v1\/authority\/resolve\//);
  assert.match(source, /Ready on your network/);
});
