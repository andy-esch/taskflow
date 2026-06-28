package domain

import "testing"

// TestIsEpicArchived pins the ONLY statuses the dashboard/epics-tab default hides:
// the two known terminals. Everything else (active, foreign, empty) is non-archived
// so visibility fails open.
func TestIsEpicArchived(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{EpicStatusRetired, true},
		{EpicStatusDeprecated, true},
		{EpicStatusActive, false},
		{"planning", false},  // foreign vocabulary — not a known terminal
		{"completed", false}, // looks finished, but isn't a canonical terminal: still shows
		{"in-progress", false},
		{"", false}, // missing status — fail open
	}
	for _, c := range cases {
		if got := IsEpicArchived(c.status); got != c.want {
			t.Errorf("IsEpicArchived(%q) = %v, want %v", c.status, got, c.want)
		}
	}
}

// TestIsKnownEpicStatus pins the conformance check the read surfaces flag on: only
// the canonical three pass; anything else is a fixable data problem (flagged, not
// dropped).
func TestIsKnownEpicStatus(t *testing.T) {
	for _, s := range AllEpicStatuses() {
		if !IsKnownEpicStatus(s) {
			t.Errorf("IsKnownEpicStatus(%q) = false, want true (canonical status)", s)
		}
	}
	for _, s := range []string{"planning", "completed", "in-progress", "done", ""} {
		if IsKnownEpicStatus(s) {
			t.Errorf("IsKnownEpicStatus(%q) = true, want false (non-canonical)", s)
		}
	}
}
