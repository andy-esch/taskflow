package core

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type lintSourceFake struct {
	taskRecords      []TaskWithBody
	taskProblems     []LintLoadProblem
	epicProblems     []LintLoadProblem
	auditProblems    []LintLoadProblem
	researchProblems []LintLoadProblem
	taskReads        int
	epicReads        int
	auditReads       int
	researchReads    int
}

func (f *lintSourceFake) ReadLintTasks() ([]TaskWithBody, []LintLoadProblem, error) {
	f.taskReads++
	return f.taskRecords, f.taskProblems, nil
}

func (f *lintSourceFake) ReadLintEpics() ([]domain.Epic, []LintLoadProblem, error) {
	f.epicReads++
	return nil, f.epicProblems, nil
}

func (f *lintSourceFake) ReadAuditSnapshot(string) (AuditSnapshot, error) {
	f.auditReads++
	return AuditSnapshot{Problems: f.auditProblems}, nil
}

func (f *lintSourceFake) ReadLintResearch() ([]domain.Research, []LintLoadProblem, error) {
	f.researchReads++
	return nil, f.researchProblems, nil
}

func TestLintPreservesPortableLoadProblemIdentityWithoutLocations(t *testing.T) {
	source := &lintSourceFake{
		taskProblems: []LintLoadProblem{{
			EntityKind: LintEntityTask, EntityID: "6g0000000001", EntitySlug: "broken-task", Message: "bad task",
		}},
		epicProblems: []LintLoadProblem{{
			EntityKind: LintEntityEpic, EntityID: "21-broken-epic", Message: "bad epic",
		}},
		auditProblems: []LintLoadProblem{{
			EntityKind: LintEntityAudit, EntityID: "6g0000000002", EntitySlug: "2026-09-23-broken-audit", Message: "bad audit",
		}},
		researchProblems: []LintLoadProblem{{
			EntityKind: LintEntityResearch, EntityID: "6g0000000003", EntitySlug: "broken-research", Message: "bad research",
		}},
	}
	threads := &threadReadFake{problems: []ThreadReadProblem{{
		ThreadID: "6g0000000004", ThreadSlug: "broken-thread", Message: "bad Thread",
	}}}

	_, problems, err := NewService(nil, WithLintSource(source), WithThreadStore(threads)).Lint()
	if err != nil {
		t.Fatal(err)
	}
	if source.taskReads != 1 || source.epicReads != 1 || source.auditReads != 1 || source.researchReads != 1 {
		t.Fatalf("lint source reads = task:%d epic:%d audit:%d research:%d; want one each",
			source.taskReads, source.epicReads, source.auditReads, source.researchReads)
	}
	if len(problems) != 5 {
		t.Fatalf("problems = %+v; want all five entity kinds", problems)
	}
	byKind := make(map[LintEntityKind]LintLoadProblem, len(problems))
	for _, problem := range problems {
		byKind[problem.EntityKind] = problem
		if problem.Location != "" || problem.LocationIsPath {
			t.Errorf("pathless problem acquired filesystem semantics: %+v", problem)
		}
	}
	for kind, wantID := range map[LintEntityKind]string{
		LintEntityTask: "6g0000000001", LintEntityEpic: "21-broken-epic",
		LintEntityAudit: "6g0000000002", LintEntityResearch: "6g0000000003",
		LintEntityThread: "6g0000000004",
	} {
		if got := byKind[kind]; got.EntityID != wantID || got.Message == "" {
			t.Errorf("%s problem = %+v; want id %q and a message", kind, got, wantID)
		}
	}
}

func TestLintRequiresDedicatedReadCapability(t *testing.T) {
	_, _, err := NewService(nopStore{}).Lint()
	if err == nil || !strings.Contains(err.Error(), "lint reads are unavailable") {
		t.Fatalf("Lint error = %v; want missing lint capability", err)
	}
}

func TestLintRejectsMissingAuditSnapshotCapability(t *testing.T) {
	source := &lintSourceFake{}
	dropAuditCapability := func(s *Service) { s.auditReads = nil }

	_, _, err := NewService(nil, WithLintSource(source), dropAuditCapability).Lint()
	if err == nil || !strings.Contains(err.Error(), "audit snapshot reads are unavailable") {
		t.Fatalf("Lint error = %v; want missing audit snapshot capability", err)
	}
	if source.taskReads != 0 || source.epicReads != 0 || source.researchReads != 0 || source.auditReads != 0 {
		t.Fatalf("lint read before capability validation: %+v", source)
	}
}

func TestExplicitAuditSnapshotSourceWinsRegardlessOfOptionOrder(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts func(*lintSourceFake, *auditSnapshotStub) []Option
	}{
		{"audit then lint", func(lint *lintSourceFake, audit *auditSnapshotStub) []Option {
			return []Option{WithAuditSnapshotSource(audit), WithLintSource(lint)}
		}},
		{"lint then audit", func(lint *lintSourceFake, audit *auditSnapshotStub) []Option {
			return []Option{WithLintSource(lint), WithAuditSnapshotSource(audit)}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			broad := &lintSourceFake{}
			dedicated := &auditSnapshotStub{all: AuditSnapshot{Problems: []LintLoadProblem{{
				EntityKind: LintEntityAudit, EntityID: "6g0000000009", Message: "dedicated source",
			}}}}
			svc := NewService(nil, tc.opts(broad, dedicated)...)

			_, problems, err := svc.QueryFindings(FindingFilter{})
			if err != nil || len(problems) != 1 || problems[0].Message != "dedicated source" {
				t.Fatalf("QueryFindings problems = %+v, err = %v", problems, err)
			}
			if len(dedicated.calls) != 1 || dedicated.calls[0] != "" || broad.auditReads != 0 {
				t.Fatalf("audit reads = dedicated:%q broad:%d", dedicated.calls, broad.auditReads)
			}
		})
	}
}

