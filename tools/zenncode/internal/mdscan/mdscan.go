// Package mdscan extracts fenced code blocks and their surrounding metadata
// from Zenn article Markdown.
//
// A block's metadata sits in its own paragraph: a zenncode directive written
// as an HTML comment above the opening fence (invisible in the rendered
// article), and a Go Playground URL on a line of its own, which is how Zenn
// renders a playground link card. The articles here place that URL above the
// block in some cases and below it in others, so both sides are scanned.
package mdscan

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Block is one fenced code block.
type Block struct {
	File      string // path as given to Scan
	Lang      string // info string up to the first space or colon, e.g. "go"
	Info      string // the whole info string, e.g. "go:main.go"
	OpenLine  int    // 1-indexed line of the opening fence
	CloseLine int    // 1-indexed line of the closing fence
	Code      string // fence contents; ends with a newline unless empty

	// NextLang and NextCode describe the fenced block that follows this one
	// in the document, which is where an article shows a program's output.
	NextLang string
	NextCode string

	Directive Directive

	// Articles in this repository put the playground link on either side of
	// the block, depending on when they were written, so both are recorded
	// and the ambiguous case -- a link sitting between two blocks -- shows
	// up as PlayBelow of one and PlayAbove of the next.
	PlayAbove Link
	PlayBelow Link
}

// Link is a Go Playground URL on a line of its own, written either bare (Zenn
// renders it as a link card) or wrapped in Markdown link syntax, which the
// book uses.
type Link struct {
	URL  string
	Line int // 1-indexed, 0 if there is no link

	// Pre and Post are the text around the URL, so that rewriting a link
	// keeps whatever the author wrote around it.
	Pre, Post string
}

// Found reports whether a link is present.
func (l Link) Found() bool { return l.Line != 0 }

// Inline reports whether the link is wrapped in Markdown syntax rather than
// written bare.
func (l Link) Inline() bool { return l.Pre != "" || l.Post != "" }

// Render writes url in this link's style.
func (l Link) Render(url string) string { return l.Pre + url + l.Post }

// Pos renders the block's location as file:line.
func (b Block) Pos() string {
	return fmt.Sprintf("%s:%d", b.File, b.OpenLine)
}

// Directive is a `<!-- zenncode: ... -->` comment attached to a block.
// The zero value means no directive was present.
type Directive struct {
	Line int // 1-indexed line, 0 if absent
	Raw  string
	Opts map[string]string // bare keys map to ""
}

// Present reports whether a directive was found.
func (d Directive) Present() bool { return d.Line != 0 }

// Has reports whether key was given, with or without a value.
func (d Directive) Has(key string) bool {
	_, ok := d.Opts[key]
	return ok
}

// Get returns the value for key, or "" if absent.
func (d Directive) Get(key string) string { return d.Opts[key] }

// Bool returns the value of a key written either bare ("run") or as
// "run=true"/"run=false". Absent keys return def.
func (d Directive) Bool(key string, def bool) bool {
	v, ok := d.Opts[key]
	if !ok {
		return def
	}
	switch strings.ToLower(v) {
	case "", "true", "yes", "1":
		return true
	case "false", "no", "0":
		return false
	}
	return def
}

var (
	fenceOpenRe = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})(.*)$")
	directiveRe = regexp.MustCompile(`^\s*<!--\s*zenncode:?\s*(.*?)\s*-->\s*$`)
	playURLRe   = regexp.MustCompile(`^(https?://(?:go\.dev/play|play\.golang\.org|gotipplay\.golang\.org)/p/[A-Za-z0-9_\-]+)/?$`)
	// A whole line that is one Markdown link to the playground, with
	// whatever decoration the author put around it.
	playMarkdownRe = regexp.MustCompile(`^([^\[\]]*\[[^\]]*\]\()(https?://(?:go\.dev/play|play\.golang\.org|gotipplay\.golang\.org)/p/[A-Za-z0-9_\-]+)(\)[^\[\]]*)$`)
	directiveTok   = regexp.MustCompile(`([A-Za-z][A-Za-z0-9_\-]*)(?:=("[^"]*"|'[^']*'|[^\s]*))?`)
)

