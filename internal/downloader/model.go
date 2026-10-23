// Package downloader runs a concurrency-limited queue of yt-dlp downloads with
// pause/resume/cancel, retries with backoff, and throttled progress events.
package downloader

import "github.com/amirreza-azadmehr/dlyven/internal/platform"

type Kind string

const (
	KindVideo     Kind = "video"
	KindAudio     Kind = "audio"
	KindSubtitles Kind = "subtitles"
	KindThumbnail Kind = "thumbnail"
)

type Status string

const (
	StatusQueued      Status = "queued"
	StatusDownloading Status = "downloading"
	StatusMerging     Status = "merging"
	StatusConverting  Status = "converting"
	StatusPaused      Status = "paused"
	StatusRetrying    Status = "retrying"
	StatusDone        Status = "done"
	StatusFailed      Status = "failed"
	StatusCanceled    Status = "canceled"
)

func (s Status) Active() bool {
	return s == StatusDownloading || s == StatusMerging || s == StatusConverting || s == StatusRetrying
}

func (s Status) Final() bool {
	return s == StatusDone || s == StatusFailed || s == StatusCanceled
}

// Request is everything needed to perform one download. Built by the frontend.
type Request struct {
	URL        string  `json:"url"`
	Title      string  `json:"title"`
	Uploader   string  `json:"uploader"`
	Thumbnail  string  `json:"thumbnail"`
	Platform   string  `json:"platform"`
	Duration   float64 `json:"duration"`
	UploadDate string  `json:"uploadDate"`
	MediaID    string  `json:"mediaId"`

	// Playlist context: when the entry has no own URL, PlaylistItem selects it from URL.
	PlaylistTitle string `json:"playlistTitle"`
	PlaylistIndex int    `json:"playlistIndex"`
	PlaylistItem  int    `json:"playlistItem"`

	Kind Kind `json:"kind"`

	// Video
	FormatID  string  `json:"formatId"`  // exact format, single items only
	HasAudio  bool    `json:"hasAudio"`  // whether FormatID already contains audio
	MaxHeight int     `json:"maxHeight"` // preset cap (0 = best); used when FormatID == ""
	Height    int     `json:"height"`
	FPS       float64 `json:"fps"`
	Container string  `json:"container"` // mp4 | mkv | webm

	// Audio
	AudioFormat   string `json:"audioFormat"`   // mp3 | m4a | opus | flac | wav
	AudioBitrate  string `json:"audioBitrate"`  // best | 320 | 256 | …
	AudioFormatID string `json:"audioFormatId"` // optional exact audio stream

	// Subtitles (with Kind=subtitles they are the only output; otherwise extra)
	SubLangs  []string `json:"subLangs"`
	SubMode   string   `json:"subMode"`   // embed | file
	SubFormat string   `json:"subFormat"` // srt | vtt
	AutoSubs  bool     `json:"autoSubs"`

	// Thumbnail
	ThumbnailURL    string `json:"thumbnailUrl"`
	ThumbnailFormat string `json:"thumbnailFormat"` // jpg | png | original

	// Embedding (default from settings)
	EmbedThumbnail bool `json:"embedThumbnail"`
	EmbedMetadata  bool `json:"embedMetadata"`
	EmbedChapters  bool `json:"embedChapters"`

	OutputDir        string `json:"outputDir"`
	FilenameTemplate string `json:"filenameTemplate"`
	QualityLabel     string `json:"qualityLabel"`

	// Spotify: audio comes from a matched public source and is tagged with this metadata.
	Spotify *platform.SpotifyTrack `json:"spotify,omitempty"`
	Album   string                 `json:"album"`
}

type Job struct {
	ID          string  `json:"id"`
	Request     Request `json:"request"`
	Status      Status  `json:"status"`
	Progress    float64 `json:"progress"` // 0..100 overall
	Downloaded  int64   `json:"downloaded"`
	Total       int64   `json:"total"`
	Speed       float64 `json:"speed"` // bytes/s
	ETA         int64   `json:"eta"`   // seconds, -1 unknown
	Part        int     `json:"part"`  // current stream (1-based) when merging video+audio
	Parts       int     `json:"parts"`
	FilePath    string  `json:"filePath"`
	OutputBase  string  `json:"outputBase"` // planned file name without extension
	Ext         string  `json:"ext"`
	ErrorCode   string  `json:"errorCode"`
	ErrorDetail string  `json:"errorDetail"`
	Attempts    int     `json:"attempts"`
	CreatedAt   int64   `json:"createdAt"`
	StartedAt   int64   `json:"startedAt"`
	FinishedAt  int64   `json:"finishedAt"`
}
