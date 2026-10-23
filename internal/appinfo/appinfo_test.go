package appinfo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateLegacyDirRenamesOnce(t *testing.T) {
	root := t.TempDir()
	oldPath := filepath.Join(root, "GrabIt")
	newPath := filepath.Join(root, "Dlyven")
	if err := os.MkdirAll(oldPath, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(oldPath, "settings.json")
	if err := os.WriteFile(marker, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := migrateLegacyDir(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("legacy dir still present: %v", err)
	}
	if _, err := os.Stat(filepath.Join(newPath, "settings.json")); err != nil {
		t.Fatal(err)
	}
	if err := migrateLegacyDir(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
}

func TestMigrateLegacyDirDoesNotClobber(t *testing.T) {
	root := t.TempDir()
	oldPath := filepath.Join(root, "GrabIt")
	newPath := filepath.Join(root, "Dlyven")
	if err := os.MkdirAll(oldPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(newPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldPath, "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newPath, "new.txt"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := migrateLegacyDir(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(oldPath, "old.txt")); err != nil {
		t.Fatal("expected legacy dir to remain when destination exists")
	}
	if _, err := os.Stat(filepath.Join(newPath, "new.txt")); err != nil {
		t.Fatal(err)
	}
}

func TestBundleID(t *testing.T) {
	if BundleID != "io.github.amirrezaazadmehr.dlyven" {
		t.Fatalf("BundleID = %q", BundleID)
	}
}
