#!/usr/bin/env node
import { initProject, listReleases, publishProject, rollbackRelease, verifyRelease, buildBundle } from "./rcp.mjs";

const [command = "help", ...args] = process.argv.slice(2);
const cwd = process.cwd();

function option(name, fallback = undefined) { const index = args.indexOf(`--${name}`); return index >= 0 ? args[index + 1] : fallback; }
function short(id) { return id ? `${id.slice(0, 18)}…${id.slice(-8)}` : "—"; }

try {
  if (command === "init") {
    const result = await initProject(cwd, { domain: option("domain"), entrypoint: option("entrypoint") });
    console.log(`Racore project initialized\nDomain: ${result.config.domain}\nProtocol: ${result.config.protocol}`);
  } else if (command === "build") {
    const bundle = await buildBundle(option("input", cwd));
    console.log(`Build verified\nContent root: ${bundle.root}\nFiles: ${bundle.index.files.length}`);
  } else if (command === "publish") {
    const result = await publishProject(cwd, { version: option("version"), commit: option("commit"), input: option("input") });
    console.log(`Release published\nVersion: ${result.manifest.version}\nManifest: ${result.manifest.manifestId}\nContent: ${result.bundle.root}\n${result.dns}`);
  } else if (command === "releases") {
    const releases = await listReleases(cwd);
    if (!releases.length) console.log("No releases yet.");
    else for (const release of releases) console.log(`${release.current ? "*" : " "} v${release.version} ${short(release.manifestId)} ${release.createdAt}`);
  } else if (command === "verify") {
    const result = await verifyRelease(cwd, args[0]);
    console.log(`${result.valid ? "VERIFIED" : "INVALID"} ${short(result.id)}\n${result.reason}`);
    if (!result.valid) process.exitCode = 1;
  } else if (command === "rollback") {
    if (!args[0]) throw new Error("Usage: racore rollback <manifest-id>");
    const result = await rollbackRelease(cwd, args[0]);
    console.log(`Rolled back and verified\nCurrent manifest: ${result.current}`);
  } else {
    console.log(`Racore Protocol CLI

Commands:
  racore init --domain example.com
  racore build [--input ./dist]
  racore publish --version 1.0.0 [--commit abc123] [--input ./dist]
  racore releases
  racore verify [manifest-id]
  racore rollback <manifest-id>`);
  }
} catch (error) {
  console.error(`Racore error: ${error.message}`);
  process.exitCode = 1;
}
