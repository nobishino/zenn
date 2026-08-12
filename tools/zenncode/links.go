package main

import "github.com/nobishino/zenn/tools/zenncode/internal/mdscan"

// side is where an article puts the playground link for a code block.
type side int

const (
	sideBelow side = iota
	sideAbove
)

func (s side) String() string {
	if s == sideAbove {
		return "above"
	}
	return "below"
}

// defaultSide is used for an article that has no links to learn from. Of the
// links already in this repository, roughly three quarters sit under the block.
const defaultSide = sideBelow

// filePlan says, for one article, which side new links go on and which block
// owns each link that is already there.
type filePlan struct {
	conv  side
	owned map[int]mdscan.Link // index into the article's Go blocks
	tmpl  mdscan.Link         // the style new links are written in
}

// link returns the link owned by the i'th Go block of the article.
func (p filePlan) link(i int) mdscan.Link { return p.owned[i] }

// render writes url the way this article writes its links: bare, or wrapped
// the way the links already there are wrapped.
func (p filePlan) render(url string) string {
	if p.tmpl.Inline() {
		return p.tmpl.Render(url)
	}
	return url
}

// planLinks works out the ownership of every playground link in an article.
//
// A link written between two code blocks is recorded by mdscan as being below
// one and above the other, and nothing in the text says which it belongs to.
// The unambiguous links -- those with a code block on one side only -- decide
// the article's convention, and the convention decides the rest.
func planLinks(blocks []mdscan.Block) filePlan {
	type candidate struct {
		below int // block index whose PlayBelow this is, -1 if none
		above int // block index whose PlayAbove this is, -1 if none
	}
	cands := map[int]*candidate{}
	get := func(line int) *candidate {
		c, ok := cands[line]
		if !ok {
			c = &candidate{below: -1, above: -1}
			cands[line] = c
		}
		return c
	}
	for i, b := range blocks {
		if l := b.PlayAbove; l.Found() {
			get(l.Line).above = i
		}
		if l := b.PlayBelow; l.Found() {
			get(l.Line).below = i
		}
	}

	var votesAbove, votesBelow int
	for _, c := range cands {
		switch {
		case c.above >= 0 && c.below < 0:
			votesAbove++
		case c.below >= 0 && c.above < 0:
			votesBelow++
		}
	}
	conv := defaultSide
	switch {
	case votesAbove > votesBelow:
		conv = sideAbove
	case votesBelow > votesAbove:
		conv = sideBelow
	}

	plan := filePlan{conv: conv, owned: map[int]mdscan.Link{}}
	for line, c := range cands {
		owner := -1
		switch {
		case c.above >= 0 && c.below >= 0:
			// Ambiguous: the article's own convention breaks the tie.
			if conv == sideAbove {
				owner = c.above
			} else {
				owner = c.below
			}
		case c.above >= 0:
			owner = c.above
		case c.below >= 0:
			owner = c.below
		}
		if owner < 0 {
			continue
		}
		owned := blocks[owner].PlayAbove
		if blocks[owner].PlayBelow.Line == line {
			owned = blocks[owner].PlayBelow
		}
		plan.owned[owner] = owned
	}
	plan.tmpl = dominantStyle(plan.owned)
	return plan
}

// dominantStyle picks the wrapper new links should use. The book writes its
// links as "[Go Playgroundで実行する](url)"; copying that keeps a page
// consistent instead of mixing styles.
func dominantStyle(owned map[int]mdscan.Link) mdscan.Link {
	var inline, bare int
	best := mdscan.Link{}
	for _, l := range owned {
		if !l.Inline() {
			bare++
			continue
		}
		inline++
		if !best.Found() || l.Line < best.Line {
			best = l
		}
	}
	if inline > bare {
		return best
	}
	return mdscan.Link{}
}
