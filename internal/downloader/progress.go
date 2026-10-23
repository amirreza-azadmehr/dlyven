package downloader

import (
	"bytes"
	"strconv"
	"strings"
)

// yt-dlp is asked to print machine-readable progress lines via --progress-template:
//
//	DLYVEN|<status>|<downloaded>|<total>|<total_estimate>|<speed>|<eta>|<format_id>|<2nd format_id>|<frag_i>|<frag_n>
//	DLYVENPP|<status>|<postprocessor>
//
// Missing values are printed as "NA". yt-dlp often uses CR (not LF) between progress
// updates, even with --newline, so the scanner must split on both.
const (
	progressPrefix = "DLYVEN|"
	ppPrefix       = "DLYVENPP|"

	DownloadTemplate    = "download:" + progressPrefix + "%(progress.status)s|%(progress.downloaded_bytes)s|%(progress.total_bytes)s|%(progress.total_bytes_estimate)s|%(progress.speed)s|%(progress.eta)s|%(info.format_id)s|%(info.requested_formats.1.format_id)s|%(progress.fragment_index)s|%(progress.fragment_count)s"
	PostprocessTemplate = "postprocess:" + ppPrefix + "%(progress.status)s|%(progress.postprocessor)s"

	// Files smaller than this (thumbnails, storyboards, sidecar JSON) must not drive the bar.
	auxMaxBytes = 1 << 20
)

type Progress struct {
	Status     string
	Downloaded int64
	Total      int64
	Speed      float64
	ETA        int64
	FormatID   string
	// SecondFormatID is the second stream of a video+audio merge ("" when not merging).
	SecondFormatID string
	Fragment       int
	Fragments      int
}

func parseNum(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "NA" || s == "None" {
		return -1
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1
	}
	return f
}

// splitProgressScan splits yt-dlp output on CR, LF, or CRLF so live `\r` progress is not
// concatenated into one record whose first DLYVEN| line (often a finished thumbnail) wins.
func splitProgressScan(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		sep := 1
		if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
			sep = 2
		}
		return i + sep, data[:i], nil
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// ParseProgressLine parses a download progress line. If several records were glued
// together, the last one is used (the most recent `\r` update).
func ParseProgressLine(line string) (Progress, bool) {
	i := strings.LastIndex(line, progressPrefix)
	if i < 0 {
		return Progress{}, false
	}
	rest := strings.TrimSpace(line[i+len(progressPrefix):])
	if j := strings.IndexByte(rest, '\r'); j >= 0 {
		rest = rest[:j]
	}
	parts := strings.Split(rest, "|")
	if len(parts) < 7 {
		return Progress{}, false
	}
	p := Progress{Status: parts[0], FormatID: parts[6], ETA: -1}
	if d := parseNum(parts[1]); d >= 0 {
		p.Downloaded = int64(d)
	}
	if t := parseNum(parts[2]); t > 0 {
		p.Total = int64(t)
	} else if t := parseNum(parts[3]); t > 0 {
		p.Total = int64(t)
	}
	if s := parseNum(parts[4]); s > 0 {
		p.Speed = s
	}
	if e := parseNum(parts[5]); e >= 0 {
		p.ETA = int64(e)
	}
	if p.FormatID == "NA" {
		p.FormatID = ""
	}
	if len(parts) >= 8 && parts[7] != "NA" {
		p.SecondFormatID = parts[7]
	}
	if len(parts) >= 10 {
		if n := parseNum(parts[8]); n >= 0 {
			p.Fragment = int(n)
		}
		if n := parseNum(parts[9]); n > 0 {
			p.Fragments = int(n)
		}
	}
	return p, true
}

func (p Progress) auxiliary() bool {
	id := strings.ToLower(p.FormatID)
	if strings.HasPrefix(id, "sb") {
		return true
	}
	return p.Total > 0 && p.Total < auxMaxBytes
}

func (p Progress) fraction() float64 {
	if p.Status == "finished" && (p.Total > 0 || p.Fragments > 0) {
		return 1
	}
	if p.Total > 0 {
		return mathMin(float64(p.Downloaded)/float64(p.Total), 1)
	}
	if p.Fragments > 0 && p.Fragment >= 0 {
		return mathMin(float64(p.Fragment)/float64(p.Fragments), 1)
	}
	return 0
}

func mathMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ParsePostprocessLine returns the job status implied by a postprocessor line.
func ParsePostprocessLine(line string) (Status, bool) {
	i := strings.Index(line, ppPrefix)
	if i >= 0 {
		parts := strings.Split(strings.TrimSpace(line[i+len(ppPrefix):]), "|")
		if len(parts) < 2 || parts[0] == "finished" {
			return "", false
		}
		return statusForPostprocessor(parts[1])
	}
	// Fallback on yt-dlp's human-readable lines, e.g. "[Merger] Merging formats into ..."
	trim := strings.TrimSpace(line)
	if strings.HasPrefix(trim, "[") {
		if j := strings.Index(trim, "]"); j > 1 {
			return statusForPostprocessor(trim[1:j])
		}
	}
	return "", false
}

func statusForPostprocessor(name string) (Status, bool) {
	switch {
	case name == "Merger" || name == "FFmpegMerger":
		return StatusMerging, true
	case strings.HasPrefix(name, "FFmpeg") || name == "EmbedThumbnail" || name == "ExtractAudio" ||
		name == "VideoConvertor" || name == "VideoRemuxer" || name == "Metadata" || name == "ThumbnailsConvertor" ||
		name == "SubtitlesConvertor" || name == "EmbedSubtitle" || name == "FixupM3u8" || strings.HasPrefix(name, "Fixup"):
		return StatusConverting, true
	}
	return "", false
}
