package core

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// These exercise the service use-cases at their own seam (via fakeStore) —
// previously they were covered only indirectly through the CLI tests, so a
// core regression surfaced as a confusing CLI-level failure.

func TestService_Lint(t *testing.T) {
	svc := NewService(&fakeStore{
		epics: []domain.Epic{{ID: "e1"}},
		tasks: []domain.Task{
			// Clean active task: no issues.
			{Slug: "clean", Status: domain.StatusInProgress, Declared: domain.StatusInProgress,
				Epic: "e1", Description: "fine", Tags: []string{"go"}, Tier: 3, Priority: "medium",
				Effort: "Unknown", Created: "2026-06-12"},
			// Active with a dangling epic + missing fields: full lint applies.
			{Slug: "dangling", Status: domain.StatusReadyToStart, Declared: domain.StatusReadyToStart,
				Epic: "ghost", Description: "d", Tags: []string{"x"}, Tier: 3, Priority: "medium"},
			// Archived: only misfiled drift is reported, not missing fields.
			{Slug: "archived-misfiled", Status: domain.StatusCompleted, Declared: domain.StatusInProgress},
			{Slug: "archived-clean", Status: domain.StatusCompleted, Declared: domain.StatusCompleted},
		},
		problems: []domain.FileProblem{{Path: "x.md", Message: "broken"}},
	})

	results, problems, err := svc.Lint()
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 {
		t.Errorf("file problems must pass through, got %d", len(problems))
	}
	got := map[string]string{}
	for _, r := range results {
		var msgs []string
		for _, i := range r.Issues {
			msgs = append(msgs, i.Field+": "+i.Message)
		}
		got[r.Slug] = strings.Join(msgs, "; ")
	}
	if _, ok := got["clean"]; ok {
		t.Errorf("clean task should have no issues: %q", got["clean"])
	}
	if !strings.Contains(got["dangling"], "epic") {
		t.Errorf("dangling epic should be flagged, got %q", got["dangling"])
	}
	if got["archived-misfiled"] == "" {
		t.Error("an archived misfiled task should still be flagged")
	}
	if _, ok := got["archived-clean"]; ok {
		t.Error("a clean archived task must not be nagged about missing fields")
	}
}

func TestService_ListAudits_BucketFilters(t *testing.T) {
	svc := NewService(&fakeStore{audits: []domain.Audit{
		{Slug: "a-open", Bucket: domain.AuditOpen},
		{Slug: "a-closed", Bucket: domain.AuditClosed},
		{Slug: "a-deferred", Bucket: domain.AuditDeferred},
	}})

	for _, tc := range []struct {
		bucket string
		all    bool
		want   []string
	}{
		{"", false, []string{"a-open"}}, // default: open only
		{"", true, []string{"a-open", "a-closed", "a-deferred"}},
		{"closed", false, []string{"a-closed"}},
	} {
		audits, _, err := svc.ListAudits(tc.bucket, tc.all)
		if err != nil {
			t.Fatal(err)
		}
		var slugs []string
		for _, a := range audits {
			slugs = append(slugs, a.Slug)
		}
		if strings.Join(slugs, ",") != strings.Join(tc.want, ",") {
			t.Errorf("ListAudits(%q, %v) = %v, want %v", tc.bucket, tc.all, slugs, tc.want)
		}
	}
}

func TestService_NewEpic(t *testing.T) {
	svc := NewService(&fakeStore{})

	// Validation failures: nothing reaches the store.
	for _, p := range []NewEpicParams{
		{Title: "X", Description: "", Priority: "medium"},                       // description required
		{Title: "X", Description: "d", Priority: "urgent"},                      // invalid priority
		{Title: "!!!", Description: "d", Priority: "medium"},                    // empty slug
		{Title: "X", Description: strings.Repeat("a", 200), Priority: "medium"}, // too long
	} {
		if _, err := svc.NewEpic(p); err == nil {
			t.Errorf("NewEpic(%+v) should fail validation", p)
		}
	}

	// The happy path stamps created and passes the slug through.
	e, err := svc.NewEpic(NewEpicParams{Title: "My Epic", Description: "d", Priority: "medium", Status: "planning"})
	if err != nil {
		t.Fatal(err)
	}
	if e.Created == "" || e.Status != "planning" {
		t.Errorf("created epic wrong: %+v", e)
	}
}
