package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nobishino/zenn/tools/zenncode/internal/mdscan"
)

// expectKind is what an article claims about a snippet.
type expectKind int

const (
	// expectBuild is the default: the code compiles. Nothing is executed,
	// because plenty of the samples here deadlock or race on purpose.
	expectBuild expectKind = iota
	// expectRun additionally runs the program and requires it to exit
	// cleanly, optionally against the output the article shows.
	expectRun
	// expectCompileError inverts the check: the article is making a point
	// out of the code being rejected.
	expectCompileError
	// expectPanic is expectRun's counterpart for code that is supposed to
	// blow up at run time -- an article demonstrating that a comparison
	// panics wants the panic, not a clean exit.
	expectPanic
)

func (k expectKind) String() string {
	switch k {
	case expectRun:
		return "run"
	case expectCompileError:
		return "compile-error"
	case expectPanic:
		return "panic"
	default:
		return "build"
	}
}

// defaultRunTimeout is generous because verify runs many builds at once and a
// program can wait a while for a core. It only costs real time when a sample
// hangs, and a sample that hangs on purpose is better marked expect=build.
// Individual blocks can lower it with timeout=.
const defaultRunTimeout = 30 * time.Second

// expectation is the parsed form of the expect-related directive keys.
type expectation struct {
	kind    expectKind
	errRe   *regexp.Regexp // compile-error only: the message must match
	panicRe *regexp.Regexp // panic only: the message must match
	output  string         // expected stdout; empty means it is not checked
	timeout time.Duration
}

// runs reports whether the expectation requires executing the program.
func (e expectation) runs() bool { return e.kind == expectRun || e.kind == expectPanic }

// parseExpectation reads the directive on a block. The keys imply one
// another -- naming an expected message means a compile error is expected,
// and naming an expected output means the program is expected to run -- so
// the common cases need only one key.
func parseExpectation(b mdscan.Block) (expectation, error) {
	d := b.Directive
	e := expectation{timeout: defaultRunTimeout}

	switch v := d.Get("expect"); v {
	case "":
		switch {
		case d.Has("error"):
			e.kind = expectCompileError
		case d.Has("panic"):
			e.kind = expectPanic
		case d.Has("output"):
			e.kind = expectRun
		}
	case "build":
		e.kind = expectBuild
	case "run":
		e.kind = expectRun
	case "compile-error":
		e.kind = expectCompileError
	case "panic":
		e.kind = expectPanic
	default:
		return e, fmt.Errorf("expect=%s: want build, run, compile-error or panic", v)
	}

	if v := d.Get("error"); v != "" {
		if e.kind != expectCompileError {
			return e, fmt.Errorf("error= only applies with expect=compile-error")
		}
		re, err := regexp.Compile(v)
		if err != nil {
			return e, fmt.Errorf("error=%s: %w", v, err)
		}
		e.errRe = re
	}

	if v := d.Get("panic"); v != "" {
		if e.kind != expectPanic {
			return e, fmt.Errorf("panic= only applies with expect=panic")
		}
		re, err := regexp.Compile(v)
		if err != nil {
			return e, fmt.Errorf("panic=%s: %w", v, err)
		}
		e.panicRe = re
	}

	if d.Has("output") {
		if v := d.Get("output"); v != "next" {
			return e, fmt.Errorf("output=%s: the only supported value is next", v)
		}
		if e.kind != expectRun {
			return e, fmt.Errorf("output=next only applies with expect=run")
		}
		if b.NextCode == "" {
			return e, fmt.Errorf("output=next: there is no block after this one to compare against")
		}
		e.output = b.NextCode
	}

	if v := d.Get("timeout"); v != "" {
		td, err := time.ParseDuration(v)
		if err != nil {
			return e, fmt.Errorf("timeout=%s: %w", v, err)
		}
		if td <= 0 {
			return e, fmt.Errorf("timeout=%s: must be positive", v)
		}
		e.timeout = td
	}
	return e, nil
}

// aborted reports whether stderr is the runtime's report of a program dying on
// its own terms. "panic:" covers explicit and runtime panics; "fatal error:"
// covers what the runtime raises instead when no recovery is possible, such as
// an all-goroutine deadlock or a concurrent map write. Both are the kind of
// failure an article demonstrates on purpose, and both are distinct from a
// program that merely exits non-zero.
func aborted(stderr string) bool {
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(line, "panic: ") || strings.HasPrefix(line, "fatal error: ") {
			return true
		}
	}
	return false
}

// diffOutput reports how got differs from want, or "" if they agree.
// Trailing whitespace is ignored: it is invisible in an article and never the
// point of the example.
func diffOutput(want, got string) string {
	wantLines, gotLines := outputLines(want), outputLines(got)
	if len(wantLines) == len(gotLines) {
		same := true
		for i := range wantLines {
			if wantLines[i] != gotLines[i] {
				same = false
				break
			}
		}
		if same {
			return ""
		}
	}

	var b strings.Builder
	b.WriteString("the program's output is not what the article shows\n")
	for i := 0; i < max(len(wantLines), len(gotLines)); i++ {
		w, g := lineAt(wantLines, i), lineAt(gotLines, i)
		switch {
		case w == g:
			fmt.Fprintf(&b, "  %s\n", w)
		case g == "":
			fmt.Fprintf(&b, "- %s\n", w)
		case w == "":
			fmt.Fprintf(&b, "+ %s\n", g)
		default:
			fmt.Fprintf(&b, "- %s\n+ %s\n", w, g)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func outputLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func lineAt(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return ""
}
