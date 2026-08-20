package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoWithFile lays out a file at rel and makes it the working directory's
// repository, the way the tool sees one.
func repoWithFile(t *testing.T, rel, content string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
}

func TestLoadFileSource(t *testing.T) {
	const code = "package main\n\nfunc main() {}\n"
	repoWithFile(t, "gosample/multi/main.go", code)

	src, err := loadFileSource("gosample/multi/main.go")
	if err != nil {
		t.Fatal(err)
	}
	if src.Code != code {
		t.Errorf("Code = %q", src.Code)
	}
	if src.Dir != filepath.FromSlash("gosample/multi") {
		t.Errorf("Dir = %q", src.Dir)
	}
	if !src.IsMain {
		t.Error("IsMain = false, want true for package main")
	}
}

func TestLoadFileSourceLibrary(t *testing.T) {
	repoWithFile(t, "gosample/greet/greet.go", "package greet\n\n// mainly a library\n")
	src, err := loadFileSource("gosample/greet/greet.go")
	if err != nil {
		t.Fatal(err)
	}
	if src.IsMain {
		t.Error("IsMain = true, want false: the file declares package greet")
	}
}

func TestLoadFileSourceErrors(t *testing.T) {
	repoWithFile(t, "gosample/main.go", "package main\n")
	for _, tt := range []struct {
		name string
		rel  string
		want string
	}{
		{"empty", "", "file= needs a path"},
		{"absolute", "/etc/hosts.go", "relative to the repository root"},
		{"not go", "gosample", "must name a .go file"},
		{"escapes the repository", "../elsewhere/main.go", "must stay inside the repository"},
		{"missing", "gosample/nosuch.go", "no such file"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadFileSource(tt.rel)
			if err == nil {
				t.Fatalf("expected an error mentioning %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to mention %q", err, tt.want)
			}
		})
	}
}

func TestFileSourceDrifted(t *testing.T) {
	src := fileSource{Code: "package main\n\nfunc main() {}\n"}
	for _, tt := range []struct {
		name  string
		block string
		want  bool
	}{
		{"identical", "package main\n\nfunc main() {}\n", false},
		{"no trailing newline", "package main\n\nfunc main() {}", false},
		{"extra trailing newlines", "package main\n\nfunc main() {}\n\n\n", false},
		{"changed", "package main\n\nfunc main() { println(1) }\n", true},
		{"truncated", "package main\n", true},
		{"leading blank line matters", "\npackage main\n\nfunc main() {}\n", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := src.drifted(tt.block); got != tt.want {
				t.Errorf("drifted = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileSourceFenceLines(t *testing.T) {
	src := fileSource{Code: "package main\n\nfunc main() {}\n"}
	got := src.fenceLines()
	want := []string{"package main", "", "func main() {}"}
	if len(got) != len(want) {
		t.Fatalf("fenceLines = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}
