package mdedit

import "testing"

func TestInsertParagraphAfter(t *testing.T) {
	src := "```go\nvar x int\n```\ntext\n"
	d := New([]byte(src))
	d.InsertParagraphAfter(3, "URL")
	got := string(d.Bytes())
	want := "```go\nvar x int\n```\n\nURL\n\ntext\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestInsertParagraphAfterExistingBlank(t *testing.T) {
	// A blank line is already there; do not add a second one.
	src := "```\n```\n\ntext\n"
	d := New([]byte(src))
	d.InsertParagraphAfter(2, "URL")
	got := string(d.Bytes())
	want := "```\n```\n\nURL\n\ntext\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestInsertParagraphAtEnd(t *testing.T) {
	src := "text\n```\n```\n"
	d := New([]byte(src))
	d.InsertParagraphAfter(3, "URL")
	got := string(d.Bytes())
	want := "text\n```\n```\n\nURL\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestInsertParagraphBefore(t *testing.T) {
	src := "text\n```go\nvar x int\n```\n"
	d := New([]byte(src))
	d.InsertParagraphBefore(2, "URL")
	got := string(d.Bytes())
	want := "text\n\nURL\n\n```go\nvar x int\n```\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestRemoveParagraph(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
		line int
		want string
	}{
		{
			"between paragraphs",
			"```\n```\n\nURL\n\ntext\n", 4,
			"```\n```\n\ntext\n",
		},
		{
			"at the end of the file",
			"```\n```\n\nURL\n", 4,
			"```\n```\n",
		},
		{
			"with no blank line around it",
			"text\nURL\ntail\n", 2,
			"text\ntail\n",
		},
		{
			"out of range is ignored",
			"text\n", 9,
			"text\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := New([]byte(tt.src))
			d.RemoveParagraph(tt.line)
			if got := string(d.Bytes()); got != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestRemoveParagraphWithOtherEdits(t *testing.T) {
	// Two links in one file: the first goes away, the second is rewritten.
	src := "```\n```\n\nOLD\n\n```\n```\n\nURL\n\ntail\n"
	d := New([]byte(src))
	d.RemoveParagraph(4)
	d.Replace(9, "NEW")
	got := string(d.Bytes())
	want := "```\n```\n\n```\n```\n\nNEW\n\ntail\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestReplaceRange(t *testing.T) {
	for _, tt := range []struct {
		name     string
		src      string
		from, to int
		lines    []string
		want     string
	}{
		{
			"same number of lines",
			"```go\nold\n```\n", 2, 2, []string{"new"},
			"```go\nnew\n```\n",
		},
		{
			"more lines than before",
			"```go\nold\n```\n", 2, 2, []string{"a", "b", "c"},
			"```go\na\nb\nc\n```\n",
		},
		{
			"fewer lines than before",
			"```go\na\nb\nc\n```\n", 2, 4, []string{"one"},
			"```go\none\n```\n",
		},
		{
			"out of range is ignored",
			"text\n", 3, 9, []string{"x"},
			"text\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := New([]byte(tt.src))
			d.ReplaceRange(tt.from, tt.to, tt.lines)
			if got := string(d.Bytes()); got != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestReplaceRangeWithLaterEdits(t *testing.T) {
	// The fence contents grow while a link below it is removed. Both edits are
	// planned against the original line numbers.
	src := "```go\nold\n```\n\nURL\n\ntail\n"
	d := New([]byte(src))
	d.ReplaceRange(2, 2, []string{"a", "b"})
	d.RemoveParagraph(5)
	got := string(d.Bytes())
	want := "```go\na\nb\n```\n\ntail\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestReplaceAndMultipleEdits(t *testing.T) {
	// Edits are planned against the original line numbers; applying the
	// later one first must not shift the earlier one.
	src := "OLD\n\n```\n```\n\ntail\n"
	d := New([]byte(src))
	d.Replace(1, "NEW")
	d.InsertParagraphAfter(4, "URL")
	got := string(d.Bytes())
	want := "NEW\n\n```\n```\n\nURL\n\ntail\n"
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestPreservesMissingTrailingNewline(t *testing.T) {
	d := New([]byte("a\nb"))
	d.Replace(1, "A")
	if got := string(d.Bytes()); got != "A\nb" {
		t.Errorf("got %q", got)
	}
}

func TestCleanRoundTrip(t *testing.T) {
	src := "# heading\n\ntext\n"
	d := New([]byte(src))
	if d.Dirty() {
		t.Error("a fresh document should not be dirty")
	}
	if got := string(d.Bytes()); got != src {
		t.Errorf("got %q, want %q", got, src)
	}
}
