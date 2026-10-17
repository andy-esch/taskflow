package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func invalidIDRepairFixture(t *testing.T) (root, source, destination string, before []byte) {
	t.Helper()
	root = t.TempDir()
	source = filepath.Join(root, domain.TasksDir, "6fbj87000lt6-bad-id.md")
	destination = filepath.Join(root, domain.TasksDir, "6fbj870001t6-bad-id.md")
	before = []byte("---\nschema: 1\nid: 6fbj87000lt6\nstatus: next-up\ndescription: keep # comment\ncustom: value\n---\n# Bad ID\n\nKeep body.\n")
	testutil.Write(t, source, string(before))
	return
}

func TestFixIDRelocationRefusesRawDestinationCollision(t *testing.T) {
	root, source, destination, before := invalidIDRepairFixture(t)
	winner := "another writer owns this destination\n"
	original := testHookBeforeFixIDExclusiveCreate
	t.Cleanup(func() { testHookBeforeFixIDExclusiveCreate = original })
	testHookBeforeFixIDExclusiveCreate = func(path string) { testutil.Write(t, path, winner) }
	results, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).FixFrontmatter(false)
	if !errors.Is(err, domain.ErrConflict) || len(results) != 0 {
		t.Errorf("destination collision result=%+v err=%v", results, err)
	}
	if got, err := os.ReadFile(destination); err != nil || string(got) != winner {
		t.Errorf("destination winner was overwritten: %q err=%v", got, err)
	}
	if got, err := os.ReadFile(source); err != nil || string(got) != string(before) {
		t.Errorf("collision changed the source: %q err=%v", got, err)
	}
}

func TestFixIDRelocationRefusesRawSourceChangeBeforeCreate(t *testing.T) {
	root, source, destination, before := invalidIDRepairFixture(t)
	concurrent := string(before) + "\nConcurrent edit before create.\n"
	original := testHookBeforeFixIDDestinationCreate
	t.Cleanup(func() { testHookBeforeFixIDDestinationCreate = original })
	testHookBeforeFixIDDestinationCreate = func(path, _ string) { testutil.Write(t, path, concurrent) }
	results, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).FixFrontmatter(false)
	if !errors.Is(err, domain.ErrConflict) || len(results) != 0 {
		t.Errorf("source conflict result=%+v err=%v", results, err)
	}
	if got, err := os.ReadFile(source); err != nil || string(got) != concurrent {
		t.Errorf("concurrent source was removed/changed: %q err=%v", got, err)
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stale destination was created: %v", err)
	}
}

func TestFixIDRelocationRechecksSpellingAuthorizationBeforeCreate(t *testing.T) {
	for _, drift := range []string{"canonical owner", "cross-kind owner", "inbound reference"} {
		t.Run(drift, func(t *testing.T) {
			root, source, destination, before := invalidIDRepairFixture(t)
			var peerPath, peerBody string
			original := testHookBeforeFixIDDestinationCreate
			t.Cleanup(func() { testHookBeforeFixIDDestinationCreate = original })
			testHookBeforeFixIDDestinationCreate = func(_, _ string) {
				switch drift {
				case "canonical owner":
					peerPath = filepath.Join(root, domain.TasksDir, "6fbj870001t6-other.md")
					peerBody = "---\nid: 6fbj870001t6\n---\n# Existing identity\n"
				case "cross-kind owner":
					peerPath = filepath.Join(root, domain.ThreadsDir, "6fbj870001t6-other.md")
					peerBody = "---\nid: 6fbj870001t6\n---\n# Existing Thread identity\n"
				case "inbound reference":
					peerPath = filepath.Join(root, domain.TasksDir, "000000000001-reference.md")
					peerBody = "---\nid: 000000000001\n---\nSee 6fbj87000lt6-bad-id.md\n"
				}
				testutil.Write(t, peerPath, peerBody)
			}
			results, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).FixFrontmatter(false)
			if !errors.Is(err, domain.ErrConflict) || len(results) != 0 {
				t.Fatalf("stale authorization results=%+v err=%v", results, err)
			}
			if content, err := os.ReadFile(source); err != nil || string(content) != string(before) {
				t.Fatalf("refused spelling repair changed source: %s err=%v", content, err)
			}
			if content, err := os.ReadFile(peerPath); err != nil || string(content) != peerBody {
				t.Fatalf("refused spelling repair changed peer: %s err=%v", content, err)
			}
			if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("refused spelling repair created destination: %v", err)
			}
		})
	}
}

