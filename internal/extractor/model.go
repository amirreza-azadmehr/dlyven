// Package extractor turns a URL into MediaInfo by running `yt-dlp -J` (or by reading
// Spotify's public metadata) and normalizing formats for the UI.
package extractor

import "github.com/amirreza-azadmehr/dlyven/internal/platform"

type Format struct {
	ID         string  `json:"id"`
	Ext        string  `json:"ext"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	FPS        float64 `json:"fps"`
	HDR        bool    `json:"hdr"`
	DynRange   string  `json:"dynamicRange"`
	VCodec     string  `json:"vcodec"` // normalized: avc1 → h264, etc. "" when audio-only
	ACodec     string  `json:"acodec"` // "" when video-only
	TBR        float64 `json:"tbr"`    // kbit/s
	VBR        float64 `json:"vbr"`
	ABR        float64 `json:"abr"`
	ASR        int     `json:"asr"`
	Channels   int     `json:"channels"`
	Filesize   int64   `json:"filesize"`
	SizeApprox bool    `json:"sizeApprox"`
	HasVideo   bool    `json:"hasVideo"`
	HasAudio   bool    `json:"hasAudio"`
	Note       string  `json:"note"`
	Language   string  `json:"language"`
	Protocol   string  `json:"protocol"`
	Label      string  `json:"label"` // "1080p", "1080p60", "4K"
}

// ResolutionGroup groups video formats sharing a height, best first.
type ResolutionGroup struct {
	Height  int      `json:"height"`
	Label   string   `json:"label"`
	Formats []Format `json:"formats"`
}

type SubtitleTrack struct {
	Lang string   `json:"lang"`
	Name string   `json:"name"`
	Auto bool     `json:"auto"`
	Exts []string `json:"exts"`
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	ID     string `json:"id"`
}

type Entry struct {
	Index     int     `json:"index"` // 1-based playlist position
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	URL       string  `json:"url"`
	Duration  float64 `json:"duration"`
	Thumbnail string  `json:"thumbnail"`
	Uploader  string  `json:"uploader"`
	// Spotify-only: the track metadata used for matching/tagging.
	Spotify *platform.SpotifyTrack `json:"spotify,omitempty"`
}

// Highlights are format IDs the UI badges.
type Highlights struct {
	Best        string `json:"best"`
	Recommended string `json:"recommended"`
	Smallest    string `json:"smallest"`
}

type MediaInfo struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"` // the URL to download (webpage_url)
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Uploader    string            `json:"uploader"`
	Duration    float64           `json:"duration"`
	UploadDate  string            `json:"uploadDate"` // YYYY-MM-DD
	ViewCount   int64             `json:"viewCount"`
	LikeCount   int64             `json:"likeCount"`
	Thumbnail   string            `json:"thumbnail"`
	IsLive      bool              `json:"isLive"`
	Platform    platform.Info     `json:"platform"`
	Extractor   string            `json:"extractor"`
	IsPlaylist  bool              `json:"isPlaylist"`
	Entries     []Entry           `json:"entries"`
	VideoGroups []ResolutionGroup `json:"videoGroups"`
	Audio       []Format          `json:"audio"`
	Highlights  Highlights        `json:"highlights"`
	MaxHeight   int               `json:"maxHeight"`
	Subtitles   []SubtitleTrack   `json:"subtitles"`
	Thumbnails  []Thumbnail       `json:"thumbnails"`
	// Matched is true for Spotify: audio comes from a matched public source.
	Matched bool                        `json:"matched"`
	Spotify *platform.SpotifyCollection `json:"spotify,omitempty"`
}
