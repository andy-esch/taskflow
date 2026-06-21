package core

import (
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

const gatewayBody = `# Audit: gateway

#### S1. Tighten the API gateway
**Status:** open · **Component:** stravapipe · **Effort:** S · **Urgency:** soon

#### H1. Fix the auth bypass
**Status:** fixed 2026-06-15 (PR #724) · **Component:** auth · **Effort:** M · **Urgency:** acute
`

const ingestBody = `# Audit: ingest

#### M1. Backfill retries
**Status:** open · **Component:** stravapipe · **Effort:** M · **Urgency:** eventually
`

func findingsRepo() *fakeStore {
	return &fakeStore{
		audits: []domain.Audit{
			{Slug: "2026-06-14-gateway", Bucket: domain.AuditOpen},
			{Slug: "2026-06-10-ingest", Bucket: domain.AuditClosed},
		},
		auditBodies: map[string]string{
			"2026-06-14-gateway": gatewayBody,
			"2026-06-10-ingest":  ingestBody,
		},
	}
}

func codes(fs []AuditFinding) []string {
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = f.Audit + ":" + f.Code
	}
	return out
}

func TestQueryFindings_CrossAudit_NoFilter(t *testing.T) {
	got, problems, err := NewService(findingsRepo()).QueryFindings(FindingFilter{})
	if err != nil || len(problems) != 0 {
		t.Fatalf("QueryFindings: %v / %v", err, problems)
	}
	if len(got) != 3 {
		t.Fatalf("expected all 3 findings across both audits, got %v", codes(got))
	}
	// Each hit carries its audit + bucket.
	if got[0].Audit != "2026-06-14-gateway" || got[0].Bucket != "open" {
		t.Errorf("finding should be tagged with its audit/bucket, got %+v", got[0])
	}
}

func TestQueryFindings_StatusFilter(t *testing.T) {
	got, _, _ := NewService(findingsRepo()).QueryFindings(FindingFilter{Status: []string{"open"}})
	// S1 (gateway) + M1 (ingest) are open; H1 is fixed.
	if len(got) != 2 {
		t.Fatalf("status=open should match 2, got %v", codes(got))
	}
	got, _, _ = NewService(findingsRepo()).QueryFindings(FindingFilter{Status: []string{"FIXED"}})
	if len(got) != 1 || got[0].Code != "H1" {
		t.Errorf("status=FIXED (case-insensitive) should match just H1, got %v", codes(got))
	}
}

func TestQueryFindings_MultiValueAndComponent(t *testing.T) {
	// effort any-of S,M → all three (S1=S, H1=M, M1=M).
	if got, _, _ := NewService(findingsRepo()).QueryFindings(FindingFilter{Effort: []string{"S", "M"}}); len(got) != 3 {
		t.Errorf("effort S,M should match all 3, got %v", codes(got))
	}
	// component is a case-insensitive substring: "strava" → the two stravapipe findings.
	got, _, _ := NewService(findingsRepo()).QueryFindings(FindingFilter{Component: "STRAVA"})
	if len(got) != 2 {
		t.Errorf("component substring 'STRAVA' should match the 2 stravapipe findings, got %v", codes(got))
	}
}

func TestQueryFindings_SingleAudit(t *testing.T) {
	got, _, err := NewService(findingsRepo()).QueryFindings(FindingFilter{Audit: "2026-06-14-gateway"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("single-audit query should only see the gateway findings, got %v", codes(got))
	}
	// An unknown audit slug is ErrNotFound, not a silent empty result.
	if _, _, err := NewService(findingsRepo()).QueryFindings(FindingFilter{Audit: "nope"}); err == nil {
		t.Error("unknown audit slug should error, not return empty")
	}
}

func TestLintAudits(t *testing.T) {
	// gateway (open): S1 open + H1 fixed → clean. ingest (closed): M1 open → bucket drift.
	results, problems, err := NewService(findingsRepo()).LintAudits("")
	if err != nil || len(problems) != 0 {
		t.Fatalf("LintAudits: %v / %v", err, problems)
	}
	if len(results) != 1 || results[0].Slug != "2026-06-10-ingest" {
		t.Fatalf("expected only the closed-audit bucket drift, got %+v", results)
	}
	if results[0].Issues[0].Field != "bucket" {
		t.Errorf("expected a bucket issue, got %v", results[0].Issues)
	}
	// Single, clean audit → no issues.
	if clean, _, _ := NewService(findingsRepo()).LintAudits("2026-06-14-gateway"); len(clean) != 0 {
		t.Errorf("the open gateway audit should lint clean, got %+v", clean)
	}
}
