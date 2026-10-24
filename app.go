package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/amirreza-azadmehr/dlyven/internal/appinfo"
	"github.com/amirreza-azadmehr/dlyven/internal/downloader"
	"github.com/amirreza-azadmehr/dlyven/internal/engine"
	"github.com/amirreza-azadmehr/dlyven/internal/errs"
	"github.com/amirreza-azadmehr/dlyven/internal/extractor"
	"github.com/amirreza-azadmehr/dlyven/internal/filename"
	"github.com/amirreza-azadmehr/dlyven/internal/logx"
	"github.com/amirreza-azadmehr/dlyven/internal/platform"
	"github.com/amirreza-azadmehr/dlyven/internal/settings"
	"github.com/amirreza-azadmehr/dlyven/internal/store"
	"github.com/amirreza-azadmehr/dlyven/internal/sysutil"
	"github.com/amirreza-azadmehr/dlyven/internal/updater"
)

// App is bound to the frontend: every exported method is callable from JS as
// window.go.main.App.<Method>() and returns a Promise.
type App struct {
	ctx atomic.Pointer[context.Context]

	dataDir     string
	settings    *settings.Store
	log         *slog.Logger
	logFile     *logx.RotatingFile
	translucent bool

	engine    *engine.Manager
	extractor *extractor.Extractor
	downloads *downloader.Manager
	history   *store.Store
	updater   *updater.Manager

	analyzeMu     sync.Mutex
	analyzeCancel context.CancelFunc
}

func NewApp(dataDir string, st *settings.Store, log *slog.Logger, logFile *logx.RotatingFile, translucent bool) (*App, error) {
	a := &App{dataDir: dataDir, settings: st, log: log, logFile: logFile, translucent: translucent}

	binDir, err := appinfo.SubDir("bin")
	if err != nil {
		return nil, err
	}
	cacheDir, _ := appinfo.SubDir("cache")
	a.engine = engine.New(binDir, cacheDir, st.Get, a.emit, log)
	a.extractor = extractor.New(a.engine, st.Get, log)

	a.history, err = store.Open(filepath.Join(dataDir, "history.db"))
	if err != nil {
		log.Error("history unavailable", "err", err)
		a.history = nil
	}
	a.downloads = downloader.New(a.engine, st.Get, a.history, a.emit, log,
		filepath.Join(dataDir, "queue.json"), downloader.Hooks{OnFinished: a.onFinished})

	exe, err := os.Executable()
	if err != nil {
		exe = ""
	} else if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	a.updater = updater.New(updater.Config{
		RepoURL: appinfo.RepoURL, Current: appinfo.Version, Exe: exe, Dev: devBuild,
		Client: func() *http.Client {
			s := st.Get()
			return sysutil.NewHTTPClient(s.ProxyURL(), time.Duration(s.TimeoutSec)*time.Second)
		},
		Emit: a.emit, Log: log,
	})

	st.OnChange(func(settings.Settings) { a.downloads.Wake() })
	return a, nil
}

// emit sends an event to the frontend (no-op until the window exists).
func (a *App) emit(name string, data any) {
	if c := a.ctx.Load(); c != nil {
		runtime.EventsEmit(*c, name, data)
	}
}

