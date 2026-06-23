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
	// Capture the duplicate's message by value (no *LintResult): a nil-check-then-
	// dereference trips staticcheck SA5011 on older versions that don't treat
	// t.Fatalf as terminating the nil path.
	dupMsg, found := "", false
	for _, r := range results {
		if r.Slug == "solo" {
			t.Errorf("a unique slug must not be flagged: %+v", r)
		}
		if r.Slug == "dup" && len(r.Issues) > 0 && r.Issues[0].Field == "slug" {
			dupMsg, found = r.Issues[0].Message, true
		}
	}
	if !found {
		t.Fatalf("expected a duplicate-slug issue for 'dup', got %+v", results)
	}
	if !strings.Contains(dupMsg, "duplicate") ||
		!strings.Contains(dupMsg, "completed") ||
		!strings.Contains(dupMsg, "deprecated") {
		t.Errorf("the issue should name the duplicate's two dirs, got %q", dupMsg)
	}
}
