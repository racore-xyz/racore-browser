import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

function bitmapDimensions(buffer) {
  assert.equal(buffer.toString("ascii", 0, 2), "BM");
  return {
    width: buffer.readInt32LE(18),
    height: buffer.readInt32LE(22),
    bitsPerPixel: buffer.readUInt16LE(28),
  };
}

test("NSIS installer uses Racore branding and non-commercial license", async () => {
  const [config, license] = await Promise.all([
    readFile(new URL("../src-tauri/tauri.conf.json", import.meta.url), "utf8").then(JSON.parse),
    readFile(new URL("../src-tauri/licenses/INSTALLER-LICENSE.txt", import.meta.url), "utf8"),
  ]);
  const nsis = config.bundle.windows.nsis;

  assert.equal(config.bundle.publisher, "Racore");
  assert.equal(config.bundle.license, "LicenseRef-Racore-NonCommercial-1.0");
  assert.equal(config.bundle.licenseFile, "licenses/INSTALLER-LICENSE.txt");
  assert.equal(nsis.installerIcon, "icons/icon.ico");
  assert.equal(nsis.headerImage, "windows/nsis/header.bmp");
  assert.equal(nsis.sidebarImage, "windows/nsis/sidebar.bmp");
  assert.match(license, /Non-Commercial Use/);
  assert.match(license, /SYSTEM INFORMATION/);
  assert.match(license, /THIRD-PARTY SOFTWARE NOTICES/);
  assert.match(license, /CC BY 4\.0/);
  assert.match(license, /llama\.cpp/);
  assert.match(license, /IPFS Kubo/);
});

test("NSIS bitmap assets use the required dimensions and 24-bit color", async () => {
  const [header, sidebar] = await Promise.all([
    readFile(new URL("../src-tauri/windows/nsis/header.bmp", import.meta.url)),
    readFile(new URL("../src-tauri/windows/nsis/sidebar.bmp", import.meta.url)),
  ]);

  assert.deepEqual(bitmapDimensions(header), {
    width: 150,
    height: 57,
    bitsPerPixel: 24,
  });
  assert.deepEqual(bitmapDimensions(sidebar), {
    width: 164,
    height: 314,
    bitsPerPixel: 24,
  });
});

test("transparent Racore wordmark is retained as the NSIS design source", async () => {
  const logo = await readFile(
    new URL("../src-tauri/windows/nsis/racore-logo-transparent.png", import.meta.url),
  );
  assert.equal(logo.toString("hex", 1, 4), "504e47");
});
