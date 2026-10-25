package errs

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := map[string]Code{
		"ERROR: [youtube] abc: Sign in to confirm your age. This video may be inappropriate for some users.":                                        ErrAgeRestricted,
		"ERROR: [youtube] abc: Sign in to confirm you’re not a bot. Use --cookies-from-browser":                                                     ErrLoginRequired,
		"ERROR: [Instagram] xyz: Requested content is not available, rate-limit reached or login required":                                          ErrLoginRequired,
		"ERROR: [Instagram] xyz: Instagram sent an empty media response.":                                                                           ErrLoginRequired,
		"ERROR: [youtube] abc: Private video. Sign in if you've been granted access":                                                                ErrPrivate,
		"ERROR: [youtube] abc: The uploader has not made this video available in your country":                                                      ErrGeoBlocked,
		"ERROR: Unsupported URL: https://example.com/":                                                                                              ErrUnsupportedURL,
		"ERROR: unable to download video data: HTTP Error 404: Not Found":                                                                           ErrNotFound,
		"ERROR: unable to download video data: HTTP Error 429: Too Many Requests":                                                                   ErrRateLimited,
		"ERROR: Unable to download webpage: <urlopen error [Errno 11001] getaddrinfo failed>":                                                       ErrNetwork,
		"ERROR: Unable to download webpage: ('Unable to connect to proxy', OSError('Tunnel connection failed: 407 Proxy Authentication Required'))": ErrProxy,
		"ERROR: Postprocessing: ffprobe and ffmpeg not found. Please install or provide the path":                                                   ErrFFmpegMissing,
		"ERROR: Could not copy Chrome cookie database. See https://github.com/yt-dlp/yt-dlp/issues/7271":                                            ErrCookiesLocked,
		"ERROR: Failed to decrypt with DPAPI. See  https://github.com/yt-dlp/yt-dlp/issues/10927  for more info":                                    ErrCookiesEncrypted,
		"ERROR: Could not copy Edge cookie database. See  https://github.com/yt-dlp/yt-dlp/issues/7271  for more info":                              ErrCookiesLocked,
		"ERROR: could not find firefox cookies database in C:\\Users\\a\\AppData":                                                                   ErrCookies,
		"ERROR: [youtube] abc: This live event will begin in 3 hours.":                                                                              ErrLiveNotStarted,
		"ERROR: [Spotify] This video is DRM protected":                                                                                              ErrDRM,
		"ERROR: [Errno 28] No space left on device":                                                                                                 ErrDisk,
		"ERROR: Requested format is not available. Use --list-formats":                                                                              ErrFormatUnavailable,
		"something weird happened": ErrUnknown,
	}
	for in, want := range cases {
		if got := Classify(in); got != want {
			t.Errorf("Classify(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestClassifyIgnoresWarnings(t *testing.T) {
	out := "WARNING: [youtube] Use --cookies-from-browser for better results\nERROR: unable to download video data: HTTP Error 404: Not Found"
	if got := Classify(out); got != ErrNotFound {
		t.Errorf("got %s", got)
	}
}

func TestErrorJSONRoundTrip(t *testing.T) {
	e := New(ErrNetwork, "boom")
	var back Error
	if err := json.Unmarshal([]byte(e.Error()), &back); err != nil || back.Code != ErrNetwork || back.Detail != "boom" {
		t.Errorf("round trip failed: %v %+v", err, back)
	}
}

func TestCodeOf(t *testing.T) {
	if CodeOf(fmt.Errorf("wrapped: %w", New(ErrPrivate, ""))) != ErrPrivate {
		t.Error("wrapped code lost")
	}
	if CodeOf(context.Canceled) != ErrCanceled || CodeOf(nil) != "" || CodeOf(fmt.Errorf("x")) != ErrUnknown {
		t.Error("CodeOf mapping wrong")
	}
	if Wrap(ErrDisk, nil) != nil {
		t.Error("Wrap(nil) should be nil")
	}
}

func TestLastErrorLine(t *testing.T) {
	out := "[info] x\nERROR: first\nsome trailing\n"
	if got := LastErrorLine(out); got != "ERROR: first" {
		t.Errorf("got %q", got)
	}
}
