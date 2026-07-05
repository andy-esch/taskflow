package store

import (
	"os"
	"path/filepath"
	"testing"
)

// goodID2 is a second valid 12-char id (the routines-spike task's id).
const goodID2 = "6fjw1d2jm6e9"

// resolveID over flat candidates matches on either the stable id (a prefix) or
// the human slug (prefix/substring) — the ADR-0003 §4 resolution rule.
func TestResolveID_FlatByIDAndSlug(t *testing.T) {
	cands := []candidate{
		{id: goodID, slug: "retry-backoff", path: "a"},
		{id: goodID2, slug: "flatten-layout", path: "b"},
	}
	tests := []struct {
		query   string
		want    string // resolved path
		wantErr bool
	}{
		{goodID, "a", false},          // exact id
		{"6fjvr0", "a", false},        // id prefix
		{"retry-backoff", "a", false}, // exact slug
		{"retry", "a", false},         // slug substring
		{"flatten", "b", false},       // slug substring on the other candidate
		{"6fj", "", true},             // id prefix matching BOTH -> ambiguous
		{"nope", "", true},            // no match
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			c, err := resolveID("task", tc.query, cands)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveID(%q): want error, got %+v", tc.query, c)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveID(%q): %v", tc.query, err)
			}
			if c.path != tc.want {
				t.Errorf("resolveID(%q) = %q, want %q", tc.query, c.path, tc.want)
			}
		})
	}
}

// Two entities can share a slug under the flat layout (unique by id); the shared
// slug is ambiguous, but each id resolves it uniquely.
func TestResolveID_DupSlugDisambiguatedByID(t *testing.T) {
	cands := []candidate{
		{id: goodID, slug: "cleanup", path: "a"},
		{id: goodID2, slug: "cleanup", path: "b"},
	}
	if _, err := resolveID("task", "cleanup", cands); err == nil {
		t.Fatal("a dup slug should resolve ambiguously")
	}
	if c, err := resolveID("task", goodID, cands); err != nil || c.path != "a" {
		t.Fatalf("full id A: got (%+v, %v)", c, err)
	}
	if c, err := resolveID("task", "6fjw1d", cands); err != nil || c.path != "b" {
		t.Fatalf("id B prefix: got (%+v, %v)", c, err)
	}
}

// flatCandidates yields a candidate per id-led file and silently skips every
// non-entity name (the carveout gate).
func TestFlatCandidates_SkipsNonEntities(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		goodID + "-retry-backoff.md",    // entity
		goodID2 + "-2026-06-16-area.md", // entity (audit-style, date in the slug)
		"HOWTO-execute.md",              // non-entity
		"README.md",                     // non-entity
		"notes.md",                      // non-entity
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cands, err := flatCandidates(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) != 2 {
		t.Fatalf("want 2 entity candidates, got %d: %+v", len(cands), cands)
	}
	got := map[string]string{}
	for _, c := range cands {
		got[c.id] = c.slug
	}
	if got[goodID] != "retry-backoff" || got[goodID2] != "2026-06-16-area" {
		t.Errorf("unexpected candidates: %+v", cands)
	}
}