// focusWindow brings the existing window forward when a second instance is launched.
func (a *App) focusWindow() {
	if c := a.ctx.Load(); c != nil {
		runtime.WindowUnminimise(*c)
		runtime.WindowShow(*c)
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx.Store(&ctx)
	a.updater.Cleanup()
	a.downloads.Start()
	go a.prepareEngine()
	go a.checkAppUpdate()
}

func (a *App) shutdown(context.Context) {
	a.downloads.Shutdown()
	// A downloaded update is installed on any quit, not only through "Restart now".
	if err := a.updater.ApplyStaged(); err != nil {
		a.log.Error("applying update failed", "err", err)
	}
	if a.history != nil {
		_ = a.history.Close()
	}
	a.log.Info("shutdown")
	if a.logFile != nil {
		_ = a.logFile.Close()
	}
}

// prepareEngine installs missing tools on first run and checks for yt-dlp updates.
func (a *App) prepareEngine() {
	ctx := context.Background()
	st := a.engine.Status(ctx)
	a.emit("engine:status", st)
	if !st.Ready || !st.Deno.Installed {
		if err := a.engine.EnsureInstalled(ctx); err != nil {
			a.log.Error("engine install failed", "err", err)
		}
		a.emit("engine:status", a.engine.Status(ctx))
		a.engine.FillVersions(ctx)
		a.emit("engine:status", a.engine.Status(ctx))
		return
	}
	a.engine.FillVersions(ctx)
	a.emit("engine:status", a.engine.Status(ctx))
	if a.settings.Get().AutoUpdateEngine && st.YtDlp.Source == "managed" {
		cst, err := a.engine.CheckUpdates(ctx)
		if err == nil && cst.YtDlp.UpdateAvailable {
			a.log.Info("updating yt-dlp", "from", cst.YtDlp.Version, "to", cst.YtDlp.Latest)
			if err := a.engine.Update(ctx, false); err != nil {
				a.log.Warn("auto-update failed", "err", err)
			}
		}
		a.emit("engine:status", a.engine.Status(ctx))
	}
}

// checkAppUpdate looks for a new Dlyven release a few seconds after launch, off the
// startup path. It only checks: nothing is downloaded until the user asks.
func (a *App) checkAppUpdate() {
	time.Sleep(5 * time.Second)
	if !a.settings.Get().AutoCheckUpdates {
		return
	}
	st, err := a.updater.Check(context.Background())
	if err != nil {
		a.log.Warn("update check failed", "err", err)
	} else if st.Available {
		a.log.Info("update available", "version", st.Latest, "mode", st.Mode)
	}
}

func (a *App) onFinished(j downloader.Job) {
	a.emit("download:finished", j)
	if j.Status != downloader.StatusDone {
		return
	}
	s := a.settings.Get()
	if s.Notify {
		_ = sysutil.Notify(appinfo.Name, j.Request.Title)
	}
	if s.OpenFolderOnDone && j.FilePath != "" {
		_ = sysutil.RevealInFolder(j.FilePath)
	}
}

// ───────────────────────────── App info & settings ─────────────────────────────

type AppInfo struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	OS          string         `json:"os"`
	Arch        string         `json:"arch"`
	Translucent bool           `json:"translucent"`
	DataDir     string         `json:"dataDir"`
	LogPath     string         `json:"logPath"`
	Platforms   []platform.Def `json:"platforms"`
	Fields      []string       `json:"templateFields"`
	Browsers    []string       `json:"cookieBrowsers"`
}

func (a *App) GetAppInfo() AppInfo {
	lp := ""
	if a.logFile != nil {
		lp = a.logFile.Path()
	}
	return AppInfo{
		Name: appinfo.Name, Version: appinfo.Version, OS: goruntime.GOOS, Arch: goruntime.GOARCH,
		Translucent: a.translucent, DataDir: a.dataDir, LogPath: lp,
		Platforms: platform.Registry, Fields: filename.Fields, Browsers: settings.CookieBrowsers,
	}
}

func (a *App) GetSettings() settings.Settings { return a.settings.Get() }

func (a *App) SaveSettings(s settings.Settings) (settings.Settings, error) {
	return a.settings.Save(s)
}

func (a *App) ResetSettings() (settings.Settings, error) { return a.settings.Reset() }

// ───────────────────────────── Analysis ─────────────────────────────

func (a *App) DetectPlatform(url string) platform.Info { return platform.Detect(url) }

func (a *App) ExtractURLs(text string) []string { return platform.ExtractURLs(text) }

// Analyze cancels any analysis in flight and inspects url.
func (a *App) Analyze(url string) (*extractor.MediaInfo, error) {
	ctx, cancel := context.WithCancel(context.Background())
	a.analyzeMu.Lock()
	if a.analyzeCancel != nil {
		a.analyzeCancel()
	}
	a.analyzeCancel = cancel
	a.analyzeMu.Unlock()
	defer cancel()

	if a.engine.YtDlpPath() == "" && platform.Detect(url).ID != "spotify" {
		return nil, errs.New(errs.ErrEngineMissing, "")
	}
	return a.extractor.Analyze(ctx, url)
}

