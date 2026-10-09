package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestAuditMoveErrorPreservesPolicyAndSourceIdentity(t *testing.T) {
	a := domain.Audit{ID: "declared-other", Slug: "display", UnparsedFindings: 2}
	err := &AuditMoveError{
		Source: RecordSource{ID: "source-audit", Location: "urn:opaque:audit"},
		Cause:  a.ValidateMove(domain.AuditClosed),
	}
	var incomplete *domain.AuditIncompleteEvidenceError
	if !errors.Is(err, domain.ErrValidation) || !errors.As(err, &incomplete) || incomplete.Count != 2 {
		t.Fatalf("structured semantic refusal lost: %v", err)
	}
	if !strings.Contains(err.Error(), "audit lint source-audit") || strings.Contains(err.Error(), "audit lint display") || strings.Contains(err.Error(), "declared-other") || strings.Contains(err.Error(), "urn:") {
		t.Fatalf("diagnosis lost source identity or exposed opaque context: %v", err)
	}
}

func TestAuditUnsettledMoveErrorPreservesSourceIdentity(t *testing.T) {
	a := domain.Audit{ID: "declared-other", Slug: "display", Findings: 2, ActiveFindings: 1}
	err := &AuditMoveError{Source: RecordSource{ID: "source-audit", Location: "urn:opaque:audit"}, Cause: a.ValidateMove(domain.AuditDeferred)}
	var unsettled *domain.AuditUnsettledFindingsError
	if !errors.Is(err, domain.ErrValidation) || !errors.As(err, &unsettled) || unsettled.Count != 2 {
		t.Fatalf("structured parsed-status refusal lost: %v", err)
	}
	for _, advice := range []string{"audit lint source-audit", "audit findings source-audit"} {
		if !strings.Contains(err.Error(), advice) {
			t.Fatalf("missing source-backed advice %q: %v", advice, err)
		}
	}
	for _, wrong := range []string{"audit lint display", "audit findings display", "declared-other", "urn:"} {
		if strings.Contains(err.Error(), wrong) {
			t.Fatalf("advice exposed wrong identity/context %q: %v", wrong, err)
		}
	}
}
