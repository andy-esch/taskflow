package store

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
)

func writeTask(t *testing.T, root, status, name, content string) {
	t.Helper()
	path, out := testutil.TaskFixture(root, status, name, content)
	testutil.Write(t, path, out)
}

// TestFS_ListTasksWithBodies pins the body-carrying scan (the task twin of
// ListAuditsWithFindings): each task returns with its markdown body — which the
// acceptance-criteria lint reads — alongside the frontmatter ListTasks parses.
func TestFS_ListTasksWithBodies(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "a.md", "---\nstatus: ready-to-start\nepic: e1\ntags: [x]\n---\n# A\n\n## Acceptance criteria\n\n- [x] done\n")

	got, problems, err := NewFS(root).ListTasksWithBodies()
	if err != nil || len(problems) != 0 {
		t.Fatalf("ListTasksWithBodies: %v / %+v", err, problems)
	}
	if len(got) != 1 || got[0].Task.Slug != "a" {
		t.Fatalf("want 1 task 'a', got %+v", got)
	}
	if !strings.Contains(got[0].Body, "## Acceptance criteria") || !strings.Contains(got[0].Body, "- [x] done") {
		t.Errorf("body not carried through the scan:\n%q", got[0].Body)
	}
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

// TestFS_ListTasks_MissingFrontmatterIsLoud: a fence-less file (or a malformed
// opening fence like `---"`) is surfaced as a loud FileProblem naming the valid
// shape — not silently parsed as an empty task, which downstream would misreport
// as merely "missing id".
func TestFS_ListTasks_MissingFrontmatterIsLoud(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "completed", "no-fence.md", "# Just a heading\n\nnotes\n")
	writeTask(t, root, "completed", "bad-fence.md", "---\"\nstatus: completed\nepic: 01-x\n---\n# X\n")

	tasks, problems, err := NewFS(root).ListTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("a file with no valid frontmatter must not parse as a task, got %+v", tasks)
	}
	if len(problems) != 2 {
		t.Fatalf("want 2 loud problems, got %d: %+v", len(problems), problems)
	}
	for _, p := range problems {
		if !strings.Contains(p.Message, "missing frontmatter") || !strings.Contains(p.Message, "schema task") {
			t.Errorf("problem should name the valid shape, got %q", p.Message)
		}
	}
}
