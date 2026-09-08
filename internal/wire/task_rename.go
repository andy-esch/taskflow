package wire

import "github.com/andy-esch/taskflow/internal/core"

// TaskRenameRecoveryJSON is the structured durable-prefix diagnosis emitted
// when a task rename fails after one or more document writes committed.
type TaskRenameRecoveryJSON struct {
	TaskID             string        `json:"task_id"`
	FromSlug           string        `json:"from_slug"`
	ToSlug             string        `json:"to_slug"`
	PlannedDocuments   int           `json:"planned_documents"`
	AppliedDocuments   int           `json:"applied_documents"`
	PlannedLinks       int           `json:"planned_links"`
	AppliedLinks       int           `json:"applied_links"`
	Changed            bool          `json:"changed"`
	DryRun             bool          `json:"dry_run"`
	Committed          bool          `json:"committed"`
	Complete           bool          `json:"complete"`
	DestinationWritten bool          `json:"destination_written"`
	SourceRemoved      bool          `json:"source_removed"`
	Remedy             string        `json:"remedy"`
	Workspace          WorkspaceJSON `json:"workspace"`
}

func ToTaskRenameRecoveryJSON(receipt core.TaskRenameReceipt, workspace WorkspaceJSON) TaskRenameRecoveryJSON {
	return TaskRenameRecoveryJSON{
		TaskID: receipt.Task.ID, FromSlug: receipt.FromSlug, ToSlug: receipt.Task.Slug,
		PlannedDocuments: receipt.PlannedDocuments, AppliedDocuments: receipt.AppliedDocuments,
		PlannedLinks: receipt.PlannedLinks, AppliedLinks: receipt.AppliedLinks,
		Changed: receipt.Changed, DryRun: receipt.DryRun, Committed: receipt.Committed, Complete: receipt.Complete,
		DestinationWritten: receipt.DestinationWritten, SourceRemoved: receipt.SourceRemoved,
		Remedy: receipt.Remedy, Workspace: workspace,
	}
}
