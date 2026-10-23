//go:build smoke

// Package smoke holds an opt-in end-to-end test against the real internet:
//
//	go test -tags smoke -v ./internal/smoke -timeout 20m
//
// It installs the engine into a temp dir, analyzes a public video and downloads it.
package smoke

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amirreza-azadmehr/dlyven/internal/downloader"
	"github.com/amirreza-azadmehr/dlyven/internal/engine"
	"github.com/amirreza-azadmehr/dlyven/internal/extractor"
	"github.com/amirreza-azadmehr/dlyven/internal/settings"
)

func TestEndToEnd(t *testing.T) {
	dir := t.TempDir()
	if d := os.Getenv("SMOKE_DIR"); d != "" {
		dir = d
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	s := settings.Defaults()
	s.DownloadDir = filepath.Join(dir, "out")
	get := func() settings.Settings { return s }
	eng := engine.New(filepath.Join(dir, "bin"), filepath.Join(dir, "cache"), get, func(string, any) {}, log)

	ctx := context.Background()
	if err := eng.EnsureInstalled(ctx); err != nil {
		t.Fatalf("install: %v", err)
	}
	st := eng.Status(ctx)
	t.Logf("engine: yt-dlp=%s ffmpeg=%s deno=%s", st.YtDlp.Version, st.FFmpeg.Version, st.Deno.Version)
	if !st.Ready {
		t.Fatalf("not ready: %+v", st)
	}

	url := os.Getenv("SMOKE_URL")
	if url == "" {
		url = "https://www.youtube.com/watch?v=jNQXAC9IVRw" // "Me at the zoo": tiny, public, stable
	}
	x := extractor.New(eng, get, log)
	mi, err := x.Analyze(ctx, url)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	t.Logf("title=%q groups=%d audio=%d subs=%d highlights=%+v", mi.Title, len(mi.VideoGroups), len(mi.Audio), len(mi.Subtitles), mi.Highlights)
	if len(mi.VideoGroups) == 0 {
		t.Fatal("no video formats")
	}

	done := make(chan downloader.Job, 1)
	m := downloader.New(eng, get, nil, func(name string, data any) {}, log, "", downloader.Hooks{OnFinished: func(j downloader.Job) { done <- j }})
	m.Start()
	_, err = m.Enqueue(downloader.Request{URL: mi.URL, Title: mi.Title, Platform: mi.Platform.ID, Kind: downloader.KindVideo,
		MaxHeight: 360, Container: "mp4", QualityLabel: "360p", EmbedMetadata: true, EmbedThumbnail: true})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case j := <-done:
		t.Logf("job: status=%s file=%s size=%d err=%s %s", j.Status, j.FilePath, j.Total, j.ErrorCode, j.ErrorDetail)
		if j.Status != downloader.StatusDone {
			t.Fatal("download failed")
		}
		if _, err := os.Stat(j.FilePath); err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Minute):
		t.Fatal("timeout")
	}
}

func TestSpotifyMatch(t *testing.T) {
	dir := t.TempDir()
	if d := os.Getenv("SMOKE_DIR"); d != "" {
		dir = d
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	s := settings.Defaults()
	s.DownloadDir = filepath.Join(dir, "out")
	get := func() settings.Settings { return s }
	eng := engine.New(filepath.Join(dir, "bin"), filepath.Join(dir, "cache"), get, func(string, any) {}, log)
	ctx := context.Background()
	if err := eng.EnsureInstalled(ctx); err != nil {
		t.Fatalf("install: %v", err)
	}
	url := os.Getenv("SMOKE_SPOTIFY")
	if url == "" {
		url = "https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC"
	}
	mi, err := extractor.New(eng, get, log).Analyze(ctx, url)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if !mi.Matched || len(mi.Entries) == 0 || mi.Entries[0].Spotify == nil {
		t.Fatalf("unexpected: %+v", mi)
	}
	tr := mi.Entries[0].Spotify
	t.Logf("spotify: %q by %v (%d ms) cover=%v", tr.Title, tr.Artists, tr.DurationMs, tr.CoverURL != "")

	done := make(chan downloader.Job, 1)
	m := downloader.New(eng, get, nil, func(string, any) {}, log, "", downloader.Hooks{OnFinished: func(j downloader.Job) { done <- j }})
	m.Start()
	if _, err := m.Enqueue(downloader.Request{Title: tr.Title, Platform: "spotify", Kind: downloader.KindAudio,
		AudioFormat: "mp3", AudioBitrate: "192", Spotify: tr, FilenameTemplate: "{artist} - {title}"}); err != nil {
		t.Fatal(err)
	}
	select {
	case j := <-done:
		t.Logf("job: status=%s file=%s err=%s %s", j.Status, j.FilePath, j.ErrorCode, j.ErrorDetail)
		if j.Status != downloader.StatusDone {
			t.Fatal("download failed")
		}
	case <-time.After(10 * time.Minute):
		t.Fatal("timeout")
	}
}
