package main

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/nobishino/zenn/tools/zenncode/internal/check"
	"github.com/nobishino/zenn/tools/zenncode/internal/mdscan"
)

// fileSource is a .go file in the repository that a block mirrors, named with
// the `file=` directive.
//
// Most snippets in these articles are excerpts that the tool restores into a
// program, which only works for code that fits in one file and builds against
// the standard library. A sample that needs more than that -- several files, a
// build tag, a test -- lives in the repository as real Go code, and the block
// shows a copy of it. The file is the original: `fix` copies it into the
// article, `verify` reports when the two have drifted apart.
type fileSource struct {
	Rel    string // path as written in the directive, relative to the repo root
	Dir    string // the package directory, relative to the repo root
	Code   string // the file's contents
	IsMain bool   // the file declares package main
}

// loadFileSource reads the file a block is mirrored from. Paths are written
// relative to the repository root, which is the working directory by the time
// this runs.
func loadFileSource(rel string) (fileSource, error) {
	switch {
	case rel == "":
		return fileSource{}, fmt.Errorf("file= needs a path, as in file=gosample/main.go")
	case filepath.IsAbs(rel):
		return fileSource{}, fmt.Errorf("file=%s: use a path relative to the repository root", rel)
	case !strings.HasSuffix(rel, ".go"):
		return fileSource{}, fmt.Errorf("file=%s: must name a .go file", rel)
	}
	clean := filepath.Clean(rel)
	if strings.HasPrefix(clean, "..") {
		return fileSource{}, fmt.Errorf("file=%s: must stay inside the repository", rel)
	}
	data, err := os.ReadFile(clean)
	if err != nil {
		return fileSource{}, fmt.Errorf("file=%s: %w", rel, err)
	}
	code := string(data)
	return fileSource{
		Rel:    clean,
		Dir:    filepath.Dir(clean),
		Code:   code,
		IsMain: isCommand(clean, data),
	}, nil
}

// isCommand reports whether the file declares package main. The package
// clause is parsed rather than matched as text, so a file carrying an import
// comment (`package main // import "example/cmd"`) is still recognized as a
// command; getting this wrong would leave expect=run with no binary to run.
// A file that does not parse is nobody's command.
func isCommand(path string, src []byte) bool {
	f, err := parser.ParseFile(token.NewFileSet(), path, src, parser.PackageClauseOnly)
	return err == nil && f.Name.Name == "main"
}

// drifted reports whether the block shows something other than the file.
// Trailing blank lines are ignored: a fence and a file end differently and
// neither is visible to a reader.
func (f fileSource) drifted(blockCode string) bool {
	return strings.TrimRight(blockCode, "\n") != strings.TrimRight(f.Code, "\n")
}

// fenceLines is the file's content as the lines that belong between the
// fences.
func (f fileSource) fenceLines() []string {
	return strings.Split(strings.TrimRight(f.Code, "\n"), "\n")
}

// processFileBlock verifies a block whose code lives in the repository.
//
// The package is compiled where it is, so build tags, sibling files and tests
// work as they do for the author. The article's copy is checked last: a stale
// copy is worth reporting, but a package that does not build is the bigger
// news, and `fix` would otherwise paste broken code into the article.
func processFileBlock(builder *check.Builder, b mdscan.Block, exp expectation, r result) result {
	src, err := loadFileSource(b.Directive.Get("file"))
	if err != nil {
		return r.fail("directive", err.Error())
	}
	r.src = &src

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()
	res, err := builder.Package(ctx, check.PackageRequest{
		Dir:    src.Dir,
		IsMain: src.IsMain,
		Test:   exp.kind == expectTest,
		Binary: exp.runs(),
		GOOS:   b.Directive.Get("goos"),
		GOARCH: b.Directive.Get("goarch"),
	})
	stage := "build"
	if exp.kind == expectTest {
		stage = "test"
	}
	if err != nil {
		return r.fail(stage, fmt.Sprintf("%v\n%s", err, res.Output))
	}
	if !res.OK {
		return r.fail(stage, res.Output)
	}

	if exp.runs() {
		if r = checkRun(ctx, r, exp, res.Bin); r.status != statusOK {
			return r
		}
	}

	if src.drifted(b.Code) {
		return r.fail("file", fmt.Sprintf("this block no longer matches %s; run `make fix`", src.Rel))
	}
	r.status = statusOK
	return r
}
