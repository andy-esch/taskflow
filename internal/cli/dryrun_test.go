package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// notExist asserts a path was NOT created/changed on disk by a dry run.
func notExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("--dry-run wrote to disk: %s exists", path)
	}
}

// noTaskFile asserts no flat id-led task file (tasks/<id>-<slug>.md) landed for
// slug — the dry-run equivalent of notExist under the flat layout, where the id is
// minted so the file can only be found by its slug suffix (ADR-0003 §4).
func noTaskFile(t *testing.T, root, slug string) {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(root, "tasks", "*-"+slug+".md"))
	if err != nil {
		t.Fatalf("glob task %q: %v", slug, err)
	}
	if len(m) != 0 {
		t.Errorf("--dry-run wrote to disk: task file(s) for %q exist: %v", slug, m)
	}
}

// runRootRC runs the root command and returns output + error (for exit-code
// assertions a would-fail dry run needs).
func runRootRC(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	// Execute() BEFORE reading out: `return out.String(), cmd.Execute()` evaluates
	// out.String() first (left-to-right), capturing the buffer pre-Execute (empty).
	err := cmd.Execute()
	return out.String(), err
}

// H4 (2026-06-22 audit): `task edit` is interactive and has no preview, so it
// must REJECT --dry-run rather than silently accept-and-ignore it (the flag would
// otherwise open an editor whose save lands on disk — the opposite of the flag).
func TestDryRun_TaskEditRejected(t *testing.T) {
	root := setupRepo(t)
	_, err := runRootRC(t, "-C", root, "task", "edit", "alpha", "--dry-run")
	if err == nil {
		t.Fatal("`task edit --dry-run` must be rejected, not silently accepted")
	}
	if !strings.Contains(err.Error(), "dry-run") {
		t.Errorf("the rejection should explain the no-preview contract, got %v", err)
	}
}

func TestDryRun_TaskNew(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: active\n---\n# E1\n")

	out := runRoot(t, "-C", root, "task", "new", "Preview Me", "--epic", "e1", "--tags", "a", "--dry-run")
	if !strings.Contains(out, "would create") {
		t.Errorf("dry-run should report the intended create, got %q", out)
	}
	noTaskFile(t, root, "preview-me")

	// JSON envelope carries dry_run and the would-be path.
	js := runRoot(t, "-C", root, "task", "new", "Json Preview", "--epic", "e1", "--tags", "a", "--dry-run", "--json")
	var env struct {
		DryRun  bool `json:"dry_run"`
		Created struct {
			ID, Status, Path string
		} `json:"created"`
	}
	if err := json.Unmarshal([]byte(js), &env); err != nil {
		t.Fatalf("dry-run --json invalid: %v\n%s", err, js)
	}
	if !env.DryRun || env.Created.ID != "json-preview" {
		t.Errorf("dry-run envelope wrong: %+v", env)
	}
	// status = the would-be task status (authoritative in frontmatter); the flat
	// path is tasks/<minted-id>-json-preview.md, relative to the planning root — no
	// status subdir (ADR-0003 §4). The id is minted, so match the id-led shape.
	if env.Created.Status != "ready-to-start" ||
		!strings.HasPrefix(env.Created.Path, "tasks/") ||
		!strings.HasSuffix(env.Created.Path, "-json-preview.md") ||
		strings.Contains(env.Created.Path, "ready-to-start/") {
		t.Errorf("dry-run envelope status/path wrong: %+v", env.Created)
	}
	noTaskFile(t, root, "json-preview")

	// A would-fail dry run still errors (unknown epic → validation, exit 11).
	if _, err := runRootRC(t, "-C", root, "task", "new", "Bad", "--epic", "ghost", "--tags", "a", "--dry-run"); err == nil || ExitCode(err) != 11 {
		t.Errorf("a would-fail dry-run must still error (exit 11), got %v", err)
	}
}

func TestDryRun_TaskMoveAndSet(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: active\n---\n# E1\n")
	runRoot(t, "-C", root, "task", "new", "Mover", "--epic", "e1", "--tags", "a")
	orig := taskPath(t, root, "mover")

	// Move (in-place under the flat layout): previews, the file stays put and its
	// frontmatter status is unchanged on disk (ADR-0003 §4 — status lives in the
	// file, the path never moves).
	before, _ := os.ReadFile(orig)
	out := runRoot(t, "-C", root, "task", "start", "mover", "--dry-run")
	if !strings.Contains(out, "would move") {
		t.Errorf("dry-run start should report the intended move, got %q", out)
	}
	if _, err := os.Stat(orig); err != nil {
		t.Error("dry-run move must leave the file at its original flat path")
	}
	if after, _ := os.ReadFile(orig); string(before) != string(after) {
		t.Error("dry-run move must not change the file's status on disk")
	}

	// Set: previews, frontmatter unchanged on disk.
	before, _ = os.ReadFile(orig)
	out = runRoot(t, "-C", root, "task", "set", "mover", "--priority", "high", "--dry-run")
	if !strings.Contains(out, "would update") {
		t.Errorf("dry-run set should report the intended update, got %q", out)
	}
	after, _ := os.ReadFile(orig)
	if string(before) != string(after) {
		t.Error("dry-run set must not modify the file")
	}

	// A would-fail set dry-run still errors (unknown epic).
	if _, err := runRootRC(t, "-C", root, "task", "set", "mover", "--epic", "ghost", "--dry-run"); err == nil || ExitCode(err) != 11 {
		t.Errorf("dry-run set to a bad epic must still error, got %v", err)
	}
}

func TestDryRun_EpicNewAndAuditMove(t *testing.T) {
	root := freshRepo(t)

	out := runRoot(t, "-C", root, "epic", "new", "Preview Epic", "--description", "d", "--dry-run")
	if !strings.Contains(out, "would create") {
		t.Errorf("dry-run epic new should preview, got %q", out)
	}
	// No epic .md file written (the would-be id is NN-preview-epic). The dir holds
	// init's .gitkeep, so assert specifically that no markdown landed.
	entries, _ := os.ReadDir(filepath.Join(root, "epics"))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			t.Errorf("dry-run epic new must not write a file, found %s", e.Name())
		}
	}

	// Audit move: seed an open audit, dry-run close it.
	mustWrite(t, filepath.Join(root, "audits", "open", "2026-06-01-x.md"), "---\narea: store\n---\n# A\n")
	out = runRoot(t, "-C", root, "audit", "close", "2026-06-01-x", "--dry-run")
	if !strings.Contains(out, "would move") {
		t.Errorf("dry-run audit close should preview, got %q", out)
	}
	if _, err := os.Stat(filepath.Join(root, "audits", "open", "2026-06-01-x.md")); err != nil {
		t.Error("dry-run audit close must leave the file in open/")
	}
	notExist(t, filepath.Join(root, "audits", "closed", "2026-06-01-x.md"))
}

func TestDryRun_Init(t *testing.T) {
	root := t.TempDir()
	out := runRoot(t, "init", "--path", root, "--dry-run")
	if !strings.Contains(out, "would initialize") {
		t.Errorf("dry-run init should preview, got %q", out)
	}
	if _, err := os.Stat(filepath.Join(root, ".tskflwctl.toml")); err == nil {
		t.Error("dry-run init must not write the config")
	}
	if _, err := os.Stat(filepath.Join(root, "tasks")); err == nil {
		t.Error("dry-run init must not create the tree")
	}
}
