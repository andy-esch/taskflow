package store

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

type taskTargetOperation struct {
	name string
	run  func(*core.Service, string, bool) error
}

func taskTargetOperations() []taskTargetOperation {
	return []taskTargetOperation{
		{"checkbox", func(s *core.Service, ref string, dry bool) error {
			_, _, _, err := s.SetAcceptanceCriterion(ref, 1, true, dry)
			return err
		}},
		{"state", func(s *core.Service, ref string, dry bool) error {
			_, _, _, err := s.SetCriteriaState(ref, []int{1, 2}, domain.CriterionDeferred, "later", dry)
			return err
		}},
		{"add", func(s *core.Service, ref string, dry bool) error {
			_, _, _, err := s.EditCriteria(ref, dry, func(body string) (string, error) { return domain.AddCriterion(body, "third") })
			return err
		}},
		{"remove", func(s *core.Service, ref string, dry bool) error {
			_, _, _, err := s.EditCriteria(ref, dry, func(body string) (string, error) { return domain.RemoveCriterion(body, 2) })
			return err
		}},
		{"replace", func(s *core.Service, ref string, dry bool) error {
			_, _, _, err := s.EditCriteria(ref, dry, func(body string) (string, error) { return domain.ReplaceCriterionText(body, 1, "rewritten") })
			return err
		}},
		{"tracked", func(s *core.Service, ref string, dry bool) error {
			_, _, _, err := s.SetCriteriaTracked(ref, []int{1}, "alpha-two", "follow up", dry)
			return err
		}},
		{"append", func(s *core.Service, ref string, dry bool) error {
			_, _, err := s.AppendBody(ref, "a new note", dry)
			return err
		}},
		{"fields", func(s *core.Service, ref string, dry bool) error {
			_, err := s.SetFields(ref, map[string]any{"priority": "high"}, false, dry)
			return err
		}},
	}
}

func taskTargetSeed(slug string) string {
	return fmt.Sprintf("---\nid: %s\nstatus: ready-to-start\ndescription: original\ntags: [go]\ntier: 2\n---\n# %s\n\n## Acceptance criteria\n\n- [ ] one\n- [ ] two\n", testutil.TaskID(slug), slug)
}

func taskTargetFixture(t *testing.T) (*FS, string, string) {
	t.Helper()
	root := t.TempDir()
	for _, slug := range []string{"alpha", "alpha-two"} {
		writeTask(t, root, "ready-to-start", slug+".md", taskTargetSeed(slug))
	}
	return testutil.Must(NewFS(root, core.UnrestrictedMutations())),
		filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md"),
		filepath.Join(root, "tasks", testutil.TaskID("alpha-two")+"-alpha-two.md")
}

func TestTaskMutationRetryKeepsSelectedIdentityAfterRename(t *testing.T) {
	for _, op := range taskTargetOperations() {
		for _, reuse := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/reuse=%v", op.name, reuse), func(t *testing.T) {
				fs, source, sibling := taskTargetFixture(t)
				root := filepath.Dir(filepath.Dir(source))
				renamed := filepath.Join(filepath.Dir(source), testutil.TaskID("alpha")+"-omega.md")
				beforeSibling := readFile(t, sibling)
				var beforeTarget, replacement string
				injected := false
				hook := func() {
					if injected {
						return
					}
					injected = true
					if _, err := fs.RenameTask(testutil.TaskID("alpha"), "Omega", false); err != nil {
						t.Fatal(err)
					}
					beforeTarget = readFile(t, renamed)
					if reuse {
						replacement = filepath.Join(root, "tasks", testutil.TaskID("replacement")+"-alpha.md")
						testutil.Write(t, replacement, taskTargetSeed("replacement"))
					}
				}
				testHookBeforeBodyWrite, testHookBeforeSetFieldsWrite = hook, hook
				t.Cleanup(func() { testHookBeforeBodyWrite, testHookBeforeSetFieldsWrite = nil, nil })
				retries := 0
				svc := core.MustNewService(fs, core.WithRetry(1, func(int) { retries++ }))
				if err := op.run(svc, "alpha", false); err != nil || !injected || retries != 1 {
					t.Fatalf("mutation did not retry the selected identity: err=%v injected=%v retries=%d", err, injected, retries)
				}
				if readFile(t, renamed) == beforeTarget || readFile(t, sibling) != beforeSibling {
					t.Fatal("retry changed a sibling or failed to change the selected task")
				}
				if replacement != "" && readFile(t, replacement) != taskTargetSeed("replacement") {
					t.Fatal("retry changed the replacement holding the old alias")
				}
			})
		}
	}
}

