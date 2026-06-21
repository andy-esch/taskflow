package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// New scaffolds stamp the reserved schema-version key, written first.
func TestCreate_StampsSchemaVersion(t *testing.T) {
	fs := NewFS(t.TempDir())
	taskC, err := fs.CreateTask(domain.Task{Slug: "t", Status: domain.StatusReadyToStart, Epic: "e1", Tags: []string{"x"}, Created: "2026-01-01"}, "# T\n", false)
	if err != nil {
		t.Fatal(err)
	}
	epicC, err := fs.CreateEpic("alpha", domain.Epic{Status: "planning", Description: "d", Priority: "medium", Created: "2026-01-01"}, "# E\n", false)
	if err != nil {
		t.Fatal(err)
	}
	auditC, err := fs.CreateAudit(domain.Audit{Slug: "2026-01-01-a", Area: "a", Date: "2026-01-01"}, "# A\n", false)
	if err != nil {
		t.Fatal(err)
	}

	prefix := fmt.Sprintf("---\nschema: %d", domain.FileSchemaVersion)
	for _, p := range []string{taskC.Path, epicC.Path, auditC.Path} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(b), prefix) {
			t.Errorf("%s: schema:%d should be the first frontmatter key:\n%s", p, domain.FileSchemaVersion, b)
		}
	}
}

// A file carrying the reserved schema key parses (the loader ignores it) and both
// a surgical field edit and a relocating move preserve it — it is reserved, not a
// managed field.
func TestSchemaVersion_ParsesAndSurvivesEdits(t *testing.T) {
	root := t.TempDir()
	seed := "---\nschema: 1\nstatus: ready-to-start\ndescription: d\ntier: 2\n---\n# T\n\nbody\n"
	writeTask(t, root, "ready-to-start", "keep.md", seed)
	fs := NewFS(root)

	if _, _, err := fs.GetTask("keep"); err != nil {
		t.Fatalf("a file with schema:1 should load: %v", err)
	}
	if _, err := fs.SetFields("keep", map[string]any{"tier": 3}, false); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(root, domain.TasksDir, "ready-to-start", "keep.md")); !strings.Contains(got, "schema: 1") {
		t.Errorf("SetFields dropped the reserved schema key:\n%s", got)
	}
	if _, err := fs.Move("keep", domain.StatusInProgress, bodyNow, false); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(root, domain.TasksDir, "in-progress", "keep.md")); !strings.Contains(got, "schema: 1") {
		t.Errorf("Move dropped the reserved schema key:\n%s", got)
	}
}

// The reserved key also survives a body edit (EditBody append + replace), which
// rewrites the body through the same surgical frontmatter path.
func TestSchemaVersion_SurvivesBodyEdit(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "keep.md", "---\nschema: 1\nstatus: ready-to-start\ndescription: d\n---\n# T\n\nbody\n")
	fs := NewFS(root)
	path := filepath.Join(root, domain.TasksDir, "ready-to-start", "keep.md")

	if _, _, err := fs.EditBody("keep", "## Notes\n- x", true, bodyNow, false); err != nil { // append
		t.Fatal(err)
	}
	if got := readFile(t, path); !strings.Contains(got, "schema: 1") {
		t.Errorf("append dropped the reserved schema key:\n%s", got)
	}
	if _, _, err := fs.EditBody("keep", "# Rewritten", false, bodyNow, false); err != nil { // replace
		t.Fatal(err)
	}
	if got := readFile(t, path); !strings.Contains(got, "schema: 1") {
		t.Errorf("replace dropped the reserved schema key:\n%s", got)
	}
}