func TestLintSourceSuppliesDefaultAuditSnapshotSource(t *testing.T) {
	source := &lintSourceFake{}
	if _, _, err := NewService(nil, WithLintSource(source)).QueryFindings(FindingFilter{}); err != nil {
		t.Fatal(err)
	}
	if source.auditReads != 1 {
		t.Fatalf("audit reads = %d, want one embedded default read", source.auditReads)
	}
}

func TestLintAttributesPathlessGraphDiagnosticsToReadableRecords(t *testing.T) {
	t.Run("dependency and lifecycle", func(t *testing.T) {
		prerequisite := graphRecord("pathless-prerequisite", domain.StatusNextUp)
		dependent := graphRecord("pathless-in-flight", domain.StatusInProgress, prerequisite.ID)
		invalid := graphRecord("pathless-invalid", domain.StatusReadyToStart, "not-a-stable-id")
		prerequisite.Path, dependent.Path, invalid.Path = "", "", ""

		results := lintTaskRecords(t, prerequisite, dependent, invalid)
		assertLintIssue(t, results, invalid.Slug, "depends_on", "not a stable task id")
		assertLintIssue(t, results, dependent.Slug, "status", "dependency gate")
	})

	t.Run("cycle", func(t *testing.T) {
		left := graphRecord("pathless-cycle-left", domain.StatusReadyToStart)
		right := graphRecord("pathless-cycle-right", domain.StatusReadyToStart, left.ID)
		left.DependsOn = []string{right.ID}
		left.Path, right.Path = "", ""

		results := lintTaskRecords(t, left, right)
		assertLintIssue(t, results, left.Slug, "depends_on", "dependency cycle")
		assertLintIssue(t, results, right.Slug, "depends_on", "dependency cycle")
	})

	t.Run("legacy declaration", func(t *testing.T) {
		prerequisite := graphRecord("pathless-legacy-prerequisite", domain.StatusCompleted)
		owner := graphRecord("pathless-legacy-owner", domain.StatusReadyToStart)
		owner.LegacyBlockedBy = []string{prerequisite.Slug}
		owner.LegacyDependencyFields = []string{"blocked_by"}
		prerequisite.Path, owner.Path = "", ""

		results := lintTaskRecords(t, prerequisite, owner)
		assertLintIssue(t, results, owner.Slug, "blocked_by", "legacy dependency field")
	})
}

func TestLintRecordAttributionDoesNotCollideOnIDOrLocation(t *testing.T) {
	first := graphRecord("portable-duplicate-first", domain.StatusReadyToStart)
	second := graphRecord("portable-duplicate-second", domain.StatusReadyToStart, "bad-reference")
	second.ID, second.FilenameID = first.ID, first.FilenameID
	// An opaque or contradictory location is context, not the record join key.
	first.Path, second.Path = "opaque://same", "opaque://same"

	results := lintTaskRecords(t, first, second)
	assertLintIssue(t, results, first.Slug, "id", "duplicate stable task id")
	assertLintIssue(t, results, second.Slug, "id", "duplicate stable task id")
	assertLintIssue(t, results, second.Slug, "depends_on", "bad-reference")
	if lintResultHas(results, first.Slug, "depends_on", "bad-reference") {
		t.Fatalf("second record dependency defect leaked onto first record: %+v", results)
	}
}

func TestLintUsesPathlessUnreadableIdentityInLifecycleDiagnosis(t *testing.T) {
	unreadableID := "6g0000000005"
	dependent := graphRecord("depends-on-pathless-unreadable", domain.StatusInProgress, unreadableID)
	dependent.Path = ""
	source := &lintSourceFake{
		taskRecords: []TaskWithBody{{Task: dependent}},
		taskProblems: []LintLoadProblem{{
			EntityKind: LintEntityTask, EntityID: unreadableID,
			EntitySlug: "unreadable", Message: "remote decode failed",
		}},
	}

	results, problems, err := NewService(nil, WithLintSource(source)).Lint()
	if err != nil {
		t.Fatal(err)
	}
	assertLintIssue(t, results, dependent.Slug, "status", unreadableID)
	if len(problems) != 1 || problems[0].EntityID != unreadableID || problems[0].Location != "" {
		t.Fatalf("load problems = %+v", problems)
	}
}

func lintTaskRecords(t *testing.T, tasks ...domain.Task) []LintResult {
	t.Helper()
	records := make([]TaskWithBody, len(tasks))
	for index, task := range tasks {
		records[index] = TaskWithBody{Task: task}
	}
	results, _, err := NewService(nil, WithLintSource(&lintSourceFake{taskRecords: records})).Lint()
	if err != nil {
		t.Fatal(err)
	}
	return results
}

func assertLintIssue(t *testing.T, results []LintResult, slug, field, messagePart string) {
	t.Helper()
	if !lintResultHas(results, slug, field, messagePart) {
		t.Fatalf("missing %s issue containing %q for %s in %+v", field, messagePart, slug, results)
	}
}

func lintResultHas(results []LintResult, slug, field, messagePart string) bool {
	for _, result := range results {
		if result.Slug != slug {
			continue
		}
		for _, issue := range result.Issues {
			if issue.Field == field && strings.Contains(issue.Message, messagePart) {
				return true
			}
		}
	}
	return false
}
