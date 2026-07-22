import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const auditUrl = new URL("../docs/tauri-migration-audit.md", import.meta.url);

test("Tauri migration audit records every Electron IPC channel and event", async () => {
  const audit = await readFile(auditUrl, "utf8");
  const channels = [
    "racore:daemon-status",
    "racore:api",
    "racore:open-browser",
    "racore:open-external",
    "racore:platform",
    "racore:daemon-exit",
  ];

  for (const channel of channels) {
    assert.ok(audit.includes(`\`${channel}\``), `missing ${channel}`);
  }
});

test("Tauri migration contract preserves React, schemas, and mesh code", async () => {
  const audit = await readFile(auditUrl, "utf8");

  assert.match(audit, /Keep React and the existing Racore UI\./);
  assert.match(audit, /Preserve the frontend-to-daemon HTTP workflow and JSON schemas\./);
  assert.match(audit, /Do not modify `god\/internal\/mesh\/\*\*`/);
  assert.match(audit, /Do not ship Node\.js, Electron, `\.node` native bindings/);
  assert.match(audit, /Add or update documentation and automated tests in every implementation step/);
});

test("Tauri migration roadmap defines five independently committed steps", async () => {
  const audit = await readFile(auditUrl, "utf8");
  const headings = audit.match(/^### Step \d+ — .+$/gm) ?? [];
  const commits = audit.match(/^- Commit: `[^`]+`\.$/gm) ?? [];

  assert.equal(headings.length, 5);
  assert.equal(commits.length, 5);
});
