// Package sysutil contains OS integration helpers (process control, file reveal,
// notifications, HTTP client construction).
package sysutil

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Command creates an exec.Cmd that never flashes a console window on Windows and gets a
// UTF-8 environment (important for Persian/Kurdish titles and paths).
func Command(name string, args ...string) *exec.Cmd {
	return CommandContext(context.Background(), name, args...)
}

// CommandContext is like Command but kills the whole process tree when ctx is canceled.
func CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	configureCmd(cmd)
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8", "PYTHONUTF8=1")
	cmd.WaitDelay = 3 * time.Second
	return cmd
}

// Kill terminates a started command and all its children (yt-dlp spawns ffmpeg).
func Kill(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return killTree(cmd)
}

// NewHTTPClient returns a client honouring the user's proxy (http, https, socks5).
func NewHTTPClient(proxyURL string, timeout time.Duration) *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if proxyURL != "" {
		if u, err := url.Parse(proxyURL); err == nil {
			tr.Proxy = http.ProxyURL(u)
		}
	} else {
		tr.Proxy = http.ProxyFromEnvironment
	}
	tr.ResponseHeaderTimeout = timeout
	return &http.Client{Transport: tr}
}

// OpenPath opens a file or folder with the OS default handler.
func OpenPath(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = Command("rundll32", "url.dll,FileProtocolHandler", path)
	case "darwin":
		cmd = Command("open", path)
	default:
		cmd = Command("xdg-open", path)
	}
	return cmd.Start()
}

// RevealInFolder opens the containing folder and selects the file when the OS supports it.
func RevealInFolder(path string) error {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return OpenPath(path)
	}
	if _, err := os.Stat(path); err != nil {
		return OpenPath(filepath.Dir(path))
	}
	switch runtime.GOOS {
	case "windows":
		return revealWindows(path)
	case "darwin":
		return Command("open", "-R", path).Start()
	default:
		uri := (&url.URL{Scheme: "file", Path: path}).String()
		err := Command("dbus-send", "--session", "--print-reply", "--dest=org.freedesktop.FileManager1",
			"--type=method_call", "/org/freedesktop/FileManager1", "org.freedesktop.FileManager1.ShowItems",
			"array:string:"+uri, "string:").Run()
		if err != nil {
			return OpenPath(filepath.Dir(path))
		}
		return nil
	}
}

// Notify shows a native desktop notification. Failures are ignored by callers on purpose:
// notifications are a nicety and must never break a download.
func Notify(title, body string) error {
	switch runtime.GOOS {
	case "windows":
		return notifyWindows(title, body)
	case "darwin":
		cmd := Command("osascript", "-e", `on run argv
display notification (item 2 of argv) with title (item 1 of argv)
end run`, title, body)
		return cmd.Start()
	default:
		return Command("notify-send", "--app-name", title, title, body).Start()
	}
}

// LookPathExtra searches PATH plus common install locations that GUI apps don't inherit
// on macOS (Homebrew) and Linux (snap/flatpak/local).
func LookPathExtra(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	if runtime.GOOS == "windows" {
		return ""
	}
	home, _ := os.UserHomeDir()
	for _, dir := range []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin", "/snap/bin", filepath.Join(home, ".local/bin")} {
		p := filepath.Join(dir, name)
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	return ""
}
