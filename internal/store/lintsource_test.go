package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	auditSnapshot, auditErr := fs.ReadAuditSnapshot("")
	auditProblems := auditSnapshot.Problems
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

func TestFSAuditSnapshotPreservesSingleAuditResolutionSemantics(t *testing.T) {
	root := t.TempDir()
	writeAudit(t, root, "open", "2026-09-20-shared-alpha.md", "---\narea: cli\ndate: 2026-09-20\n---\n\n#### H1. Alpha\n**Status:** open\n")
	writeAudit(t, root, "closed", "2026-09-21-shared-beta.md", "---\narea: cli\ndate: 2026-09-21\n---\n\n#### M1. Beta\n**Status:** fixed\n")
	writeAudit(t, root, "open", "2026-09-22-unrelated-corrupt.md", "---\nid: [unterminated\n---\n")
	fs := NewFS(root)
	reads := map[string]int{}
	fs.auditReadFile = func(path string) ([]byte, error) {
		reads[path]++
		return os.ReadFile(path)
	}

	snapshot, err := fs.ReadAuditSnapshot("shared-alpha")
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Problems) != 0 || len(snapshot.Audits) != 1 || snapshot.Audits[0].Audit.Slug != "2026-09-20-shared-alpha" ||
		len(snapshot.Audits[0].Findings) != 1 || snapshot.Audits[0].Findings[0].Code != "H1" {
		t.Fatalf("single audit snapshot = %+v", snapshot)
	}
	if len(reads) != 1 {
		t.Fatalf("selected snapshot opened %d sources: %+v", len(reads), reads)
	}
	for path, count := range reads {
		if count != 1 || !strings.Contains(filepath.Base(path), "shared-alpha") {
			t.Fatalf("selected snapshot reads = %+v; want shared-alpha exactly once", reads)
		}
	}

	clear(reads)
	all, err := fs.ReadAuditSnapshot("")
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Audits) != 2 || len(all.Problems) != 1 || len(reads) != 3 {
		t.Fatalf("unfiltered snapshot = %+v, reads = %+v", all, reads)
	}
	for path, count := range reads {
		if count != 1 {
			t.Fatalf("unfiltered snapshot read %s %d times; want once", path, count)
		}
	}
	if _, err := fs.ReadAuditSnapshot("shared"); !errors.Is(err, domain.ErrAmbiguous) {
		t.Fatalf("ambiguous selector error = %v, want ErrAmbiguous", err)
	}
	if _, err := fs.ReadAuditSnapshot("missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing selector error = %v, want ErrNotFound", err)
	}
}
