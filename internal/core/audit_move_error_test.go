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
