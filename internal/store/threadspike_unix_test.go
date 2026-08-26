//go:build unix

package store

import (
	"strings"
	"sync"
	"testing"

	"github.com/andy-esch/taskflow/internal/threadspike"
)

func TestThreadSpikeConcurrentOppositeEdgesCannotCommitACycle(t *testing.T) {
	repo, ids := spikeRepo(t, "alpha", "beta")
	adapter := NewThreadSpikeRepository(repo.Root, "planning-space")
	plans := []threadspike.MaterializedPlan{
		{Schema: threadspike.PlanSchema, RepoID: "planning-space", Edges: []threadspike.Edge{{From: ids["alpha"], To: ids["beta"]}}},
		{Schema: threadspike.PlanSchema, RepoID: "planning-space", Edges: []threadspike.Edge{{From: ids["beta"], To: ids["alpha"]}}},
	}
	type result struct {
		receipt threadspike.Receipt
		err     error
	}
	start := make(chan struct{})
	results := make(chan result, len(plans))
	var ready sync.WaitGroup
	ready.Add(len(plans))
	for _, plan := range plans {
		plan := plan
		go func() {
			ready.Done()
			<-start
			receipt, err := adapter.Apply(plan, threadspike.ApplyOptions{})
			results <- result{receipt: receipt, err: err}
		}()
	}
	ready.Wait()
	close(start)

	var succeeded, rejected int
	for range plans {
		result := <-results
		if result.err == nil {
			succeeded++
			if !result.receipt.Complete {
				t.Fatalf("successful receipt incomplete: %+v", result.receipt)
			}
		} else {
			rejected++
			if !strings.Contains(result.err.Error(), "dependency cycle") {
				t.Fatalf("rejection = %v", result.err)
			}
		}
	}
	if succeeded != 1 || rejected != 1 {
		t.Fatalf("succeeded=%d rejected=%d", succeeded, rejected)
	}
	final, err := adapter.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(final.Problems) != 0 {
		t.Fatalf("snapshot problems = %+v", final.Problems)
	}
	if err := threadspike.NewGraph(final.Tasks).Validate(); err != nil {
		t.Fatal(err)
	}
	if got := len(final.Tasks[ids["alpha"]].DependsOn) + len(final.Tasks[ids["beta"]].DependsOn); got != 1 {
		t.Fatalf("committed edge count = %d", got)
	}
}
