# Releasing

Releases are decided by `scripts/auto-release.py` and guarded by `scripts/releasectl.py`
(tests: `python3 -m unittest discover -b -s scripts -p 'test_*.py'`).

This repository has **immutable releases** enabled
(Settings → General → Releases → *Enable release immutability*). That is a
repository setting, not an account-wide lock. After a release is published, its
files and git tag cannot be changed. Keep it on.

| Trigger | Result |
| --- | --- |
| Push to `main` with app changes | next patch version, built and published |
| Push to `main`, docs/site/locales only | nothing |
| Commit message contains `[skip release]` | nothing, unless the run is forced |
| Actions → Release → Run workflow (`force`) | next patch version, even with `[skip release]` or docs-only changes |
| Pull request (forks included) | unsigned build for every OS, nothing published, no secrets |

Do not `git push` a version tag. GitHub rejects it with GH013 while immutability
is on; the tag is created only when the draft release is published.

## Why `git push origin vX.Y.Z` fails

```
GH013: Repository rule violations found for refs/tags/v1.0.1
- Cannot create ref due to creations being restricted.
```

That message is from **immutable releases**, not from a Ruleset. The Rulesets
pages are empty on purpose. Do not create a tag ruleset, and do not turn
immutability off.

The workflow uploads every asset to a **draft**, then publishes. Publishing is
what creates `vX.Y.Z`. After that the files and tag are locked.

## What the workflow does

- Push to `main` (app changes) or **Run workflow** → **Decide** picks the version
  → every OS builds → assets go on a draft → the draft is published → the git
  tag appears → `releasectl.py verify` checks the tag points at the built commit.
- An existing published release is never touched. A pre-existing *draft* for the
  same version is reused only if that provably loses nothing: it targets the
  commit being released, holds no assets, has no hand-written notes (only
  GitHub's "Full Changelog" line), and no conflicting tag exists. Otherwise the
  run stops with the reasons and leaves the draft untouched.
- Publishing also creates the git tag, which can fire a second tag-event run.
  That run sees the release already published and skips.

## Failure messages

| Message | Meaning |
| --- | --- |
| `tag vX.Y.Z already exists on … not …` | an existing tag is never moved; pick a new version |
| `a draft release vX.Y.Z already exists and cannot be reused safely: …` | the listed reasons (other commit, assets, hand-written notes, conflicting tag) would lose data; publish or delete the draft yourself, then re-run |
| `vX.Y.Z is already published` | releases are immutable; cut a new version |
| `tag_name was used by an immutable release` | that tag name is permanently reserved; add it to `.github/consumed-release-tags` and cut the next patch |

## Versions

Published: `v1.0.0`. `v1.0.1` was consumed by a failed immutable publish and cannot
be reused. The next cut is `v1.0.2`. Do not create the tag by hand, and do not
publish an empty GitHub Release: once published, binaries cannot be added.
