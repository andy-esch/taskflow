package domain

import "testing"

// TestEpicRefKey pins the Scheme-2 rule: an `epic:` reference resolves on its leading
// NN number, so a drifted slug still matches the epic (and a legacy non-NN stem falls
// back to exact-match on the whole string).
func TestEpicRefKey(t *testing.T) {
	cases := []struct{ ref, want string }{
		{"24-data-model-x", "24"},
		{"24-renamed-later", "24"}, // same epic despite a different slug
		{"24", "24"},
		{"00-taskflow-v1-core", "00"},
		{"100-scale", "100"},
		{"taskflow-v1-core", "taskflow-v1-core"}, // no NN → whole-string fallback
		{"", ""},
	}
	for _, c := range cases {
		if got := EpicRefKey(c.ref); got != c.want {
			t.Errorf("EpicRefKey(%q) = %q, want %q", c.ref, got, c.want)
		}
	}
}

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
