package engine

import (
	"os"
	"path/filepath"
	"runtime"
)

// browserDirs lists where each browser keeps its profile data, relative to the OS's usual
// roots. They mirror the locations yt-dlp's --cookies-from-browser searches, so "installed"
// here means "yt-dlp can look for cookies there".
func browserDirs(goos, home string, env func(string) string) map[string][]string {
	join := filepath.Join
	switch goos {
	case "windows":
		local, roaming := env("LOCALAPPDATA"), env("APPDATA")
		if local == "" {
			local = join(home, "AppData", "Local")
		}
		if roaming == "" {
			roaming = join(home, "AppData", "Roaming")
		}
		return map[string][]string{
			"chrome":   {join(local, "Google", "Chrome", "User Data")},
			"edge":     {join(local, "Microsoft", "Edge", "User Data")},
			"brave":    {join(local, "BraveSoftware", "Brave-Browser", "User Data")},
			"chromium": {join(local, "Chromium", "User Data")},
			"opera":    {join(roaming, "Opera Software", "Opera Stable")},
			"vivaldi":  {join(local, "Vivaldi", "User Data")},
			"firefox":  {join(roaming, "Mozilla", "Firefox", "Profiles")},
		}
	case "darwin":
		app := join(home, "Library", "Application Support")
		return map[string][]string{
			"chrome":   {join(app, "Google", "Chrome")},
			"edge":     {join(app, "Microsoft Edge")},
			"brave":    {join(app, "BraveSoftware", "Brave-Browser")},
			"chromium": {join(app, "Chromium")},
			"opera":    {join(app, "com.operasoftware.Opera")},
			"vivaldi":  {join(app, "Vivaldi")},
			"firefox":  {join(app, "Firefox", "Profiles")},
			"safari":   {join(home, "Library", "Containers", "com.apple.Safari", "Data", "Library", "Cookies"), join(home, "Library", "Cookies")},
		}
	default:
		cfg := env("XDG_CONFIG_HOME")
		if cfg == "" {
			cfg = join(home, ".config")
		}
		return map[string][]string{
			"chrome":   {join(cfg, "google-chrome")},
			"edge":     {join(cfg, "microsoft-edge")},
			"brave":    {join(cfg, "BraveSoftware", "Brave-Browser")},
			"chromium": {join(cfg, "chromium"), join(home, "snap", "chromium", "common", "chromium")},
			"opera":    {join(cfg, "opera")},
			"vivaldi":  {join(cfg, "vivaldi")},
			"firefox": {join(home, ".mozilla", "firefox"), join(home, "snap", "firefox", "common", ".mozilla", "firefox"),
				join(home, ".var", "app", "org.mozilla.firefox", ".mozilla", "firefox")},
		}
	}
}

// installedBrowsers returns the subset of order that has profile data on this machine.
func installedBrowsers(order []string, dirs map[string][]string) []string {
	var out []string
	for _, b := range order {
		for _, d := range dirs[b] {
			if fi, err := os.Stat(d); err == nil && fi.IsDir() {
				out = append(out, b)
				break
			}
		}
	}
	return out
}

// InstalledBrowsers reports which supported browsers have a profile on this machine, so the
// UI can offer only real choices and default to one that works.
func InstalledBrowsers(order []string) []string {
	home, _ := os.UserHomeDir()
	return installedBrowsers(order, browserDirs(runtime.GOOS, home, os.Getenv))
}
