package engine

import (
	"fmt"
	"runtime"
)

// archiveKind tells the installer how to unpack a download.
type archiveKind int

const (
	archiveRaw archiveKind = iota // the download is the binary itself
	archiveZip
	archiveTarXz
)

// source describes one downloadable artifact and where its SHA-256 checksum lives.
type source struct {
	URL string
	// SumsURL is fetched to verify the download. When SumsRelative is true it is a suffix
	// appended to the final (post-redirect) download URL instead of an absolute URL.
	SumsURL      string
	SumsRelative bool
	// SumName is the file name looked up inside a multi-entry checksum file ("" = first hash).
	SumName string
	Kind    archiveKind
	// Extract maps file base names inside the archive to installed names.
	Extract map[string]string
}

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func ytDlpAssetName() (string, error) {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "windows/amd64":
		return "yt-dlp.exe", nil
	case "windows/arm64":
		return "yt-dlp_arm64.exe", nil
	case "windows/386":
		return "yt-dlp_x86.exe", nil
	case "darwin/amd64", "darwin/arm64":
		return "yt-dlp_macos", nil
	case "linux/amd64":
		return "yt-dlp_linux", nil
	case "linux/arm64":
		return "yt-dlp_linux_aarch64", nil
	}
	return "", fmt.Errorf("no yt-dlp build for %s/%s", runtime.GOOS, runtime.GOARCH)
}

func ytDlpSource(tag string) (source, error) {
	name, err := ytDlpAssetName()
	if err != nil {
		return source{}, err
	}
	base := "https://github.com/yt-dlp/yt-dlp/releases/download/" + tag + "/"
	return source{
		URL:     base + name,
		SumsURL: base + "SHA2-256SUMS",
		SumName: name,
		Kind:    archiveRaw,
		Extract: map[string]string{name: exe("yt-dlp")},
	}, nil
}

// ffmpegSources returns one or more artifacts that together provide ffmpeg + ffprobe.
//
//	Windows / Linux: BtbN static GPL builds (checksums.sha256 in the same release)
//	macOS:           Martin Riedl's signed static builds (.sha256 next to each file)
func ffmpegSources() ([]source, error) {
	const btbn = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/"
	both := map[string]string{exe("ffmpeg"): exe("ffmpeg"), exe("ffprobe"): exe("ffprobe")}
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "windows/amd64", "windows/386":
		n := "ffmpeg-master-latest-win64-gpl.zip"
		return []source{{URL: btbn + n, SumsURL: btbn + "checksums.sha256", SumName: n, Kind: archiveZip, Extract: both}}, nil
	case "windows/arm64":
		n := "ffmpeg-master-latest-winarm64-gpl.zip"
		return []source{{URL: btbn + n, SumsURL: btbn + "checksums.sha256", SumName: n, Kind: archiveZip, Extract: both}}, nil
	case "linux/amd64":
		n := "ffmpeg-master-latest-linux64-gpl.tar.xz"
		return []source{{URL: btbn + n, SumsURL: btbn + "checksums.sha256", SumName: n, Kind: archiveTarXz, Extract: both}}, nil
	case "linux/arm64":
		n := "ffmpeg-master-latest-linuxarm64-gpl.tar.xz"
		return []source{{URL: btbn + n, SumsURL: btbn + "checksums.sha256", SumName: n, Kind: archiveTarXz, Extract: both}}, nil
	case "darwin/amd64", "darwin/arm64":
		arch := runtime.GOARCH
		var out []source
		for _, bin := range []string{"ffmpeg", "ffprobe"} {
			out = append(out, source{
				URL:          fmt.Sprintf("https://ffmpeg.martin-riedl.de/redirect/latest/macos/%s/release/%s.zip", arch, bin),
				SumsURL:      ".sha256",
				SumsRelative: true,
				Kind:         archiveZip,
				Extract:      map[string]string{bin: bin},
			})
		}
		return out, nil
	}
	return nil, fmt.Errorf("no ffmpeg build for %s/%s", runtime.GOOS, runtime.GOARCH)
}

// denoSource: yt-dlp needs a JavaScript runtime to solve YouTube's player challenges and
// list every format. Deno is the runtime recommended by the yt-dlp project.
func denoSource() (source, error) {
	var target string
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "windows/amd64", "windows/arm64":
		target = "x86_64-pc-windows-msvc"
	case "darwin/amd64":
		target = "x86_64-apple-darwin"
	case "darwin/arm64":
		target = "aarch64-apple-darwin"
	case "linux/amd64":
		target = "x86_64-unknown-linux-gnu"
	case "linux/arm64":
		target = "aarch64-unknown-linux-gnu"
	default:
		return source{}, fmt.Errorf("no deno build for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	n := "deno-" + target + ".zip"
	base := "https://github.com/denoland/deno/releases/latest/download/"
	return source{
		URL:     base + n,
		SumsURL: base + n + ".sha256sum",
		Kind:    archiveZip,
		Extract: map[string]string{exe("deno"): exe("deno")},
	}, nil
}
