package downloader

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/amirreza-azadmehr/dlyven/internal/engine"
	"github.com/amirreza-azadmehr/dlyven/internal/errs"
	"github.com/amirreza-azadmehr/dlyven/internal/filename"
	"github.com/amirreza-azadmehr/dlyven/internal/platform"
	"github.com/amirreza-azadmehr/dlyven/internal/settings"
	"github.com/amirreza-azadmehr/dlyven/internal/store"
)

// Event names emitted to the frontend.
const (
	EventUpdate  = "download:update"  // payload: Job
	EventRemoved = "download:removed" // payload: job ID
)

// partialDirName holds in-progress files next to the destination so the final move is a
// cheap same-filesystem rename (not a multi-GB cross-drive copy).
const partialDirName = ".dlyven-partial"

const emitInterval = 100 * time.Millisecond // ≈10 Hz per job

type stopReason int

const (
	stopNone stopReason = iota
	stopPause
	stopCancel
	stopRemove
	stopShutdown
)

type jobState struct {
	Job
	running  bool
	cancel   context.CancelFunc
	stop     stopReason
	lastEmit time.Time
	tmpDir   string
	outDir   string

	// progress bookkeeping for multi-stream (video+audio) downloads
	seen      []string
	curPart   int
	curTotal  int64
	doneBytes int64
	errTail   []string
}

type Hooks struct {
	// OnFinished is called (outside the lock) when a job reaches done or failed.
	OnFinished func(Job)
}

type Manager struct {
	mu    sync.Mutex
	jobs  map[string]*jobState
	order []string

	eng       *engine.Manager
	settings  func() settings.Settings
	history   *store.Store
	log       *slog.Logger
	emit      func(string, any)
	hooks     Hooks
	queueFile string

	ctx    context.Context
	cancel context.CancelFunc
	wake   chan struct{}
	wg     sync.WaitGroup

	// historyImported is persisted so a one-time backfill from SQLite history
	// does not re-add items the user already cleared from the download list.
	historyImported bool
}

func New(eng *engine.Manager, st func() settings.Settings, history *store.Store, emit func(string, any), log *slog.Logger, queueFile string, hooks Hooks) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		jobs: map[string]*jobState{}, eng: eng, settings: st, history: history, emit: emit, log: log,
		queueFile: queueFile, hooks: hooks, ctx: ctx, cancel: cancel, wake: make(chan struct{}, 1),
	}
}

// Start restores the persisted queue and launches the scheduler.
func (m *Manager) Start() {
	m.restore()
	go m.scheduler()
	m.Wake()
}

// Wake asks the scheduler to re-evaluate (e.g. after a concurrency change).
func (m *Manager) Wake() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Enqueue validates a request, plans its output file name and queues it.
func (m *Manager) Enqueue(req Request) (Job, error) {
	if strings.TrimSpace(req.URL) == "" && req.Spotify == nil && req.Kind != KindThumbnail {
		return Job{}, errs.New(errs.ErrInvalidURL, "")
	}
	if req.Kind == "" {
		req.Kind = KindVideo
	}
	if req.Kind == KindSubtitles && len(req.SubLangs) == 0 {
		return Job{}, errs.New(errs.ErrFormatUnavailable, "no subtitle languages selected")
	}
	if req.Kind == KindThumbnail && req.ThumbnailURL == "" {
		return Job{}, errs.New(errs.ErrFormatUnavailable, "no thumbnail selected")
	}
	if req.Spotify != nil {
		req.Kind = KindAudio
	}
	s := m.settings()
	outDir := strings.TrimSpace(req.OutputDir)
	if outDir == "" {
		outDir = s.DirFor(req.Platform)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return Job{}, errs.Wrap(errs.ErrDisk, err)
	}
	req.OutputDir = outDir
	if req.FilenameTemplate == "" {
		req.FilenameTemplate = s.FilenameTemplate
	}

	ext := expectedExt(req)
	m.mu.Lock()
	defer m.mu.Unlock()
	base := filename.Render(req.FilenameTemplate, TemplateVars(req))
	base = filename.Unique(outDir, base, ext, m.reservedLocked)

	j := &jobState{
		Job: Job{
			ID: newID(), Request: req, Status: StatusQueued, ETA: -1, Parts: 1, Part: 1,
			OutputBase: base, Ext: ext, CreatedAt: time.Now().UnixMilli(),
		},
		outDir: outDir,
	}
	j.tmpDir = filepath.Join(outDir, partialDirName, j.ID)
	m.jobs[j.ID] = j
	m.order = append(m.order, j.ID)
	m.emitLocked(j, true)
	m.persistLocked()
	m.Wake()
	return j.Job, nil
}

