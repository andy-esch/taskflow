package store

import (
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestThreadPlannerCannotRewriteTerminalLifecycleAuthorization(t *testing.T) {
	root := t.TempDir()
	memberID := testutil.TaskID("terminal-source-member")
	writeGraphMutationTask(t, root, "terminal-source-member", domain.StatusCompleted, nil, "")
	created, svc := createThreadForMutation(t, root, "terminal-source-thread", memberID)
	if _, err := svc.StartThread(created.Thread.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CancelThread(created.Thread.ID, false); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(created.Local.CommittedPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, dry := range []bool{true, false} {
		result, err := testutil.Must(NewFS(root, core.UnrestrictedMutations())).MutateThread(threadMutationStoreNow, dry, func(snapshot core.ThreadMutationSnapshot) (core.ThreadMutationPlan, error) {
			snapshot.Threads[0].Status = domain.ThreadStatusUnstarted
			return core.ThreadMutationPlan{ThreadID: created.Thread.ID, Operation: core.ThreadMutationStart}, nil
		})
		if !errors.Is(err, domain.ErrValidation) || result.Committed {
			t.Fatalf("callback rewrote authorization: result=%+v err=%v", result, err)
		}
		after, err := os.ReadFile(created.Local.CommittedPath)
		if err != nil || !slices.Equal(before, after) {
			t.Fatalf("cancelled Thread changed: err=%v", err)
		}
	}
}

func TestThreadPlannerDispatchersIsolateNestedValuesAndBodies(t *testing.T) {
	for _, route := range []string{"mutation", "creation", "apply"} {
		t.Run(route, func(t *testing.T) {
			thread := domain.Thread{ID: testutil.TaskID("snapshot-owner"), Tags: []string{"original"}, Tasks: []string{testutil.TaskID("snapshot-member")}}
			threads := []domain.Thread{thread}
			bodies := map[string]string{thread.ID: "Original body"}
			fs := testutil.Must(NewFS(t.TempDir(), core.ReadOnlyMutations()))
			mutate := func(input []domain.Thread) {
				input[0].ID = "rewritten"
				input[0].Tags[0] = "rewritten"
				input[0].Tasks[0] = "rewritten"
			}
			var err error
			switch route {
			case "mutation":
				_, err = callThreadMutationPlanner(fs, func(snapshot core.ThreadMutationSnapshot) (core.ThreadMutationPlan, error) {
					mutate(snapshot.Threads)
					return core.ThreadMutationPlan{}, nil
				}, core.ThreadMutationSnapshot{Threads: threads})
			case "creation":
				_, err = callThreadCreationPlanner(fs, func(snapshot core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
					mutate(snapshot.Threads)
					return core.ThreadCreationPlan{}, nil
				}, core.ThreadCreationSnapshot{Threads: threads})
			case "apply":
				_, err = callThreadApplyPlanner(fs, func(snapshot core.ThreadApplySnapshot) (core.ThreadApplyPlan, error) {
					mutate(snapshot.Threads)
					snapshot.ThreadBodies[thread.ID] = "Rewritten body"
					return core.ThreadApplyPlan{}, nil
				}, core.ThreadApplySnapshot{Threads: threads, ThreadBodies: bodies})
			}
			if err != nil || threads[0].ID != thread.ID || !reflect.DeepEqual(threads[0].Tags, []string{"original"}) || !reflect.DeepEqual(threads[0].Tasks, []string{testutil.TaskID("snapshot-member")}) || bodies[thread.ID] != "Original body" {
				t.Fatalf("%s callback rewrote owner evidence: threads=%+v bodies=%+v err=%v", route, threads, bodies, err)
			}
		})
	}
}
