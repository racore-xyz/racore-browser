import assert from "node:assert/strict";
import { mkdtemp, mkdir, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { findForbiddenArtifacts } from "../scripts/verify-tauri-bundle.mjs";

test("bundle verifier accepts Tauri assets and named sidecars", async () => {
  const root = await mkdtemp(join(tmpdir(), "racore-tauri-clean-"));
  try {
    await writeFile(join(root, "racore-browser.exe"), "tauri application");
    await writeFile(join(root, "racored.exe"), "go daemon");
    await writeFile(join(root, "ipfs.exe"), "kubo executable");
    assert.deepEqual(findForbiddenArtifacts(root), []);
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});

test("bundle verifier rejects native Node and Electron artifacts", async () => {
  const root = await mkdtemp(join(tmpdir(), "racore-tauri-dirty-"));
  try {
    await mkdir(join(root, "node_modules"));
    await writeFile(join(root, "node_modules", "binding.node"), "native binding");
    await writeFile(join(root, "payload.bin"), "contains electron.exe runtime");
    const findings = findForbiddenArtifacts(root);
    assert.ok(findings.some((finding) => finding.includes("binding.node")));
    assert.ok(findings.some((finding) => finding.includes("electron.exe")));
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});