func expectedExt(req Request) string {
	switch req.Kind {
	case KindAudio:
		return audioFormatOf(req)
	case KindSubtitles:
		return subFormat(req)
	case KindThumbnail:
		switch req.ThumbnailFormat {
		case "jpg", "png":
			return req.ThumbnailFormat
		}
		return guessImageExt(req.ThumbnailURL)
	}
	return containerOf(req)
}

// TemplateVars builds the values for filename templates.
func TemplateVars(req Request) filename.Vars {
	v := filename.Vars{
		"title":    req.Title,
		"uploader": req.Uploader,
		"quality":  req.QualityLabel,
		"id":       req.MediaID,
		"date":     req.UploadDate,
		"playlist": req.PlaylistTitle,
	}
	if d, ok := platform.Lookup(req.Platform); ok {
		v["platform"] = d.Name
	}
	if req.Height > 0 {
		v["resolution"] = fmt.Sprintf("%dp", req.Height)
	} else if req.MaxHeight > 0 {
		v["resolution"] = fmt.Sprintf("%dp", req.MaxHeight)
	}
	if req.FPS > 0 {
		v["fps"] = fmt.Sprintf("%dfps", int(math.Round(req.FPS)))
	}
	if req.PlaylistIndex > 0 {
		v["index"] = fmt.Sprintf("%02d", req.PlaylistIndex)
	}
	if t := req.Spotify; t != nil {
		v["artist"] = strings.Join(t.Artists, ", ")
		v["album"] = firstNonEmpty(t.Album, req.Album)
		if t.TrackNumber > 0 {
			v["track"] = fmt.Sprintf("%02d", t.TrackNumber)
		}
		if v["uploader"] == "" {
			v["uploader"] = v["artist"]
		}
	}
	return v
}

func firstNonEmpty(a ...string) string {
	for _, s := range a {
		if s != "" {
			return s
		}
	}
	return ""
}

func (m *Manager) reservedLocked(lowerPath string) bool {
	for _, j := range m.jobs {
		if j.Status.Final() && j.Status != StatusDone {
			continue
		}
		p := strings.ToLower(filepath.Join(j.outDir, j.OutputBase+"."+j.Ext))
		if p == lowerPath {
			return true
		}
	}
	return false
}

func (m *Manager) List() []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Job, 0, len(m.order))
	for _, id := range m.order {
		out = append(out, m.jobs[id].Job)
	}
	return out
}

func (m *Manager) get(id string) (*jobState, error) {
	j, ok := m.jobs[id]
	if !ok {
		return nil, errs.New(errs.ErrNotFoundJob, id)
	}
	return j, nil
}

func (m *Manager) Pause(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, err := m.get(id)
	if err != nil {
		return err
	}
	switch {
	case j.running:
		j.stop = stopPause
		j.cancel()
	case j.Status == StatusQueued || j.Status == StatusRetrying:
		j.Status = StatusPaused
		j.Speed, j.ETA = 0, -1
		m.emitLocked(j, true)
		m.persistLocked()
	}
	return nil
}

func (m *Manager) Resume(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, err := m.get(id)
	if err != nil {
		return err
	}
	if j.Status == StatusPaused {
		j.Status = StatusQueued
		m.emitLocked(j, true)
		m.persistLocked()
		m.Wake()
	}
	return nil
}

func (m *Manager) Cancel(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, err := m.get(id)
	if err != nil {
		return err
	}
	if j.running {
		j.stop = stopCancel
		j.cancel()
		return nil
	}
	if !j.Status.Final() {
		j.Status = StatusCanceled
		j.Speed, j.ETA = 0, -1
		j.FinishedAt = time.Now().UnixMilli()
		m.cleanupPartial(j)
		m.emitLocked(j, true)
		m.persistLocked()
	}
	return nil
}

