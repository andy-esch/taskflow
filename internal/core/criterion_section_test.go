package core

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

type criterionReadStore struct{ *fakeStore }

func (s criterionReadStore) ReadTask(ref string) (LoadedRecord[TaskWithBody], error) {
	record, err := s.fakeStore.ReadTask(ref)
	if err == nil {
		record.Value.Body = s.taskBodies[record.Value.Task.Slug]
	}
	return record, err
}

type criterionRetryStore struct {
	selectedTaskStore
	commits int
}

func (s *criterionRetryStore) TransformTaskBody(target TaskMutationTarget, now time.Time, _ bool, transform func(string) (string, error)) (domain.Task, string, bool, error) {
	s.targets = append(s.targets, target)
	s.times = append(s.times, now)
	body := "## Acceptance criteria\n- [ ] first\n"
	if len(s.targets) > 1 {
		body += "## Acceptance\n- [ ] second\n"
	}
	out, err := transform(body)
	if err != nil {
		return domain.Task{}, "", false, err
	}
	if len(s.targets) == 1 {
		return domain.Task{}, "", false, domain.ErrConflict
	}
	s.commits++
	return domain.Task{}, out, true, nil
}

func TestCriterionRetryRevalidatesFreshBodyStructure(t *testing.T) {
	for _, multi := range []bool{false, true} {
		st := &criterionRetryStore{selectedTaskStore: selectedTaskStore{sourceID: "portable-task"}}
		svc := MustNewService(st, WithRetry(1, func(int) {}))
		var err error
		if multi {
			_, _, _, err = svc.SetCriteriaState("alias", []int{1}, domain.CriterionMet, "", false)
		} else {
			_, _, _, err = svc.SetAcceptanceCriterion("alias", 1, true, false)
		}
		if !errors.Is(err, domain.ErrValidation) || len(st.targets) != 2 || st.commits != 0 || st.reads != 1 {
			t.Fatalf("CAS retry reused stale criterion structure: err=%v targets=%v commits=%d", err, st.targets, st.commits)
		}
		if st.targets[0] != st.targets[1] || st.times[0] != st.times[1] {
			t.Fatal("criterion retry lost bound source identity or operation time")
		}
	}
}

func TestCriterionSectionPortableReadsLintAndCompletion(t *testing.T) {
	for _, body := range []string{
		"## Acceptance\n- [x] first\n## Acceptance criteria\n- [ ] second\n",
		"## Notes / acceptance\n- [ ] unknown\n",
	} {
		for _, status := range []domain.Status{domain.StatusInProgress, domain.StatusCompleted} {
			task := graphRecord("criteria", status)
			st := &fakeStore{tasks: []domain.Task{task}, taskBodies: map[string]string{task.Slug: body}}
			svc := MustNewService(criterionReadStore{st})
			if _, list, err := svc.AcceptanceCriteria(task.Slug); !errors.Is(err, domain.ErrValidation) || list != nil {
				t.Fatalf("portable criterion read concealed structure: %+v, %v", list, err)
			}
			results, _, err := svc.Lint()
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, result := range results {
				for _, issue := range result.Issues {
					found = found || issue.Field == "acceptance"
				}
			}
			if !found {
				t.Fatalf("lint omitted structural error for status %s: %+v", status, results)
			}
			for _, override := range []TaskLifecycleOverride{TaskLifecycleOverrideNone, TaskLifecycleOverrideAcceptanceCriteria} {
				_, _, err := ValidateTaskLifecyclePlan(NewTaskGraph([]domain.Task{task}, nil), TaskLifecyclePlan{
					TaskID: task.ID, To: domain.StatusCompleted, Override: override,
				}, body)
				if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "acceptance") {
					t.Fatalf("completion bypassed structure: status=%s override=%s err=%v", status, override, err)
				}
			}
		}
	}
}
