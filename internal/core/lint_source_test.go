package core

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type lintSourceFake struct {
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
	return nil, f.taskProblems, nil
}

func (f *lintSourceFake) ReadLintEpics() ([]domain.Epic, []LintLoadProblem, error) {
	f.epicReads++
	return nil, f.epicProblems, nil
}

func (f *lintSourceFake) ReadLintAudits() ([]AuditWithFindings, []LintLoadProblem, error) {
	f.auditReads++
	return nil, f.auditProblems, nil
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