func (m *Manager) Retry(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, err := m.get(id)
	if err != nil {
		return err
	}
	if j.Status != StatusFailed && j.Status != StatusCanceled {
		return nil
	}
	j.Status = StatusQueued
	j.Progress, j.Downloaded, j.Total, j.Speed, j.ETA = 0, 0, 0, 0, -1
	j.ErrorCode, j.ErrorDetail = "", ""
	j.Attempts = 0
	j.FinishedAt = 0
	m.emitLocked(j, true)
	m.persistLocked()
	m.Wake()
	return nil
}

// Remove deletes a job from the list; a running job is canceled first.
func (m *Manager) Remove(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, err := m.get(id)
	if err != nil {
		return err
	}
	if j.running {
		j.stop = stopRemove
		j.cancel()
		return nil
	}
	if j.Status != StatusDone {
		m.cleanupPartial(j)
	}
	m.removeLocked(id)
	return nil
}

// ClearFinished removes done, failed and canceled jobs from the list.
func (m *Manager) ClearFinished() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range append([]string{}, m.order...) {
		if j := m.jobs[id]; j.Status.Final() {
			if j.Status != StatusDone {
				m.cleanupPartial(j)
			}
			m.removeLocked(id)
		}
	}
}

func (m *Manager) removeLocked(id string) {
	delete(m.jobs, id)
	for i, o := range m.order {
		if o == id {
			m.order = append(m.order[:i], m.order[i+1:]...)
			break
		}
	}
	m.emit(EventRemoved, id)
	m.persistLocked()
}

// Shutdown stops running jobs (keeping partial files so they resume next launch).
func (m *Manager) Shutdown() {
	m.mu.Lock()
	for _, j := range m.jobs {
		if j.running {
			j.stop = stopShutdown
			j.cancel()
		}
	}
	m.mu.Unlock()
	done := make(chan struct{})
	go func() { m.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	m.cancel()
	m.mu.Lock()
	m.persistLocked()
	m.mu.Unlock()
}

func (m *Manager) scheduler() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.wake:
		}
		m.mu.Lock()
		limit := m.settings().MaxConcurrent
		running := 0
		for _, j := range m.jobs {
			if j.running {
				running++
			}
		}
		for _, id := range m.order {
			if running >= limit {
				break
			}
			if j := m.jobs[id]; j.Status == StatusQueued && !j.running {
				m.startLocked(j)
				running++
			}
		}
		m.mu.Unlock()
	}
}

func (m *Manager) startLocked(j *jobState) {
	ctx, cancel := context.WithCancel(m.ctx)
	j.cancel = cancel
	j.running = true
	j.stop = stopNone
	j.Status = StatusDownloading
	j.Attempts++
	j.StartedAt = time.Now().UnixMilli()
	j.Speed, j.ETA = 0, -1
	j.seen, j.curPart, j.curTotal, j.doneBytes, j.errTail = nil, 0, 0, 0, nil
	m.emitLocked(j, true)
	m.persistLocked()
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		path, err := m.execute(ctx, j)
		m.finish(j, path, err)
	}()
}

func (m *Manager) finish(j *jobState, path string, runErr error) {
	m.mu.Lock()
	j.running = false
	j.cancel = nil
	j.Speed, j.ETA = 0, -1
	reason := j.stop
	j.stop = stopNone
	var finished bool

	switch reason {
	case stopRemove:
		m.cleanupPartial(j)
		m.removeLocked(j.ID)
		m.mu.Unlock()
		m.Wake()
		return
	case stopCancel:
		j.Status = StatusCanceled
		j.FinishedAt = time.Now().UnixMilli()
		m.cleanupPartial(j)
	case stopPause, stopShutdown:
		j.Status = StatusPaused
		if reason == stopShutdown {
			j.Status = StatusQueued // resume automatically on next launch
		}
	default:
		if runErr == nil {
			j.Status = StatusDone
			j.Progress = 100
			j.FilePath = path
			if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
				j.Total, j.Downloaded = fi.Size(), fi.Size()
			}
			j.FinishedAt = time.Now().UnixMilli()
			finished = true
		} else {
			code := errs.CodeOf(runErr)
			j.ErrorCode, j.ErrorDetail = string(code), errs.DetailOf(runErr)
			retries := m.settings().Retries
			if errs.Retryable(code) && j.Attempts <= retries {
				j.Status = StatusRetrying
				delay := time.Duration(math.Min(math.Pow(2, float64(j.Attempts)), 30)) * time.Second
				m.log.Warn("download failed, retrying", "id", j.ID, "code", code, "attempt", j.Attempts, "delay", delay)
				time.AfterFunc(delay, func() {
					m.mu.Lock()
					if j.Status == StatusRetrying && !j.running {
						j.Status = StatusQueued
						m.emitLocked(j, true)
					}
					m.mu.Unlock()
					m.Wake()
				})
			} else {
				j.Status = StatusFailed
				j.FinishedAt = time.Now().UnixMilli()
				finished = true
				m.log.Error("download failed", "id", j.ID, "code", code, "detail", j.ErrorDetail)
			}
		}
	}
	m.emitLocked(j, true)
	m.persistLocked()
	snapshot := j.Job
	m.mu.Unlock()
	m.Wake()

	if finished {
		m.recordHistory(snapshot)
		if m.hooks.OnFinished != nil {
			m.hooks.OnFinished(snapshot)
		}
	}
}

