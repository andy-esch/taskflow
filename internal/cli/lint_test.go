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
	write("tasks/in-progress/dup.md", "---\nstatus: active\nepic: e1\ntier: 2\npriority: high\neffort: 1h\ncreated: 2026-01-01\ntags: [a]\ndescription: d\n---\n# Dup\n")
	write("tasks/completed/dup.md", "---\nstatus: completed\n---\n# Dup\n")
	write("epics/e1.md", "---\nstatus: active\n---\n# E1\n")

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
