import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const root = new URL("../", import.meta.url);

async function text(path) {
  return readFile(new URL(path, root), "utf8");
}

async function jpegDimensions(path) {
  const file = await readFile(new URL(path, root));
  assert.deepEqual([...file.subarray(0, 3)], [255, 216, 255]);
  let offset = 2;
  while (offset < file.length - 9) {
    if (file[offset] !== 255) {
      offset += 1;
      continue;
    }
    const marker = file[offset + 1];
    const length = file.readUInt16BE(offset + 2);
    if ([192, 193, 194].includes(marker)) {
      return { height: file.readUInt16BE(offset + 5), width: file.readUInt16BE(offset + 7) };
    }
    offset += 2 + length;
  }
  throw new Error(`No JPEG dimensions found in ${path}`);
}

test("README presents the product with real captures and generated artwork", async () => {
  const readme = await text("README.md");
  const screenshots = [
    "docs/assets/readme/landing.jpg",
    "docs/assets/readme/onboarding.jpg",
    "docs/assets/readme/workspace.jpg",
  ];

  for (const asset of screenshots) {
    assert.match(readme, new RegExp(asset.replaceAll("/", "\\/")));
    const dimensions = await jpegDimensions(asset);
    assert.ok(dimensions.width >= 1200, `${asset} should be README-banner width`);
    assert.ok(dimensions.height >= 700, `${asset} should preserve desktop detail`);
  }

  const heroPath = "docs/assets/readme/hero.png";
  const hero = await readFile(new URL(heroPath, root));
  assert.match(readme, new RegExp(heroPath.replaceAll("/", "\\/")));
  assert.deepEqual([...hero.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
  assert.ok(hero.byteLength > 1_000_000, "generated hero should preserve presentation detail");

  assert.match(readme, /```mermaid/);
  assert.match(readme, /Security by default/);
  assert.match(readme, /76\.7% smaller/);
});

test("README signal is self-contained animated SVG", async () => {
  const readme = await text("README.md");
  const signal = await text("docs/assets/readme/signal.svg");

  assert.match(readme, /docs\/assets\/readme\/signal\.svg/);
  assert.match(signal, /<animate\b/g);
  assert.match(signal, /LOCAL-FIRST/);
  assert.match(signal, /AGENTIC/);
  assert.match(signal, /VERIFIABLE/);
  assert.doesNotMatch(signal, /(?:href|src)="https?:\/\//);
});
