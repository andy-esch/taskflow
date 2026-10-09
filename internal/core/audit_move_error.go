package core

import (
	"errors"
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// AuditMoveError carries the adapter-established source of a guarded refusal.
// Neither the user's possibly fuzzy selector nor the audit's declaration/display
// slug is authoritative recovery identity. Cause retains the domain policy error.
type AuditMoveError struct {
	Source RecordSource
	Cause  error
}

func (e *AuditMoveError) Error() string {
	var incomplete *domain.AuditIncompleteEvidenceError
	var unsettled *domain.AuditUnsettledFindingsError
	if errors.As(e.Cause, &unsettled) && e.Source.ID != "" {
		return fmt.Sprintf("%v; inspect with `tskflwctl audit findings %s` and `tskflwctl audit lint %s`", e.Cause, e.Source.ID, e.Source.ID)
	}
	if errors.As(e.Cause, &incomplete) && e.Source.ID != "" {
		return fmt.Sprintf("%v; inspect with `tskflwctl audit lint %s`", e.Cause, e.Source.ID)
	}
	return e.Cause.Error()
}

func (e *AuditMoveError) Unwrap() error { return e.Cause }
