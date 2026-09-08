package core

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TaskRenameMutationResult is the store-owned outcome of a task rename and its
// multi-document inbound-link cascade. Planned counts describe the guarded plan;
// applied counts are its durable prefix when a later write or cleanup fails.
type TaskRenameMutationResult struct {
	Task               domain.Task
	FromSlug           string
	PlannedDocuments   int
	AppliedDocuments   int
	PlannedLinks       int
	AppliedLinks       int
	Changed            bool
	DryRun             bool
	Committed          bool
	Complete           bool
	DestinationWritten bool
	SourceRemoved      bool
}

// TaskRenameReceipt is the adapter-neutral public result. Committed means at
// least one document changed durably; Complete means the destination is the sole
// authoritative task path and the entire guarded cascade plan was applied.
type TaskRenameReceipt struct {
	Task               domain.Task
	FromSlug           string
	PlannedDocuments   int
	AppliedDocuments   int
	PlannedLinks       int
	AppliedLinks       int
	Changed            bool
	DryRun             bool
	Committed          bool
	Complete           bool
	DestinationWritten bool
	SourceRemoved      bool
	Remedy             string
}

// TaskRenameFailure preserves a durable rename prefix so adapters can distinguish
// a safe retry from the narrower destination-written/source-retained state that
// needs inspection before another mutation.
type TaskRenameFailure struct {
	Cause   error
	Receipt TaskRenameReceipt
}

func (e *TaskRenameFailure) Error() string {
	if e == nil {
		return "task rename partially committed"
	}
	return fmt.Sprintf("task rename committed %d of %d planned document write(s) and %d of %d link rewrite(s) before failing: %v; %s",
		e.Receipt.AppliedDocuments, e.Receipt.PlannedDocuments,
		e.Receipt.AppliedLinks, e.Receipt.PlannedLinks,
		e.Cause, e.Receipt.Remedy)
}

func (e *TaskRenameFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func taskRenameReceipt(result TaskRenameMutationResult) TaskRenameReceipt {
	return TaskRenameReceipt{
		Task: result.Task, FromSlug: result.FromSlug,
		PlannedDocuments: result.PlannedDocuments, AppliedDocuments: result.AppliedDocuments,
		PlannedLinks: result.PlannedLinks, AppliedLinks: result.AppliedLinks,
		Changed: result.Changed, DryRun: result.DryRun, Committed: result.Committed, Complete: result.Complete,
		DestinationWritten: result.DestinationWritten, SourceRemoved: result.SourceRemoved,
	}
}

func taskRenameFailureRemedy(result TaskRenameMutationResult) string {
	switch {
	case result.Complete:
		return "inspect the renamed task by stable id before retrying; the planned rename is already complete"
	case result.DestinationWritten && !result.SourceRemoved:
		return "inspect the old and destination task files before retrying; if the destination is correct, remove the retained old source to restore one stable-id owner"
	case result.Committed:
		return "resolve the reported error, then rerun the same rename by stable task id; the durable link-rewrite prefix is convergent"
	default:
		return ""
	}
}
