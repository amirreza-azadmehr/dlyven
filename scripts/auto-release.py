#!/usr/bin/env python3
"""Decide whether this push should cut a GitHub Release, and the next tag.

App code (Go, UI, engines, build scripts) → yes, bump patch.
Docs / site / README / locales-only → no.

Also validates a pushed v*.*.* tag (it must be this commit and reachable from main)
and refuses anything that would reuse or collide with an existing version tag.

Env:
  GITHUB_REF   ref that triggered the run (refs/heads/main or refs/tags/vX.Y.Z)
  BEFORE       previous branch tip (github.event.before)
  GITHUB_SHA   commit being released
  HEAD_MSG     latest commit message ([skip release] bails out unless FORCE)
  FORCE        1/true to release even if only docs changed or [skip release] is set
  MAIN_REF     what counts as main (default origin/main)
  HAS_TAG_TOKEN true when a RELEASE_TAG_TOKEN secret exists (tags are then pushed by
               the "tag" job and the tag-triggered run builds, see release.yml)
  GITHUB_OUTPUT Actions output file (optional; still prints)

Exit status 1 means "refuse loudly": the run must not publish.
"""
from __future__ import annotations

import json
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


def consumed_tag_names() -> set[str]:
    """Tag names GitHub has reserved (immutable release attempted). They cannot be reused."""
    try:
        with open(".github/consumed-release-tags", encoding="utf-8") as fh:
            raw = fh.read()
    except OSError:
        return set()
    names: set[str] = set()
    for line in raw.splitlines():
        line = line.split("#", 1)[0].strip()
        if not line:
            continue
        names.add(line if line.startswith("v") else f"v{line}")
    return names


def next_version(tags: list[tuple[int, int, int]]) -> tuple[str, str]:
    """(last_tag or '', next X.Y.Z), skipping consumed immutable tag names.

    If skipping those names would leave a hole in the patch line (v1.0.0 then
    v1.0.3), jump to the next minor (v1.1.0) so the public history does not
    look like failed releases. After that, ordinary runs keep patch-bumping
    (v1.1.0 → v1.1.1).
    """
    consumed = consumed_tag_names()
    if tags:
        a, b, c = tags[-1]
        last = f"v{a}.{b}.{c}"
        c += 1
    else:
        last = ""
        a, b, c = 1, 0, 0
    while f"v{a}.{b}.{c}" in consumed:
        c += 1
    if last:
        la, lb, lc = (int(x) for x in last[1:].split("."))
        if (a, b) == (la, lb) and c > lc + 1:
            b += 1
            c = 0
            while f"v{a}.{b}.{c}" in consumed:
                c += 1
    return last, f"{a}.{b}.{c}"


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


class Refuse(Exception):
    """A release that must fail the run rather than be skipped."""


def tag_commit(tag: str) -> str:
    """Commit a tag points at ('' when the tag doesn't exist)."""
    try:
        return sh("git", "rev-list", "-n", "1", f"refs/tags/{tag}")
    except subprocess.CalledProcessError:
        return ""


def is_ancestor(a: str, b: str) -> bool:
    return subprocess.call(["git", "merge-base", "--is-ancestor", a, b], stderr=subprocess.DEVNULL) == 0


def write_out(should: bool, version: str, reason: str) -> None:
    print(reason)
    tagger = "app" if os.environ.get("HAS_TAG_TOKEN", "").strip().lower() == "true" else "builtin"
    print(f"should={str(should).lower()} version={version} tagger={tagger}")
    path = os.environ.get("GITHUB_OUTPUT")
    if not path:
        return
    with open(path, "a", encoding="utf-8") as fh:
        fh.write(f"should={str(should).lower()}\n")
        fh.write(f"version={version}\n")
        fh.write(f"tagger={tagger}\n")


def decide_tag(ref: str, sha: str, main: str) -> int:
    """A pushed v*.*.* tag: build exactly that commit, and only if it is releasable."""
    name = ref.removeprefix("refs/tags/")
    m = re.fullmatch(r"v(\d+)\.(\d+)\.(\d+)", name)
    if not m:
        raise Refuse(f"{name} is not a vX.Y.Z tag")
    version = f"{m[1]}.{m[2]}.{m[3]}"
    pointed = tag_commit(name)
    if pointed != sha:
        raise Refuse(f"tag {name} points at {pointed or 'nothing'}, but this run is for {sha}")
    if not exists(main) or not is_ancestor(sha, main):
        raise Refuse(f"{sha} is not on {main}; refusing to release a commit that never reached main")
    # Publishing an immutable release creates the git tag and re-fires this workflow.
    # A published release for this tag is that echo, not a new cut.
    if os.environ.get("GH_TOKEN"):
        viewed = subprocess.run(
            ["gh", "release", "view", name, "--json", "isDraft"],
            text=True, capture_output=True,
        )
        if viewed.returncode == 0:
            try:
                info = json.loads(viewed.stdout)
            except ValueError:
                info = {}
            if info.get("isDraft") is False:
                write_out(False, version, f"{name} is already published; skipping the tag-event echo")
                return 0
    write_out(True, version, f"tag {name} on {sha[:7]} (on main) → publish v{version}")
    return 0


def main() -> int:
    ref = (os.environ.get("GITHUB_REF") or "refs/heads/main").strip()
    force = os.environ.get("FORCE", "").strip().lower() in {"1", "true", "yes"}
    before = (os.environ.get("BEFORE") or "").strip()
    sha = (os.environ.get("GITHUB_SHA") or "").strip() or sh("git", "rev-parse", "HEAD")
    main_ref = (os.environ.get("MAIN_REF") or "origin/main").strip()
    msg = os.environ.get("HEAD_MSG") or ""
    if not msg.strip():
        try:
            msg = sh("git", "log", "-1", "--format=%B", sha)
        except subprocess.CalledProcessError:
            msg = ""

    if ref.startswith("refs/tags/"):
        return decide_tag(ref, sha, main_ref)
    if ref != "refs/heads/main":
        raise Refuse(f"releases are cut from main, not {ref}")

    tags = semver_tags()
    last, version = next_version(tags)
    if tags:
        a, b, c = tags[-1]
    else:
        a, b, c = 1, 0, 0

    # Never publish an older run's commit over a newer push; and never a commit that isn't main's.
    if exists(main_ref):
        tip = sh("git", "rev-parse", main_ref)
        if tip != sha:
            if is_ancestor(sha, tip):
                write_out(False, version, f"{sha[:7]} is no longer the tip of main ({tip[:7]}); the newer push decides")
                return 0
            raise Refuse(f"{sha} is not on {main_ref}")

    # The version we'd cut must be unused; an existing tag elsewhere is a collision, not a re-release.
    clash = tag_commit(f"v{version}")
    if clash and clash != sha:
        raise Refuse(f"tag v{version} already exists on {clash[:7]}; refusing to reuse a version")

    # An explicit force outranks the skip marker; an ordinary run honours it.
    if not force and re.search(r"\[skip[ -]?release\]", msg, re.I):
        write_out(False, version, "commit asked to skip the release (run with force to override)")
        return 0

    if last and tag_commit(last) == sha:
        # Already tagged: the tag-triggered run owns publishing it. Re-run that one.
        write_out(False, f"{a}.{b}.{c}", f"{last} is already this commit; publishing belongs to the {last} tag run")
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
    try:
        sys.exit(main())
    except Refuse as e:
        print(f"::error::{e}")
        sys.exit(1)
