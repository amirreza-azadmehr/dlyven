package store

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHistory(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "h.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(s.Upsert(ctx, Entry{ID: "1", URL: "u1", Title: "آهنگ فارسی", Platform: "youtube", Status: "done", FinishedAt: 1}))
	must(s.Upsert(ctx, Entry{ID: "2", URL: "u2", Title: "100% Kurdish ڕۆژ", Platform: "instagram", Status: "failed", FinishedAt: 2}))
	must(s.Upsert(ctx, Entry{ID: "1", URL: "u1", Title: "آهنگ فارسی (updated)", Platform: "youtube", Status: "done", FinishedAt: 3}))

	p, err := s.List(ctx, Query{})
	must(err)
	if p.Total != 2 || p.Items[0].ID != "1" || p.Items[0].Title != "آهنگ فارسی (updated)" {
		t.Fatalf("list: %+v", p)
	}
	p, _ = s.List(ctx, Query{Search: "فارسی"})
	if p.Total != 1 {
		t.Errorf("unicode search: %+v", p)
	}
	p, _ = s.List(ctx, Query{Search: "100%"})
	if p.Total != 1 || p.Items[0].ID != "2" {
		t.Errorf("escaped %% search: %+v", p)
	}
	p, _ = s.List(ctx, Query{Platform: "instagram", Status: "failed"})
	if p.Total != 1 {
		t.Errorf("filter: %+v", p)
	}
	plats, _ := s.Platforms(ctx)
	if len(plats) != 2 {
		t.Errorf("platforms: %v", plats)
	}
	must(s.Delete(ctx, "2"))
	must(s.Clear(ctx))
	p, _ = s.List(ctx, Query{})
	if p.Total != 0 {
		t.Errorf("after clear: %+v", p)
	}
}

func TestHistorySurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "h.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := s.Upsert(ctx, Entry{ID: "keep", URL: "https://ex", Title: "saved", Status: "done", FinishedAt: 9}); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	for _, e := range mustReadDir(t, dir) {
		if strings.Contains(e, "?") {
			t.Fatalf("dsn leaked into filename: %s", e)
		}
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p, err := s.List(ctx, Query{})
	if err != nil {
		t.Fatal(err)
	}
	if p.Total != 1 || len(p.Items) != 1 || p.Items[0].ID != "keep" || p.Items[0].Title != "saved" {
		t.Fatalf("after reopen: %+v", p)
	}
}

func mustReadDir(t *testing.T, dir string) []string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(ents))
	for _, e := range ents {
		names = append(names, e.Name())
	}
	return names
}
