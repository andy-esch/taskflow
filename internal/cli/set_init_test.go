package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskSet(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "set", "alpha", "--priority", "low", "--tags", "x,y")
	if !strings.Contains(out, "updated alpha") {
		t.Errorf("unexpected output: %q", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "priority: low") {
		t.Errorf("priority not set:\n%s", b)
	}
}

func TestTaskSet_ArbitraryKeyValue(t *testing.T) {
	root := setupRepo(t)
	// Unknown keys are rejected without --force (decided 2026-06-12): a typo'd
	// field name must not silently persist.
	{
		var out bytes.Buffer
		cmd := NewRootCmd(&out, &out)
		cmd.SetArgs([]string{"-C", root, "task", "set", "alpha", "--set", "owner=me"})
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		err := cmd.Execute()
		if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "--force") {
			t.Fatalf("unknown key without --force should exit 11 mentioning --force, got %v", err)
		}
	}
	out := runRoot(t, "-C", root, "task", "set", "alpha", "--force",
		"--set", "owner=me", "--set", "custom_field=keep me")
	if !strings.Contains(out, "updated alpha") {
		t.Errorf("unexpected output: %q", out)
	}
	b, err := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "owner: me") {
		t.Errorf("arbitrary field not written:\n%s", s)
	}
	// A value with a space (the colon-free case) must round-trip readably.
	if !strings.Contains(s, "custom_field:") || !strings.Contains(s, "keep me") {
		t.Errorf("arbitrary field with space not written:\n%s", s)
	}
	// The body and existing fields must survive the surgical write.
	if !strings.Contains(s, "# Alpha") || !strings.Contains(s, "status: ready-to-start") {
		t.Errorf("body or existing field lost:\n%s", s)
	}
}

func TestTaskSet_MalformedSet_Errors(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "set", "alpha", "--set", "noequals"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error for --set without '='")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit code 11 (validation), got %d", ExitCode(err))
	}
}

func TestTaskSet_InvalidPriority_Exit11(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "set", "alpha", "--priority", "urgent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit code 11 (validation), got %d", ExitCode(err))
	}
}

func TestTaskSet_NoFields_Errors(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "set", "alpha"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected an error when no fields are given")
	}
}

func TestInit_ThenList(t *testing.T) {
	root := t.TempDir()
	out := runRoot(t, "init", "--path", root)
	if !strings.Contains(out, "initialized") {
		t.Errorf("unexpected init output: %q", out)
	}
	// list should now work (and be empty) without a "not a planning repo" error.
	if listOut := runRoot(t, "-C", root, "task", "list"); listOut != "" {
		t.Errorf("expected empty list, got %q", listOut)
	}
}
