package engine

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/amirreza-azadmehr/dlyven/internal/settings"
)

func TestParseChecksum(t *testing.T) {
	sums := `b7af77f686e95f7845ce8d09d21957718febb8b48b30555adb9dd2292ad7bf14  yt-dlp
33121d192d706e62f684a06f80918e3f3a8ba52ac3a65213b6951c39c7d9f555  yt-dlp.exe
d8bd33fca9d5b23bd5b9952b3fca36db5c46331933fda9f598a3da14ce2c243f *yt-dlp_linux`
	if got := ParseChecksum(sums, "yt-dlp.exe"); got != "33121d192d706e62f684a06f80918e3f3a8ba52ac3a65213b6951c39c7d9f555" {
		t.Errorf("got %q", got)
	}
	if got := ParseChecksum(sums, "yt-dlp_linux"); got != "d8bd33fca9d5b23bd5b9952b3fca36db5c46331933fda9f598a3da14ce2c243f" {
		t.Errorf("binary-mode marker: got %q", got)
	}
	if got := ParseChecksum(sums, "missing"); got != "" {
		t.Errorf("ambiguous file must not match: %q", got)
	}

	ps := "\r\nAlgorithm : SHA256\r\nHash      : CBA84B26E0DF5AE7F74FC66B86A95B866872A86292EBD84FA1E959927EDD93BF\r\nPath      : C:\\a\\deno.zip\r\n"
	if got := ParseChecksum(ps, ""); got != "cba84b26e0df5ae7f74fc66b86a95b866872a86292ebd84fa1e959927edd93bf" {
		t.Errorf("powershell format: got %q", got)
	}
	single := "c8ed4c4e6978a03c485edbfe4e0a5dc2380f8a30bba5150531b31b094492d924  ffmpeg.zip\n"
	if got := ParseChecksum(single, ""); got != "c8ed4c4e6978a03c485edbfe4e0a5dc2380f8a30bba5150531b31b094492d924" {
		t.Errorf("single: got %q", got)
	}
}

func TestParseVersion(t *testing.T) {
	if v := parseVersion(CompFFmpeg, "ffmpeg version N-117000-gabc Copyright (c) 2000-2025\nbuilt with gcc"); v != "N-117000-gabc" {
		t.Error(v)
	}
	if v := parseVersion(CompYtDlp, "2026.10.21\n"); v != "2026.10.21" {
		t.Error(v)
	}
	if v := parseVersion(CompDeno, "deno 2.1.4 (stable, release, x86_64-unknown-linux-gnu)\nv8 13"); v != "2.1.4" {
		t.Error(v)
	}
}

func TestStatusDoesNotExecVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script stub")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "yt-dlp")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := New(dir, dir, func() settings.Settings { return settings.Settings{} }, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	start := time.Now()
	st := m.Status(context.Background())
	if time.Since(start) > time.Second {
		t.Fatalf("Status launched the binary (%s)", time.Since(start))
	}
	if !st.YtDlp.Installed {
		t.Fatal("expected yt-dlp present")
	}
	if st.YtDlp.Version != "" {
		t.Fatalf("version should stay lazy, got %q", st.YtDlp.Version)
	}
}

func TestImpersonateArgsDoesNotBlock(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script stub")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "yt-dlp")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := New(dir, dir, func() settings.Settings { return settings.Settings{} }, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	start := time.Now()
	args := m.impersonateArgs(context.Background())
	if time.Since(start) > time.Second {
		t.Fatalf("impersonateArgs probed synchronously (%s)", time.Since(start))
	}
	if len(args) != 2 || args[0] != "--impersonate" || args[1] != "chrome" {
		t.Fatalf("got %v", args)
	}
}
