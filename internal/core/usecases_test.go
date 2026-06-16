package core

import (
	"errors"
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

// TestService_NewTask_RequiresTags pins the D1 decision: `new` must not
// scaffold a file its own linter rejects, so tags are required at creation.
func TestService_NewTask_RequiresTags(t *testing.T) {
	svc := NewService(&fakeStore{epics: []domain.Epic{{ID: "e1"}}})
	_, err := svc.NewTask(NewTaskParams{Title: "X", Epic: "e1", Tier: 3, Autonomy: 3, Priority: "medium"})
	if err == nil || !strings.Contains(err.Error(), "tag") {
		t.Errorf("tagless create should fail mentioning tags, got %v", err)
	}
}

// TestService_Create_RejectsHostileTitle pins that the create path hard-fails on
// filename-hostile title characters (a colon + em-dash) rather than silently
// slugifying to a different name, for both tasks and epics.
func TestService_Create_RejectsHostileTitle(t *testing.T) {
	svc := NewService(&fakeStore{epics: []domain.Epic{{ID: "e1"}}})
	_, err := svc.NewTask(NewTaskParams{Title: "Fix: the thing — now", Epic: "e1", Tags: []string{"x"}, Tier: 3, Autonomy: 3, Priority: "medium"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("task new with a hostile title should be ErrValidation, got %v", err)
	}
	_, err = svc.NewEpic(NewEpicParams{Title: "Plan: phase — two", Description: "d", Priority: "medium", Status: "planning"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("epic new with a hostile title should be ErrValidation, got %v", err)
	}
	// A clean title still creates.
	if _, err := svc.NewTask(NewTaskParams{Title: "Fix the thing now", Epic: "e1", Tags: []string{"x"}, Tier: 3, Autonomy: 3, Priority: "medium"}); err != nil {
		t.Errorf("a clean title should create, got %v", err)
	}
}

// TestService_NewAudit pins the create logic that lives in core: the slug is
// <date>-<area-slug>, the date defaults to today when omitted, and area/date are
// validated (nothing reaches the store on bad input).
func TestService_NewAudit(t *testing.T) {
	fs := &fakeStore{}
	svc := NewService(fs)

	// Explicit date → slug is <date>-<area-slug>; area kept verbatim, bucket open.
	a, err := svc.NewAudit(NewAuditParams{Area: "Arch Data Flow", Date: "2026-06-16"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Slug != "2026-06-16-arch-data-flow" {
		t.Errorf("slug = %q, want 2026-06-16-arch-data-flow", a.Slug)
	}
	if a.Bucket != domain.AuditOpen || a.Area != "Arch Data Flow" || a.Date != "2026-06-16" {
		t.Errorf("audit fields wrong: %+v", a)
	}
	if len(fs.createdAudits) != 1 {
		t.Fatalf("expected one CreateAudit call, got %d", len(fs.createdAudits))
	}

	// Empty date defaults to today, so the slug still carries a valid date prefix.
	a2, err := svc.NewAudit(NewAuditParams{Area: "dispatcher"})
	if err != nil {
		t.Fatal(err)
	}
	datePart := strings.TrimSuffix(a2.Slug, "-dispatcher")
	if datePart == a2.Slug || domain.ValidateDate(datePart) != nil {
		t.Errorf("defaulted slug is not <today>-dispatcher: %q", a2.Slug)
	}

	// A missing area and a malformed date are both ErrValidation.
	if _, err := svc.NewAudit(NewAuditParams{Area: "   "}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("empty area should be ErrValidation, got %v", err)
	}
	if _, err := svc.NewAudit(NewAuditParams{Area: "x", Date: "06-16-2026"}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("bad date should be ErrValidation, got %v", err)
	}
}

// TestService_Lint_FlagsInvalidEpicStatus pins the D2 decision: epic status is
// a closed vocabulary, and files outside it surface in lint.
func TestService_Lint_FlagsInvalidEpicStatus(t *testing.T) {
	svc := NewService(&fakeStore{epics: []domain.Epic{
		{ID: "good", Status: "planning"},
		{ID: "weird", Status: "bananas"},
	}})
	results, _, err := svc.Lint()
	if err != nil {
		t.Fatal(err)
	}
	var flagged []string
	for _, r := range results {
		flagged = append(flagged, r.Slug)
	}
	if len(flagged) != 1 || flagged[0] != "weird" {
		t.Errorf("only the invalid epic status should be flagged, got %v", flagged)
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
