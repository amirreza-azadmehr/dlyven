// Package updater keeps Dlyven itself current from GitHub Releases. It finds the latest
// release, downloads this platform's asset, verifies it against the release's
// SHA256SUMS.txt and swaps it in. Nothing is downloaded until the user asks for it.
package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amirreza-azadmehr/dlyven/internal/engine"
	"github.com/amirreza-azadmehr/dlyven/internal/errs"
	"github.com/amirreza-azadmehr/dlyven/internal/sysutil"
)

const (
	EventStatus   = "update:status"
	EventProgress = "update:progress"

	// stagePrefix names the folders downloads are staged in (next to the app, or in the
	// temp folder for installers). Cleanup removes them on the next launch.
	stagePrefix = ".dlyven-update-"
)

type Status struct {
	Current    string `json:"current"`
	Latest     string `json:"latest"` // "" until a check succeeds
	Available  bool   `json:"available"`
	Mode       Mode   `json:"mode"`
	Asset      string `json:"asset"`
	ReleaseURL string `json:"releaseUrl"`
	Ready      bool   `json:"ready"` // downloaded and verified: restart to finish
	Busy       bool   `json:"busy"`
}

// Progress is emitted as the "update:progress" event while downloading.
type Progress struct {
	Stage      string  `json:"stage"` // downloading | verifying | preparing | ready | error
	Percent    float64 `json:"percent"`
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	ErrorCode  string  `json:"errorCode,omitempty"`
	Error      string  `json:"error,omitempty"`
}

type Config struct {
	RepoURL string // https://github.com/<owner>/<repo>
	Current string // running version
	Exe     string // running executable; "" forces ModeManual
	Dev     bool   // `wails dev` build: never replace it
	Client  func() *http.Client
	Emit    func(name string, data any)
	Log     *slog.Logger
}

type Manager struct {
	cfg    Config
	target target

	mu       sync.Mutex
	busy     bool
	tag      string // latest release tag, e.g. v1.2.3
	staged   string // verified new executable, .app or installer
	stageDir string
	applied  bool
}

func New(cfg Config) *Manager {
	t := manual(cfg.Exe, "executable path unknown")
	if cfg.Exe != "" {
		t = detect(cfg.Exe, runtime.GOOS, runtime.GOARCH)
	}
	if cfg.Dev {
		t = manual(t.path, "development build")
	}
	cfg.Log.Info("updater", "mode", t.mode, "asset", t.asset, "path", t.path, "why", t.why)
	return &Manager{cfg: cfg, target: t}
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statusLocked()
}

func (m *Manager) statusLocked() Status {
	latest := strings.TrimPrefix(m.tag, "v")
	st := Status{
		Current: m.cfg.Current, Latest: latest, Available: latest != "" && Newer(latest, m.cfg.Current),
		Mode: m.target.mode, Asset: m.target.asset, Ready: m.staged != "", Busy: m.busy,
		ReleaseURL: m.cfg.RepoURL + "/releases/latest",
	}
	if m.tag != "" {
		st.ReleaseURL = m.cfg.RepoURL + "/releases/tag/" + m.tag
	}
	return st
}

// Check looks up the latest release. It never downloads anything.
func (m *Manager) Check(ctx context.Context) (Status, error) {
	tag, err := engine.LatestReleaseTag(ctx, m.cfg.Client(), m.cfg.RepoURL)
	if err != nil {
		return m.Status(), err
	}
	m.mu.Lock()
	m.tag = tag
	st := m.statusLocked()
	m.mu.Unlock()
	m.cfg.Emit(EventStatus, st)
	return st, nil
}

