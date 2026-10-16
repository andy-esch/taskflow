package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestRepairTaskGraphRetryNeverReselectsReusedSlug(t *testing.T) {
	root := t.TempDir()
	selected := writeGraphMutationTask(t, root, "alpha", domain.StatusReadyToStart, []string{"invalid-token"}, "")
	other := writeGraphMutationTask(t, root, "bravo", domain.StatusReadyToStart, []string{"invalid-token"}, "")
	renamed := filepath.Join(filepath.Dir(selected), testutil.TaskID("alpha")+"-renamed.md")
	replacement := filepath.Join(filepath.Dir(other), testutil.TaskID("bravo")+"-alpha.md")
	otherBefore, err := os.ReadFile(other)
	if err != nil {
		t.Fatal(err)
	}
	original := testHookBeforeGraphRepairVerify
	t.Cleanup(func() { testHookBeforeGraphRepairVerify = original })
	testHookBeforeGraphRepairVerify = func() {
		testHookBeforeGraphRepairVerify = nil
		if err := os.Rename(selected, renamed); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(other, replacement); err != nil {
			t.Fatal(err)
		}
	}
	svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(1, func(int) {}))
	receipt, repairErr := svc.RepairTaskGraph(core.TaskGraphRepairRequest{Edits: []core.TaskGraphSourceEdit{{
		Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskSlug: "alpha"},
		Field: core.TaskDependencyDependsOn, Value: "invalid-token",
	}}}, false)
	otherAfter, err := os.ReadFile(replacement)
	if err != nil {
		t.Fatal(err)
	}
	if string(otherAfter) != string(otherBefore) {
		t.Fatalf("retry repaired the reused slug's different source: applied=%v err=%v", receipt.AppliedFiles, repairErr)
	}
	if !errors.Is(repairErr, domain.ErrNotFound) || receipt.Changed || receipt.Committed || len(receipt.AppliedFiles) != 0 {
		t.Fatalf("moved selection was not refused without writes: receipt=%+v err=%v", receipt, repairErr)
	}
	selectedAfter, err := os.ReadFile(renamed)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(selectedAfter), "invalid-token") {
		t.Fatal("refused repair changed the moved source")
	}
}

func TestRepairTaskGraphRetryRefusesMovedMissingAndUnreadableSources(t *testing.T) {
	for _, selector := range []string{"ID", "slug", "path"} {
		for _, drift := range []string{"renamed", "deleted", "unreadable"} {
			t.Run(selector+"/"+drift, func(t *testing.T) {
				root := t.TempDir()
				path := writeGraphMutationTask(t, root, "pinned-repair", domain.StatusNextUp, []string{"invalid-token"}, "")
				source := core.TaskGraphSourceRef{}
				switch selector {
				case "ID":
					source.TaskID = testutil.TaskID("pinned-repair")
				case "slug":
					source.TaskSlug = "pinned-repair"
				case "path":
					source.LocalPath = path
				}
				original := testHookBeforeGraphRepairVerify
				t.Cleanup(func() { testHookBeforeGraphRepairVerify = original })
				var afterDrift []byte
				remainingPath := path
				testHookBeforeGraphRepairVerify = func() {
					testHookBeforeGraphRepairVerify = nil
					var err error
					switch drift {
					case "renamed":
						remainingPath = filepath.Join(filepath.Dir(path), testutil.TaskID("pinned-repair")+"-new-name.md")
						err = os.Rename(path, remainingPath)
					case "deleted":
						err = os.Remove(path)
					case "unreadable":
						testutil.Write(t, path, "---\nid: [unterminated\n---\n# Raw repair in progress\n")
					}
					if err != nil {
						t.Fatal(err)
					}
					if drift != "deleted" {
						afterDrift, err = os.ReadFile(remainingPath)
						if err != nil {
							t.Fatal(err)
						}
					}
				}
				svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(2, func(int) {}))
				receipt, err := svc.RepairTaskGraph(core.TaskGraphRepairRequest{Edits: []core.TaskGraphSourceEdit{{
					Action: core.TaskGraphSourceDropDeclaration, Source: source, Field: core.TaskDependencyDependsOn, Value: "invalid-token",
				}}}, false)
				if !errors.Is(err, domain.ErrNotFound) || receipt.Committed || receipt.Changed || len(receipt.AppliedFiles) != 0 {
					t.Fatalf("receipt=%+v err=%v", receipt, err)
				}
				if drift == "unreadable" && !strings.Contains(err.Error(), "became unreadable") {
					t.Fatalf("refusal omits unreadable source advice: %v", err)
				}
				if drift == "deleted" {
					if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("deleted source was resurrected: %v", err)
					}
				} else if content, err := os.ReadFile(remainingPath); err != nil || string(content) != string(afterDrift) {
					t.Fatalf("refusal changed concurrent source bytes: %v", err)
				}
			})
		}
	}
}

