// Package lock records which playground link belongs to which snippet.
//
// The key is the hash of the normalized program, so the lock file answers the
// question verify needs to ask -- "is the link in the article still the link
// for this code?" -- without talking to the network. Only fix shares, and only
// for snippets whose hash is not recorded yet.
package lock

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"time"
)

const comment = "Playground links by snippet hash. Written by `zenncode fix`, read by `zenncode verify`. " +
	"The hash covers the normalized program, so editing a snippet retires its entry."

// Entry is one shared snippet.
type Entry struct {
	URL      string `json:"url"`
	File     string `json:"file"`     // where it was found when shared, for humans
	SharedAt string `json:"sharedAt"` // RFC 3339 date
}

// File is the on-disk lock.
type File struct {
	Comment string           `json:"_comment"`
	Entries map[string]Entry `json:"entries"`

	dirty bool
}

// Load reads a lock file. A missing file is not an error.
func Load(path string) (*File, error) {
	f := &File{Entries: map[string]Entry{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return f, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, f); err != nil {
		return nil, err
	}
	if f.Entries == nil {
		f.Entries = map[string]Entry{}
	}
	return f, nil
}

// URL returns the recorded link for hash.
func (f *File) URL(hash string) (string, bool) {
	e, ok := f.Entries[hash]
	return e.URL, ok
}

// Set records a link, marking the file as needing a save.
func (f *File) Set(hash, url, file string) {
	if e, ok := f.Entries[hash]; ok && e.URL == url && e.File == file {
		return
	}
	f.Entries[hash] = Entry{
		URL:      url,
		File:     file,
		SharedAt: time.Now().Format(time.DateOnly),
	}
	f.dirty = true
}

// Prune drops entries whose hash is no longer used by any snippet, so the file
// does not grow forever as articles are edited.
func (f *File) Prune(live map[string]bool) int {
	var n int
	for hash := range f.Entries {
		if !live[hash] {
			delete(f.Entries, hash)
			f.dirty = true
			n++
		}
	}
	return n
}

// Dirty reports whether Save would change anything.
func (f *File) Dirty() bool { return f.dirty }

// Save writes the lock file. Go marshals maps with sorted keys, so the output
// is stable across runs.
func (f *File) Save(path string) error {
	f.Comment = comment
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return err
	}
	f.dirty = false
	return nil
}
