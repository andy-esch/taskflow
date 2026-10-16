package store

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// Existing adapter fixtures establish identity from their seed, not from a
// runtime reference lookup. Keep their write tests on that explicit contract.
func testAuditTarget(seed string) core.AuditMutationTarget {
	if id.Valid(seed) {
		return core.AuditMutationTarget{ID: seed}
	}
	return core.AuditMutationTarget{ID: testutil.TaskID(seed)}
}

func testResearchTarget(seed string) core.ResearchMutationTarget {
	if id.Valid(seed) {
		return core.ResearchMutationTarget{ID: seed}
	}
	return core.ResearchMutationTarget{ID: testutil.TaskID(seed)}
}

func TestNonTaskRetriesRetainSelectedSourceIdentity(t *testing.T) {
	tests := []struct {
		name string
		kind string
		hook *func()
		run  func(*core.Service) error
	}{
		{"audit move", "audit", &testHookBeforeMoveAuditWrite, func(s *core.Service) error {
			_, err := s.MoveAudit("alpha", domain.AuditClosed, false)
			return err
		}},
		{"audit append", "audit", &testHookBeforeBodyWrite, func(s *core.Service) error {
			_, _, err := s.AppendAuditBody("alpha", "new narrative", false)
			return err
		}},
		{"audit finding", "audit", &testHookBeforeBodyWrite, func(s *core.Service) error {
			note := "updated note"
			_, _, err := s.EditFinding("alpha", "M1", core.FindingEdit{Note: &note}, false)
			return err
		}},
		{"audit finding creation", "audit", &testHookBeforeBodyWrite, func(s *core.Service) error {
			_, err := s.NewFinding("alpha", core.NewFindingParams{Band: "L", Title: "new finding"})
			return err
		}},
		{"audit headers", "audit", &testHookBeforeBodyWrite, func(s *core.Service) error {
			_, err := s.FixFindingHeaders(false)
			return err
		}},
		{"research fields", "research", &testHookBeforeResearchWrite, func(s *core.Service) error {
			_, err := s.SetResearchFields("alpha", map[string]any{"description": "changed"}, false, false)
			return err
		}},
		{"research append", "research", &testHookBeforeBodyWrite, func(s *core.Service) error {
			_, _, err := s.AppendResearchBody("alpha", "new narrative", false)
			return err
		}},
		{"epic fields", "epic", &testHookBeforeEpicWrite, func(s *core.Service) error {
			_, err := s.SetEpicFields("01-alpha", map[string]any{"priority": "high"}, false, false)
			return err
		}},
		{"epic move", "epic", &testHookBeforeEpicWrite, func(s *core.Service) error {
			_, err := s.MoveEpic("01-alpha", "retired", false)
			return err
		}},
	}
	for _, tt := range tests {
		for _, mode := range []string{"renamed", "missing"} {
			t.Run(tt.name+"/"+mode, func(t *testing.T) {
				root := t.TempDir()
				dir := filepath.Join(root, tt.kind+"s")
				if tt.kind == "research" {
					dir = filepath.Join(root, domain.ResearchDir)
				}
				name := func(slug string) string { return testutil.TaskID(slug) + "-" + slug + ".md" }
				body := func(slug string) string {
					return "---\nid: " + testutil.TaskID(slug) + "\nbucket: open\narea: identity\ndate: 2026-10-15\ncreated: 2026-10-15\nstatus: active\npriority: medium\ndescription: source fixture\n---\n# " + slug + "\n\n#### M1. Existing · **Status:** fixed\n"
				}
				if tt.kind == "epic" {
					name = func(slug string) string { return "01-" + slug + ".md" }
				}
				selected := filepath.Join(dir, name("alpha"))
				sibling := filepath.Join(dir, name("alpha-two"))
				testutil.Write(t, selected, body("alpha"))
				testutil.Write(t, sibling, body("alpha-two"))
				if tt.name == "audit headers" {
					testutil.Write(t, selected, body("alpha")+"\n#### H-2. Repair me · **Status:** open\n")
				}
				store := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
				var afterDrift map[string]testutil.TreeEntry
				fired := false
				oldHook := *tt.hook
				t.Cleanup(func() { *tt.hook = oldHook })
				*tt.hook = func() {
					if fired {
						return
					}
					fired = true
					if mode == "renamed" {
						renamed := filepath.Join(dir, strings.TrimSuffix(filepath.Base(selected), "alpha.md")+"omega.md")
						if err := os.Rename(selected, renamed); err != nil {
							t.Fatal(err)
						}
						selected = renamed
						if tt.kind != "epic" {
							sibling = filepath.Join(dir, testutil.TaskID("replacement")+"-alpha.md")
							testutil.Write(t, sibling, body("replacement"))
						}
					} else if err := os.Remove(selected); err != nil {
						t.Fatal(err)
					}
					afterDrift = testutil.SnapshotTree(t, root)
				}
				err := tt.run(core.MustNewService(store, core.WithRetry(2, func(int) {}), core.WithClock(func() time.Time { return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC) })))
				if !fired {
					t.Fatalf("probe never reached pre-write seam: %v", err)
				}
				after := testutil.SnapshotTree(t, root)
				if mode == "missing" || tt.kind == "epic" {
					if !errors.Is(err, domain.ErrNotFound) || !maps.Equal(afterDrift, after) {
						t.Fatalf("missing selected identity was reselected: err=%v unchanged=%v", err, maps.Equal(afterDrift, after))
					}
					return
				}
				if err != nil {
					t.Fatalf("renamed selected source failed: %v", err)
				}
				key := filepath.Join(tt.kind+"s", filepath.Base(selected))
				if tt.kind == "research" {
					key = filepath.Join(domain.ResearchDir, filepath.Base(selected))
				}
				if afterDrift[key] == after[key] {
					t.Fatal("selected source did not receive the write")
				}
				delete(afterDrift, key)
				delete(after, key)
				if !maps.Equal(afterDrift, after) {
					t.Fatal("retry changed a sibling rather than the selected source")
				}
			})
		}
	}
}