func TestTaskMutationRetryNeverFallsBackAfterSelectedIDDisappears(t *testing.T) {
	for _, op := range taskTargetOperations() {
		for _, label := range []string{testutil.TaskID("alpha"), strings.ToUpper(testutil.TaskID("alpha"))} {
			t.Run(op.name+"/"+label, func(t *testing.T) {
				fs, source, sibling := taskTargetFixture(t)
				impostor := filepath.Join(filepath.Dir(source), testutil.TaskID("impostor")+"-"+label+".md")
				testutil.Write(t, impostor, taskTargetSeed("impostor"))
				beforeSibling, beforeImpostor := readFile(t, sibling), readFile(t, impostor)
				injected := false
				hook := func() {
					if !injected {
						injected = true
						if err := os.Remove(source); err != nil {
							t.Fatal(err)
						}
					}
				}
				testHookBeforeBodyWrite, testHookBeforeSetFieldsWrite = hook, hook
				t.Cleanup(func() { testHookBeforeBodyWrite, testHookBeforeSetFieldsWrite = nil, nil })
				retries := 0
				svc := core.MustNewService(fs, core.WithRetry(1, func(int) { retries++ }))
				err := op.run(svc, "alpha", false)
				if !errors.Is(err, domain.ErrNotFound) || !injected || retries != 1 {
					t.Fatalf("missing selected identity was not refused: err=%v injected=%v retries=%d", err, injected, retries)
				}
				if readFile(t, sibling) != beforeSibling || readFile(t, impostor) != beforeImpostor {
					t.Fatal("missing identity fell back to a human label and changed a sibling")
				}
			})
		}
	}
}

func TestSelectedTaskPortsRefuseMissingDuplicateAndInvalidTargetsBeforeCallbacks(t *testing.T) {
	operations := []struct {
		name string
		run  func(*FS, core.TaskMutationTarget, bool, func()) error
	}{
		{"fields", func(fs *FS, target core.TaskMutationTarget, dry bool, _ func()) error {
			_, err := fs.SetFields(target, map[string]any{}, dry)
			return err
		}},
		{"body no-op", func(fs *FS, target core.TaskMutationTarget, dry bool, _ func()) error {
			_, _, err := fs.EditBody(target, "", true, bodyNow, dry)
			return err
		}},
		{"transform no-op", func(fs *FS, target core.TaskMutationTarget, dry bool, callback func()) error {
			_, _, _, err := fs.TransformTaskBody(target, bodyNow, dry, func(body string) (string, error) {
				callback()
				return body, nil
			})
			return err
		}},
	}
	for _, mode := range []string{"missing", "duplicate", "empty", "label", "uppercase"} {
		for _, op := range operations {
			for _, dry := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/preview=%v", mode, op.name, dry), func(t *testing.T) {
					fs, source, _ := taskTargetFixture(t)
					root := filepath.Dir(filepath.Dir(source))
					target := core.TaskMutationTarget{ID: testutil.TaskID("alpha")}
					want := domain.ErrValidation
					switch mode {
					case "missing":
						if err := os.Remove(source); err != nil {
							t.Fatal(err)
						}
						impostor := filepath.Join(filepath.Dir(source), testutil.TaskID("impostor")+"-"+target.ID+".md")
						testutil.Write(t, impostor, taskTargetSeed("impostor"))
						want = domain.ErrNotFound
					case "duplicate":
						testutil.Write(t, filepath.Join(filepath.Dir(source), target.ID+"-copy.md"), readFile(t, source))
						want = domain.ErrAmbiguous
					case "empty":
						target.ID = ""
					case "label":
						target.ID = "alpha"
					case "uppercase":
						target.ID = strings.ToUpper(target.ID)
					}
					before := testutil.SnapshotTree(t, root)
					callbacks := 0
					err := op.run(fs, target, dry, func() { callbacks++ })
					if !errors.Is(err, want) || callbacks != 0 || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
						t.Fatalf("bad selected target reached callbacks/effects: err=%v callbacks=%d want=%v", err, callbacks, want)
					}
					if mode == "duplicate" && (!strings.Contains(err.Error(), "alpha") || !strings.Contains(err.Error(), "copy")) {
						t.Fatalf("collision lost candidate diagnostics: %v", err)
					}
				})
			}
		}
	}
}

