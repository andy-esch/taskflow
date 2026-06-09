package store

import (
	"strings"
	"testing"
)

func TestFrontmatterError_GitConflictMarkers(t *testing.T) {
	fm := []byte("status: open\n<<<<<<< HEAD\npriority: high\n=======\npriority: low\n>>>>>>> branch\n")
	msg := frontmatterError(fm, errBadFrontmatter)
	if !strings.Contains(msg, "conflict") {
		t.Errorf("a conflicted file should report a merge conflict, got: %q", msg)
	}
}

func TestDiagnoseFrontmatter_TagsAsString(t *testing.T) {
	// The exact pm-written shape: bare comma string instead of a list.
	msg := diagnoseFrontmatter([]byte("status: ready-to-start\ntags: dispatcher,refactoring,tech-debt\n"))
	if !strings.Contains(msg, `"tags"`) || !strings.Contains(msg, "must be a YAML list") {
		t.Errorf("unhelpful message: %q", msg)
	}
	if !strings.Contains(msg, "[dispatcher, refactoring, tech-debt]") {
		t.Errorf("should suggest the concrete list fix: %q", msg)
	}
}

func TestDiagnoseFrontmatter_UnquotedColon(t *testing.T) {
	// A description with an embedded ": " makes the whole block unparseable.
	msg := diagnoseFrontmatter([]byte("status: ready-to-start\nepic: e\ndescription: Phase 1: do the thing\n"))
	if !strings.Contains(msg, `"description"`) || !strings.Contains(msg, "wrap the value in quotes") {
		t.Errorf("unhelpful message: %q", msg)
	}
	if !strings.Contains(msg, "line 3") {
		t.Errorf("should pinpoint the line: %q", msg)
	}
}

func TestFS_ListTasks_SkipsBadFileWithProblem(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "good.md", "---\nstatus: ready-to-start\ntags: [a]\n---\n# G\n")
	writeTask(t, root, "ready-to-start", "bad.md", "---\nstatus: ready-to-start\ntags: a,b,c\n---\n# B\n")

	tasks, problems, err := NewFS(root).ListTasks()
	if err != nil {
		t.Fatal(err) // a single bad file must NOT be fatal
	}
	if len(tasks) != 1 || tasks[0].Slug != "good" {
		t.Errorf("want the good task only, got %+v", tasks)
	}
	if len(problems) != 1 {
		t.Fatalf("want 1 problem, got %d", len(problems))
	}
	if !strings.Contains(problems[0].Path, "bad.md") || !strings.Contains(problems[0].Message, "tags") {
		t.Errorf("unactionable problem: %+v", problems[0])
	}
}
