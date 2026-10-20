package store

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

const selectedCreationEpicBody = "---\nstatus: active\npriority: medium\ndescription: Selected parent\ncustom: preserved\n---\n# Parent\n"

func selectedCreationService(t *testing.T, fs *FS) *core.Service {
	t.Helper()
	return core.MustNewService(fs,
		core.WithIDGen(func() string { return "6gmy4c0x6e31" }),
		core.WithClock(func() time.Time { return time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC) }),
	)
}

func TestSelectedEpicTaskCreationUsesWholeSourceIdentity(t *testing.T) {
	for _, stem := range []string{"01-parent", "202601010000-parent"} {
		for _, dry := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dry=%t", stem, dry), func(t *testing.T) {
				root := t.TempDir()
				writeEpic(t, root, stem+".md", selectedCreationEpicBody)
				writeEpic(t, root, stem+"-sibling.md", selectedCreationEpicBody)
				if err := os.MkdirAll(filepath.Join(root, domain.TasksDir), 0o755); err != nil {
					t.Fatal(err)
				}
				fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
				before := testutil.SnapshotTree(t, root)
				got, err := selectedCreationService(t, fs).NewTaskForSelectedEpic(core.EpicMutationTarget{ID: stem}, core.NewTaskParams{
					Title: "Created", Tags: []string{"test"}, DryRun: dry,
				})
				if err != nil || got.Task.Epic != stem || got.Task.Status != domain.StatusReadyToStart || got.Committed == dry ||
					got.DryRun != dry || got.Local.PlannedPath == "" {
					t.Fatalf("selected create=%+v err=%v", got, err)
				}
				after := testutil.SnapshotTree(t, root)
				if dry {
					if got.Local.CommittedPath != "" || !maps.Equal(before, after) {
						t.Fatal("preview changed files or reported a committed path")
					}
					return
				}
				record, err := fs.ReadTask(got.Task.ID)
				if err != nil || record.Value.Task.Epic != stem || got.Local.CommittedPath != got.Local.PlannedPath {
					t.Fatalf("created record=%+v receipt=%+v err=%v", record, got, err)
				}
				delete(after, filepath.Join(domain.TasksDir, got.Task.ID+"-created.md"))
				if !maps.Equal(before, after) {
					t.Fatal("create changed existing parent/sibling bytes or modes")
				}
			})
		}
	}
}

func TestSelectedEpicTaskCreationRefusesFallbackAndUnreadableParents(t *testing.T) {
	for _, tc := range []struct {
		name, target, parentName, parentBody string
		want                                 error
	}{
		{"prefix is not identity", "01", "01-parent", selectedCreationEpicBody, domain.ErrNotFound},
		{"impostor sibling", "01-parent", "01-parent-two", selectedCreationEpicBody, domain.ErrNotFound},
		{"changed full stem", "01-parent", "01-renamed", selectedCreationEpicBody, domain.ErrNotFound},
		{"declaration is not identity", "01-parent", "02-other", "---\nid: 01-parent\nstatus: active\n---\n", domain.ErrNotFound},
		{"malformed parent", "01-parent", "01-parent", "---\nstatus: [broken\n---\n", domain.ErrValidation},
		{"missing frontmatter", "01-parent", "01-parent", "# Unreadable\n", domain.ErrValidation},
	} {
		for _, dry := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/dry=%t", tc.name, dry), func(t *testing.T) {
				root := t.TempDir()
				writeEpic(t, root, tc.parentName+".md", tc.parentBody)
				before := testutil.SnapshotTree(t, root)
				fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
				got, err := selectedCreationService(t, fs).NewTaskForSelectedEpic(core.EpicMutationTarget{ID: tc.target}, core.NewTaskParams{
					Title: "Created", Tags: []string{"test"}, DryRun: dry,
				})
				if !errors.Is(err, tc.want) || got.Committed || got.Task.ID != "" || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatalf("refusal changed files/receipt: got=%+v err=%v", got, err)
				}
			})
		}
	}
}