func (m *Manager) recordHistory(j Job) {
	if m.history == nil {
		return
	}
	r := j.Request
	err := m.history.Upsert(context.Background(), store.Entry{
		ID: j.ID, URL: r.URL, Title: r.Title, Uploader: r.Uploader, Platform: r.Platform, Kind: string(r.Kind),
		Quality: r.QualityLabel, Thumbnail: r.Thumbnail, FilePath: j.FilePath, Size: j.Total,
		Status: string(j.Status), ErrorCode: j.ErrorCode, CreatedAt: j.CreatedAt, FinishedAt: j.FinishedAt,
	})
	if err != nil {
		m.log.Error("history write failed", "err", err)
	}
}

// cleanupPartial removes the job's temp directory (partial downloads) and the shared
// partial folder when it becomes empty.
func (m *Manager) cleanupPartial(j *jobState) {
	if j.tmpDir == "" {
		return
	}
	_ = os.RemoveAll(j.tmpDir)
	_ = os.Remove(filepath.Dir(j.tmpDir)) // only succeeds when empty
}

func (m *Manager) emitLocked(j *jobState, force bool) {
	if !force && time.Since(j.lastEmit) < emitInterval {
		return
	}
	j.lastEmit = time.Now()
	m.emit(EventUpdate, j.Job)
}

// execute performs the download and returns the final file path.
func (m *Manager) execute(ctx context.Context, j *jobState) (string, error) {
	m.mu.Lock()
	req, outDir, tmpDir, base := j.Request, j.outDir, j.tmpDir, j.OutputBase
	m.mu.Unlock()

	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", errs.Wrap(errs.ErrDisk, err)
	}
	if req.Kind == KindThumbnail {
		return m.downloadThumbnail(ctx, j, req, outDir, tmpDir, base)
	}
	if m.eng.FFmpegPath() == "" && req.Kind != KindSubtitles {
		return "", errs.New(errs.ErrFFmpegMissing, "")
	}

	inputs := []string{req.URL}
	if req.Spotify != nil {
		inputs = platform.MatchSources(*req.Spotify)
	}
	var lastErr error
	for i, input := range inputs {
		lastErr = m.runYtDlp(ctx, j, req, input, outDir, tmpDir, base)
		if lastErr == nil {
			break
		}
		if ctx.Err() != nil {
			return "", errs.New(errs.ErrCanceled, "")
		}
		if req.Spotify == nil || i == len(inputs)-1 {
			break
		}
		m.log.Info("spotify match source failed, trying next", "id", j.ID, "err", lastErr)
	}
	if lastErr != nil {
		if req.Spotify != nil && !errs.Retryable(errs.CodeOf(lastErr)) {
			return "", errs.New(errs.ErrSpotifyMatch, errs.DetailOf(lastErr))
		}
		return "", lastErr
	}

	path := findOutput(req, outDir, tmpDir, base)
	if path == "" {
		return "", errs.New(errs.ErrUnknown, "download finished but the output file was not found")
	}
	if req.Spotify != nil {
		m.setStatus(j, StatusConverting)
		if err := m.tagSpotify(ctx, req, tmpDir, path); err != nil {
			// Tagging is best-effort: the audio file is still valid.
			m.log.Warn("spotify tagging failed", "id", j.ID, "err", err)
		}
	}
	m.cleanupPartial(j)
	return path, nil
}

func (m *Manager) setStatus(j *jobState, st Status) {
	m.mu.Lock()
	j.Status = st
	m.emitLocked(j, true)
	m.mu.Unlock()
}

