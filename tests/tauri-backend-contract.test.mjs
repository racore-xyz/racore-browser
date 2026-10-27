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

test("browser windows use one persistent full-featured profile", async () => {
  const source = await readFile(
    new URL("../src-tauri/src/windows.rs", import.meta.url),
    "utf8",
  );

  assert.match(source, /\.join\("browser-profile"\)/);
  assert.match(source, /\.data_directory\(data_directory\)/);
  assert.match(source, /\.incognito\(false\)/);
  assert.match(source, /\.enable_clipboard_access\(\)/);
  assert.match(source, /\.zoom_hotkeys_enabled\(true\)/);
  assert.match(source, /\.general_autofill_enabled\(true\)/);
  assert.match(source, /NewWindowResponse::Allow/);
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

test("Tauri bundle declares named sidecars, Kubo, and local AI resources", async () => {
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
  assert.equal(windows.bundle.resources["resources/local-ai/"], "local-ai/");
  for (const platform of [linux, macos]) {
    assert.equal(platform.bundle.resources["resources/kubo/ipfs"], "kubo/ipfs");
    assert.equal(platform.bundle.resources["resources/local-ai/"], "local-ai/");
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
  assert.match(transport, /Control: func/);
  assert.match(transport, /setReuseAddress\(fd\)/);
  assert.match(transport, /listenConfig\.ListenPacket/);
  assert.ok(
    transport.indexOf("listenConfig.ListenPacket") <
      transport.indexOf("ipv4.NewPacketConn"),
    "socket options must be applied before the multicast bind",
  );
  assert.match(transport, /runtime\.GOOS != "windows"/);
  assert.match(windows, /syscall\.Handle\(fd\)/);
  assert.match(unix, /int\(fd\)/);
});

test("window commands defer creation to the event loop", async () => {
  const [commands, windows] = await Promise.all([
    readFile(new URL("../src-tauri/src/commands.rs", import.meta.url), "utf8"),
    readFile(new URL("../src-tauri/src/windows.rs", import.meta.url), "utf8"),
  ]);
  assert.match(commands, /pub async fn open_browser/);
  assert.match(commands, /pub async fn browser_action/);
  assert.match(windows, /run_on_main_thread/);
});

test("desktop detection requires the live Tauri IPC bridge", async () => {
  const source = await readFile(
    new URL("../app/lib/desktop.ts", import.meta.url),
    "utf8",
  );
  assert.match(source, /"__TAURI_INTERNALS__" in window/);
  assert.doesNotMatch(source, /dataset\.desktop === "tauri"/);
});

