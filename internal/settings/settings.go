// Package settings loads, validates and persists user preferences as JSON.
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Settings struct {
	// Appearance
	Language     string `json:"language"` // "" = follow system
	Theme        string `json:"theme"`    // system | dark | light
	Accent       string `json:"accent"`   // indigo | violet | blue | teal | rose | amber
	NativeDigits bool   `json:"nativeDigits"`
	Calendar     string `json:"calendar"` // gregorian | jalali
	Translucency bool   `json:"translucency"`

	// Downloads
	DownloadDir         string            `json:"downloadDir"`
	RememberLastFolder  bool              `json:"rememberLastFolder"`
	LastFolder          string            `json:"lastFolder"`
	PlatformDirs        map[string]string `json:"platformDirs"`
	FilenameTemplate    string            `json:"filenameTemplate"`
	MaxConcurrent       int               `json:"maxConcurrent"`
	SpeedLimitKBps      int               `json:"speedLimitKBps"` // 0 = unlimited
	DefaultContainer    string            `json:"defaultContainer"`
	DefaultAudioFormat  string            `json:"defaultAudioFormat"`
	DefaultAudioBitrate string            `json:"defaultAudioBitrate"`
	EmbedThumbnail      bool              `json:"embedThumbnail"`
	EmbedMetadata       bool              `json:"embedMetadata"`
	EmbedChapters       bool              `json:"embedChapters"`

	// Network
	ProxyEnabled bool   `json:"proxyEnabled"`
	ProxyType    string `json:"proxyType"` // http | https | socks5
	ProxyHost    string `json:"proxyHost"`
	ProxyPort    int    `json:"proxyPort"`
	ProxyUser    string `json:"proxyUser"`
	ProxyPass    string `json:"proxyPass"`
	UserAgent    string `json:"userAgent"`
	Retries      int    `json:"retries"`
	TimeoutSec   int    `json:"timeoutSec"`

	// Cookies
	CookiesMode    string `json:"cookiesMode"` // none | browser | file
	CookiesBrowser string `json:"cookiesBrowser"`
	CookiesFile    string `json:"cookiesFile"`

	// Engine
	YtDlpPath        string `json:"ytDlpPath"`  // "" = auto
	FfmpegPath       string `json:"ffmpegPath"` // "" = auto
	AutoUpdateEngine bool   `json:"autoUpdateEngine"`

	// Behaviour
	ClipboardWatch   bool `json:"clipboardWatch"`
	OpenFolderOnDone bool `json:"openFolderOnDone"`
	Notify           bool `json:"notify"`
	PlaySound        bool `json:"playSound"`
}

var (
	Themes         = []string{"system", "dark", "light"}
	Accents        = []string{"indigo", "violet", "blue", "teal", "rose", "amber"}
	Containers     = []string{"mp4", "mkv", "webm"}
	AudioFormats   = []string{"mp3", "m4a", "opus", "flac", "wav"}
	AudioBitrates  = []string{"best", "320", "256", "192", "128", "96"}
	ProxyTypes     = []string{"http", "https", "socks5"}
	CookiesModes   = []string{"none", "browser", "file"}
	CookieBrowsers = []string{"chrome", "edge", "firefox", "brave", "chromium", "opera", "vivaldi", "safari"}
)