func (m *Manager) runYtDlp(ctx context.Context, j *jobState, req Request, input, outDir, tmpDir, base string) error {
	s := m.settings()
	if err := m.eng.CheckCookies(s); err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(tmpDir, resultFile))
	args, _ := buildArgs(req, s, input, outDir, tmpDir, base)
	args = append(m.eng.BaseArgs(ctx, s), args...)
	cmd, err := m.eng.Command(ctx, args...)
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	m.log.Debug("yt-dlp start", "id", j.ID, "kind", req.Kind)
	if err := cmd.Start(); err != nil {
		return errs.Wrap(errs.ErrEngineMissing, err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); m.consume(j, stdout) }()
	go func() { defer wg.Done(); m.consume(j, stderr) }()
	wg.Wait()
	waitErr := cmd.Wait()

	if ctx.Err() != nil {
		return errs.New(errs.ErrCanceled, "")
	}
	if waitErr != nil {
		m.mu.Lock()
		out := strings.Join(j.errTail, "\n")
		m.mu.Unlock()
		code := errs.Classify(out)
		m.log.Warn("yt-dlp failed", "id", j.ID, "code", code, "output", out)
		return errs.New(code, errs.LastErrorLine(out))
	}
	return nil
}

func (m *Manager) consume(j *jobState, r io.Reader) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	sc.Split(splitProgressScan)
	for sc.Scan() {
		m.handleLine(j, strings.TrimSpace(sc.Text()))
	}
}

func (m *Manager) handleLine(j *jobState, line string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := ParseProgressLine(line); ok {
		m.applyProgress(j, p)
		return
	}
	if st, ok := ParsePostprocessLine(line); ok {
		if j.Status != st {
			j.Status = st
			j.Speed, j.ETA = 0, -1
			m.emitLocked(j, true)
		}
		return
	}
	if strings.TrimSpace(line) == "" {
		return
	}
	j.errTail = append(j.errTail, line)
	if len(j.errTail) > 60 {
		j.errTail = j.errTail[len(j.errTail)-60:]
	}
}

func (m *Manager) applyProgress(j *jobState, p Progress) {
	if p.auxiliary() {
		return
	}
	part, parts := 1, 1
	if p.SecondFormatID != "" {
		parts = 2
		if p.FormatID == p.SecondFormatID {
			part = 2
		}
	} else if p.FormatID != "" {
		idx := -1
		for i, f := range j.seen {
			if f == p.FormatID {
				idx = i
			}
		}
		if idx < 0 {
			j.seen = append(j.seen, p.FormatID)
			idx = len(j.seen) - 1
		}
		part, parts = idx+1, max(len(j.seen), 1)
	}
	if part != j.curPart {
		if j.curPart > 0 && part > j.curPart {
			j.doneBytes += j.curTotal
		}
		j.curPart, j.curTotal = part, 0
	}
	if p.Total > 0 {
		j.curTotal = p.Total
	}
	frac := p.fraction()
	overall := (float64(part-1) + frac) / float64(parts) * 100
	// Allow the bar to drop when a sidecar file (thumb, json) was followed by the real media.
	if part != j.Part || overall >= j.Progress-0.3 || j.Progress-overall > 2 {
		j.Progress = math.Min(overall, 99)
	}
	j.Part, j.Parts = part, parts
	if p.Total > 0 || p.Downloaded > 0 {
		j.Downloaded = j.doneBytes + p.Downloaded
	}
	switch {
	case p.Total > 0:
		j.Total = j.doneBytes + p.Total
	case j.Downloaded > j.Total:
		// Stale tiny total from an earlier sidecar; unknown overall size (typical HLS).
		j.Total = 0
	}
	j.Speed, j.ETA = p.Speed, p.ETA
	if j.Status != StatusDownloading && p.Status == "downloading" {
		j.Status = StatusDownloading
		m.emitLocked(j, true)
		return
	}
	m.emitLocked(j, false)
}

// findOutput locates the produced file: first from yt-dlp's after_move print, then by name.
func findOutput(req Request, outDir, tmpDir, base string) string {
	if b, err := os.ReadFile(filepath.Join(tmpDir, resultFile)); err == nil {
		lines := strings.Split(strings.TrimSpace(string(b)), "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			p := strings.TrimSpace(lines[i])
			if p != "" && fileExists(p) {
				return p
			}
		}
	}
	matches, _ := filepath.Glob(filepath.Join(outDir, globEscape(base)+".*"))
	best := ""
	for _, mt := range matches {
		ext := strings.ToLower(filepath.Ext(mt))
		if ext == ".part" || ext == ".ytdl" || ext == ".tmp" || strings.Contains(mt, ".tagging.") {
			continue
		}
		if req.Kind == KindSubtitles && (ext == ".srt" || ext == ".vtt") {
			return mt
		}
		if best == "" {
			best = mt
		}
	}
	return best
}

