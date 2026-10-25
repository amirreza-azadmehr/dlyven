"""Tests for scripts/releasectl.py (real git repos; a bare remote that can refuse tag creation)
and for the security-relevant shape of .github/workflows/release.yml.

    python3 -m unittest discover -b -s scripts -p 'test_*.py'
"""
import importlib.util
import json
import pathlib
import subprocess
import tempfile
import unittest
from types import SimpleNamespace

try:
    import yaml
except ImportError:  # the shape tests need PyYAML; the git tests don't
    yaml = None

HERE = pathlib.Path(__file__).parent
spec = importlib.util.spec_from_file_location("releasectl", HERE / "releasectl.py")
rc = importlib.util.module_from_spec(spec)
spec.loader.exec_module(rc)

REJECT_TAGS = """#!/bin/sh
while read old new ref; do
  case "$ref" in
    refs/tags/*)
      echo "remote: error: GH013: Repository rule violations found for $ref." >&2
      echo "remote: - Cannot create ref due to creations being restricted." >&2
      exit 1;;
  esac
done
"""


def git(cwd, *args):
    return subprocess.check_output(["git", *args], cwd=cwd, text=True).strip()


class Env:
    def __init__(self, test, reject_tags=False):
        self.tmp = tempfile.TemporaryDirectory()
        test.addCleanup(self.tmp.cleanup)
        root = pathlib.Path(self.tmp.name)
        self.remote = root / "remote.git"
        self.work = root / "work"
        subprocess.check_call(["git", "init", "-q", "--bare", "-b", "main", str(self.remote)])
        subprocess.check_call(["git", "clone", "-q", str(self.remote), str(self.work)], stderr=subprocess.DEVNULL)
        for k, v in (("user.email", "t@example.com"), ("user.name", "t"), ("commit.gpgsign", "false"), ("tag.gpgsign", "false")):
            git(self.work, "config", k, v)
        (self.work / "f").write_text("1")
        git(self.work, "add", "f")
        git(self.work, "commit", "-q", "-m", "one")
        git(self.work, "push", "-q", "origin", "HEAD:main")
        self.sha = git(self.work, "rev-parse", "HEAD")
        if reject_tags:
            hook = self.remote / "hooks" / "pre-receive"
            hook.write_text(REJECT_TAGS)
            hook.chmod(0o755)
        self.cwd = test.enterContext(_chdir(self.work))

    def env(self, **kw):
        base = {"VERSION": "1.0.2", "SHA": self.sha, "TAG_SOURCE": "app", "HAS_TAG_TOKEN": "true", "REMOTE": "origin"}
        base.update(kw)
        return base

    def remote_tags(self):
        return git(self.remote, "tag", "-l")


class _chdir:
    def __init__(self, p):
        self.p = p

    def __enter__(self):
        import os
        self.old = os.getcwd()
        os.chdir(self.p)

    def __exit__(self, *a):
        import os
        os.chdir(self.old)


class TagTest(unittest.TestCase):
    def test_creates_and_verifies_tag(self):
        e = Env(self)
        self.assertEqual(rc.main(["tag"], e.env()), 0)
        self.assertEqual(e.remote_tags(), "v1.0.2")
        self.assertEqual(git(e.remote, "rev-list", "-n1", "v1.0.2"), e.sha)

    def test_ruleset_rejection_fails_and_creates_nothing(self):
        e = Env(self, reject_tags=True)
        for source in ("app", "builtin"):
            self.assertEqual(rc.main(["tag"], e.env(TAG_SOURCE=source)), 1)
        self.assertEqual(e.remote_tags(), "")
        self.assertEqual(rc.main(["verify"], e.env()), 1)  # so nothing would be published

    def test_failure_messages_name_the_fix(self):
        gh013 = "remote: error: GH013: Repository rule violations found\nCannot create ref due to creations being restricted."
        self.assertIn("bypass actor", rc.explain_push_failure(gh013, "app"))
        self.assertIn("RELEASE_TAG_TOKEN secret", rc.explain_push_failure(gh013, "builtin"))
        for bad in ("remote: Invalid username or password.", "fatal: Authentication failed for", "HTTP 403", "Permission to x denied to y"):
            self.assertIn("invalid, expired", rc.explain_push_failure(bad, "app"))
        self.assertEqual(rc.explain_push_failure("boom", "app"), "git push failed")

    def test_existing_tag_same_commit_is_a_noop(self):
        e = Env(self)
        git(e.work, "tag", "v1.0.2")
        git(e.work, "push", "-q", "origin", "v1.0.2")
        self.assertEqual(rc.main(["tag"], e.env()), 0)

    def test_existing_tag_on_other_commit_is_never_moved(self):
        e = Env(self)
        git(e.work, "tag", "v1.0.2")
        git(e.work, "push", "-q", "origin", "v1.0.2")
        (e.work / "f").write_text("2")
        git(e.work, "commit", "-qam", "two")
        new = git(e.work, "rev-parse", "HEAD")
        git(e.work, "push", "-q", "origin", "HEAD:main")
        self.assertEqual(rc.main(["tag"], e.env(SHA=new)), 1)
        self.assertEqual(git(e.remote, "rev-list", "-n1", "v1.0.2"), e.sha)

    def test_annotated_tag_is_peeled(self):
        e = Env(self)
        git(e.work, "tag", "-a", "-m", "m", "v1.0.2")
        git(e.work, "push", "-q", "origin", "v1.0.2")
        self.assertEqual(rc.main(["verify"], e.env()), 0)

    def test_missing_or_empty_token_is_refused_before_pushing(self):
        e = Env(self)
        self.assertEqual(rc.main(["tag"], e.env(HAS_TAG_TOKEN="false")), 1)
        self.assertEqual(rc.main(["tag"], e.env(HAS_TAG_TOKEN="")), 1)
        self.assertEqual(e.remote_tags(), "")
        # the default token needs no secret
        self.assertEqual(rc.main(["tag"], e.env(TAG_SOURCE="builtin", HAS_TAG_TOKEN="")), 0)

    def test_bad_inputs(self):
        e = Env(self)
        for bad in ({"VERSION": "1.0"}, {"VERSION": "v1.0.2"}, {"SHA": "abc"}, {"SHA": ""}, {"TAG_SOURCE": "magic"}):
            self.assertEqual(rc.main(["tag"], e.env(**bad)), 1, bad)
        self.assertEqual(rc.main(["nope"], e.env()), 1)
        self.assertEqual(e.remote_tags(), "")

    def test_verify_fails_when_tag_missing_or_elsewhere(self):
        e = Env(self)
        self.assertEqual(rc.main(["verify"], e.env()), 1)
        git(e.work, "tag", "v1.0.2")
        git(e.work, "push", "-q", "origin", "v1.0.2")
        self.assertEqual(rc.main(["verify"], e.env(SHA="1" * 40)), 1)
        self.assertEqual(rc.main(["verify"], e.env()), 0)


