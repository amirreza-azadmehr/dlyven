// Package platform detects which service a URL belongs to. It is the single source of truth
// for platform metadata: the frontend fetches the list via App.GetPlatforms() for live badges.
package platform

import (
	"net/url"
	"regexp"
	"strings"
)

// Def describes a known platform. To add a platform, append an entry to Registry.
// Names are brand names and are intentionally not translated.
type Def struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Color string   `json:"color"` // brand-tinted accent used for the badge (hex)
	Icon  string   `json:"icon"`  // generic lucide icon name (no brand logos) used by the frontend
	Hosts []string `json:"hosts"` // matched as exact host or subdomain suffix
	// AudioOnly platforms hide the video/subtitle tabs.
	AudioOnly bool `json:"audioOnly"`
	// Matched platforms are not downloaded directly; tracks are matched to a public source.
	Matched bool `json:"matched"`
}

// Info is the result of Detect.
type Info struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	Icon      string `json:"icon"`
	Valid     bool   `json:"valid"`
	Known     bool   `json:"known"`
	AudioOnly bool   `json:"audioOnly"`
	Matched   bool   `json:"matched"`
	URL       string `json:"url"` // normalized URL
}

// Registry lists the platforms with first-class detection. Anything else that yt-dlp
// supports still works and is shown as "generic" with its host name.
var Registry = []Def{
	{ID: "youtube", Name: "YouTube", Color: "#FF4E45", Icon: "square-play", Hosts: []string{"youtube.com", "youtu.be", "youtube-nocookie.com", "music.youtube.com"}},
	{ID: "instagram", Name: "Instagram", Color: "#E1306C", Icon: "camera", Hosts: []string{"instagram.com", "instagr.am", "ddinstagram.com"}},
	{ID: "spotify", Name: "Spotify", Color: "#1DB954", Icon: "music", Hosts: []string{"open.spotify.com", "spotify.com", "spotify.link"}, AudioOnly: true, Matched: true},
	{ID: "tiktok", Name: "TikTok", Color: "#25F4EE", Icon: "music-2", Hosts: []string{"tiktok.com", "vm.tiktok.com", "vt.tiktok.com"}},
	{ID: "twitter", Name: "X / Twitter", Color: "#8899A6", Icon: "at-sign", Hosts: []string{"twitter.com", "x.com", "mobile.twitter.com", "fxtwitter.com", "vxtwitter.com"}},
	{ID: "facebook", Name: "Facebook", Color: "#1877F2", Icon: "users", Hosts: []string{"facebook.com", "fb.watch", "fb.com", "m.facebook.com"}},
	{ID: "vimeo", Name: "Vimeo", Color: "#1AB7EA", Icon: "video", Hosts: []string{"vimeo.com", "player.vimeo.com"}},
	{ID: "soundcloud", Name: "SoundCloud", Color: "#FF5500", Icon: "cloud", Hosts: []string{"soundcloud.com", "on.soundcloud.com", "snd.sc"}, AudioOnly: true},
	{ID: "reddit", Name: "Reddit", Color: "#FF4500", Icon: "message-circle", Hosts: []string{"reddit.com", "redd.it", "v.redd.it", "old.reddit.com"}},
	{ID: "twitch", Name: "Twitch", Color: "#9146FF", Icon: "gamepad-2", Hosts: []string{"twitch.tv", "clips.twitch.tv", "m.twitch.tv"}},
	{ID: "dailymotion", Name: "Dailymotion", Color: "#0A84FF", Icon: "play", Hosts: []string{"dailymotion.com", "dai.ly"}},
	{ID: "bandcamp", Name: "Bandcamp", Color: "#1DA0C3", Icon: "disc-3", Hosts: []string{"bandcamp.com"}, AudioOnly: true},
	{ID: "threads", Name: "Threads", Color: "#A0A0A0", Icon: "at-sign", Hosts: []string{"threads.net", "threads.com"}},
	{ID: "pinterest", Name: "Pinterest", Color: "#E60023", Icon: "pin", Hosts: []string{"pinterest.com", "pin.it"}},
	{ID: "bilibili", Name: "Bilibili", Color: "#00A1D6", Icon: "tv", Hosts: []string{"bilibili.com", "b23.tv"}},
	{ID: "aparat", Name: "Aparat", Color: "#ED145B", Icon: "film", Hosts: []string{"aparat.com"}},
}

