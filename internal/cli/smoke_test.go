package cli

import (
	"encoding/json"
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binPath is the real binary built once in TestMain; buildErr explains an empty
// path. The subprocess smokes are the only tests that exercise main.go wiring —
// os.Exit codes, the ldflags --version, and the JSON error envelope, which
// cli.WriteError emits from main, not from inside Execute().
var (
	binPath  string
	buildErr string
)

func TestMain(m *testing.M) {
	flag.Parse() // so testing.Short() is readable before m.Run()
	var cleanup func()
	if !testing.Short() {
		if dir, err := os.MkdirTemp("", "tskflwctl-smoke"); err == nil {
			cleanup = func() { _ = os.RemoveAll(dir) }
			bp := filepath.Join(dir, "tskflwctl")
			if out, berr := exec.Command("go", "build", "-o", bp, "github.com/andy-esch/taskflow/cmd/tskflwctl").CombinedOutput(); berr != nil {
				buildErr = berr.Error() + ": " + string(out)
			} else {
				binPath = bp
			}
		}
	}
	code := m.Run()
	if cleanup != nil {
		cleanup()
	}
	os.Exit(code)
}

func smokeBin(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("subprocess smoke skipped under -short")
	}
	if binPath == "" {
		t.Skipf("binary not built: %s", buildErr)
	}
	return binPath
}

// --version comes from ldflags via main (not deterministic) — assert shape, not bytes.
func TestSmoke_Version(t *testing.T) {
	out, err := exec.Command(smokeBin(t), "version").CombinedOutput()
	if err != nil {
		t.Fatalf("version exited non-zero: %v\n%s", err, out)
	}
	if !strings.HasPrefix(string(out), "tskflwctl ") {
		t.Errorf("version output unexpected: %q", out)
	}
}

// A not-found resolves to exit code 10 — the agent-routable contract, which only
// the real binary (os.Exit in main) actually produces.
func TestSmoke_NotFoundExitCode(t *testing.T) {
	cmd := exec.Command(smokeBin(t), "-C", fixtureRepo, "task", "show", "no-such-task")
	out, err := cmd.CombinedOutput()
	ee, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected a non-zero exit, got %v\n%s", err, out)
	}
	if ee.ExitCode() != 10 {
		t.Errorf("not-found should exit 10, got %d\n%s", ee.ExitCode(), out)
	}
}

// Under --json a failure is a parseable error envelope on stderr with a stable
// code — produced by main.go's WriteError, invisible to the in-process tests.
func TestSmoke_JSONErrorEnvelope(t *testing.T) {
	cmd := exec.Command(smokeBin(t), "-C", fixtureRepo, "--json", "task", "show", "no-such-task")
	out, err := cmd.CombinedOutput()
	ee, ok := err.(*exec.ExitError)
	if !ok || ee.ExitCode() != 10 {
		t.Fatalf("expected exit 10, got %v\n%s", err, out)
	}
	var env struct {
		SchemaVersion string `json:"schema_version"`
		Error         struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if jerr := json.Unmarshal(out, &env); jerr != nil {
		t.Fatalf("--json failure should be a parseable envelope: %v\n%s", jerr, out)
	}
	if env.SchemaVersion == "" || env.Error.Code != "not-found" {
		t.Errorf("error envelope wrong (want code not-found):\n%s", out)
	}
}

// A clean end-to-end read through the real binary against the committed fixture.
func TestSmoke_ListJSON(t *testing.T) {
	out, err := exec.Command(smokeBin(t), "-C", fixtureRepo, "task", "list", "--all", "--json").Output()
	if err != nil {
		t.Fatalf("list --json exited non-zero: %v", err)
	}
	if !strings.Contains(string(out), `"schema_version"`) || !strings.Contains(string(out), "alpha-task") {
		t.Errorf("list --json unexpected:\n%s", out)
	}
}

// The create→start→complete lifecycle through the real binary on an isolated copy
// of the fixture: each step exits 0 and the file ends up in completed/ — the
// noun-verb wiring and atomic moves working end to end (file state, not bytes, so
// the today-stamped dates don't make it flaky).
func TestSmoke_Lifecycle(t *testing.T) {
	bin := smokeBin(t)
	repo := t.TempDir()
	copyTree(t, fixtureRepo, repo)
	run := func(args ...string) (string, int) {
		out, err := exec.Command(bin, append([]string{"-C", repo}, args...)...).CombinedOutput()
		if ee, ok := err.(*exec.ExitError); ok {
			return string(out), ee.ExitCode()
		} else if err != nil {
			t.Fatalf("exec %v: %v", args, err)
		}
		return string(out), 0
	}
	const slug = "smoke-lifecycle-task"
	for _, step := range [][]string{
		{"task", "new", "Smoke Lifecycle Task", "--epic", "01-fixture-epic", "--tags", "x"},
		{"task", "start", slug},
		{"task", "complete", slug},
	} {
		if out, code := run(step...); code != 0 {
			t.Fatalf("`%s` exited %d:\n%s", strings.Join(step, " "), code, out)
		}
	}
	if _, err := os.Stat(filepath.Join(repo, "tasks", "completed", slug+".md")); err != nil {
		t.Errorf("task should have moved to completed/ through the binary: %v", err)
	}
}

// copyTree copies the committed fixture into a writable temp repo so a mutation
// sequence runs isolated (never touching the committed testdata).
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	if err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	}); err != nil {
		t.Fatal(err)
	}
}