REAL_DRAFT = {  # the stale v1.0.1 draft, as GitHub reports it
    "isDraft": True, "targetCommitish": "8915822a5243ab149a5f3d6a1ee4242d217576c2",
    "body": "**Full Changelog**: https://github.com/amirreza-azadmehr/dlyven/compare/v1.0.0...v1.0.1",
    "assets": [{"name": n} for n in ("Dlyven-linux-amd64.sha256", "Dlyven-linux-arm64.sha256",
                                      "Dlyven-macos-universal.zip.sha256", "Dlyven-macos-universal.dmg.sha256")],
}


class FakeGh:
    """Stands in for gh and records every call, so tests can prove the guard is read-only."""

    def __init__(self, info=None, stderr="", code=0):
        self.calls = []
        self.info, self.stderr, self.code = info, stderr, code

    def __call__(self, *args):
        self.calls.append(args)
        if self.info is None and not self.code:
            return SimpleNamespace(returncode=1, stdout="", stderr="release not found")
        out = json.dumps(self.info) if self.info is not None else ""
        return SimpleNamespace(returncode=self.code, stdout=out, stderr=self.stderr)


def draft(**kw):
    return {**{"isDraft": True, "targetCommitish": "", "body": "", "assets": []}, **kw}


class GuardTest(unittest.TestCase):
    def guard(self, e, gh, **env):
        return rc.main(["guard"], e.env(**env), gh)

    def assertReadOnly(self, gh):
        for call in gh.calls:
            self.assertEqual(call[:3], ("gh", "release", "view"), call)

    def test_no_release_passes(self):
        e = Env(self)
        gh = FakeGh()
        self.assertEqual(self.guard(e, gh), 0)
        self.assertReadOnly(gh)

    def test_published_release_always_refused(self):
        e = Env(self)
        self.assertEqual(self.guard(e, FakeGh({"isDraft": False})), 1)

    def test_the_real_stale_draft_is_refused_and_left_alone(self):
        e = Env(self)
        gh = FakeGh(REAL_DRAFT)
        self.assertEqual(self.guard(e, gh), 1)
        self.assertReadOnly(gh)
        problems = rc.draft_problems(REAL_DRAFT, "v1.0.1", e.sha, "origin")
        text = "; ".join(problems)
        self.assertIn("targets 8915822", text)
        self.assertIn("4 asset(s)", text)

    def test_empty_draft_on_this_commit_is_reused(self):
        e = Env(self)
        gh = FakeGh(draft(targetCommitish=e.sha, body="**Full Changelog**: https://x/compare/a...b"))
        self.assertEqual(self.guard(e, gh), 0)
        self.assertReadOnly(gh)

    def test_draft_with_branch_target_must_resolve_to_this_commit(self):
        e = Env(self)
        self.assertEqual(self.guard(e, FakeGh(draft(targetCommitish="main"))), 0)
        (e.work / "f").write_text("2")
        git(e.work, "commit", "-qam", "two")
        git(e.work, "push", "-q", "origin", "HEAD:main")  # main moved past e.sha
        self.assertEqual(self.guard(e, FakeGh(draft(targetCommitish="main"))), 1)
        self.assertEqual(self.guard(e, FakeGh(draft(targetCommitish="no-such-branch"))), 1)

    def test_draft_with_assets_or_handwritten_notes_is_refused(self):
        e = Env(self)
        with_asset = draft(targetCommitish=e.sha, assets=[{"name": "Dlyven-Linux-x64"}])
        notes = draft(targetCommitish=e.sha, body="## Highlights\nsomething I wrote")
        for d in (with_asset, notes):
            gh = FakeGh(d)
            self.assertEqual(self.guard(e, gh), 1)
            self.assertReadOnly(gh)

    def test_matching_tag_with_draft_is_fine(self):
        e = Env(self)
        git(e.work, "tag", "v1.0.2")
        git(e.work, "push", "-q", "origin", "v1.0.2")
        self.assertEqual(self.guard(e, FakeGh(draft(targetCommitish=e.sha))), 0)

    def test_conflicting_tag_is_refused_with_or_without_a_draft(self):
        e = Env(self)
        git(e.work, "tag", "v1.0.2")
        git(e.work, "push", "-q", "origin", "v1.0.2")
        other = "1" * 40
        self.assertEqual(self.guard(e, FakeGh(draft(targetCommitish=other)), SHA=other), 1)
        self.assertEqual(self.guard(e, FakeGh(), SHA=other), 1)
        self.assertEqual(git(e.remote, "rev-list", "-n1", "v1.0.2"), e.sha)

    def test_missing_tag_with_clean_draft_is_fine(self):
        e = Env(self)
        self.assertEqual(e.remote_tags(), "")
        self.assertEqual(self.guard(e, FakeGh(draft(targetCommitish=e.sha))), 0)

    def test_gh_errors_fail_closed(self):
        e = Env(self)
        self.assertEqual(self.guard(e, FakeGh(stderr="HTTP 401: Bad credentials", code=1)), 1)
        self.assertEqual(self.guard(e, FakeGh(stderr="", code=1)), 1)

    def test_guard_requires_a_commit(self):
        e = Env(self)
        self.assertEqual(self.guard(e, FakeGh(), SHA=""), 1)


