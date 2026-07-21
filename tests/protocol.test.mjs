import test from "node:test";
import assert from "node:assert/strict";
import { mkdtemp, writeFile, readFile } from "node:fs/promises";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { canonicalize, initProject, publishProject, listReleases, verifyRelease, rollbackRelease } from "../protocol/rcp.mjs";

test("canonical encoding sorts object keys", () => {
  assert.equal(canonicalize({ z: 1, a: { y: 2, b: 3 } }), '{"a":{"b":3,"y":2},"z":1}');
});

test("publish, verify, version, and rollback lifecycle", async () => {
  const root = await mkdtemp(join(tmpdir(), "racore-test-"));
  await writeFile(join(root, "index.html"), "<h1>Version one</h1>");
  await initProject(root, { domain: "example.test" });
  const first = await publishProject(root, { version: "1.0.0", commit: "abc123" });
  assert.equal((await verifyRelease(root)).valid, true);
  await writeFile(join(root, "index.html"), "<h1>Version two</h1>");
  const second = await publishProject(root, { version: "1.1.0", commit: "def456" });
  assert.notEqual(first.bundle.root, second.bundle.root);
  const releases = await listReleases(root);
  assert.equal(releases.length, 2);
  assert.equal(releases[0].version, "1.1.0");
  await rollbackRelease(root, first.manifest.manifestId);
  const rolledBack = await listReleases(root);
  assert.equal(rolledBack[0].version, "1.0.0");
});

test("tampered manifest fails verification", async () => {
  const root = await mkdtemp(join(tmpdir(), "racore-tamper-"));
  await writeFile(join(root, "index.html"), "safe");
  await initProject(root, { domain: "safe.test" });
  const release = await publishProject(root, { version: "1.0.0" });
  const path = join(root, ".racore", "releases", `${release.manifest.manifestId}.json`);
  const manifest = JSON.parse(await readFile(path, "utf8"));
  manifest.domain = "evil.test";
  await writeFile(path, JSON.stringify(manifest));
  assert.equal((await verifyRelease(root)).valid, false);
});
