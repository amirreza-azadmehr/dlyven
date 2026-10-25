package engine

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amirreza-azadmehr/dlyven/internal/errs"
	"github.com/amirreza-azadmehr/dlyven/internal/settings"
)

const (
	netscapeHeader   = "# Netscape HTTP Cookie File"
	cookiesFileLimit = 32 << 20
	normalizedPrefix = "cookies-repaired-"
)

// cookieArgs returns the yt-dlp flags for the configured cookie source. A cookies file is
// passed through as one argument (spaces and Windows backslashes are fine: os/exec quotes
// it), after any repair the file needs to be readable by yt-dlp (see prepareCookiesFile).
func (m *Manager) cookieArgs(s settings.Settings) []string {
	switch s.CookiesMode {
	case "browser":
		return []string{"--cookies-from-browser", s.CookiesBrowser}
	case "file":
		if p := strings.TrimSpace(s.CookiesFile); p != "" {
			if fixed, _, err := m.prepareCookiesFile(p); err == nil {
				p = fixed
			}
			return []string{"--cookies", p}
		}
	}
	return nil
}

// CheckCookies reports a problem with the configured cookies file before yt-dlp runs, so the
// user gets an actionable message instead of a raw Python traceback. Browser mode and "none"
// are validated by yt-dlp itself.
func (m *Manager) CheckCookies(s settings.Settings) error {
	if s.CookiesMode != "file" {
		return nil
	}
	p := strings.TrimSpace(s.CookiesFile)
	if p == "" {
		return errs.New(errs.ErrCookies, "no cookies.txt file is selected")
	}
	_, _, err := m.prepareCookiesFile(p)
	return err
}

// prepareCookiesFile validates a user-exported cookies file and returns the path yt-dlp should
// read. yt-dlp only accepts Netscape files that start with the header and use tabs, and
// Windows editors and browser extensions routinely produce files that don't: a UTF-8 BOM
// (Notepad), a missing header, or tabs turned into spaces.
//
// Only that framing is repaired; cookie names, values, domains, paths and expiry times are
// never altered. Anything ambiguous is rejected with the offending line number. The original
// file is only ever read. A repaired copy (0600, named by content hash, written atomically)
// lives in the cache dir and is removed on the next start and on quit (CleanupCookies); a
// file that is already fine is used as is, so yt-dlp can keep refreshing it.
func (m *Manager) prepareCookiesFile(path string) (string, bool, error) {
	path = filepath.Clean(path)
	fi, err := os.Stat(path)
	switch {
	case err != nil:
		return "", false, errs.New(errs.ErrCookies, fmt.Sprintf("cookies file not found: %s", path))
	case fi.IsDir():
		return "", false, errs.New(errs.ErrCookies, fmt.Sprintf("cookies path is a folder, not a file: %s", path))
	case fi.Size() > cookiesFileLimit:
		return "", false, errs.New(errs.ErrCookies, "cookies file is too large to be a cookies.txt export")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", false, errs.New(errs.ErrCookies, fmt.Sprintf("cannot read cookies file: %v", err))
	}
	fixed, err := normalizeNetscape(raw)
	if err != nil {
		return "", false, errs.New(errs.ErrCookies, err.Error())
	}
	if bytes.Equal(fixed, raw) {
		return path, false, nil
	}
	dir := m.cacheDir
	if dir == "" {
		return "", false, errs.New(errs.ErrCookies, "cookies file needs repair but there is no cache folder to write the copy to")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", false, errs.New(errs.ErrCookies, fmt.Sprintf("cannot prepare cookies file: %v", err))
	}
	sum := sha256.Sum256(fixed)
	out := filepath.Join(dir, normalizedPrefix+hex.EncodeToString(sum[:8])+".txt")
	if _, err := os.Stat(out); err != nil {
		tmp, err := os.CreateTemp(dir, normalizedPrefix+"tmp-*")
		if err != nil {
			return "", false, errs.New(errs.ErrCookies, fmt.Sprintf("cannot prepare cookies file: %v", err))
		}
		_, werr := tmp.Write(fixed)
		cerr := tmp.Close()
		if werr != nil || cerr != nil || os.Rename(tmp.Name(), out) != nil {
			_ = os.Remove(tmp.Name())
			return "", false, errs.New(errs.ErrCookies, "cannot write the repaired cookies copy")
		}
		_ = os.Chmod(out, 0o600)
	}
	// Other copies belong to files that are no longer selected.
	pruneNormalized(dir, filepath.Base(out))
	return out, true, nil
}

// CleanupCookies removes repaired cookie copies (they hold session secrets).
func (m *Manager) CleanupCookies() {
	if m.cacheDir != "" {
		pruneNormalized(m.cacheDir, "")
	}
}

func pruneNormalized(dir, keep string) {
	if dir == "" {
		return
	}
	matches, _ := filepath.Glob(filepath.Join(dir, normalizedPrefix+"*"))
	for _, f := range matches {
		if filepath.Base(f) != keep {
			_ = os.Remove(f)
		}
	}
}

// normalizeNetscape repairs the framing of a cookies.txt (BOM, line endings, missing header,
// spaces instead of tabs) and refuses anything it cannot repair without guessing.
func normalizeNetscape(raw []byte) ([]byte, error) {
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))
	if bytes.IndexByte(raw, 0) >= 0 {
		return nil, fmt.Errorf("not a text cookies.txt (binary data found); export it again in Netscape format")
	}
	text := strings.ReplaceAll(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\r", "\n")
	var out []string
	cookies := 0
	hasHeader := false
	for i, line := range strings.Split(text, "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" {
			out = append(out, "")
			continue
		}
		if strings.HasPrefix(trim, "#") && !strings.HasPrefix(trim, "#HttpOnly_") {
			if strings.EqualFold(trim, netscapeHeader) || strings.EqualFold(trim, "# HTTP Cookie File") {
				hasHeader = true
			}
			out = append(out, line)
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 7 {
			// Tabs turned into spaces. Only an exact 7-field split is unambiguous: an empty
			// value or a value with spaces would shift fields, so those lines are refused.
			fields = strings.Fields(trim)
			if len(fields) != 7 {
				return nil, fmt.Errorf("line %d is not a Netscape cookie (7 tab-separated fields expected); export the file again", i+1)
			}
		}
		if !isInt(fields[4]) { // the one field yt-dlp itself refuses to load
			return nil, fmt.Errorf("line %d is not a Netscape cookie (expiry is not a number); export the file again", i+1)
		}
		cookies++
		out = append(out, strings.Join(fields, "\t"))
	}
	if cookies == 0 {
		return nil, fmt.Errorf("no cookies found: expected a Netscape-format cookies.txt")
	}
	body := strings.Join(out, "\n")
	if !hasHeader {
		body = netscapeHeader + "\n" + body
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return []byte(body), nil
}

func isInt(s string) bool {
	s = strings.TrimPrefix(s, "-")
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
