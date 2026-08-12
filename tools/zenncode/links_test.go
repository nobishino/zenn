package main

import (
	"testing"

	"github.com/nobishino/zenn/tools/zenncode/internal/mdscan"
)

// goBlocks scans src and keeps the Go blocks, which is what planLinks sees.
func goBlocks(t *testing.T, src string) []mdscan.Block {
	t.Helper()
	var out []mdscan.Block
	for _, b := range mdscan.Scan("a.md", []byte(src)) {
		if b.Lang == "go" {
			out = append(out, b)
		}
	}
	return out
}

const fence = "```go\nvar x int\n```\n"

func TestPlanLinksBelowConvention(t *testing.T) {
	// Two blocks, each with its link underneath. The middle link is
	// ambiguous, but the last one is not and settles the convention.
	src := fence + "\nhttps://go.dev/play/p/aaa\n\n" + fence + "\nhttps://go.dev/play/p/bbb\n"
	plan := planLinks(goBlocks(t, src))
	if plan.conv != sideBelow {
		t.Fatalf("conv = %v, want below", plan.conv)
	}
	if got := plan.link(0).URL; got != "https://go.dev/play/p/aaa" {
		t.Errorf("block 0 owns %q", got)
	}
	if got := plan.link(1).URL; got != "https://go.dev/play/p/bbb" {
		t.Errorf("block 1 owns %q", got)
	}
}

func TestPlanLinksAboveConvention(t *testing.T) {
	src := "https://go.dev/play/p/aaa\n\n" + fence + "\nhttps://go.dev/play/p/bbb\n\n" + fence
	plan := planLinks(goBlocks(t, src))
	if plan.conv != sideAbove {
		t.Fatalf("conv = %v, want above", plan.conv)
	}
	if got := plan.link(0).URL; got != "https://go.dev/play/p/aaa" {
		t.Errorf("block 0 owns %q", got)
	}
	if got := plan.link(1).URL; got != "https://go.dev/play/p/bbb" {
		t.Errorf("block 1 owns %q", got)
	}
}

func TestPlanLinksAmbiguousFollowsConvention(t *testing.T) {
	// Only one link, sitting between two blocks: below the first, above the
	// second. With nothing to learn from, the repository default applies and
	// it belongs to the block above it.
	src := fence + "\nhttps://go.dev/play/p/aaa\n\n" + fence
	plan := planLinks(goBlocks(t, src))
	if plan.conv != defaultSide {
		t.Fatalf("conv = %v, want the default", plan.conv)
	}
	if got := plan.link(0).URL; got != "https://go.dev/play/p/aaa" {
		t.Errorf("block 0 owns %q, want the ambiguous link", got)
	}
	if plan.link(1).Found() {
		t.Errorf("block 1 owns %+v, want nothing", plan.link(1))
	}
}

func TestPlanLinksNone(t *testing.T) {
	plan := planLinks(goBlocks(t, fence+"\ntext\n\n"+fence))
	if plan.conv != defaultSide {
		t.Errorf("conv = %v", plan.conv)
	}
	if plan.link(0).Found() || plan.link(1).Found() {
		t.Errorf("expected no owned links, got %+v %+v", plan.link(0), plan.link(1))
	}
}
