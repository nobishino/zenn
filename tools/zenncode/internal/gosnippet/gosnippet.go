// Package gosnippet turns the Go excerpts written in articles into complete,
// compilable programs.
//
// Articles show the interesting part of a program and leave out the
// boilerplate: the package clause and the imports are usually missing, and a
// snippet about a declaration often has no main function at all. Normalize
// puts that boilerplate back so the snippet can be compiled locally and, later,
// shared to the Go Playground as a program a reader can actually run.
package gosnippet

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/imports"
)

// Options controls normalization of a single snippet.
type Options struct {
	// Imports are package paths to add before goimports runs, for snippets
	// where the import is ambiguous (math/rand vs crypto/rand, and friends).
	Imports []string
	// Filename is the name goimports resolves imports relative to. It also
	// appears in error messages.
	Filename string
}

// Program is a normalized snippet.
type Program struct {
	Source    string // complete, gofmt-ed program
	Package   string
	HasMain   bool // the program has a func main, added or not
	AddedPkg  bool // a package clause was added
	AddedMain bool // an empty func main was added
}

// Hash identifies a program by content. It is the key under which a
// playground URL and a verification result are recorded. A program that could
// not be normalized has no source and hashes to the empty string.
func (p Program) Hash() string {
	if p.Source == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(p.Source))
	return hex.EncodeToString(sum[:])[:16]
}

var packageClauseRe = regexp.MustCompile(`(?m)^package\s+[A-Za-z_]`)

// Normalize returns code as a complete program.
//
// The returned error describes why a snippet could not be made into one --
// most often a syntax error, which for this repository is usually deliberate
// (an article demonstrating that some code does not compile) rather than a
// mistake.
func Normalize(code string, opts Options) (Program, error) {
	var p Program

	src := code
	var offset int // lines added above the snippet, for error positions
	if !packageClauseRe.MatchString(src) {
		src = "package main\n" + src
		p.AddedPkg = true
		offset = 1
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, opts.filename(), src, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return p, shiftErr(err, offset)
	}
	p.Package = file.Name.Name
	p.HasMain = hasMainFunc(file)

	if p.Package == "main" && !p.HasMain {
		src = strings.TrimRight(src, "\n") + "\n\nfunc main() {}\n"
		p.AddedMain = true
		p.HasMain = true
	}
	if len(opts.Imports) > 0 {
		src = addImports(src, opts.Imports)
	}

	out, err := imports.Process(opts.filename(), []byte(src), &imports.Options{
		Comments:  true,
		TabIndent: true,
		TabWidth:  8,
	})
	if err != nil {
		return p, shiftErr(err, offset)
	}
	p.Source = string(out)
	return p, nil
}

func (o Options) filename() string {
	if o.Filename == "" {
		return "snippet.go"
	}
	return o.Filename
}

func hasMainFunc(file *ast.File) bool {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "main" || fn.Recv != nil {
			continue
		}
		if fn.Type.Params.NumFields() == 0 && fn.Type.Results.NumFields() == 0 {
			return true
		}
	}
	return false
}

// addImports inserts an import block right below the package clause. goimports
// runs afterwards and merges it with whatever the snippet already imported.
func addImports(src string, paths []string) string {
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "package ") {
			continue
		}
		var b strings.Builder
		b.WriteString("\nimport (\n")
		for _, path := range paths {
			fmt.Fprintf(&b, "\t%s\n", strconv.Quote(path))
		}
		b.WriteString(")")
		rest := append([]string{b.String()}, lines[i+1:]...)
		return strings.Join(append(lines[:i+1:i+1], rest...), "\n")
	}
	return src
}

// maxProblems caps how many diagnostics a ParseError carries. A snippet that
// is a bare fragment produces one error per following line, and none of them
// say anything the first few do not.
const maxProblems = 6

// Problem is one diagnostic, positioned relative to the first line of the
// snippet as the article shows it.
type Problem struct {
	Line, Col int
	Msg       string
}

// ParseError reports that a snippet is not valid Go.
type ParseError struct {
	Problems []Problem
	Extra    int // problems dropped by maxProblems
}

func (e *ParseError) Error() string {
	msgs := make([]string, 0, len(e.Problems))
	for _, p := range e.Problems {
		msgs = append(msgs, fmt.Sprintf("%d:%d: %s", p.Line, p.Col, p.Msg))
	}
	if e.Extra > 0 {
		msgs = append(msgs, fmt.Sprintf("(and %d more)", e.Extra))
	}
	return strings.Join(msgs, "; ")
}

// shiftErr converts a parser error into a ParseError whose positions point at
// the snippet as the article shows it, not at the wrapped source.
func shiftErr(err error, offset int) error {
	var list scanner.ErrorList
	if !errors.As(err, &list) {
		return err
	}
	pe := &ParseError{}
	for _, e := range list {
		if len(pe.Problems) == maxProblems {
			pe.Extra = len(list) - maxProblems
			break
		}
		pe.Problems = append(pe.Problems, Problem{
			Line: e.Pos.Line - offset,
			Col:  e.Pos.Column,
			Msg:  e.Msg,
		})
	}
	return pe
}
