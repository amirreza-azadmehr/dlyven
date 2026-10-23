package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/amirreza-azadmehr/dlyven/internal/errs"
)

// Spotify content is DRM-protected and is never downloaded from Spotify. Instead we read the
// public metadata shown on Spotify's embed pages, match each track to a publicly available
// audio source (YouTube Music search through yt-dlp) and tag the result with that metadata.

type SpotifyTrack struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Artists     []string `json:"artists"`
	Album       string   `json:"album"`
	DurationMs  int64    `json:"durationMs"`
	CoverURL    string   `json:"coverUrl"`
	TrackNumber int      `json:"trackNumber"`
	ReleaseDate string   `json:"releaseDate"`
	Explicit    bool     `json:"explicit"`
}

type SpotifyCollection struct {
	Kind     string         `json:"kind"` // track | album | playlist | artist
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Owner    string         `json:"owner"`
	CoverURL string         `json:"coverUrl"`
	Release  string         `json:"releaseDate"`
	Tracks   []SpotifyTrack `json:"tracks"`
}

var spotifyPathRe = regexp.MustCompile(`^/(?:intl-[a-z-]+/)?(track|album|playlist|artist|episode|show)/([A-Za-z0-9]+)`)

// ParseSpotifyURL extracts kind and ID from an open.spotify.com URL or a spotify: URI.
func ParseSpotifyURL(raw string) (kind, id string, ok bool) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(strings.ToLower(raw), "spotify:") {
		parts := strings.Split(raw, ":")
		if len(parts) >= 3 {
			return strings.ToLower(parts[1]), parts[2], true
		}
		return "", "", false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", false
	}
	m := spotifyPathRe.FindStringSubmatch(u.Path)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}

