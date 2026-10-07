package core

import (
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestPortableAuditReadsRetainIncompleteFindingEvidence(t *testing.T) {
	audit := domain.Audit{ID: "opaque-audit", Slug: "portable", Bucket: domain.AuditOpen,
		Findings: 1, DoneFindings: 1, UnparsedFindings: 2}
	fake := &fakeStore{audits: []domain.Audit{audit}}
	svc := MustNewService(fake)
	listed, problems, err := svc.ListAudits("", false)
	if err != nil || len(problems) != 0 || len(listed) != 1 {
		t.Fatalf("list: %+v problems=%+v err=%v", listed, problems, err)
	}
	shown, err := svc.ShowAudit("portable")
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []domain.Audit{listed[0].Value, shown.Value.Audit} {
		if value.UnparsedFindings != 2 || value.Findings != 1 || value.Percent() != 100 || value.ReadyToClose() {
			t.Fatalf("portable read invented counts or readiness: %+v", value)
		}
	}
}
