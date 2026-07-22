import { readFileSync, readdirSync, statSync } from "node:fs";
import { basename, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const forbiddenNames = [
  /(^|[\\/])electron(?:\.exe)?$/i,
  /(^|[\\/])node\.exe$/i,
  /(^|[\\/])node_modules([\\/]|$)/i,
  /\.node$/i,
  /node-gyp/i,
];
const forbiddenBinaryMarkers = ["electron.exe", "node_modules", "node-gyp"];

export function findForbiddenArtifacts(root) {
  const findings = [];
  const visit = (path) => {
    const info = statSync(path);
    if (info.isDirectory()) {
      for (const entry of readdirSync(path)) visit(join(path, entry));
      return;
    }
    const normalized = path.replaceAll("\\", "/");
    if (forbiddenNames.some((pattern) => pattern.test(normalized))) {
      findings.push(normalized);
      return;
    }
    if (info.size <= 512 * 1024 * 1024) {
      const content = readFileSync(path).toString("latin1").toLowerCase();
      for (const marker of forbiddenBinaryMarkers) {
        if (content.includes(marker)) findings.push(`${normalized}: contains ${marker}`);
      }
    }
  };
  visit(resolve(root));
  return findings;
}

function run() {
  const projectRoot = resolve(fileURLToPath(new URL("..", import.meta.url)));
  const releaseRoot = resolve(
    process.argv[2] || join(projectRoot, "src-tauri", "target", "release"),
  );
  const deliverables = readdirSync(releaseRoot, { withFileTypes: true })
    .filter((entry) => entry.isFile() || entry.name === "bundle")
    .map((entry) => join(releaseRoot, entry.name));
  for (const generated of ["binaries", "resources"]) {
    const path = join(projectRoot, "src-tauri", generated);
    try {
      if (statSync(path).isDirectory()) deliverables.push(path);
    } catch {
      // A source-only verification may run before sidecars are prepared.
    }
  }
  const findings = deliverables.flatMap(findForbiddenArtifacts);
  if (findings.length) {
    throw new Error(`Forbidden desktop artifacts:\n${findings.join("\n")}`);
  }
  console.log(`Verified ${basename(releaseRoot)}: no Electron or Node runtime artifacts.`);
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  run();
}
