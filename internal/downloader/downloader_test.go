package downloader

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amirreza-azadmehr/dlyven/internal/platform"
	"github.com/amirreza-azadmehr/dlyven/internal/settings"
	"github.com/amirreza-azadmehr/dlyven/internal/store"
)

func TestParseProgressLine(t *testing.T) {
	p, ok := ParseProgressLine("DLYVEN|downloading|1048576|4194304|NA|524288.5|6|137|140")
	if !ok {
		t.Fatal("not parsed")
	}
	if p.Status != "downloading" || p.Downloaded != 1048576 || p.Total != 4194304 || p.Speed != 524288.5 || p.ETA != 6 || p.FormatID != "137" || p.SecondFormatID != "140" {
		t.Errorf("unexpected %+v", p)
	}

	// total missing → falls back to estimate; NA speed/eta; no merge
	p, ok = ParseProgressLine("[download] DLYVEN|downloading|100|NA|1000.0|NA|NA|18|NA\r")
	if !ok || p.Total != 1000 || p.Speed != 0 || p.ETA != -1 || p.SecondFormatID != "" {
		t.Errorf("unexpected %+v", p)
	}

	if _, ok := ParseProgressLine("[youtube] abc: Downloading webpage"); ok {
		t.Error("non-progress line parsed")
	}
	if _, ok := ParseProgressLine("DLYVEN|short"); ok {
		t.Error("truncated line parsed")
	}

	// Last record wins when CR-glued updates share one buffer.
	p, ok = ParseProgressLine("DLYVEN|finished|500|500|NA|NA|NA|sb0|NA\rDLYVEN|downloading|10485760|83886080|NA|524288|20|137|140|3|50")
	if !ok || p.Status != "downloading" || p.Downloaded != 10485760 || p.Total != 83886080 || p.Fragment != 3 || p.Fragments != 50 {
		t.Errorf("last record: %+v", p)
	}
}

func TestSplitProgressScan(t *testing.T) {
	in := []byte("DLYVEN|finished|500|500|NA|NA|NA|sb0|NA\rDLYVEN|downloading|100|800|NA|1|2|137|140\n")
	sc := bufio.NewScanner(bytes.NewReader(in))
	sc.Split(splitProgressScan)
	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if len(lines) != 2 || !strings.Contains(lines[0], "finished") || !strings.Contains(lines[1], "downloading") {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestParsePostprocessLine(t *testing.T) {
	cases := map[string]Status{
		"DLYVENPP|started|Merger":                          StatusMerging,
		"DLYVENPP|processing|FFmpegExtractAudio":           StatusConverting,
		"DLYVENPP|started|EmbedThumbnail":                  StatusConverting,
		"[Merger] Merging formats into \"x.mp4\"":          StatusMerging,
		"[ExtractAudio] Destination: x.mp3":                StatusConverting,
		"[VideoRemuxer] Not remuxing media file \"x.mp4\"": StatusConverting,
	}
	for line, want := range cases {
		got, ok := ParsePostprocessLine(line)
		if !ok || got != want {
			t.Errorf("%q → %v %v, want %v", line, got, ok, want)
		}
	}
	for _, line := range []string{"DLYVENPP|finished|Merger", "DLYVENPP|started|MoveFiles", "[youtube] Extracting URL", "plain"} {
		if _, ok := ParsePostprocessLine(line); ok {
			t.Errorf("%q should not change status", line)
		}
	}
}

func TestVideoFormatSelector(t *testing.T) {
	f, s := VideoFormatSelector(Request{FormatID: "137", HasAudio: false, Container: "mp4"})
	if f != "137+ba[ext=m4a]/137+ba/137" || s != "" {
		t.Errorf("got %q %q", f, s)
	}
	f, _ = VideoFormatSelector(Request{FormatID: "22", HasAudio: true})
	if f != "22" {
		t.Errorf("got %q", f)
	}
	f, s = VideoFormatSelector(Request{MaxHeight: 1080, Container: "mkv"})
	if f != "bv*+ba/b" || s != "res:1080,fps" {
		t.Errorf("got %q %q", f, s)
	}
	_, s = VideoFormatSelector(Request{Container: "mp4"})
	if s != "res,fps,vcodec:h264,ext:mp4:m4a" {
		t.Errorf("got %q", s)
	}
}

func TestAudioQualityArg(t *testing.T) {
	if AudioQualityArg("best") != "0" || AudioQualityArg("320") != "320K" || AudioQualityArg("") != "0" {
		t.Error("audio quality mapping wrong")
	}
}

func TestBuildArgsAudioSpotify(t *testing.T) {
	req := Request{Kind: KindAudio, AudioFormat: "m4a", AudioBitrate: "256", Spotify: &platform.SpotifyTrack{Title: "x"}, EmbedMetadata: true}
	args, ext := buildArgs(req, settings.Defaults(), "ytsearch1:a - x", "/out", "/out/.dlyven-partial/1", "50% Song")
	joined := strings.Join(args, " ")
	if ext != "m4a" {
		t.Errorf("ext = %s", ext)
	}
	for _, want := range []string{"-x --audio-format m4a", "--audio-quality 256K", "--playlist-items 1", "-o 50%% Song.%(ext)s", "-- ytsearch1:a - x"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in %s", want, joined)
		}
	}
	if strings.Contains(joined, "--embed-metadata") {
		t.Error("spotify jobs are tagged by Dlyven, not yt-dlp")
	}
	if args[len(args)-2] != "--" {
		t.Error("input must follow -- so URLs can't be parsed as options")
	}
}

func TestBuildArgsVideoWithSubs(t *testing.T) {
	req := Request{Kind: KindVideo, FormatID: "299", Container: "mkv", SubLangs: []string{"en", "fa"}, SubMode: "embed", AutoSubs: true, EmbedThumbnail: true, PlaylistItem: 3}
	s := settings.Defaults()
	s.SpeedLimitKBps = 500
	args, ext := buildArgs(req, s, "https://example.com/list", "/o", "/t", "base")
	joined := strings.Join(args, " ")
	for _, want := range []string{"--merge-output-format mkv", "--write-subs --sub-langs en,fa", "--write-auto-subs", "--embed-subs", "--embed-thumbnail", "--limit-rate 500K", "--playlist-items 3"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in %s", want, joined)
		}
	}
	if ext != "mkv" {
		t.Errorf("ext=%s", ext)
	}
}

