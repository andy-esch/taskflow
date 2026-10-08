package domain

import (
	"errors"
	"testing"
)

func TestAuditMoveRequiresCompleteEvidence(t *testing.T) {
	for _, to := range []AuditBucket{AuditClosed, AuditDeferred} {
		for _, unparsed := range []int{1, 2} {
			a := Audit{Slug: "portable", Bucket: to, Findings: 1, DoneFindings: 1, UnparsedFindings: unparsed}
			err := a.ValidateMove(to)
			var incomplete *AuditIncompleteEvidenceError
			if !errors.Is(err, ErrValidation) || !errors.As(err, &incomplete) || incomplete.Count != unparsed || incomplete.Slug != "portable" || incomplete.Target != to {
				t.Fatalf("same-bucket %s with %d unparsed: %v", to, unparsed, err)
			}
			if err := a.ValidateMove(AuditOpen); err != nil {
				t.Fatalf("reopen for repair: %v", err)
			}
		}
		for _, a := range []Audit{{Slug: "empty"}, {Slug: "settled", Findings: 2, DoneFindings: 1, DroppedFindings: 1}} {
			if err := a.ValidateMove(to); err != nil {
				t.Fatalf("%s to %s: %v", a.Slug, to, err)
			}
		}
		if err := (Audit{Slug: "open", Findings: 1, OpenFindings: 1}).ValidateMove(to); !errors.Is(err, ErrValidation) {
			t.Fatalf("parsed open finding to %s: %v", to, err)
		}
	}
	if err := (Audit{}).ValidateMove("unknown"); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid bucket: %v", err)
	}
}
