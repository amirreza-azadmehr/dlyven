// Package filename renders user filename templates and makes the result safe on every OS
// while preserving Unicode (Persian, Kurdish, Turkish, CJK, emoji…).
package filename

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Vars are the values available to templates as {name}.
type Vars map[string]string

// Fields lists the supported template fields (shown in the settings UI).
var Fields = []string{"title", "uploader", "quality", "resolution", "fps", "platform", "id", "date", "playlist", "index", "artist", "album", "track"}

var fieldRe = regexp.MustCompile(`\{([a-z]+)\}`)

// MaxBytes keeps names comfortably under the 255-byte limit of most filesystems
// (leaves room for " (99)", ".ext", ".part", yt-dlp's ".f137" and subtitle suffixes).
const MaxBytes = 180

// Render substitutes {field} placeholders. A trailing ".{ext}" is dropped because the
// extension is decided by the download (container / audio format).
// Separators left dangling by empty fields (" - ", "__") are collapsed.
func Render(tpl string, v Vars) string {
	tpl = strings.TrimSuffix(strings.TrimSpace(tpl), ".{ext}")
	tpl = strings.ReplaceAll(tpl, "{ext}", "")
	out := fieldRe.ReplaceAllStringFunc(tpl, func(m string) string {
		key := m[1 : len(m)-1]
		return strings.TrimSpace(v[key])
	})
	out = collapseSeparators(out)
	return Sanitize(out)
}

var (
	multiSpaceRe = regexp.MustCompile(`\s{2,}`)
	danglingRe   = regexp.MustCompile(`(\s*[-_–—|·]\s*){2,}`)
	emptyParenRe = regexp.MustCompile(`\(\s*\)|\[\s*\]`)
)

func collapseSeparators(s string) string {
	s = emptyParenRe.ReplaceAllString(s, "")
	s = danglingRe.ReplaceAllStringFunc(s, func(m string) string {
		// keep the first separator, normalized with spaces
		for _, r := range m {
			if !unicode.IsSpace(r) {
				return " " + string(r) + " "
			}
		}
		return " "
	})
	s = multiSpaceRe.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "-_–—|· ")
	return s
}

var reservedWin = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// Sanitize makes a single path component safe on Windows, macOS and Linux:
//   - NFC-normalizes (macOS hands out NFD; Persian/Kurdish combining marks stay intact)
//   - replaces characters illegal on Windows  < > : " / \ | ? *  with look-alike safe ones
//   - removes control characters and bidi override/embedding controls (which can be used
//     to disguise extensions, e.g. "gpj.exe" shown as "exe.jpg"), but KEEPS ZWNJ/ZWJ,
//     which are required for correct Persian and Kurdish spelling
//   - trims trailing dots/spaces (Windows) and avoids reserved device names
//   - truncates to MaxBytes without splitting a UTF-8 sequence
func Sanitize(s string) string {
	s = norm.NFC.String(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '/' || r == '\\':
			b.WriteRune('-')
		case r == ':':
			b.WriteRune('-')
		case r == '*':
			b.WriteRune('_')
		case r == '?':
			b.WriteRune('？')
		case r == '"':
			b.WriteRune('\'')
		case r == '<':
			b.WriteRune('(')
		case r == '>':
			b.WriteRune(')')
		case r == '|':
			b.WriteRune('-')
		case r == '‌' || r == '‍':
			b.WriteRune(r) // ZWNJ / ZWJ — keep
		case isBidiControl(r):
			// drop
		case unicode.IsControl(r) || r == utf8.RuneError:
			b.WriteRune(' ')
		case unicode.Is(unicode.Cf, r) && r != '‌' && r != '‍':
			// other invisible format characters (soft hyphen, BOM…)
		default:
			b.WriteRune(r)
		}
	}
	out := multiSpaceRe.ReplaceAllString(b.String(), " ")
	out = strings.TrimSpace(out)
	out = truncateBytes(out, MaxBytes)
	out = strings.TrimRight(out, ". ")
	out = strings.TrimLeft(out, " ")
	if out == "" || out == "." || out == ".." {
		out = "download"
	}
	base := out
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	if reservedWin[strings.ToUpper(strings.TrimSpace(base))] {
		out = "_" + out
	}
	return out
}

func isBidiControl(r rune) bool {
	return (r >= '‪' && r <= '‮') || (r >= '⁦' && r <= '⁩') || r == '‎' || r == '‏' || r == '؜'
}

func truncateBytes(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := 0
	for i := range s {
		if i > max {
			break
		}
		cut = i
	}
	// make sure we don't end on a ZWNJ or combining mark orphan
	out := strings.TrimRightFunc(s[:cut], func(r rune) bool {
		return r == '‌' || r == '‍' || unicode.Is(unicode.Mn, r) || unicode.IsSpace(r)
	})
	return out
}

// Unique returns base (or "base (n)") such that dir/base.ext does not exist and isn't
// reserved by another queued job. Comparison is case-insensitive to be safe on
// Windows/macOS filesystems.
func Unique(dir, base, ext string, reserved func(path string) bool) string {
	candidate := base
	for i := 1; i < 10000; i++ {
		p := filepath.Join(dir, candidate+dotExt(ext))
		if !exists(p) && (reserved == nil || !reserved(strings.ToLower(p))) {
			return candidate
		}
		candidate = base + " (" + strconv.Itoa(i) + ")"
	}
	return candidate
}

func dotExt(ext string) string {
	if ext == "" {
		return ""
	}
	return "." + strings.TrimPrefix(ext, ".")
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// EscapeOutputTemplate escapes a literal string for use inside a yt-dlp output template.
func EscapeOutputTemplate(s string) string {
	return strings.ReplaceAll(s, "%", "%%")
}