// AnalyzeForBatch analyzes without sharing the interactive cancel slot, so a batch import
// (dropped .txt with many links) and the Home screen can analyze independently.
func (a *App) AnalyzeForBatch(url string) (*extractor.MediaInfo, error) {
	return a.extractor.Analyze(context.Background(), url)
}

func (a *App) CancelAnalyze() {
	a.analyzeMu.Lock()
	defer a.analyzeMu.Unlock()
	if a.analyzeCancel != nil {
		a.analyzeCancel()
		a.analyzeCancel = nil
	}
}

// ───────────────────────────── Downloads ─────────────────────────────

func (a *App) rememberFolder(dir string) {
	if dir == "" {
		return
	}
	if s := a.settings.Get(); s.RememberLastFolder && s.LastFolder != dir {
		_, _ = a.settings.Update(func(s *settings.Settings) { s.LastFolder = dir })
	}
}

func (a *App) Enqueue(req downloader.Request) (downloader.Job, error) {
	j, err := a.downloads.Enqueue(req)
	if err == nil {
		a.rememberFolder(req.OutputDir)
	}
	return j, err
}

// EnqueueMany queues several requests (playlist items); stops at the first error.
func (a *App) EnqueueMany(reqs []downloader.Request) ([]downloader.Job, error) {
	out := make([]downloader.Job, 0, len(reqs))
	for _, r := range reqs {
		j, err := a.downloads.Enqueue(r)
		if err != nil {
			return out, err
		}
		out = append(out, j)
	}
	if len(reqs) > 0 {
		a.rememberFolder(reqs[0].OutputDir)
	}
	return out, nil
}

func (a *App) ListDownloads() []downloader.Job { return a.downloads.List() }
func (a *App) PauseDownload(id string) error   { return a.downloads.Pause(id) }
func (a *App) ResumeDownload(id string) error  { return a.downloads.Resume(id) }
func (a *App) CancelDownload(id string) error  { return a.downloads.Cancel(id) }
func (a *App) RetryDownload(id string) error   { return a.downloads.Retry(id) }
func (a *App) RemoveDownload(id string) error  { return a.downloads.Remove(id) }
func (a *App) ClearFinishedDownloads()         { a.downloads.ClearFinished() }
func (a *App) PreviewFilename(tpl string, vars map[string]string, ext string) string {
	name := filename.Render(tpl, filename.Vars(vars))
	if ext != "" {
		name += "." + ext
	}
	return name
}

// DefaultFolderFor returns the destination folder the app would use for a platform.
func (a *App) DefaultFolderFor(platformID string) string { return a.settings.Get().DirFor(platformID) }

// ───────────────────────────── Files & dialogs ─────────────────────────────

func (a *App) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (a *App) OpenFile(path string) error {
	if !a.FileExists(path) {
		return errs.New(errs.ErrNotFound, path)
	}
	return sysutil.OpenPath(path)
}

func (a *App) ShowInFolder(path string) error { return sysutil.RevealInFolder(path) }

func (a *App) OpenFolder(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return errs.Wrap(errs.ErrDisk, err)
	}
	return sysutil.OpenPath(path)
}

func (a *App) OpenLogsFolder() error { return sysutil.OpenPath(filepath.Join(a.dataDir, "logs")) }

func (a *App) dialogCtx() (context.Context, error) {
	c := a.ctx.Load()
	if c == nil {
		return nil, errs.New(errs.ErrUnknown, "window not ready")
	}
	return *c, nil
}

// ChooseFolder opens the native folder picker. Returns "" if the user cancels.
func (a *App) ChooseFolder(current string, title string) (string, error) {
	ctx, err := a.dialogCtx()
	if err != nil {
		return "", err
	}
	return runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		Title: title, DefaultDirectory: current, CanCreateDirectories: true,
	})
}

func (a *App) ChooseCookiesFile(title string) (string, error) {
	ctx, err := a.dialogCtx()
	if err != nil {
		return "", err
	}
	return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: []runtime.FileFilter{{DisplayName: "cookies.txt (Netscape)", Pattern: "*.txt"}, {DisplayName: "*", Pattern: "*"}},
	})
}

