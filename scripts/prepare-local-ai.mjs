import { execFileSync } from "node:child_process";
import { createHash } from "node:crypto";
import {
  chmodSync,
  createReadStream,
  createWriteStream,
  existsSync,
  mkdirSync,
  readFileSync,
  renameSync,
  rmSync,
  statSync,
} from "node:fs";
import { dirname, join, resolve } from "node:path";
import { Readable } from "node:stream";
import { pipeline } from "node:stream/promises";
import { fileURLToPath } from "node:url";

const projectRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const manifest = JSON.parse(
  readFileSync(join(projectRoot, "config", "local-ai.json"), "utf8"),
);
const target =
  process.env.RACORE_TAURI_TARGET ||
  execFileSync("rustc", ["--print", "host-tuple"], {
    encoding: "utf8",
  }).trim();

function platformKey(triple) {
  const os = triple.includes("windows")
    ? "windows"
    : triple.includes("apple-darwin")
      ? "darwin"
      : triple.includes("linux")
        ? "linux"
        : null;
  const arch = triple.startsWith("x86_64")
    ? "x64"
    : triple.startsWith("aarch64")
      ? "arm64"
      : null;
  if (!os || !arch) throw new Error(`Unsupported local AI target: ${triple}`);
  return `${os}-${arch}`;
}

async function sha256(path) {
  const hash = createHash("sha256");
  for await (const chunk of createReadStream(path)) hash.update(chunk);
  return hash.digest("hex");
}

async function verified(path, expectedHash, expectedBytes) {
  if (!existsSync(path)) return false;
  if (expectedBytes && statSync(path).size !== expectedBytes) return false;
  return (await sha256(path)) === expectedHash;
}

async function download(url, destination, expectedHash, expectedBytes) {
  if (await verified(destination, expectedHash, expectedBytes)) {
    console.log(`Using verified ${destination}`);
    return;
  }

  mkdirSync(dirname(destination), { recursive: true });
  const temporary = `${destination}.download`;
  rmSync(temporary, { force: true });
  console.log(`Downloading ${url}`);
  const response = await fetch(url, { redirect: "follow" });
  if (!response.ok || !response.body) {
    throw new Error(`Download failed with HTTP ${response.status}: ${url}`);
  }
  await pipeline(Readable.fromWeb(response.body), createWriteStream(temporary));
  if (!(await verified(temporary, expectedHash, expectedBytes))) {
    rmSync(temporary, { force: true });
    throw new Error(`Checksum or size mismatch for ${url}`);
  }
  renameSync(temporary, destination);
}

const key = platformKey(target);
const artifact = manifest.runtime.artifacts[key];
if (!artifact) throw new Error(`No llama.cpp runtime configured for ${key}`);

const localRoot = join(projectRoot, "desktop", "runtime", "local-ai");
const modelPath = join(localRoot, "models", manifest.model.file);
const modelUrl =
  `https://huggingface.co/${manifest.model.artifactRepo}/resolve/` +
  `${manifest.model.revision}/${manifest.model.file}`;
await download(
  modelUrl,
  modelPath,
  manifest.model.sha256,
  manifest.model.bytes,
);

const archivePath = join(localRoot, "downloads", artifact.file);
const runtimeUrl =
  `https://github.com/ggml-org/llama.cpp/releases/download/` +
  `${manifest.runtime.version}/${artifact.file}`;
await download(runtimeUrl, archivePath, artifact.sha256);

const runtimeDir = join(localRoot, "runtime", key);
const serverName = key.startsWith("windows") ? "llama-server.exe" : "llama-server";
const serverPath = join(runtimeDir, serverName);
if (!existsSync(serverPath)) {
  rmSync(runtimeDir, { recursive: true, force: true });
  mkdirSync(runtimeDir, { recursive: true });
  execFileSync("tar", ["-xf", archivePath, "-C", runtimeDir], {
    stdio: "inherit",
  });
}
if (!existsSync(serverPath)) {
  throw new Error(`llama-server was not found after extracting ${archivePath}`);
}
if (!key.startsWith("windows")) chmodSync(serverPath, 0o755);

console.log(`Local AI ready: ${manifest.model.id} (${manifest.model.quantization})`);
console.log(`Model: ${modelPath}`);
console.log(`Runtime: ${serverPath}`);
