//go:build windows

package sysutil

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows"
)

const createNoWindow = 0x08000000

func configureCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	cmd.Cancel = func() error { return killTree(cmd) }
}

func killTree(cmd *exec.Cmd) error {
	kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	kill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	if err := kill.Run(); err != nil {
		return cmd.Process.Kill()
	}
	return nil
}

func revealWindows(path string) error {
	cmd := exec.Command("explorer")
	// explorer.exe needs the raw form /select,"C:\path with spaces\file.mp4".
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf(`explorer /select,"%s"`, path)}
	return cmd.Start()
}

const toastScript = `[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null
$t = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$n = $t.GetElementsByTagName('text')
$n.Item(0).AppendChild($t.CreateTextNode($env:DLYVEN_TITLE)) > $null
$n.Item(1).AppendChild($t.CreateTextNode($env:DLYVEN_BODY)) > $null
$toast = [Windows.UI.Notifications.ToastNotification]::new($t)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe').Show($toast)`

func notifyWindows(title, body string) error {
	// Text is passed via environment variables so it can never be interpreted as script.
	cmd := Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", toastScript)
	cmd.Env = append(os.Environ(), "DLYVEN_TITLE="+title, "DLYVEN_BODY="+body)
	return cmd.Start()
}

// SupportsTranslucency reports whether Mica is available (Windows 11, build 22000+).
func SupportsTranslucency() bool {
	v := windows.RtlGetVersion()
	return v != nil && v.BuildNumber >= 22000
}
