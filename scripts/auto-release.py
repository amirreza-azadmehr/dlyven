#!/usr/bin/env python3
"""Decide whether this push should cut a GitHub Release, and the next tag.

App code (Go, UI, engines, build scripts) → yes, bump patch.
Docs / site / README / locales-only → no.

Env:
  BEFORE       previous branch tip (github.event.before)
  GITHUB_SHA   commit being released
  HEAD_MSG     latest commit message ([skip release] bails out)
  FORCE        1/true to release even if only docs changed
  GITHUB_OUTPUT Actions output file (optional; still prints)
"""
from __future__ import annotations

import os
import re
import subprocess
import sys

APP_EXACT = {
    "go.mod",
    "go.sum",
    "wails.json",
    "main.go",
    "app.go",
    "frontend/index.html",
    "frontend/package.json",
    "frontend/package-lock.json",
    "frontend/vite.config.ts",
    "frontend/tailwind.config.js",
    "frontend/postcss.config.js",
    "frontend/tsconfig.json",
    "scripts/build.sh",
    "scripts/build.ps1",
}
APP_PREFIX = (
    "internal/",
    "frontend/src/",
    "build/",
)
SKIP_PREFIX = (
    "frontend/src/locales/",
    "docs/",
    "site/",
    ".github/",
)


def sh(*args: str) -> str:
    return subprocess.check_output(args, text=True).strip()


def exists(rev: str) -> bool:
    if not rev or set(rev) == {"0"}:
        return False
    try:
        subprocess.check_call(["git", "cat-file", "-e", f"{rev}^{{commit}}"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        return True
    except subprocess.CalledProcessError:
        return False


def semver_tags() -> list[tuple[int, int, int]]:
    raw = subprocess.check_output(["git", "tag", "-l", "v*"], text=True)
    found: list[tuple[int, int, int]] = []
    for line in raw.split():
        m = re.fullmatch(r"v(\d+)\.(\d+)\.(\d+)", line.strip())
        if m:
            found.append((int(m[1]), int(m[2]), int(m[3])))
    found.sort()
    return found


def is_app(path: str) -> bool:
    path = path.replace("\\", "/")
    if path.endswith(".md"):
        return False
    for skip in SKIP_PREFIX:
        if path == skip.rstrip("/") or path.startswith(skip):
            return False
    if path.endswith(".go") or path in APP_EXACT:
        return True
    return any(path.startswith(p) for p in APP_PREFIX)


def write_out(should: bool, version: str, reason: str) -> None:
    print(reason)
    print(f"should={str(should).lower()} version={version}")
    path = os.environ.get("GITHUB_OUTPUT")
    if not path:
        return
    with open(path, "a", encoding="utf-8") as fh:
        fh.write(f"should={str(should).lower()}\n")
        fh.write(f"version={version}\n")


def main() -> int:
    force = os.environ.get("FORCE", "").strip().lower() in {"1", "true", "yes"}
    before = (os.environ.get("BEFORE") or "").strip()
    sha = (os.environ.get("GITHUB_SHA") or "").strip() or sh("git", "rev-parse", "HEAD")
    msg = os.environ.get("HEAD_MSG") or ""
    if not msg.strip():
        try:
            msg = sh("git", "log", "-1", "--format=%B", sha)
        except subprocess.CalledProcessError:
            msg = ""

    tags = semver_tags()
    if tags:
        a, b, c = tags[-1]
        last = f"v{a}.{b}.{c}"
        bumped = f"{a}.{b}.{c + 1}"
    else:
        last = ""
        bumped = "1.0.0"

    last_sha = sh("git", "rev-list", "-n", "1", last) if last else ""
    on_latest_tag = bool(last) and last_sha == sha
    version = f"{a}.{b}.{c}" if on_latest_tag else bumped

    if re.search(r"\[skip[ -]?release\]", msg, re.I):
        write_out(False, version, "commit asked to skip the release")
        return 0

    if on_latest_tag:
        write_out(True, version, f"{last} is this commit — publish v{version}")
        return 0

    if force:
        write_out(True, version, f"forced → v{version}")
        return 0

    base = before if exists(before) else last
    if not base:
        write_out(True, version, f"no previous tag → v{version}")
        return 0

    diff = sh("git", "diff", "--name-only", base, sha)
    files = [ln for ln in diff.splitlines() if ln]
    app_files = [p for p in files if is_app(p)]
    if not app_files:
        write_out(False, version, f"no app changes since {base} ({len(files)} other files)")
        return 0

    print("app files:")
    for p in app_files:
        print(f"  {p}")
    write_out(True, version, f"{len(app_files)} app file(s) since {base} → v{version}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
