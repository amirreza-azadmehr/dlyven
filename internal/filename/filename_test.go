package filename

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSanitize(t *testing.T) {
	cases := map[string]string{
		`What? A "great" <video>: part 1/2 | HD*`: `What？ A 'great' (video)- part 1-2 - HD_`,
		"trailing dots...  ":                      "trailing dots",
		"CON":                                     "_CON",
		"aux.mp4":                                 "_aux.mp4",
		"":                                        "download",
		"..":                                      "download",
		"tab\there\nnewline":                      "tab here newline",
		// Persian with ZWNJ (می‌خواهم) must keep U+200C
		"می‌خواهم":          "می‌خواهم",
		"کوردی: ڕۆژ ڵ ێ ە":  "کوردی- ڕۆژ ڵ ێ ە",
		"Türkçe şarkı İğne": "Türkçe şarkı İğne",
		// RTL override used to disguise extensions is removed
		"invoice‮gpj.exe": "invoicegpj.exe",
		"emoji 🎵 ok":      "emoji 🎵 ok",
	}
	for in, want := range cases {
		if got := Sanitize(in); got != want {
			t.Errorf("Sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeNFC(t *testing.T) {
	nfd := "Café" // e + combining acute
	if got := Sanitize(nfd); got != "Café" {
		t.Errorf("expected NFC, got %q", got)
	}
}

func TestSanitizeTruncatesOnRuneBoundary(t *testing.T) {
	long := strings.Repeat("ژ", 200) // 2 bytes each
	got := Sanitize(long)
	if len(got) > MaxBytes || !utf8.ValidString(got) {
		t.Fatalf("len=%d valid=%v", len(got), utf8.ValidString(got))
	}
}

func TestRender(t *testing.T) {
	v := Vars{"title": "My Video", "quality": "1080p", "uploader": ""}
	cases := map[string]string{
		"{title} - {quality}":          "My Video - 1080p",
		"{title} - {resolution}.{ext}": "My Video",
		"{uploader} - {title}":         "My Video",
		"{title} [{uploader}]":         "My Video",
		"{index}. {title}":             ". My Video",
		"{title}/{quality}":            "My Video-1080p",
	}
	for tpl, want := range cases {
		if got := Render(tpl, v); got != want {
			t.Errorf("Render(%q) = %q, want %q", tpl, got, want)
		}
	}
}

func TestUnique(t *testing.T) {
	dir := t.TempDir()
	if got := Unique(dir, "song", "mp3", nil); got != "song" {
		t.Fatalf("got %q", got)
	}
	os.WriteFile(filepath.Join(dir, "song.mp3"), nil, 0o644)
	if got := Unique(dir, "song", "mp3", nil); got != "song (1)" {
		t.Fatalf("got %q", got)
	}
	reserved := func(p string) bool { return p == strings.ToLower(filepath.Join(dir, "song (1).mp3")) }
	if got := Unique(dir, "song", "mp3", reserved); got != "song (2)" {
		t.Fatalf("got %q", got)
	}
}

func TestEscapeOutputTemplate(t *testing.T) {
	if got := EscapeOutputTemplate("100% real"); got != "100%% real" {
		t.Errorf("got %q", got)
	}
}
