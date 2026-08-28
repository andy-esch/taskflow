package store

import (
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

	svc := core.NewService(NewFS(root), core.WithClock(func() time.Time { return graphMutationNow }))
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
	svc := core.NewService(NewFS(root), core.WithClock(func() time.Time { return graphMutationNow }))

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
	svc := core.NewService(NewFS(root), core.WithClock(func() time.Time { return graphMutationNow }))

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
	svc := core.NewService(NewFS(root), core.WithClock(func() time.Time { return graphMutationNow }))
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

func TestDependencyMigrationEveryDurablePrefixStaysSoundAndResumes(t *testing.T) {
	for failAfter := 1; failAfter <= 3; failAfter++ {
		t.Run(fmt.Sprintf("after-%d", failAfter), func(t *testing.T) {
			root := t.TempDir()
			writeGraphMutationTask(t, root, "prefix-a", domain.StatusReadyToStart, nil, "blocked_by: [prefix-b]\n")
			writeGraphMutationTask(t, root, "prefix-b", domain.StatusCompleted, nil, "blocked_by: [prefix-c]\n")
			writeGraphMutationTask(t, root, "prefix-c", domain.StatusCompleted, nil, "blocked_by: [prefix-d]\n")
			writeGraphMutationTask(t, root, "prefix-d", domain.StatusCompleted, nil, "")
			svc := core.NewService(NewFS(root), core.WithClock(func() time.Time { return graphMutationNow }))

			original := testHookAfterGraphWrite
			defer func() { testHookAfterGraphWrite = original }()
			writes := 0
			testHookAfterGraphWrite = func(string) error {
				writes++
				if writes == failAfter {
					testHookAfterGraphWrite = nil
					return errors.New("injected interruption")
				}
				return nil
			}
			partial, err := svc.MigrateTaskDependencies(false)
			if err == nil || len(partial.AppliedTaskIDs) != failAfter || len(partial.RemainingTaskIDs) != 3-failAfter {
				t.Fatalf("prefix %d receipt=%+v err=%v", failAfter, partial, err)
			}
			if failAfter == 3 && !strings.Contains(err.Error(), "all planned dependency task files were durably applied") {
				t.Fatalf("final-write failure did not explain fully durable result: %v", err)
			}
			graph, loadErr := core.LoadTaskGraph(NewFS(root))
			if loadErr != nil || graph.Health() == core.GraphBroken {
				t.Fatalf("prefix %d left broken graph: health=%s err=%v problems=%+v", failAfter, graph.Health(), loadErr, graph.Problems())
			}
			completed, err := svc.MigrateTaskDependencies(false)
			if err != nil || len(completed.AppliedTaskIDs) != 3-failAfter {
				t.Fatalf("prefix %d rerun=%+v err=%v", failAfter, completed, err)
			}
			graph, loadErr = core.LoadTaskGraph(NewFS(root))
			if loadErr != nil || graph.Health() != core.GraphHealthy {
				t.Fatalf("prefix %d rerun health=%s err=%v", failAfter, graph.Health(), loadErr)
			}
		})
	}
}
