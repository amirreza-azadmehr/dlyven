# Releasing

Releases are decided by `scripts/auto-release.py` (tests: `python3 -m unittest discover -s scripts -p 'test_*.py'`).

| Trigger | Result |
| --- | --- |
| Push to `main` with app changes | next patch version, built and published |
| Push to `main`, docs/site/locales only | nothing |
| Commit message contains `[skip release]` | nothing, unless the run is forced |
| Actions → Release → Run workflow (`force`) | next patch version, even with `[skip release]` or docs-only changes |
| Push of a `vX.Y.Z` tag | that exact commit is built and published, if it is on `main` |

Refused (the run fails, nothing is published): a version that is already published, a tag that
exists on a different commit, a tag whose commit is not on `main`, a run for a commit that is not
`main`'s tip (skipped, because the newer push decides), and runs from any other branch.

## Tag creation under repository rules

If a ruleset restricts tag creation, the default `GITHUB_TOKEN` cannot create the release tag
(`GH013: Cannot create ref due to creations being restricted`). Two ways out, neither weakens
other protections:

1. **Push the tag yourself** from an account that may bypass the rule:
   `git tag vX.Y.Z <commit> && git push origin vX.Y.Z`. The tag run does the rest.
2. **Automate it**: add a bypass actor to the *tag-creation* ruleset only (a dedicated GitHub App
   or a machine user), put that identity's token in the `RELEASE_TAG_TOKEN` secret. A push to
   `main` then only pushes the tag, and the tag-triggered run builds and publishes, so a release is
   never built twice. Leave branch rules, tag update/deletion rules and signing requirements as
   they are.
