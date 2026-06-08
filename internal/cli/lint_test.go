package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	write("epics/e1.md", "---\nstatus: in-progress\n---\n# E1\n")
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
	cmd := NewRootCmd(&out, &out)
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