func TestTaskInitialReferenceStillResolvesAnIDShapedHumanSlug(t *testing.T) {
	root := t.TempDir()
	label := testutil.TaskID("absent")
	path := filepath.Join(root, "tasks", testutil.TaskID("impostor")+"-"+label+".md")
	testutil.Write(t, path, taskTargetSeed("impostor"))
	svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())))
	if _, err := svc.SetFields(label, map[string]any{"priority": "high"}, false, false); err != nil {
		t.Fatalf("initial human reference semantics changed: %v", err)
	}
	if !strings.Contains(readFile(t, path), "priority: high") {
		t.Fatal("initial reference did not select the ID-shaped human slug")
	}
}

type driftAfterTaskSelection struct {
	*FS
	afterRead func()
}

func (s *driftAfterTaskSelection) ReadTask(ref string) (core.LoadedRecord[core.TaskWithBody], error) {
	record, err := s.FS.ReadTask(ref)
	s.afterSelection(ref, err)
	return record, err
}

func (s *driftAfterTaskSelection) ResolveTaskSource(ref string) (core.RecordSource, error) {
	source, err := s.FS.ResolveTaskSource(ref)
	s.afterSelection(ref, err)
	return source, err
}

func (s *driftAfterTaskSelection) afterSelection(ref string, err error) {
	if err == nil && ref == "alpha" && s.afterRead != nil {
		hook := s.afterRead
		s.afterRead = nil
		hook()
	}
}

func TestTaskSelectedIdentitySurvivesDriftBeforeFirstMutation(t *testing.T) {
	operations := append(taskTargetOperations(), taskTargetOperation{"body replace", func(s *core.Service, ref string, dry bool) error {
		_, _, err := s.ReplaceBody(ref, "# Replacement\n", dry)
		return err
	}})
	for _, mode := range []string{"rename and reuse", "delete", "duplicate"} {
		for _, op := range operations {
			for _, dry := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/preview=%v", mode, op.name, dry), func(t *testing.T) {
					fs, source, _ := taskTargetFixture(t)
					root := filepath.Dir(filepath.Dir(source))
					sourceID := testutil.TaskID("alpha")
					targetPath := filepath.Join(filepath.Dir(source), sourceID+"-omega.md")
					var before map[string]testutil.TreeEntry
					st := &driftAfterTaskSelection{FS: fs}
					st.afterRead = func() {
						switch mode {
						case "rename and reuse":
							if _, err := fs.RenameTask(sourceID, "Omega", false); err != nil {
								t.Fatal(err)
							}
							testutil.Write(t, filepath.Join(filepath.Dir(source), testutil.TaskID("replacement")+"-alpha.md"), taskTargetSeed("replacement"))
						case "delete":
							if err := os.Remove(source); err != nil {
								t.Fatal(err)
							}
							testutil.Write(t, filepath.Join(filepath.Dir(source), testutil.TaskID("impostor")+"-"+sourceID+".md"), taskTargetSeed("impostor"))
						case "duplicate":
							testutil.Write(t, filepath.Join(filepath.Dir(source), sourceID+"-copy.md"), readFile(t, source))
						}
						before = testutil.SnapshotTree(t, root)
					}
					err := op.run(core.MustNewService(st), "alpha", dry)
					if before == nil {
						t.Fatal("selection drift was not exercised")
					}
					want := map[string]error{"delete": domain.ErrNotFound, "duplicate": domain.ErrAmbiguous}[mode]
					if (want == nil && err != nil) || (want != nil && !errors.Is(err, want)) {
						t.Fatalf("wrong selection outcome: got=%v want=%v", err, want)
					}
					if dry || want != nil {
						if !maps.Equal(before, testutil.SnapshotTree(t, root)) {
							t.Fatal("preview or refused selection drift changed files")
						}
					} else {
						// Only the identity selected before drift may change.
						after := testutil.SnapshotTree(t, root)
						key := filepath.Join("tasks", filepath.Base(targetPath))
						if before[key] == after[key] {
							t.Fatal("selected renamed task was not changed")
						}
						delete(before, key)
						delete(after, key)
						if !maps.Equal(before, after) {
							t.Fatal("selection drift changed another record")
						}
					}
				})
			}
		}
	}
}

