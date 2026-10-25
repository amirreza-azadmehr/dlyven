package engine

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var allBrowsers = []string{"chrome", "edge", "firefox", "brave", "chromium", "opera", "vivaldi", "safari"}

func mkdirs(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInstalledBrowsersWindows(t *testing.T) {
	home := t.TempDir()
	local := filepath.Join(home, "Local Data é")
	roaming := filepath.Join(home, "Roaming Data")
	mkdirs(t,
		filepath.Join(local, "Google", "Chrome", "User Data"),
		filepath.Join(local, "Microsoft", "Edge", "User Data"),
		filepath.Join(roaming, "Mozilla", "Firefox", "Profiles"),
	)
	env := func(k string) string { return map[string]string{"LOCALAPPDATA": local, "APPDATA": roaming}[k] }
	got := installedBrowsers(allBrowsers, browserDirs("windows", home, env))
	if want := []string{"chrome", "edge", "firefox"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInstalledBrowsersWindowsEnvFallback(t *testing.T) {
	home := t.TempDir()
	mkdirs(t, filepath.Join(home, "AppData", "Local", "BraveSoftware", "Brave-Browser", "User Data"))
	got := installedBrowsers(allBrowsers, browserDirs("windows", home, func(string) string { return "" }))
	if want := []string{"brave"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInstalledBrowsersDarwin(t *testing.T) {
	home := t.TempDir()
	mkdirs(t,
		filepath.Join(home, "Library", "Application Support", "Google", "Chrome"),
		filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles"),
		filepath.Join(home, "Library", "Cookies"),
	)
	got := installedBrowsers(allBrowsers, browserDirs("darwin", home, func(string) string { return "" }))
	if want := []string{"chrome", "firefox", "safari"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestInstalledBrowsersLinuxIncludingSnapAndFlatpak(t *testing.T) {
	home := t.TempDir()
	mkdirs(t,
		filepath.Join(home, ".config", "google-chrome"),
		filepath.Join(home, "snap", "firefox", "common", ".mozilla", "firefox"),
		filepath.Join(home, "snap", "chromium", "common", "chromium"),
	)
	got := installedBrowsers(allBrowsers, browserDirs("linux", home, func(string) string { return "" }))
	if want := []string{"chrome", "firefox", "chromium"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	mkdirs(t, filepath.Join(home, "xdg", "microsoft-edge"))
	got = installedBrowsers(allBrowsers, browserDirs("linux", home, func(k string) string {
		if k == "XDG_CONFIG_HOME" {
			return filepath.Join(home, "xdg")
		}
		return ""
	}))
	if want := []string{"edge", "firefox", "chromium"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("XDG_CONFIG_HOME ignored: got %v want %v", got, want)
	}
}

func TestSafariOnlyOnDarwin(t *testing.T) {
	for _, goos := range []string{"windows", "linux"} {
		if _, ok := browserDirs(goos, "/h", func(string) string { return "" })["safari"]; ok {
			t.Errorf("%s must not list safari", goos)
		}
	}
}