// Download fetches, verifies and stages the latest release for this platform. Restart
// installs it; so does quitting Dlyven (ApplyStaged), in replace mode.
func (m *Manager) Download(ctx context.Context) (Status, error) {
	if m.target.mode == ModeManual {
		return m.Status(), errs.New(errs.ErrUnknown, "this copy can't update itself: "+m.target.why)
	}
	m.mu.Lock()
	if m.busy {
		m.mu.Unlock()
		return m.Status(), errs.New(errs.ErrBusy, "update in progress")
	}
	m.busy = true
	m.mu.Unlock()

	err := m.download(ctx)

	m.mu.Lock()
	m.busy = false
	st := m.statusLocked()
	m.mu.Unlock()
	if err != nil {
		m.cfg.Log.Error("update failed", "err", err)
		m.progress(Progress{Stage: "error", ErrorCode: string(errs.CodeOf(err)), Error: errs.DetailOf(err)})
	}
	m.cfg.Emit(EventStatus, st)
	return st, err
}

func (m *Manager) download(ctx context.Context) error {
	if m.Status().Ready {
		return nil
	}
	if _, err := m.Check(ctx); err != nil {
		return err
	}
	m.mu.Lock()
	tag, available := m.tag, m.statusLocked().Available
	m.mu.Unlock()
	if !available {
		return nil
	}

	parent := filepath.Dir(m.target.path) // same volume, so applying is a rename
	if m.target.mode == ModeInstaller {
		parent = os.TempDir()
	}
	dir, err := os.MkdirTemp(parent, stagePrefix)
	if err != nil {
		return errs.Wrap(errs.ErrDisk, err)
	}
	ok := false
	defer func() {
		if !ok {
			_ = os.RemoveAll(dir)
		}
	}()

	base := m.cfg.RepoURL + "/releases/download/" + tag + "/"
	file := filepath.Join(dir, m.target.asset)
	sum, err := m.fetch(ctx, base+m.target.asset, file)
	if err != nil {
		return err
	}
	m.progress(Progress{Stage: "verifying", Percent: 100})
	want, err := m.checksum(ctx, base+SumsFile, m.target.asset)
	if err != nil {
		return err
	}
	if !strings.EqualFold(want, sum) {
		return errs.New(errs.ErrChecksum, fmt.Sprintf("%s: expected %s, got %s", m.target.asset, want, sum))
	}
	m.progress(Progress{Stage: "preparing", Percent: 100})
	staged, err := prepare(file, dir, m.target)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.staged, m.stageDir = staged, dir
	m.mu.Unlock()
	ok = true
	m.cfg.Log.Info("update downloaded", "tag", tag, "asset", m.target.asset)
	m.progress(Progress{Stage: "ready", Percent: 100})
	return nil
}

// ApplyStaged moves a downloaded update into place (replace mode). Restart calls it, and
// so does shutdown, so a downloaded update is installed whenever Dlyven quits. The running
// process keeps working: it only loses its path, not its open files.
func (m *Manager) ApplyStaged() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.staged == "" || m.applied || m.target.mode != ModeReplace {
		return nil
	}
	prev := filepath.Join(m.stageDir, "previous-"+filepath.Base(m.target.path))
	if err := os.Rename(m.target.path, prev); err != nil {
		return errs.Wrap(errs.ErrDisk, err)
	}
	if err := os.Rename(m.staged, m.target.path); err != nil {
		_ = os.Rename(prev, m.target.path)
		return errs.Wrap(errs.ErrDisk, err)
	}
	m.applied = true
	m.cfg.Log.Info("update applied", "tag", m.tag, "path", m.target.path)
	return nil
}

// Restart installs the staged update and starts the new version once this process has
// exited. The caller quits the app right after.
func (m *Manager) Restart() error {
	m.mu.Lock()
	staged := m.staged
	m.mu.Unlock()
	if staged == "" {
		return errs.New(errs.ErrUnknown, "no update downloaded")
	}
	installer := ""
	if m.target.mode == ModeInstaller {
		installer = staged
	} else if err := m.ApplyStaged(); err != nil {
		return err
	}
	if err := relaunch(m.target.path, installer); err != nil {
		return errs.Wrap(errs.ErrUnknown, err)
	}
	return nil
}