func TestSelectedEpicTaskCreationChecksParentAfterWaitingForGuard(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, domain.EpicsDir, "01-parent.md")
	writeEpic(t, root, "01-parent.md", selectedCreationEpicBody)
	writeEpic(t, root, "01-parent-sibling.md", selectedCreationEpicBody)
	// Hold the same process guard as a cooperating writer. The second policy
	// check proves the create reached checkedWriteLock before we change the parent.
	release := processRepositoryLock(root)
	released := false
	t.Cleanup(func() {
		if !released {
			release()
		}
	})
	waiting := make(chan struct{})
	authorizations := 0
	fs := testutil.Must(NewFS(root, core.GuardedMutations(func() error {
		authorizations++
		if authorizations == 2 {
			close(waiting)
		}
		return nil
	})))
	type outcome struct {
		receipt core.TaskCreationReceipt
		err     error
	}
	done := make(chan outcome, 1)
	go func() {
		r, err := selectedCreationService(t, fs).NewTaskForSelectedEpic(core.EpicMutationTarget{ID: "01-parent"}, core.NewTaskParams{
			Title: "Created", Tags: []string{"test"},
		})
		done <- outcome{r, err}
	}()
	select {
	case <-waiting:
	case <-time.After(5 * time.Second):
		t.Fatal("create did not reach guard acquisition")
	}
	if err := os.Rename(parent, filepath.Join(root, domain.EpicsDir, "01-renamed.md")); err != nil {
		t.Fatal(err)
	}
	afterDrift := testutil.SnapshotTree(t, root)
	release()
	released = true
	select {
	case got := <-done:
		if authorizations != 3 {
			t.Fatalf("expected entry, lock-entry and in-guard checks; got %d", authorizations)
		}
		if !errors.Is(got.err, domain.ErrNotFound) || got.receipt.Committed || !maps.Equal(afterDrift, testutil.SnapshotTree(t, root)) {
			t.Fatalf("stale parent was accepted: %+v err=%v", got.receipt, got.err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("create did not return after releasing guard")
	}
}

func TestSelectedEpicTaskCreationReauthorizesAfterWaiting(t *testing.T) {
	root := t.TempDir()
	writeEpic(t, root, "01-parent.md", selectedCreationEpicBody)
	before := testutil.SnapshotTree(t, root)
	release := processRepositoryLock(root)
	released := false
	t.Cleanup(func() {
		if !released {
			release()
		}
	})
	var denied atomic.Bool
	waiting := make(chan struct{})
	checks := 0
	fs := testutil.Must(NewFS(root, core.GuardedMutations(func() error {
		checks++
		if denied.Load() {
			return core.ErrReadOnlyPersistence
		}
		// Evaluate the pre-lock permit before signaling. Otherwise revocation
		// can refuse at check 2 and accidentally hide a missing in-guard check.
		if checks == 2 {
			close(waiting)
		}
		return nil
	})))
	svc := selectedCreationService(t, fs)
	done := make(chan error, 1)
	go func() {
		_, err := svc.NewTaskForSelectedEpic(core.EpicMutationTarget{ID: "01-parent"}, core.NewTaskParams{Title: "Created", Tags: []string{"test"}})
		done <- err
	}()
	select {
	case <-waiting:
	case <-time.After(5 * time.Second):
		t.Fatal("create did not reach guard acquisition")
	}
	denied.Store(true)
	release()
	released = true
	select {
	case err := <-done:
		if checks != 3 {
			t.Fatalf("expected entry, lock-entry and in-guard checks; got %d", checks)
		}
		if !errors.Is(err, core.ErrReadOnlyPersistence) || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
			t.Fatalf("expired policy permitted create: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("create did not return after releasing guard")
	}
}

func TestSelectedEpicTaskCreationKeepsCollisionAndFinalizationReceipts(t *testing.T) {
	for _, collision := range []bool{false, true} {
		t.Run(fmt.Sprintf("collision=%t", collision), func(t *testing.T) {
			root := t.TempDir()
			writeEpic(t, root, "01-parent.md", selectedCreationEpicBody)
			if collision {
				testutil.Write(t, filepath.Join(root, domain.TasksDir, "6gmy4c0x6e31-existing.md"), "unreadable but owns identity")
			}
			fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
			previous := testHookRepositoryUnlockError
			t.Cleanup(func() { testHookRepositoryUnlockError = previous })
			releases, minted := 0, 0
			testHookRepositoryUnlockError = func() error { releases++; return domain.ErrConflict }
			svc := core.MustNewService(fs, core.WithIDGen(func() string { minted++; return "6gmy4c0x6e31" }))
			got, err := svc.NewTaskForSelectedEpic(core.EpicMutationTarget{ID: "01-parent"}, core.NewTaskParams{Title: "Created", Tags: []string{"test"}})
			var cleanup *core.CreateFinalizationError
			if !errors.Is(err, domain.ErrConflict) || !errors.As(err, &cleanup) || releases != 1 || minted != 1 || got.Committed == collision {
				t.Fatalf("lost create failure evidence: got=%+v err=%v releases=%d minted=%d", got, err, releases, minted)
			}
			if !collision {
				if got.Local.CommittedPath == "" || got.Local.CommittedPath != got.Local.PlannedPath {
					t.Fatalf("missing durable destination: %+v", got)
				}
				if _, err := os.Stat(got.Local.CommittedPath); err != nil {
					t.Fatal(err)
				}
			}
			files, err := os.ReadDir(filepath.Join(root, domain.TasksDir))
			if err != nil || len(files) != 1 {
				t.Fatalf("unexpected created/reminted files=%v err=%v", files, err)
			}
		})
	}
}

func TestSelectedEpicTaskCreationRefusesReadOnlyAndMismatchedParent(t *testing.T) {
	for _, dry := range []bool{false, true} {
		root := t.TempDir()
		writeEpic(t, root, "01-parent.md", "unreadable")
		before := testutil.SnapshotTree(t, root)
		fs := testutil.Must(NewFS(root, core.ReadOnlyMutations()))
		if _, err := selectedCreationService(t, fs).NewTaskForSelectedEpic(core.EpicMutationTarget{ID: "01-parent"}, core.NewTaskParams{Title: "Created", Tags: []string{"test"}, DryRun: dry}); !errors.Is(err, core.ErrReadOnlyPersistence) {
			t.Fatalf("read-only policy did not precede parent read: %v", err)
		}
		writable := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
		_, err := writable.CreateTaskForSelectedEpic(core.EpicMutationTarget{ID: "01-parent"}, domain.Task{Epic: "02-other"}, "", dry)
		if !errors.Is(err, domain.ErrValidation) || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
			t.Fatalf("mismatched parent changed files: %v", err)
		}
	}
}
