import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

test("Rust registers the complete Electron replacement command surface", async () => {
  const source = await readFile(
    new URL("../src-tauri/src/lib.rs", import.meta.url),
    "utf8",
  );
  for (const command of [
    "daemon_status",
    "daemon_request",
    "platform_info",
    "open_browser",
    "open_external",
  ]) {
    assert.match(source, new RegExp(`commands::${command}`));
  }
});

test("sidecar preparation compiles Go code without touching mesh sources", async () => {
  const source = await readFile(
    new URL("../scripts/prepare-tauri-sidecars.mjs", import.meta.url),
    "utf8",
  );
  assert.match(source, /for \(const name of \["racored", "racore"\]\)/);
  assert.match(source, /CGO_ENABLED: "0"/);
  assert.doesNotMatch(source, /internal[\\/]mesh/);
});

test("Tauri bundle declares only named Racore sidecars and platform Kubo resources", async () => {
  const [config, windows, linux, macos] = await Promise.all([
    "tauri.conf.json",
    "tauri.windows.conf.json",
    "tauri.linux.conf.json",
    "tauri.macos.conf.json",
  ].map(async (name) =>
    JSON.parse(
      await readFile(new URL(`../src-tauri/${name}`, import.meta.url), "utf8"),
    ),
  ));
  assert.deepEqual(config.bundle.externalBin, [
    "binaries/racored",
    "binaries/racore",
  ]);
  assert.equal(windows.bundle.resources["resources/kubo/ipfs.exe"], "kubo/ipfs.exe");
  for (const platform of [linux, macos]) {
    assert.equal(platform.bundle.resources["resources/kubo/ipfs"], "kubo/ipfs");
  }
});

test("UDP syscall adapters preserve the no-mesh-change boundary", async () => {
  const [transport, windows, unix] = await Promise.all([
    readFile(new URL("../god/internal/transport/udp.go", import.meta.url), "utf8"),
    readFile(
      new URL("../god/internal/transport/socket_windows.go", import.meta.url),
      "utf8",
    ),
    readFile(
      new URL("../god/internal/transport/socket_unix.go", import.meta.url),
      "utf8",
    ),
  ]);
  assert.match(transport, /setReuseAddress\(file\.Fd\(\)\)/);
  assert.match(windows, /syscall\.Handle\(fd\)/);
  assert.match(unix, /int\(fd\)/);
});
