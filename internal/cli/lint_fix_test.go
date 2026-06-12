package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintFix_DryRunThenFix(t *testing.T) {
	root := t.TempDir()
	bad := filepath.Join(root, "tasks", "ready-to-start", "bad.md")
	if err := os.MkdirAll(filepath.Dir(bad), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("---\nstatus: ready-to-start\ndescription: A: B\ntags: x,y\n---\n# Bad\n"), 0o644); err != nil {
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

	// real fix: writes; the file becomes readable.
	if out := runRoot(t, "-C", root, "lint", "--fix"); !strings.Contains(out, "fixed") {
		t.Errorf("expected a fix report: %q", out)
	}
	if listOut := runRoot(t, "-C", root, "task", "list"); !strings.Contains(listOut, "bad") {
		t.Errorf("task should be readable after fix: %q", listOut)
	}
}

// TestLintFix_UnrepairableFileExitsNonZero pins B4: a file the fixer can't
// repair must be surfaced with a non-zero exit — `lint --fix` previously said
// nothing and exited 0, leaving the tree broken while claiming success.
func TestLintFix_UnrepairableFileExitsNonZero(t *testing.T) {
	root := t.TempDir()
	broken := filepath.Join(root, "tasks", "ready-to-start", "broken.md")
	if err := os.MkdirAll(filepath.Dir(broken), 0o755); err != nil {
		t.Fatal(err)
	}
	// Unterminated frontmatter: nothing the text fixer can do with it.
	if err := os.WriteFile(broken, []byte("---\nstatus: ready-to-start\n# no closing fence\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
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
	dry := NewRootCmd(&out, &out)
	dry.SetArgs([]string{"-C", root, "lint", "--fix", "--dry-run"})
	dry.SetOut(&out)
	dry.SetErr(&out)
	if err := dry.Execute(); err != nil {
		t.Errorf("dry-run should not fail on unrepairable files: %v", err)
	}
}
