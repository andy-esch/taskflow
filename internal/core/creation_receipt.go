package core

import "github.com/andy-esch/taskflow/internal/domain"

// LocalCreateOutcome is optional adapter-owned destination evidence. A dry run
// can know the planned path without creating a resolvable record; CommittedPath
// is set only after that destination became durable. Both remain empty for a
// pathless adapter. Neither field is semantic entity identity.
type LocalCreateOutcome struct {
	PlannedPath   string
	CommittedPath string
}

// DisplayPath preserves the existing local create UI contract: previews show
// the planned destination; successful writes show the durable one.
func (o LocalCreateOutcome) DisplayPath(dryRun bool) string {
	if dryRun {
		return o.PlannedPath
	}
	return o.CommittedPath
}

// Create receipts remain entity-specific even though their optional local
// destination evidence has the same shape. Remote adapters need not invent a
// path, and future removal of domain Path does not alter these outcomes.
type TaskCreationReceipt struct {
	Task      domain.Task
	Local     LocalCreateOutcome
	DryRun    bool
	Committed bool
}

type EpicCreationReceipt struct {
	Epic      domain.Epic
	Local     LocalCreateOutcome
	DryRun    bool
	Committed bool
}

type AuditCreationReceipt struct {
	Audit     domain.Audit
	Local     LocalCreateOutcome
	DryRun    bool
	Committed bool
}

type ResearchCreationReceipt struct {
	Research  domain.Research
	Local     LocalCreateOutcome
	DryRun    bool
	Committed bool
}
