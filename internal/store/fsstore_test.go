package store

import (
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func writeTask(t *testing.T, root, status, name, content string) {
	t.Helper()
	testutil.Write(t, filepath.Join(root, domain.TasksDir, status, name), content)
}

func TestFS_ListTasks(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "alpha.md",
		"---\nstatus: ready-to-start\nepic: 01-x\ntier: 2\npriority: high\ntags: [a, b]\ndescription: do alpha\n---\n# Alpha\n")
	writeTask(t, root, "in-progress", "beta.md",
		"---\nstatus: in-progress\ndescription: do beta\n---\n# Beta\n")

	tasks, _, err := NewFS(root).ListTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Fatalf("want 2 tasks, got %d", len(tasks))
	}

	seen := map[string]bool{}
	for _, task := range tasks {
		seen[string(task.Status)] = true
		if task.Slug == "alpha" {
			if task.Epic != "01-x" || task.Tier != 2 || task.Priority != "high" || task.Description != "do alpha" {
				t.Errorf("alpha parsed wrong: %+v", task)
			}
			if len(task.Tags) != 2 {
				t.Errorf("alpha tags = %v", task.Tags)
			}
		}
	}
	if !seen["ready-to-start"] || !seen["in-progress"] {
		t.Errorf("missing statuses, saw: %v", seen)
	}
}

func TestFS_ListTasks_StatusFromDirWhenMissing(t *testing.T) {
	root := t.TempDir()
	// No status in frontmatter → directory is the source of truth.
	writeTask(t, root, "completed", "gamma.md", "---\ndescription: g\n---\n# Gamma\n")
	tasks, _, err := NewFS(root).ListTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].Status != "completed" {
		t.Fatalf("got %+v", tasks)
	}
}