func TestRepairTaskGraphRetryKeepsDuplicateIDOccurrenceAndConcurrentContent(t *testing.T) {
	for _, duplicateInitially := range []bool{false, true} {
		t.Run(fmt.Sprintf("duplicate-initially=%t", duplicateInitially), func(t *testing.T) {
			root := t.TempDir()
			path := writeGraphMutationTask(t, root, "repair-duplicate-source", domain.StatusNextUp, []string{"invalid-token"}, "custom: keep # comment\n")
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			shadow := filepath.Join(filepath.Dir(path), testutil.TaskID("repair-duplicate-source")+"-shadow.md")
			if duplicateInitially {
				testutil.Write(t, shadow, string(before))
			}
			shadowBytes := string(before) + "\n<!-- concurrent shadow edit -->\n"
			original := testHookBeforeGraphRepairVerify
			t.Cleanup(func() { testHookBeforeGraphRepairVerify = original })
			testHookBeforeGraphRepairVerify = func() {
				testHookBeforeGraphRepairVerify = nil
				testutil.Write(t, shadow, shadowBytes)
				updated := strings.Replace(string(before), "depends_on: [invalid-token]", "depends_on: [invalid-token, concurrent-intent]", 1)
				testutil.Write(t, path, updated+"\n<!-- concurrent selected-source body -->\n")
			}
			source := core.TaskGraphSourceRef{TaskID: testutil.TaskID("repair-duplicate-source")}
			if duplicateInitially {
				// Duplicate IDs are deliberately disambiguated by physical occurrence.
				source = core.TaskGraphSourceRef{LocalPath: path}
			}
			svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(1, func(int) {}))
			receipt, err := svc.RepairTaskGraph(core.TaskGraphRepairRequest{Edits: []core.TaskGraphSourceEdit{{
				Action: core.TaskGraphSourceDropDeclaration, Source: source, Field: core.TaskDependencyDependsOn, Value: "invalid-token",
			}}}, false)
			if err != nil || !receipt.Committed || len(receipt.AppliedFiles) != 1 || receipt.AppliedFiles[0] != path || receipt.FinalHealth != core.GraphBroken {
				t.Fatalf("receipt=%+v err=%v", receipt, err)
			}
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, expected := range []string{"concurrent-intent", "custom: keep # comment", "concurrent selected-source body"} {
				if !strings.Contains(string(content), expected) {
					t.Fatalf("repair lost %q:\n%s", expected, content)
				}
			}
			if strings.Contains(string(content), "invalid-token") {
				t.Fatal("selected defect was not repaired")
			}
			if content, err := os.ReadFile(shadow); err != nil || string(content) != shadowBytes {
				t.Fatalf("repair changed the other duplicate occurrence: %v", err)
			}
		})
	}
}

func TestRepairTaskGraphRetryReauthorizesCurrentDeclaration(t *testing.T) {
	for _, concurrentRepair := range []bool{false, true} {
		t.Run(fmt.Sprintf("already-repaired=%t", concurrentRepair), func(t *testing.T) {
			root := t.TempDir()
			value := testutil.TaskID("repair-now-existing")
			path := writeGraphMutationTask(t, root, "repair-reauthorize", domain.StatusNextUp, []string{value}, "")
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			current := string(before)
			original := testHookBeforeGraphRepairVerify
			t.Cleanup(func() { testHookBeforeGraphRepairVerify = original })
			testHookBeforeGraphRepairVerify = func() {
				testHookBeforeGraphRepairVerify = nil
				if concurrentRepair {
					current = strings.Replace(current, "depends_on: ["+value+"]", "depends_on: []", 1) + "\n<!-- concurrent repair -->\n"
					testutil.Write(t, path, current)
				} else {
					writeGraphMutationTask(t, root, "repair-now-existing", domain.StatusCompleted, nil, "")
				}
			}
			svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(1, func(int) {}))
			receipt, err := svc.RepairTaskGraph(core.TaskGraphRepairRequest{Edits: []core.TaskGraphSourceEdit{{
				Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskSlug: "repair-reauthorize"}, Field: core.TaskDependencyDependsOn, Value: value,
			}}}, false)
			if concurrentRepair {
				if err != nil || len(receipt.Selected) != 1 || len(receipt.Operations) != 0 || receipt.FinalHealth != core.GraphHealthy {
					t.Fatalf("already-satisfied receipt=%+v err=%v", receipt, err)
				}
			} else if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("now-valid dependency removal was allowed: receipt=%+v err=%v", receipt, err)
			}
			if receipt.Changed || receipt.Committed || len(receipt.AppliedFiles) != 0 {
				t.Fatalf("replanned no-op/refusal claimed a write: %+v", receipt)
			}
			if content, err := os.ReadFile(path); err != nil || string(content) != current {
				t.Fatalf("replanned no-op/refusal changed source: %v", err)
			}
		})
	}
}

