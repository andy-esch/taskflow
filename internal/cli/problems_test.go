package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestPortableProblemsErrorPrefersIdentityOverLocation(t *testing.T) {
	err := portableProblemsError("planning", []core.LintLoadProblem{{
		EntityKind: core.LintEntityTask, EntityID: "6gknown000001", EntitySlug: "known-task",
		Location: "db://misleading/not-the-task", Message: "bad record",
	}})
	if err == nil || !strings.Contains(err.Error(), "known-task") || strings.Contains(err.Error(), "not-the-task") {
		t.Fatalf("portable error should prefer explicit identity: %v", err)
	}
}

func TestPortableProblemsErrorRetainsLocalRepairLocation(t *testing.T) {
	err := portableProblemsError("task", []core.LintLoadProblem{{
		EntityKind: core.LintEntityTask, EntitySlug: "known-task",
		Location: "db://tasks/known-task", Path: "/planning/tasks/6gknown000001-known-task.md",
		Message: "bad record",
	}})
	if err == nil || !strings.Contains(err.Error(), "known-task") ||
		!strings.Contains(err.Error(), "6gknown000001-known-task.md") {
		t.Fatalf("portable local error should retain identity and repair location: %v", err)
	}
}

func TestTaskList_ReportsBadFileButShowsGood(t *testing.T) {
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
	write("tasks/"+testutil.TaskID("good")+"-good.md", "---\nstatus: ready-to-start\ndescription: ok\n---\n# Good\n")
	write("tasks/"+testutil.TaskID("bad")+"-bad.md", "---\nstatus: ready-to-start\ntags: a,b,c\n---\n# Bad\n")

	res, err := runRootStreams(t, "-C", root, "task", "list")

	if err == nil {
		t.Fatal("expected a non-zero result for the unreadable file")
	}
	if ExitCode(err) != 11 {
		t.Errorf("want exit 11, got %d", ExitCode(err))
	}
	// Best-effort listing on stdout, per-file guidance on stderr: the point of the
	// command is that one bad file neither hides the good rows nor pollutes them.
	if !strings.Contains(res.Out, "good") {
		t.Errorf("the good task should still be listed:\n%s", res.Out)
	}
	if !strings.Contains(res.Err, "tags") || !strings.Contains(res.Err, "bad.md") {
		t.Errorf("the bad file should be reported with guidance:\n%s", res.Err)
	}
}
