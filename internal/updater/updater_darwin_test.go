package updater

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amirreza-azadmehr/dlyven/internal/errs"
)

// makeBundle builds a minimal ad-hoc signed Dlyven.app in dir. version lands in Info.plist
// so the test can tell which bundle ended up installed.
func makeBundle(t *testing.T, dir, id, version string) string {
	t.Helper()
	app := filepath.Join(dir, "Dlyven.app")
	macos := filepath.Join(app, "Contents", "MacOS")
	if err := os.MkdirAll(macos, 0o755); err != nil {
		t.Fatal(err)
	}
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleExecutable</key><string>Dlyven</string>
<key>CFBundleIdentifier</key><string>%s</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleShortVersionString</key><string>%s</string>
</dict></plist>
`, id, version)
	if err := os.WriteFile(filepath.Join(app, "Contents", "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatal(err)
	}
	bin, err := os.ReadFile("/usr/bin/true")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(macos, "Dlyven"), bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("codesign", "--force", "--sign", "-", app).CombinedOutput(); err != nil {
		t.Fatalf("codesign: %v %s", err, out)
	}
	return app
}

func zipBundle(t *testing.T, app string) []byte {
	t.Helper()
	zip := filepath.Join(t.TempDir(), AssetMacZip)
	if out, err := exec.Command("ditto", "-c", "-k", "--keepParent", app, zip).CombinedOutput(); err != nil {
		t.Fatalf("ditto: %v %s", err, out)
	}
	b, err := os.ReadFile(zip)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func installedVersion(t *testing.T, app string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(app, "Contents", "Info.plist"))
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []string{"1.0.0", "2.0.0"} {
		if strings.Contains(string(b), "<string>"+v+"</string>") {
			return v
		}
	}
	return ""
}

func TestMacBundleUpdate(t *testing.T) {
	const id = "io.github.amirrezaazadmehr.dlyven"
	apps := t.TempDir() // stands in for /Applications
	current := makeBundle(t, apps, id, "1.0.0")
	srv := fakeGitHub(t, "v2.0.0", AssetMacZip, zipBundle(t, makeBundle(t, t.TempDir(), id, "2.0.0")), "")
	m, _ := newTestManager(t, srv, "1.0.0", target{mode: ModeReplace, asset: AssetMacZip, path: current})

	st, err := m.Download(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !st.Ready {
		t.Fatalf("not staged: %+v", st)
	}
	if v := installedVersion(t, current); v != "1.0.0" {
		t.Fatalf("download replaced the running app early: %s", v)
	}
	if err := m.ApplyStaged(); err != nil {
		t.Fatal(err)
	}
	if v := installedVersion(t, current); v != "2.0.0" {
		t.Fatalf("installed version %q after apply", v)
	}
	if out, err := exec.Command("codesign", "--verify", "--deep", "--strict", current).CombinedOutput(); err != nil {
		t.Fatalf("swapped bundle fails codesign: %s", out)
	}
	m.Cleanup()
	if left, _ := filepath.Glob(filepath.Join(apps, stagePrefix+"*")); len(left) != 0 {
		t.Errorf("cleanup left %v", left)
	}
}

func TestMacBundleFromOtherSignerIsRejected(t *testing.T) {
	apps := t.TempDir()
	current := makeBundle(t, apps, "io.github.amirrezaazadmehr.dlyven", "1.0.0")
	srv := fakeGitHub(t, "v2.0.0", AssetMacZip, zipBundle(t, makeBundle(t, t.TempDir(), "com.example.impostor", "2.0.0")), "")
	m, _ := newTestManager(t, srv, "1.0.0", target{mode: ModeReplace, asset: AssetMacZip, path: current})

	_, err := m.Download(context.Background())
	if errs.CodeOf(err) != errs.ErrChecksum || !strings.Contains(errs.DetailOf(err), "com.example.impostor") {
		t.Fatalf("want ERR_CHECKSUM naming the other signer, got %v", err)
	}
	if v := installedVersion(t, current); v != "1.0.0" {
		t.Fatalf("app changed: %s", v)
	}
	if left, _ := filepath.Glob(filepath.Join(apps, stagePrefix+"*")); len(left) != 0 {
		t.Errorf("rejected update left %v", left)
	}
}