func TestTaskMutationCASStillRefusesAnIdentityDuplicatedAfterLookup(t *testing.T) {
	for _, op := range taskTargetOperations() {
		t.Run(op.name, func(t *testing.T) {
			fs, source, _ := taskTargetFixture(t)
			root := filepath.Dir(filepath.Dir(source))
			var before map[string]testutil.TreeEntry
			hook := func() {
				if before != nil {
					return
				}
				testutil.Write(t, filepath.Join(filepath.Dir(source), testutil.TaskID("alpha")+"-copy.md"), readFile(t, source))
				before = testutil.SnapshotTree(t, root)
			}
			testHookBeforeBodyWrite, testHookBeforeSetFieldsWrite = hook, hook
			t.Cleanup(func() { testHookBeforeBodyWrite, testHookBeforeSetFieldsWrite = nil, nil })
			sleeps := 0
			err := op.run(core.MustNewService(fs, core.WithRetry(2, func(int) { sleeps++ })), "alpha", false)
			if !errors.Is(err, domain.ErrAmbiguous) || sleeps != 0 || before == nil || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
				t.Fatalf("late collision escaped CAS or was retried: err=%v sleeps=%d exercised=%v", err, sleeps, before != nil)
			}
		})
	}
}

func TestTaskMutationSelectionAndCASContentionShareRepositoryRetryBudget(t *testing.T) {
	fs, source, sibling := taskTargetFixture(t)
	beforeSibling := readFile(t, sibling)
	// Deterministically exercise the real repository-scoped pre-selection
	// conflict, then two post-selection content conflicts. A split retry budget
	// would allow a third write attempt to succeed instead of stopping here.
	release := testutil.Must(fs.enterRepositoryPlanner())
	plannerActive := true
	t.Cleanup(func() {
		if plannerActive {
			release()
		}
		testHookBeforeSetFieldsWrite = nil
	})
	writes, sleeps := 0, 0
	testHookBeforeSetFieldsWrite = func() {
		writes++
		if writes <= 2 {
			testutil.Write(t, source, readFile(t, source)+fmt.Sprintf("\nConcurrent note %d\n", writes))
		}
	}
	svc := core.MustNewService(fs, core.WithRetry(2, func(int) {
		sleeps++
		if plannerActive {
			release()
			plannerActive = false
		}
	}))
	_, err := svc.SetFields("alpha", map[string]any{"priority": "high"}, false, false)
	if !errors.Is(err, domain.ErrConflict) || writes != 2 || sleeps != 2 {
		t.Fatalf("selection and CAS used separate budgets: err=%v writes=%d sleeps=%d", err, writes, sleeps)
	}
	if strings.Contains(readFile(t, source), "priority: high") || readFile(t, sibling) != beforeSibling {
		t.Fatal("exhausted retry committed a field update or changed a sibling")
	}
}
