import { createHash, generateKeyPairSync, sign, verify } from "node:crypto";
import { mkdir, readFile, readdir, writeFile } from "node:fs/promises";
import { basename, join, relative, resolve } from "node:path";

export const PROTOCOL_VERSION = "rcp/0.1";
export const STATE_DIR = ".racore";

export function canonicalize(value) {
  if (Array.isArray(value)) return `[${value.map(canonicalize).join(",")}]`;
  if (value && typeof value === "object") {
    return `{${Object.keys(value).sort().map(key => `${JSON.stringify(key)}:${canonicalize(value[key])}`).join(",")}}`;
  }
  return JSON.stringify(value);
}

export function digest(input) {
  return createHash("sha256").update(input).digest("base64url");
}

export function contentId(input) {
  return `rcp1-${digest(input)}`;
}

async function walk(root, current = root) {
  const entries = await readdir(current, { withFileTypes: true });
  const files = [];
  for (const entry of entries.sort((a, b) => a.name.localeCompare(b.name))) {
    if ([STATE_DIR, ".git", "node_modules"].includes(entry.name)) continue;
    const absolute = join(current, entry.name);
    if (entry.isDirectory()) files.push(...await walk(root, absolute));
    else files.push({ absolute, path: relative(root, absolute).replaceAll("\\", "/") });
  }
  return files;
}

export async function buildBundle(inputDirectory) {
  const root = resolve(inputDirectory);
  const files = [];
  for (const file of await walk(root)) {
    const bytes = await readFile(file.absolute);
    files.push({ path: file.path, size: bytes.byteLength, hash: digest(bytes) });
  }
  const index = { algorithm: "sha256", files };
  return { root: contentId(canonicalize(index)), index };
}

export function createPublisherIdentity() {
  const { publicKey, privateKey } = generateKeyPairSync("ed25519");
  return {
    publicKey: publicKey.export({ type: "spki", format: "pem" }),
    privateKey: privateKey.export({ type: "pkcs8", format: "pem" }),
  };
}

export function createManifest({ domain, version, root, parent = null, commit = "working-tree", entrypoint = "/index.html", mirrors = [] }) {
  return { protocol: PROTOCOL_VERSION, domain, version, root, parent, commit, createdAt: new Date().toISOString(), entrypoint, mirrors };
}

export function signManifest(manifest, privateKey, publicKey) {
  const payload = Buffer.from(canonicalize(manifest));
  const signature = sign(null, payload, privateKey).toString("base64url");
  const signed = { ...manifest, signature: { algorithm: "ed25519", publicKey, value: signature } };
  return { ...signed, manifestId: contentId(canonicalize(signed)) };
}

export function verifyManifest(signedManifest) {
  const { signature, manifestId, ...manifest } = signedManifest;
  if (!signature || signature.algorithm !== "ed25519") return { valid: false, reason: "Missing or unsupported signature" };
  const signatureValid = verify(null, Buffer.from(canonicalize(manifest)), signature.publicKey, Buffer.from(signature.value, "base64url"));
  const expectedId = contentId(canonicalize({ ...manifest, signature }));
  if (!signatureValid) return { valid: false, reason: "Publisher signature is invalid" };
  if (manifestId !== expectedId) return { valid: false, reason: "Manifest content ID does not match" };
  if (manifest.protocol !== PROTOCOL_VERSION) return { valid: false, reason: `Unsupported protocol ${manifest.protocol}` };
  return { valid: true, reason: "Signature and content ID verified", manifest };
}

async function readJson(path, fallback = null) {
  try { return JSON.parse(await readFile(path, "utf8")); } catch (error) { if (error.code === "ENOENT") return fallback; throw error; }
}

async function writeJson(path, value) {
  await writeFile(path, `${JSON.stringify(value, null, 2)}\n`, "utf8");
}

