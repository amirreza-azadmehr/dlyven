//go:build !windows

package updater

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// The relaunch helper must hold off until the old process is gone, or the single-instance
// lock hands the launch straight back to the exiting app.
func TestWaitScriptWaitsForExit(t *testing.T) {
	old := exec.Command("sleep", "1")
	if err := old.Start(); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "launched")
	helper := exec.Command("sh", "-c", waitScript, "sh", strconv.Itoa(old.Process.Pid), "touch", marker)
	if err := helper.Start(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(400 * time.Millisecond)
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("helper launched while the old process was still running")
	}
	_ = old.Wait() // reap it, as launchd/init would for the real app
	if err := helper.Wait(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("helper never launched after the old process exited")
	}
}