// ResolveSpotifyShortLink follows spotify.link redirects to the canonical open.spotify.com URL.
func ResolveSpotifyShortLink(ctx context.Context, client *http.Client, raw string) (string, error) {
	if !strings.Contains(strings.ToLower(raw), "spotify.link") {
		return raw, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", browserUA)
	resp, err := client.Do(req)
	if err != nil {
		return "", errs.Wrap(errs.ErrNetwork, err)
	}
	defer resp.Body.Close()
	return resp.Request.URL.String(), nil
}

const browserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36"

// FetchSpotify downloads the public embed page for an item and parses its metadata.
func FetchSpotify(ctx context.Context, client *http.Client, kind, id string) (*SpotifyCollection, error) {
	switch kind {
	case "track", "album", "playlist", "artist":
	default:
		return nil, errs.New(errs.ErrSpotifyUnsupported, kind)
	}
	endpoint := fmt.Sprintf("https://open.spotify.com/embed/%s/%s", kind, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept-Language", "en")
	resp, err := client.Do(req)
	if err != nil {
		return nil, errs.Wrap(errs.ErrNetwork, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errs.New(errs.ErrNotFound, endpoint)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errs.New(errs.ErrNetwork, fmt.Sprintf("spotify embed: HTTP %d", resp.StatusCode))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, errs.Wrap(errs.ErrNetwork, err)
	}
	col, err := ParseSpotifyEmbed(body)
	if err != nil {
		return nil, err
	}
	if col.Kind == "" {
		col.Kind = kind
	}
	if col.ID == "" {
		col.ID = id
	}
	return col, nil
}

var nextDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json"[^>]*>(.*?)</script>`)

// ParseSpotifyEmbed parses the __NEXT_DATA__ JSON embedded in Spotify's embed page.
func ParseSpotifyEmbed(html []byte) (*SpotifyCollection, error) {
	m := nextDataRe.FindSubmatch(html)
	if m == nil {
		return nil, errs.New(errs.ErrSpotifyMatch, "embed metadata not found")
	}
	var root map[string]any
	if err := json.Unmarshal(m[1], &root); err != nil {
		return nil, errs.New(errs.ErrSpotifyMatch, "embed metadata: "+err.Error())
	}
	entity := dig(root, "props", "pageProps", "state", "data", "entity")
	ent, ok := entity.(map[string]any)
	if !ok {
		return nil, errs.New(errs.ErrSpotifyMatch, "embed entity missing")
	}

	col := &SpotifyCollection{
		Kind:     str(ent["type"]),
		ID:       str(ent["id"]),
		Title:    firstNonEmpty(str(ent["name"]), str(ent["title"])),
		CoverURL: bestCover(ent),
		Release:  str(dig(ent, "releaseDate", "isoString")),
	}
	artists := artistNames(ent["artists"])
	col.Owner = firstNonEmpty(strings.Join(artists, ", "), str(ent["subtitle"]), str(dig(ent, "authors")))

	if col.Kind == "track" || (ent["trackList"] == nil && ent["duration"] != nil) {
		col.Kind = "track"
		col.Tracks = []SpotifyTrack{{
			ID:          col.ID,
			Title:       col.Title,
			Artists:     nonEmptyOr(artists, splitArtists(str(ent["subtitle"]))),
			DurationMs:  num(ent["duration"]),
			CoverURL:    col.CoverURL,
			ReleaseDate: col.Release,
			Explicit:    ent["isExplicit"] == true,
			TrackNumber: 1,
		}}
		return col, nil
	}

	list, _ := ent["trackList"].([]any)
	for i, it := range list {
		t, ok := it.(map[string]any)
		if !ok {
			continue
		}
		uri := str(t["uri"])
		tid := uri
		if idx := strings.LastIndex(uri, ":"); idx >= 0 {
			tid = uri[idx+1:]
		}
		track := SpotifyTrack{
			ID:          tid,
			Title:       str(t["title"]),
			Artists:     splitArtists(str(t["subtitle"])),
			DurationMs:  num(t["duration"]),
			CoverURL:    col.CoverURL,
			TrackNumber: i + 1,
			ReleaseDate: col.Release,
			Explicit:    t["isExplicit"] == true,
		}
		if col.Kind == "album" {
			track.Album = col.Title
		}
		if track.Title != "" {
			col.Tracks = append(col.Tracks, track)
		}
	}
	if len(col.Tracks) == 0 {
		return nil, errs.New(errs.ErrSpotifyMatch, "no tracks in embed metadata")
	}
	return col, nil
}

// MatchQuery builds the search text used to find a public source for a track.
func MatchQuery(t SpotifyTrack) string {
	artist := ""
	if len(t.Artists) > 0 {
		artist = t.Artists[0]
	}
	q := strings.TrimSpace(artist + " - " + t.Title)
	return strings.Trim(q, " -")
}

// MatchSources returns candidate yt-dlp inputs in order of preference:
// YouTube Music "songs" search first (official audio), then a plain YouTube search.
func MatchSources(t SpotifyTrack) []string {
	q := MatchQuery(t)
	return []string{
		"https://music.youtube.com/search?q=" + url.QueryEscape(q) + "#songs",
		"ytsearch1:" + q + " audio",
	}
}

func dig(v any, keys ...string) any {
	for _, k := range keys {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[k]
	}
	return v
}

func str(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func num(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	}
	return 0
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func nonEmptyOr(a, b []string) []string {
	if len(a) > 0 {
		return a
	}
	return b
}

func artistNames(v any) []string {
	arr, _ := v.([]any)
	var out []string
	for _, a := range arr {
		if m, ok := a.(map[string]any); ok {
			if n := str(m["name"]); n != "" {
				out = append(out, n)
			}
		}
	}
	return out
}

func splitArtists(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, " ", " ")
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func bestCover(ent map[string]any) string {
	var sources []any
	if s, ok := dig(ent, "coverArt", "sources").([]any); ok {
		sources = s
	} else if s, ok := dig(ent, "visualIdentity", "image").([]any); ok {
		sources = s
	}
	best, bestW := "", int64(-1)
	for _, s := range sources {
		m, ok := s.(map[string]any)
		if !ok {
			continue
		}
		w := num(m["width"])
		if w == 0 {
			w = num(m["maxWidth"])
		}
		if u := str(m["url"]); u != "" && w > bestW {
			best, bestW = u, w
		}
	}
	return best
}