func TestFixIDRelocationKeepsRawSourceChangeBeforeRemove(t *testing.T) {
	root, source, destination, before := invalidIDRepairFixture(t)
	concurrent := string(before) + "\nConcurrent edit before unlink.\n"
	original := testHookBeforeFixIDSourceRemove
	t.Cleanup(func() { testHookBeforeFixIDSourceRemove = original })
	testHookBeforeFixIDSourceRemove = func(path, _ string) error {
		testutil.Write(t, path, concurrent)
		return nil
	}
	results, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).FixFrontmatter(false)
	if !errors.Is(err, domain.ErrConflict) || len(results) != 0 {
		t.Errorf("post-create conflict result=%+v err=%v", results, err)
	}
	assertPartialFixRelocation(t, err, source, destination)
	if got, err := os.ReadFile(source); err != nil || string(got) != concurrent {
		t.Errorf("concurrent source was deleted/changed: %q err=%v", got, err)
	}
	if _, err := os.Stat(destination); err != nil {
		t.Errorf("durable destination lost instead of preserved for inspection: %v", err)
	}
}

func assertPartialFixRelocation(t *testing.T, err error, source, destination string) {
	t.Helper()
	var failure *core.FrontmatterRepairFailure
	if !errors.As(err, &failure) {
		t.Fatalf("missing typed partial relocation: %v", err)
	}
	if failure.Relocation.SourcePath != source || failure.Relocation.DestinationPath != destination || !failure.Relocation.DestinationWritten || failure.Relocation.SourceRemoved {
		t.Fatalf("incorrect relocation progress: %+v", failure.Relocation)
	}
	for _, expected := range []string{source, destination, "compare and merge both copies", "refuses while both paths exist"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("partial error omits %q: %v", expected, err)
		}
	}
}

func TestFixIDRelocationRemovalFailureRetainsEarlierRepairsAndLocalOutcome(t *testing.T) {
	root, source, destination, before := invalidIDRepairFixture(t)
	earlier := filepath.Join(root, domain.TasksDir, "000000000001-earlier.md")
	testutil.Write(t, earlier, "---\nid: 000000000001\nstatus: next-up\ntags: a,b\n---\n# Earlier\n")
	original := testHookFixIDRemoveSource
	t.Cleanup(func() { testHookFixIDRemoveSource = original })
	testHookFixIDRemoveSource = func(path string) error {
		return &os.PathError{Op: "remove", Path: path, Err: syscall.EPERM}
	}
	result, err := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations()))).RepairPlanning(false)
	if !errors.Is(err, syscall.EPERM) || len(result.Fixes) != 1 || result.Fixes[0].Path != earlier || result.Fixes[0].Skipped || result.Relocation == nil {
		t.Fatalf("earlier prefix/local progress lost or partial move counted fixed: result=%+v err=%v", result, err)
	}
	assertPartialFixRelocation(t, err, source, destination)
	if result.Relocation.SourcePath != source || result.Relocation.DestinationPath != destination || !result.Relocation.DestinationWritten || result.Relocation.SourceRemoved {
		t.Fatalf("application lost the real adapter outcome: %+v", result.Relocation)
	}
	if content, err := os.ReadFile(source); err != nil || string(content) != string(before) {
		t.Fatalf("failed unlink changed source: %s err=%v", content, err)
	}
	content, err := os.ReadFile(destination)
	want := strings.Replace(string(before), "id: 6fbj87000lt6", "id: 6fbj870001t6", 1)
	if err != nil || string(content) != want {
		t.Fatalf("destination changed unrelated bytes: %s err=%v", content, err)
	}
}

func TestFixIDRelocationRetainsSourceWhenCreatedDestinationDrifts(t *testing.T) {
	for _, deleted := range []bool{false, true} {
		t.Run(map[bool]string{false: "edited", true: "deleted"}[deleted], func(t *testing.T) {
			root, source, destination, before := invalidIDRepairFixture(t)
			original := testHookBeforeFixIDSourceRemove
			t.Cleanup(func() { testHookBeforeFixIDSourceRemove = original })
			testHookBeforeFixIDSourceRemove = func(_, path string) error {
				if deleted {
					return os.Remove(path)
				}
				testutil.Write(t, path, "concurrent destination bytes\n")
				return nil
			}
			results, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).FixFrontmatter(false)
			if !errors.Is(err, domain.ErrConflict) || len(results) != 0 {
				t.Fatalf("destination drift results=%+v err=%v", results, err)
			}
			assertPartialFixRelocation(t, err, source, destination)
			if content, err := os.ReadFile(source); err != nil || string(content) != string(before) {
				t.Fatalf("old copy was lost after destination drift: %s err=%v", content, err)
			}
			if deleted {
				if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("deleted destination was recreated: %v", err)
				}
			} else if content, err := os.ReadFile(destination); err != nil || string(content) != "concurrent destination bytes\n" {
				t.Fatalf("raw destination edit was clobbered: %s err=%v", content, err)
			}
		})
	}
}

