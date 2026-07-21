from __future__ import annotations

import argparse
import hashlib
import io
import json
import re
import sys
import zipfile
from pathlib import Path

import httpx

DAEMON = "http://127.0.0.1:47831"
DOMAIN_RE = re.compile(r"^(?=.{1,253}$)(?!-)(?:[a-z0-9-]{1,63}\.)+[a-z]{2,63}$")


def api(method: str, path: str, **kwargs):
    try:
        response = httpx.request(method, f"{DAEMON}{path}", timeout=120, **kwargs)
    except httpx.HTTPError as exc:
        raise RuntimeError("Racore Desktop is not running. Open the app and retry.") from exc
    if not response.is_success:
        try:
            detail = response.json().get("detail", response.text)
        except ValueError:
            detail = response.text
        raise RuntimeError(str(detail))
    return response.json()


def normalize_domain(value: str) -> str:
    domain = value.lower().strip().strip(".")
    if not DOMAIN_RE.match(domain):
        raise RuntimeError("Use a valid fully-qualified domain such as app.example.com")
    return domain


def bundle(directory: Path) -> tuple[bytes, str, int, int]:
    root = directory.resolve()
    if not root.is_dir():
        raise RuntimeError(f"Build directory does not exist: {root}")
    files = sorted(path for path in root.rglob("*") if path.is_file())
    if not files:
        raise RuntimeError("Build directory is empty")
    if not (root / "index.html").exists():
        raise RuntimeError("Build directory must contain index.html")
    index = []
    output = io.BytesIO()
    with zipfile.ZipFile(output, "w", zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
        for path in files:
            relative = path.relative_to(root).as_posix()
            data = path.read_bytes()
            index.append({"path": relative, "size": len(data), "sha256": hashlib.sha256(data).hexdigest()})
            archive.writestr(relative, data)
    canonical = json.dumps(index, sort_keys=True, separators=(",", ":")).encode()
    return output.getvalue(), hashlib.sha256(canonical).hexdigest(), len(files), sum(item["size"] for item in index)


def cmd_status(_):
    return api("GET", "/health")


def cmd_domains(_):
    return api("GET", "/v1/authority/domains")


def cmd_claim(args):
    domain = normalize_domain(args.domain)
    status = api("GET", f"/v1/authority/domains/{domain}/available")
    if not status["available"]:
        raise RuntimeError(f"{domain} is already claimed on the active known mesh")
    return api("POST", "/v1/authority/domains", json={"domain": domain})


def cmd_publish(args):
    domain = normalize_domain(args.domain)
    archive, content_root, file_count, total_size = bundle(Path(args.build))
    known = {item["domain"] for item in api("GET", "/v1/authority/domains")}
    if domain not in known:
        availability = api("GET", f"/v1/authority/domains/{domain}/available")
        if not availability["available"]:
            raise RuntimeError(f"{domain} belongs to another authority on the active known mesh")
        api("POST", "/v1/authority/domains", json={"domain": domain})
    uploaded = api("POST", "/v1/ipfs/add", files={"file": (f"{domain}-{args.version}.zip", archive, "application/zip")})
    return api("POST", f"/v1/authority/domains/{domain}/releases", json={"version": args.version, "cid": uploaded["cid"], "contentRoot": content_root, "entrypoint": "index.html", "files": file_count, "size": total_size})


def cmd_releases(args):
    return api("GET", f"/v1/authority/domains/{normalize_domain(args.domain)}/releases")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="racore", description="Publish verified frontend builds to Racore/IPFS")
    parser.add_argument("--json", action="store_true")
    commands = parser.add_subparsers(dest="command", required=True)
    commands.add_parser("status", help="Check daemon, mesh, and IPFS").set_defaults(run=cmd_status)
    commands.add_parser("domains", help="List domains controlled by this device").set_defaults(run=cmd_domains)
    claim = commands.add_parser("claim", help="Claim an available domain")
    claim.add_argument("domain")
    claim.set_defaults(run=cmd_claim)
    publish = commands.add_parser("publish", help="Publish a JavaScript framework build directory")
    publish.add_argument("--domain", required=True)
    publish.add_argument("--build", required=True)
    publish.add_argument("--version", required=True)
    publish.set_defaults(run=cmd_publish)
    releases = commands.add_parser("releases", help="List signed releases")
    releases.add_argument("domain")
    releases.set_defaults(run=cmd_releases)
    return parser


def main() -> None:
    args = build_parser().parse_args()
    try:
        result = args.run(args)
        if args.json or not isinstance(result, list):
            print(json.dumps(result, indent=2))
        elif not result:
            print("No records found.")
        else:
            for item in result:
                print(item.get("domain") or f"v{item.get('version')}  {item.get('cid')}")
    except RuntimeError as exc:
        print(f"Racore error: {exc}", file=sys.stderr)
        raise SystemExit(1) from exc


if __name__ == "__main__":
    main()
