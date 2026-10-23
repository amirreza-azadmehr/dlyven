package settings

import (
	"path/filepath"
	"testing"
)

func TestValidateClamps(t *testing.T) {
	s := Settings{Theme: "neon", MaxConcurrent: 99, Retries: -1, TimeoutSec: 1, Accent: "x", ProxyType: "ftp"}.Validate()
	if s.Theme != "system" || s.MaxConcurrent != 10 || s.Retries != 0 || s.TimeoutSec != 5 || s.Accent != "indigo" || s.ProxyType != "http" {
		t.Errorf("validate: %+v", s)
	}
	if s.DownloadDir == "" || s.FilenameTemplate == "" || s.PlatformDirs == nil {
		t.Error("defaults not filled")
	}
}

func TestValidateLanguage(t *testing.T) {
	for in, want := range map[string]string{"": "", "fa": "fa", "ckb": "ckb", "pt-BR": "pt-BR", "x": "", "../etc": "", "EN": ""} {
		if got := (Settings{Language: in}).Validate().Language; got != want {
			t.Errorf("language %q → %q, want %q", in, got, want)
		}
	}
}

func TestProxyURL(t *testing.T) {
	s := Defaults()
	if s.ProxyURL() != "" {
		t.Error("disabled proxy must be empty")
	}
	s.ProxyEnabled, s.ProxyType, s.ProxyHost, s.ProxyPort = true, "socks5", "127.0.0.1", 1080
	if got := s.ProxyURL(); got != "socks5://127.0.0.1:1080" {
		t.Error(got)
	}
	s.ProxyUser, s.ProxyPass = "u", "p@ss"
	if got := s.ProxyURL(); got != "socks5://u:p%40ss@127.0.0.1:1080" {
		t.Error(got)
	}
	s.ProxyHost = "http://10.0.0.1:8080"
	if got := s.ProxyURL(); got != "http://10.0.0.1:8080" {
		t.Error(got)
	}
}

func TestDirFor(t *testing.T) {
	s := Defaults()
	s.DownloadDir = "/d"
	s.LastFolder = "/last"
	s.PlatformDirs = map[string]string{"youtube": "/yt"}
	if s.DirFor("youtube") != "/yt" || s.DirFor("vimeo") != "/last" {
		t.Error("dir resolution wrong")
	}
	s.RememberLastFolder = false
	if s.DirFor("vimeo") != "/d" {
		t.Error("should fall back to download dir")
	}
}

func TestStoreRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "settings.json")
	st, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	st.OnChange(func(Settings) { called = true })
	if _, err := st.Update(func(s *Settings) { s.Language = "ckb"; s.MaxConcurrent = 5 }); err != nil {
		t.Fatal(err)
	}
	st2, _ := Load(p)
	if got := st2.Get(); got.Language != "ckb" || got.MaxConcurrent != 5 || !called {
		t.Errorf("reloaded: %+v called=%v", got, called)
	}
	r, _ := st2.Reset()
	if r.Language != "ckb" || r.MaxConcurrent != 3 {
		t.Errorf("reset should keep language: %+v", r)
	}
}
