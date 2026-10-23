// Package store persists download history in SQLite (modernc.org/sqlite — pure Go, no CGO).
package store

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Entry struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	Uploader   string `json:"uploader"`
	Platform   string `json:"platform"`
	Kind       string `json:"kind"`
	Quality    string `json:"quality"`
	Thumbnail  string `json:"thumbnail"`
	FilePath   string `json:"filePath"`
	Size       int64  `json:"size"`
	Status     string `json:"status"` // done | failed | canceled
	ErrorCode  string `json:"errorCode"`
	CreatedAt  int64  `json:"createdAt"`  // unix ms
	FinishedAt int64  `json:"finishedAt"` // unix ms
}

type Query struct {
	Search   string `json:"search"`
	Platform string `json:"platform"` // "" = all
	Status   string `json:"status"`   // "" = all
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

type Page struct {
	Items []Entry `json:"items"`
	Total int     `json:"total"`
}

type Store struct{ db *sql.DB }

const schema = `
CREATE TABLE IF NOT EXISTS history (
  id          TEXT PRIMARY KEY,
  url         TEXT NOT NULL,
  title       TEXT NOT NULL DEFAULT '',
  uploader    TEXT NOT NULL DEFAULT '',
  platform    TEXT NOT NULL DEFAULT '',
  kind        TEXT NOT NULL DEFAULT '',
  quality     TEXT NOT NULL DEFAULT '',
  thumbnail   TEXT NOT NULL DEFAULT '',
  file_path   TEXT NOT NULL DEFAULT '',
  size        INTEGER NOT NULL DEFAULT 0,
  status      TEXT NOT NULL DEFAULT '',
  error_code  TEXT NOT NULL DEFAULT '',
  created_at  INTEGER NOT NULL DEFAULT 0,
  finished_at INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_history_finished ON history(finished_at DESC);
CREATE INDEX IF NOT EXISTS idx_history_platform ON history(platform);
`

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", sqliteDSN(path))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite: serialize writers, avoids SQLITE_BUSY
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("history schema: %w", err)
	}
	return &Store{db: db}, nil
}

// sqliteDSN builds a file: URI so query pragmas are not treated as part of the filename
// (a trailing "?" is illegal on Windows and used to create a uniquely-named file on Unix).
func sqliteDSN(path string) string {
	abs := path
	if a, err := filepath.Abs(path); err == nil {
		abs = a
	}
	p := filepath.ToSlash(abs)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Path: p}
	q := u.Query()
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "synchronous(NORMAL)")
	q.Add("_pragma", "foreign_keys(1)")
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *Store) Close() error {
	_, _ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return s.db.Close()
}

func (s *Store) Upsert(ctx context.Context, e Entry) error {
	if e.FinishedAt == 0 {
		e.FinishedAt = time.Now().UnixMilli()
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO history (id,url,title,uploader,platform,kind,quality,thumbnail,file_path,size,status,error_code,created_at,finished_at)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  title=excluded.title, uploader=excluded.uploader, quality=excluded.quality, thumbnail=excluded.thumbnail,
  file_path=excluded.file_path, size=excluded.size, status=excluded.status, error_code=excluded.error_code,
  finished_at=excluded.finished_at`,
		e.ID, e.URL, e.Title, e.Uploader, e.Platform, e.Kind, e.Quality, e.Thumbnail, e.FilePath, e.Size, e.Status, e.ErrorCode, e.CreatedAt, e.FinishedAt)
	return err
}

func (s *Store) List(ctx context.Context, q Query) (Page, error) {
	if q.Limit <= 0 || q.Limit > 500 {
		q.Limit = 100
	}
	var where []string
	var args []any
	if t := strings.TrimSpace(q.Search); t != "" {
		like := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(t) + "%"
		where = append(where, `(title LIKE ? ESCAPE '\' OR uploader LIKE ? ESCAPE '\' OR url LIKE ? ESCAPE '\' OR file_path LIKE ? ESCAPE '\')`)
		args = append(args, like, like, like, like)
	}
	if q.Platform != "" {
		where = append(where, "platform = ?")
		args = append(args, q.Platform)
	}
	if q.Status != "" {
		where = append(where, "status = ?")
		args = append(args, q.Status)
	}
	cond := ""
	if len(where) > 0 {
		cond = " WHERE " + strings.Join(where, " AND ")
	}
	var page Page
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM history"+cond, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,url,title,uploader,platform,kind,quality,thumbnail,file_path,size,status,error_code,created_at,finished_at
FROM history`+cond+` ORDER BY finished_at DESC LIMIT ? OFFSET ?`, append(args, q.Limit, q.Offset)...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	page.Items = []Entry{}
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.URL, &e.Title, &e.Uploader, &e.Platform, &e.Kind, &e.Quality, &e.Thumbnail, &e.FilePath, &e.Size, &e.Status, &e.ErrorCode, &e.CreatedAt, &e.FinishedAt); err != nil {
			return page, err
		}
		page.Items = append(page.Items, e)
	}
	return page, rows.Err()
}

func (s *Store) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM history WHERE id = ?", id)
	return err
}

func (s *Store) Clear(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM history")
	return err
}

// Platforms returns the distinct platforms present in history (for the filter dropdown).
func (s *Store) Platforms(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT DISTINCT platform FROM history WHERE platform <> '' ORDER BY platform")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
