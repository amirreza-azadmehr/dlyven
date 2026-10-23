// Package logx provides structured logging to a size-rotated log file.
package logx

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// RotatingFile is an io.Writer that rotates when the file exceeds MaxBytes,
// keeping Keep old files (name.1.log ... name.N.log).
type RotatingFile struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	keep     int
	f        *os.File
	size     int64
}

func NewRotatingFile(path string, maxBytes int64, keep int) (*RotatingFile, error) {
	r := &RotatingFile{path: path, maxBytes: maxBytes, keep: keep}
	if err := r.open(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RotatingFile) open() error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(r.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	r.f, r.size = f, fi.Size()
	return nil
}

func (r *RotatingFile) rotated(i int) string {
	ext := filepath.Ext(r.path)
	return fmt.Sprintf("%s.%d%s", r.path[:len(r.path)-len(ext)], i, ext)
}

func (r *RotatingFile) rotate() error {
	r.f.Close()
	for i := r.keep - 1; i >= 1; i-- {
		_ = os.Rename(r.rotated(i), r.rotated(i+1))
	}
	_ = os.Rename(r.path, r.rotated(1))
	return r.open()
}

func (r *RotatingFile) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.size+int64(len(p)) > r.maxBytes {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := r.f.Write(p)
	r.size += int64(n)
	return n, err
}

func (r *RotatingFile) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.f.Close()
}

// Path returns the active log file path.
func (r *RotatingFile) Path() string { return r.path }

// New creates a JSON slog logger writing to dir/<name>.log (2 MB × 3 files) and,
// in development, to stderr as well.
func New(dir, name string, alsoStderr bool) (*slog.Logger, *RotatingFile, error) {
	rf, err := NewRotatingFile(filepath.Join(dir, name+".log"), 2<<20, 3)
	if err != nil {
		return slog.New(slog.NewTextHandler(os.Stderr, nil)), nil, err
	}
	var w io.Writer = rf
	if alsoStderr {
		w = io.MultiWriter(rf, os.Stderr)
	}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})), rf, nil
}

// Tail returns up to n last lines of a file (used by "Copy diagnostic info").
func Tail(path string, n int) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	ring := make([]string, 0, n)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	for sc.Scan() {
		if len(ring) == n {
			ring = ring[1:]
		}
		ring = append(ring, sc.Text())
	}
	return ring
}
