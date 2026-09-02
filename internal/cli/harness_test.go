package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// The in-process CLI harness. Stream discipline is a stated contract — payload on
// stdout, diagnostics on stderr (see problems.go) — so the harness must be able to
// tell them apart. A single shared buffer cannot: with one writer behind both
// streams, moving a diagnostic from ErrOut to Out is invisible to every assertion
// in the package, and the `-o name` / `--json` consumers that contract exists for
// break silently.
//
// runResult therefore keeps both streams separate AND keeps a faithfully
// interleaved merged view, because most tests legitimately don't care which
// stream a line arrived on and shouldn't be forced to.

// runResult is one root-command execution's captured output.
type runResult struct {
	Out    string // stdout: the payload `-o name`/`--json` consumers parse
	Err    string // stderr: diagnostics, per problems.go
	Merged string // both, in the order they were actually written
}

// runRootStreams executes the root command in-process against args and returns
// its streams separately. It is the primitive the other helpers are built on;
// reach for it directly when a test's claim is about WHICH stream output landed
// on. The returned error is the command's, so callers can assert on exit codes.
func runRootStreams(t *testing.T, args ...string) (runResult, error) {
	t.Helper()
	return runRootStreamsIn(t, strings.NewReader(""), args...)
}

// runRootStreamsIn is runRootStreams with stdin supplied, for the commands that
// read a body from it.
func runRootStreamsIn(t *testing.T, stdin io.Reader, args ...string) (runResult, error) {
	t.Helper()
	// Each stream gets its own buffer and a tee into a shared one, so Merged
	// preserves true interleaving rather than concatenating one stream after the
	// other — a test that asserts a diagnostic appears between two payload lines
	// keeps working.
	var out, errOut, merged bytes.Buffer
	stdout := io.MultiWriter(&out, &merged)
	stderr := io.MultiWriter(&errOut, &merged)

	cmd := NewRootCmd(stdin, stdout, stderr)
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	// Execute BEFORE reading the buffers: `return result, cmd.Execute()` would
	// evaluate the buffers first (left-to-right) and capture them pre-Execute.
	err := cmd.Execute()
	return runResult{Out: out.String(), Err: errOut.String(), Merged: merged.String()}, err
}

// runRoot executes the root command and returns its merged output, failing the
// test if the command errored. The default for the many assertions that only care
// that some text was produced; use runRootStreams when the stream matters.
func runRoot(t *testing.T, args ...string) string {
	t.Helper()
	res, err := runRootStreams(t, args...)
	if err != nil {
		t.Fatalf("execute %v: %v\noutput:\n%s", args, err, res.Merged)
	}
	return res.Merged
}

// runRootRC executes the root command and returns its merged output plus the
// error, for tests asserting on failure and exit codes.
func runRootRC(t *testing.T, args ...string) (string, error) {
	t.Helper()
	res, err := runRootStreams(t, args...)
	return res.Merged, err
}
