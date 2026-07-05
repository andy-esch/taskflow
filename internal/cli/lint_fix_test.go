package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestLintFix_DryRunThenFix(t *testing.T) {
	root := t.TempDir()
	// The only issues are the two FIXABLE ones (an unquoted ':' in description, a
	// comma-joined tags list); every required field is present, so the post-fix
	// re-lint is clean and `lint --fix` exits 0 (Fix 1 keys the exit off the leftover
	// findings — a tree the fixer fully repairs must still come back green).
	bad, badContent := testutil.TaskFixture(root, "ready-to-start", "bad.md",
		"---\nstatus: ready-to-start\nepic: 01-e1\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ndescription: A: B\ntags: x,y\n---\n# Bad\n")
	testutil.Write(t, bad, badContent)
	epic := filepath.Join(root, "epics", "01-e1.md")
	if err := os.MkdirAll(filepath.Dir(epic), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(epic, []byte("---\nstatus: active\npriority: high\ndescription: the epic\n---\n# E1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// dry-run: reports, doesn't write.
	out := runRoot(t, "-C", root, "lint", "--fix", "--dry-run")
	if !strings.Contains(out, "would fix") {
		t.Errorf("expected a dry-run report: %q", out)
	}
	if raw, _ := os.ReadFile(bad); !strings.Contains(string(raw), "description: A: B") {
		t.Error("dry-run modified the file")
	}

	// real fix: writes; the file becomes readable and the tree comes back lint-clean.
	if out := runRoot(t, "-C", root, "lint", "--fix"); !strings.Contains(out, "fixed") {
		t.Errorf("expected a fix report: %q", out)
	}
	if listOut := runRoot(t, "-C", root, "task", "list"); !strings.Contains(listOut, "bad") {
		t.Errorf("task should be readable after fix: %q", listOut)
	}
}

// TestLintFix_BackfillsMissingID pins the ADR-0003 backfill: plain lint flags a
// pre-id task, `--fix` mints one from its created date, and the tree comes back
// clean with the id visible in --json.
func TestLintFix_BackfillsMissingID(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("epics/01-e1.md", "---\nstatus: active\npriority: high\ndescription: the epic\n---\n# E1\n")
	// Fully valid EXCEPT it predates ids (no id: field). It lands at a flat id-led
	// path via its filename slug, but carries no id: field to backfill.
	tPath, tContent := testutil.TaskFixture(root, "ready-to-start", "t.md",
		"---\nstatus: ready-to-start\nepic: 01-e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-05\ntags: [a]\n---\n# T\n")
	testutil.Write(t, tPath, tContent)

	// Plain lint flags the missing id (and exits non-zero).
	out, err := runRootRC(t, "-C", root, "lint")
	if err == nil {
		t.Error("plain lint must flag a missing id")
	}
	if !strings.Contains(out, "missing stable id") {
		t.Errorf("expected a missing-id finding, got: %q", out)
	}

	// --fix backfills it.
	if fixOut := runRoot(t, "-C", root, "lint", "--fix"); !strings.Contains(fixOut, "fixed") {
		t.Errorf("expected a fix report: %q", fixOut)
	}

	// The task now carries an id, visible in --json.
	var got struct {
		Task struct {
			ID string `json:"id"`
		} `json:"task"`
	}
	showOut := runRoot(t, "-C", root, "task", "show", "t", "--json")
	if err := json.Unmarshal([]byte(showOut), &got); err != nil {
		t.Fatalf("task show --json invalid: %v\n%s", err, showOut)
	}
	if got.Task.ID == "" {
		t.Errorf("id should be backfilled, got empty:\n%s", showOut)
	}

	// And plain lint is clean again.
	if _, err := runRootRC(t, "-C", root, "lint"); err != nil {
		t.Errorf("lint should pass after the backfill, got: %v", err)
	}
}

// TestLintFix_BackfillsDatelessFromFilename: a completed, dateless, non-date-prefixed
// task was "unrepairable" pre-flatten (nothing to mint an id from). Post-flatten the id
// already leads the flat filename, so `--fix` backfills it from there and the tree lints
// clean — no date required, and no "could not auto-repair" leftovers.
func TestLintFix_BackfillsDatelessFromFilename(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("epics/01-e1.md", "---\nstatus: active\npriority: high\ndescription: the epic\n---\n# E1\n")
	ndPath, ndContent := testutil.TaskFixture(root, "completed", "nodate.md",
		"---\nstatus: completed\nepic: 01-e1\n---\n# ND\n")
	testutil.Write(t, ndPath, ndContent)

	out, err := runRootRC(t, "-C", root, "lint", "--fix")
	if err != nil {
		t.Fatalf("lint --fix should repair a dateless task from its filename, got err %v\n%s", err, out)
	}
	if strings.Contains(out, "could not auto-repair") {
		t.Errorf("nothing should be left unrepaired:\n%s", out)
	}
	// The backfilled id is exactly the one that already led the flat filename.
	if b, _ := os.ReadFile(ndPath); !strings.Contains(string(b), "id: "+testutil.TaskID("nodate")) {
		t.Errorf("id should be backfilled from the filename:\n%s", b)
	}
}

// TestLintFix_InvalidFrontmatterFailsLoud pins the contract: a task file with no
// `---` block fails loudly (the file is named, the valid shape is described, and
// `schema task` is pointed to) rather than being misreported as a fixable "missing
// id" — `lint --fix` must not attempt to synthesize frontmatter for it.
func TestLintFix_InvalidFrontmatterFailsLoud(t *testing.T) {
	root := t.TempDir()
	// A fence-less body stays a loud FileProblem; it lands at a flat id-led path.
	p := filepath.Join(root, "tasks", testutil.TaskID("no-fence")+"-no-fence.md")
	testutil.Write(t, p, "# Just a heading\n\nnotes\n")

	out, err := runRootRC(t, "-C", root, "lint", "--fix")
	if err == nil || ExitCode(err) != 11 {
		t.Fatalf("want exit 11 for invalid frontmatter, got %v", err)
	}
	if !strings.Contains(out, "no-fence.md") || !strings.Contains(out, "missing frontmatter") || !strings.Contains(out, "schema task") {
		t.Errorf("expected a loud, shape-naming problem:\n%s", out)
	}
	if strings.Contains(out, "missing stable id") || strings.Contains(out, "no date to mint") {
		t.Errorf("invalid frontmatter must not be misreported as a missing id:\n%s", out)
	}
}

// TestLintFix_UnrepairableFileExitsNonZero pins B4: a file the fixer can't
// repair must be surfaced with a non-zero exit — `lint --fix` previously said
// nothing and exited 0, leaving the tree broken while claiming success.
func TestLintFix_UnrepairableFileExitsNonZero(t *testing.T) {
	root := t.TempDir()
	// Unterminated frontmatter: nothing the text fixer can do with it; flat id-led path.
	broken := filepath.Join(root, "tasks", testutil.TaskID("broken")+"-broken.md")
	testutil.Write(t, broken, "---\nstatus: ready-to-start\n# no closing fence\n")

	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "lint", "--fix"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("lint --fix must fail when a file remains unrepairable")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11, got %d (%v)", ExitCode(err), err)
	}
	if !strings.Contains(out.String(), "broken.md") {
		t.Errorf("the unrepairable file should be named in the output:\n%s", out.String())
	}
	// --dry-run stays exit 0 (it promises nothing about the result).
	out.Reset()
	dry := NewRootCmd(strings.NewReader(""), &out, &out)
	dry.SetArgs([]string{"-C", root, "lint", "--fix", "--dry-run"})
	dry.SetOut(&out)
	dry.SetErr(&out)
	if err := dry.Execute(); err != nil {
		t.Errorf("dry-run should not fail on unrepairable files: %v", err)
	}
}

// TestLintFix_ReportOnlyEpicExitsNonZero pins Fix 1: `lint --fix` on a tree whose
// ONLY issue is a report-only epic (an active epic missing required fields — never
// auto-fixed, surfaced as a `result`, never a `problem`) must still exit 11 and
// name the epic. The post-fix re-lint previously discarded `results`, so this tree
// exited 0 in a false green: the fixer touched nothing, the unreadable list was
// empty, and the leftover epic finding was dropped.
func TestLintFix_ReportOnlyEpicExitsNonZero(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A clean task plus an active epic missing priority/description — the epic is the
	// sole, fix-immune failure.
	write("epics/01-e1.md", "---\nstatus: active\n---\n# E1\n")
	goodPath, goodContent := testutil.TaskFixture(root, "ready-to-start", "good.md",
		"---\nstatus: ready-to-start\nepic: 01-e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-01\ntags: [a]\n---\n# Good\n")
	testutil.Write(t, goodPath, goodContent)

	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "lint", "--fix"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("lint --fix must fail when a report-only epic remains broken")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11, got %d (%v)", ExitCode(err), err)
	}
	o := out.String()
	if !strings.Contains(o, "e1") {
		t.Errorf("the leftover epic should be named in the output:\n%s", o)
	}
	if !strings.Contains(o, "could not auto-repair") {
		t.Errorf("the human output should flag what --fix could not repair:\n%s", o)
	}

	// --json: the leftover epic finding must land in `remaining`, not only in the
	// prose error — a --json consumer must never parse prose to learn it stayed broken.
	out.Reset()
	jc := NewRootCmd(strings.NewReader(""), &out, &out)
	jc.SetArgs([]string{"-C", root, "lint", "--fix", "--json"})
	jc.SetOut(&out)
	jc.SetErr(&out)
	if err := jc.Execute(); err == nil || ExitCode(err) != 11 {
		t.Fatalf("want exit 11 for a report-only epic, got %v", err)
	}
	var env struct {
		Remaining []struct {
			Slug   string `json:"slug"`
			Issues []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			} `json:"issues"`
		} `json:"remaining"`
	}
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("fix --json invalid: %v\n%s", err, out.String())
	}
	if len(env.Remaining) != 1 || env.Remaining[0].Slug != "01-e1" || len(env.Remaining[0].Issues) == 0 {
		t.Errorf("remaining should carry the epic 01-e1 + its issues, got %+v", env.Remaining)
	}
}

