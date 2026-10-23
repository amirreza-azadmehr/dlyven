package extractor

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/amirreza-azadmehr/dlyven/internal/errs"
)

// rawInfo mirrors the subset of yt-dlp's info dict we use. Pointers distinguish
// "absent/null" from zero.
type rawInfo struct {
	Type               string                   `json:"_type"`
	ID                 string                   `json:"id"`
	Title              string                   `json:"title"`
	Description        string                   `json:"description"`
	Uploader           string                   `json:"uploader"`
	Channel            string                   `json:"channel"`
	Creator            string                   `json:"creator"`
	Artist             string                   `json:"artist"`
	Duration           float64                  `json:"duration"`
	UploadDate         string                   `json:"upload_date"`
	ViewCount          int64                    `json:"view_count"`
	LikeCount          int64                    `json:"like_count"`
	Thumbnail          string                   `json:"thumbnail"`
	Thumbnails         []rawThumb               `json:"thumbnails"`
	WebpageURL         string                   `json:"webpage_url"`
	OriginalURL        string                   `json:"original_url"`
	URL                string                   `json:"url"`
	IEKey              string                   `json:"ie_key"`
	Extractor          string                   `json:"extractor_key"`
	IsLive             bool                     `json:"is_live"`
	LiveStatus         string                   `json:"live_status"`
	Formats            []rawFormat              `json:"formats"`
	Subtitles          map[string][]rawSubtitle `json:"subtitles"`
	AutomaticCaptions  map[string][]rawSubtitle `json:"automatic_captions"`
	Entries            []json.RawMessage        `json:"entries"`
	PlaylistCount      int                      `json:"playlist_count"`
	Width              int                      `json:"width"`
	Height             int                      `json:"height"`
	Ext                string                   `json:"ext"`
	VCodec             *string                  `json:"vcodec"`
	ACodec             *string                  `json:"acodec"`
	FormatID           string                   `json:"format_id"`
	RequestedDownloads []json.RawMessage        `json:"requested_downloads"`
}

