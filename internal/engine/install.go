package engine

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ulikunitz/xz"

	"github.com/amirreza-azadmehr/dlyven/internal/errs"
)

func (m *Manager) installSource(ctx context.Context, comp string, src source) error {
	tmpDir, err := os.MkdirTemp(m.binDir, ".install-"+comp+"-")
	if err != nil {
		return errs.Wrap(errs.ErrDisk, err)
	}
	defer os.RemoveAll(tmpDir)

	archive := filepath.Join(tmpDir, "download")
	finalURL, sum, err := m.download(ctx, comp, src.URL, archive)
	if err != nil {
		return err
	}

	m.progress(Progress{Component: comp, Stage: "verifying", Percent: 100})
	sumsURL := src.SumsURL
	if src.SumsRelative {
		sumsURL = finalURL + src.SumsURL
	}
	want, err := m.fetchChecksum(ctx, sumsURL, src.SumName)
	if err != nil {
		return err
	}
	if !strings.EqualFold(want, sum) {
		return errs.New(errs.ErrChecksum, fmt.Sprintf("%s: expected %s, got %s", path.Base(src.URL), want, sum))
	}

	m.progress(Progress{Component: comp, Stage: "extracting", Percent: 100})
	staged := map[string]string{} // installed name -> staged path
	switch src.Kind {
	case archiveRaw:
		for _, installed := range src.Extract {
			staged[installed] = archive
		}
	case archiveZip:
		if err := extractZip(archive, tmpDir, src.Extract, staged); err != nil {
			return err
		}
	case archiveTarXz:
		if err := extractTarXz(archive, tmpDir, src.Extract, staged); err != nil {
			return err
		}
	}
	if len(staged) != len(src.Extract) {
		return errs.New(errs.ErrEngineInstall, fmt.Sprintf("archive %s did not contain %v", path.Base(src.URL), keys(src.Extract)))
	}
	for installed, p := range staged {
		if err := os.Chmod(p, 0o755); err != nil {
			return errs.Wrap(errs.ErrDisk, err)
		}
		dst := filepath.Join(m.binDir, installed)
		if err := replaceFile(p, dst); err != nil {
			return errs.Wrap(errs.ErrDisk, err)
		}
		clearQuarantine(dst)
	}
	return nil
}

// replaceFile moves src over dst. On Windows a running .exe can't be overwritten but can
// be renamed, so the old binary is moved aside first.
func replaceFile(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		old := dst + ".old"
		_ = os.Remove(old)
		if err := os.Rename(dst, old); err != nil {
			return err
		}
		defer os.Remove(old)
	}
	return os.Rename(src, dst)
}

type progressWriter struct {
	total, done int64
	last        time.Time
	report      func(done, total int64)
}

func (w *progressWriter) Write(p []byte) (int, error) {
	w.done += int64(len(p))
	if time.Since(w.last) > 150*time.Millisecond {
		w.last = time.Now()
		w.report(w.done, w.total)
	}
	return len(p), nil
}

// download streams url to dest, hashing as it goes. Returns the final URL after redirects.
func (m *Manager) download(ctx context.Context, comp, url, dest string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Dlyven-engine-installer")
	resp, err := m.httpClient().Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", "", errs.New(errs.ErrCanceled, "")
		}
		return "", "", errs.Wrap(errs.ErrNetwork, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", errs.New(errs.ErrNetwork, fmt.Sprintf("GET %s: HTTP %d", url, resp.StatusCode))
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", "", errs.Wrap(errs.ErrDisk, err)
	}
	defer f.Close()

	h := sha256.New()
	pw := &progressWriter{total: resp.ContentLength, report: func(done, total int64) {
		pct := 0.0
		if total > 0 {
			pct = float64(done) / float64(total) * 100
		}
		m.progress(Progress{Component: comp, Stage: "downloading", Percent: pct, Downloaded: done, Total: total})
	}}
	if _, err := io.Copy(io.MultiWriter(f, h, pw), resp.Body); err != nil {
		if ctx.Err() != nil {
			return "", "", errs.New(errs.ErrCanceled, "")
		}
		return "", "", errs.Wrap(errs.ErrNetwork, err)
	}
	pw.report(pw.done, pw.done)
	if err := f.Close(); err != nil {
		return "", "", errs.Wrap(errs.ErrDisk, err)
	}
	return resp.Request.URL.String(), hex.EncodeToString(h.Sum(nil)), nil
}

var hexHashRe = regexp.MustCompile(`(?i)\b[0-9a-f]{64}\b`)

// fetchChecksum downloads a checksum file and returns the SHA-256 for name.
// Supported formats: "hash  name" lines (sha256sum), a single bare hash, and
// PowerShell's Get-FileHash output (Deno's Windows checksums).
func (m *Manager) fetchChecksum(ctx context.Context, url, name string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := m.httpClient().Do(req)
	if err != nil {
		return "", errs.Wrap(errs.ErrNetwork, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errs.New(errs.ErrChecksum, fmt.Sprintf("checksum file %s: HTTP %d", url, resp.StatusCode))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", errs.Wrap(errs.ErrNetwork, err)
	}
	if sum := ParseChecksum(string(body), name); sum != "" {
		return sum, nil
	}
	return "", errs.New(errs.ErrChecksum, "no checksum for "+name+" in "+url)
}

// ParseChecksum extracts the hash for name from checksum-file content.
func ParseChecksum(content, name string) string {
	if name != "" {
		sc := bufio.NewScanner(strings.NewReader(content))
		for sc.Scan() {
			f := strings.Fields(sc.Text())
			if len(f) >= 2 && hexHashRe.MatchString(f[0]) && strings.TrimPrefix(f[len(f)-1], "*") == name {
				return strings.ToLower(f[0])
			}
		}
	}
	all := hexHashRe.FindAllString(content, -1)
	if len(all) == 1 {
		return strings.ToLower(all[0])
	}
	return ""
}

func extractZip(archive, dir string, want map[string]string, staged map[string]string) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return errs.Wrap(errs.ErrEngineInstall, err)
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		installed, ok := want[path.Base(f.Name)]
		if !ok {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return errs.Wrap(errs.ErrEngineInstall, err)
		}
		out := filepath.Join(dir, "x-"+installed)
		err = writeFile(out, rc)
		rc.Close()
		if err != nil {
			return err
		}
		staged[installed] = out
	}
	return nil
}

func extractTarXz(archive, dir string, want map[string]string, staged map[string]string) error {
	f, err := os.Open(archive)
	if err != nil {
		return errs.Wrap(errs.ErrEngineInstall, err)
	}
	defer f.Close()
	xr, err := xz.NewReader(bufio.NewReader(f))
	if err != nil {
		return errs.Wrap(errs.ErrEngineInstall, err)
	}
	tr := tar.NewReader(xr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errs.Wrap(errs.ErrEngineInstall, err)
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		installed, ok := want[path.Base(h.Name)]
		if !ok {
			continue
		}
		out := filepath.Join(dir, "x-"+installed)
		if err := writeFile(out, tr); err != nil {
			return err
		}
		staged[installed] = out
	}
	return nil
}

func writeFile(p string, r io.Reader) error {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return errs.Wrap(errs.ErrDisk, err)
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return errs.Wrap(errs.ErrDisk, err)
	}
	return f.Close()
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
