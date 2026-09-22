package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func TestEveryFSMutationEntryRequiresAuthorizationBeforeFilesystemEffects(t *testing.T) {
	blocked := errors.New("mutation denied")
	now := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)

	type mutation struct {
		name string
		run  func(*FS, bool) error
	}
	mutations := []mutation{
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
			_, err := s.MoveAudit("missing", domain.AuditOpen, dryRun)
			return err
		}},
		{"AppendAuditBody", func(s *FS, dryRun bool) error {
			_, _, err := s.AppendAuditBody("missing", "text", now, dryRun)
			return err
		}},
		{"EditBody", func(s *FS, dryRun bool) error {
			_, _, err := s.EditBody("missing", "text", false, now, dryRun)
			return err
		}},
		{"TransformAuditBody", func(s *FS, dryRun bool) error {
			_, _, _, err := s.TransformAuditBody("missing", now, dryRun, func(domain.Audit, string) (string, error) { return "", nil })
			return err
		}},
		{"TransformTaskBody", func(s *FS, dryRun bool) error {
			_, _, _, err := s.TransformTaskBody("missing", now, dryRun, func(string) (string, error) { return "", nil })
			return err
		}},
		{"EditTask", func(s *FS, _ bool) error {
			_, _, err := s.EditTask("missing", now, func(current string, _ error) (string, error) { return current, nil })
			return err
		}},
		{"EditAudit", func(s *FS, _ bool) error {
			_, _, err := s.EditAudit("missing", now, func(current string, _ error) (string, error) { return current, nil })
			return err
		}},
		{"FixFrontmatter", func(s *FS, dryRun bool) error {
			_, err := s.FixFrontmatter(dryRun)
			return err
		}},
		{"SetFields", func(s *FS, dryRun bool) error {
			_, err := s.SetFields("missing", map[string]any{"priority": "high"}, dryRun)
			return err
		}},
		{"RenameTask", func(s *FS, dryRun bool) error {
			_, err := s.RenameTask("missing", "Renamed", dryRun)
			return err
		}},
		{"MoveEpic", func(s *FS, dryRun bool) error {
			_, err := s.MoveEpic("missing", "active", now, dryRun)
			return err
		}},
		{"SetEpicFields", func(s *FS, dryRun bool) error {
			_, err := s.SetEpicFields("missing", map[string]any{"priority": "high"}, dryRun)
			return err
		}},
		{"EditEpic", func(s *FS, _ bool) error {
			_, _, err := s.EditEpic("missing", now, func(current string, _ error) (string, error) { return current, nil })
			return err
		}},
		{"SetResearchFields", func(s *FS, dryRun bool) error {
			_, err := s.SetResearchFields("missing", map[string]any{"description": "probe"}, dryRun)
			return err
		}},
		{"EditResearch", func(s *FS, _ bool) error {
			_, _, err := s.EditResearch("missing", now, func(current string, _ error) (string, error) { return current, nil })
			return err
		}},
		{"AppendResearchBody", func(s *FS, dryRun bool) error {
			_, _, err := s.AppendResearchBody("missing", "text", now, dryRun)
			return err
		}},
		{"MutateTaskGraph", func(s *FS, dryRun bool) error {
			_, err := s.MutateTaskGraph(now, dryRun, func(*core.TaskGraph) (core.TaskGraphMutationPlan, error) { return core.TaskGraphMutationPlan{}, nil })
			return err
		}},
		{"MutateTaskGraphRepair", func(s *FS, dryRun bool) error {
			_, err := s.MutateTaskGraphRepair(now, dryRun, func(*core.TaskGraph) (core.TaskGraphRepairPlan, error) { return core.TaskGraphRepairPlan{}, nil })
			return err
		}},
		{"MutateTaskLifecycle", func(s *FS, dryRun bool) error {
			_, err := s.MutateTaskLifecycle(now, dryRun, func(*core.TaskGraph) (core.TaskLifecyclePlan, error) { return core.TaskLifecyclePlan{}, nil })
			return err
		}},
		{"MutateThreadApply", func(s *FS, dryRun bool) error {
			_, err := s.MutateThreadApply(now, dryRun, func(core.ThreadApplySnapshot) (core.ThreadApplyPlan, error) { return core.ThreadApplyPlan{}, nil })
			return err
		}},
		{"MutateThread", func(s *FS, dryRun bool) error {
			_, err := s.MutateThread(now, dryRun, func(core.ThreadMutationSnapshot) (core.ThreadMutationPlan, error) {
				return core.ThreadMutationPlan{}, nil
			})
			return err
		}},
		{"MutateThreadCreation", func(s *FS, dryRun bool) error {
			_, err := s.MutateThreadCreation(now, dryRun, func(core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
				return core.ThreadCreationPlan{}, nil
			})
			return err
		}},
	}

	for _, mutation := range mutations {
		for _, dryRun := range []bool{true, false} {
			mode := map[bool]string{true: "dry-run", false: "write"}[dryRun]
			t.Run(mutation.name+"/"+mode, func(t *testing.T) {
				root := filepath.Join(t.TempDir(), "not-yet-created")
				store := NewFS(root, WithMutationAuthorization(func() error { return blocked }))
				if err := mutation.run(store, dryRun); !errors.Is(err, blocked) {
					t.Fatalf("error = %v, want authorization error", err)
				}
				if _, err := os.Stat(root); !os.IsNotExist(err) {
					t.Fatalf("denied mutation touched planning root: stat error = %v", err)
				}
			})
		}
	}
}
