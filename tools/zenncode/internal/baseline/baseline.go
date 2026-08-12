// Package baseline records the snippets that are known not to verify.
//
// The repository has years of articles written against older toolchains, and
// many snippets fail to compile on purpose. Rather than annotate all of them
// up front, verify tolerates the failures listed here and rejects any new one.
// Entries are keyed by the hash of the normalized program, so editing a
// snippet drops it out of the baseline and it has to pass -- or be recorded
// again deliberately.
package baseline

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"
)

const comment = "Snippets that do not verify today. Regenerate with `zenncode baseline`. " +
	"Entries are matched by hash, so an edited snippet must pass on its own."

// Entry is one known-failing snippet.
type Entry struct {
	File   string `json:"file"`
	Line   int    `json:"line"`           // informational; not used for matching
	Hash   string `json:"hash,omitempty"` // gosnippet.Program.Hash, empty if normalization failed
	Raw    string `json:"raw,omitempty"`  // hash of the snippet as written, used when Hash is empty
	Stage  string `json:"stage"`          // "normalize" or "build"
	Reason string `json:"reason"`         // first line of the failure, for humans
}

func (e Entry) key() string { return e.File + "\x00" + e.Hash + "\x00" + e.Raw }

// File is the on-disk baseline.
type File struct {
	Comment   string  `json:"_comment"`
	Generated string  `json:"generated"`
	Entries   []Entry `json:"entries"`

	index map[string]Entry
}

// Load reads a baseline. A missing file is not an error; it loads as empty.
func Load(path string) (*File, error) {
	f := &File{index: map[string]Entry{}}
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
	for _, e := range f.Entries {
		f.index[e.key()] = e
	}
	return f, nil
}

// Has reports whether the given snippet is a known failure.
func (f *File) Has(e Entry) bool {
	_, ok := f.index[e.key()]
	return ok
}

// Save writes entries to path, sorted so the file stays diff-friendly.
func Save(path string, entries []Entry) error {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].File != entries[j].File {
			return entries[i].File < entries[j].File
		}
		return entries[i].Line < entries[j].Line
	})
	f := File{
		Comment:   comment,
		Generated: time.Now().Format(time.RFC3339),
		Entries:   entries,
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// Reason condenses a multi-line failure into a single line.
func Reason(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "(no output)"
	}
	line, _, _ := strings.Cut(s, "\n")
	if len(line) > 200 {
		line = line[:200] + "..."
	}
	return strings.TrimSpace(line)
}
