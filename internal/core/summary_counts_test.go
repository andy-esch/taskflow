package core

import (
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// SplitCounts is the partition BOTH dashboards render, so its three properties —
// active/archived membership, dropping empty buckets, and preserving Counts' display
// order — are pinned here rather than re-asserted per surface.
func TestSummarySplitCounts(t *testing.T) {
	s := Summary{Counts: []StatusCount{
		{Status: domain.StatusNextUp, Count: 4},
		{Status: domain.StatusReadyToStart, Count: 0}, // empty bucket: dropped, not zero-rendered
		{Status: domain.StatusInProgress, Count: 3},
		{Status: domain.StatusCompleted, Count: 9},
		{Status: domain.StatusDeprecated, Count: 0}, // empty bucket on the archived side too
		{Status: domain.StatusDeferred, Count: 2},
	}}
	active, archived := s.SplitCounts()

	wantActive := []StatusCount{
		{Status: domain.StatusNextUp, Count: 4},
		{Status: domain.StatusInProgress, Count: 3},
	}
	wantArchived := []StatusCount{
		{Status: domain.StatusCompleted, Count: 9},
		{Status: domain.StatusDeferred, Count: 2},
	}
	assertCounts(t, "active", active, wantActive)
	assertCounts(t, "archived", archived, wantArchived)
}

// An all-empty Counts yields nothing at all, so a surface can ask "is there anything
// to show?" by length instead of scanning for a non-zero.
func TestSummarySplitCountsAllEmpty(t *testing.T) {
	s := Summary{Counts: []StatusCount{
		{Status: domain.StatusNextUp, Count: 0},
		{Status: domain.StatusCompleted, Count: 0},
	}}
	if active, archived := s.SplitCounts(); len(active) != 0 || len(archived) != 0 {
		t.Errorf("all-empty counts should split to nothing, got active=%v archived=%v", active, archived)
	}
}

// Every status the domain knows lands on exactly one side — so a status added later
// can't silently vanish from both dashboards.
func TestSummarySplitCountsCoversEveryStatus(t *testing.T) {
	all := domain.AllStatuses()
	counts := make([]StatusCount, 0, len(all))
	for _, st := range all {
		counts = append(counts, StatusCount{Status: st, Count: 1})
	}
	active, archived := Summary{Counts: counts}.SplitCounts()
	if got := len(active) + len(archived); got != len(all) {
		t.Errorf("split kept %d of %d statuses; every status must land on exactly one side", got, len(all))
	}
	for _, c := range active {
		if !c.Status.IsActive() {
			t.Errorf("%s is on the active side but IsActive() is false", c.Status)
		}
	}
	for _, c := range archived {
		if c.Status.IsActive() {
			t.Errorf("%s is on the archived side but IsActive() is true", c.Status)
		}
	}
}

func assertCounts(t *testing.T, side string, got, want []StatusCount) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", side, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %v, want %v (Counts' display order is preserved)", side, i, got[i], want[i])
		}
	}
}