func TestTemplateVars(t *testing.T) {
	v := TemplateVars(Request{Title: "T", Platform: "youtube", MaxHeight: 720, FPS: 59.94, PlaylistIndex: 4,
		Spotify: &platform.SpotifyTrack{Artists: []string{"A", "B"}, TrackNumber: 7}})
	if v["platform"] != "YouTube" || v["resolution"] != "720p" || v["fps"] != "60fps" || v["index"] != "04" || v["artist"] != "A, B" || v["track"] != "07" {
		t.Errorf("vars = %v", v)
	}
}

func TestApplyProgressMergesParts(t *testing.T) {
	m := &Manager{emit: func(string, any) {}}
	j := &jobState{Job: Job{Status: StatusDownloading}}
	m.applyProgress(j, Progress{Status: "downloading", Downloaded: 50_000_000, Total: 100_000_000, FormatID: "137", SecondFormatID: "140"})
	if j.Parts != 2 || j.Part != 1 || j.Progress != 25 {
		t.Fatalf("part1: %+v", j.Job)
	}
	m.applyProgress(j, Progress{Status: "downloading", Downloaded: 10_000_000, Total: 20_000_000, FormatID: "140", SecondFormatID: "140"})
	if j.Part != 2 || j.Progress != 75 || j.Downloaded != 110_000_000 || j.Total != 120_000_000 {
		t.Fatalf("part2: %+v", j.Job)
	}
}

func TestApplyProgressIgnoresThumbnailAndTracksHLS(t *testing.T) {
	m := &Manager{emit: func(string, any) {}}
	j := &jobState{Job: Job{Status: StatusDownloading}}
	m.applyProgress(j, Progress{Status: "finished", Downloaded: 40_000, Total: 40_000, FormatID: "137"})
	if j.Progress != 0 || j.Total != 0 {
		t.Fatalf("thumbnail should be ignored: %+v", j.Job)
	}
	m.applyProgress(j, Progress{Status: "downloading", Downloaded: 8_000_000, Total: 80_000_000, Speed: 1_000_000, ETA: 72, FormatID: "137"})
	if j.Progress < 9 || j.Progress > 11 || j.Downloaded != 8_000_000 || j.Total != 80_000_000 {
		t.Fatalf("media: %+v", j.Job)
	}
	m.applyProgress(j, Progress{Status: "downloading", Downloaded: 20_000_000, Fragment: 10, Fragments: 40, FormatID: "137"})
	if j.Progress < 24 || j.Progress > 26 {
		t.Fatalf("hls fragments should drive percent when total is missing: %+v", j.Job)
	}
}

func TestFinishedJobsSurviveRestore(t *testing.T) {
	dir := t.TempDir()
	queue := filepath.Join(dir, "queue.json")
	st, err := store.Open(filepath.Join(dir, "h.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	silent := slog.New(slog.NewTextHandler(io.Discard, nil))
	m := New(nil, func() settings.Settings { return settings.Settings{} }, st, func(string, any) {}, silent, queue, Hooks{})
	m.mu.Lock()
	m.jobs["done1"] = &jobState{Job: Job{ID: "done1", Status: StatusDone, Request: Request{Title: "clip", URL: "https://ex"}, FilePath: filepath.Join(dir, "clip.mp4"), Total: 42, Progress: 100}}
	m.order = []string{"done1"}
	m.historyImported = true
	m.persistLocked()
	m.mu.Unlock()

	m2 := New(nil, func() settings.Settings { return settings.Settings{} }, st, func(string, any) {}, silent, queue, Hooks{})
	m2.restore()
	list := m2.List()
	if len(list) != 1 || list[0].ID != "done1" || list[0].Status != StatusDone || list[0].Request.Title != "clip" {
		t.Fatalf("restored: %+v", list)
	}
}

func TestImportHistoryWhenQueueEmpty(t *testing.T) {
	dir := t.TempDir()
	queue := filepath.Join(dir, "queue.json")
	st, err := store.Open(filepath.Join(dir, "h.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if err := st.Upsert(ctx, store.Entry{ID: "old", URL: "https://ex", Title: "from db", Status: "done", FilePath: filepath.Join(dir, "a.mp4"), Size: 9, FinishedAt: 11}); err != nil {
		t.Fatal(err)
	}
	silent := slog.New(slog.NewTextHandler(io.Discard, nil))
	m := New(nil, func() settings.Settings { return settings.Settings{} }, st, func(string, any) {}, silent, queue, Hooks{})
	m.restore()
	list := m.List()
	if len(list) != 1 || list[0].ID != "old" || list[0].Status != StatusDone || list[0].Request.Title != "from db" {
		t.Fatalf("imported: %+v", list)
	}
	// Second launch must not duplicate, even after ClearFinished.
	m.ClearFinished()
	m3 := New(nil, func() settings.Settings { return settings.Settings{} }, st, func(string, any) {}, silent, queue, Hooks{})
	m3.restore()
	if n := len(m3.List()); n != 0 {
		t.Fatalf("cleared finished should stay gone, got %d", n)
	}
}
