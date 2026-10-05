package store

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestDependencyMigrationPreservesBodyCommentsAndConverges(t *testing.T) {
	root := t.TempDir()
	prerequisiteID := testutil.TaskID("legacy-prerequisite")
	secondID := testutil.TaskID("legacy-second")
	dependentID := testutil.TaskID("legacy-dependent")
	prerequisitePath := writeGraphMutationTask(t, root, "legacy-prerequisite", domain.StatusCompleted, nil,
		"blocks: [legacy-dependent]\ncustom_key: keep-me # keep this comment\n")
	dependentPath := writeGraphMutationTask(t, root, "legacy-dependent", domain.StatusReadyToStart, nil,
		"blocked_by: [legacy-prerequisite]\ndependencies: ["+secondID+"]\n")
	writeGraphMutationTask(t, root, "legacy-second", domain.StatusCompleted, nil, "")

	svc := core.MustNewService(NewFS(root), core.WithClock(func() time.Time { return graphMutationNow }))
	receipt, err := svc.MigrateTaskDependencies(false)
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Changed || len(receipt.ClearedLegacyFields) != 3 || len(receipt.AppliedTaskIDs) != 2 {
		t.Fatalf("migration receipt = %+v", receipt)
	}
	for _, path := range []string{prerequisitePath, dependentPath} {
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		for _, legacy := range []string{"blocked_by:", "dependencies:", "blocks:"} {
			if strings.Contains(string(content), legacy) {
				t.Fatalf("legacy field %q remains in %s:\n%s", legacy, path, content)
			}
		}
		if !strings.Contains(string(content), "Body stays intact.") || !strings.Contains(string(content), "updated_at: \"2026-08-27\"") {
			t.Fatalf("migration did not preserve body/stamp update in %s:\n%s", path, content)
		}
	}
	prerequisiteContent, _ := os.ReadFile(prerequisitePath)
	if !strings.Contains(string(prerequisiteContent), "custom_key: keep-me # keep this comment") {
		t.Fatalf("frontmatter comment was lost:\n%s", prerequisiteContent)
	}
	dependent, _, err := NewFS(root).GetTask(dependentID)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{prerequisiteID, secondID}
	slices.Sort(want)
	if !slices.Equal(dependent.DependsOn, want) {
		t.Fatalf("depends_on = %v, want %v", dependent.DependsOn, want)
	}
	graph, err := core.LoadTaskGraph(NewFS(root))
	if err != nil || graph.Health() != core.GraphHealthy {
		t.Fatalf("post-migration graph = %v health=%s", err, graph.Health())
	}
}

func TestDependencyMigrationFailureCarriesDurablePrefixAndRerunConverges(t *testing.T) {
	root := t.TempDir()
	writeGraphMutationTask(t, root, "prefix-prerequisite", domain.StatusCompleted, nil,
		"blocks: [prefix-dependent]\n")
	writeGraphMutationTask(t, root, "prefix-dependent", domain.StatusReadyToStart, nil,
		"blocked_by: [prefix-prerequisite]\n")
	svc := core.MustNewService(NewFS(root), core.WithClock(func() time.Time { return graphMutationNow }))

	original := testHookAfterGraphWrite
	defer func() { testHookAfterGraphWrite = original }()
	testHookAfterGraphWrite = func(string) error {
		testHookAfterGraphWrite = nil
		return errors.New("injected interruption")
	}
	partial, err := svc.MigrateTaskDependencies(false)
	var failure *core.DependencyMutationFailure
	if !errors.As(err, &failure) || len(partial.AppliedTaskIDs) != 1 || len(partial.RemainingTaskIDs) != 1 {
		t.Fatalf("partial receipt=%+v failure=%+v err=%v", partial, failure, err)
	}
	if !strings.Contains(err.Error(), "durable dependency prefix") {
		t.Fatalf("human recovery guidance missing: %v", err)
	}

	completed, err := svc.MigrateTaskDependencies(false)
	if err != nil || !completed.Changed || len(completed.AppliedTaskIDs) != 1 {
		t.Fatalf("convergent rerun=%+v err=%v", completed, err)
	}
	graph, loadErr := core.LoadTaskGraph(NewFS(root))
	if loadErr != nil || graph.Health() != core.GraphHealthy {
		t.Fatalf("rerun graph health=%s err=%v", graph.Health(), loadErr)
	}
}

