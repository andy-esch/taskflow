package store

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func inplaceRepairFixture(t *testing.T, dir string) (root, path string, before []byte) {
	t.Helper()
	root = t.TempDir()
	name := "000000000002-repair.md"
	if dir == domain.EpicsDir {
		name = "01-repair.md"
	}
	path = filepath.Join(root, dir, name)
	before = []byte("---\nid: 000000000002\ndescription: Phase 1: do it # keep comment\ntags: one,two\ncustom: preserve\n---\n# Repair\n\nKeep body.\n")
	testutil.Write(t, path, string(before))
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	return
}

func TestFixInPlaceRefusesRawDriftAcrossEntityKinds(t *testing.T) {
	for _, dir := range []string{domain.TasksDir, domain.EpicsDir, domain.AuditsDir, domain.ResearchDir, domain.ThreadsDir} {
		for _, drift := range []string{"edit", "delete", "symlink", "directory"} {
			t.Run(dir+"/"+drift, func(t *testing.T) {
				root, path, before := inplaceRepairFixture(t, dir)
				winner := append(bytes.Clone(before), []byte("\nUnique raw edit.\n")...)
				foreign := filepath.Join(t.TempDir(), "foreign.md")
				testutil.Write(t, foreign, string(before)) // byte-identical links still are not records
				original := testHookBeforeFixInPlaceWrite
				t.Cleanup(func() { testHookBeforeFixInPlaceWrite = original })
				calls := 0
				testHookBeforeFixInPlaceWrite = func(target string) {
					calls++
					if target != path {
						t.Fatalf("unexpected target: %s", target)
					}
					if drift == "edit" {
						if err := os.WriteFile(path, winner, 0o600); err != nil {
							t.Fatal(err)
						}
						return
					}
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					switch drift {
					case "symlink":
						if err := os.Symlink(foreign, path); err != nil {
							t.Fatal(err)
						}
					case "directory":
						if err := os.Mkdir(path, 0o700); err != nil {
							t.Fatal(err)
						}
					}
				}
				results, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).FixFrontmatter(false)
				if !errors.Is(err, domain.ErrConflict) || len(results) != 0 || calls != 1 {
					t.Fatalf("stale repair accepted/replayed: results=%+v err=%v calls=%d", results, err, calls)
				}
				switch drift {
				case "edit":
					if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, winner) {
						t.Fatalf("raw edit lost: %q err=%v", got, err)
					}
					if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
						t.Fatalf("raw mode lost: info=%v err=%v", info, err)
					}
				case "delete":
					if _, err := os.Lstat(path); !os.IsNotExist(err) {
						t.Fatalf("deleted source resurrected: %v", err)
					}
				case "symlink":
					if target, err := os.Readlink(path); err != nil || target != foreign {
						t.Fatalf("foreign entry overwritten: %s err=%v", target, err)
					}
				case "directory":
					if info, err := os.Lstat(path); err != nil || !info.IsDir() {
						t.Fatalf("foreign directory overwritten: info=%v err=%v", info, err)
					}
				}
				if got, err := os.ReadFile(foreign); err != nil || !bytes.Equal(got, before) {
					t.Fatalf("foreign content changed: %q err=%v", got, err)
				}
			})
		}
	}
}

func TestFixInPlaceRechecksAuthorizationAfterWaiting(t *testing.T) {
	root, path, before := inplaceRepairFixture(t, domain.TasksDir)
	var denied atomic.Bool
	var calls atomic.Int32
	entered := make(chan struct{})
	policyErr := errors.New("policy revoked while waiting")
	fs := testutil.Must(NewFS(root, core.GuardedMutations(func() error {
		if denied.Load() {
			return policyErr
		}
		if calls.Add(1) == 2 { // entry + lock-entry checks, both before waiting
			close(entered)
		}
		return nil
	})))
	release := processRepositoryLock(root)
	released := false
	defer func() {
		if !released {
			release()
		}
	}()
	done := make(chan error, 1)
	go func() {
		_, err := fs.FixFrontmatter(false)
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(10 * time.Second):
		t.Fatal("repair did not reach the guard")
	}
	denied.Store(true)
	release()
	released = true
	select {
	case err := <-done:
		if !errors.Is(err, policyErr) {
			t.Fatalf("post-wait denial ignored: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("repair did not release the guard")
	}
	if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, before) {
		t.Fatalf("denied repair changed source: %q err=%v", got, err)
	}
}

func TestFixInPlaceRetainsCompletedPrefixAndReleaseFailure(t *testing.T) {
	root, later, before := inplaceRepairFixture(t, domain.TasksDir)
	earlier := filepath.Join(root, domain.TasksDir, "000000000001-earlier.md")
	testutil.Write(t, earlier, strings.Replace(string(before), "000000000002", "000000000001", 1))
	winner := string(before) + "\nConcurrent edit.\n"
	original := testHookBeforeFixInPlaceWrite
	t.Cleanup(func() { testHookBeforeFixInPlaceWrite = original })
	testHookBeforeFixInPlaceWrite = func(path string) {
		if path == later {
			testutil.Write(t, path, winner)
		}
	}
	releaseErr := errors.New("guard finalization failed")
	unlockHook := testHookRepositoryUnlockError
	t.Cleanup(func() { testHookRepositoryUnlockError = unlockHook })
	testHookRepositoryUnlockError = func() error { return releaseErr }
	results, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).FixFrontmatter(false)
	if !errors.Is(err, domain.ErrConflict) || !errors.Is(err, releaseErr) || len(results) != 1 || results[0].Path != earlier || results[0].Skipped {
		t.Fatalf("lost primary error, cleanup error or completed prefix: %+v err=%v", results, err)
	}
	if got, err := os.ReadFile(later); err != nil || string(got) != winner {
		t.Fatalf("stale write overwrote newer bytes: %q err=%v", got, err)
	}
	if got, err := os.ReadFile(earlier); err != nil || !strings.Contains(string(got), "tags: [one, two]") {
		t.Fatalf("durable prefix missing: %q err=%v", got, err)
	}
}