export async function initProject(projectDirectory, options = {}) {
  const root = resolve(projectDirectory);
  const state = join(root, STATE_DIR);
  await mkdir(join(state, "releases"), { recursive: true });
  const identity = createPublisherIdentity();
  const config = { protocol: PROTOCOL_VERSION, domain: options.domain || `${basename(root)}.local`, entrypoint: options.entrypoint || "/index.html", version: "0.1.0", mirrors: [] };
  await writeJson(join(state, "config.json"), config);
  await writeFile(join(state, "publisher-private.pem"), identity.privateKey, { mode: 0o600 });
  await writeFile(join(state, "publisher-public.pem"), identity.publicKey);
  await writeJson(join(state, "registry.json"), { domain: config.domain, current: null, history: [] });
  return { root, config, publicKey: identity.publicKey };
}

export async function publishProject(projectDirectory, overrides = {}) {
  const rootDirectory = resolve(projectDirectory);
  const state = join(rootDirectory, STATE_DIR);
  const config = await readJson(join(state, "config.json"));
  if (!config) throw new Error("Not a Racore project. Run `racore init` first.");
  const registry = await readJson(join(state, "registry.json"), { domain: config.domain, current: null, history: [] });
  const bundle = await buildBundle(overrides.input || rootDirectory);
  const privateKey = await readFile(join(state, "publisher-private.pem"), "utf8");
  const publicKey = await readFile(join(state, "publisher-public.pem"), "utf8");
  const version = overrides.version || config.version;
  const manifest = createManifest({ domain: config.domain, version, root: bundle.root, parent: registry.current, commit: overrides.commit || "working-tree", entrypoint: config.entrypoint, mirrors: config.mirrors });
  const signed = signManifest(manifest, privateKey, publicKey);
  await writeJson(join(state, "releases", `${signed.manifestId}.json`), signed);
  await writeJson(join(state, "bundles", `${bundle.root}.json`), bundle).catch(async error => { if (error.code !== "ENOENT") throw error; await mkdir(join(state, "bundles")); await writeJson(join(state, "bundles", `${bundle.root}.json`), bundle); });
  registry.current = signed.manifestId;
  registry.history = [signed.manifestId, ...registry.history.filter(id => id !== signed.manifestId)];
  await writeJson(join(state, "registry.json"), registry);
  return { manifest: signed, bundle, dns: `_racore.${config.domain} TXT \"rcp=0.1 manifest=${signed.manifestId}\"` };
}

export async function listReleases(projectDirectory) {
  const state = join(resolve(projectDirectory), STATE_DIR);
  const registry = await readJson(join(state, "registry.json"));
  if (!registry) throw new Error("Not a Racore project.");
  const releases = [];
  for (const id of registry.history) {
    const manifest = await readJson(join(state, "releases", `${id}.json`));
    if (manifest) releases.push({ ...manifest, current: id === registry.current });
  }
  return releases;
}

export async function verifyRelease(projectDirectory, id = null) {
  const state = join(resolve(projectDirectory), STATE_DIR);
  const registry = await readJson(join(state, "registry.json"));
  if (!registry?.current) throw new Error("No published release.");
  const target = id || registry.current;
  const manifest = await readJson(join(state, "releases", `${target}.json`));
  if (!manifest) throw new Error(`Release ${target} not found.`);
  return { id: target, ...verifyManifest(manifest) };
}

export async function rollbackRelease(projectDirectory, id) {
  const state = join(resolve(projectDirectory), STATE_DIR);
  const registry = await readJson(join(state, "registry.json"));
  if (!registry?.history.includes(id)) throw new Error(`Release ${id} is not in this site's history.`);
  const verification = await verifyRelease(projectDirectory, id);
  if (!verification.valid) throw new Error(`Refusing rollback: ${verification.reason}`);
  registry.current = id;
  registry.history = [id, ...registry.history.filter(item => item !== id)];
  await writeJson(join(state, "registry.json"), registry);
  return { current: id, verification };
}
