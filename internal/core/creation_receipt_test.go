package core

import (
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Portable creation outcomes carry semantic values without a local outcome.
// Thread keeps a transitional path until its separate mutation migration.
type pathlessCreationStore struct{ *fakeStore }

func (s *pathlessCreationStore) CreateTask(task domain.Task, _ string, dry bool) (TaskCreationReceipt, error) {
	return TaskCreationReceipt{Task: task, DryRun: dry, Committed: !dry}, nil
}

func (s *pathlessCreationStore) CreateEpic(slug string, epic domain.Epic, _ string, dry bool) (EpicCreationReceipt, error) {
	epic.ID = slug
	return EpicCreationReceipt{Epic: epic, DryRun: dry, Committed: !dry}, nil
}

func (s *pathlessCreationStore) CreateAudit(audit domain.Audit, _ string, dry bool) (AuditCreationReceipt, error) {
	return AuditCreationReceipt{Audit: audit, DryRun: dry, Committed: !dry}, nil
}

func (s *pathlessCreationStore) CreateResearch(research domain.Research, _ string, dry bool) (ResearchCreationReceipt, error) {
	return ResearchCreationReceipt{Research: research, DryRun: dry, Committed: !dry}, nil
}

func (s *pathlessCreationStore) MutateThreadCreation(_ time.Time, dry bool, _ ThreadCreationPlanner) (ThreadCreationMutationResult, error) {
	return ThreadCreationMutationResult{
		Thread: domain.Thread{ID: "6gdx7mn9f0a4", Slug: "created", Path: "/not-the-thread-outcome"},
		DryRun: dry, Changed: true, Committed: !dry,
	}, nil
}

func (s *pathlessCreationStore) MutateThread(_ time.Time, dry bool, _ ThreadMutationPlanner) (ThreadMutationResult, error) {
	return ThreadMutationResult{
		Plan:   ThreadMutationPlan{Operation: ThreadMutationStart},
		Thread: domain.Thread{ID: "6gdx7mn9f0a4", Slug: "created", Path: "/not-the-thread-update-outcome"},
		DryRun: dry, Changed: true, Committed: !dry,
	}, nil
}

func TestCreateReceiptsDoNotInferLocalPathsFromDomainRecords(t *testing.T) {
	adapter := &pathlessCreationStore{fakeStore: &fakeStore{epics: []domain.Epic{{ID: "01-domain"}}}}
	svc := MustNewService(adapter, WithClock(fixedClock("2026-09-27")), WithIDGen(func() string { return "6gdx7mn9f0a5" }))
	for _, dry := range []bool{true, false} {
		task, err := svc.NewTask(NewTaskParams{Title: "Created", Epic: "01-domain", Tags: []string{"test"}, DryRun: dry})
		if err != nil || task.Task.ID == "" || task.Local != (LocalCreateOutcome{}) || task.Committed == dry {
			t.Fatalf("pathless task receipt=%+v err=%v", task, err)
		}
		epic, err := svc.NewEpic(NewEpicParams{Title: "Created", Description: "goal", Status: domain.EpicStatusActive, Priority: "medium", DryRun: dry})
		if err != nil || epic.Epic.ID == "" || epic.Local != (LocalCreateOutcome{}) || epic.Committed == dry {
			t.Fatalf("pathless epic receipt=%+v err=%v", epic, err)
		}
		audit, err := svc.NewAudit(NewAuditParams{Area: "Created", DryRun: dry})
		if err != nil || audit.Audit.ID == "" || audit.Local != (LocalCreateOutcome{}) || audit.Committed == dry {
			t.Fatalf("pathless audit receipt=%+v err=%v", audit, err)
		}
		research, err := svc.NewResearch(NewResearchParams{Title: "Created", DryRun: dry})
		if err != nil || research.Research.ID == "" || research.Local != (LocalCreateOutcome{}) || research.Committed == dry {
			t.Fatalf("pathless research receipt=%+v err=%v", research, err)
		}
		thread, err := svc.NewThread(NewThreadParams{Title: "Created", Description: "goal", Goal: "done", DryRun: dry})
		if err != nil || thread.Thread.Path == "" || thread.Local != (LocalCreateOutcome{}) || thread.Committed == dry {
			t.Fatalf("pathless Thread receipt=%+v err=%v", thread, err)
		}
		updated, err := svc.StartThread("created", dry)
		if err != nil || updated.Thread.Path == "" || updated.LocalPath != "" || updated.Committed == dry {
			t.Fatalf("pathless Thread update receipt=%+v err=%v", updated, err)
		}
	}
}

func TestRenameRecoveryMentionsOnlyAvailableLocalPaths(t *testing.T) {
	for _, tc := range []struct {
		local       TaskRenameReceipt
		want, avoid string
	}{
		{TaskRenameReceipt{}, "", "source"},
		{TaskRenameReceipt{SourcePath: "/planning/tasks/old.md"}, "source \"/planning/tasks/old.md\"", "destination"},
		{TaskRenameReceipt{DestinationPath: "/planning/tasks/new.md"}, "destination \"/planning/tasks/new.md\"", "source"},
	} {
		got := taskRenamePathDiagnosis(tc.local)
		if (tc.want == "" && got != "") || (tc.want != "" && !strings.Contains(got, tc.want)) ||
			(tc.avoid != "" && strings.Contains(got, tc.avoid)) {
			t.Errorf("path diagnosis for %+v = %q", tc.local, got)
		}
	}
}
