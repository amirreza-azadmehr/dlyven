//go:build !windows

package sysutil

import (
	"errors"
	"os/exec"
	"runtime"
	"syscall"
)

func configureCmd(cmd *exec.Cmd) {
	// Own process group so the whole tree (yt-dlp + ffmpeg) can be signalled at once.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return killTree(cmd) }
}

func killTree(cmd *exec.Cmd) error {
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
		return cmd.Process.Kill()
	}
	return nil
}

func revealWindows(string) error { return errors.New("not windows") }

func notifyWindows(string, string) error { return errors.New("not windows") }

// SupportsTranslucency: vibrancy works on macOS; on Linux it depends on the compositor,
// so we use the solid fallback there.
func SupportsTranslucency() bool { return runtime.GOOS == "darwin" }
