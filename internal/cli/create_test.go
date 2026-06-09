package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// freshRepo inits an empty planning tree and returns its root.
func freshRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runRoot(t, "init", "--path", root)
	return root
}

func TestTaskNew_HappyPath(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n# E1\n")

	out := runRoot(t, "-C", root, "task", "new", "My Brand New Task", "--epic", "e1", "--tags", "a,b")
	if !strings.Contains(out, "created") {
		t.Errorf("unexpected output: %q", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "my-brand-new-task.md"))
	if err != nil {
		t.Fatalf("task file not created: %v", err)
	}
	s := string(b)
	for _, want := range []string{
		"status: ready-to-start", "epic: e1", "tier: 3", "priority: medium",
		"effort: Unknown", "## Acceptance criteria", "Epic [[e1]]",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("created task missing %q:\n%s", want, s)
		}
	}
	// Round-trips through show, and the repo is lint-clean (only this task).
	if show := runRoot(t, "-C", root, "task", "show", "my-brand-new-task"); !strings.Contains(show, "e1") {
		t.Errorf("show failed: %q", show)
	}
	runRoot(t, "-C", root, "lint") // would Fatalf if exit != 0
}

func TestTaskNew_Next(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")
	runRoot(t, "-C", root, "task", "new", "Soon", "--epic", "e1", "--next")
	if _, err := os.Stat(filepath.Join(root, "tasks", "next-up", "soon.md")); err != nil {
		t.Errorf("--next should land in next-up/: %v", err)
	}
}

func TestTaskNew_UnknownEpic_Exit11(t *testing.T) {
	root := freshRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "new", "X", "--epic", "nope"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown epic")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11, got %d", ExitCode(err))
	}
}

func TestTaskNew_RefusesClobber(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "epics", "e1.md"), "---\nstatus: in-progress\n---\n")
	runRoot(t, "-C", root, "task", "new", "Dup", "--epic", "e1")
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "new", "Dup", "--epic", "e1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected refusal to clobber an existing task")
	}
	if ExitCode(err) != 14 {
		t.Errorf("clobber should exit 14 (conflict), got %d", ExitCode(err))
	}
}

func TestEpicNew(t *testing.T) {
	root := freshRepo(t)
	out := runRoot(t, "-C", root, "epic", "new", "Payments Revamp", "--description", "Overhaul payments")
	if !strings.Contains(out, "created") {
		t.Errorf("unexpected output: %q", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "epics", "01-payments-revamp.md"))
	if err != nil {
		t.Fatalf("epic not created (auto-number): %v", err)
	}
	s := string(b)
	for _, want := range []string{"status: planning", "description: Overhaul payments", "priority: medium", "**Goal.**"} {
		if !strings.Contains(s, want) {
			t.Errorf("epic missing %q:\n%s", want, s)
		}
	}
}

func TestList_MisfiledMarker(t *testing.T) {
	root := freshRepo(t)
	// A file in ready-to-start/ (active, so it shows) whose frontmatter claims a
	// different recognized status → misfiled, marked with ⚠.
	mustWrite(t, filepath.Join(root, "tasks", "ready-to-start", "drift.md"),
		"---\nstatus: completed\nepic: e\ndescription: d\ntier: 3\npriority: low\neffort: x\ncreated: 2026-06-09\ntags: [a]\n---\n# x\n")
	out := runRoot(t, "-C", root, "task", "list")
	if !strings.Contains(out, "⚠") {
		t.Errorf("expected a ⚠ misfiled marker:\n%q", out)
	}
	if !strings.Contains(out, "misfiled") {
		t.Errorf("expected a misfiled footer note:\n%q", out)
	}
}

func TestLint_FlagsMisfiledArchivedTask(t *testing.T) {
	root := freshRepo(t)
	// A completed/ file with a stale active status — archived, so field lint is
	// skipped, but the drift must still surface.
	mustWrite(t, filepath.Join(root, "tasks", "completed", "old.md"),
		"---\nstatus: in-progress\n---\n# x\n")
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "lint"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected lint to fail on a misfiled archived task")
	}
	if !strings.Contains(out.String(), "old") || !strings.Contains(out.String(), "folder") {
		t.Errorf("expected a misfiled report for 'old':\n%s", out.String())
	}
}

func TestEpicNew_RequiresDescription(t *testing.T) {
	root := freshRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "epic", "new", "X"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when --description is missing")
	}
}