type rawThumb struct {
	URL        string  `json:"url"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	ID         string  `json:"id"`
	Preference float64 `json:"preference"`
}

type rawSubtitle struct {
	Ext  string `json:"ext"`
	URL  string `json:"url"`
	Name string `json:"name"`
}

type rawFormat struct {
	FormatID       string   `json:"format_id"`
	Ext            string   `json:"ext"`
	Width          *int     `json:"width"`
	Height         *int     `json:"height"`
	FPS            *float64 `json:"fps"`
	DynamicRange   *string  `json:"dynamic_range"`
	VCodec         *string  `json:"vcodec"`
	ACodec         *string  `json:"acodec"`
	TBR            *float64 `json:"tbr"`
	VBR            *float64 `json:"vbr"`
	ABR            *float64 `json:"abr"`
	ASR            *int     `json:"asr"`
	AudioChannels  *int     `json:"audio_channels"`
	Filesize       *float64 `json:"filesize"`
	FilesizeApprox *float64 `json:"filesize_approx"`
	FormatNote     string   `json:"format_note"`
	Language       *string  `json:"language"`
	Protocol       string   `json:"protocol"`
	Resolution     string   `json:"resolution"`
}

// ParseInfoJSON converts `yt-dlp -J` output into MediaInfo.
func ParseInfoJSON(data []byte) (*MediaInfo, error) {
	var raw rawInfo
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, errs.New(errs.ErrUnknown, "invalid yt-dlp JSON: "+err.Error())
	}
	if (raw.Type == "playlist" || raw.Type == "multi_video") && len(raw.Entries) == 1 {
		// Single-item "playlists" (e.g. one-image carousels) are shown as a single item.
		var only rawInfo
		if err := json.Unmarshal(raw.Entries[0], &only); err == nil && len(only.Formats) > 0 {
			if only.WebpageURL == "" {
				only.WebpageURL = raw.WebpageURL
			}
			raw = only
		}
	}

	mi := &MediaInfo{
		ID:          raw.ID,
		URL:         firstNonEmpty(raw.WebpageURL, raw.OriginalURL, raw.URL),
		Title:       strings.TrimSpace(raw.Title),
		Description: truncateRunes(raw.Description, 600),
		Uploader:    firstNonEmpty(raw.Uploader, raw.Channel, raw.Creator, raw.Artist),
		Duration:    raw.Duration,
		UploadDate:  formatDate(raw.UploadDate),
		ViewCount:   raw.ViewCount,
		LikeCount:   raw.LikeCount,
		IsLive:      raw.IsLive || raw.LiveStatus == "is_live",
		Extractor:   raw.Extractor,
		Thumbnails:  normalizeThumbs(raw.Thumbnails, raw.Thumbnail),
		Entries:     []Entry{},
		VideoGroups: []ResolutionGroup{},
		Audio:       []Format{},
		Subtitles:   []SubtitleTrack{},
	}
	mi.Thumbnail = raw.Thumbnail
	if mi.Thumbnail == "" && len(mi.Thumbnails) > 0 {
		mi.Thumbnail = mi.Thumbnails[0].URL
	}

	if raw.Type == "playlist" || raw.Type == "multi_video" {
		mi.IsPlaylist = true
		for i, e := range raw.Entries {
			var re rawInfo
			if err := json.Unmarshal(e, &re); err != nil {
				continue
			}
			entry := Entry{
				Index:     i + 1,
				ID:        re.ID,
				Title:     firstNonEmpty(strings.TrimSpace(re.Title), re.ID, fmt.Sprintf("#%d", i+1)),
				URL:       entryURL(re),
				Duration:  re.Duration,
				Thumbnail: re.Thumbnail,
				Uploader:  firstNonEmpty(re.Uploader, re.Channel),
			}
			if entry.Thumbnail == "" {
				if th := normalizeThumbs(re.Thumbnails, ""); len(th) > 0 {
					entry.Thumbnail = th[len(th)/2].URL // a mid-size thumbnail for list rows
				}
			}
			mi.Entries = append(mi.Entries, entry)
		}
		if mi.Thumbnail == "" && len(mi.Entries) > 0 {
			mi.Thumbnail = mi.Entries[0].Thumbnail
		}
		return mi, nil
	}

	formats := make([]Format, 0, len(raw.Formats))
	for _, rf := range raw.Formats {
		if f, ok := normalizeFormat(rf, raw.Duration); ok {
			formats = append(formats, f)
		}
	}
	if len(formats) == 0 && raw.FormatID != "" {
		// Single-format extractors put the format fields on the root object.
		f, ok := normalizeFormat(rawFormat{FormatID: raw.FormatID, Ext: raw.Ext, Width: &raw.Width, Height: &raw.Height, VCodec: raw.VCodec, ACodec: raw.ACodec}, raw.Duration)
		if ok {
			formats = append(formats, f)
		}
	}
	mi.VideoGroups, mi.Audio = GroupFormats(formats)
	mi.Highlights = PickHighlights(mi.VideoGroups)
	if len(mi.VideoGroups) > 0 {
		mi.MaxHeight = mi.VideoGroups[0].Height
	}
	mi.Subtitles = normalizeSubs(raw.Subtitles, raw.AutomaticCaptions)
	return mi, nil
}

func entryURL(re rawInfo) string {
	for _, u := range []string{re.URL, re.WebpageURL, re.OriginalURL} {
		if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
			return u
		}
	}
	if re.IEKey == "Youtube" && re.ID != "" {
		return "https://www.youtube.com/watch?v=" + re.ID
	}
	return ""
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func normalizeFormat(rf rawFormat, duration float64) (Format, bool) {
	vc, ac := deref(rf.VCodec), deref(rf.ACodec)
	if rf.Ext == "mhtml" || strings.Contains(strings.ToLower(rf.FormatNote), "storyboard") || rf.Protocol == "mhtml" {
		return Format{}, false
	}
	w, h := deref(rf.Width), deref(rf.Height)
	audioExt := map[string]bool{"m4a": true, "mp3": true, "opus": true, "ogg": true, "aac": true, "flac": true, "wav": true, "oga": true, "weba": true}
	imageExt := map[string]bool{"jpg": true, "jpeg": true, "png": true, "webp": true, "gif": true}
	still := imageExt[strings.ToLower(rf.Ext)] && (vc == "none" || vc == "") && (ac == "none" || ac == "")
	if vc == "none" && ac == "none" && !still {
		return Format{}, false
	}
	hasVideo := vc != "none" || still
	if !still && vc == "" && w == 0 && h == 0 && (audioExt[rf.Ext] || (ac != "" && ac != "none")) {
		hasVideo = false
	}
	hasAudio := ac != "none" && !still
	if !hasVideo && !hasAudio {
		return Format{}, false
	}
	f := Format{
		ID:       rf.FormatID,
		Ext:      rf.Ext,
		Width:    w,
		Height:   h,
		FPS:      math.Round(deref(rf.FPS)*100) / 100,
		DynRange: deref(rf.DynamicRange),
		TBR:      deref(rf.TBR),
		VBR:      deref(rf.VBR),
		ABR:      deref(rf.ABR),
		ASR:      deref(rf.ASR),
		Channels: deref(rf.AudioChannels),
		HasVideo: hasVideo,
		HasAudio: hasAudio,
		Note:     rf.FormatNote,
		Language: deref(rf.Language),
		Protocol: rf.Protocol,
	}
	f.HDR = f.DynRange != "" && f.DynRange != "SDR"
	if hasVideo {
		f.VCodec = NormalizeCodec(vc)
	}
	if hasAudio {
		f.ACodec = NormalizeCodec(ac)
	}
	switch {
	case rf.Filesize != nil && *rf.Filesize > 0:
		f.Filesize = int64(*rf.Filesize)
	case rf.FilesizeApprox != nil && *rf.FilesizeApprox > 0:
		f.Filesize, f.SizeApprox = int64(*rf.FilesizeApprox), true
	case f.TBR > 0 && duration > 0:
		f.Filesize, f.SizeApprox = int64(f.TBR*1000/8*duration), true
	}
	if hasVideo {
		f.Label = ResolutionLabel(w, h, f.FPS)
	}
	return f, true
}

// NormalizeCodec maps codec strings like "avc1.640028" to display names.
func NormalizeCodec(c string) string {
	l := strings.ToLower(c)
	switch {
	case l == "" || l == "none":
		return ""
	case strings.HasPrefix(l, "avc") || strings.HasPrefix(l, "h264"):
		return "H.264"
	case strings.HasPrefix(l, "hvc") || strings.HasPrefix(l, "hev") || strings.HasPrefix(l, "h265"):
		return "H.265"
	case strings.HasPrefix(l, "vp09") || strings.HasPrefix(l, "vp9"):
		return "VP9"
	case strings.HasPrefix(l, "vp8"):
		return "VP8"
	case strings.HasPrefix(l, "av01") || l == "av1":
		return "AV1"
	case strings.HasPrefix(l, "mp4a") || l == "aac":
		return "AAC"
	case strings.HasPrefix(l, "opus"):
		return "Opus"
	case strings.HasPrefix(l, "vorbis"):
		return "Vorbis"
	case strings.HasPrefix(l, "mp3"):
		return "MP3"
	case strings.HasPrefix(l, "flac"):
		return "FLAC"
	case strings.HasPrefix(l, "ec-3") || strings.HasPrefix(l, "eac3"):
		return "E-AC-3"
	case strings.HasPrefix(l, "ac-3") || strings.HasPrefix(l, "ac3"):
		return "AC-3"
	}
	if i := strings.IndexByte(c, '.'); i > 0 {
		return strings.ToUpper(c[:i])
	}
	return strings.ToUpper(c)
}

// ResolutionHeight returns the "p" value: the short side, so vertical 1080×1920 is 1080p.
func ResolutionHeight(w, h int) int {
	if w > 0 && h > 0 && w < h {
		return w
	}
	return h
}

// ResolutionLabel renders "1080p", "1080p60", "2160p" …
func ResolutionLabel(w, h int, fps float64) string {
	p := ResolutionHeight(w, h)
	if p <= 0 {
		return ""
	}
	label := fmt.Sprintf("%dp", p)
	if fps > 30.5 {
		label += fmt.Sprintf("%d", int(math.Round(fps)))
	}
	return label
}

// GroupFormats splits formats into resolution groups (highest first) and audio-only formats.
func GroupFormats(formats []Format) ([]ResolutionGroup, []Format) {
	groups := map[int][]Format{}
	audio := []Format{}
	unknown := []Format{}
	for _, f := range formats {
		switch {
		case f.HasVideo && ResolutionHeight(f.Width, f.Height) > 0:
			p := ResolutionHeight(f.Width, f.Height)
			groups[p] = append(groups[p], f)
		case f.HasVideo:
			unknown = append(unknown, f)
		case f.HasAudio:
			audio = append(audio, f)
		}
	}
	heights := make([]int, 0, len(groups))
	for h := range groups {
		heights = append(heights, h)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(heights)))
	out := make([]ResolutionGroup, 0, len(heights)+1)
	for _, h := range heights {
		fs := groups[h]
		sort.SliceStable(fs, func(i, j int) bool { return betterVideo(fs[i], fs[j]) })
		out = append(out, ResolutionGroup{Height: h, Label: fmt.Sprintf("%dp", h), Formats: fs})
	}
	if len(unknown) > 0 {
		sort.SliceStable(unknown, func(i, j int) bool { return betterVideo(unknown[i], unknown[j]) })
		out = append(out, ResolutionGroup{Height: 0, Label: "", Formats: unknown})
	}
	sort.SliceStable(audio, func(i, j int) bool {
		ai, aj := bitrate(audio[i]), bitrate(audio[j])
		if ai != aj {
			return ai > aj
		}
		return audio[i].Filesize > audio[j].Filesize
	})
	return out, audio
}

func bitrate(f Format) float64 {
	if f.ABR > 0 {
		return f.ABR
	}
	return f.TBR
}

var codecRank = map[string]int{"AV1": 4, "VP9": 3, "H.265": 2, "H.264": 1}

func betterVideo(a, b Format) bool {
	if a.FPS != b.FPS {
		return a.FPS > b.FPS
	}
	if a.HDR != b.HDR {
		return a.HDR
	}
	if codecRank[a.VCodec] != codecRank[b.VCodec] {
		return codecRank[a.VCodec] > codecRank[b.VCodec]
	}
	if a.TBR != b.TBR {
		return a.TBR > b.TBR
	}
	return a.Filesize > b.Filesize
}

// PickHighlights chooses the "Best", "Recommended" (most compatible: H.264 ≤1080p) and
// "Smallest" video formats.
func PickHighlights(groups []ResolutionGroup) Highlights {
	var h Highlights
	if len(groups) == 0 {
		return h
	}
	h.Best = groups[0].Formats[0].ID

	var rec *Format
	for gi := range groups {
		if groups[gi].Height > 1080 {
			continue
		}
		for fi := range groups[gi].Formats {
			f := &groups[gi].Formats[fi]
			if f.VCodec == "H.264" && !f.HDR {
				if rec == nil || f.FPS > rec.FPS {
					rec = f
				}
			}
		}
		if rec != nil {
			break
		}
	}
	if rec == nil {
		for gi := range groups {
			if groups[gi].Height <= 1080 && len(groups[gi].Formats) > 0 {
				rec = &groups[gi].Formats[0]
				break
			}
		}
	}
	if rec != nil {
		h.Recommended = rec.ID
	} else {
		h.Recommended = h.Best
	}

	var small *Format
	for gi := range groups {
		for fi := range groups[gi].Formats {
			f := &groups[gi].Formats[fi]
			if f.Filesize > 0 && (small == nil || f.Filesize < small.Filesize) {
				small = f
			}
		}
	}
	if small == nil {
		last := groups[len(groups)-1]
		small = &last.Formats[len(last.Formats)-1]
	}
	h.Smallest = small.ID
	return h
}

func normalizeSubs(manual, auto map[string][]rawSubtitle) []SubtitleTrack {
	out := []SubtitleTrack{}
	add := func(m map[string][]rawSubtitle, isAuto bool) {
		langs := make([]string, 0, len(m))
		for l := range m {
			if l == "live_chat" || l == "rechat" || l == "danmaku" {
				continue
			}
			langs = append(langs, l)
		}
		sort.Strings(langs)
		for _, l := range langs {
			subs := m[l]
			t := SubtitleTrack{Lang: l, Auto: isAuto}
			seen := map[string]bool{}
			for _, s := range subs {
				if t.Name == "" && s.Name != "" {
					t.Name = s.Name
				}
				if s.Ext != "" && !seen[s.Ext] {
					seen[s.Ext] = true
					t.Exts = append(t.Exts, s.Ext)
				}
			}
			out = append(out, t)
		}
	}
	add(manual, false)
	add(auto, true)
	return out
}

func normalizeThumbs(ts []rawThumb, main string) []Thumbnail {
	seen := map[string]bool{}
	out := []Thumbnail{}
	for _, t := range ts {
		if t.URL == "" || seen[t.URL] {
			continue
		}
		seen[t.URL] = true
		out = append(out, Thumbnail{URL: t.URL, Width: t.Width, Height: t.Height, ID: t.ID})
	}
	if main != "" && !seen[main] {
		out = append(out, Thumbnail{URL: main})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Width > out[j].Width })
	if len(out) > 12 {
		// keep the largest ones plus the main thumbnail
		out = out[:12]
	}
	return out
}

func formatDate(d string) string {
	if len(d) == 8 {
		return d[:4] + "-" + d[4:6] + "-" + d[6:]
	}
	return d
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