func TestAuditHeaderRepairUsesOccurrenceIdentityWithDuplicateLabels(t *testing.T) {
	for _, dry := range []bool{false, true} {
		t.Run(map[bool]string{false: "write", true: "preview"}[dry], func(t *testing.T) {
			root := t.TempDir()
			for _, seed := range []string{"first", "second"} {
				testutil.Write(t, filepath.Join(root, domain.AuditsDir, testutil.TaskID(seed)+"-same-label.md"), "---\nid: "+testutil.TaskID(seed)+"\nbucket: open\narea: identity\ndate: 2026-10-15\n---\n# Audit\n\n#### H-1. Repair me · **Status:** open\n")
			}
			before := testutil.SnapshotTree(t, root)
			svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())))
			fixes, err := svc.FixFindingHeaders(dry)
			if err != nil || len(fixes) != 2 {
				t.Fatalf("repair confused equal labels: fixes=%+v err=%v", fixes, err)
			}
			after := testutil.SnapshotTree(t, root)
			if dry {
				if !maps.Equal(before, after) {
					t.Fatal("preview changed an occurrence")
				}
				return
			}
			for _, seed := range []string{"first", "second"} {
				record, err := svc.ShowAudit(testutil.TaskID(seed))
				if err != nil || len(domain.ParseFindings(record.Value.Body)) != 1 || len(domain.NearMissFindingHeaders(record.Value.Body)) != 0 {
					t.Fatalf("occurrence %s not repaired: record=%+v err=%v", seed, record, err)
				}
			}
		})
	}
}

func TestSelectedAuditPortsRefuseDuplicateIdentityBeforeCallbackOrNoop(t *testing.T) {
	root := t.TempDir()
	selectedID := testutil.TaskID("duplicate")
	for _, label := range []string{"one", "two"} {
		testutil.Write(t, filepath.Join(root, domain.AuditsDir, selectedID+"-"+label+".md"), "---\nbucket: open\narea: identity\ndate: 2026-10-15\n---\n# Audit\n")
	}
	fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
	before := testutil.SnapshotTree(t, root)
	for _, dry := range []bool{false, true} {
		if _, err := fs.MoveAudit(core.AuditMutationTarget{ID: selectedID}, domain.AuditOpen, dry); !errors.Is(err, domain.ErrAmbiguous) {
			t.Fatalf("ambiguous noop accepted: %v", err)
		}
		called := false
		_, _, _, err := fs.TransformAuditBody(core.AuditMutationTarget{ID: selectedID}, time.Now(), dry, func(_ domain.Audit, body string) (string, error) { called = true; return body, nil })
		if !errors.Is(err, domain.ErrAmbiguous) || called {
			t.Fatalf("ambiguous identity reached callback: called=%v err=%v", called, err)
		}
	}
	if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("ambiguous identity changed records")
	}
}
