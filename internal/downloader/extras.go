package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/amirreza-azadmehr/dlyven/internal/appinfo"
	"github.com/amirreza-azadmehr/dlyven/internal/errs"
	"github.com/amirreza-azadmehr/dlyven/internal/filename"
	"github.com/amirreza-azadmehr/dlyven/internal/media"
	"github.com/amirreza-azadmehr/dlyven/internal/sysutil"
)

func guessImageExt(u string) string {
	if p, err := url.Parse(u); err == nil {
		switch strings.ToLower(strings.TrimPrefix(path.Ext(p.Path), ".")) {
		case "jpg", "jpeg":
			return "jpg"
		case "png":
			return "png"
		case "webp":
			return "webp"
		}
	}
	return "jpg"
}

func extFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "jpeg"):
		return "jpg"
	case strings.Contains(ct, "png"):
		return "png"
	case strings.Contains(ct, "webp"):
		return "webp"
	}
	return ""
}

// fetchToFile downloads an HTTP resource (through the configured proxy) and reports
// progress on j. Returns the detected image extension.
func (m *Manager) fetchToFile(ctx context.Context, j *jobState, rawURL, dest string) (string, error) {
	s := m.settings()
	client := sysutil.NewHTTPClient(s.ProxyURL(), time.Duration(s.TimeoutSec)*time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", errs.Wrap(errs.ErrInvalidURL, err)
	}
	if s.UserAgent != "" {
		req.Header.Set("User-Agent", s.UserAgent)
	}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", errs.New(errs.ErrCanceled, "")
		}
		return "", errs.New(errs.Classify(err.Error()), err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		code := errs.ErrNetwork
		if resp.StatusCode == http.StatusNotFound {
			code = errs.ErrNotFound
		}
		return "", errs.New(code, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", errs.Wrap(errs.ErrDisk, err)
	}
	defer f.Close()
	total := resp.ContentLength
	var done int64
	buf := make([]byte, 64*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return "", errs.Wrap(errs.ErrDisk, werr)
			}
			done += int64(n)
			if j != nil {
				m.mu.Lock()
				j.Downloaded = done
				if total > 0 {
					j.Total = total
					j.Progress = float64(done) / float64(total) * 99
				}
				m.emitLocked(j, false)
				m.mu.Unlock()
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			if ctx.Err() != nil {
				return "", errs.New(errs.ErrCanceled, "")
			}
			return "", errs.Wrap(errs.ErrNetwork, rerr)
		}
	}
	if err := f.Close(); err != nil {
		return "", errs.Wrap(errs.ErrDisk, err)
	}
	ext := extFromContentType(resp.Header.Get("Content-Type"))
	if ext == "" {
		ext = guessImageExt(rawURL)
	}
	return ext, nil
}

func (m *Manager) downloadThumbnail(ctx context.Context, j *jobState, req Request, outDir, tmpDir, base string) (string, error) {
	src := filepath.Join(tmpDir, "thumbnail.src")
	srcExt, err := m.fetchToFile(ctx, j, req.ThumbnailURL, src)
	if err != nil {
		return "", err
	}
	want := req.ThumbnailFormat
	if want != "jpg" && want != "png" {
		want = srcExt
	}
	m.mu.Lock()
	if want != j.Ext {
		// The real type differed from the URL guess: pick a fresh non-clashing name.
		j.Ext = want
		j.OutputBase = filename.Unique(outDir, base, want, nil)
		base = j.OutputBase
	}
	m.mu.Unlock()
	dst := filepath.Join(outDir, base+"."+want)
	if want == srcExt {
		if err := os.Rename(src, dst); err != nil {
			return "", errs.Wrap(errs.ErrDisk, err)
		}
	} else {
		m.setStatus(j, StatusConverting)
		if err := media.ConvertImage(ctx, m.eng.FFmpegPath(), src, dst); err != nil {
			return "", err
		}
	}
	m.cleanupPartial(j)
	return dst, nil
}

// tagSpotify writes Spotify metadata and cover art onto the matched audio file.
func (m *Manager) tagSpotify(ctx context.Context, req Request, tmpDir, path string) error {
	t := req.Spotify
	cover := ""
	if t.CoverURL != "" {
		p := filepath.Join(tmpDir, "cover.img")
		if _, err := m.fetchToFile(ctx, nil, t.CoverURL, p); err == nil {
			cover = p
		}
	}
	tags := media.Tags{
		"title":        t.Title,
		"artist":       strings.Join(t.Artists, ", "),
		"album":        firstNonEmpty(t.Album, req.Album),
		"album_artist": firstOr(t.Artists, ""),
		"comment":      "Audio matched from a public source by " + appinfo.Name,
	}
	if t.TrackNumber > 0 {
		tags["track"] = fmt.Sprint(t.TrackNumber)
	}
	if len(t.ReleaseDate) >= 4 {
		tags["date"] = t.ReleaseDate[:4]
	}
	return media.TagAudio(ctx, m.eng.FFmpegPath(), path, tags, cover)
}

func firstOr(a []string, def string) string {
	if len(a) > 0 {
		return a[0]
	}
	return def
}
