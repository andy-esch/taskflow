package cli

import (
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