// ParseLink reports whether the whole line is a Go Playground link, bare or in
// Markdown syntax, and returns it.
func ParseLink(line string) (Link, bool) {
	trimmed := strings.TrimSpace(line)
	if m := playURLRe.FindStringSubmatch(trimmed); m != nil {
		return Link{URL: m[1]}, true
	}
	if m := playMarkdownRe.FindStringSubmatch(trimmed); m != nil {
		return Link{URL: m[2], Pre: m[1], Post: m[3]}, true
	}
	return Link{}, false
}

// ScanFile reads path and returns its fenced code blocks.
func ScanFile(path string) ([]Block, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Scan(path, src), nil
}

// Scan returns the fenced code blocks in src. path is recorded on each block
// and is not read from disk.
func Scan(path string, src []byte) []Block {
	lines := splitLines(string(src))

	var blocks []Block
	for i := 0; i < len(lines); i++ {
		m := fenceOpenRe.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		marker, info := m[1], strings.TrimSpace(m[2])

		// An info string is only valid on an opening fence, and a fence
		// closing an earlier one has already been consumed below, so any
		// fence we reach here opens a block.
		open := i
		end := len(lines) // unterminated fence: run to end of file
		var body []string
		for j := i + 1; j < len(lines); j++ {
			if isFenceClose(lines[j], marker) {
				end = j
				break
			}
			body = append(body, lines[j])
		}

		b := Block{
			File:      path,
			Lang:      langOf(info),
			Info:      info,
			OpenLine:  open + 1,
			CloseLine: end + 1,
			Code:      joinLines(body),
		}
		b.Directive, b.PlayAbove = scanAbove(lines, open)
		b.PlayBelow = scanBelow(lines, end)
		if n := len(blocks); n > 0 {
			blocks[n-1].NextLang, blocks[n-1].NextCode = b.Lang, b.Code
		}
		blocks = append(blocks, b)

		i = end // resume after the closing fence
	}
	return blocks
}

// scanAbove walks upwards from the opening fence collecting the directive and
// playground link that belong to the block. Blank lines are skipped; anything
// else stops the walk, so metadata must sit in the block's own paragraph.
func scanAbove(lines []string, open int) (Directive, Link) {
	var d Directive
	var link Link
	for i := open - 1; i >= 0; i-- {
		line := lines[i]
		switch {
		case strings.TrimSpace(line) == "":
			continue
		case directiveRe.MatchString(line):
			if !d.Present() {
				d = parseDirective(line, i+1)
			}
		default:
			l, ok := ParseLink(line)
			if !ok {
				return d, link
			}
			if !link.Found() {
				l.Line = i + 1
				link = l
			}
		}
	}
	return d, link
}

// scanBelow is scanAbove's mirror: it finds a playground link placed under the
// closing fence, which is the convention the older articles use.
func scanBelow(lines []string, end int) Link {
	for i := end + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" || directiveRe.MatchString(line) {
			continue
		}
		if l, ok := ParseLink(line); ok {
			l.Line = i + 1
			return l
		}
		return Link{}
	}
	return Link{}
}

func parseDirective(line string, lineNo int) Directive {
	m := directiveRe.FindStringSubmatch(line)
	if m == nil {
		return Directive{}
	}
	d := Directive{Line: lineNo, Raw: strings.TrimSpace(m[1]), Opts: map[string]string{}}
	for _, tok := range directiveTok.FindAllStringSubmatch(d.Raw, -1) {
		d.Opts[tok[1]] = strings.Trim(tok[2], `"'`)
	}
	return d
}

func isFenceClose(line, marker string) bool {
	s := strings.TrimRight(strings.TrimLeft(line, " "), " \t")
	if !strings.HasPrefix(s, marker[:1]) {
		return false
	}
	if len(s) < len(marker) {
		return false
	}
	return strings.Trim(s, marker[:1]) == ""
}

func langOf(info string) string {
	if i := strings.IndexAny(info, " :\t"); i >= 0 {
		return info[:i]
	}
	return info
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
