#!/usr/bin/env python3
"""Release safety checks that must run before and after the tag exists.

    releasectl.py tag      create vX.Y.Z on SHA and push it, then verify the remote
    releasectl.py verify   fail unless the remote tag vX.Y.Z points at SHA
    releasectl.py guard    fail if a release for vX.Y.Z already exists (see below)

Env:
  VERSION        X.Y.Z (no leading v)
  SHA            commit the tag must point at
  TAG_SOURCE     "app" when RELEASE_TAG_TOKEN pushes the tag, "builtin" for GITHUB_TOKEN
  HAS_TAG_TOKEN  "true" when the RELEASE_TAG_TOKEN secret is non-empty (tag, source=app)
  REMOTE         git remote to push to (default origin)
  GH_TOKEN       guard only: token for `gh release view` (read-only calls; the guard never
                 creates, edits or deletes anything)

An existing tag is never moved, deleted or overwritten. An existing published release is
never touched. An existing draft is reused only when it provably loses nothing: it targets
the commit being released, holds no assets, has no hand-written notes, and no conflicting tag
exists. Anything else is a failure that says why and leaves the draft untouched.
"""
from __future__ import annotations

import json
import os
import re
import subprocess
import sys


class Fail(Exception):
    pass


def run(*args: str) -> subprocess.CompletedProcess:
    return subprocess.run(args, text=True, capture_output=True)


def need(env, name: str) -> str:
    v = (env.get(name) or "").strip()
    if not v:
        raise Fail(f"{name} is not set")
    return v


def parse(env) -> tuple[str, str, str]:
    version = need(env, "VERSION")
    if not re.fullmatch(r"\d+\.\d+\.\d+", version):
        raise Fail(f"VERSION {version!r} is not X.Y.Z")
    sha = need(env, "SHA")
    if not re.fullmatch(r"[0-9a-f]{40}", sha):
        raise Fail(f"SHA {sha!r} is not a full commit hash")
    return f"v{version}", version, sha


def remote_tag_commit(remote: str, tag: str) -> str:
    """Commit the remote tag points at (peeled if annotated), '' if it doesn't exist."""
    p = run("git", "ls-remote", "--tags", remote, f"refs/tags/{tag}", f"refs/tags/{tag}^{{}}")
    if p.returncode != 0:
        raise Fail(f"cannot read tags from {remote}: {p.stderr.strip()}")
    plain = peeled = ""
    for line in p.stdout.splitlines():
        sha, ref = line.split("\t")
        if ref.endswith("^{}"):
            peeled = sha
        else:
            plain = sha
    return peeled or plain


def explain_push_failure(stderr: str, source: str) -> str:
    low = stderr.lower()
    if "gh013" in low or "repository rule" in low or "creations being restricted" in low:
        if source == "app":
            return ("the tag ruleset rejected RELEASE_TAG_TOKEN: add the token's owner (or its GitHub App) "
                    "as a bypass actor on the tag-creation ruleset (docs/RELEASING.md)")
        return ("the tag ruleset forbids the default GITHUB_TOKEN from creating tags: set the "
                "RELEASE_TAG_TOKEN secret and authorize its owner as a bypass actor (docs/RELEASING.md)")
    if any(s in low for s in ("bad credentials", "authentication failed", "invalid username or password",
                              "could not read username", "403", "permission to", "denied to", "401")):
        return ("the token was refused: RELEASE_TAG_TOKEN is invalid, expired, or lacks write access "
                "to Contents on this repository (docs/RELEASING.md)")
    return "git push failed"


def cmd_tag(env) -> int:
    tag, _, sha = parse(env)
    source = (env.get("TAG_SOURCE") or "").strip()
    if source not in {"app", "builtin"}:
        raise Fail(f"TAG_SOURCE must be app or builtin, got {source!r}")
    if source == "app" and (env.get("HAS_TAG_TOKEN") or "").strip().lower() != "true":
        raise Fail("RELEASE_TAG_TOKEN is empty or missing, but this run was set up to use it")
    remote = (env.get("REMOTE") or "origin").strip()

    existing = remote_tag_commit(remote, tag)
    if existing:
        if existing != sha:
            raise Fail(f"{tag} already exists on {existing[:7]}, not {sha[:7]}; existing tags are never moved")
        print(f"{tag} already exists on {sha[:7]}")
        return 0

    local = run("git", "rev-parse", "-q", "--verify", f"refs/tags/{tag}^{{commit}}")
    if local.returncode == 0 and local.stdout.strip() != sha:
        raise Fail(f"a local {tag} exists on {local.stdout.strip()[:7]}, not {sha[:7]}")
    if local.returncode != 0:
        mk = run("git", "tag", tag, sha)
        if mk.returncode != 0:
            raise Fail(f"cannot create tag {tag} on {sha[:7]}: {mk.stderr.strip()}")
    push = run("git", "push", remote, f"refs/tags/{tag}")
    if push.returncode != 0:
        sys.stderr.write(push.stderr)
        raise Fail(f"could not create tag {tag}: {explain_push_failure(push.stderr, source)}")
    return cmd_verify(env)


