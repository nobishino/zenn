package main

import (
	"strings"
	"testing"
	"time"

	"github.com/nobishino/zenn/tools/zenncode/internal/mdscan"
)

// block scans a fragment and returns its first Go block.
func block(t *testing.T, src string) mdscan.Block {
	t.Helper()
	for _, b := range mdscan.Scan("a.md", []byte(src)) {
		if b.Lang == "go" {
			return b
		}
	}
	t.Fatalf("no Go block in %q", src)
	return mdscan.Block{}
}

func TestParseExpectation(t *testing.T) {
	code := "```go\nvar x int\n```\n"
	out := "```\nhello\n```\n"

	for _, tt := range []struct {
		name string
		src  string
		want expectKind
	}{
		{"default", code, expectBuild},
		{"explicit build", "<!-- zenncode: expect=build -->\n" + code, expectBuild},
		{"explicit run", "<!-- zenncode: expect=run -->\n" + code, expectRun},
		{"compile error", "<!-- zenncode: expect=compile-error -->\n" + code, expectCompileError},
		{"error implies compile-error", `<!-- zenncode: error="cannot use" -->` + "\n" + code, expectCompileError},
		{"output implies run", "<!-- zenncode: output=next -->\n" + code + "\n" + out, expectRun},
		{"explicit panic", "<!-- zenncode: expect=panic -->\n" + code, expectPanic},
		{"panic implies panic", `<!-- zenncode: panic="index out of range" -->` + "\n" + code, expectPanic},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e, err := parseExpectation(block(t, tt.src))
			if err != nil {
				t.Fatal(err)
			}
			if e.kind != tt.want {
				t.Errorf("kind = %v, want %v", e.kind, tt.want)
			}
		})
	}
}

func TestParseExpectationOutputNext(t *testing.T) {
	src := "<!-- zenncode: output=next -->\n```go\nvar x int\n```\n\n```\nhello\nworld\n```\n"
	e, err := parseExpectation(block(t, src))
	if err != nil {
		t.Fatal(err)
	}
	if e.output != "hello\nworld\n" {
		t.Errorf("output = %q", e.output)
	}
	if e.timeout != defaultRunTimeout {
		t.Errorf("timeout = %v", e.timeout)
	}
}

func TestParseExpectationErrors(t *testing.T) {
	code := "```go\nvar x int\n```\n"
	for _, tt := range []struct {
		name string
		src  string
		want string
	}{
		{"bad expect", "<!-- zenncode: expect=nonsense -->\n" + code, "want build, run, compile-error or panic"},
		{"error without compile-error", `<!-- zenncode: expect=run error=x -->` + "\n" + code, "only applies with expect=compile-error"},
		{"panic without expect=panic", `<!-- zenncode: expect=run panic=x -->` + "\n" + code, "only applies with expect=panic"},
		{"bad panic regexp", `<!-- zenncode: expect=panic panic="[" -->` + "\n" + code, "error parsing regexp"},
		{"output without a next block", "<!-- zenncode: output=next -->\n" + code, "no block after this one"},
		{"output with compile-error", "<!-- zenncode: expect=compile-error output=next -->\n" + code, "only applies with expect=run"},
		{"bad regexp", `<!-- zenncode: expect=compile-error error="[" -->` + "\n" + code, "error parsing regexp"},
		{"bad timeout", "<!-- zenncode: expect=run timeout=soon -->\n" + code, "timeout=soon"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseExpectation(block(t, tt.src))
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to mention %q", err, tt.want)
			}
		})
	}
}

func TestParseExpectationTimeout(t *testing.T) {
	e, err := parseExpectation(block(t, "<!-- zenncode: expect=run timeout=2s -->\n```go\nvar x int\n```\n"))
	if err != nil {
		t.Fatal(err)
	}
	if e.timeout != 2*time.Second {
		t.Errorf("timeout = %v", e.timeout)
	}
}

func TestAborted(t *testing.T) {
	for _, tt := range []struct {
		name   string
		stderr string
		want   bool
	}{
		{"panic", "panic: runtime error: index out of range [1]\n\ngoroutine 1 [running]:\n", true},
		{"fatal error", "fatal error: all goroutines are asleep - deadlock!\n", true},
		{"panic after output on stderr", "log line\npanic: boom\n", true},
		{"clean", "", false},
		{"plain message", "flag provided but not defined: -x\n", false},
		{"the word panic in prose", "the program did not panic: good\n", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := aborted(tt.stderr); got != tt.want {
				t.Errorf("aborted(%q) = %v, want %v", tt.stderr, got, tt.want)
			}
		})
	}
}

func TestDiffOutput(t *testing.T) {
	for _, tt := range []struct {
		name       string
		want, got  string
		wantAgrees bool
	}{
		{"identical", "a\nb\n", "a\nb\n", true},
		{"trailing newlines ignored", "a\nb", "a\nb\n\n", true},
		{"trailing spaces ignored", "a  \nb\n", "a\nb\n", true},
		{"different line", "a\nb\n", "a\nc\n", false},
		{"missing line", "a\nb\n", "a\n", false},
		{"extra line", "a\n", "a\nb\n", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := diffOutput(tt.want, tt.got)
			if (d == "") != tt.wantAgrees {
				t.Errorf("diffOutput = %q, agreement expected: %v", d, tt.wantAgrees)
			}
		})
	}
}

func TestDiffOutputReport(t *testing.T) {
	d := diffOutput("a\nb\n", "a\nc\n")
	for _, want := range []string{"  a", "- b", "+ c"} {
		if !strings.Contains(d, want) {
			t.Errorf("report is missing %q:\n%s", want, d)
		}
	}
}
