package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// buildBinary compiles the real tskflwctl once per test run. Everything else
// in the suite tests packages in-process; only here are the os.Exit wiring and
// the semantic exit codes (10–14) exercised through an actual process.
var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

func binary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "tskflwctl-smoke-*")
		if err != nil {
			buildErr = err
			return
		}
		binPath = filepath.Join(dir, "tskflwctl")
		out, err := exec.Command("go", "build", "-o", binPath, ".").CombinedOutput()
		if err != nil {
			buildErr = err
			t.Logf("build output:\n%s", out)
		}
	})
	if buildErr != nil {
		t.Fatalf("build binary: %v", buildErr)
	}
	return binPath
}

// run executes the binary against root, returning combined output and the real
// process exit code.
func run(t *testing.T, root string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(binary(t), append([]string{"-C", root}, args...)...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	t.Fatalf("run %v: %v\n%s", args, err, out)
	return "", -1
}

func TestSmoke_LifecycleAndExitCodes(t *testing.T) {
	root := t.TempDir()

	// init scaffolds the tree.
	if out, code := run(t, root, "init", "--path", root); code != 0 {
		t.Fatalf("init: exit %d\n%s", code, out)
	}
	if out, code := run(t, root, "epic", "new", "Smoke Epic", "--description", "smoke"); code != 0 {
		t.Fatalf("epic new: exit %d\n%s", code, out)
	}
	out, code := run(t, root, "task", "new", "Smoke Task",
		"--epic", "01-smoke-epic", "--description", "smoke", "--tags", "smoke", "--json")
	if code != 0 {
		t.Fatalf("task new: exit %d\n%s", code, out)
	}
	var created struct {
		SchemaVersion string `json:"schema_version"`
		Created       struct {
			ID string `json:"id"`
		} `json:"created"`
	}
	if err := json.Unmarshal([]byte(out), &created); err != nil || created.Created.ID == "" {
		t.Fatalf("task new --json should return the resolved id: %v\n%s", err, out)
	}
	slug := created.Created.ID

	// Lifecycle: start → complete, then lint must be clean.
	if out, code := run(t, root, "task", "start", slug); code != 0 {
		t.Fatalf("task start: exit %d\n%s", code, out)
	}
	if out, code := run(t, root, "task", "complete", slug); code != 0 {
		t.Fatalf("task complete: exit %d\n%s", code, out)
	}
	if out, code := run(t, root, "lint"); code != 0 {
		t.Fatalf("lint after lifecycle: exit %d\n%s", code, out)
	}

	// Semantic exit codes through a real process exit.
	if out, code := run(t, root, "task", "show", "no-such-task"); code != 10 {
		t.Errorf("not-found should exit 10, got %d\n%s", code, out)
	}
	if out, code := run(t, root, "task", "list", "--status", "bogus"); code != 11 {
		t.Errorf("invalid status filter should exit 11, got %d\n%s", code, out)
	}
	if out, code := run(t, root, "task", "move", slug, "limbo"); code != 11 {
		t.Errorf("invalid move target should exit 11, got %d\n%s", code, out)
	}

	// Errors go to stderr prefixed "error:" (main's contract).
	cmd := exec.Command(binary(t), "-C", root, "task", "show", "ghost")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	_ = cmd.Run()
	if !strings.Contains(stderr.String(), "error:") {
		t.Errorf("errors should print to stderr with the error: prefix, got %q", stderr.String())
	}

	// Under --json, the failure is a machine-readable envelope on stderr and
	// stdout stays empty — agents must never parse prose (schema 1.1).
	cmd = exec.Command(binary(t), "-C", root, "task", "show", "ghost", "--json")
	var jsonOut, jsonErr strings.Builder
	cmd.Stdout = &jsonOut
	cmd.Stderr = &jsonErr
	_ = cmd.Run()
	if jsonOut.Len() != 0 {
		t.Errorf("stdout must stay empty on a --json failure, got %q", jsonOut.String())
	}
	var env struct {
		SchemaVersion string `json:"schema_version"`
		Error         struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(jsonErr.String()), &env); err != nil {
		t.Fatalf("--json error should be a JSON envelope: %v\n%s", err, jsonErr.String())
	}
	if env.SchemaVersion == "" || env.Error.Code != "not-found" || env.Error.Message == "" {
		t.Errorf("error envelope wrong: %+v", env)
	}
}

func TestSmoke_VersionStamp(t *testing.T) {
	out, code := run(t, t.TempDir(), "version")
	if code != 0 || strings.TrimSpace(out) == "" {
		t.Errorf("version: exit %d output %q", code, out)
	}
}
