package mdscan

import "testing"

func TestScan(t *testing.T) {
	src := `---
title: "sample"
---

# heading

https://go.dev/play/p/abc123XYZ

` + "```go" + `
func main() {}
` + "```" + `

text

<!-- zenncode: expect=compile-error imports="math/rand, fmt" skip=false -->
https://play.golang.org/p/def456
` + "```go:main.go" + `
var x int
` + "```" + `

` + "```mermaid" + `
graph TD;
` + "```" + `
`

	blocks := Scan("a.md", []byte(src))
	if len(blocks) != 3 {
		t.Fatalf("got %d blocks, want 3", len(blocks))
	}

	b := blocks[0]
	if b.Lang != "go" || b.Code != "func main() {}\n" {
		t.Errorf("block 0 = %q / %q", b.Lang, b.Code)
	}
	if b.OpenLine != 9 || b.CloseLine != 11 {
		t.Errorf("block 0 lines = %d..%d, want 9..11", b.OpenLine, b.CloseLine)
	}
	if b.PlayAbove.URL != "https://go.dev/play/p/abc123XYZ" || b.PlayAbove.Line != 7 {
		t.Errorf("block 0 play above = %+v", b.PlayAbove)
	}
	if b.PlayBelow.Found() {
		t.Errorf("block 0 play below = %+v", b.PlayBelow)
	}
	if b.Directive.Present() {
		t.Errorf("block 0 has an unexpected directive: %+v", b.Directive)
	}

	b = blocks[1]
	if b.Lang != "go" || b.Info != "go:main.go" {
		t.Errorf("block 1 lang/info = %q / %q", b.Lang, b.Info)
	}
	if got := b.Directive.Get("expect"); got != "compile-error" {
		t.Errorf("expect = %q", got)
	}
	if got := b.Directive.Get("imports"); got != "math/rand, fmt" {
		t.Errorf("imports = %q", got)
	}
	if b.Directive.Bool("skip", true) {
		t.Errorf("skip=false was read as true")
	}
	if b.PlayAbove.URL != "https://play.golang.org/p/def456" {
		t.Errorf("block 1 play above = %+v", b.PlayAbove)
	}

	if blocks[2].Lang != "mermaid" {
		t.Errorf("block 2 lang = %q", blocks[2].Lang)
	}
}

func TestScanMetaStopsAtProse(t *testing.T) {
	// A playground link belonging to an earlier paragraph must not attach
	// itself to this block.
	src := "https://go.dev/play/p/aaa\n\nsome prose\n\n```go\nvar x int\n```\n"
	blocks := Scan("a.md", []byte(src))
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks", len(blocks))
	}
	if blocks[0].PlayAbove.Found() {
		t.Errorf("PlayAbove = %+v, want none", blocks[0].PlayAbove)
	}
}

func TestScanNestedFence(t *testing.T) {
	// A go block inside a four-backtick fence is content, not a block.
	src := "````markdown\n```go\nvar x int\n```\n````\n\n```go\nvar y int\n```\n"
	blocks := Scan("a.md", []byte(src))
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	if blocks[0].Lang != "markdown" {
		t.Errorf("block 0 lang = %q", blocks[0].Lang)
	}
	if blocks[1].Code != "var y int\n" {
		t.Errorf("block 1 code = %q", blocks[1].Code)
	}
}

func TestParseLink(t *testing.T) {
	for _, tt := range []struct {
		in     string
		want   string
		inline bool
	}{
		{"https://go.dev/play/p/abc-1_2", "https://go.dev/play/p/abc-1_2", false},
		{"https://gotipplay.golang.org/p/abc/", "https://gotipplay.golang.org/p/abc", false},
		{"  https://go.dev/play/p/abc  ", "https://go.dev/play/p/abc", false},
		{"see https://go.dev/play/p/abc", "", false},
		{"https://go.dev/doc", "", false},
		{"[run](https://go.dev/play/p/abc)", "https://go.dev/play/p/abc", true},
		{"\U0001F449 [Go Playgroundで実行する](https://go.dev/play/p/abc)", "https://go.dev/play/p/abc", true},
	} {
		got, _ := ParseLink(tt.in)
		if got.URL != tt.want {
			t.Errorf("ParseLink(%q).URL = %q, want %q", tt.in, got.URL, tt.want)
		}
		if got.Inline() != tt.inline {
			t.Errorf("ParseLink(%q).Inline() = %v", tt.in, got.Inline())
		}
	}
}

func TestLinkRenderKeepsWrapper(t *testing.T) {
	l, ok := ParseLink("\U0001F449 [Go Playgroundで実行する](https://go.dev/play/p/old)")
	if !ok {
		t.Fatal("did not parse")
	}
	got := l.Render("https://go.dev/play/p/new")
	want := "\U0001F449 [Go Playgroundで実行する](https://go.dev/play/p/new)"
	if got != want {
		t.Errorf("Render = %q, want %q", got, want)
	}
}

func TestScanLinkBelowBlock(t *testing.T) {
	// Older articles put the link under the block. A link between two
	// blocks is recorded on both, and which one owns it is not mdscan's
	// call to make.
	src := "```go\nvar x int\n```\n\nhttps://go.dev/play/p/aaa\n\n```go\nvar y int\n```\n"
	blocks := Scan("a.md", []byte(src))
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks", len(blocks))
	}
	if got := blocks[0].PlayBelow.URL; got != "https://go.dev/play/p/aaa" {
		t.Errorf("block 0 below = %q", got)
	}
	if got := blocks[1].PlayAbove.URL; got != "https://go.dev/play/p/aaa" {
		t.Errorf("block 1 above = %q", got)
	}
	if blocks[0].PlayAbove.Found() || blocks[1].PlayBelow.Found() {
		t.Errorf("unexpected links on the far sides")
	}
}
