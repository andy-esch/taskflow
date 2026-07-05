package store

import "testing"

// goodID is a real, valid 12-char id (the carveout task's own id) — Crockford
// base32, lowercase, no i/l/o/u.
const goodID = "6fjvr03mr9zg"

func TestSplitFlatName(t *testing.T) {
	cases := []struct {
		name     string
		stem     string
		wantID   string
		wantSlug string
		wantOK   bool
	}{
		// Entities: id leads, the slug (dashes and all) is the remainder.
		{"simple slug", goodID + "-retry-backoff", goodID, "retry-backoff", true},
		{"slug with many dashes", goodID + "-add-retry-to-the-webhook", goodID, "add-retry-to-the-webhook", true},
		{"audit date stays in the slug", goodID + "-2026-06-16-dispatcher", goodID, "2026-06-16-dispatcher", true},
		{"single-char slug", goodID + "-x", goodID, "x", true},

		// Carveout: non-entity names are not id-led -> ok=false, no id/slug.
		{"non-entity HOWTO", "HOWTO-execute", "", "", false},
		{"non-entity README", "README", "", "", false},
		{"bare id, no separator or slug", goodID, "", "", false},
		{"id + separator but empty slug", goodID + "-", "", "", false},
		{"invalid id chars (i excluded from Crockford)", "iiiiiiiiiiii-slug", "", "", false},
		{"too-short leading field", "abc-slug", "", "", false},
		{"wrong separator", goodID + "_slug", "", "", false},
		{"uppercase id (strict lowercase)", "6FJVR03MR9ZG-slug", "", "", false},
		{"empty", "", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotID, gotSlug, gotOK := splitFlatName(tc.stem)
			if gotID != tc.wantID || gotSlug != tc.wantSlug || gotOK != tc.wantOK {
				t.Errorf("splitFlatName(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tc.stem, gotID, gotSlug, gotOK, tc.wantID, tc.wantSlug, tc.wantOK)
			}
		})
	}
}