func TestRepairTaskGraphAutoRetryCanFollowNewSourceSet(t *testing.T) {
	root := t.TempDir()
	id := testutil.TaskID("auto-repair-moved")
	path := writeGraphMutationTask(t, root, "auto-repair-moved", domain.StatusNextUp, []string{id}, "")
	renamed := filepath.Join(filepath.Dir(path), id+"-renamed.md")
	original := testHookBeforeGraphRepairVerify
	t.Cleanup(func() { testHookBeforeGraphRepairVerify = original })
	testHookBeforeGraphRepairVerify = func() {
		testHookBeforeGraphRepairVerify = nil
		if err := os.Rename(path, renamed); err != nil {
			t.Fatal(err)
		}
		writeGraphMutationTask(t, root, "auto-repair-new", domain.StatusNextUp, []string{testutil.TaskID("auto-repair-new")}, "")
	}
	svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(1, func(int) {}))
	receipt, err := svc.RepairTaskGraph(core.TaskGraphRepairRequest{Auto: true}, false)
	if err != nil || !receipt.Committed || len(receipt.AppliedFiles) != 2 || receipt.FinalHealth != core.GraphHealthy {
		t.Fatalf("auto did not rediagnose renamed/new sources: receipt=%+v err=%v", receipt, err)
	}
}

func TestRepairTaskGraphMixedRetryKeepsExplicitSelectionSeparateFromAuto(t *testing.T) {
	for _, moveExplicit := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit-moved=%t", moveExplicit), func(t *testing.T) {
			root := t.TempDir()
			explicitID := testutil.TaskID("mixed-explicit")
			explicitPath := writeGraphMutationTask(t, root, "mixed-explicit", domain.StatusNextUp, []string{"invalid-token"}, "custom: keep # selected comment\n")
			explicitBefore, err := os.ReadFile(explicitPath)
			if err != nil {
				t.Fatal(err)
			}
			autoID := testutil.TaskID("mixed-auto")
			autoPath := writeGraphMutationTask(t, root, "mixed-auto", domain.StatusNextUp, []string{autoID}, "")
			autoRenamed := filepath.Join(filepath.Dir(autoPath), autoID+"-renamed.md")
			original := testHookBeforeGraphRepairVerify
			t.Cleanup(func() { testHookBeforeGraphRepairVerify = original })
			var newAutoPath string
			afterDrift := make(map[string][]byte)
			testHookBeforeGraphRepairVerify = func() {
				testHookBeforeGraphRepairVerify = nil
				if err := os.Rename(autoPath, autoRenamed); err != nil {
					t.Fatal(err)
				}
				newAutoPath = writeGraphMutationTask(t, root, "mixed-auto-new", domain.StatusNextUp, []string{testutil.TaskID("mixed-auto-new")}, "")
				if moveExplicit {
					renamed := filepath.Join(filepath.Dir(explicitPath), explicitID+"-renamed.md")
					if err := os.Rename(explicitPath, renamed); err != nil {
						t.Fatal(err)
					}
					explicitPath = renamed
				}
				testutil.Write(t, explicitPath, string(explicitBefore)+"\n<!-- concurrent selected body -->\n")
				for _, path := range []string{explicitPath, autoRenamed, newAutoPath} {
					content, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					afterDrift[path] = content
				}
			}
			svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(1, func(int) {}))
			receipt, err := svc.RepairTaskGraph(core.TaskGraphRepairRequest{Auto: true, Edits: []core.TaskGraphSourceEdit{{
				Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskSlug: "mixed-explicit"},
				Field: core.TaskDependencyDependsOn, Value: "invalid-token",
			}}}, false)
			if moveExplicit {
				if !errors.Is(err, domain.ErrNotFound) || receipt.Changed || receipt.Committed || len(receipt.AppliedFiles) != 0 {
					t.Fatalf("mixed refusal receipt=%+v err=%v", receipt, err)
				}
				for path, before := range afterDrift {
					if content, err := os.ReadFile(path); err != nil || string(content) != string(before) {
						t.Fatalf("mixed refusal still wrote %s: %v", path, err)
					}
				}
				return
			}
			if err != nil || !receipt.Committed || len(receipt.AppliedFiles) != 3 || len(receipt.Operations) != 3 || receipt.FinalHealth != core.GraphHealthy {
				t.Fatalf("mixed receipt=%+v err=%v", receipt, err)
			}
			for _, operation := range receipt.Operations {
				if operation.Edit.Source.TaskID == explicitID {
					if operation.Automatic || operation.Edit.Source.LocalPath != explicitPath {
						t.Fatalf("explicit intent expanded or moved: %+v", operation)
					}
				} else if !operation.Automatic || (operation.Edit.Source.LocalPath != autoRenamed && operation.Edit.Source.LocalPath != newAutoPath) {
					t.Fatalf("auto did not rediagnose the current sources: %+v", operation)
				}
			}
			content, err := os.ReadFile(explicitPath)
			if err != nil || strings.Contains(string(content), "invalid-token") || !strings.Contains(string(content), "custom: keep # selected comment") || !strings.Contains(string(content), "concurrent selected body") {
				t.Fatalf("mixed repair lost selected source intent: %s, %v", content, err)
			}
			if _, err := os.Stat(autoPath); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("old auto path was recreated: %v", err)
			}
		})
	}
}