@unittest.skipUnless(yaml, "PyYAML not installed")
class WorkflowShapeTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.text = (HERE.parent / ".github" / "workflows" / "release.yml").read_text()
        cls.wf = yaml.safe_load(cls.text)
        cls.jobs = cls.wf["jobs"]

    def test_no_pull_request_target_and_read_only_by_default(self):
        triggers = self.wf.get("on", self.wf.get(True))
        self.assertNotIn("pull_request_target", triggers)
        self.assertEqual(self.wf["permissions"], {"contents": "read"})

    def test_release_token_only_in_jobs_that_never_run_for_pull_requests(self):
        for name, job in self.jobs.items():
            if "RELEASE_TAG_TOKEN" in yaml.safe_dump(job):
                cond = str(job.get("if", ""))
                self.assertIn("!= 'pull_request'", cond, f"{name} can run for pull requests")

    def test_write_permissions_are_confined_to_publish(self):
        writers = [n for n, j in self.jobs.items() if "write" in yaml.safe_dump(j.get("permissions", {}))]
        self.assertEqual(writers, ["publish"])

    def test_tag_job_uses_release_token_and_script(self):
        job = self.jobs["tag"]
        steps = yaml.safe_dump(job["steps"])
        self.assertIn("secrets.RELEASE_TAG_TOKEN", steps)
        self.assertIn("releasectl.py tag", steps)
        self.assertIn("tagger == 'app'", job["if"])
        self.assertIn("TAG_SOURCE: app", steps)
        names = [st["name"] for st in job["steps"] if "name" in st]
        self.assertLess(names.index("Refuse to tag over an existing release or draft"),
                        names.index("Push the tag (the tag run builds and publishes it)"))

    def test_publish_verifies_tag_before_releasing(self):
        names = [s.get("name", s.get("uses", "")) for s in self.jobs["publish"]["steps"]]
        verify = names.index("Verify the tag before publishing")
        release = next(i for i, s in enumerate(self.jobs["publish"]["steps"]) if "action-gh-release" in s.get("uses", ""))
        self.assertLess(verify, release)
        self.assertLess(names.index("Refuse to overwrite an existing release or draft"), release)

    def test_publish_only_for_decided_releases(self):
        cond = self.jobs["publish"]["if"]
        for part in ("!= 'pull_request'", "needs.decide.outputs.should == 'true'", "needs.build.result != 'skipped'"):
            self.assertIn(part, cond)

    def test_builds_wait_for_tag_job_only_via_tag_run(self):
        cond = self.jobs["build"]["if"]
        self.assertIn("startsWith(github.ref, 'refs/tags/v') || needs.decide.outputs.tagger != 'app'", cond)

    def test_release_action_is_strict(self):
        rel = next(s for s in self.jobs["publish"]["steps"] if "action-gh-release" in s.get("uses", ""))
        self.assertTrue(rel["with"]["fail_on_unmatched_files"])
        self.assertFalse(rel["with"]["draft"])


if __name__ == "__main__":
    unittest.main()
