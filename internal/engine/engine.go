// Package engine installs, updates and locates the external tools Dlyven drives:
// yt-dlp (extraction/downloads), ffmpeg + ffprobe (merging/conversion) and Deno
// (JavaScript runtime yt-dlp uses for YouTube). Every download is SHA-256 verified.
package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/amirreza-azadmehr/dlyven/internal/errs"
	"github.com/amirreza-azadmehr/dlyven/internal/settings"
	"github.com/amirreza-azadmehr/dlyven/internal/sysutil"
)

const (
	CompYtDlp  = "yt-dlp"
	CompFFmpeg = "ffmpeg"
	CompDeno   = "deno"
)

type ComponentStatus struct {
	Name            string `json:"name"`
	Installed       bool   `json:"installed"`
	Path            string `json:"path"`
	Version         string `json:"version"`
	Source          string `json:"source"` // custom | managed | system | ""
	Latest          string `json:"latest"`
	UpdateAvailable bool   `json:"updateAvailable"`
	Optional        bool   `json:"optional"`
}

type Status struct {
	YtDlp  ComponentStatus `json:"ytDlp"`
	FFmpeg ComponentStatus `json:"ffmpeg"`
	Deno   ComponentStatus `json:"deno"`
	Busy   bool            `json:"busy"`
	Ready  bool            `json:"ready"` // yt-dlp and ffmpeg present
}

// Progress is emitted as the "engine:progress" event while installing.
type Progress struct {
	Component  string  `json:"component"`
	Stage      string  `json:"stage"` // checking | downloading | verifying | extracting | done | error
	Percent    float64 `json:"percent"`
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	ErrorCode  string  `json:"errorCode,omitempty"`
	Error      string  `json:"error,omitempty"`
}

type Manager struct {
	binDir   string
	cacheDir string
	settings func() settings.Settings
	emit     func(name string, data any)
	log      *slog.Logger

	mu          sync.Mutex
	busy        bool
	versions    map[string]cachedVersion
	latest      map[string]string
	impersonate string // "chrome", "off", or "" if not probed yet
}

type cachedVersion struct {
	mod     time.Time
	version string
}

func New(binDir, cacheDir string, st func() settings.Settings, emit func(string, any), log *slog.Logger) *Manager {
	_ = os.MkdirAll(binDir, 0o755)
	clearQuarantine(binDir)
	pruneNormalized(cacheDir, "") // repaired cookie copies from a previous run
	return &Manager{
		binDir: binDir, cacheDir: cacheDir, settings: st, emit: emit, log: log,
		versions: map[string]cachedVersion{}, latest: map[string]string{},
	}
}

func (m *Manager) BinDir() string { return m.binDir }

func (m *Manager) httpClient() *http.Client {
	s := m.settings()
	return sysutil.NewHTTPClient(s.ProxyURL(), time.Duration(s.TimeoutSec)*time.Second)
}

func isFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

// resolve returns (path, source) for a component.
func (m *Manager) resolve(comp string) (string, string) {
	s := m.settings()
	custom := ""
	switch comp {
	case CompYtDlp:
		custom = s.YtDlpPath
	case CompFFmpeg:
		custom = s.FfmpegPath
	}
	if custom != "" {
		if fi, err := os.Stat(custom); err == nil {
			if fi.IsDir() {
				if p := filepath.Join(custom, exe(comp)); isFile(p) {
					return p, "custom"
				}
			} else {
				return custom, "custom"
			}
		}
	}
	if p := filepath.Join(m.binDir, exe(comp)); isFile(p) {
		return p, "managed"
	}
	if p := sysutil.LookPathExtra(exe(comp)); p != "" {
		return p, "system"
	}
	return "", ""
}

// YtDlpPath returns the yt-dlp executable or "".
func (m *Manager) YtDlpPath() string { p, _ := m.resolve(CompYtDlp); return p }

// FFmpegPath returns the ffmpeg executable or "".
func (m *Manager) FFmpegPath() string { p, _ := m.resolve(CompFFmpeg); return p }

// DenoPath returns the deno executable or "".
func (m *Manager) DenoPath() string { p, _ := m.resolve(CompDeno); return p }

