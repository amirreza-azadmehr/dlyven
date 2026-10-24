package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/amirreza-azadmehr/dlyven/internal/errs"
)

func TestNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.0.3", "1.0.2", true},
		{"v1.0.3", "1.0.2", true},
		{"1.0.2", "1.0.2", false},
		{"1.0.2", "1.0.10", false},
		{"1.10.0", "1.9.9", true},
		{"2.0.0", "1.99.99", true},
		{"1.0.2", "1.0.2-pr.7", true}, // a PR build of 1.0.2 is older than 1.0.2
		{"1.0.2-pr.7", "1.0.2", false},
		{"1.1", "1.0.9", true},
		{"dev", "1.0.0", false},
		{"1.0.0", "dev", false},
		{"", "1.0.0", false},
	}
	for _, c := range cases {
		if got := Newer(c.a, c.b); got != c.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestDetect(t *testing.T) {
	dir := t.TempDir()
	touch := func(p string) string {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}

	macExe := touch(filepath.Join(dir, "Applications", "Dlyven.app", "Contents", "MacOS", "Dlyven"))
	if got := detect(macExe, "darwin", "arm64"); got.mode != ModeReplace || got.asset != AssetMacZip || got.path != filepath.Join(dir, "Applications", "Dlyven.app") {
		t.Errorf("mac bundle: %+v", got)
	}
	if got := detect("/private/var/folders/x/AppTranslocation/ABC/d/Dlyven.app/Contents/MacOS/Dlyven", "darwin", "arm64"); got.mode != ModeManual {
		t.Errorf("translocated app must be manual: %+v", got)
	}
	if got := detect(touch(filepath.Join(dir, "bare", "Dlyven")), "darwin", "arm64"); got.mode != ModeManual {
		t.Errorf("binary outside a bundle must be manual: %+v", got)
	}

	portable := touch(filepath.Join(dir, "portable", "Dlyven.exe"))
	if got := detect(portable, "windows", "amd64"); got.mode != ModeReplace || got.asset != AssetWindowsPortable {
		t.Errorf("windows portable: %+v", got)
	}
	installed := touch(filepath.Join(dir, "installed", "Dlyven.exe"))
	touch(filepath.Join(dir, "installed", "uninstall.exe"))
	if got := detect(installed, "windows", "amd64"); got.mode != ModeInstaller || got.asset != AssetWindowsSetup {
		t.Errorf("windows installed: %+v", got)
	}
	if got := detect(portable, "windows", "arm64"); got.mode != ModeManual {
		t.Errorf("windows arm64 has no build: %+v", got)
	}

	linux := touch(filepath.Join(dir, "linux", "Dlyven"))
	if got := detect(linux, "linux", "amd64"); got.mode != ModeReplace || got.asset != AssetLinuxX64 {
		t.Errorf("linux amd64: %+v", got)
	}
	if got := detect(linux, "linux", "arm64"); got.asset != AssetLinuxARM64 {
		t.Errorf("linux arm64: %+v", got)
	}

	if runtime.GOOS != "windows" && os.Geteuid() != 0 {
		ro := touch(filepath.Join(dir, "readonly", "Dlyven"))
		if err := os.Chmod(filepath.Dir(ro), 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(filepath.Dir(ro), 0o755) })
		if got := detect(ro, "linux", "amd64"); got.mode != ModeManual {
			t.Errorf("read-only folder must be manual: %+v", got)
		}
	}
}