// languageRe accepts any BCP-47-ish language code, so adding a language needs no Go change
// (the frontend registry in src/i18n/index.ts is the source of truth).
var languageRe = regexp.MustCompile(`^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

const DefaultTemplate = "{title} - {quality}"

func defaultDownloadDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, "Downloads")
}

func Defaults() Settings {
	return Settings{
		Language:            "",
		Theme:               "system",
		Accent:              "indigo",
		Calendar:            "gregorian",
		Translucency:        true,
		DownloadDir:         defaultDownloadDir(),
		RememberLastFolder:  true,
		PlatformDirs:        map[string]string{},
		FilenameTemplate:    DefaultTemplate,
		MaxConcurrent:       3,
		DefaultContainer:    "mp4",
		DefaultAudioFormat:  "mp3",
		DefaultAudioBitrate: "320",
		EmbedThumbnail:      true,
		EmbedMetadata:       true,
		EmbedChapters:       true,
		ProxyType:           "http",
		Retries:             3,
		TimeoutSec:          30,
		CookiesMode:         "none",
		CookiesBrowser:      "chrome",
		AutoUpdateEngine:    true,
		ClipboardWatch:      false,
		Notify:              true,
		PlaySound:           true,
	}
}

func oneOf(v string, allowed []string, def string) string {
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	return def
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Validate returns a copy with every field coerced into its valid range.
func (s Settings) Validate() Settings {
	d := Defaults()
	if s.Language != "" && !languageRe.MatchString(s.Language) {
		s.Language = ""
	}
	s.Theme = oneOf(s.Theme, Themes, d.Theme)
	s.Accent = oneOf(s.Accent, Accents, d.Accent)
	s.Calendar = oneOf(s.Calendar, []string{"gregorian", "jalali"}, d.Calendar)
	if strings.TrimSpace(s.DownloadDir) == "" {
		s.DownloadDir = d.DownloadDir
	}
	if s.PlatformDirs == nil {
		s.PlatformDirs = map[string]string{}
	}
	for k, v := range s.PlatformDirs {
		if strings.TrimSpace(v) == "" {
			delete(s.PlatformDirs, k)
		}
	}
	if strings.TrimSpace(s.FilenameTemplate) == "" {
		s.FilenameTemplate = d.FilenameTemplate
	}
	s.MaxConcurrent = clamp(s.MaxConcurrent, 1, 10)
	s.SpeedLimitKBps = clamp(s.SpeedLimitKBps, 0, 1_000_000)
	s.DefaultContainer = oneOf(s.DefaultContainer, Containers, d.DefaultContainer)
	s.DefaultAudioFormat = oneOf(s.DefaultAudioFormat, AudioFormats, d.DefaultAudioFormat)
	s.DefaultAudioBitrate = oneOf(s.DefaultAudioBitrate, AudioBitrates, d.DefaultAudioBitrate)
	s.ProxyType = oneOf(s.ProxyType, ProxyTypes, d.ProxyType)
	s.ProxyHost = strings.TrimSpace(s.ProxyHost)
	s.ProxyPort = clamp(s.ProxyPort, 0, 65535)
	s.UserAgent = strings.TrimSpace(s.UserAgent)
	s.Retries = clamp(s.Retries, 0, 10)
	s.TimeoutSec = clamp(s.TimeoutSec, 5, 300)
	s.CookiesMode = oneOf(s.CookiesMode, CookiesModes, d.CookiesMode)
	s.CookiesBrowser = oneOf(s.CookiesBrowser, CookieBrowsers, d.CookiesBrowser)
	s.YtDlpPath = strings.TrimSpace(s.YtDlpPath)
	s.FfmpegPath = strings.TrimSpace(s.FfmpegPath)
	return s
}

// ProxyURL returns the proxy URL for yt-dlp and Go's HTTP client, or "" if disabled.
func (s Settings) ProxyURL() string {
	if !s.ProxyEnabled || s.ProxyHost == "" {
		return ""
	}
	host := s.ProxyHost
	// Accept a full URL pasted into the host field.
	if strings.Contains(host, "://") {
		if u, err := url.Parse(host); err == nil && u.Host != "" {
			return u.String()
		}
	}
	if s.ProxyPort > 0 {
		host = host + ":" + strconv.Itoa(s.ProxyPort)
	}
	u := url.URL{Scheme: s.ProxyType, Host: host}
	if s.ProxyUser != "" {
		u.User = url.UserPassword(s.ProxyUser, s.ProxyPass)
	}
	return u.String()
}

// DirFor returns the destination folder for a platform, honouring per-platform overrides.
func (s Settings) DirFor(platformID string) string {
	if d := s.PlatformDirs[platformID]; d != "" {
		return d
	}
	if s.RememberLastFolder && s.LastFolder != "" {
		return s.LastFolder
	}
	return s.DownloadDir
}

// Redacted returns a copy safe to include in diagnostics.
func (s Settings) Redacted() Settings {
	if s.ProxyPass != "" {
		s.ProxyPass = "***"
	}
	if s.ProxyUser != "" {
		s.ProxyUser = "***"
	}
	if s.CookiesFile != "" {
		s.CookiesFile = "(set)"
	}
	return s
}

// Store persists Settings to a JSON file with atomic writes.
type Store struct {
	path string
	mu   sync.RWMutex
	cur  Settings
	subs []func(Settings)
}

func Load(path string) (*Store, error) {
	st := &Store{path: path, cur: Defaults()}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return st, st.write(st.cur)
		}
		return st, err
	}
	s := Defaults()
	if err := json.Unmarshal(b, &s); err != nil {
		// Corrupt file: keep a backup and start from defaults rather than failing to launch.
		_ = os.Rename(path, path+".bak")
		return st, st.write(st.cur)
	}
	st.cur = s.Validate()
	return st, nil
}

func (st *Store) Get() Settings {
	st.mu.RLock()
	defer st.mu.RUnlock()
	c := st.cur
	c.PlatformDirs = make(map[string]string, len(st.cur.PlatformDirs))
	for k, v := range st.cur.PlatformDirs {
		c.PlatformDirs[k] = v
	}
	return c
}

// OnChange registers a callback invoked (synchronously) after each successful save.
func (st *Store) OnChange(fn func(Settings)) {
	st.mu.Lock()
	st.subs = append(st.subs, fn)
	st.mu.Unlock()
}

func (st *Store) Save(s Settings) (Settings, error) {
	s = s.Validate()
	if err := st.write(s); err != nil {
		return st.Get(), err
	}
	st.mu.Lock()
	st.cur = s
	subs := append([]func(Settings){}, st.subs...)
	st.mu.Unlock()
	for _, fn := range subs {
		fn(s)
	}
	return s, nil
}

// Update applies a mutation to the current settings and saves.
func (st *Store) Update(fn func(*Settings)) (Settings, error) {
	s := st.Get()
	fn(&s)
	return st.Save(s)
}

func (st *Store) Reset() (Settings, error) {
	cur := st.Get()
	d := Defaults()
	// Keep the language so the UI doesn't jump unexpectedly after a reset.
	d.Language = cur.Language
	return st.Save(d)
}

func (st *Store) write(s Settings) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(st.path), 0o755); err != nil {
		return err
	}
	tmp := st.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return os.Rename(tmp, st.path)
}
