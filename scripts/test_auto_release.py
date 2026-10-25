"""Tests for scripts/auto-release.py, run against throwaway git repos.

    python3 -m unittest discover -s scripts -p 'test_*.py'
"""
import importlib.util
import os
import pathlib
import subprocess
import tempfile
import unittest

SCRIPT = pathlib.Path(__file__).with_name("auto-release.py")
spec = importlib.util.spec_from_file_location("auto_release", SCRIPT)
ar = importlib.util.module_from_spec(spec)
spec.loader.exec_module(ar)


class Repo:
    def __init__(self, path: pathlib.Path):
        self.path = path
        self.git("init", "-q", "-b", "main")
        self.git("config", "user.email", "t@example.com")
        self.git("config", "user.name", "t")
        self.git("config", "commit.gpgsign", "false")

    def git(self, *args: str) -> str:
        return subprocess.check_output(["git", *args], cwd=self.path, text=True).strip()

    def commit(self, file: str, msg: str = "change") -> str:
        p = self.path / file
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_text(p.read_text() + "x\n" if p.exists() else "x\n")
        self.git("add", "-A")
        self.git("commit", "-q", "-m", msg)
        sha = self.git("rev-parse", "HEAD")
        self.git("update-ref", "refs/remotes/origin/main", sha)  # stands in for the fetched main
        return sha

    def tag(self, name: str, rev: str = "HEAD") -> None:
        self.git("tag", name, rev)


class DecideTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.repo = Repo(pathlib.Path(self.tmp.name))
        self.out = pathlib.Path(self.tmp.name) / "out.txt"
        self.cwd = os.getcwd()
        os.chdir(self.repo.path)
        self.saved = dict(os.environ)
        for k in ("GITHUB_REF", "BEFORE", "FORCE", "HEAD_MSG", "HAS_TAG_TOKEN", "MAIN_REF", "GITHUB_SHA"):
            os.environ.pop(k, None)
        os.environ["GITHUB_OUTPUT"] = str(self.out)

    def tearDown(self):
        os.chdir(self.cwd)
        os.environ.clear()
        os.environ.update(self.saved)
        self.tmp.cleanup()

    def run_main(self, **env):
        os.environ.setdefault("GITHUB_SHA", self.repo.git("rev-parse", "HEAD"))
        os.environ.update({k: v for k, v in env.items()})
        if self.out.exists():
            self.out.unlink()
        try:
            rc = ar.main()
        except ar.Refuse as e:
            return "refused", str(e)
        outs = dict(l.split("=", 1) for l in self.out.read_text().split())
        self.assertEqual(rc, 0)
        return outs["should"], outs["version"]

    # --- ordinary pushes
    def test_app_change_bumps_patch(self):
        self.repo.commit("main.go")
        self.repo.tag("v1.0.0", "HEAD")
        sha = self.repo.commit("internal/x.go")
        self.assertEqual(self.run_main(GITHUB_SHA=sha, BEFORE=self.repo.git("rev-parse", "HEAD~1")), ("true", "1.0.1"))

    def test_docs_only_is_skipped(self):
        self.repo.commit("main.go")
        self.repo.tag("v1.0.0")
        sha = self.repo.commit("README.md")
        self.assertEqual(self.run_main(GITHUB_SHA=sha, BEFORE=self.repo.git("rev-parse", "HEAD~1"))[0], "false")

    # --- skip marker vs force
    def test_skip_marker_honoured_without_force(self):
        self.repo.commit("main.go")
        self.repo.tag("v1.0.0")
        self.repo.commit("internal/x.go", "fix [skip release]")
        self.assertEqual(self.run_main(FORCE="false")[0], "false")

    def test_force_overrides_skip_marker(self):
        self.repo.commit("main.go")
        self.repo.tag("v1.0.0")
        self.repo.commit("internal/x.go", "fix [skip release]")
        self.assertEqual(self.run_main(FORCE="true"), ("true", "1.0.1"))

    def test_force_releases_docs_only(self):
        self.repo.commit("main.go")
        self.repo.tag("v1.0.0")
        self.repo.commit("README.md")
        self.assertEqual(self.run_main(FORCE="true"), ("true", "1.0.1"))

    # --- duplicates / reuse
    def test_already_tagged_commit_is_left_to_tag_run(self):
        self.repo.commit("main.go")
        self.repo.tag("v1.0.0")
        self.assertEqual(self.run_main(FORCE="true")[0], "false")

    def test_clash_check_refuses(self):
        self.repo.commit("main.go")
        self.repo.tag("v1.0.0")
        sha = self.repo.commit("internal/x.go")
        orig = ar.semver_tags
        ar.semver_tags = lambda: [(1, 0, 0)]  # pretend v1.0.1 isn't visible as a semver tag
        try:
            self.repo.tag("v1.0.1", self.repo.git("rev-parse", "HEAD~1"))
            status, msg = self.run_main(GITHUB_SHA=sha, FORCE="true")
        finally:
            ar.semver_tags = orig
        self.assertEqual(status, "refused")
        self.assertIn("already exists", msg)

    # --- commit validation
    def test_superseded_commit_is_skipped(self):
        self.repo.commit("main.go")
        self.repo.tag("v1.0.0")
        old = self.repo.commit("internal/a.go")
        self.repo.commit("internal/b.go")
        self.assertEqual(self.run_main(GITHUB_SHA=old)[0], "false")

    def test_commit_not_on_main_is_refused(self):
        self.repo.commit("main.go")
        self.repo.tag("v1.0.0")
        self.repo.git("checkout", "-q", "-b", "feature")
        self.repo.git("commit", "-q", "--allow-empty", "-m", "off main")
        sha = self.repo.git("rev-parse", "HEAD")
        status, _ = self.run_main(GITHUB_SHA=sha)
        self.assertEqual(status, "refused")

    def test_non_main_branch_is_refused(self):
        self.repo.commit("main.go")
        self.assertEqual(self.run_main(GITHUB_REF="refs/heads/feature")[0], "refused")

    # --- tag-triggered runs
    def test_tag_run_on_main_commit(self):
        sha = self.repo.commit("main.go")
        self.repo.tag("v1.0.2")
        self.assertEqual(self.run_main(GITHUB_REF="refs/tags/v1.0.2", GITHUB_SHA=sha), ("true", "1.0.2"))

    def test_tag_run_for_other_commit_is_refused(self):
        a = self.repo.commit("main.go")
        self.repo.commit("internal/x.go")
        self.repo.tag("v1.0.2", a)
        status, msg = self.run_main(GITHUB_REF="refs/tags/v1.0.2", GITHUB_SHA=self.repo.git("rev-parse", "HEAD"))
        self.assertEqual(status, "refused")
        self.assertIn("points at", msg)

    def test_tag_not_on_main_is_refused(self):
        self.repo.commit("main.go")
        self.repo.git("checkout", "-q", "-b", "feature")
        self.repo.git("commit", "-q", "--allow-empty", "-m", "off main")
        self.repo.tag("v9.9.9")
        status, msg = self.run_main(GITHUB_REF="refs/tags/v9.9.9", GITHUB_SHA=self.repo.git("rev-parse", "HEAD"))
        self.assertEqual(status, "refused")
        self.assertIn("never reached main", msg)

    def test_tag_run_ignores_skip_marker(self):
        sha = self.repo.commit("main.go", "docs [skip release]")
        self.repo.tag("v1.0.2")
        self.assertEqual(self.run_main(GITHUB_REF="refs/tags/v1.0.2", GITHUB_SHA=sha)[0], "true")

    def test_non_semver_tag_is_refused(self):
        sha = self.repo.commit("main.go")
        self.repo.tag("v1.0")
        self.assertEqual(self.run_main(GITHUB_REF="refs/tags/v1.0", GITHUB_SHA=sha)[0], "refused")

    # --- tagger output
    def test_tagger_output(self):
        sha = self.repo.commit("main.go")
        self.run_main(GITHUB_SHA=sha, HAS_TAG_TOKEN="true")
        self.assertIn("tagger=app", self.out.read_text())
        self.run_main(GITHUB_SHA=sha, HAS_TAG_TOKEN="false")
        self.assertIn("tagger=builtin", self.out.read_text())


if __name__ == "__main__":
    unittest.main()
