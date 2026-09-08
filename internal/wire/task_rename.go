package wire

import "github.com/andy-esch/taskflow/internal/core"

// TaskRenameJSON is the reusable presentation-independent wire receipt. Its
// fields have the same meaning in successful, dry-run, and post-commit failure
// output; Workspace is supplied by the invoking adapter.
type TaskRenameJSON struct {
	TaskID             string        `json:"task_id" jsonschema:"description=stable identity retained across the rename"`
	FromSlug           string        `json:"from_slug" jsonschema:"description=source slug observed when the rename was planned"`
	ToSlug             string        `json:"to_slug" jsonschema:"description=destination slug derived from the requested title"`
	PlannedDocuments   int           `json:"planned_documents" jsonschema:"description=document writes in the guarded rename plan"`
	AppliedDocuments   int           `json:"applied_documents" jsonschema:"description=planned document writes that became durable"`
	PlannedLinks       int           `json:"planned_links" jsonschema:"description=inbound Markdown links the guarded plan will repoint"`
	AppliedLinks       int           `json:"applied_links" jsonschema:"description=planned inbound link rewrites that became durable"`
	Changed            bool          `json:"changed" jsonschema:"description=true when the plan changes at least one document"`
	DryRun             bool          `json:"dry_run" jsonschema:"description=true for a non-durable preview; applied counts and committed remain zero or false"`
	Committed          bool          `json:"committed" jsonschema:"description=true after at least one planned document write became durable"`
	Complete           bool          `json:"complete" jsonschema:"description=true when the entire plan is applied and the destination is the sole authoritative task path; an exact no-op is complete without a commit"`
	DestinationWritten bool          `json:"destination_written" jsonschema:"description=true after the renamed task content became durable at its destination path"`
	SourceRemoved      bool          `json:"source_removed" jsonschema:"description=true when a distinct old source path was removed after destination creation"`
	Workspace          WorkspaceJSON `json:"workspace" jsonschema:"description=planning tree this receipt describes"`
}

// TaskRenameEnvelope is the non-error `task rename --json` receipt for committed
// outcomes and dry-run previews. Task is retained from the earlier generic
// task-mutation shape, making the new receipt fields an additive contract change
// for existing consumers.
type TaskRenameEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	TaskRenameJSON
	Task TaskJSON `json:"task" jsonschema:"description=resulting task on success or prospective task on dry-run"`
}

// TaskRenameRecoveryJSON is the structured durable-prefix diagnosis emitted
// when a task rename fails after one or more document writes committed.
type TaskRenameRecoveryJSON struct {
	TaskRenameJSON
	Remedy string `json:"remedy" jsonschema:"description=operator action appropriate to the durable prefix"`
}

// ToTaskRenameJSON maps the core receipt without adding adapter presentation.
func ToTaskRenameJSON(receipt core.TaskRenameReceipt, workspace WorkspaceJSON) TaskRenameJSON {
	return TaskRenameJSON{
		TaskID: receipt.Task.ID, FromSlug: receipt.FromSlug, ToSlug: receipt.Task.Slug,
		PlannedDocuments: receipt.PlannedDocuments, AppliedDocuments: receipt.AppliedDocuments,
		PlannedLinks: receipt.PlannedLinks, AppliedLinks: receipt.AppliedLinks,
		Changed: receipt.Changed, DryRun: receipt.DryRun, Committed: receipt.Committed, Complete: receipt.Complete,
		DestinationWritten: receipt.DestinationWritten, SourceRemoved: receipt.SourceRemoved,
		Workspace: workspace,
	}
}

// ToTaskRenameEnvelope builds the successful rename envelope value.
func ToTaskRenameEnvelope(receipt core.TaskRenameReceipt, workspace WorkspaceJSON) TaskRenameEnvelope {
	return TaskRenameEnvelope{
		SchemaVersion:  SchemaVersion,
		TaskRenameJSON: ToTaskRenameJSON(receipt, workspace),
		Task:           ToTaskJSON(receipt.Task),
	}
}

// ToTaskRenameRecoveryJSON adds failure-only recovery guidance to the shared
// receipt fields without changing their meanings.
func ToTaskRenameRecoveryJSON(receipt core.TaskRenameReceipt, workspace WorkspaceJSON) TaskRenameRecoveryJSON {
	return TaskRenameRecoveryJSON{
		TaskRenameJSON: ToTaskRenameJSON(receipt, workspace),
		Remedy:         receipt.Remedy,
	}
}
