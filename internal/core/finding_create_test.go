package core

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

type findingCreationStore struct {
	nopStore
	body             string
	bucket           domain.AuditBucket
	conflictBody     string
	transformCalls   int
	lastTransformDry bool
}

func (s *findingCreationStore) TransformAuditBody(
	slug string,
	_ time.Time,
	dryRun bool,
	transform func(domain.Audit, string) (string, error),
) (domain.Audit, string, bool, error) {
	s.transformCalls++
	s.lastTransformDry = dryRun
	current := s.body
	bucket := s.bucket
	if bucket == "" {
		bucket = domain.AuditOpen
	}
	next, err := transform(domain.Audit{Slug: slug, Bucket: bucket}, current)
	if err != nil {
		return domain.Audit{}, "", false, err
	}
	if s.conflictBody != "" {
		s.body = s.conflictBody
		s.conflictBody = ""
		return domain.Audit{}, "", false, domain.ErrConflict
	}
	findings := domain.ParseFindings(next)
	tally := domain.TallyFindings(findings)
	audit := domain.Audit{
		Slug: slug, Bucket: bucket, Findings: len(findings), OpenFindings: tally.Open,
		ActiveFindings: tally.Active, DoneFindings: tally.Done, DroppedFindings: tally.Dropped,
	}
	if !dryRun {
		s.body = next
	}
	return audit, next, next != current, nil
}

func TestNewFindingRefusesNonOpenAuditFromGuardedSnapshot(t *testing.T) {
	body := managedFindingCreationBody()
	store := &findingCreationStore{body: body, bucket: domain.AuditClosed}
	svc := NewService(store)
	_, err := svc.NewFinding("audit", NewFindingParams{Band: "H", Title: "Would reopen work"})
	if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "reopen the audit first") {
		t.Fatalf("closed-audit creation should be refused, got %v", err)
	}
	if store.body != body {
		t.Fatal("closed-audit refusal changed the body")
	}
}

func managedFindingCreationBody() string {
	return "# Audit\n\n## Findings\n\n#### H1. Existing · **Status:** open\n\nEvidence.\n\n" +
		"## Candidate tasks\n\n" + domain.CandidateTasksMarkerComment() + "\n"
}

func TestNewFindingCreatesFindingAndCandidateInOneTransform(t *testing.T) {
	store := &findingCreationStore{body: managedFindingCreationBody()}
	svc := NewService(store)
	candidate := "Create a focused follow-up"
	receipt, err := svc.NewFinding("audit", NewFindingParams{
		Band: "H", Title: "New issue", Effort: "S", Urgency: "soon",
		Body: "The new evidence.", Recommendation: "Fix it.", Candidate: &candidate,
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.transformCalls != 1 || receipt.Finding.Code != "H2" || receipt.DryRun {
		t.Fatalf("receipt=%+v calls=%d", receipt, store.transformCalls)
	}
	for _, want := range []string{
		"#### H2. New issue · **Status:** open",
		"- ○ H2 · open — Create a focused follow-up",
	} {
		if !strings.Contains(store.body, want) {
			t.Errorf("atomic body missing %q:\n%s", want, store.body)
		}
	}
}

func TestNewFindingRetriesAllocationAgainstFreshBody(t *testing.T) {
	initial := managedFindingCreationBody()
	concurrent, _, err := domain.CreateFinding(initial, domain.FindingDraft{Band: "H", Title: "Concurrent issue"})
	if err != nil {
		t.Fatal(err)
	}
	store := &findingCreationStore{body: initial, conflictBody: concurrent}
	svc := NewService(store, WithRetry(2, func(int) {}))
	receipt, err := svc.NewFinding("audit", NewFindingParams{Band: "H", Title: "Retried issue"})
	if err != nil {
		t.Fatal(err)
	}
	if store.transformCalls != 2 || receipt.Finding.Code != "H3" {
		t.Fatalf("fresh allocation receipt=%+v calls=%d", receipt, store.transformCalls)
	}
	if !strings.Contains(store.body, "#### H2. Concurrent issue") || !strings.Contains(store.body, "#### H3. Retried issue") {
		t.Fatalf("retry lost one creation:\n%s", store.body)
	}
}

func TestNewFindingCandidateRefusalIsAtomicAndDryRunDoesNotPersist(t *testing.T) {
	legacy := "## Findings\n\n#### M1. Existing · **Status:** open\n\n## Candidate tasks\n\n- legacy prose\n"
	store := &findingCreationStore{body: legacy}
	svc := NewService(store)
	candidate := "Needs a managed section"
	if _, err := svc.NewFinding("audit", NewFindingParams{Band: "M", Title: "No partial write", Candidate: &candidate}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("legacy candidate section should refuse creation, got %v", err)
	}
	if store.body != legacy {
		t.Fatalf("failed candidate creation partially changed body:\n%s", store.body)
	}

	receipt, err := svc.NewFinding("audit", NewFindingParams{Band: "M", Title: "Preview", DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.DryRun || receipt.Finding.Code != "M2" || !store.lastTransformDry {
		t.Fatalf("dry-run receipt=%+v storeDry=%v", receipt, store.lastTransformDry)
	}
	if store.body != legacy {
		t.Fatalf("dry run persisted body:\n%s", store.body)
	}

	empty := "  "
	if _, err := svc.NewFinding("audit", NewFindingParams{Band: "M", Title: "x", Candidate: &empty}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("empty creation candidate should be rejected, got %v", err)
	}
}
