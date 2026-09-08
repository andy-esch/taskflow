package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestWriteErrorCarriesStructuredTaskRenameRecovery(t *testing.T) {
	receipt := core.TaskRenameReceipt{
		Task:               domain.Task{ID: "6g7wxs43g7nh", Slug: "new-title"},
		FromSlug:           "old-title",
		PlannedDocuments:   3,
		AppliedDocuments:   2,
		PlannedLinks:       2,
		AppliedLinks:       2,
		Changed:            true,
		Committed:          true,
		DestinationWritten: true,
		Remedy:             "inspect both task paths before retrying",
	}
	cause := &core.TaskRenameFailure{Cause: domain.ErrConflict, Receipt: receipt}
	err := &taskRenameCommandFailure{
		cause: cause, receipt: receipt,
		workspace: wire.WorkspaceJSON{PlanningRoot: "/repo/planning", Source: wire.WorkspaceSourceConfig},
	}

	var out bytes.Buffer
	WriteError(&out, err, true)
	var envelope wire.ErrorEnvelope
	if decodeErr := json.Unmarshal(out.Bytes(), &envelope); decodeErr != nil {
		t.Fatalf("decode task rename recovery: %v\n%s", decodeErr, out.String())
	}
	got := envelope.Error.TaskRename
	if envelope.Error.Code != "conflict" || got == nil || got.TaskID != receipt.Task.ID ||
		got.FromSlug != receipt.FromSlug || got.ToSlug != receipt.Task.Slug ||
		got.PlannedDocuments != 3 || got.AppliedDocuments != 2 ||
		!got.Committed || got.Complete || !got.DestinationWritten || got.SourceRemoved ||
		got.Remedy != receipt.Remedy || got.Workspace.PlanningRoot != "/repo/planning" ||
		!errors.Is(err, domain.ErrConflict) {
		t.Fatalf("task rename recovery envelope = %+v", envelope)
	}
}
