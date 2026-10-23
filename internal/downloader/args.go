package downloader

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/amirreza-azadmehr/dlyven/internal/filename"
	"github.com/amirreza-azadmehr/dlyven/internal/settings"
)

// buildArgs returns the yt-dlp arguments (after the shared base args) for a request,
// plus the expected final extension. input is the URL or search query to download.
func buildArgs(req Request, s settings.Settings, input, outDir, tmpDir, outBase string) ([]string, string) {
	args := []string{
		"--newline", "--progress",
		"--progress-template", DownloadTemplate,
		"--progress-template", PostprocessTemplate,
		"--no-mtime",
		"--continue",
		"-P", "home:" + outDir,
		"-P", "temp:" + tmpDir,
		"-o", filename.EscapeOutputTemplate(outBase) + ".%(ext)s",
		"--print-to-file", "after_move:filepath", filename.EscapeOutputTemplate(filepath.Join(tmpDir, resultFile)),
	}
	if s.SpeedLimitKBps > 0 {
		args = append(args, "--limit-rate", fmt.Sprintf("%dK", s.SpeedLimitKBps))
	}

	switch {
	case req.Spotify != nil:
		// Search inputs are playlists of results; take the top match only.
		args = append(args, "--yes-playlist", "--playlist-items", "1")
	case req.PlaylistItem > 0:
		args = append(args, "--yes-playlist", "--playlist-items", fmt.Sprint(req.PlaylistItem))
	default:
		args = append(args, "--no-playlist")
	}

	var ext string
	switch req.Kind {
	case KindAudio:
		args, ext = audioArgs(args, req)
	case KindSubtitles:
		args, ext = subtitleOnlyArgs(args, req)
	default:
		args, ext = videoArgs(args, req)
	}
	args = append(args, "--", input)
	return args, ext
}

const resultFile = "result.txt"

func containerOf(req Request) string {
	switch req.Container {
	case "mkv", "webm", "mp4":
		return req.Container
	}
	return "mp4"
}

// VideoFormatSelector returns -f and -S values for a video request.
func VideoFormatSelector(req Request) (format string, sort string) {
	c := containerOf(req)
	audio := "ba"
	switch c {
	case "mp4":
		audio = "ba[ext=m4a]"
	case "webm":
		audio = "ba[ext=webm]"
	}
	if req.FormatID != "" {
		if req.HasAudio {
			return req.FormatID, ""
		}
		return fmt.Sprintf("%s+%s/%s+ba/%s", req.FormatID, audio, req.FormatID, req.FormatID), ""
	}
	res := "res"
	if req.MaxHeight > 0 {
		res = fmt.Sprintf("res:%d", req.MaxHeight)
	}
	switch c {
	case "mp4":
		// Prefer H.264 at the chosen resolution for maximum player compatibility;
		// resolution ranks first, so 4K still falls back to VP9/AV1.
		return "bv*+ba/b", res + ",fps,vcodec:h264,ext:mp4:m4a"
	case "webm":
		return "bv*[ext=webm]+ba[ext=webm]/bv*+ba/b", res + ",fps,ext:webm:webm"
	}
	return "bv*+ba/b", res + ",fps"
}

func videoArgs(args []string, req Request) ([]string, string) {
	c := containerOf(req)
	f, sortSpec := VideoFormatSelector(req)
	args = append(args, "-f", f)
	if sortSpec != "" {
		args = append(args, "-S", sortSpec)
	}
	args = append(args, "--merge-output-format", c)
	if c == "webm" {
		// webm only holds VP8/VP9/AV1 + Opus/Vorbis; anything else falls back to mkv.
		args = append(args, "--remux-video", "webm>webm/mkv")
	} else {
		args = append(args, "--remux-video", c)
	}
	if req.EmbedMetadata {
		args = append(args, "--embed-metadata")
	}
	if req.EmbedThumbnail && c != "webm" {
		args = append(args, "--embed-thumbnail", "--convert-thumbnails", "jpg")
	}
	if req.EmbedChapters {
		args = append(args, "--embed-chapters")
	}
	if len(req.SubLangs) > 0 {
		args = append(args, "--write-subs", "--sub-langs", strings.Join(req.SubLangs, ","))
		if req.AutoSubs {
			args = append(args, "--write-auto-subs")
		}
		if req.SubMode == "embed" {
			args = append(args, "--embed-subs")
		} else {
			args = append(args, "--convert-subs", subFormat(req))
		}
	}
	return args, c
}

func subFormat(req Request) string {
	if req.SubFormat == "vtt" {
		return "vtt"
	}
	return "srt"
}

// AudioQualityArg maps a bitrate choice to yt-dlp's --audio-quality.
func AudioQualityArg(bitrate string) string {
	if bitrate == "" || bitrate == "best" {
		return "0"
	}
	return bitrate + "K"
}

func audioFormatOf(req Request) string {
	switch req.AudioFormat {
	case "mp3", "m4a", "opus", "flac", "wav":
		return req.AudioFormat
	}
	return "mp3"
}

func audioArgs(args []string, req Request) ([]string, string) {
	af := audioFormatOf(req)
	sel := "ba/b"
	switch {
	case req.AudioFormatID != "":
		sel = req.AudioFormatID + "/ba/b"
	case af == "m4a":
		sel = "ba[ext=m4a]/ba/b"
	case af == "opus":
		sel = "ba[acodec=opus]/ba/b"
	}
	args = append(args, "-f", sel, "-x", "--audio-format", af)
	if af != "flac" && af != "wav" {
		args = append(args, "--audio-quality", AudioQualityArg(req.AudioBitrate))
	}
	if req.Spotify != nil {
		// Tags and cover come from Spotify metadata afterwards (see tagSpotify).
		return args, af
	}
	if req.EmbedMetadata {
		args = append(args, "--embed-metadata")
	}
	if req.EmbedThumbnail && af != "wav" {
		args = append(args, "--embed-thumbnail", "--convert-thumbnails", "jpg")
	}
	return args, af
}

func subtitleOnlyArgs(args []string, req Request) ([]string, string) {
	sf := subFormat(req)
	args = append(args, "--skip-download", "--write-subs", "--sub-langs", strings.Join(req.SubLangs, ","),
		"--sub-format", sf+"/best", "--convert-subs", sf)
	if req.AutoSubs {
		args = append(args, "--write-auto-subs")
	}
	return args, sf
}
