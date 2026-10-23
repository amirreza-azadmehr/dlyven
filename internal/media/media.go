// Package media runs ffmpeg for post-processing Dlyven does itself (yt-dlp handles the rest):
// writing Spotify metadata + cover art onto matched audio, and converting thumbnails.
package media

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/amirreza-azadmehr/dlyven/internal/errs"
	"github.com/amirreza-azadmehr/dlyven/internal/sysutil"
)

// Tags are written as container metadata (ID3 for mp3, iTunes atoms for m4a, Vorbis comments…).
type Tags map[string]string

// coverSupported lists containers where ffmpeg can attach cover art as a picture stream.
var coverSupported = map[string]bool{".mp3": true, ".m4a": true, ".flac": true, ".mp4": true}

// TagAudio rewrites path in place with the given tags and optional cover image.
func TagAudio(ctx context.Context, ffmpeg, path string, tags Tags, cover string) error {
	if ffmpeg == "" {
		return errs.New(errs.ErrFFmpegMissing, "")
	}
	ext := strings.ToLower(filepath.Ext(path))
	tmp := strings.TrimSuffix(path, filepath.Ext(path)) + ".tagging" + ext

	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", path}
	withCover := cover != "" && coverSupported[ext]
	if withCover {
		args = append(args, "-i", cover, "-map", "0:a", "-map", "1:v", "-c:v", "mjpeg", "-disposition:v:0", "attached_pic",
			"-metadata:s:v", "title=Album cover", "-metadata:s:v", "comment=Cover (front)")
	} else {
		args = append(args, "-map", "0:a")
	}
	args = append(args, "-c:a", "copy", "-map_metadata", "-1")
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if v := strings.TrimSpace(tags[k]); v != "" {
			args = append(args, "-metadata", k+"="+v)
		}
	}
	if ext == ".mp3" {
		args = append(args, "-id3v2_version", "3")
	}
	args = append(args, tmp)

	var stderr bytes.Buffer
	cmd := sysutil.CommandContext(ctx, ffmpeg, args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(tmp)
		return errs.New(errs.Classify(stderr.String()), fmt.Sprintf("ffmpeg tag: %v: %s", err, strings.TrimSpace(stderr.String())))
	}
	return os.Rename(tmp, path)
}

// ConvertImage converts src to dst (format from dst's extension: .jpg or .png).
func ConvertImage(ctx context.Context, ffmpeg, src, dst string) error {
	if ffmpeg == "" {
		return errs.New(errs.ErrFFmpegMissing, "")
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", src}
	if strings.EqualFold(filepath.Ext(dst), ".jpg") {
		args = append(args, "-q:v", "2")
	}
	args = append(args, "-frames:v", "1", "-update", "1", dst)
	var stderr bytes.Buffer
	cmd := sysutil.CommandContext(ctx, ffmpeg, args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errs.New(errs.ErrUnknown, fmt.Sprintf("ffmpeg convert: %v: %s", err, strings.TrimSpace(stderr.String())))
	}
	return nil
}