// fakeGitHub serves /releases/latest (redirect), the asset and SHA256SUMS.txt like GitHub does.
func fakeGitHub(t *testing.T, tag, asset string, body []byte, sum string) *httptest.Server {
	t.Helper()
	if sum == "" {
		h := sha256.Sum256(body)
		sum = hex.EncodeToString(h[:])
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/releases/tag/"+tag, http.StatusFound)
	})
	mux.HandleFunc("/releases/download/"+tag+"/"+asset, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	})
	mux.HandleFunc("/releases/download/"+tag+"/"+SumsFile, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  %s\n%s  Dlyven-Other\n", sum, asset, strings.Repeat("0", 64))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newTestManager(t *testing.T, srv *httptest.Server, current string, tg target) (*Manager, *[]string) {
	t.Helper()
	var events []string
	m := &Manager{
		cfg: Config{
			RepoURL: srv.URL, Current: current, Client: srv.Client,
			Emit: func(name string, data any) {
				if p, ok := data.(Progress); ok {
					name += ":" + p.Stage
				}
				events = append(events, name)
			},
			Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		target: tg,
	}
	return m, &events
}

func TestDownloadAndApply(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "Dlyven")
	if err := os.WriteFile(exe, []byte("\x7fELF old"), 0o755); err != nil {
		t.Fatal(err)
	}
	newBin := []byte("\x7fELF new build")
	srv := fakeGitHub(t, "v1.2.0", AssetLinuxX64, newBin, "")
	m, events := newTestManager(t, srv, "1.1.0", target{mode: ModeReplace, asset: AssetLinuxX64, path: exe})

	st, err := m.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !st.Available || st.Latest != "1.2.0" || !strings.HasSuffix(st.ReleaseURL, "/releases/tag/v1.2.0") {
		t.Fatalf("check: %+v", st)
	}

	st, err = m.Download(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !st.Ready || st.Busy {
		t.Fatalf("download: %+v", st)
	}
	if b, _ := os.ReadFile(exe); string(b) != "\x7fELF old" {
		t.Fatal("download must not touch the running executable")
	}
	if got := strings.Join(*events, " "); !strings.Contains(got, "update:progress:verifying") || !strings.Contains(got, "update:progress:ready") {
		t.Errorf("events: %s", got)
	}

	if err := m.ApplyStaged(); err != nil {
		t.Fatal(err)
	}
	if err := m.ApplyStaged(); err != nil { // shutdown calls it again after Restart
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(exe); string(b) != string(newBin) {
		t.Fatalf("not replaced: %q", b)
	}
	if fi, _ := os.Stat(exe); runtime.GOOS != "windows" && fi.Mode().Perm()&0o100 == 0 { // no exec bit on Windows
		t.Error("new executable is not executable")
	}
	prev, _ := filepath.Glob(filepath.Join(dir, stagePrefix+"*", "previous-Dlyven"))
	if len(prev) != 1 {
		t.Fatalf("previous version not kept for cleanup: %v", prev)
	}

	m.Cleanup()
	if left, _ := filepath.Glob(filepath.Join(dir, stagePrefix+"*")); len(left) != 0 {
		t.Errorf("cleanup left %v", left)
	}
}

func TestDownloadRejectsBadChecksum(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "Dlyven")
	if err := os.WriteFile(exe, []byte("\x7fELF old"), 0o755); err != nil {
		t.Fatal(err)
	}
	srv := fakeGitHub(t, "v1.2.0", AssetLinuxX64, []byte("\x7fELF tampered"), strings.Repeat("ab", 32))
	m, events := newTestManager(t, srv, "1.1.0", target{mode: ModeReplace, asset: AssetLinuxX64, path: exe})

	st, err := m.Download(context.Background())
	if errs.CodeOf(err) != errs.ErrChecksum {
		t.Fatalf("want ERR_CHECKSUM, got %v", err)
	}
	if st.Ready || st.Busy {
		t.Errorf("status after failure: %+v", st)
	}
	if !strings.Contains(strings.Join(*events, " "), "update:progress:error") {
		t.Errorf("no error event: %v", *events)
	}
	if left, _ := filepath.Glob(filepath.Join(dir, stagePrefix+"*")); len(left) != 0 {
		t.Errorf("failed download left %v", left)
	}
	if err := m.ApplyStaged(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "\x7fELF old" {
		t.Fatal("executable changed after a failed download")
	}
}

func TestDownloadRejectsNonExecutable(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "Dlyven.exe")
	if err := os.WriteFile(exe, []byte("MZ old"), 0o755); err != nil {
		t.Fatal(err)
	}
	srv := fakeGitHub(t, "v1.2.0", AssetWindowsPortable, []byte("<html>not a program</html>"), "")
	m, _ := newTestManager(t, srv, "1.1.0", target{mode: ModeReplace, asset: AssetWindowsPortable, path: exe})
	if _, err := m.Download(context.Background()); errs.CodeOf(err) != errs.ErrChecksum {
		t.Fatalf("want ERR_CHECKSUM, got %v", err)
	}
}

func TestUpToDateDownloadsNothing(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "Dlyven")
	srv := fakeGitHub(t, "v1.2.0", AssetLinuxX64, []byte("\x7fELF"), "")
	m, _ := newTestManager(t, srv, "1.2.0", target{mode: ModeReplace, asset: AssetLinuxX64, path: exe})
	st, err := m.Download(context.Background())
	if err != nil || st.Available || st.Ready {
		t.Fatalf("up to date: %+v %v", st, err)
	}
	if left, _ := filepath.Glob(filepath.Join(dir, stagePrefix+"*")); len(left) != 0 {
		t.Errorf("left %v", left)
	}
}

func TestManualModeRefusesDownload(t *testing.T) {
	srv := fakeGitHub(t, "v1.2.0", AssetLinuxX64, []byte("\x7fELF"), "")
	m, _ := newTestManager(t, srv, "1.1.0", manual("/usr/bin/dlyven", "app folder is not writable"))
	if _, err := m.Download(context.Background()); err == nil {
		t.Fatal("manual mode must not download")
	}
	if st := m.Status(); st.Mode != ModeManual || st.Ready {
		t.Errorf("%+v", st)
	}
}

// The workflow, the release notes and the updater must agree on asset names, or the
// updater 404s on every release.
func TestAssetNamesMatchWorkflow(t *testing.T) {
	for _, f := range []string{"../../.github/workflows/release.yml", "../../.github/scripts/release-notes.sh"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{AssetWindowsSetup, AssetWindowsPortable, AssetMacZip, AssetLinuxX64, AssetLinuxARM64} {
			if !strings.Contains(string(b), name) {
				t.Errorf("%s does not mention %s", f, name)
			}
		}
	}
}
