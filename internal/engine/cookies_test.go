package engine

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/amirreza-azadmehr/dlyven/internal/errs"
	"github.com/amirreza-azadmehr/dlyven/internal/settings"
)

const goodLine = ".youtube.com\tTRUE\t/\tTRUE\t1893456000\tSID\tabc123"

func newTestManager(t *testing.T, s settings.Settings) *Manager {
	t.Helper()
	dir := t.TempDir()
	m := New(dir, filepath.Join(dir, "cache dir"), func() settings.Settings { return s }, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	m.impersonate = "off" // no yt-dlp probe in tests
	return m
}

func writeCookies(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func argAfter(args []string, flag string) (string, bool) {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1], true
		}
	}
	return "", false
}

func TestBrowserArgsPerBrowser(t *testing.T) {
	for _, b := range settings.CookieBrowsers {
		m := newTestManager(t, settings.Settings{CookiesMode: "browser", CookiesBrowser: b})
		got, ok := argAfter(m.BaseArgs(context.Background(), m.settings()), "--cookies-from-browser")
		if !ok || got != b {
			t.Errorf("browser %q: got %q (present=%v)", b, got, ok)
		}
	}
}

func TestNoCookieArgsWhenOff(t *testing.T) {
	m := newTestManager(t, settings.Settings{CookiesMode: "none", CookiesBrowser: "chrome", CookiesFile: "x.txt"})
	args := m.BaseArgs(context.Background(), m.settings())
	for _, a := range args {
		if a == "--cookies" || a == "--cookies-from-browser" {
			t.Fatalf("cookie flag present with mode none: %v", args)
		}
	}
}

func TestCookiesFileWithSpacesIsOneArgument(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "My Cookies Folder", "نمونه")
	_ = os.MkdirAll(dir, 0o755)
	p := writeCookies(t, dir, "my cookies.txt", netscapeHeader+"\n"+goodLine+"\n")
	m := newTestManager(t, settings.Settings{CookiesMode: "file", CookiesFile: "  " + p + " "})
	got, ok := argAfter(m.BaseArgs(context.Background(), m.settings()), "--cookies")
	if !ok || got != p {
		t.Fatalf("got %q, want the untouched original %q", got, p)
	}
}

func TestPrepareCookiesFileRepairs(t *testing.T) {
	cases := map[string]string{
		"bom":        "\xef\xbb\xbf" + netscapeHeader + "\n" + goodLine + "\n",
		"no header":  goodLine + "\n",
		"crlf":       netscapeHeader + "\r\n" + goodLine + "\r\n",
		"spaces":     netscapeHeader + "\n" + strings.ReplaceAll(goodLine, "\t", "  ") + "\n",
		"neg expiry": netscapeHeader + "\n.a.com\tTRUE\t/\tFALSE\t-1\tN\tv\n",
		"httponly":   netscapeHeader + "\n#HttpOnly_" + goodLine + "\n",
		"cr only":    netscapeHeader + "\r" + goodLine + "\r",
		"no newline": netscapeHeader + "\n" + goodLine,
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			m := newTestManager(t, settings.Settings{})
			src := writeCookies(t, t.TempDir(), "c.txt", content)
			out, changed, err := m.prepareCookiesFile(src)
			if err != nil {
				t.Fatal(err)
			}
			if !changed && name != "httponly" && name != "neg expiry" {
				t.Fatalf("expected a repaired copy, got the original")
			}
			b, _ := os.ReadFile(out)
			text := string(b)
			if !strings.HasPrefix(text, netscapeHeader+"\n") || strings.ContainsAny(text, "\r") || strings.HasPrefix(text, "\xef") {
				t.Fatalf("not normalized: %q", text)
			}
			if !strings.Contains(text, "\tSID\tabc123") && name != "neg expiry" {
				t.Fatalf("cookie lost: %q", text)
			}
			if changed && strings.Contains(src, filepath.Dir(out)) && out == src {
				t.Fatal("original must not be overwritten")
			}
		})
	}
}

func TestPrepareCookiesFileKeepsGoodFile(t *testing.T) {
	m := newTestManager(t, settings.Settings{})
	src := writeCookies(t, t.TempDir(), "c.txt", netscapeHeader+"\n"+goodLine+"\n")
	out, changed, err := m.prepareCookiesFile(src)
	if err != nil || changed || out != src {
		t.Fatalf("out=%q changed=%v err=%v", out, changed, err)
	}
}

