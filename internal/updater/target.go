package updater

import (
	"os"
	"path/filepath"
	"strings"
)

// Release asset names. They must match the matrix in .github/workflows/release.yml
// (TestAssetNamesMatchWorkflow checks).
const (
	AssetWindowsSetup    = "Dlyven-Windows-x64-Setup.exe"
	AssetWindowsPortable = "Dlyven-Windows-x64-Portable.exe"
	AssetMacZip          = "Dlyven-macOS-Universal.zip"
	AssetLinuxX64        = "Dlyven-Linux-x64"
	AssetLinuxARM64      = "Dlyven-Linux-ARM64"

	// SumsFile is the combined checksum file published with every release.
	SumsFile = "SHA256SUMS.txt"
)

// Mode says how this copy of Dlyven can be updated.
type Mode string

const (
	// ModeReplace swaps our own executable (or .app bundle on macOS) for the new one.
	ModeReplace Mode = "replace"
	// ModeInstaller runs the new Setup.exe silently (Windows copies installed with Setup).
	ModeInstaller Mode = "installer"
	// ModeManual means this copy can't replace itself; the user downloads the release.
	ModeManual Mode = "manual"
)

// target is what an update replaces and with which release asset.
type target struct {
	mode  Mode
	asset string
	// path is replaced by the update: the executable, or the .app bundle on macOS.
	path string
	// why explains ModeManual (logged, not shown).
	why string
}

func manual(path, why string) target { return target{mode: ModeManual, path: path, why: why} }

// detect works out how the executable at exe can be updated on goos/goarch.
func detect(exe, goos, goarch string) target {
	switch goos {
	case "darwin":
		app := bundleOf(exe)
		if app == "" {
			return manual(exe, "not inside an .app bundle")
		}
		// Gatekeeper runs quarantined apps from a read-only random path until they're moved.
		if strings.Contains(app, "/AppTranslocation/") {
			return manual(app, "app is translocated; move it to Applications")
		}
		if !writable(filepath.Dir(app)) {
			return manual(app, "app folder is not writable")
		}
		return target{mode: ModeReplace, asset: AssetMacZip, path: app}
	case "windows":
		if goarch != "amd64" {
			return manual(exe, "no release build for windows/"+goarch)
		}
		dir := filepath.Dir(exe)
		// Setup (NSIS) writes uninstall.exe next to the app; portable copies have none.
		if fileExists(filepath.Join(dir, "uninstall.exe")) {
			return target{mode: ModeInstaller, asset: AssetWindowsSetup, path: exe}
		}
		if !writable(dir) {
			return manual(exe, "app folder is not writable")
		}
		return target{mode: ModeReplace, asset: AssetWindowsPortable, path: exe}
	case "linux":
		asset := map[string]string{"amd64": AssetLinuxX64, "arm64": AssetLinuxARM64}[goarch]
		if asset == "" {
			return manual(exe, "no release build for linux/"+goarch)
		}
		if !writable(filepath.Dir(exe)) {
			return manual(exe, "app folder is not writable")
		}
		return target{mode: ModeReplace, asset: asset, path: exe}
	}
	return manual(exe, "no release build for "+goos)
}

// bundleOf returns the .app bundle containing exe (…/Dlyven.app/Contents/MacOS/Dlyven), or "".
func bundleOf(exe string) string {
	macos := filepath.Dir(exe)
	contents := filepath.Dir(macos)
	app := filepath.Dir(contents)
	if filepath.Base(macos) != "MacOS" || filepath.Base(contents) != "Contents" || !strings.HasSuffix(app, ".app") {
		return ""
	}
	return app
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

// writable reports whether files can be created in dir. Probing beats checking permission
// bits: ACLs, read-only mounts (snap, flatpak, dmg) and UAC all show up the same way.
func writable(dir string) bool {
	f, err := os.CreateTemp(dir, ".dlyven-write-test-*")
	if err != nil {
		return false
	}
	f.Close()
	_ = os.Remove(f.Name())
	return true
}
