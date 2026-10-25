# Releasing

Releases are decided by `scripts/auto-release.py` and guarded by `scripts/releasectl.py`
(tests: `python3 -m unittest discover -s scripts -p 'test_*.py'`).

| Trigger | Result |
| --- | --- |
| Push to `main` with app changes | next patch version, built and published |
| Push to `main`, docs/site/locales only | nothing |
| Commit message contains `[skip release]` | nothing, unless the run is forced |
| Actions → Release → Run workflow (`force`) | next patch version, even with `[skip release]` or docs-only changes |
| Push of a `vX.Y.Z` tag | that exact commit is built and published, if it is on `main` |
| Pull request (forks included) | unsigned build for every OS, nothing published, no secrets |

## Why tag creation fails

```
GH013: Repository rule violations found for refs/tags/v1.0.1
- Cannot create ref due to creations being restricted.
```

A tag ruleset with **Restrict creations** applies to everyone who is not on its bypass list.
The workflow's default `GITHUB_TOKEN` (the `github-actions[bot]` identity) cannot be put on a
bypass list, so it can never create a tag while that rule is active. Granting the job
`contents: write` does not help; the rule is evaluated separately from permissions. (The rule
that blocked the tag is shown at `https://github.com/amirreza-azadmehr/dlyven/rules?ref=refs%2Ftags%2Fv1.0.1`;
it is not visible through the REST/GraphQL API with a repo-scoped token, so check it in
Settings → Rules.)

Keep the rule. Give the *release identity*, and only it, permission to create tags.

## One-time setup (repository owner)

1. **Create the token** (a fine-grained personal access token owned by the account that is an
   admin of this repo, or a GitHub App installed only on this repo):
   - Resource owner: `amirreza-azadmehr`; Repository access: *Only select repositories → dlyven*.
   - Repository permissions: **Contents: Read and write**, **Metadata: Read** (nothing else).
   - Expiration: at most one year; set a reminder to rotate it. A GitHub App installation
     token avoids rotation by a person but needs a small token-minting step in the workflow.
2. **Store it**: Settings → Secrets and variables → Actions → *New repository secret*, name
   **`RELEASE_TAG_TOKEN`**. (Not as an environment-less variable, not in code.)
3. **Authorize it on the tag ruleset**: Settings → Rules → Rulesets → the ruleset that
   restricts tag creation → *Bypass list* → *Add bypass*:
   - for a PAT owned by you: the **Repository admin** role, mode **Always** (the token acts as
     you, so the role covers it); or
   - for a GitHub App: that app, mode **Always**.

   Do not remove *Restrict creations*, *Restrict updates* or *Restrict deletions*, and do not
   widen the target pattern. Branch rules are untouched.
4. Optionally restrict the secret to the `main` branch with an environment (the `tag` job would
   then declare `environment:`); the workflow already never exposes it to pull requests.

## What happens once it is set

- Push to `main` → **Decide** picks the version → **Tag** job pushes `vX.Y.Z` with
  `RELEASE_TAG_TOKEN` (a PAT/App push, unlike `GITHUB_TOKEN`, triggers workflows) → the
  tag-triggered run builds every OS and publishes. The `main` run itself builds nothing, so a
  release is never built twice.
- `scripts/releasectl.py tag` never moves, deletes or overwrites a tag, then reads the remote
  back and fails unless the tag points at the intended commit.
- **GitHub Release** runs `releasectl.py verify` before publishing: no verified tag, no
  release. It also refuses to touch an existing release; a pre-existing *draft* is adopted only
  when you push its tag deliberately, never by an automatic `main` run.
- Without `RELEASE_TAG_TOKEN` the old behaviour applies (tag with `GITHUB_TOKEN` after the
  build). That only works while no ruleset restricts tag creation; otherwise it fails with
  an explanation and nothing is published.

## Failure messages

| Message | Meaning |
| --- | --- |
| `the tag ruleset rejected RELEASE_TAG_TOKEN` | token owner/app is not on the ruleset bypass list |
| `the tag ruleset forbids the default GITHUB_TOKEN` | secret not set (or empty) |
| `the token was refused … invalid, expired` | rotate the token / check Contents: write on this repo |
| `tag vX.Y.Z already exists on … not …` | an existing tag is never moved; pick a new version |
| `a draft release vX.Y.Z already exists` | publish or delete the draft yourself, or push the tag on purpose |
| `vX.Y.Z is already published` | releases are immutable; cut a new version |

## Releasing by hand (no token)

From an account on the bypass list: `git tag vX.Y.Z <commit-on-main> && git push origin vX.Y.Z`.
The tag run validates that the commit is on `main`, builds and publishes it.

## Versions and the existing draft

Tags: `v1.0.0` only. There is a stale draft `v1.0.1` (no git tag). An automatic run from the
current `main` would compute `v1.0.1`, and the guard will refuse it because that draft exists:
publish, edit or delete the draft yourself first, or release `v1.0.2` by pushing that tag
(skipping `v1.0.1` is then your explicit choice).
