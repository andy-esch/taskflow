package core

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

type entityTargetStore struct {
	nopStore
	sourceCalls      int
	initialConflicts int
	emptySource      bool
	targets          []string
	stamps           []any
}

func (s *entityTargetStore) source() (RecordSource, error) {
	s.sourceCalls++
	if s.sourceCalls <= s.initialConflicts {
		return RecordSource{}, domain.ErrConflict
	}
	if s.emptySource {
		return RecordSource{}, nil
	}
	return RecordSource{ID: "opaque:selected", Location: "not-an-id"}, nil
}

func (s *entityTargetStore) ResolveAuditSource(string) (RecordSource, error)    { return s.source() }
func (s *entityTargetStore) ResolveEpicSource(string) (RecordSource, error)     { return s.source() }
func (s *entityTargetStore) ResolveResearchSource(string) (RecordSource, error) { return s.source() }

func (s *entityTargetStore) mutate(id string, fields map[string]any) error {
	s.targets = append(s.targets, id)
	if fields != nil {
		s.stamps = append(s.stamps, fields["updated_at"])
	}
	if len(s.targets) == 1 {
		return domain.ErrConflict // no identity in the result to recover accidentally
	}
	return nil
}

func (s *entityTargetStore) MoveAudit(target AuditMutationTarget, _ domain.AuditBucket, _ bool) (domain.Audit, error) {
	return domain.Audit{}, s.mutate(target.ID, nil)
}

func (s *entityTargetStore) SetEpicFields(target EpicMutationTarget, fields map[string]any, _ bool) (domain.Epic, error) {
	return domain.Epic{}, s.mutate(target.ID, fields)
}

func (s *entityTargetStore) SetResearchFields(target ResearchMutationTarget, fields map[string]any, _ bool) (domain.Research, error) {
	return domain.Research{}, s.mutate(target.ID, fields)
}

func TestEntityMutationsShareSelectionAndWriteRetryBudget(t *testing.T) {
	for _, family := range []string{"audit", "epic", "research"} {
		for _, dry := range []bool{false, true} {
			for _, initialConflicts := range []int{0, 1, 2} {
				t.Run(fmt.Sprintf("%s/dry=%t/initial=%d", family, dry, initialConflicts), func(t *testing.T) {
					st := &entityTargetStore{initialConflicts: initialConflicts}
					clocks, sleeps := 0, 0
					svc := MustNewService(st, WithRetry(2, func(int) { sleeps++ }), WithClock(func() time.Time {
						clocks++
						return time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
					}))
					var err error
					switch family {
					case "audit":
						_, err = svc.MoveAudit("human-ref", domain.AuditClosed, dry)
					case "epic":
						_, err = svc.SetEpicFields("human-ref", map[string]any{"priority": "high"}, false, dry)
					case "research":
						_, err = svc.SetResearchFields("human-ref", map[string]any{"description": "changed"}, false, dry)
					}
					wantSources, wantWrites, wantSleeps := 1+initialConflicts, 2, 1+initialConflicts
					if dry || initialConflicts == 2 {
						wantSources, wantWrites, wantSleeps = 3, 1, 2
						if dry {
							wantSources, wantWrites, wantSleeps = 1, 0, 0
							if initialConflicts == 0 {
								wantWrites = 1
							}
						}
						if !errors.Is(err, domain.ErrConflict) {
							t.Fatalf("preview should return its first conflict: %v", err)
						}
					} else if err != nil {
						t.Fatal(err)
					}
					if st.sourceCalls != wantSources || len(st.targets) != wantWrites || sleeps != wantSleeps {
						t.Fatalf("selection/retry drift: store=%+v sleeps=%d", st, sleeps)
					}
					for _, target := range st.targets {
						if target != "opaque:selected" {
							t.Fatalf("target drift: %q", target)
						}
					}
					if family != "audit" && clocks != 1 {
						t.Fatalf("clock called %d times", clocks)
					}
					for _, stamp := range st.stamps {
						if stamp != "2026-10-15" {
							t.Fatalf("retry timestamp drift: %v", stamp)
						}
					}
				})
			}
		}
	}
}

func TestSelectedNonTaskActionsRefuseEmptyEvidenceBeforeAdapter(t *testing.T) {
	for _, dry := range []bool{false, true} {
		st := &entityTargetStore{}
		svc := MustNewService(st)
		for _, run := range []func() error{
			func() error { _, e := svc.MoveSelectedAudit(AuditMutationTarget{}, domain.AuditClosed, dry); return e },
			func() error { _, e := svc.MoveSelectedEpic(EpicMutationTarget{}, "retired", dry); return e },
			func() error {
				_, e := svc.SetSelectedEpicFields(EpicMutationTarget{}, map[string]any{"priority": "high"}, false, dry)
				return e
			},
		} {
			if err := run(); !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("empty evidence accepted: %v", err)
			}
		}
		if st.sourceCalls != 0 || len(st.targets) != 0 {
			t.Fatal("empty selected target reached adapter")
		}
	}
}

func TestEntityMutationSelectionRequiresSourceEvidence(t *testing.T) {
	for _, family := range []string{"audit", "epic", "research"} {
		t.Run(family, func(t *testing.T) {
			st := &entityTargetStore{emptySource: true}
			svc := MustNewService(st)
			var err error
			switch family {
			case "audit":
				_, err = svc.MoveAudit("human-ref", domain.AuditClosed, false)
			case "epic":
				_, err = svc.SetEpicFields("human-ref", map[string]any{"priority": "high"}, false, false)
			case "research":
				_, err = svc.SetResearchFields("human-ref", map[string]any{"description": "changed"}, false, false)
			}
			if !errors.Is(err, domain.ErrValidation) || st.sourceCalls != 1 || len(st.targets) != 0 {
				t.Fatalf("missing source evidence reached mutation: err=%v store=%+v", err, st)
			}
		})
	}
}
