package store

import (
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestFSLintSourceTranslatesLocalProblemsAtAdapterBoundary(t *testing.T) {
	root := t.TempDir()
	taskID := testutil.TaskID("lint-task")
	auditID := testutil.TaskID("lint-audit")
	researchID := testutil.TaskID("lint-research")
	paths := map[core.LintEntityKind]string{
		core.LintEntityTask:     filepath.Join(root, domain.TasksDir, taskID+"-broken-task.md"),
		core.LintEntityEpic:     filepath.Join(root, domain.EpicsDir, "21-broken-epic.md"),
		core.LintEntityAudit:    filepath.Join(root, domain.AuditsDir, auditID+"-2026-09-23-broken-audit.md"),
		core.LintEntityResearch: filepath.Join(root, domain.ResearchDir, researchID+"-broken-research.md"),
	}
	for _, path := range paths {
		testutil.Write(t, path, "---\nid: [unterminated\n---\n")
	}

	fs := NewFS(root)
	_, taskProblems, taskErr := fs.ReadLintTasks()
	_, epicProblems, epicErr := fs.ReadLintEpics()
	_, auditProblems, auditErr := fs.ReadLintAudits()
	_, researchProblems, researchErr := fs.ReadLintResearch()
	for kind, err := range map[core.LintEntityKind]error{
		core.LintEntityTask: taskErr, core.LintEntityEpic: epicErr,
		core.LintEntityAudit: auditErr, core.LintEntityResearch: researchErr,
	} {
		if err != nil {
			t.Fatalf("ReadLint%s: %v", kind, err)
		}
	}

	wants := map[core.LintEntityKind]struct {
		problems []core.LintLoadProblem
		id       string
		slug     string
	}{
		core.LintEntityTask:     {taskProblems, taskID, "broken-task"},
		core.LintEntityEpic:     {epicProblems, "21-broken-epic", ""},
		core.LintEntityAudit:    {auditProblems, auditID, "2026-09-23-broken-audit"},
		core.LintEntityResearch: {researchProblems, researchID, "broken-research"},
	}
	for kind, want := range wants {
		if len(want.problems) != 1 {
			t.Fatalf("%s problems = %+v; want one", kind, want.problems)
		}
		got := want.problems[0]
		if got.EntityKind != kind || got.EntityID != want.id || got.EntitySlug != want.slug ||
			got.Location != paths[kind] || !got.LocationIsPath || got.Message == "" {
			t.Errorf("%s problem = %+v; want id=%q slug=%q path=%q", kind, got, want.id, want.slug, paths[kind])
		}
	}
}
