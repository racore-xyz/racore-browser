import { execFileSync } from "node:child_process";
import { copyFileSync, existsSync, mkdirSync } from "node:fs";
import { basename, dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const projectRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const tauriRoot = join(projectRoot, "src-tauri");
const target =
  process.env.RACORE_TAURI_TARGET ||
  execFileSync("rustc", ["--print", "host-tuple"], {
    encoding: "utf8",
  }).trim();

function targetPlatform(triple) {
  const goos = triple.includes("windows")
    ? "windows"
    : triple.includes("apple-darwin")
      ? "darwin"
      : triple.includes("linux")
        ? "linux"
        : null;
  const goarch = triple.startsWith("x86_64")
    ? "amd64"
    : triple.startsWith("aarch64")
      ? "arm64"
      : null;

  if (!goos || !goarch) {
    throw new Error(`Unsupported Tauri target triple: ${triple}`);
  }
  return { goos, goarch, executableSuffix: goos === "windows" ? ".exe" : "" };
}

const platform = targetPlatform(target);
const go = process.env.RACORE_GO || "go";
const binaryDir = join(tauriRoot, "binaries");
mkdirSync(binaryDir, { recursive: true });

for (const name of ["racored", "racore"]) {
  const output = join(
    binaryDir,
    `${name}-${target}${platform.executableSuffix}`,
  );
  execFileSync(
    go,
    [
      "build",
      "-trimpath",
      "-ldflags=-s -w",
      "-o",
      output,
      `./cmd/${name}`,
    ],
    {
      cwd: join(projectRoot, "god"),
      env: {
        ...process.env,
        CGO_ENABLED: "0",
        GOOS: platform.goos,
        GOARCH: platform.goarch,
      },
      stdio: "inherit",
    },
  );
}

const defaultKubo = join(
  projectRoot,
  "desktop",
  "runtime",
  "kubo",
  `ipfs${platform.executableSuffix}`,
);
const kuboSource = resolve(process.env.RACORE_KUBO_BINARY || defaultKubo);
if (!existsSync(kuboSource)) {
  throw new Error(
    `Kubo binary not found at ${kuboSource}. Set RACORE_KUBO_BINARY to a ${platform.goos}/${platform.goarch} Kubo executable.`,
  );
}

const kuboDestination = join(
  tauriRoot,
  "resources",
  "kubo",
  `ipfs${platform.executableSuffix}`,
);
mkdirSync(dirname(kuboDestination), { recursive: true });
copyFileSync(kuboSource, kuboDestination);

console.log(
  `Prepared ${target} sidecars and ${basename(kuboSource)} for Tauri.`,
);
