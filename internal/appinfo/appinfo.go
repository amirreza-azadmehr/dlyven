// Package appinfo holds the product identity in one place so the app is easy to rename.
package appinfo

import (
	"os"
	"path/filepath"
)

const (
	// Name is the human-facing product name (window title, notifications, diagnostics).
	Name = "Dlyven"
	// ID is the machine-facing identifier (data directory, log file names).
	ID = "dlyven"
	// BundleID is the reverse-DNS identifier shared by Android, macOS, and Windows.
	// Hyphens from the GitHub username are stripped (invalid in those ID formats).
	BundleID = "io.github.amirrezaazadmehr.dlyven"

	// legacyName/ID are the previous product folder names. Used only for a one-time data migration.
	legacyName = "GrabIt"
	legacyID   = "grabit"
)

// Version is a var so release builds can stamp it:
//
//	wails build -ldflags "-X github.com/amirreza-azadmehr/dlyven/internal/appinfo.Version=1.2.3"
var Version = "1.0.2"

// DataDir returns (and creates) the per-user application data directory.
//
//	Windows: %AppData%\Dlyven
//	macOS:   ~/Library/Application Support/Dlyven
//	Linux:   ~/.config/Dlyven
//
// If a leftover GrabIt folder exists and the Dlyven folder does not, it is renamed once
// so settings, history, and engine binaries survive the product rename.
func DataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", err
		}
		dir := filepath.Join(home, "."+ID, Name)
		if err := migrateLegacyDir(filepath.Join(home, "."+legacyID, legacyName), dir); err != nil {
			return "", err
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
		return dir, nil
	}
	dir := filepath.Join(base, Name)
	if err := migrateLegacyDir(filepath.Join(base, legacyName), dir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func migrateLegacyDir(oldPath, newPath string) error {
	if oldPath == newPath {
		return nil
	}
	info, err := os.Stat(oldPath)
	if err != nil || !info.IsDir() {
		return nil
	}
	if _, err := os.Stat(newPath); err == nil {
		return nil
	}
	return os.Rename(oldPath, newPath)
}

// SubDir returns (and creates) a named directory inside DataDir.
func SubDir(name string) (string, error) {
	root, err := DataDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
