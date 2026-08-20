// Command zenncode verifies the Go sample code embedded in the articles of
// this repository.
//
// Every ```go block in an article is turned into a complete program (see
// package gosnippet) and compiled. Snippets that do not compile today are
// tolerated only while they are listed in the baseline file.
//
//	zenncode verify [paths...]    compile every snippet; non-zero exit on new failures
//	zenncode baseline [paths...]  record today's failures as the baseline
//	zenncode list [paths...]      one line per snippet
//	zenncode show file.md:123     print the normalized program for one snippet
//
// Paths default to the articles and books directories and are interpreted
// relative to the repository root.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nobishino/zenn/tools/zenncode/internal/baseline"
	"github.com/nobishino/zenn/tools/zenncode/internal/check"
	"github.com/nobishino/zenn/tools/zenncode/internal/gosnippet"
	"github.com/nobishino/zenn/tools/zenncode/internal/lock"
	"github.com/nobishino/zenn/tools/zenncode/internal/mdscan"
)

const (
	baselineName = "zenncode-baseline.json"
	lockName     = "zenncode-lock.json"
	buildTimeout = 90 * time.Second
	// sharePause keeps a bulk run gentle on the playground service.
	sharePause = 200 * time.Millisecond
)

var defaultPaths = []string{"articles", "books"}

// errFailed reports verification failures without printing a message of its
// own; the failures have already been listed.
var errFailed = errors.New("verification failed")

