package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nobishino/zenn/tools/zenncode/internal/lock"
	"github.com/nobishino/zenn/tools/zenncode/internal/mdedit"
	"github.com/nobishino/zenn/tools/zenncode/internal/mdscan"
	"github.com/nobishino/zenn/tools/zenncode/internal/playground"
)

// action is one planned change to an article.
type action struct {
	res  result
	kind actionKind
	link mdscan.Link // the link that is there now, if any
	url  string      // the link it should have; empty until shared
	text string      // the line to write, the URL in the article's own style
	side side        // where a new link goes
}

type actionKind int

const (
	actionAdd actionKind = iota
	actionUpdate
)

func (a action) String() string {
	switch a.kind {
	case actionUpdate:
		return fmt.Sprintf("%s: update link %s -> %s", posOf(a.res.block, a.link.Line), a.link.URL, a.url)
	default:
		return fmt.Sprintf("%s: add link (%s) %s", a.res.block.Pos(), a.side, a.url)
	}
}

func posOf(b mdscan.Block, line int) string {
	return fmt.Sprintf("%s:%d", b.File, line)
}

// cmdFix shares every snippet that verifies and writes its link into the
// article. Snippets that do not verify are left alone: a link is a promise
// that the code runs, and this command only makes that promise for code it
// has compiled.
func cmdFix(paths []string, jobs int, dryRun bool) error {
	// Pruning is only safe when the run saw every article; a run over a
	// subset knows nothing about the hashes the rest of the repository uses.
	fullRun := len(paths) == 0
	results, err := process(paths, jobs)
	if err != nil {
		return err
	}
	lk, err := lock.Load(lockName)
	if err != nil {
		return err
	}

	var (
		client  = &playground.Client{Pause: sharePause}
		ctx     = context.Background()
		shared  int
		changed int
		files   int
	)

	for _, group := range groupByFile(results) {
		plan := planLinks(blocksOf(group))
		var actions []action

		for i, r := range group {
			if r.status != statusOK || r.block.Directive.Get("playground") == "none" {
				continue
			}
			link := plan.link(i)
			url, known := lk.URL(r.prog.Hash())
			if known && link.Found() && link.URL == url {
				continue // already correct
			}
			a := action{res: r, link: link, url: url, side: plan.conv}
			if link.Found() {
				a.kind = actionUpdate
			} else {
				a.kind = actionAdd
			}
			if !known {
				if dryRun {
					a.url = "(share pending)"
					actions = append(actions, a)
					continue
				}
				url, err = client.Share(ctx, r.prog.Source)
				if err != nil {
					return fmt.Errorf("%s: %w", r.block.Pos(), err)
				}
				shared++
				lk.Set(r.prog.Hash(), url, r.block.File)
				a.url = url
			}
			if a.kind == actionUpdate && a.link.URL == a.url {
				continue // the article already had the right link
			}
			if a.kind == actionUpdate {
				a.text = a.link.Render(a.url)
			} else {
				a.text = plan.render(a.url)
			}
			actions = append(actions, a)
		}

		if len(actions) == 0 {
			continue
		}
		files++
		changed += len(actions)
		for _, a := range actions {
			fmt.Println(a)
		}
		if dryRun {
			continue
		}
		if err := applyActions(group[0].block.File, actions); err != nil {
			return err
		}
	}

	var pruned int
	if fullRun {
		live := map[string]bool{}
		for _, r := range results {
			if h := r.prog.Hash(); h != "" {
				live[h] = true
			}
		}
		pruned = lk.Prune(live)
	}

	if dryRun {
		fmt.Printf("\n%d links to write in %d files (dry run; nothing shared or written)\n", changed, files)
		return nil
	}
	if lk.Dirty() {
		if err := lk.Save(lockName); err != nil {
			return err
		}
	}
	fmt.Printf("\n%d links written in %d files, %d snippets shared, %d stale lock entries pruned\n",
		changed, files, shared, pruned)
	return nil
}

// applyActions rewrites one article.
func applyActions(path string, actions []action) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	doc := mdedit.New(src)
	for _, a := range actions {
		switch {
		case a.kind == actionUpdate:
			doc.Replace(a.link.Line, a.text)
		case a.side == sideAbove:
			doc.InsertParagraphBefore(blockStart(a.res.block), a.text)
		default:
			doc.InsertParagraphAfter(a.res.block.CloseLine, a.text)
		}
	}
	if !doc.Dirty() {
		return nil
	}
	return os.WriteFile(path, doc.Bytes(), 0o644)
}

// groupByFile splits results into per-article runs, keeping the order process
// produced so that block indices line up with planLinks.
func groupByFile(results []result) [][]result {
	var groups [][]result
	for i := 0; i < len(results); {
		j := i
		for j < len(results) && results[j].block.File == results[i].block.File {
			j++
		}
		groups = append(groups, results[i:j])
		i = j
	}
	return groups
}

func blocksOf(results []result) []mdscan.Block {
	blocks := make([]mdscan.Block, len(results))
	for i, r := range results {
		blocks[i] = r.block
	}
	return blocks
}