func (m *Manager) version(ctx context.Context, comp, path string) string {
	if path == "" {
		return ""
	}
	fi, err := os.Stat(path)
	if err != nil {
		return ""
	}
	m.mu.Lock()
	if c, ok := m.versions[path]; ok && c.mod.Equal(fi.ModTime()) {
		m.mu.Unlock()
		return c.version
	}
	m.mu.Unlock()

	// Status used to sit on this for 45s when a cold yt-dlp hung (macOS Gatekeeper /
	// first unpack). Home and Analyze waited on that. A short bound is enough: version
	// is display-only and cached after the first success.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var args []string
	switch comp {
	case CompYtDlp:
		args = []string{"--version"}
	case CompFFmpeg:
		args = []string{"-hide_banner", "-version"}
	case CompDeno:
		args = []string{"--version"}
	}
	out, err := sysutil.CommandContext(ctx, path, args...).Output()
	v := ""
	if err != nil {
		m.log.Warn("version check failed", "component", comp, "path", path, "err", err)
	} else {
		v = parseVersion(comp, string(out))
	}
	m.mu.Lock()
	m.versions[path] = cachedVersion{mod: fi.ModTime(), version: v}
	m.mu.Unlock()
	return v
}

func parseVersion(comp, out string) string {
	line := strings.TrimSpace(strings.SplitN(out, "\n", 2)[0])
	f := strings.Fields(line)
	switch comp {
	case CompFFmpeg:
		// "ffmpeg version N-117000-g1234 Copyright (c) ..."  or "ffmpeg version 7.1 ..."
		if len(f) >= 3 {
			return f[2]
		}
	case CompDeno:
		// "deno 2.1.4 (stable, release, x86_64-unknown-linux-gnu)"
		if len(f) >= 2 {
			return f[1]
		}
	}
	return line
}

func (m *Manager) cachedVersion(path string) string {
	if path == "" {
		return ""
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.versions[path]; ok {
		return c.version
	}
	return ""
}

// FillVersions runs --version on installed tools in the background. Status() itself
// never execs — that was blocking the Home screen and Instagram analysis for ~45s.
func (m *Manager) FillVersions(ctx context.Context) {
	var wg sync.WaitGroup
	for _, comp := range []string{CompYtDlp, CompFFmpeg, CompDeno} {
		p, _ := m.resolve(comp)
		if p == "" {
			continue
		}
		wg.Add(1)
		go func(c, path string) {
			defer wg.Done()
			_ = m.version(ctx, c, path)
		}(comp, p)
	}
	wg.Wait()
}

// Status reports the installation state of every component without launching binaries.
func (m *Manager) Status(_ context.Context) Status {
	st := Status{}
	build := func(comp string, optional bool) ComponentStatus {
		p, src := m.resolve(comp)
		cs := ComponentStatus{Name: comp, Path: p, Source: src, Installed: p != "", Optional: optional, Version: m.cachedVersion(p)}
		m.mu.Lock()
		cs.Latest = m.latest[comp]
		m.mu.Unlock()
		if comp == CompYtDlp && cs.Latest != "" && cs.Version != "" && cs.Source == "managed" {
			cs.UpdateAvailable = cs.Latest != cs.Version
		}
		return cs
	}
	st.YtDlp = build(CompYtDlp, false)
	st.FFmpeg = build(CompFFmpeg, false)
	st.Deno = build(CompDeno, true)
	m.mu.Lock()
	st.Busy = m.busy
	m.mu.Unlock()
	st.Ready = st.YtDlp.Installed && st.FFmpeg.Installed
	return st
}

func (m *Manager) progress(p Progress) {
	if m.emit != nil {
		m.emit("engine:progress", p)
	}
}

func (m *Manager) begin() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.busy {
		return errs.New(errs.ErrBusy, "engine operation in progress")
	}
	m.busy = true
	return nil
}

func (m *Manager) end() {
	m.mu.Lock()
	m.busy = false
	m.mu.Unlock()
}

// EnsureInstalled installs every missing component. Deno failures are non-fatal.
func (m *Manager) EnsureInstalled(ctx context.Context) error {
	if err := m.begin(); err != nil {
		return err
	}
	defer m.end()
	if p, _ := m.resolve(CompYtDlp); p == "" {
		if err := m.installYtDlp(ctx, ""); err != nil {
			return err
		}
	}
	if p, _ := m.resolve(CompFFmpeg); p == "" {
		if err := m.installFFmpeg(ctx); err != nil {
			return err
		}
	}
	if p, _ := m.resolve(CompDeno); p == "" {
		if err := m.installDeno(ctx); err != nil {
			m.log.Warn("deno install failed (YouTube may list fewer formats)", "err", err)
		}
	}
	return nil
}