func TestLintAfterPartialIDRelocationNeverAdvisesOverwritingTheSibling(t *testing.T) {
	for _, dir := range []string{domain.TasksDir, domain.AuditsDir, domain.ResearchDir, domain.ThreadsDir} {
		for _, edited := range []bool{false, true} {
			t.Run(dir+map[bool]string{false: "/unchanged", true: "/edited"}[edited], func(t *testing.T) {
				root := t.TempDir()
				source := filepath.Join(root, dir, "6fbj87000lt6-bad-id.md")
				destination := filepath.Join(root, dir, "6fbj870001t6-bad-id.md")
				before := "---\nschema: 1\nid: 6fbj87000lt6\nstatus: next-up\n---\n# Keep\n"
				testutil.Write(t, source, before)
				original := testHookFixIDRemoveSource
				t.Cleanup(func() { testHookFixIDRemoveSource = original })
				testHookFixIDRemoveSource = func(path string) error {
					return &os.PathError{Op: "remove", Path: path, Err: syscall.EBUSY}
				}
				fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
				_, err := fs.FixFrontmatter(false)
				if !errors.Is(err, syscall.EBUSY) {
					t.Fatalf("fixture did not reach partial relocation: %v", err)
				}
				testHookFixIDRemoveSource = nil
				if edited {
					testutil.Write(t, destination, strings.Replace(before, "6fbj87000lt6", "6fbj870001t6", 1)+"\nUnique destination edit.\n")
				}
				destinationBefore, err := os.ReadFile(destination)
				if err != nil {
					t.Fatal(err)
				}
				_, problems, err := core.MustNewService(fs).Lint()
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for _, problem := range problems {
					if problem.Location != source {
						continue
					}
					found = true
					for _, want := range []string{"already exists", filepath.Base(destination), "compare and merge both copies", "refuses while both paths exist"} {
						if !strings.Contains(problem.Message, want) {
							t.Errorf("recovery diagnostic omits %q: %s", want, problem.Message)
						}
					}
					if strings.Contains(problem.Message, "rename to") {
						t.Errorf("recovery advises overwriting the occupied sibling: %s", problem.Message)
					}
				}
				if !found {
					t.Fatalf("plain lint lost retained source: %+v", problems)
				}
				results, err := fs.FixFrontmatter(false)
				if err != nil || len(results) != 1 || !results[0].Skipped || !strings.Contains(strings.Join(results[0].Changes, " "), "already exists") {
					t.Fatalf("retry should refuse reconciliation: results=%+v err=%v", results, err)
				}
				if content, err := os.ReadFile(source); err != nil || string(content) != before {
					t.Fatalf("lint/retry changed source: %q err=%v", content, err)
				}
				if content, err := os.ReadFile(destination); err != nil || string(content) != string(destinationBefore) {
					t.Fatalf("lint/retry changed destination: %q err=%v", content, err)
				}
			})
		}
	}
}

func TestFixIDRelocationDryRunWritesNothingAndPreservesModeForAllIDLedKinds(t *testing.T) {
	for _, dir := range []string{domain.TasksDir, domain.AuditsDir, domain.ResearchDir, domain.ThreadsDir} {
		t.Run(dir, func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, dir, "6fbj87000lt6-bad-id.md")
			destination := filepath.Join(root, dir, "6fbj870001t6-bad-id.md")
			before := "---\nid: 6fbj87000lt6\ncustom: keep # comment\n---\n# Keep\n"
			testutil.Write(t, source, before)
			if err := os.Chmod(source, 0o600); err != nil {
				t.Fatal(err)
			}
			original := testHookBeforeFixIDDestinationCreate
			t.Cleanup(func() { testHookBeforeFixIDDestinationCreate = original })
			testHookBeforeFixIDDestinationCreate = func(_, _ string) { t.Fatal("preview reached persistence") }
			fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
			proposed, err := fs.FixFrontmatter(true)
			if err != nil || len(proposed) != 1 || proposed[0].Skipped {
				t.Fatalf("preview=%+v err=%v", proposed, err)
			}
			if content, err := os.ReadFile(source); err != nil || string(content) != before {
				t.Fatalf("preview changed source: %s err=%v", content, err)
			}
			if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("preview created destination: %v", err)
			}
			testHookBeforeFixIDDestinationCreate = nil
			fixed, err := fs.FixFrontmatter(false)
			if err != nil || len(fixed) != 1 || fixed[0].Skipped {
				t.Fatalf("fixed=%+v err=%v", fixed, err)
			}
			if info, err := os.Stat(destination); err != nil || info.Mode().Perm() != 0o600 {
				t.Fatalf("source mode changed: info=%v err=%v", info, err)
			}
		})
	}
}
