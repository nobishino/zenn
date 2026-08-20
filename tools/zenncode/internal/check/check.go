// Package check compiles normalized snippets in a throwaway module.
//
// Builds are hermetic: GOPROXY is off and no go.sum is written, so a snippet
// can only use the standard library. That matches the Go Playground closely
// enough for the sample code in this repository and keeps verification usable
// offline.
package check

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

// Request is one snippet to compile.
type Request struct {
	Program   string // complete program, as produced by gosnippet.Normalize
	IsMain    bool   // package main with a func main
	GoVersion string // go directive for the scratch module; defaults to the toolchain's
	Binary    bool   // keep the executable so it can be run
	GOOS      string // cross-compilation target; empty means the host's
	GOARCH    string
}

// Result is the outcome of a build.
type Result struct {
	OK     bool
	Output string // compiler diagnostics, with scratch paths stripped
	Bin    string // the executable, when Request.Binary was set and OK
}

// RunResult is the outcome of running a built program.
type RunResult struct {
	OK       bool
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
}

// Builder compiles snippets under a shared temporary directory.
// A Builder is safe for concurrent use.
type Builder struct {
	root      string
	goVersion string
	n         atomic.Int64
}

// NewBuilder creates a Builder. Close must be called to remove its scratch
// directory.
func NewBuilder() (*Builder, error) {
	root, err := os.MkdirTemp("", "zenncode-")
	if err != nil {
		return nil, err
	}
	// On macOS the temp directory is reached through a symlink, and a panic
	// trace names the resolved path. Resolve it now so the two agree and
	// scratch paths can be stripped from diagnostics.
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	b := &Builder{root: root, goVersion: toolchainVersion()}
	return b, nil
}

// Close removes the Builder's scratch directory.
func (b *Builder) Close() error { return os.RemoveAll(b.root) }

// GoVersion reports the go directive used when a Request does not name one.
func (b *Builder) GoVersion() string { return b.goVersion }

// Build compiles req and reports whether it succeeded.
// The returned error is non-nil only when the build could not be run at all.
func (b *Builder) Build(ctx context.Context, req Request) (Result, error) {
	dir := filepath.Join(b.root, fmt.Sprintf("b%04d", b.n.Add(1)))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Result{}, err
	}

	version := req.GoVersion
	if version == "" {
		version = b.goVersion
	}
	gomod := fmt.Sprintf("module zenncodesample\n\ngo %s\n", version)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, "sample.go"), []byte(req.Program), 0o644); err != nil {
		return Result{}, err
	}

	args := []string{"build"}
	bin := os.DevNull
	if req.IsMain {
		// For a library build there is nothing to write and -o would be
		// rejected, so only a main package names an output.
		if req.Binary {
			bin = filepath.Join(dir, "prog")
		}
		args = append(args, "-o", bin)
	}
	args = append(args, ".")

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GOPROXY=off",
		"GOFLAGS=-mod=mod",
		"GOWORK=off",
		"GOTOOLCHAIN=local",
	)
	// A snippet written for wasm only compiles under its own GOOS/GOARCH:
	// on the host the build constraints exclude it, or a //go:wasmimport
	// declaration is rejected for having no body.
	if req.GOOS != "" {
		cmd.Env = append(cmd.Env, "GOOS="+req.GOOS)
	}
	if req.GOARCH != "" {
		cmd.Env = append(cmd.Env, "GOARCH="+req.GOARCH)
	}
	out, err := cmd.CombinedOutput()
	res := Result{OK: err == nil, Output: clean(string(out), dir)}
	if res.OK && req.Binary && req.IsMain {
		res.Bin = bin
	}
	if err != nil && res.Output == "" {
		res.Output = err.Error()
	}
	if ctx.Err() != nil {
		return res, ctx.Err()
	}
	return res, nil
}

// Run executes a program built with Request.Binary and captures its output.
// A program that outlives timeout is killed and reported as timed out, which
// is the normal fate of a sample that deadlocks on purpose.
func Run(ctx context.Context, bin string, timeout time.Duration) (RunResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var stdout, stderr strings.Builder
	cmd := exec.CommandContext(ctx, bin)
	cmd.Dir = filepath.Dir(bin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	dir := filepath.Dir(bin)
	res := RunResult{
		Stdout: stdout.String(),
		// A panic trace names the scratch file it came from; strip the
		// path so the message reads as if it came from the article.
		Stderr: clean(stderr.String(), dir),
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		res.TimedOut = true
		return res, nil
	}
	var exit *exec.ExitError
	switch {
	case err == nil:
		res.OK = true
	case errors.As(err, &exit):
		res.ExitCode = exit.ExitCode()
	default:
		return res, err
	}
	return res, nil
}

// clean strips scratch paths so diagnostics read as if they came from the
// article itself.
func clean(out, dir string) string {
	out = strings.ReplaceAll(out, dir+string(filepath.Separator), "")
	out = strings.ReplaceAll(out, dir, ".")
	out = strings.ReplaceAll(out, "# zenncodesample\n", "")
	out = strings.ReplaceAll(out, "./sample.go", "sample.go")
	return strings.TrimRight(out, "\n")
}

// toolchainVersion returns the go directive to use by default, e.g. "1.24.0".
// The same toolchain runs the builds, so its own version is always usable.
func toolchainVersion() string {
	out, err := exec.Command("go", "env", "GOVERSION").Output()
	if err != nil {
		return "1.21"
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "go")
}