func TestFixInPlacePreservesPreviewAndRepairSemantics(t *testing.T) {
	for _, dir := range []string{domain.TasksDir, domain.EpicsDir, domain.AuditsDir, domain.ResearchDir, domain.ThreadsDir} {
		t.Run(dir, func(t *testing.T) {
			root, path, before := inplaceRepairFixture(t, dir)
			fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
			preview, err := fs.FixFrontmatter(true)
			if err != nil || len(preview) != 1 || preview[0].Skipped {
				t.Fatalf("malformed-text preview failed: %+v err=%v", preview, err)
			}
			if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, before) {
				t.Fatalf("preview changed bytes: %q err=%v", got, err)
			}
			results, err := fs.FixFrontmatter(false)
			if err != nil || len(results) != 1 || results[0].Skipped {
				t.Fatalf("repair failed: %+v err=%v", results, err)
			}
			want := strings.Replace(string(before), "description: Phase 1: do it", `description: "Phase 1: do it"`, 1)
			want = strings.Replace(want, "tags: one,two", "tags: [one, two]", 1)
			if got, err := os.ReadFile(path); err != nil || string(got) != want {
				t.Fatalf("repair changed unrelated content: %q err=%v", got, err)
			}
			if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
				t.Fatalf("repair changed restrictive mode: info=%v err=%v", info, err)
			}
			if results, err := fs.FixFrontmatter(false); err != nil || len(results) != 0 {
				t.Fatalf("repair is not idempotent: %+v err=%v", results, err)
			}
		})
	}
}

func TestFixFrontmatterReportsGuardReleaseFailureEvenWithoutWrites(t *testing.T) {
	for _, scenario := range []string{"repaired", "no-op", "skipped", "preview"} {
		t.Run(scenario, func(t *testing.T) {
			root, path, before := inplaceRepairFixture(t, domain.TasksDir)
			switch scenario {
			case "no-op":
				before, _ = fixFrontmatterText(before)
				testutil.Write(t, path, string(before))
			case "skipped":
				before = bytes.Replace(before, []byte("tags: one,two"), []byte("depends_on: one,two"), 1)
				testutil.Write(t, path, string(before))
			}
			releaseErr := errors.New("release failed")
			original := testHookRepositoryUnlockError
			t.Cleanup(func() { testHookRepositoryUnlockError = original })
			calls := 0
			testHookRepositoryUnlockError = func() error { calls++; return releaseErr }
			results, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).FixFrontmatter(scenario == "preview")
			if scenario == "preview" {
				if err != nil || calls != 0 {
					t.Fatalf("preview acquired/released write guard: %+v err=%v calls=%d", results, err, calls)
				}
			} else if !errors.Is(err, releaseErr) || calls != 1 {
				t.Fatalf("release failure lost: %+v err=%v calls=%d", results, err, calls)
			}
			wantRows := 1
			if scenario == "no-op" {
				wantRows = 0
			}
			if len(results) != wantRows || (wantRows > 0 && results[0].Skipped != (scenario == "skipped")) {
				t.Fatalf("release changed completed/proposed/skipped rows: %+v", results)
			}
			if scenario != "repaired" {
				if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, before) {
					t.Fatalf("non-writing repair changed bytes: %q err=%v", got, err)
				}
			}
		})
	}
}

func TestFixFrontmatterReleaseFailurePreservesPartialRelocation(t *testing.T) {
	root, source, destination, _ := invalidIDRepairFixture(t)
	unlinkErr, releaseErr := errors.New("unlink failed"), errors.New("release failed")
	removeHook, unlockHook := testHookFixIDRemoveSource, testHookRepositoryUnlockError
	t.Cleanup(func() {
		testHookFixIDRemoveSource, testHookRepositoryUnlockError = removeHook, unlockHook
	})
	testHookFixIDRemoveSource = func(string) error { return unlinkErr }
	testHookRepositoryUnlockError = func() error { return releaseErr }
	result, err := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations()))).RepairPlanning(false)
	var failure *core.FrontmatterRepairFailure
	if !errors.Is(err, unlinkErr) || !errors.Is(err, releaseErr) || !errors.As(err, &failure) || len(result.Fixes) != 0 || result.Relocation == nil || result.Relocation.SourcePath != source || result.Relocation.DestinationPath != destination {
		t.Fatalf("joined failure lost relocation distinction: %+v err=%v", result, err)
	}
}

func TestVerifyFixInPlacePathRetainsFilesystemCause(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "not-a-directory")
	testutil.Write(t, parent, "keep")
	path := filepath.Join(parent, "repair.md")
	err := verifyFixInPlacePath(path, "irrelevant")
	var pathErr *os.PathError
	if !errors.Is(err, syscall.ENOTDIR) || !errors.As(err, &pathErr) || pathErr.Path != path {
		t.Fatalf("source inspection lost actual filesystem cause/path: %v", err)
	}
}
