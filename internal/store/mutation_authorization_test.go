package store

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

type authorizationMutation struct {
	name string
	run  func(*FS, bool) error
}

// Share the entry inventory between absent and populated fixtures, but make
// callback/no-op behavior observable rather than treating a missing record as proof.
func authorizationMutations(now time.Time, ref string, callback func()) []authorizationMutation {
	return []authorizationMutation{
		{"CreateTask", func(s *FS, dryRun bool) error {
			_, err := s.CreateTask(domain.Task{ID: "6fjangd7kvh0", Slug: "probe", Status: domain.StatusReadyToStart, Created: "2026-09-22"}, "", dryRun)
			return err
		}},
		{"CreateAudit", func(s *FS, dryRun bool) error {
			_, err := s.CreateAudit(domain.Audit{ID: "6fjangd7kvh1", Slug: "probe", Area: "test", Date: "2026-09-22"}, "", dryRun)
			return err
		}},
		{"CreateResearch", func(s *FS, dryRun bool) error {
			_, err := s.CreateResearch(domain.Research{ID: "6fjangd7kvh2", Slug: "probe", Created: "2026-09-22"}, "", dryRun)
			return err
		}},
		{"CreateEpic", func(s *FS, dryRun bool) error {
			_, err := s.CreateEpic("probe", domain.Epic{Status: "active", Created: "2026-09-22"}, "", dryRun)
			return err
		}},
		{"MoveAudit", func(s *FS, dryRun bool) error {
			_, err := s.MoveAudit(ref, domain.AuditOpen, dryRun)
			return err
		}},
		{"AppendAuditBody", func(s *FS, dryRun bool) error {
			_, _, err := s.AppendAuditBody(ref, "", now, dryRun)
			return err
		}},
		{"EditBody", func(s *FS, dryRun bool) error {
			_, _, err := s.EditBody(ref, "", true, now, dryRun)
			return err
		}},
		{"TransformAuditBody", func(s *FS, dryRun bool) error {
			_, _, _, err := s.TransformAuditBody(ref, now, dryRun, func(_ domain.Audit, body string) (string, error) { callback(); return body, nil })
			return err
		}},
		{"TransformTaskBody", func(s *FS, dryRun bool) error {
			_, _, _, err := s.TransformTaskBody(ref, now, dryRun, func(body string) (string, error) { callback(); return body, nil })
			return err
		}},
		{"EditTask", func(s *FS, _ bool) error {
			_, _, err := s.EditTask(ref, now, func(current string, _ error) (string, error) { callback(); return current, nil })
			return err
		}},
		{"EditAudit", func(s *FS, _ bool) error {
			_, _, err := s.EditAudit(ref, now, func(current string, _ error) (string, error) { callback(); return current, nil })
			return err
		}},
		{"FixFrontmatter", func(s *FS, dryRun bool) error {
			_, err := s.FixFrontmatter(dryRun)
			return err
		}},
		{"SetFields", func(s *FS, dryRun bool) error {
			_, err := s.SetFields(ref, map[string]any{}, dryRun)
			return err
		}},
		{"RenameTask", func(s *FS, dryRun bool) error {
			_, err := s.RenameTask(ref, "Renamed", dryRun)
			return err
		}},
		{"MoveEpic", func(s *FS, dryRun bool) error {
			_, err := s.MoveEpic(ref, "active", now, dryRun)
			return err
		}},
		{"SetEpicFields", func(s *FS, dryRun bool) error {
			_, err := s.SetEpicFields(ref, map[string]any{}, dryRun)
			return err
		}},
		{"EditEpic", func(s *FS, _ bool) error {
			_, _, err := s.EditEpic(ref, now, func(current string, _ error) (string, error) { callback(); return current, nil })
			return err
		}},
		{"SetResearchFields", func(s *FS, dryRun bool) error {
			_, err := s.SetResearchFields(ref, map[string]any{}, dryRun)
			return err
		}},
		{"EditResearch", func(s *FS, _ bool) error {
			_, _, err := s.EditResearch(ref, now, func(current string, _ error) (string, error) { callback(); return current, nil })
			return err
		}},
		{"AppendResearchBody", func(s *FS, dryRun bool) error {
			_, _, err := s.AppendResearchBody(ref, "", now, dryRun)
			return err
		}},
		{"MutateTaskGraph", func(s *FS, dryRun bool) error {
			_, err := s.MutateTaskGraph(now, dryRun, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) {
				callback()
				return core.TaskGraphMutationPlan{}, nil
			})
			return err
		}},
		{"MutateTaskGraphRepair", func(s *FS, dryRun bool) error {
			_, err := s.MutateTaskGraphRepair(now, dryRun, func(*core.TaskGraph) (core.TaskGraphRepairPlan, error) {
				callback()
				return core.TaskGraphRepairPlan{}, nil
			})
			return err
		}},
		{"MutateTaskLifecycle", func(s *FS, dryRun bool) error {
			_, err := s.MutateTaskLifecycle(now, dryRun, func(*core.TaskGraph) (core.TaskLifecyclePlan, error) {
				callback()
				return core.TaskLifecyclePlan{}, nil
			})
			return err
		}},
		{"MutateThreadApply", func(s *FS, dryRun bool) error {
			_, err := s.MutateThreadApply(now, dryRun, func(core.ThreadApplySnapshot) (core.ThreadApplyPlan, error) {
				callback()
				return core.ThreadApplyPlan{}, nil
			})
			return err
		}},
		{"MutateThread", func(s *FS, dryRun bool) error {
			_, err := s.MutateThread(now, dryRun, func(core.ThreadMutationSnapshot) (core.ThreadMutationPlan, error) {
				callback()
				return core.ThreadMutationPlan{}, nil
			})
			return err
		}},
		{"MutateThreadCreation", func(s *FS, dryRun bool) error {
			_, err := s.MutateThreadCreation(now, dryRun, func(core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
				callback()
				return core.ThreadCreationPlan{}, nil
			})
			return err
		}},
	}
}

func TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects(t *testing.T) {
	blocked := errors.New("mutation denied")
	now := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	for _, policy := range []struct {
		name  string
		value core.MutationPolicy
		want  error
	}{
		{"guarded", core.GuardedMutations(func() error { return blocked }), blocked},
		{"read-only", core.ReadOnlyMutations(), core.ErrReadOnlyPersistence},
		{"zero-value adapter", core.MutationPolicy{}, core.ErrInvalidMutationPolicy},
	} {
		for _, mutation := range authorizationMutations(now, "missing", func() { t.Error("denied mutation invoked callback") }) {
			for _, dryRun := range []bool{true, false} {
				mode := map[bool]string{true: "dry-run", false: "write"}[dryRun]
				t.Run(policy.name+"/"+mutation.name+"/"+mode, func(t *testing.T) {
					root := filepath.Join(t.TempDir(), "not-yet-created")
					store := &FS{}
					if policy.name != "zero-value adapter" {
						store = testutil.Must(NewFS(root, policy.value))
					}
					if err := mutation.run(store, dryRun); !errors.Is(err, policy.want) {
						t.Fatalf("error = %v, want %v", err, policy.want)
					}
					if _, err := os.Stat(root); !os.IsNotExist(err) {
						t.Fatalf("denied mutation touched planning root: stat error = %v", err)
					}
				})
			}
		}
	}
}

func TestPopulatedFSMutationsRefuseBeforeCallbacksAndNoops(t *testing.T) {
	denied := errors.New("mutation denied")
	now := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	for _, mode := range []string{"read-only", "guarded", "zero policy"} {
		for _, dryRun := range []bool{true, false} {
			t.Run(mode+map[bool]string{true: "/preview", false: "/commit"}[dryRun], func(t *testing.T) {
				r := testutil.NewRepo(t)
				header := fmt.Sprintf("---\nschema: 1\nid: %s\n", testutil.TaskID("probe"))
				r.Task("ready-to-start", "probe.md", header+"created: \"2026-10-05\"\ndescription: probe\ntags: [test]\n---\n# Task\n")
				r.Audit("open", "probe.md", header+"area: test\ndate: \"2026-10-05\"\n---\n# Audit\n")
				r.Epic("01-probe.md", "---\nstatus: active\ndescription: probe\n---\n# Epic\n")
				r.Research("probe.md", header+"created: \"2026-10-05\"\ndescription: probe\n---\n# Research\n")
				policy, want := core.ReadOnlyMutations(), core.ErrReadOnlyPersistence
				if mode == "guarded" {
					policy, want = core.GuardedMutations(func() error { return denied }), denied
				}
				s := testutil.Must(NewFS(r.Root, policy))
				if mode == "zero policy" {
					s.mutationPolicy, want = core.MutationPolicy{}, core.ErrInvalidMutationPolicy
				}
				// Missing references can hide premature callbacks. Require readable
				// records in every family before using this fixture as denial evidence.
				if _, _, err := s.GetTask("probe"); err != nil {
					t.Fatal(err)
				}
				if _, _, err := s.GetAudit("probe"); err != nil {
					t.Fatal(err)
				}
				if _, _, err := s.GetEpic("probe"); err != nil {
					t.Fatal(err)
				}
				if _, _, err := s.GetResearch("probe"); err != nil {
					t.Fatal(err)
				}
				if graph, err := core.LoadTaskGraph(s); err != nil || graph.Health() != core.GraphHealthy {
					t.Fatalf("fixture must not mask planner execution with an invalid graph: graph=%v err=%v", graph, err)
				}
				before := testutil.SnapshotTree(t, r.Root)
				callbacks := 0
				for _, mutation := range authorizationMutations(now, "probe", func() { callbacks++ }) {
					t.Run(mutation.name, func(t *testing.T) {
						if err := mutation.run(s, dryRun); !errors.Is(err, want) {
							t.Fatalf("mutation error=%v; want %v", err, want)
						}
						if callbacks != 0 {
							t.Fatalf("denied mutation executed %d callback(s)", callbacks)
						}
						if !maps.Equal(before, testutil.SnapshotTree(t, r.Root)) {
							t.Fatal("denied mutation changed planning bytes, entries, or modes")
						}
					})
				}
			})
		}
	}
}

func TestInvalidPolicyDoesNotExecuteConstructorOptions(t *testing.T) {
	for _, policy := range []core.MutationPolicy{{}, core.GuardedMutations(nil)} {
		called := false
		s, err := NewFS(t.TempDir(), policy, func(*FS) { called = true })
		if s != nil || !errors.Is(err, core.ErrInvalidMutationPolicy) || called {
			t.Fatalf("invalid construction: store=%v err=%v optionCalled=%v", s, err, called)
		}
	}
}