func TestDependencyMigrationBlocksOnlyWritesDependentBeforeClearingOwner(t *testing.T) {
	root := t.TempDir()
	ownerID := testutil.TaskID("blocks-only-owner")
	dependentID := testutil.TaskID("blocks-only-dependent")
	writeGraphMutationTask(t, root, "blocks-only-owner", domain.StatusCompleted, nil,
		"blocks: [blocks-only-dependent]\n")
	writeGraphMutationTask(t, root, "blocks-only-dependent", domain.StatusReadyToStart, nil, "")
	svc := core.MustNewService(NewFS(root), core.WithClock(func() time.Time { return graphMutationNow }))

	original := testHookAfterGraphWrite
	defer func() { testHookAfterGraphWrite = original }()
	testHookAfterGraphWrite = func(string) error {
		testHookAfterGraphWrite = nil
		return errors.New("injected interruption")
	}
	partial, err := svc.MigrateTaskDependencies(false)
	if err == nil || !slices.Equal(partial.PlannedTaskIDs, []string{dependentID, ownerID}) ||
		!slices.Equal(partial.AppliedTaskIDs, []string{dependentID}) {
		t.Fatalf("blocks-only prefix receipt=%+v err=%v", partial, err)
	}
	dependent, _, getErr := NewFS(root).GetTask(dependentID)
	if getErr != nil || !slices.Equal(dependent.DependsOn, []string{ownerID}) {
		t.Fatalf("dependent canonical prefix=%v err=%v", dependent.DependsOn, getErr)
	}
	owner, _, getErr := NewFS(root).GetTask(ownerID)
	if getErr != nil || !slices.Equal(owner.LegacyBlocks, []string{"blocks-only-dependent"}) {
		t.Fatalf("owner legacy prefix=%v err=%v", owner.LegacyBlocks, getErr)
	}
	graph, loadErr := core.LoadTaskGraph(NewFS(root))
	if loadErr != nil || graph.Health() != core.GraphDegraded {
		t.Fatalf("blocks-only prefix health=%s err=%v problems=%+v", graph.Health(), loadErr, graph.Problems())
	}
	completed, err := svc.MigrateTaskDependencies(false)
	if err != nil || !slices.Equal(completed.AppliedTaskIDs, []string{ownerID}) {
		t.Fatalf("blocks-only rerun=%+v err=%v", completed, err)
	}
	graph, loadErr = core.LoadTaskGraph(NewFS(root))
	if loadErr != nil || graph.Health() != core.GraphHealthy {
		t.Fatalf("blocks-only final health=%s err=%v", graph.Health(), loadErr)
	}
}

func TestDependencyMigrationClearsAndReportsPresentEmptyLegacyFields(t *testing.T) {
	root := t.TempDir()
	taskID := testutil.TaskID("empty-legacy-owner")
	path := writeGraphMutationTask(t, root, "empty-legacy-owner", domain.StatusReadyToStart, nil,
		"blocked_by: []\ndependencies: []\nblocks: []\n")
	svc := core.MustNewService(NewFS(root), core.WithClock(func() time.Time { return graphMutationNow }))
	receipt, err := svc.MigrateTaskDependencies(false)
	if err != nil || !receipt.Changed || len(receipt.ClearedLegacyFields) != 3 ||
		!slices.Equal(receipt.AppliedTaskIDs, []string{taskID}) {
		t.Fatalf("empty legacy migration=%+v err=%v", receipt, err)
	}
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	for _, field := range []string{"blocked_by:", "dependencies:", "blocks:"} {
		if strings.Contains(string(content), field) {
			t.Fatalf("empty legacy field %s remains:\n%s", field, content)
		}
	}
}

// Count Service attempts without replacing any filesystem behavior. A conflict
// after an actual write must not disappear into an automatic resume/no-op.
type dependencyMutationCountingFS struct {
	*FS
	mutationCalls int
}

func (s *dependencyMutationCountingFS) MutateTaskGraph(now time.Time, dryRun bool, planner core.TaskGraphPlanner) (core.TaskGraphMutationResult, error) {
	s.mutationCalls++
	return s.FS.MutateTaskGraph(now, dryRun, planner)
}

func TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes(t *testing.T) {
	for failAfter := 1; failAfter <= 3; failAfter++ {
		t.Run(fmt.Sprintf("after-%d", failAfter), func(t *testing.T) {
			root := t.TempDir()
			paths, before := map[string]string{}, map[string][]byte{}
			for i, slug := range []string{"prefix-a", "prefix-b", "prefix-c", "prefix-d"} {
				status, legacy := domain.StatusCompleted, ""
				if i == 0 {
					status = domain.StatusReadyToStart
				}
				if i < 3 {
					legacy = "blocked_by: [" + []string{"prefix-b", "prefix-c", "prefix-d"}[i] + "]\n"
				}
				taskID := testutil.TaskID(slug)
				paths[taskID] = writeGraphMutationTask(t, root, slug, status, nil, legacy)
				var readErr error
				before[taskID], readErr = os.ReadFile(paths[taskID])
				if readErr != nil {
					t.Fatal(readErr)
				}
			}
			fs := &dependencyMutationCountingFS{FS: NewFS(root)}
			svc := core.MustNewService(fs, core.WithClock(func() time.Time { return graphMutationNow }), core.WithRetry(4, func(int) {}))

			original := testHookAfterGraphWrite
			defer func() { testHookAfterGraphWrite = original }()
			writes := 0
			testHookAfterGraphWrite = func(string) error {
				writes++
				if writes == failAfter {
					testHookAfterGraphWrite = nil
					return fmt.Errorf("injected interruption: %w", domain.ErrConflict)
				}
				return nil
			}
			partial, err := svc.MigrateTaskDependencies(false)
			var failure *core.DependencyMutationFailure
			if !errors.Is(err, domain.ErrConflict) || !errors.As(err, &failure) || fs.mutationCalls != 1 ||
				len(partial.AppliedTaskIDs) != failAfter || len(partial.RemainingTaskIDs) != 3-failAfter {
				t.Fatalf("prefix %d receipt=%+v err=%v", failAfter, partial, err)
			}
			for taskID, path := range paths {
				after, readErr := os.ReadFile(path)
				if readErr != nil {
					t.Fatal(readErr)
				}
				changed := !bytes.Equal(before[taskID], after)
				if changed != slices.Contains(partial.AppliedTaskIDs, taskID) ||
					!strings.Contains(string(after), "Body stays intact.") {
					t.Fatalf("durable bytes disagree with receipt for %s: %+v", taskID, partial)
				}
			}
			if failAfter == 3 && !strings.Contains(err.Error(), "all planned dependency task files were durably applied") {
				t.Fatalf("final-write failure did not explain fully durable result: %v", err)
			}
			if partial.Remedy == "" || !strings.Contains(err.Error(), partial.Remedy) || strings.Contains(err.Error(), "retry the same command to converge") {
				t.Fatalf("prefix %d lost core inspection guidance: receipt=%+v err=%v", failAfter, partial, err)
			}
			if failAfter == 3 && !strings.Contains(partial.Remedy, "writes already landed") {
				t.Fatalf("final-write recovery implied outstanding writes: %+v", partial)
			}
			if failAfter < 3 && !strings.Contains(partial.Remedy, "applied/remaining task IDs") {
				t.Fatalf("partial-write recovery omitted inspection of durable progress: %+v", partial)
			}
			graph, loadErr := core.LoadTaskGraph(NewFS(root))
			if loadErr != nil || graph.Health() == core.GraphBroken {
				t.Fatalf("prefix %d left broken graph: health=%s err=%v problems=%+v", failAfter, graph.Health(), loadErr, graph.Problems())
			}
			// An operator can change remaining intent after inspecting the prefix.
			// Explicit resume must plan from those live bytes, not replay the old plan.
			var editedID string
			if len(partial.RemainingTaskIDs) > 0 {
				editedID = partial.RemainingTaskIDs[0]
				content, readErr := os.ReadFile(paths[editedID])
				if readErr != nil {
					t.Fatal(readErr)
				}
				for _, prerequisite := range []string{"prefix-b", "prefix-c", "prefix-d"} {
					content = bytes.ReplaceAll(content, []byte("blocked_by: ["+prerequisite+"]"), []byte("blocked_by: []"))
				}
				content = append(content, []byte("\nOperator changed remaining intent.\n")...)
				if writeErr := os.WriteFile(paths[editedID], content, 0o644); writeErr != nil {
					t.Fatal(writeErr)
				}
			}
			completed, err := svc.MigrateTaskDependencies(false)
			if err != nil || fs.mutationCalls != 2 || len(completed.AppliedTaskIDs) != 3-failAfter {
				t.Fatalf("prefix %d rerun=%+v err=%v", failAfter, completed, err)
			}
			if editedID != "" {
				task, body, readErr := fs.GetTask(editedID)
				if readErr != nil || len(task.DependsOn) != 0 || !strings.Contains(body, "Operator changed remaining intent.") {
					t.Fatalf("resume ignored current operator intent: task=%+v body=%s err=%v", task, body, readErr)
				}
			}
			graph, loadErr = core.LoadTaskGraph(NewFS(root))
			if loadErr != nil || graph.Health() != core.GraphHealthy {
				t.Fatalf("prefix %d rerun health=%s err=%v", failAfter, graph.Health(), loadErr)
			}
		})
	}
}