def cmd_verify(env) -> int:
    tag, _, sha = parse(env)
    remote = (env.get("REMOTE") or "origin").strip()
    got = remote_tag_commit(remote, tag)
    if got != sha:
        raise Fail(f"{tag} is {'missing on the remote' if not got else 'on ' + got[:7]}, expected {sha[:7]}; not publishing")
    print(f"{tag} → {sha[:7]} verified on {remote}")
    return 0


AUTO_BODY = re.compile(r"\s*(\*\*Full Changelog\*\*:\s*\S+)?\s*")


def release_info(tag: str, runner=run):
    """The release for tag as a dict (gh JSON), or None. Anything gh can't answer is an error,
    never "no release": a false 'none' would let a run overwrite something."""
    p = runner("gh", "release", "view", tag, "--json", "isDraft,targetCommitish,body,assets")
    if p.returncode == 0:
        try:
            return json.loads(p.stdout)
        except ValueError:
            raise Fail(f"cannot parse the release data for {tag}")
    if "release not found" in p.stderr.lower():
        return None
    raise Fail(f"cannot tell whether {tag} has a release: {p.stderr.strip() or 'gh failed'}")


def draft_problems(info: dict, tag: str, sha: str, remote: str) -> list[str]:
    """Reasons the existing draft cannot be filled in without losing or mixing information.
    An empty list means: same commit, nothing in it yet, nothing hand-written to overwrite."""
    out = []
    target = (info.get("targetCommitish") or "").strip()
    if re.fullmatch(r"[0-9a-f]{40}", target):
        if target != sha:
            out.append(f"it targets {target[:7]}, but this release is {sha[:7]}")
    else:
        p = run("git", "ls-remote", remote, f"refs/heads/{target}") if target else None
        tip = p.stdout.split("\t")[0] if p is not None and p.returncode == 0 and p.stdout else ""
        if tip != sha:
            out.append(f"it targets '{target or '?'}', which is not {sha[:7]}")
    names = [a.get("name", "?") for a in info.get("assets") or []]
    if names:
        out.append(f"it already holds {len(names)} asset(s) ({', '.join(names[:4])}{'…' if len(names) > 4 else ''}) that would be mixed with or replaced by this build")
    if not AUTO_BODY.fullmatch(info.get("body") or ""):
        out.append("its notes were written by hand and would be replaced")
    existing = remote_tag_commit(remote, tag)
    if existing and existing != sha:
        out.append(f"the tag {tag} already exists on {existing[:7]}")
    return out


def cmd_guard(env, runner=run) -> int:
    tag, _, sha = parse(env)
    remote = (env.get("REMOTE") or "origin").strip()
    existing = remote_tag_commit(remote, tag)
    if existing and existing != sha:
        raise Fail(f"tag {tag} already exists on {existing[:7]}, not {sha[:7]}; existing tags are never moved")
    info = release_info(tag, runner)
    if info is None:
        print(f"no release for {tag}")
        return 0
    if not info.get("isDraft"):
        raise Fail(f"{tag} is already published; releases are immutable, cut a new version")
    problems = draft_problems(info, tag, sha, remote)
    if problems:
        raise Fail(f"a draft release {tag} already exists and cannot be reused safely: " + "; ".join(problems) +
                   f". The draft is untouched. Publish or delete it yourself (or release a different version), then re-run")
    print(f"draft {tag} is empty, targets {sha[:7]} and has no hand-written notes; it will be filled in")
    return 0


def main(argv=None, env=None, runner=run) -> int:
    argv = sys.argv[1:] if argv is None else argv
    env = os.environ if env is None else env
    cmds = {"tag": cmd_tag, "verify": cmd_verify}
    try:
        if len(argv) != 1:
            raise Fail("usage: releasectl.py tag|verify|guard")
        if argv[0] == "guard":
            return cmd_guard(env, runner)
        if argv[0] not in cmds:
            raise Fail(f"unknown command {argv[0]!r}")
        return cmds[argv[0]](env)
    except Fail as e:
        print(f"::error::{e}")
        return 1


if __name__ == "__main__":
    sys.exit(main())
