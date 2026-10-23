// Package errs defines the stable error codes the backend returns to the frontend.
// The frontend maps every code to a localized, human-friendly message (errors.<CODE> keys).
package errs

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

type Code string

const (
	ErrUnknown            Code = "ERR_UNKNOWN"
	ErrInvalidURL         Code = "ERR_INVALID_URL"
	ErrUnsupportedURL     Code = "ERR_UNSUPPORTED_URL"
	ErrLoginRequired      Code = "ERR_LOGIN_REQUIRED"
	ErrAgeRestricted      Code = "ERR_AGE_RESTRICTED"
	ErrPrivate            Code = "ERR_PRIVATE"
	ErrGeoBlocked         Code = "ERR_GEO_BLOCKED"
	ErrNotFound           Code = "ERR_NOT_FOUND"
	ErrRateLimited        Code = "ERR_RATE_LIMITED"
	ErrDRM                Code = "ERR_DRM"
	ErrNetwork            Code = "ERR_NETWORK"
	ErrTimeout            Code = "ERR_TIMEOUT"
	ErrProxy              Code = "ERR_PROXY"
	ErrCookies            Code = "ERR_COOKIES"
	ErrFFmpegMissing      Code = "ERR_FFMPEG_MISSING"
	ErrEngineMissing      Code = "ERR_ENGINE_MISSING"
	ErrEngineInstall      Code = "ERR_ENGINE_INSTALL"
	ErrChecksum           Code = "ERR_CHECKSUM"
	ErrDisk               Code = "ERR_DISK"
	ErrCanceled           Code = "ERR_CANCELED"
	ErrLiveNotStarted     Code = "ERR_LIVE_NOT_STARTED"
	ErrFormatUnavailable  Code = "ERR_FORMAT_UNAVAILABLE"
	ErrSpotifyUnsupported Code = "ERR_SPOTIFY_UNSUPPORTED"
	ErrSpotifyMatch       Code = "ERR_SPOTIFY_MATCH"
	ErrBusy               Code = "ERR_BUSY"
	ErrNotFoundJob        Code = "ERR_JOB_NOT_FOUND"
)

// Error is the wire format. Error() serializes to JSON so Wails rejects the JS promise
// with a string the frontend can parse back into {code, detail}.
type Error struct {
	Code   Code   `json:"code"`
	Detail string `json:"detail,omitempty"`
}

func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func New(code Code, detail string) *Error { return &Error{Code: code, Detail: detail} }

// Wrap attaches a code to an underlying error, preserving an existing code if present.
func Wrap(code Code, err error) error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	if errors.Is(err, context.Canceled) {
		return New(ErrCanceled, "")
	}
	return New(code, err.Error())
}

// CodeOf extracts the code from err, or ErrUnknown.
func CodeOf(err error) Code {
	if err == nil {
		return ""
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	if errors.Is(err, context.Canceled) {
		return ErrCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}
	return ErrUnknown
}

// DetailOf returns the human-readable detail for logging/diagnostics.
func DetailOf(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Detail
	}
	if err == nil {
		return ""
	}
	return err.Error()
}

// Retryable reports whether an automatic retry has a reasonable chance to succeed.
func Retryable(c Code) bool {
	switch c {
	case ErrNetwork, ErrTimeout, ErrRateLimited, ErrUnknown:
		return true
	}
	return false
}

type rule struct {
	code    Code
	needles []string
}

// Order matters: more specific rules come first.
var rules = []rule{
	{ErrFFmpegMissing, []string{"ffmpeg not found", "ffprobe and ffmpeg not found", "ffmpeg is not installed", "ffprobe/avprobe and ffmpeg/avconv not found"}},
	{ErrCookies, []string{"could not copy chrome cookie database", "could not copy", "failed to decrypt", "cookies database", "could not find firefox cookies", "could not find chrome cookies", "unsupported browser specified for cookies", "failed to load cookies"}},
	{ErrDRM, []string{"drm protected", "this video is drm", " drm", "widevine"}},
	{ErrAgeRestricted, []string{"sign in to confirm your age", "age-restricted", "age restricted", "inappropriate for some users", "confirm your age"}},
	{ErrPrivate, []string{"private video", "this video is private", "this account is private", "private account", "is private"}},
	{ErrLoginRequired, []string{"login required", "log in", "sign in to confirm", "requires authentication", "use --cookies", "--cookies-from-browser", "logged-in", "authentication", "members-only", "join this channel", "account to view", "rate-limit reached or login required", "empty media response", "unable to extract shared data", "accounts/login"}},
	{ErrGeoBlocked, []string{"not available in your country", "available in your country", "geo restrict", "geo-restrict", "blocked it in your country", "not available from your location", "in your region"}},
	{ErrLiveNotStarted, []string{"live event will begin", "premieres in", "this live event", "is not currently live", "is upcoming"}},
	{ErrRateLimited, []string{"http error 429", "too many requests", "rate limit", "rate-limit"}},
	{ErrFormatUnavailable, []string{"requested format is not available", "requested format not available"}},
	{ErrUnsupportedURL, []string{"unsupported url", "no suitable extractor", "is not a supported url"}},
	{ErrInvalidURL, []string{"is not a valid url", "invalid url", "not a valid url"}},
	{ErrNotFound, []string{"http error 404", "video unavailable", "does not exist", "has been removed", "not found", "no video formats found", "this content isn't available", "no video could be found"}},
	{ErrProxy, []string{"proxy", "tunnel connection failed", "socks"}},
	{ErrTimeout, []string{"timed out", "timeout"}},
	{ErrDisk, []string{"no space left", "errno 28", "disk full", "permission denied", "access is denied", "read-only file system", "errno 13"}},
	{ErrNetwork, []string{"unable to download", "connection", "network is unreachable", "name or service not known", "getaddrinfo failed", "temporary failure in name resolution", "nodename nor servname", "ssl", "urlopen error", "remote end closed", "eof occurred", "http error 5", "certificate", "no route to host", "dial tcp", "i/o timeout"}},
}

// Classify maps raw yt-dlp/ffmpeg/network output to a code. When the output contains
// yt-dlp "ERROR:" lines only those are considered, so warnings don't cause misclassification.
func Classify(output string) Code {
	var errLines []string
	for _, l := range strings.Split(output, "\n") {
		if strings.Contains(l, "ERROR") {
			errLines = append(errLines, l)
		}
	}
	if len(errLines) > 0 {
		output = strings.Join(errLines, "\n")
	}
	low := strings.ToLower(output)
	if strings.Contains(low, "proxy authentication") || strings.Contains(low, "unable to connect to proxy") {
		return ErrProxy
	}
	for _, r := range rules {
		for _, n := range r.needles {
			if strings.Contains(low, n) {
				return r.code
			}
		}
	}
	return ErrUnknown
}

// LastErrorLine returns the most relevant "ERROR:" line from yt-dlp output (or the last non-empty line).
func LastErrorLine(output string) string {
	lines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	last := ""
	for i := len(lines) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines[i])
		if l == "" {
			continue
		}
		if strings.HasPrefix(l, "ERROR:") {
			return l
		}
		if last == "" {
			last = l
		}
	}
	return last
}
