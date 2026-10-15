package core

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

type selectedTaskStore struct {
	nopStore
	sourceID, readRef string
	readErr           error
	readConflicts     int
	reads             int
	typedReads        int
	targets           []TaskMutationTarget
	times             []time.Time
	updates           []map[string]any
}

func (s *selectedTaskStore) ReadTask(string) (LoadedRecord[TaskWithBody], error) {
	s.typedReads++
	// A malformed semantic record must not make its identity unaddressable for
	// field repair. Selection tests fail if core goes through the typed read.
	return LoadedRecord[TaskWithBody]{}, domain.ErrValidation
}

func (s *selectedTaskStore) ResolveTaskSource(ref string) (RecordSource, error) {
	s.reads++
	s.readRef = ref
	if s.reads <= s.readConflicts {
		return RecordSource{}, domain.ErrConflict
	}
	return RecordSource{ID: s.sourceID, Location: "opaque:remote/task"}, s.readErr
}

func (s *selectedTaskStore) attempt(target TaskMutationTarget) error {
	s.targets = append(s.targets, target)
	if len(s.targets) <= 2 {
		return domain.ErrConflict
	}
	return nil
}

func (s *selectedTaskStore) SetFields(target TaskMutationTarget, updates map[string]any, _ bool) (domain.Task, error) {
	s.updates = append(s.updates, updates)
	return domain.Task{}, s.attempt(target)
}

func (s *selectedTaskStore) EditBody(target TaskMutationTarget, _ string, _ bool, now time.Time, _ bool) (domain.Task, string, error) {
	s.times = append(s.times, now)
	return domain.Task{}, "", s.attempt(target)
}

func (s *selectedTaskStore) TransformTaskBody(target TaskMutationTarget, now time.Time, _ bool, transform func(string) (string, error)) (domain.Task, string, bool, error) {
	s.times = append(s.times, now)
	body, err := transform("## Acceptance criteria\n\n- [ ] one\n")
	if err != nil {
		return domain.Task{}, "", false, err
	}
	return domain.Task{}, body, true, s.attempt(target)
}

func TestTaskMutationsSelectSourceOnceAndRetainTargetAndClock(t *testing.T) {
	for _, verb := range []string{"fields", "append", "criterion", "replace"} {
		for _, dryRun := range []bool{false, true} {
			t.Run(verb+"/"+map[bool]string{false: "write", true: "preview"}[dryRun], func(t *testing.T) {
				st := &selectedTaskStore{sourceID: "portable-source-identity"}
				now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
				clocks, sleeps := 0, 0
				svc := MustNewService(st, WithClock(func() time.Time { clocks++; return now }), WithRetry(2, func(int) { sleeps++ }))
				var err error
				switch verb {
				case "fields":
					_, err = svc.SetFields("user-alias", map[string]any{"priority": "high"}, false, dryRun)
				case "append":
					_, _, err = svc.AppendBody("user-alias", "note", dryRun)
				case "criterion":
					_, _, _, err = svc.SetAcceptanceCriterion("user-alias", 1, true, dryRun)
				case "replace":
					_, _, err = svc.ReplaceBody("user-alias", "replacement", dryRun)
				}
				wantAttempts := 3
				if dryRun || verb == "replace" {
					wantAttempts = 1
					if !errors.Is(err, domain.ErrConflict) {
						t.Fatalf("non-retried mutation swallowed conflict: %v", err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
				if st.reads != 1 || st.typedReads != 0 || st.readRef != "user-alias" || clocks != 1 || len(st.targets) != wantAttempts || sleeps != wantAttempts-1 {
					t.Fatalf("selection/retry/clock drift: reads=%d ref=%q clocks=%d targets=%v sleeps=%d", st.reads, st.readRef, clocks, st.targets, sleeps)
				}
				for _, target := range st.targets {
					if target.ID != st.sourceID {
						t.Fatalf("used a declaration, alias or location instead of source identity: %+v", target)
					}
				}
				for _, stamp := range st.times {
					if stamp != now {
						t.Fatalf("retry changed timestamp: %v", stamp)
					}
				}
				for _, updates := range st.updates {
					if updates["updated_at"] != "2026-10-14" {
						t.Fatalf("retry changed update stamp: %+v", updates)
					}
				}
			})
		}
	}
}

func TestTaskMutationSelectionRefusesMissingEvidenceAndReadErrors(t *testing.T) {
	for _, tc := range []struct {
		name, sourceID string
		readErr, want  error
	}{
		{"missing", "", nil, domain.ErrValidation},
		{"blank", "  ", nil, domain.ErrValidation},
		{"ambiguous", "id", domain.ErrAmbiguous, domain.ErrAmbiguous},
		{"missing record", "id", domain.ErrNotFound, domain.ErrNotFound},
	} {
		for _, verb := range []string{"fields", "append", "criterion", "replace"} {
			for _, dry := range []bool{false, true} {
				t.Run(tc.name+"/"+verb+"/"+map[bool]string{false: "write", true: "preview"}[dry], func(t *testing.T) {
					st := &selectedTaskStore{sourceID: tc.sourceID, readErr: tc.readErr}
					svc := MustNewService(st)
					var err error
					switch verb {
					case "fields":
						_, err = svc.SetFields("alias", map[string]any{"priority": "high"}, false, dry)
					case "append":
						_, _, err = svc.AppendBody("alias", "note", dry)
					case "criterion":
						_, _, _, err = svc.SetAcceptanceCriterion("alias", 1, true, dry)
					case "replace":
						_, _, err = svc.ReplaceBody("alias", "replacement", dry)
					}
					if !errors.Is(err, tc.want) || len(st.targets) != 0 || st.reads != 1 {
						t.Fatalf("invalid selection reached mutation: err=%v reads=%d targets=%v", err, st.reads, st.targets)
					}
				})
			}
		}
	}
}

func TestTaskMutationTargetDoesNotCarryLocationOrDeclaration(t *testing.T) {
	// Keep the portable contract focused: this is an ID, not a URI/path parser.
	st := &selectedTaskStore{sourceID: "urn:task:opaque"}
	_, _, err := MustNewService(st, WithRetry(0, noopSleep)).AppendBody("human ref", "note", false)
	if !errors.Is(err, domain.ErrConflict) || len(st.targets) != 1 || !strings.HasPrefix(st.targets[0].ID, "urn:") {
		t.Fatalf("core imposed local filename semantics on portable identity: targets=%+v err=%v", st.targets, err)
	}
}

func TestTaskMutationSelectionAndWritesShareOneRetryBudget(t *testing.T) {
	for _, dry := range []bool{false, true} {
		st := &selectedTaskStore{sourceID: "selected", readConflicts: 1}
		sleeps := 0
		_, _, err := MustNewService(st, WithRetry(2, func(int) { sleeps++ })).AppendBody("alias", "note", dry)
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("exhausted combined budget should surface conflict: %v", err)
		}
		if dry {
			if st.reads != 1 || len(st.targets) != 0 || sleeps != 0 {
				t.Fatalf("preview retried selection: reads=%d targets=%v sleeps=%d", st.reads, st.targets, sleeps)
			}
		} else if st.reads != 2 || len(st.targets) != 2 || sleeps != 2 {
			t.Fatalf("selection and writes spent separate budgets: reads=%d targets=%v sleeps=%d", st.reads, st.targets, sleeps)
		}
	}
}
