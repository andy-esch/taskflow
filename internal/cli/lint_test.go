package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
)

// TestLint_EpicSoleFailure pins the end-to-end epic-lint path: a clean task plus
// an active epic missing its required fields makes `lint` exit 11 and name the
// epic + its issue. Epics are report-only `results` (never `problems`), so this is
// the path Fix 1 keys `lint --fix`'s exit off — the CLI seam must surface it.
func TestLint_EpicSoleFailure(t *testing.T) {
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
	// The task is fully lint-clean; the epic is the SOLE failure (active, but missing
	// the required priority + description).
	write("epics/01-e1.md", "---\nstatus: active\n---\n# E1\n")
	goodPath, goodOut := testutil.TaskFixture(root, "ready-to-start", "good.md",
		"---\nid: "+testutil.TaskID("good")+"\nstatus: ready-to-start\nepic: 01-e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-01\ntags: [a]\n---\n# Good\n")
	testutil.Write(t, goodPath, goodOut)

	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "lint"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("an epic missing required fields should fail lint")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11 for a failing epic, got %d", ExitCode(err))
	}
	o := out.String()
	if !strings.Contains(o, "e1") {
		t.Errorf("lint should name the failing epic id:\n%s", o)
	}
	if !strings.Contains(o, "priority") && !strings.Contains(o, "description") {
		t.Errorf("lint should name the epic's missing field:\n%s", o)
	}
}

func TestLint_Clean(t *testing.T) {
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
	// The epic must itself be lint-clean now (status + priority + description),
	// or `lint` would flag it and never report a pass.
	write("epics/01-e1.md", "---\nstatus: active\npriority: high\ndescription: the epic\n---\n# E1\n")
	goodPath, goodOut := testutil.TaskFixture(root, "ready-to-start", "good.md",
		"---\nid: "+testutil.TaskID("good")+"\nstatus: ready-to-start\nepic: 01-e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-01\ntags: [a]\n---\n# Good\n")
	testutil.Write(t, goodPath, goodOut)

	out := runRoot(t, "-C", root, "lint")
	if !strings.Contains(out, "pass lint") {
		t.Errorf("expected pass, got: %q", out)
	}
}

// TestLint_FlagsNonNNEpicFailOpen pins the epic NN- gate end-to-end: a non-NN-<slug>
// epic is lint-flagged (exit 11, naming the convention) yet STILL lists/resolves — the
// fail-open contract, not a dropped FileProblem.
func TestLint_FlagsNonNNEpicFailOpen(t *testing.T) {
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
	write("tasks/.gitkeep", "") // anchor discovery
	// Otherwise lint-clean (status/priority/description) so the NN- name is the sole issue.
	write("epics/legacy-epic.md", "---\nstatus: active\npriority: high\ndescription: the goal\n---\n# Legacy\n")

	out, err := runRootRC(t, "-C", root, "lint")
	if err == nil || ExitCode(err) != 11 {
		t.Fatalf("a non-NN epic must fail lint (exit 11), got %v", err)
	}
	if !strings.Contains(out, "legacy-epic") || !strings.Contains(out, "NN-") {
		t.Errorf("lint should name the epic and the NN- convention:\n%s", out)
	}
	// Fail-open: still lists/resolves despite the flag.
	if list := runRoot(t, "-C", root, "epic", "list", "-q"); !strings.Contains(list, "legacy-epic") {
		t.Errorf("a non-NN epic must still list (fail-open), got:\n%s", list)
	}
}

func TestLint_Dirty_Exit11(t *testing.T) {
	// setupRepo's tasks have only status+description → missing required fields.
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "lint"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected lint issues")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit code 11, got %d", ExitCode(err))
	}
	if !strings.Contains(out.String(), "issues") {
		t.Errorf("expected an issues report, got: %q", out.String())
	}
}

// A task with NO frontmatter status still lists (raw status) but is flagged —
// frontmatter is authoritative under the flat layout, so a missing/unrecognized
// status is a real defect (StatusFellBack), not a silent folder fallback.
func TestLint_FlagsMissingFrontmatterStatus(t *testing.T) {
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
	write("epics/01-e1.md", "---\nstatus: active\npriority: high\ndescription: e\n---\n# E1\n")
	// A raw flat id-led path with NO status in frontmatter (TaskFixture would inject
	// one; here the missing status is the whole point).
	xPath := filepath.Join(root, "tasks", testutil.TaskID("x")+"-x.md")
	testutil.Write(t,
		xPath,
		"---\nid: 6fjangd7kvca\nepic: 01-e1\ntier: 3\npriority: high\neffort: x\ncreated: 2026-06-09\ntags: [a]\ndescription: d\n---\n# x\n")

	out, err := runRootRC(t, "-C", root, "lint")
	if err == nil {
		t.Error("lint must exit non-zero on a missing frontmatter status")
	}
	if !strings.Contains(out, "frontmatter status missing or unrecognized") {
		t.Errorf("expected the missing-status flag:\n%s", out)
	}
}

// TestLint_FlagsArchivedTaskIDDrift verifies that archived tasks with a drifted ID
// in the frontmatter vs the filename are flagged by lint.
func TestLint_FlagsArchivedTaskIDDrift(t *testing.T) {
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
	write("epics/01-e1.md", "---\nstatus: active\npriority: high\ndescription: e\n---\n# E1\n")

	tID := testutil.TaskID("drifted")
	xPath := filepath.Join(root, "tasks", tID+"-drifted.md")
	testutil.Write(t,
		xPath,
		"---\nid: differentid12\nstatus: completed\nepic: 01-e1\ntier: 3\npriority: high\neffort: x\ncreated: 2026-06-09\ntags: [a]\ndescription: d\n---\n# drifted\n")

	out, err := runRootRC(t, "-C", root, "lint")
	if err == nil {
		t.Error("lint must exit non-zero on a drifted task ID")
	}
	if !strings.Contains(out, "disagrees with the filename id") {
		t.Errorf("expected the id-drift flag:\n%s", out)
	}
}

// TestLint_LinksFlagsDangler pins the opt-in Scheme-2 dangler check: plain `lint` ignores
// body links, but `lint --links` flags a markdown link whose target file is missing (exit
// 11) while leaving external and resolvable links alone.
func TestLint_LinksFlagsDangler(t *testing.T) {
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
	write("epics/01-e.md", "---\nstatus: active\npriority: high\ndescription: e\n---\n# E\n")
	p, out := testutil.TaskFixture(root, "ready-to-start", "a.md",
		"---\nid: "+testutil.TaskID("a")+"\nstatus: ready-to-start\nepic: 01-e\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\ndescription: d\n---\n# A\n\nDead: [gone](6fjangd7kvzz-gone.md). Ext: [x](https://e.com/x.md).\n")
	testutil.Write(t, p, out)

	// Plain lint passes — body links aren't checked by default (routines stay clean).
	if _, err := runRootRC(t, "-C", root, "lint"); err != nil {
		t.Fatalf("plain lint should pass (danglers not checked): %v", err)
	}
	// --links flags the missing local link (exit 11), not the external one.
	o, err := runRootRC(t, "-C", root, "lint", "--links")
	if err == nil || ExitCode(err) != 11 {
		t.Fatalf("lint --links should flag the dangler (exit 11), got %v", err)
	}
	if !strings.Contains(o, "6fjangd7kvzz-gone.md") || strings.Contains(o, "e.com") {
		t.Errorf("should flag only the missing local link:\n%s", o)
	}
}
