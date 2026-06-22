package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestScanDir_OverDescriptorDir demonstrates the M1 entity-descriptor claim from
// the store side: a kind's listing is the generic scanDir over its descriptor Dir
// plus a per-kind parse func — so lighting up a new entity's store scan is a
// scanDir call, not bespoke directory-walk machinery. It points scanDir at a
// directory taken from the registry (not a hardcoded literal) and confirms the
// resilient-read contract every entity listing inherits: good files list, a
// malformed one becomes a non-fatal FileProblem.
func TestScanDir_OverDescriptorDir(t *testing.T) {
	var dir string
	for _, d := range domain.Descriptors() {
		if d.Kind == "audit" {
			dir = d.Dir
		}
	}
	if dir == "" {
		t.Fatal("no audit descriptor in the registry")
	}

	root := t.TempDir()
	scanPath := filepath.Join(root, dir, "open")
	if err := os.MkdirAll(scanPath, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"good.md", "also.md"} {
		if err := os.WriteFile(filepath.Join(scanPath, name), []byte("# "+name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	parse := func(_ string, content []byte) (string, error) {
		if len(content) == 0 {
			return "", errBadFrontmatter // an empty file is "malformed" for this demo
		}
		return string(content), nil
	}

	got, problems, err := scanDir(scanPath, parse)
	if err != nil {
		t.Fatalf("scanDir: %v", err)
	}
	if len(got) != 2 || len(problems) != 0 {
		t.Fatalf("want 2 items + 0 problems, got %d + %d", len(got), len(problems))
	}

	// A malformed file is a non-fatal FileProblem (skipped, not an error) — the same
	// resilient-read contract a new entity's listing inherits for free.
	if err := os.WriteFile(filepath.Join(scanPath, "bad.md"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got, problems, err = scanDir(scanPath, parse)
	if err != nil {
		t.Fatalf("scanDir (with a bad file): %v", err)
	}
	if len(got) != 2 || len(problems) != 1 {
		t.Fatalf("want 2 items + 1 problem, got %d + %d", len(got), len(problems))
	}
}
