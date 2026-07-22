import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

async function readJson(path) {
  return JSON.parse(await readFile(new URL(path, import.meta.url), "utf8"));
}

test("Tauri uses the static React desktop build and preserves window geometry", async () => {
  const config = await readJson("../src-tauri/tauri.conf.json");
  const [window] = config.app.windows;

  assert.equal(config.identifier, "xyz.racore.browser");
  assert.equal(config.build.frontendDist, "../dist-desktop");
  assert.equal(config.build.devUrl, "http://127.0.0.1:1420");
  assert.deepEqual(
    [window.width, window.height, window.minWidth, window.minHeight],
    [1480, 940, 980, 680],
  );
  assert.equal(config.app.withGlobalTauri, false);
  assert.match(config.app.security.csp, /object-src 'none'/);
});

test("main capability grants no filesystem, OS, or shell plugin access", async () => {
  const capability = await readJson("../src-tauri/capabilities/main.json");
  const permissions = JSON.stringify(capability.permissions);

  assert.deepEqual(capability.windows, ["main"]);
  assert.doesNotMatch(permissions, /(?:fs|os|shell|opener):/);
});

test("desktop Vite entry reuses the React product workspace", async () => {
  const [entry, viteConfig] = await Promise.all([
    readFile(new URL("../desktop-ui/main.tsx", import.meta.url), "utf8"),
    readFile(new URL("../desktop-ui/vite.config.ts", import.meta.url), "utf8"),
  ]);

  assert.match(entry, /RacoreProductApp/);
  assert.match(viteConfig, /"next\/image"/);
  assert.match(viteConfig, /dist-desktop/);
});
