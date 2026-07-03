package core

import (
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestBoard_ActivePipelineOnlyInOrder(t *testing.T) {
	svc := NewService(&fakeStore{tasks: []domain.Task{
		{Slug: "a", Status: domain.StatusInProgress},
		{Slug: "b", Status: domain.StatusNextUp},
		{Slug: "c", Status: domain.StatusInProgress},
		{Slug: "d", Status: domain.StatusReadyToStart},
		{Slug: "done", Status: domain.StatusCompleted},  // terminal → excluded
		{Slug: "dead", Status: domain.StatusDeprecated}, // terminal → excluded
		{Slug: "parked", Status: domain.StatusDeferred}, // parked → excluded
	}})
	b, err := svc.Board()
	if err != nil {
		t.Fatal(err)
	}

	want := []domain.Status{domain.StatusNextUp, domain.StatusReadyToStart, domain.StatusInProgress}
	if len(b.Columns) != len(want) {
		t.Fatalf("got %d columns, want %d", len(b.Columns), len(want))
	}
	counts := map[domain.Status]int{}
	for i, c := range b.Columns {
		if c.Status != want[i] {
			t.Errorf("column %d = %q, want %q", i, c.Status, want[i])
		}
		counts[c.Status] = len(c.Tasks)
		for _, tk := range c.Tasks {
			if !tk.Status.IsActive() {
				t.Errorf("non-active task %q leaked onto the board (status %q)", tk.Slug, tk.Status)
			}
		}
	}
	if counts[domain.StatusInProgress] != 2 || counts[domain.StatusNextUp] != 1 || counts[domain.StatusReadyToStart] != 1 {
		t.Errorf("wrong per-column counts: %v", counts)
	}
}

func TestBoard_EmptyColumnsStillPresent(t *testing.T) {
	svc := NewService(&fakeStore{tasks: []domain.Task{
		{Slug: "a", Status: domain.StatusInProgress},
	}})
	b, err := svc.Board()
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Columns) != 3 {
		t.Fatalf("got %d columns, want 3 (empty statuses still show a column)", len(b.Columns))
	}
	for _, c := range b.Columns {
		if c.Status != domain.StatusInProgress && len(c.Tasks) != 0 {
			t.Errorf("column %q should be empty, has %d", c.Status, len(c.Tasks))
		}
	}
}