func globEscape(s string) string {
	r := strings.NewReplacer(`[`, `[[]`, `*`, `[*]`, `?`, `[?]`)
	if os.PathSeparator != '\\' {
		r = strings.NewReplacer(`\`, `\\`, `[`, `\[`, `*`, `\*`, `?`, `\?`)
	}
	return r.Replace(s)
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

type persisted struct {
	Jobs            []persistedJob `json:"jobs"`
	HistoryImported bool           `json:"historyImported"`
}

type persistedJob struct {
	Job    Job    `json:"job"`
	OutDir string `json:"outDir"`
	TmpDir string `json:"tmpDir"`
}

// persistLocked writes the whole list (active and finished) so completed
// downloads are still there after the app is closed and opened again.
func (m *Manager) persistLocked() {
	if m.queueFile == "" {
		return
	}
	p := persisted{HistoryImported: m.historyImported}
	for _, id := range m.order {
		j := m.jobs[id]
		p.Jobs = append(p.Jobs, persistedJob{Job: j.Job, OutDir: j.outDir, TmpDir: j.tmpDir})
	}
	b, err := json.Marshal(p)
	if err != nil {
		return
	}
	tmp := m.queueFile + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err == nil {
		_ = os.Rename(tmp, m.queueFile)
	}
}

func (m *Manager) restore() {
	if m.queueFile == "" {
		return
	}
	imported := false
	if b, err := os.ReadFile(m.queueFile); err == nil {
		var p persisted
		if err := json.Unmarshal(b, &p); err != nil {
			m.log.Warn("queue restore failed", "err", err)
		} else {
			imported = p.HistoryImported
			m.mu.Lock()
			for _, pj := range p.Jobs {
				j := &jobState{Job: pj.Job, outDir: pj.OutDir, tmpDir: pj.TmpDir}
				if j.Status.Active() {
					j.Status = StatusQueued
				}
				j.Speed, j.ETA = 0, -1
				m.jobs[j.ID] = j
				m.order = append(m.order, j.ID)
			}
			m.mu.Unlock()
			m.log.Info("queue restored", "jobs", len(p.Jobs))
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !imported {
		n := m.importHistoryLocked()
		if n > 0 {
			m.log.Info("history imported into queue", "jobs", n)
		}
	}
	m.historyImported = true
	m.persistLocked()
}

func (m *Manager) importHistoryLocked() int {
	if m.history == nil {
		return 0
	}
	page, err := m.history.List(context.Background(), store.Query{Limit: 500})
	if err != nil {
		m.log.Warn("history import failed", "err", err)
		return 0
	}
	added := 0
	for i := len(page.Items) - 1; i >= 0; i-- {
		e := page.Items[i]
		if _, ok := m.jobs[e.ID]; ok {
			continue
		}
		j := jobFromHistory(e)
		m.jobs[j.ID] = j
		m.order = append(m.order, j.ID)
		added++
	}
	return added
}

func jobFromHistory(e store.Entry) *jobState {
	kind := Kind(e.Kind)
	if kind == "" {
		kind = KindVideo
	}
	st := Status(e.Status)
	progress := 0.0
	if st == StatusDone {
		progress = 100
	}
	outDir := ""
	if e.FilePath != "" {
		outDir = filepath.Dir(e.FilePath)
	}
	return &jobState{
		Job: Job{
			ID: e.ID,
			Request: Request{
				URL: e.URL, Title: e.Title, Uploader: e.Uploader, Thumbnail: e.Thumbnail,
				Platform: e.Platform, Kind: kind, QualityLabel: e.Quality, OutputDir: outDir,
			},
			Status: st, Progress: progress, Downloaded: e.Size, Total: e.Size, ETA: -1,
			FilePath: e.FilePath, ErrorCode: e.ErrorCode, CreatedAt: e.CreatedAt, FinishedAt: e.FinishedAt,
		},
		outDir: outDir,
	}
}
