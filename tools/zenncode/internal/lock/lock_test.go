package lock

import (
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock.json")

	f, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Dirty() {
		t.Error("a missing lock file should load clean")
	}
	if _, ok := f.URL("nope"); ok {
		t.Error("empty lock returned a URL")
	}

	f.Set("h1", "https://go.dev/play/p/aaa", "a.md")
	if !f.Dirty() {
		t.Error("Set should mark the file dirty")
	}
	if err := f.Save(path); err != nil {
		t.Fatal(err)
	}
	if f.Dirty() {
		t.Error("Save should clear dirty")
	}

	f2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if url, ok := f2.URL("h1"); !ok || url != "https://go.dev/play/p/aaa" {
		t.Errorf("URL(h1) = %q, %v", url, ok)
	}

	// Setting the same values again is not a change.
	f2.Set("h1", "https://go.dev/play/p/aaa", "a.md")
	if f2.Dirty() {
		t.Error("re-Setting identical values should not dirty the file")
	}
}

func TestPrune(t *testing.T) {
	f := &File{Entries: map[string]Entry{
		"live": {URL: "https://go.dev/play/p/aaa"},
		"dead": {URL: "https://go.dev/play/p/bbb"},
	}}
	// Prune drops everything not named, so a caller that has only looked at
	// part of the repository must not call it -- see cmdFix.
	if n := f.Prune(map[string]bool{"live": true}); n != 1 {
		t.Errorf("pruned %d, want 1", n)
	}
	if _, ok := f.URL("live"); !ok {
		t.Error("live entry was pruned")
	}
	if _, ok := f.URL("dead"); ok {
		t.Error("dead entry survived")
	}
}