// TestLintFix_JSONReportsUnreadable pins the --json contract: an unrepairable
// file must appear in the fix report's `unreadable` array, not only as a count
// in the prose error — a --json consumer must never parse prose to learn what
// stayed broken.
func TestLintFix_JSONReportsUnreadable(t *testing.T) {
	root := t.TempDir()
	// Unterminated frontmatter: unrepairable; flat id-led path.
	broken := filepath.Join(root, "tasks", testutil.TaskID("broken")+"-broken.md")
	testutil.Write(t, broken, "---\nstatus: ready-to-start\n# no closing fence\n")

	// stdout carries the fix envelope; the error is returned (not written, since
	// the root silences errors), so the buffer holds only the JSON report.
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "lint", "--fix", "--json"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err == nil || ExitCode(err) != 11 {
		t.Fatalf("want exit 11 for an unrepairable file, got %v", err)
	}

	var env struct {
		DryRun     bool `json:"dry_run"`
		Unreadable []struct {
			Path    string `json:"path"`
			Message string `json:"message"`
		} `json:"unreadable"`
	}
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("fix --json invalid: %v\n%s", err, out.String())
	}
	if len(env.Unreadable) != 1 || !strings.Contains(env.Unreadable[0].Path, "broken.md") {
		t.Errorf("unreadable array should name broken.md, got %+v", env.Unreadable)
	}
}