// CheckUpdates queries the latest yt-dlp release tag.
func (m *Manager) CheckUpdates(ctx context.Context) (Status, error) {
	tag, err := m.latestYtDlpTag(ctx)
	if err != nil {
		return m.Status(ctx), err
	}
	m.mu.Lock()
	m.latest[CompYtDlp] = tag
	m.mu.Unlock()
	return m.Status(ctx), nil
}

// Update refreshes managed components: yt-dlp to the latest release (if newer), and
// ffmpeg/deno when missing or when force is set.
func (m *Manager) Update(ctx context.Context, force bool) error {
	if err := m.begin(); err != nil {
		return err
	}
	defer m.end()

	tag, err := m.latestYtDlpTag(ctx)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.latest[CompYtDlp] = tag
	m.mu.Unlock()

	p, src := m.resolve(CompYtDlp)
	if p == "" || src == "managed" {
		if force || p == "" || m.version(ctx, CompYtDlp, p) != tag {
			if err := m.installYtDlp(ctx, tag); err != nil {
				return err
			}
		} else {
			m.progress(Progress{Component: CompYtDlp, Stage: "done", Percent: 100})
		}
	}
	if p, src := m.resolve(CompFFmpeg); p == "" || (force && src == "managed") {
		if err := m.installFFmpeg(ctx); err != nil {
			return err
		}
	}
	if p, src := m.resolve(CompDeno); p == "" || (force && src == "managed") {
		if err := m.installDeno(ctx); err != nil {
			m.log.Warn("deno install failed", "err", err)
		}
	}
	return nil
}

func (m *Manager) latestYtDlpTag(ctx context.Context) (string, error) {
	return LatestReleaseTag(ctx, m.httpClient(), "https://github.com/yt-dlp/yt-dlp")
}

// LatestReleaseTag returns the tag of a GitHub repository's latest published release.
// It follows the /releases/latest redirect instead of the API: no rate limit, works
// through proxies that block api.github.com.
func LatestReleaseTag(ctx context.Context, c *http.Client, repoURL string) (string, error) {
	client := *c
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, repoURL+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", errs.Wrap(errs.ErrNetwork, err)
	}
	resp.Body.Close()
	loc := resp.Header.Get("Location")
	i := strings.LastIndex(loc, "/tag/")
	if i < 0 {
		return "", errs.New(errs.ErrNetwork, fmt.Sprintf("unexpected release redirect (HTTP %d)", resp.StatusCode))
	}
	return loc[i+len("/tag/"):], nil
}

func (m *Manager) installYtDlp(ctx context.Context, tag string) error {
	m.progress(Progress{Component: CompYtDlp, Stage: "checking"})
	if tag == "" {
		t, err := m.latestYtDlpTag(ctx)
		if err != nil {
			m.fail(CompYtDlp, err)
			return err
		}
		tag = t
	}
	src, err := ytDlpSource(tag)
	if err != nil {
		e := errs.Wrap(errs.ErrEngineInstall, err)
		m.fail(CompYtDlp, e)
		return e
	}
	if err := m.install(ctx, CompYtDlp, src); err != nil {
		return err
	}
	m.mu.Lock()
	m.latest[CompYtDlp] = tag
	m.mu.Unlock()
	return nil
}