// Cleanup deletes what earlier updates left behind: staging folders next to the app
// (holding the previous version) and downloaded installers in the temp folder.
func (m *Manager) Cleanup() {
	var dirs []string
	if m.target.path != "" {
		d, _ := filepath.Glob(filepath.Join(filepath.Dir(m.target.path), stagePrefix+"*"))
		dirs = append(dirs, d...)
	}
	d, _ := filepath.Glob(filepath.Join(os.TempDir(), stagePrefix+"*"))
	for _, dir := range append(dirs, d...) {
		if err := os.RemoveAll(dir); err != nil {
			m.cfg.Log.Warn("update cleanup failed", "dir", dir, "err", err)
		}
	}
}

func (m *Manager) progress(p Progress) { m.cfg.Emit(EventProgress, p) }

func (m *Manager) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Dlyven-updater/"+m.cfg.Current)
	resp, err := m.cfg.Client().Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, errs.New(errs.ErrCanceled, "")
		}
		return nil, errs.Wrap(errs.ErrNetwork, err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errs.New(errs.ErrNetwork, fmt.Sprintf("GET %s: HTTP %d", url, resp.StatusCode))
	}
	return resp, nil
}

// fetch streams url to dest and returns its SHA-256.
func (m *Manager) fetch(ctx context.Context, url, dest string) (string, error) {
	m.progress(Progress{Stage: "downloading"})
	resp, err := m.get(ctx, url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	f, err := os.Create(dest)
	if err != nil {
		return "", errs.Wrap(errs.ErrDisk, err)
	}
	defer f.Close()

	h := sha256.New()
	pw := &progressWriter{total: resp.ContentLength, report: func(done, total int64) {
		pct := 0.0
		if total > 0 {
			pct = float64(done) / float64(total) * 100
		}
		m.progress(Progress{Stage: "downloading", Percent: pct, Downloaded: done, Total: total})
	}}
	if _, err := io.Copy(io.MultiWriter(f, h, pw), resp.Body); err != nil {
		if ctx.Err() != nil {
			return "", errs.New(errs.ErrCanceled, "")
		}
		return "", errs.Wrap(errs.ErrNetwork, err)
	}
	pw.report(pw.done, pw.done)
	if err := f.Close(); err != nil {
		return "", errs.Wrap(errs.ErrDisk, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (m *Manager) checksum(ctx context.Context, url, name string) (string, error) {
	resp, err := m.get(ctx, url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", errs.Wrap(errs.ErrNetwork, err)
	}
	if sum := engine.ParseChecksum(string(body), name); sum != "" {
		return sum, nil
	}
	return "", errs.New(errs.ErrChecksum, "no checksum for "+name+" in "+url)
}

type progressWriter struct {
	total, done int64
	last        time.Time
	report      func(done, total int64)
}

func (w *progressWriter) Write(p []byte) (int, error) {
	w.done += int64(len(p))
	if time.Since(w.last) > 150*time.Millisecond {
		w.last = time.Now()
		w.report(w.done, w.total)
	}
	return len(p), nil
}

// prepare turns a verified download into something ApplyStaged or the installer can use.
func prepare(file, dir string, t target) (string, error) {
	switch {
	case t.mode == ModeInstaller:
		return file, nil
	case t.asset == AssetMacZip:
		return prepareMacApp(file, dir, t.path)
	}
	// A cheap guard against ever swapping in something that isn't a program.
	magic := "\x7fELF"
	if t.asset == AssetWindowsPortable {
		magic = "MZ"
	}
	head := make([]byte, len(magic))
	f, err := os.Open(file)
	if err != nil {
		return "", errs.Wrap(errs.ErrDisk, err)
	}
	_, err = io.ReadFull(f, head)
	f.Close()
	if err != nil || string(head) != magic {
		return "", errs.New(errs.ErrChecksum, t.asset+" is not an executable")
	}
	if err := os.Chmod(file, 0o755); err != nil {
		return "", errs.Wrap(errs.ErrDisk, err)
	}
	return file, nil
}

// prepareMacApp unpacks the release zip next to the installed app, then checks the new
// bundle's code signature and that it comes from the same signer as the running app.
func prepareMacApp(zipPath, dir, current string) (string, error) {
	out := filepath.Join(dir, "new")
	// ditto keeps the bundle's symlinks, permissions and signature intact.
	if b, err := sysutil.Command("ditto", "-x", "-k", zipPath, out).CombinedOutput(); err != nil {
		return "", errs.New(errs.ErrDisk, "unzip: "+strings.TrimSpace(string(b)))
	}
	apps, _ := filepath.Glob(filepath.Join(out, "*.app"))
	if len(apps) != 1 {
		return "", errs.New(errs.ErrChecksum, "release zip does not hold exactly one .app")
	}
	app := apps[0]
	if b, err := sysutil.Command("codesign", "--verify", "--deep", "--strict", app).CombinedOutput(); err != nil {
		return "", errs.New(errs.ErrChecksum, "code signature: "+strings.TrimSpace(string(b)))
	}
	// An ad-hoc app may move to Developer ID, never the other way or to another team.
	have, want := signer(app), signer(current)
	if (want.team != "" && have.team != want.team) || (want.id != "" && have.id != want.id) {
		return "", errs.New(errs.ErrChecksum, fmt.Sprintf("signed as %s/%s, expected %s/%s", have.id, have.team, want.id, want.team))
	}
	_ = exec.Command("xattr", "-dr", "com.apple.quarantine", app).Run()
	return app, nil
}

type signature struct{ id, team string }

func signer(app string) signature {
	out, _ := sysutil.Command("codesign", "-dv", app).CombinedOutput()
	var s signature
	for _, line := range strings.Split(string(out), "\n") {
		if v, ok := strings.CutPrefix(line, "Identifier="); ok {
			s.id = strings.TrimSpace(v)
		}
		if v, ok := strings.CutPrefix(line, "TeamIdentifier="); ok && strings.TrimSpace(v) != "not set" {
			s.team = strings.TrimSpace(v)
		}
	}
	return s
}

// waitScript waits (up to 30s) for the process in $1 to exit, then runs the rest of argv.
const waitScript = `pid=$1; shift; i=0
while kill -0 "$pid" 2>/dev/null && [ "$i" -lt 150 ]; do sleep 0.2; i=$((i+1)); done
exec "$@"`

// relaunchScript is waitScript for Windows, plus the silent Setup run in installer mode.
// /D= must be last and unquoted (NSIS reads it to the end of the command line). Paths
// come in through the environment so they can never be read as script.
const relaunchScript = `$ErrorActionPreference = 'SilentlyContinue'
Wait-Process -Id ([int]$env:DLYVEN_PID) -Timeout 30
if ($env:DLYVEN_INSTALLER) {
  try { Start-Process -FilePath $env:DLYVEN_INSTALLER -ArgumentList ('/S /D=' + $env:DLYVEN_DIR) -Verb RunAs -Wait -ErrorAction Stop } catch {}
}
Start-Process -FilePath $env:DLYVEN_EXE`

// relaunch starts a detached helper that waits for this process to exit, runs installer
// if one is given (Windows Setup), then starts the app at path. Starting it any sooner
// would hand the launch straight back to us through the single-instance lock.
func relaunch(path, installer string) error {
	pid := strconv.Itoa(os.Getpid())
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = sysutil.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", relaunchScript)
		cmd.Env = append(cmd.Env, "DLYVEN_PID="+pid, "DLYVEN_EXE="+path, "DLYVEN_INSTALLER="+installer, "DLYVEN_DIR="+filepath.Dir(path))
	case "darwin":
		cmd = sysutil.Command("sh", "-c", waitScript, "sh", pid, "open", path)
	default:
		cmd = sysutil.Command("sh", "-c", waitScript, "sh", pid, path)
	}
	return cmd.Start()
}
