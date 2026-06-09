package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

func TestFS_SetFields(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "alpha.md",
		"---\nstatus: ready-to-start\nepic: 01-x\ntier: 2\ncustom: keep\n---\n# Alpha\nbody\n")

	task, err := NewFS(root).SetFields("alpha", map[string]any{
		"priority":    "high",
		"tags":        []string{"a", "b"},
		"description": "new desc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if task.Priority != "high" || task.Description != "new desc" || len(task.Tags) != 2 {
		t.Errorf("task not updated: %+v", task)
	}

	b, err := os.ReadFile(filepath.Join(root, "tasks", "ready-to-start", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "custom: keep") || !strings.Contains(s, "epic: 01-x") {
		t.Errorf("unknown/other fields lost (surgical write failed):\n%s", s)
	}
	if !strings.Contains(s, "# Alpha\nbody\n") {
		t.Errorf("body not preserved:\n%s", s)
	}
	// Output is valid YAML (lists encoded properly).
	fm, _ := splitFrontmatter(b)
	var m map[string]any
	if err := yaml.Unmarshal(fm, &m); err != nil {
		t.Fatalf("output not valid yaml: %v\n%s", err, fm)
	}
	if tags, ok := m["tags"].([]any); !ok || len(tags) != 2 {
		t.Errorf("tags not a 2-element list: %#v", m["tags"])
	}
}

func TestFS_SetFields_NotFound(t *testing.T) {
	_, err := NewFS(t.TempDir()).SetFields("ghost", map[string]any{"priority": "low"})
	if err == nil {
		t.Fatal("want error for missing task")
	}
}
