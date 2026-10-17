package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type renameOutcomeStore struct {
	nopStore
	outcome TaskRenameMutationResult
	cause   error
	calls   int
}

func (s *renameOutcomeStore) RenameTask(string, string, bool) (TaskRenameMutationResult, error) {
	s.calls++
	return s.outcome, s.cause
}

func TestRenameTaskPreservesCauseAndSelectsRecoveryFromDurableActions(t *testing.T) {
	for _, tc := range []struct {
		name                         string
		committed, written, complete bool
		want                         string
	}{
		{name: "uncommitted"},
		{name: "cascade prefix", committed: true, want: "rerun the same rename by stable task id"},
		{name: "destination written", committed: true, written: true, want: "only after verifying the destination preserves the intended content"},
		{name: "completed", committed: true, written: true, complete: true, want: "already complete"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cause := errors.New("adapter operation failed")
			appliedDocuments, appliedLinks := 0, 0
			if tc.committed {
				appliedDocuments, appliedLinks = 1, 2
			}
			if tc.written {
				appliedDocuments, appliedLinks = 2, 3
			}
			fake := &renameOutcomeStore{cause: cause, outcome: TaskRenameMutationResult{
				Task: domain.Task{ID: "opaque-task-key", Slug: "new"}, FromSlug: "old",
				PlannedDocuments: 2, AppliedDocuments: appliedDocuments, PlannedLinks: 3, AppliedLinks: appliedLinks,
				Changed: true, Committed: tc.committed, DestinationWritten: tc.written,
				Complete: tc.complete, SourceRemoved: tc.complete,
			}}
			receipt, err := MustNewService(fake).RenameTask("old", "New", false)
			if !errors.Is(err, cause) || fake.calls != 1 || receipt.Task.ID != "opaque-task-key" || receipt.Committed != tc.committed || receipt.DestinationWritten != tc.written || receipt.SourceRemoved != tc.complete || receipt.Complete != tc.complete {
				t.Fatalf("error/outcome lost or rename replayed: %+v err=%v calls=%d", receipt, err, fake.calls)
			}
			if receipt.AppliedDocuments != appliedDocuments || receipt.PlannedDocuments != 2 || receipt.AppliedLinks != appliedLinks || receipt.PlannedLinks != 3 || receipt.SourcePath != "" || receipt.DestinationPath != "" {
				t.Fatalf("portable receipt was re-derived or gained invented paths: %+v", receipt)
			}
			var partial *TaskRenameFailure
			if errors.As(err, &partial) != tc.committed {
				t.Fatalf("durable failure classification changed: %v", err)
			}
			if tc.want == "" {
				if receipt.Remedy != "" || err != cause {
					t.Fatalf("uncommitted failure gained durable recovery: %+v err=%v", receipt, err)
				}
			} else if !strings.Contains(receipt.Remedy, tc.want) {
				t.Fatalf("wrong recovery guidance for historical actions: %+v", receipt)
			}
		})
	}
}