func (a *App) ChooseExecutable(title string) (string, error) {
	ctx, err := a.dialogCtx()
	if err != nil {
		return "", err
	}
	return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{Title: title})
}

// ───────────────────────────── Engine ─────────────────────────────

func (a *App) GetEngineStatus() engine.Status { return a.engine.Status(context.Background()) }

func (a *App) InstallEngine() (engine.Status, error) {
	err := a.engine.EnsureInstalled(context.Background())
	st := a.engine.Status(context.Background())
	a.emit("engine:status", st)
	return st, err
}

func (a *App) UpdateEngine(force bool) (engine.Status, error) {
	err := a.engine.Update(context.Background(), force)
	st := a.engine.Status(context.Background())
	a.emit("engine:status", st)
	return st, err
}

func (a *App) CheckEngineUpdates() (engine.Status, error) {
	return a.engine.CheckUpdates(context.Background())
}

// ───────────────────────────── App updates ─────────────────────────────

func (a *App) GetUpdateStatus() updater.Status { return a.updater.Status() }

func (a *App) CheckForUpdates() (updater.Status, error) {
	return a.updater.Check(context.Background())
}

// DownloadUpdate downloads and verifies the latest release. Progress arrives as
// "update:progress" events.
func (a *App) DownloadUpdate() (updater.Status, error) {
	return a.updater.Download(context.Background())
}

// RestartToUpdate installs the downloaded update and quits; the new version starts once
// this process is gone. Running downloads resume after the restart, as on any quit.
func (a *App) RestartToUpdate() error {
	ctx, err := a.dialogCtx()
	if err != nil {
		return err
	}
	if err := a.updater.Restart(); err != nil {
		return err
	}
	runtime.Quit(ctx)
	return nil
}

// ───────────────────────────── History ─────────────────────────────

func (a *App) ListHistory(q store.Query) (store.Page, error) {
	if a.history == nil {
		return store.Page{Items: []store.Entry{}}, nil
	}
	return a.history.List(context.Background(), q)
}

func (a *App) HistoryPlatforms() ([]string, error) {
	if a.history == nil {
		return []string{}, nil
	}
	return a.history.Platforms(context.Background())
}

func (a *App) DeleteHistory(id string) error {
	if a.history == nil {
		return nil
	}
	return a.history.Delete(context.Background(), id)
}

func (a *App) ClearHistory() error {
	if a.history == nil {
		return nil
	}
	return a.history.Clear(context.Background())
}

// ───────────────────────────── Diagnostics ─────────────────────────────

// GetDiagnostics returns a plain-text report for bug reports. Secrets are redacted.
func (a *App) GetDiagnostics() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n", appinfo.Name, appinfo.Version)
	fmt.Fprintf(&b, "OS: %s/%s  Go: %s\n", goruntime.GOOS, goruntime.GOARCH, goruntime.Version())
	fmt.Fprintf(&b, "Time: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&b, "Translucency: %v\n", a.translucent)
	us := a.updater.Status()
	fmt.Fprintf(&b, "Updates: mode=%s asset=%s latest=%q ready=%v\n\n", us.Mode, us.Asset, us.Latest, us.Ready)

	a.engine.FillVersions(context.Background())
	st := a.engine.Status(context.Background())
	for _, c := range []engine.ComponentStatus{st.YtDlp, st.FFmpeg, st.Deno} {
		fmt.Fprintf(&b, "%-7s installed=%v version=%q source=%s path=%s\n", c.Name, c.Installed, c.Version, c.Source, c.Path)
	}
	jobs := a.downloads.List()
	counts := map[downloader.Status]int{}
	for _, j := range jobs {
		counts[j.Status]++
	}
	fmt.Fprintf(&b, "\nQueue: %d jobs %v\n", len(jobs), counts)

	sj, _ := json.MarshalIndent(a.settings.Get().Redacted(), "", "  ")
	fmt.Fprintf(&b, "\nSettings:\n%s\n", sj)

	if a.logFile != nil {
		b.WriteString("\nRecent log:\n")
		for _, l := range logx.Tail(a.logFile.Path(), 150) {
			b.WriteString(l)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
