// Package mdedit applies line edits to a Markdown file.
//
// Edits are recorded against the line numbers of the original document and
// applied from the bottom up, so a caller can plan every change against one
// scan of the file without tracking how earlier edits shifted the rest.
package mdedit

import (
	"sort"
	"strings"
)

type opKind int

const (
	opReplace opKind = iota
	opInsertAfter
	opInsertBefore
	opRemove
)

type op struct {
	kind opKind
	line int // 1-indexed, in the original document
	text string
}

// Doc is a Markdown document with pending edits.
type Doc struct {
	lines    []string
	trailing bool // the original ended with a newline
	ops      []op
}

// New reads a document.
func New(src []byte) *Doc {
	s := strings.ReplaceAll(string(src), "\r\n", "\n")
	d := &Doc{trailing: strings.HasSuffix(s, "\n")}
	if d.trailing {
		s = strings.TrimSuffix(s, "\n")
	}
	if s != "" {
		d.lines = strings.Split(s, "\n")
	}
	return d
}

// Len is the number of lines in the original document.
func (d *Doc) Len() int { return len(d.lines) }

// Replace overwrites line n.
func (d *Doc) Replace(n int, text string) {
	d.ops = append(d.ops, op{opReplace, n, text})
}

// InsertParagraphAfter puts text on its own line below line n, separated from
// its neighbours by a blank line.
func (d *Doc) InsertParagraphAfter(n int, text string) {
	d.ops = append(d.ops, op{opInsertAfter, n, text})
}

// InsertParagraphBefore puts text on its own line above line n, separated from
// its neighbours by a blank line.
func (d *Doc) InsertParagraphBefore(n int, text string) {
	d.ops = append(d.ops, op{opInsertBefore, n, text})
}

// RemoveParagraph deletes line n along with the blank line that separated it
// from what follows, so removing a one-line paragraph leaves the text around
// it spaced as it was.
func (d *Doc) RemoveParagraph(n int) {
	d.ops = append(d.ops, op{opRemove, n, ""})
}

// Dirty reports whether any edit is pending.
func (d *Doc) Dirty() bool { return len(d.ops) > 0 }

// Bytes renders the document with its edits applied.
func (d *Doc) Bytes() []byte {
	lines := make([]string, len(d.lines))
	copy(lines, d.lines)

	ops := make([]op, len(d.ops))
	copy(ops, d.ops)
	// Bottom-up, so an edit never invalidates the line number of the next.
	sort.SliceStable(ops, func(i, j int) bool { return ops[i].line > ops[j].line })

	for _, o := range ops {
		i := o.line - 1 // index of the anchor line
		switch o.kind {
		case opReplace:
			if i >= 0 && i < len(lines) {
				lines[i] = o.text
			}
		case opInsertAfter:
			ins := []string{"", o.text}
			if i+1 < len(lines) && !blank(lines[i+1]) {
				ins = append(ins, "")
			}
			lines = splice(lines, i+1, ins)
		case opInsertBefore:
			ins := []string{o.text, ""}
			if i-1 >= 0 && !blank(lines[i-1]) {
				ins = append([]string{""}, ins...)
			}
			lines = splice(lines, i, ins)
		case opRemove:
			if i < 0 || i >= len(lines) {
				break
			}
			lo, hi := i, i+1
			switch {
			case i+1 < len(lines) && blank(lines[i+1]):
				// Take the blank line below with it; the one above stays as
				// the separator between the surrounding paragraphs.
				hi++
			case i-1 >= 0 && blank(lines[i-1]):
				// Nothing below to take, so take the blank above instead.
				lo--
			}
			lines = append(lines[:lo:lo], lines[hi:]...)
		}
	}

	out := strings.Join(lines, "\n")
	if d.trailing {
		out += "\n"
	}
	return []byte(out)
}

func splice(lines []string, at int, ins []string) []string {
	out := make([]string, 0, len(lines)+len(ins))
	out = append(out, lines[:at]...)
	out = append(out, ins...)
	return append(out, lines[at:]...)
}

func blank(s string) bool { return strings.TrimSpace(s) == "" }
