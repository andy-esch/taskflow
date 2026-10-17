package store

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestRenameTask_PreexistingDanglingDestinationRejectsBeforeCascade(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		t.Run(map[bool]string{false: "write", true: "preview"}[dryRun], func(t *testing.T) {
			root, source, inbound := renameRepo(t)
			destination := filepath.Join(root, domain.TasksDir, "6fjangd7kva1-new-title.md")
			missing := filepath.Join(root, "missing-target")
			if err := os.Symlink(missing, destination); err != nil {
				t.Fatal(err)
			}
			before := map[string][]byte{}
			for _, path := range []string{source, inbound} {
				before[path] = testutil.Must(os.ReadFile(path))
			}
			result, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).RenameTask("old", "New title", dryRun)
			if !errors.Is(err, domain.ErrConflict) || result.Committed || result.DestinationWritten || result.SourceRemoved || result.Complete || result.AppliedDocuments != 0 || result.AppliedLinks != 0 {
				t.Fatalf("known occupied target must refuse before any writes: %+v err=%v", result, err)
			}
			for path, content := range before {
				if after, err := os.ReadFile(path); err != nil || !bytes.Equal(after, content) {
					t.Fatalf("preflight changed %s: %q err=%v", path, after, err)
				}
			}
			if target, err := os.Readlink(destination); err != nil || target != missing {
				t.Fatalf("preflight changed occupied entry: target=%s err=%v", target, err)
			}
			if _, err := os.Lstat(missing); !os.IsNotExist(err) {
				t.Fatalf("preflight created symlink target: %v", err)
			}
		})
	}
}

func TestRenameTask_DestinationDriftRetainsSourceAndPartialReceipt(t *testing.T) {
	for _, cascade := range []bool{false, true} {
		for _, drift := range []string{"deleted", "edited", "directory", "symlink"} {
			t.Run(map[bool]string{false: "no-cascade/", true: "cascade/"}[cascade]+drift, func(t *testing.T) {
				root, oldPath, inboundPath := renameRepo(t)
				newPath := filepath.Join(root, domain.TasksDir, "6fjangd7kva1-new-title.md")
				if !cascade {
					if err := os.Remove(inboundPath); err != nil {
						t.Fatal(err)
					}
				}
				before, err := os.ReadFile(oldPath)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(oldPath, 0o600); err != nil {
					t.Fatal(err)
				}
				calls := 0
				var edited []byte
				var linkTarget string
				original := testHookBeforeTaskRenameSourceRemove
				t.Cleanup(func() { testHookBeforeTaskRenameSourceRemove = original })
				testHookBeforeTaskRenameSourceRemove = func(source, destination string) error {
					calls++
					if source != oldPath || destination != newPath {
						t.Fatalf("unexpected relocation: %s -> %s", source, destination)
					}
					if drift == "edited" {
						content, err := os.ReadFile(destination)
						if err != nil {
							return err
						}
						edited = append(content, []byte("\nUnique raw destination edit.\n")...)
						return os.WriteFile(destination, edited, 0o600)
					}
					if drift == "symlink" {
						edited, err = os.ReadFile(destination)
						if err != nil {
							return err
						}
						linkTarget = filepath.Join(t.TempDir(), "renamed-copy.md")
						if err := os.WriteFile(linkTarget, edited, 0o600); err != nil {
							return err
						}
					}
					if err := os.Remove(destination); err != nil {
						return err
					}
					if drift == "directory" {
						return os.Mkdir(destination, 0o700)
					}
					if drift == "symlink" {
						return os.Symlink(linkTarget, destination)
					}
					return nil
				}

				receipt, err := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations()))).RenameTask("old", "New title", false)
				var partial *core.TaskRenameFailure
				if !errors.As(err, &partial) || !receipt.Committed || !receipt.DestinationWritten || receipt.SourceRemoved || receipt.Complete || calls != 1 {
					t.Fatalf("destination drift lost source protection/progress or replayed rename: %+v err=%v calls=%d", receipt, err, calls)
				}
				if !errors.Is(err, domain.ErrConflict) {
					t.Fatalf("destination drift should be a conflict: %v", err)
				}
				wantDocuments, wantLinks := 1, 0
				if cascade {
					wantDocuments, wantLinks = 2, 1
					if content, err := os.ReadFile(inboundPath); err != nil || !bytes.Contains(content, []byte(filepath.Base(newPath))) {
						t.Fatalf("durable cascade prefix lost: %q err=%v", content, err)
					}
				}
				if receipt.AppliedDocuments != wantDocuments || receipt.PlannedDocuments != wantDocuments || receipt.AppliedLinks != wantLinks || receipt.PlannedLinks != wantLinks || receipt.SourcePath != oldPath || receipt.DestinationPath != newPath {
					t.Fatalf("incorrect historical receipt: %+v", receipt)
				}
				if !reflect.DeepEqual(partial.Receipt, receipt) || !strings.Contains(receipt.Remedy, "old and destination") || !strings.Contains(receipt.Remedy, "before retrying") {
					t.Fatalf("partial error lost receipt or inspection guidance: %+v err=%v", receipt, err)
				}
				if content, err := os.ReadFile(oldPath); err != nil || !bytes.Equal(content, before) {
					t.Fatalf("destination drift deleted/changed intact source: %q err=%v", content, err)
				}
				if info, err := os.Stat(oldPath); err != nil || info.Mode().Perm() != 0o600 {
					t.Fatalf("destination drift changed source mode: info=%v err=%v", info, err)
				}
				switch drift {
				case "deleted":
					if _, err := os.Stat(newPath); !os.IsNotExist(err) {
						t.Fatalf("raw deletion was undone: %v", err)
					}
				case "edited":
					if content, err := os.ReadFile(newPath); err != nil || !bytes.Equal(content, edited) {
						t.Fatalf("raw destination edit was clobbered: %q err=%v", content, err)
					}
				case "directory":
					if info, err := os.Stat(newPath); err != nil || !info.IsDir() {
						t.Fatalf("raw destination replacement was changed: info=%v err=%v", info, err)
					}
				case "symlink":
					if target, err := os.Readlink(newPath); err != nil || target != linkTarget {
						t.Fatalf("raw symlink was changed: target=%s err=%v", target, err)
					}
					if content, err := os.ReadFile(linkTarget); err != nil || !bytes.Equal(content, edited) {
						t.Fatalf("foreign symlink target was changed: %q err=%v", content, err)
					}
				}
			})
		}
	}
}

func TestVerifyTaskRenameDestinationPreservesFilesystemCause(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parent, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(parent, "task.md")
	err := verifyTaskRenameDestination(path, hashContent([]byte("expected")))
	var pathErr *os.PathError
	if !errors.Is(err, syscall.ENOTDIR) || !errors.As(err, &pathErr) || pathErr.Path != path {
		t.Fatalf("verification discarded filesystem cause: %v", err)
	}
}
