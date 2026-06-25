package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLint_DuplicateSlug_Exit11 pins M11 end-to-end: the same slug in two status
// dirs (the Ctrl-C-in-Move leftover) makes `lint` exit 11 and name both dirs, so
// the otherwise-silent permanent ErrAmbiguous is discoverable and hand-repairable.
func TestLint_DuplicateSlug_Exit11(t *testing.T) {
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
	// Both copies folder-matching (the Move-crash shape): only the duplicate is wrong.
	// The seeded epic must itself be lint-clean (priority + description), or it would
	// fail lint too and the duplicate wouldn't be the SOLE failure exit 11 keys off.
	write("tasks/in-progress/dup.md", "---\nstatus: in-progress\nepic: e1\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\ndescription: d\n---\n# Dup\n")
	write("tasks/completed/dup.md", "---\nstatus: completed\n---\n# Dup\n")
	write("epics/e1.md", "---\nstatus: active\npriority: high\ndescription: the epic\n---\n# E1\n")

	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "lint"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("a duplicate slug should fail lint")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11 for a duplicate slug, got %d", ExitCode(err))
	}
	if o := out.String(); !strings.Contains(o, "duplicate") || !strings.Contains(o, "dup") {
		t.Errorf("lint should name the duplicate slug:\n%s", o)
	}
}

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
	write("epics/e1.md", "---\nstatus: active\n---\n# E1\n")
	write("tasks/ready-to-start/good.md",
		"---\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-01\ntags: [a]\n---\n# Good\n")

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
	write("epics/e1.md", "---\nstatus: active\npriority: high\ndescription: the epic\n---\n# E1\n")
	write("tasks/ready-to-start/good.md",
		"---\nstatus: ready-to-start\nepic: e1\ntier: 2\npriority: high\neffort: 2h\ncreated: 2026-01-01\ntags: [a]\n---\n# Good\n")

	out := runRoot(t, "-C", root, "lint")
	if !strings.Contains(out, "pass lint") {
		t.Errorf("expected pass, got: %q", out)
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