func main() {
	if err := run(os.Args[1:]); err != nil {
		if !errors.Is(err, errFailed) {
			fmt.Fprintln(os.Stderr, "zenncode:", err)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return errors.New("no command given")
	}
	cmd, rest := args[0], args[1:]

	fset := flag.NewFlagSet("zenncode "+cmd, flag.ExitOnError)
	root := fset.String("root", "", "repository root (default: nearest ancestor containing articles/)")
	jobs := fset.Int("j", runtime.NumCPU(), "number of concurrent builds")
	verbose := fset.Bool("v", false, "list every snippet, not just failures")
	dryRun := fset.Bool("n", false, "fix: report what would change without sharing or writing")
	if err := fset.Parse(rest); err != nil {
		return err
	}

	dir, err := repoRoot(*root)
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}

	switch cmd {
	case "verify":
		return cmdVerify(fset.Args(), *jobs, *verbose)
	case "fix":
		return cmdFix(fset.Args(), *jobs, *dryRun)
	case "baseline":
		return cmdBaseline(fset.Args(), *jobs)
	case "list":
		return cmdList(fset.Args(), *jobs, *verbose)
	case "show":
		return cmdShow(fset.Args())
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: zenncode <command> [flags] [paths...]

commands:
  verify     compile every Go snippet and check its playground link
  fix        share snippets and write their playground links into the articles
  baseline   record today's failures in `+baselineName+`
  list       print one line per snippet
  show       print the normalized program for one snippet (file.md:line)

flags:
  -root dir  repository root (default: nearest ancestor containing articles/)
  -j n       concurrent builds (default: NumCPU)
  -v         list every snippet, not just failures
  -n         fix: report what would change without sharing or writing
`)
}

// status is the outcome of processing one snippet.
type status int

const (
	statusOK status = iota
	statusSkipped
	statusFailed
)

// result is one processed snippet.
type result struct {
	block  mdscan.Block
	prog   gosnippet.Program
	src    *fileSource // set for file= blocks, where the repository holds the code
	stage  string      // "normalize", "build", "run", ...; set when failed
	output string      // failure detail, positioned for a human reader
	reason string      // failure detail without article line numbers, for the baseline
	status status
	expect expectKind
	known  bool // matched a baseline entry
}

// linkable reports whether a snippet should carry a playground link. A link
// is a promise that the code is real, so it is only made for code that
// verified, and only when there is a program to share -- a snippet that does
// not even parse has none.
func (r result) linkable() bool {
	return r.status == statusOK && r.prog.Hash() != "" && !r.unlinked() && !r.kept()
}

// kept reports whether the article's own link is to be left exactly as it is,
// with `playground=keep`. The old generics articles link to gotipplay, whose
// snippets were written for a development build of Go and no longer compile
// anywhere; the link is a record of what the article said at the time, so the
// tool neither checks it nor rewrites it.
func (r result) kept() bool { return r.block.Directive.Get("playground") == "keep" }

// unlinked reports whether this block should carry no playground link at all.
// Unlike the other reasons a snippet goes unlinked, these come from the block
// itself rather than from a verdict on the code, so they also mean an existing
// link should go away.
func (r result) unlinked() bool { return r.unlinkReason() != "" }

// unlinkReason says why a block carries no link, phrased to be read at the
// start of a sentence. It returns "" for the blocks that should have one.
func (r result) unlinkReason() string {
	switch {
	case r.block.Directive.Get("playground") == "none":
		return "this block asks for no playground link"
	case r.block.Directive.Has("file"):
		// A file= sample exists precisely because it does not fit in one
		// self-contained file, which is all the playground can hold.
		return "this block mirrors " + r.block.Directive.Get("file") + ", which the playground cannot build"
	case crossTarget(r.block.Directive) != "":
		// The playground builds for its own platform, so a link to code
		// written for another one would not even compile there.
		return "this snippet is built for " + crossTarget(r.block.Directive) + ", which the playground cannot build"
	}
	return ""
}

// fail records a failure and returns the result, so a caller can write
// `return r.fail(...)` at each point a snippet can go wrong.
func (r result) fail(stage, detail string) result {
	r.status, r.stage = statusFailed, stage
	r.output, r.reason = detail, detail
	return r
}

func (r result) entry() baseline.Entry {
	return baseline.Entry{
		File:   r.block.File,
		Line:   r.block.OpenLine,
		Hash:   r.prog.Hash(),
		Raw:    rawHash(r),
		Stage:  r.stage,
		Reason: baseline.Reason(r.reason),
	}
}

// rawHash keys snippets that never became a program, where there is no
// normalized source to hash.
func rawHash(r result) string {
	if r.prog.Source != "" {
		return ""
	}
	p := gosnippet.Program{Source: r.block.Code}
	return p.Hash()
}

func cmdVerify(paths []string, jobs int, verbose bool) error {
	results, err := process(paths, jobs)
	if err != nil {
		return err
	}
	base, err := baseline.Load(baselineName)
	if err != nil {
		return err
	}
	lk, err := lock.Load(lockName)
	if err != nil {
		return err
	}

	var failed, known, skipped, ok int
	var stale []result
	for i := range results {
		r := &results[i]
		switch r.status {
		case statusSkipped:
			skipped++
		case statusOK:
			ok++
			if base.Has(r.entry()) {
				stale = append(stale, *r)
			}
		case statusFailed:
			if base.Has(r.entry()) {
				r.known = true
				known++
			} else {
				failed++
			}
		}
	}

	for _, r := range results {
		if r.status == statusFailed && !r.known {
			printFailure(r)
		} else if verbose {
			fmt.Printf("%s: %s\n", r.block.Pos(), describe(r))
		}
	}
	for _, r := range stale {
		fmt.Printf("%s: now verifies; remove it from %s\n", r.block.Pos(), baselineName)
	}

	issues := checkLinks(results, lk)
	for _, is := range issues {
		fmt.Printf("%s: %s\n", is.pos, is.msg)
	}

	fmt.Printf("\n%d Go snippets: %d ok, %d known-failing, %d skipped, %d failed, %d link problems\n",
		len(results), ok, known, skipped, failed, len(issues))
	if failed > 0 || len(issues) > 0 {
		return errFailed
	}
	return nil
}

// unlinkMsg explains a link that should not be there. An ambiguous link is one
// written between two code blocks, where the tool only guessed which block it
// belongs to; that guess is not good enough to delete on, so the author is
// asked to do it.
func unlinkMsg(reason string, ambiguous bool) string {
	if ambiguous {
		return reason + ", but this link could belong to either neighbouring block; remove it by hand if it is this one's"
	}
	return reason + "; run `make fix` to remove it"
}

// linkIssue is a mismatch between an article's playground link and the code
// next to it.
type linkIssue struct {
	pos string
	msg string
}

// checkLinks compares each verified snippet with the link the article shows
// for it. It reads only the lock file, so it needs no network.
func checkLinks(results []result, lk *lock.File) []linkIssue {
	var issues []linkIssue
	for _, group := range groupByFile(results) {
		plan := planLinks(blocksOf(group))
		for i, r := range group {
			if r.kept() {
				continue
			}
			if r.unlinked() {
				if link := plan.link(i); link.Found() {
					issues = append(issues, linkIssue{posOf(r.block, link.Line), unlinkMsg(r.unlinkReason(), plan.ambiguous(i))})
				}
				continue
			}
			if !r.linkable() {
				// The snippet did not verify, so there is no code this link
				// can be checked against -- which is exactly the state this
				// command exists to surface. Saying so is the only honest
				// option: the tool cannot tell whether the link still shows
				// what the article claims.
				if link := plan.link(i); link.Found() {
					issues = append(issues, linkIssue{posOf(r.block, link.Line),
						"this block does not verify, so its link is unchecked; fix the block, or mark it `playground=keep` if the link is deliberate"})
				}
				continue
			}
			link := plan.link(i)
			url, known := lk.URL(r.prog.Hash())
			switch {
			case !known && link.Found():
				// There is a link, but this exact code has never been
				// shared: either the snippet was edited, or the link was
				// made by hand before this tool existed.
				issues = append(issues, linkIssue{posOf(r.block, link.Line),
					"this link was not made from the code next to it; run `make fix`"})
			case !known:
				issues = append(issues, linkIssue{r.block.Pos(),
					"no playground link for this code yet; run `make fix`"})
			case !link.Found():
				issues = append(issues, linkIssue{r.block.Pos(),
					"playground link is missing; run `make fix`"})
			case link.URL != url:
				issues = append(issues, linkIssue{posOf(r.block, link.Line),
					fmt.Sprintf("playground link does not match the code above it (%s, want %s); run `make fix`", link.URL, url)})
			}
		}
	}
	return issues
}

func cmdBaseline(paths []string, jobs int) error {
	results, err := process(paths, jobs)
	if err != nil {
		return err
	}
	// A run over a subset of the articles only speaks for the files it saw,
	// so the entries for every other file are carried over untouched.
	// Rewriting the whole baseline from a partial run would silently forget
	// the rest of the repository's known failures.
	old, err := baseline.Load(baselineName)
	if err != nil {
		return err
	}
	var fresh []baseline.Entry
	scanned := map[string]bool{}
	for _, r := range results {
		scanned[r.block.File] = true
		if r.status == statusFailed {
			fresh = append(fresh, r.entry())
		}
	}
	entries, kept := mergeBaseline(old.Entries, scanned, fresh)
	if err := baseline.Save(baselineName, entries); err != nil {
		return err
	}
	fmt.Printf("wrote %s: %d of %d snippets recorded as failing, %d entries kept from files not scanned\n",
		baselineName, len(entries)-kept, len(results), kept)
	return nil
}

// mergeBaseline combines the failures just recorded with the entries for
// files this run did not look at, and reports how many of the latter there
// were. It is what keeps `zenncode baseline articles/one.md` from forgetting
// everything known about the other articles.
func mergeBaseline(old []baseline.Entry, scanned map[string]bool, fresh []baseline.Entry) (entries []baseline.Entry, kept int) {
	for _, e := range old {
		if !scanned[e.File] {
			entries = append(entries, e)
			kept++
		}
	}
	return append(entries, fresh...), kept
}

func cmdList(paths []string, jobs int, verbose bool) error {
	results, err := process(paths, jobs)
	if err != nil {
		return err
	}
	for _, r := range results {
		play := "-"
		if l := r.block.PlayAbove; l.Found() {
			play = "^ " + l.URL
		}
		if l := r.block.PlayBelow; l.Found() {
			if play != "-" {
				play += "  "
			} else {
				play = ""
			}
			play += "v " + l.URL
		}
		hash := r.prog.Hash()
		if hash == "" {
			hash = "-"
		}
		expect := r.expect.String()
		if r.status == statusSkipped {
			expect = "-"
		}
		fmt.Printf("%-52s %-16s %-8s %-13s %s\n", r.block.Pos(), hash, label(r.status), expect, play)
		if verbose && r.output != "" {
			fmt.Printf("    %s\n", indent(r.output))
		}
	}
	return nil
}

func cmdShow(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: zenncode show file.md:line")
	}
	file, lineStr, ok := strings.Cut(args[0], ":")
	if !ok {
		return errors.New("usage: zenncode show file.md:line")
	}
	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return fmt.Errorf("bad line number %q", lineStr)
	}
	blocks, err := mdscan.ScanFile(file)
	if err != nil {
		return err
	}
	for _, b := range blocks {
		// Accept any line of the block, including the directive and
		// playground link above it, so the line the reader is looking at
		// works whichever one it is.
		if b.Lang != "go" || line < blockStart(b) || line > b.CloseLine {
			continue
		}
		prog, err := gosnippet.Normalize(b.Code, normalizeOpts(b))
		if err != nil {
			return fmt.Errorf("%s: %w", b.Pos(), err)
		}
		fmt.Print(prog.Source)
		return nil
	}
	return fmt.Errorf("no Go block at %s:%d", file, line)
}

// blockStart is the first line belonging to a block, counting the metadata
// lines above its opening fence.
func blockStart(b mdscan.Block) int {
	start := b.OpenLine
	for _, line := range []int{b.Directive.Line, b.PlayAbove.Line} {
		if line != 0 && line < start {
			start = line
		}
	}
	return start
}

// process scans paths and compiles every Go snippet it finds.
func process(paths []string, jobs int) ([]result, error) {
	if len(paths) == 0 {
		paths = defaultPaths
	}
	files, err := markdownFiles(paths)
	if err != nil {
		return nil, err
	}

	var blocks []mdscan.Block
	for _, f := range files {
		bs, err := mdscan.ScanFile(f)
		if err != nil {
			return nil, err
		}
		for _, b := range bs {
			if b.Lang == "go" {
				blocks = append(blocks, b)
			}
		}
	}

	builder, err := check.NewBuilder()
	if err != nil {
		return nil, err
	}
	defer builder.Close()

	if jobs < 1 {
		jobs = 1
	}
	results := make([]result, len(blocks))
	sem := make(chan struct{}, jobs)
	var wg sync.WaitGroup
	for i, b := range blocks {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = processBlock(builder, b)
		}()
	}
	wg.Wait()
	return results, nil
}

// supportedKeys are the directive keys this build honors.
var supportedKeys = map[string]bool{
	"skip":       true,
	"imports":    true,
	"goversion":  true,
	"goos":       true,
	"goarch":     true,
	"playground": true,
	"file":       true,
	"expect":     true,
	"error":      true,
	"panic":      true,
	"output":     true,
	"timeout":    true,
}

// plannedKeys are designed but not implemented. Rejecting them keeps a
// directive from silently doing nothing in an article.
// isBoolWord reports whether v is one of the spellings Directive.Bool knows,
// the empty string standing for a key written on its own.
func isBoolWord(v string) bool {
	switch strings.ToLower(v) {
	case "", "true", "yes", "1", "false", "no", "0":
		return true
	}
	return false
}

var plannedKeys = map[string]string{
	"run": "there is no run= key; building without running is the default, and expect=run asks for a run",
}

func validateDirective(d mdscan.Directive) error {
	if d.Has("playground") {
		switch v := d.Get("playground"); v {
		case "none", "keep":
		default:
			return fmt.Errorf("playground=%s: want none or keep", v)
		}
	}
	if v := d.Get("skip"); d.Has("skip") && !isBoolWord(v) {
		// Directive.Bool falls back to its default for anything it does not
		// recognize, so `skip=treu` would quietly verify the block after all.
		return fmt.Errorf("skip=%s: write skip on its own, or skip=true/false", v)
	}
	for _, key := range []string{"goos", "goarch"} {
		if d.Has(key) && d.Get(key) == "" {
			return fmt.Errorf("%s needs a value, as in goos=wasip1 goarch=wasm", key)
		}
	}
	if d.Has("file") {
		// A file= block is compiled as the package it belongs to, so its
		// module -- not the directive -- settles imports and language version.
		for _, key := range []string{"imports", "goversion"} {
			if d.Has(key) {
				return fmt.Errorf("%s does not apply with file=: the package's own module decides that", key)
			}
		}
	}
	for key := range d.Opts {
		if supportedKeys[key] {
			continue
		}
		if msg, ok := plannedKeys[key]; ok {
			return fmt.Errorf("%s: %s", key, msg)
		}
		return fmt.Errorf("unknown directive key %q", key)
	}
	return nil
}

func processBlock(builder *check.Builder, b mdscan.Block) result {
	r := result{block: b}
	if err := validateDirective(b.Directive); err != nil {
		return r.fail("directive", err.Error())
	}
	if b.Directive.Bool("skip", false) {
		r.status = statusSkipped
		return r
	}
	exp, err := parseExpectation(b)
	if err != nil {
		return r.fail("directive", err.Error())
	}
	r.expect = exp.kind

	if b.Directive.Has("file") {
		return processFileBlock(builder, b, exp, r)
	}

	prog, err := gosnippet.Normalize(b.Code, normalizeOpts(b))
	r.prog = prog
	if err != nil {
		if exp.kind == expectCompileError {
			// Code that does not even parse does not compile, which is what
			// the article claims. There is no program to share, so this
			// block gets no playground link. A named diagnostic still has
			// to appear, or `error=` would pass on any typo.
			if exp.errRe != nil && !exp.errRe.MatchString(err.Error()) {
				return r.fail("expect", fmt.Sprintf("the parse error does not match %s:\n%s", exp.errRe, locate(b, err)))
			}
			r.status = statusOK
			return r
		}
		r.status, r.stage = statusFailed, "normalize"
		r.output, r.reason = locate(b, err), err.Error()
		return r
	}

	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()
	res, err := builder.Build(ctx, check.Request{
		Program:   prog.Source,
		IsMain:    prog.HasMain,
		GoVersion: b.Directive.Get("goversion"),
		Binary:    exp.runs(),
		GOOS:      b.Directive.Get("goos"),
		GOARCH:    b.Directive.Get("goarch"),
	})
	if err != nil {
		return r.fail("build", fmt.Sprintf("%v\n%s", err, res.Output))
	}

	if exp.kind == expectCompileError {
		switch {
		case res.OK:
			return r.fail("expect", "the article expects a compile error, but this code compiles")
		case exp.errRe != nil && !exp.errRe.MatchString(res.Output):
			return r.fail("expect", fmt.Sprintf("the compiler error does not match %s:\n%s", exp.errRe, res.Output))
		}
		r.status = statusOK
		return r
	}
	if !res.OK {
		return r.fail("build", res.Output)
	}
	if !exp.runs() {
		r.status = statusOK
		return r
	}
	return checkRun(ctx, r, exp, res.Bin)
}

// checkRun executes a built program and judges it against the expectation.
// It serves both kinds of snippet: the ones restored from the article text and
// the ones built from a file= package.
func checkRun(ctx context.Context, r result, exp expectation, bin string) result {
	run, err := check.Run(ctx, bin, exp.timeout)
	if err != nil {
		return r.fail("run", err.Error())
	}
	if run.TimedOut {
		return r.fail("run", fmt.Sprintf("the program did not finish within %s", exp.timeout))
	}
	stderr := strings.TrimRight(run.Stderr, "\n")

	if exp.kind == expectPanic {
		switch {
		case run.OK:
			return r.fail("expect", "the article expects a run-time panic, but the program ran to completion")
		case !aborted(run.Stderr):
			return r.fail("expect", fmt.Sprintf("the article expects a run-time panic, but the program exited with status %d without one\n%s", run.ExitCode, stderr))
		case exp.panicRe != nil && !exp.panicRe.MatchString(run.Stderr):
			return r.fail("expect", fmt.Sprintf("the panic message does not match %s:\n%s", exp.panicRe, stderr))
		}
		r.status = statusOK
		return r
	}

	if !run.OK {
		return r.fail("run", fmt.Sprintf("the program exited with status %d\n%s", run.ExitCode, stderr))
	}
	if exp.output != "" {
		if d := diffOutput(exp.output, run.Stdout); d != "" {
			return r.fail("output", d)
		}
	}
	r.status = statusOK
	return r
}

func normalizeOpts(b mdscan.Block) gosnippet.Options {
	var opts gosnippet.Options
	if v := b.Directive.Get("imports"); v != "" {
		for _, p := range strings.Split(v, ",") {
			if p = strings.TrimSpace(p); p != "" {
				opts.Imports = append(opts.Imports, p)
			}
		}
	}
	opts.Filename = fmt.Sprintf("%s-%d.go", strings.TrimSuffix(filepath.Base(b.File), ".md"), b.OpenLine)
	return opts
}

func printFailure(r result) {
	fmt.Printf("%s: %s failed\n", r.block.Pos(), r.stage)
	fmt.Printf("    %s\n", indent(r.output))
	if r.stage == "build" {
		fmt.Printf("    (line numbers refer to the normalized program: zenncode show %s)\n", r.block.Pos())
	}
}

// locate rewrites parse diagnostics into positions in the article, so an
// editor can jump straight to the offending line of the fenced block.
func locate(b mdscan.Block, err error) string {
	var pe *gosnippet.ParseError
	if !errors.As(err, &pe) {
		return err.Error()
	}
	var lines []string
	for _, p := range pe.Problems {
		lines = append(lines, fmt.Sprintf("%s:%d:%d: %s", b.File, b.OpenLine+p.Line, p.Col, p.Msg))
	}
	if pe.Extra > 0 {
		lines = append(lines, fmt.Sprintf("(and %d more)", pe.Extra))
	}
	return strings.Join(lines, "\n")
}

func describe(r result) string {
	switch r.status {
	case statusSkipped:
		return "skipped"
	case statusFailed:
		if r.known {
			return "known-failing (" + baseline.Reason(r.reason) + ")"
		}
		return r.stage + " failed"
	default:
		return "ok"
	}
}

func label(s status) string {
	switch s {
	case statusSkipped:
		return "skip"
	case statusFailed:
		return "fail"
	default:
		return "ok"
	}
}

func indent(s string) string {
	return strings.ReplaceAll(strings.TrimRight(s, "\n"), "\n", "\n    ")
}

// markdownFiles expands the given paths into a sorted list of .md files.
func markdownFiles(paths []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if !seen[p] {
				seen[p], out = true, append(out, p)
			}
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "node_modules" || strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) == ".md" && !seen[path] {
				seen[path], out = true, append(out, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(out)
	return out, nil
}

// repoRoot returns the given root, or the nearest ancestor of the working
// directory that looks like this repository.
func repoRoot(given string) (string, error) {
	if given != "" {
		return filepath.Abs(given)
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "articles")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no articles/ directory found; pass -root")
		}
		dir = parent
	}
}