func (m *Manager) installFFmpeg(ctx context.Context) error {
	m.progress(Progress{Component: CompFFmpeg, Stage: "checking"})
	srcs, err := ffmpegSources()
	if err != nil {
		e := errs.Wrap(errs.ErrEngineInstall, err)
		m.fail(CompFFmpeg, e)
		return e
	}
	for _, s := range srcs {
		if err := m.install(ctx, CompFFmpeg, s); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) installDeno(ctx context.Context) error {
	m.progress(Progress{Component: CompDeno, Stage: "checking"})
	src, err := denoSource()
	if err != nil {
		e := errs.Wrap(errs.ErrEngineInstall, err)
		m.fail(CompDeno, e)
		return e
	}
	return m.install(ctx, CompDeno, src)
}

func (m *Manager) fail(comp string, err error) {
	m.log.Error("engine install failed", "component", comp, "err", err)
	m.progress(Progress{Component: comp, Stage: "error", ErrorCode: string(errs.CodeOf(err)), Error: errs.DetailOf(err)})
}

func (m *Manager) install(ctx context.Context, comp string, src source) error {
	err := m.installSource(ctx, comp, src)
	if err != nil {
		if errs.CodeOf(err) == errs.ErrUnknown {
			err = errs.Wrap(errs.ErrEngineInstall, err)
		}
		m.fail(comp, err)
		return err
	}
	m.log.Info("engine component installed", "component", comp, "url", src.URL)
	m.progress(Progress{Component: comp, Stage: "done", Percent: 100})
	return nil
}

// Command builds a yt-dlp command. The context kills the whole process tree when canceled.
func (m *Manager) Command(ctx context.Context, args ...string) (*exec.Cmd, error) {
	p := m.YtDlpPath()
	if p == "" {
		return nil, errs.New(errs.ErrEngineMissing, "yt-dlp not installed")
	}
	return sysutil.CommandContext(ctx, p, args...), nil
}

// FFmpegCommand builds an ffmpeg command.
func (m *Manager) FFmpegCommand(ctx context.Context, args ...string) (*exec.Cmd, error) {
	p := m.FFmpegPath()
	if p == "" {
		return nil, errs.New(errs.ErrFFmpegMissing, "ffmpeg not installed")
	}
	return sysutil.CommandContext(ctx, p, args...), nil
}

// BaseArgs returns yt-dlp flags shared by analysis and downloads: network, cookies,
// tool locations and output encoding.
func (m *Manager) BaseArgs(ctx context.Context, s settings.Settings) []string {
	args := []string{
		"--ignore-config",
		"--no-colors",
		"--encoding", "utf-8",
		"--retries", fmt.Sprint(s.Retries),
		"--fragment-retries", fmt.Sprint(max(s.Retries, 3)),
		"--extractor-retries", fmt.Sprint(s.Retries),
		"--socket-timeout", fmt.Sprint(s.TimeoutSec),
	}
	if m.cacheDir != "" {
		args = append(args, "--cache-dir", m.cacheDir)
	}
	if p := s.ProxyURL(); p != "" {
		args = append(args, "--proxy", p)
	}
	if s.UserAgent != "" {
		args = append(args, "--user-agent", s.UserAgent)
	} else {
		args = append(args, m.impersonateArgs(ctx)...)
	}
	args = append(args, m.cookieArgs(s)...)
	if ff := m.FFmpegPath(); ff != "" {
		args = append(args, "--ffmpeg-location", ff)
	}
	if deno := m.DenoPath(); deno != "" {
		args = append(args, "--js-runtimes", "deno:"+deno)
	}
	return args
}

// impersonateArgs returns --impersonate chrome when this yt-dlp build can impersonate
// a browser (curl_cffi). Instagram, TikTok and similar sites now require that TLS fingerprint.
func (m *Manager) impersonateArgs(_ context.Context) []string {
	m.mu.Lock()
	cached := m.impersonate
	m.mu.Unlock()
	if cached == "off" {
		return nil
	}
	if cached != "" {
		return []string{"--impersonate", cached}
	}
	// Official yt-dlp builds can impersonate Chrome. Don't block Analyze on
	// --list-impersonate-targets (another cold yt-dlp start, 10s+ on macOS).
	m.mu.Lock()
	if m.impersonate == "" {
		m.impersonate = "chrome"
	}
	m.mu.Unlock()
	go m.probeImpersonate()
	return []string{"--impersonate", "chrome"}
}

func (m *Manager) probeImpersonate() {
	p := m.YtDlpPath()
	if p == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	out, err := sysutil.CommandContext(ctx, p, "--list-impersonate-targets").CombinedOutput()
	target := "off"
	if err == nil && strings.Contains(strings.ToLower(string(out)), "chrome") {
		target = "chrome"
	}
	m.mu.Lock()
	m.impersonate = target
	m.mu.Unlock()
}

// clearQuarantine drops the macOS download flag so the first exec of a managed
// yt-dlp/ffmpeg/deno is not stuck in Gatekeeper for tens of seconds.
func clearQuarantine(path string) {
	if runtime.GOOS != "darwin" || path == "" {
		return
	}
	_ = exec.Command("xattr", "-dr", "com.apple.quarantine", path).Run()
}