func TestCheckCookiesRejectsBadFiles(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]settings.Settings{
		"missing":    {CookiesMode: "file", CookiesFile: filepath.Join(dir, "nope.txt")},
		"folder":     {CookiesMode: "file", CookiesFile: dir},
		"empty":      {CookiesMode: "file", CookiesFile: ""},
		"not cookie": {CookiesMode: "file", CookiesFile: writeCookies(t, dir, "x.txt", "hello world\n")},
		"json":       {CookiesMode: "file", CookiesFile: writeCookies(t, dir, "x.json", `[{"name":"SID","value":"1"}]`)},
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) {
			err := newTestManager(t, s).CheckCookies(s)
			if errs.CodeOf(err) != errs.ErrCookies {
				t.Fatalf("got %v", err)
			}
		})
	}
	for _, s := range []settings.Settings{{CookiesMode: "none"}, {CookiesMode: "browser", CookiesBrowser: "chrome"}} {
		if err := newTestManager(t, s).CheckCookies(s); err != nil {
			t.Fatalf("%s: %v", s.CookiesMode, err)
		}
	}
}

func TestRepairNeverChangesCookieData(t *testing.T) {
	body := netscapeHeader + "\n" +
		".a.com\tTRUE\t/p a/\tFALSE\t0\tn\tv a l\n" + // value with spaces
		"b.com\tFALSE\t/\tTRUE\t1893456000\temptyval\t\n" + // empty value
		"#HttpOnly_.c.com\tTRUE\t/\tTRUE\t1893456000\tH\tx==\n"
	m := newTestManager(t, settings.Settings{})
	src := writeCookies(t, t.TempDir(), "c.txt", "\xef\xbb\xbf"+strings.ReplaceAll(body, "\n", "\r\n"))
	out, changed, err := m.prepareCookiesFile(src)
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	b, _ := os.ReadFile(out)
	if string(b) != body {
		t.Fatalf("data changed:\n got %q\nwant %q", b, body)
	}
}

func TestAmbiguousFilesAreRejectedNotGuessed(t *testing.T) {
	cases := map[string]string{
		"space file, empty value":  netscapeHeader + "\n.a.com TRUE / FALSE 0 n\n",
		"space file, spaced value": netscapeHeader + "\n.a.com TRUE / FALSE 0 n a b\n",
		"bad expiry":               netscapeHeader + "\n.a.com\tTRUE\t/\tFALSE\tsoon\tn\tv\n",
		"binary":                   netscapeHeader + "\n.a.com\tTRUE\t/\tFALSE\t0\tn\tv\x00\n",
		"one bad line among good":  netscapeHeader + "\n" + goodLine + "\ngarbage\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			m := newTestManager(t, settings.Settings{})
			src := writeCookies(t, t.TempDir(), "c.txt", content)
			if _, _, err := m.prepareCookiesFile(src); errs.CodeOf(err) != errs.ErrCookies {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestOriginalNeverModifiedAndCopyIsPrivateAndCleaned(t *testing.T) {
	m := newTestManager(t, settings.Settings{})
	content := "\xef\xbb\xbf" + goodLine + "\r\n"
	src := writeCookies(t, t.TempDir(), "c.txt", content)
	out, changed, err := m.prepareCookiesFile(src)
	if err != nil || !changed || out == src {
		t.Fatalf("out=%q changed=%v err=%v", out, changed, err)
	}
	if b, _ := os.ReadFile(src); string(b) != content {
		t.Fatal("original was modified")
	}
	if runtime.GOOS != "windows" {
		if fi, _ := os.Stat(out); fi.Mode().Perm() != 0o600 {
			t.Fatalf("copy mode %v, want 0600", fi.Mode().Perm())
		}
	}
	again, _, _ := m.prepareCookiesFile(src)
	if again != out {
		t.Fatalf("same content should reuse the same copy: %q vs %q", again, out)
	}
	m.CleanupCookies()
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("copy survived CleanupCookies")
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatal("cleanup removed the original")
	}
}

func TestStaleCopiesRemovedOnNewAndSwitch(t *testing.T) {
	m := newTestManager(t, settings.Settings{})
	a := writeCookies(t, t.TempDir(), "a.txt", goodLine+"\n")
	b := writeCookies(t, t.TempDir(), "b.txt", strings.Replace(goodLine, "abc123", "zzz", 1)+"\n")
	outA, _, _ := m.prepareCookiesFile(a)
	outB, _, _ := m.prepareCookiesFile(b)
	if _, err := os.Stat(outA); !os.IsNotExist(err) {
		t.Fatal("copy of the previously selected file was left behind")
	}
	m2 := New(m.binDir, m.cacheDir, m.settings, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if _, err := os.Stat(outB); !os.IsNotExist(err) {
		t.Fatal("copy survived a restart")
	}
	_ = m2
}
