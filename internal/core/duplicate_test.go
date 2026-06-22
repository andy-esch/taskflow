package core

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestLint_DuplicateSlug pins M11: a slug present in two status dirs (a Ctrl-C in
// Move's write-then-remove window) is flagged loudly by lint — otherwise every
// resolve(slug) is a silent ErrAmbiguous with no signal. A unique slug isn't.
func TestLint_DuplicateSlug(t *testing.T) {
	svc := NewService(&fakeStore{
		tasks: []domain.Task{
			// Same slug in two dirs; both archived + folder-matching (the Move-crash
			// shape), so neither has per-field lint noise — the only issue is the dup.
			{Slug: "dup", Status: domain.StatusCompleted, Declared: domain.StatusCompleted},
			{Slug: "dup", Status: domain.StatusDeprecated, Declared: domain.StatusDeprecated},
			{Slug: "solo", Status: domain.StatusCompleted, Declared: domain.StatusCompleted},
		},
	})
	results, _, err := svc.Lint()
	if err != nil {
		t.Fatal(err)
	}
	var dup *LintResult
	for i := range results {
		if results[i].Slug == "dup" && results[i].Issues[0].Field == "slug" {
			dup = &results[i]
		}
		if results[i].Slug == "solo" {
			t.Errorf("a unique slug must not be flagged: %+v", results[i])
		}
	}
	if dup == nil {
		t.Fatalf("expected a duplicate-slug issue for 'dup', got %+v", results)
	}
	if !strings.Contains(dup.Issues[0].Message, "duplicate") ||
		!strings.Contains(dup.Issues[0].Message, "completed") ||
		!strings.Contains(dup.Issues[0].Message, "deprecated") {
		t.Errorf("the issue should name the duplicate's two dirs, got %q", dup.Issues[0].Message)
	}
}
