package extractor

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/amirreza-azadmehr/dlyven/internal/engine"
	"github.com/amirreza-azadmehr/dlyven/internal/errs"
	"github.com/amirreza-azadmehr/dlyven/internal/platform"
	"github.com/amirreza-azadmehr/dlyven/internal/settings"
	"github.com/amirreza-azadmehr/dlyven/internal/sysutil"
)

type Extractor struct {
	eng      *engine.Manager
	settings func() settings.Settings
	log      *slog.Logger
}

func New(eng *engine.Manager, st func() settings.Settings, log *slog.Logger) *Extractor {
	return &Extractor{eng: eng, settings: st, log: log}
}

// Instagram/TikTok GraphQL can sit well past two minutes without browser impersonation.
const analyzeTimeout = 4 * time.Minute

// Analyze fetches metadata and available formats for a URL.
func (x *Extractor) Analyze(ctx context.Context, rawURL string) (*MediaInfo, error) {
	pi := platform.Detect(rawURL)
	if !pi.Valid {
		return nil, errs.New(errs.ErrInvalidURL, rawURL)
	}
	ctx, cancel := context.WithTimeout(ctx, analyzeTimeout)
	defer cancel()

	if pi.ID == "spotify" {
		return x.analyzeSpotify(ctx, pi)
	}

	s := x.settings()
	if err := x.eng.CheckCookies(s); err != nil {
		return nil, err
	}
	args := append(x.eng.BaseArgs(ctx, s), analyzeFlags(pi)...)
	args = append(args, "--", pi.URL)
	cmd, err := x.eng.Command(ctx, args...)
	if err != nil {
		return nil, err
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	start := time.Now()
	runErr := cmd.Run()
	x.log.Info("analyze", "platform", pi.ID, "ms", time.Since(start).Milliseconds(), "ok", runErr == nil)

	if ctx.Err() == context.Canceled {
		return nil, errs.New(errs.ErrCanceled, "")
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil, errs.New(errs.ErrTimeout, "analysis took too long")
	}
	if runErr != nil || stdout.Len() == 0 {
		out := stderr.String()
		code := errs.Classify(out)
		x.log.Warn("analyze failed", "code", code, "stderr", tail(out, 2000))
		return nil, errs.New(code, errs.LastErrorLine(out))
	}
	mi, err := ParseInfoJSON(stdout.Bytes())
	if err != nil {
		return nil, err
	}
	mi.Platform = pi
	if mi.URL == "" {
		mi.URL = pi.URL
	}
	if mi.IsLive {
		return nil, errs.New(errs.ErrLiveNotStarted, "live streams are not supported")
	}
	return mi, nil
}

func (x *Extractor) analyzeSpotify(ctx context.Context, pi platform.Info) (*MediaInfo, error) {
	s := x.settings()
	client := sysutil.NewHTTPClient(s.ProxyURL(), time.Duration(s.TimeoutSec)*time.Second)
	u, err := platform.ResolveSpotifyShortLink(ctx, client, pi.URL)
	if err != nil {
		return nil, err
	}
	kind, id, ok := platform.ParseSpotifyURL(u)
	if !ok {
		return nil, errs.New(errs.ErrUnsupportedURL, u)
	}
	col, err := platform.FetchSpotify(ctx, client, kind, id)
	if err != nil {
		return nil, err
	}
	mi := &MediaInfo{
		ID:          col.ID,
		URL:         u,
		Title:       col.Title,
		Uploader:    col.Owner,
		Thumbnail:   col.CoverURL,
		UploadDate:  dateOnly(col.Release),
		Platform:    pi,
		Extractor:   "Spotify",
		Matched:     true,
		Spotify:     col,
		Entries:     []Entry{},
		VideoGroups: []ResolutionGroup{},
		Audio:       []Format{},
		Subtitles:   []SubtitleTrack{},
		Thumbnails:  []Thumbnail{},
	}
	if col.CoverURL != "" {
		mi.Thumbnails = append(mi.Thumbnails, Thumbnail{URL: col.CoverURL, Width: 640, Height: 640, ID: "cover"})
	}
	if col.Kind == "track" && len(col.Tracks) == 1 {
		t := col.Tracks[0]
		mi.Duration = float64(t.DurationMs) / 1000
		mi.Uploader = strings.Join(t.Artists, ", ")
		mi.Entries = []Entry{{Index: 1, ID: t.ID, Title: t.Title, Duration: mi.Duration, Thumbnail: t.CoverURL, Uploader: mi.Uploader, Spotify: &col.Tracks[0]}}
		return mi, nil
	}
	mi.IsPlaylist = true
	var total float64
	for i := range col.Tracks {
		t := &col.Tracks[i]
		d := float64(t.DurationMs) / 1000
		total += d
		mi.Entries = append(mi.Entries, Entry{
			Index: i + 1, ID: t.ID, Title: t.Title, Duration: d,
			Thumbnail: t.CoverURL, Uploader: strings.Join(t.Artists, ", "), Spotify: t,
		})
	}
	mi.Duration = total
	return mi, nil
}

// analyzeFlags chooses playlist behaviour. --flat-playlist is for YouTube-style lists
// (metadata only). Instagram carousels need full format data, so they expand instead.
func analyzeFlags(pi platform.Info) []string {
	// Override the user socket timeout: Instagram GraphQL often needs >30s per request.
	args := []string{"--dump-single-json", "--socket-timeout", "90"}
	if platform.ExpandCollection(pi) {
		return append(args, "--yes-playlist")
	}
	return append(args, "--flat-playlist", "--no-playlist")
}

func dateOnly(iso string) string {
	if len(iso) >= 10 {
		return iso[:10]
	}
	return iso
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