var bareHostRe = regexp.MustCompile(`^(?i)([a-z0-9-]+\.)+[a-z]{2,}(/|$)`)

// Normalize trims whitespace/wrapping characters, adds a scheme to bare hosts, and validates.
func Normalize(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, "<>\"'`()[]{}‎‏‪‫‬")
	if s == "" {
		return "", false
	}
	if strings.HasPrefix(strings.ToLower(s), "spotify:") {
		return s, true
	}
	if !strings.Contains(s, "://") {
		if !bareHostRe.MatchString(s) {
			return "", false
		}
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || !strings.Contains(u.Host, ".") {
		return "", false
	}
	return u.String(), true
}

func hostOf(u string) string {
	p, err := url.Parse(u)
	if err != nil {
		return ""
	}
	h := strings.ToLower(p.Hostname())
	return strings.TrimPrefix(h, "www.")
}

func matchHost(host, pattern string) bool {
	return host == pattern || strings.HasSuffix(host, "."+pattern)
}

// Lookup returns a platform definition by ID.
func Lookup(id string) (Def, bool) {
	for _, d := range Registry {
		if d.ID == id {
			return d, true
		}
	}
	return Def{}, false
}

// Detect classifies a URL.
func Detect(raw string) Info {
	norm, ok := Normalize(raw)
	if !ok {
		return Info{}
	}
	if strings.HasPrefix(strings.ToLower(norm), "spotify:") {
		d, _ := Lookup("spotify")
		return fromDef(d, norm)
	}
	host := hostOf(norm)
	for _, d := range Registry {
		for _, h := range d.Hosts {
			if matchHost(host, h) {
				return fromDef(d, norm)
			}
		}
	}
	return Info{ID: "generic", Name: host, Color: "#8B93A7", Icon: "globe", Valid: true, URL: norm}
}

func fromDef(d Def, norm string) Info {
	return Info{ID: d.ID, Name: d.Name, Color: d.Color, Icon: d.Icon, Valid: true, Known: true, AudioOnly: d.AudioOnly, Matched: d.Matched, URL: norm}
}

// instagramItemRe matches a single post/reel/TV/story (including /share/p/… wrappers),
// not a profile URL. Those posts are often carousels, so analysis must not pass --no-playlist.
var instagramItemRe = regexp.MustCompile(`(?i)/(?:share/)?(?:p|reel|reels|tv|stories)/`)

// ExpandCollection reports whether analysis should load every item in a collection
// (Instagram carousel, TikTok photo slideshow). Watch URLs with an attached playlist
// still use --no-playlist so YouTube "Up next" lists are ignored.
func ExpandCollection(pi Info) bool {
	switch pi.ID {
	case "instagram":
		u, err := url.Parse(pi.URL)
		if err != nil {
			return false
		}
		return instagramItemRe.MatchString(u.Path)
	case "tiktok":
		u, err := url.Parse(pi.URL)
		if err != nil {
			return false
		}
		p := strings.ToLower(u.Path)
		return strings.Contains(p, "/photo/") || strings.Contains(p, "/p/")
	}
	return false
}

var urlInTextRe = regexp.MustCompile(`(?i)(https?://[^\s<>"'` + "`" + `]+|spotify:[a-z]+:[A-Za-z0-9]+)`)

// ExtractURLs finds every URL in free text (e.g. a dropped .txt file), de-duplicated, in order.
func ExtractURLs(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range urlInTextRe.FindAllString(text, -1) {
		m = strings.TrimRight(m, ".,;:!?)]}»،")
		if n, ok := Normalize(m); ok && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}
