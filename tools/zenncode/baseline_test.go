package main

import (
	"testing"

	"github.com/nobishino/zenn/tools/zenncode/internal/baseline"
)

func TestMergeBaseline(t *testing.T) {
	old := []baseline.Entry{
		{File: "articles/a.md", Line: 10, Hash: "a1"},
		{File: "articles/a.md", Line: 20, Hash: "a2"},
		{File: "articles/b.md", Line: 30, Hash: "b1"},
	}
	fresh := []baseline.Entry{{File: "articles/a.md", Line: 12, Hash: "a3"}}

	t.Run("a partial run keeps the files it did not scan", func(t *testing.T) {
		entries, kept := mergeBaseline(old, map[string]bool{"articles/a.md": true}, fresh)
		if kept != 1 {
			t.Errorf("kept = %d, want 1 (articles/b.md)", kept)
		}
		if len(entries) != 2 {
			t.Fatalf("entries = %+v, want b1 and a3", entries)
		}
		if entries[0].Hash != "b1" || entries[1].Hash != "a3" {
			t.Errorf("entries = %+v, want b1 carried over and a3 recorded", entries)
		}
	})

	t.Run("a full run replaces everything", func(t *testing.T) {
		scanned := map[string]bool{"articles/a.md": true, "articles/b.md": true}
		entries, kept := mergeBaseline(old, scanned, fresh)
		if kept != 0 {
			t.Errorf("kept = %d, want 0", kept)
		}
		if len(entries) != 1 || entries[0].Hash != "a3" {
			t.Errorf("entries = %+v, want only a3", entries)
		}
	})

	t.Run("a scanned file that now passes drops out", func(t *testing.T) {
		entries, _ := mergeBaseline(old, map[string]bool{"articles/b.md": true}, nil)
		for _, e := range entries {
			if e.File == "articles/b.md" {
				t.Errorf("%+v should have been dropped: it was scanned and did not fail", e)
			}
		}
		if len(entries) != 2 {
			t.Errorf("entries = %+v, want the two articles/a.md entries", entries)
		}
	})
}
