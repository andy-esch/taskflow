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
		epics: []domain.Epic{
			// Valid epic the tasks join against.
			{ID: "e1", Status: "active", Priority: "medium", Description: "the epic"},
			// Bad epic: a typo'd status surfaces as its own LintResult (keyed by id).
			{ID: "e2-bad", Status: "bogus", Priority: "medium", Description: "d"},
		},
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
	// Epics are linted too: the bad-status epic surfaces as a result keyed by id.
	if !strings.Contains(got["e2-bad"], "status") {
		t.Errorf("bad epic status should be flagged, got %q", got["e2-bad"])
	}
	if _, ok := got["e1"]; ok {
		t.Errorf("the valid epic must not be flagged: %q", got["e1"])
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

// TestService_ListAudits_RejectsUnknownBucket mirrors ListTasks' invalid-status
// check (TestService_ListTasks_RejectsInvalidFilters): an unrecognized bucket is
// ErrValidation, not a silently empty list an agent can't tell from an empty one.
func TestService_ListAudits_RejectsUnknownBucket(t *testing.T) {
	svc := NewService(&fakeStore{audits: []domain.Audit{{Slug: "a-open", Bucket: domain.AuditOpen}}})
	if _, _, err := svc.ListAudits("bogus", false); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unknown bucket should be ErrValidation, got %v", err)
	}
}

// countingAuditStore wraps fakeStore to count how each audit-read path is hit, so
// a test can prove Summary reads every audit body exactly once (via the single
// ListAuditsWithFindings sweep) and never re-reads through GetAuditByPath — the H2
// regression guard.
type countingAuditStore struct {
	fakeStore
	listWithFindings int            // ListAuditsWithFindings call count
	byPath           map[string]int // GetAuditByPath reads, per path
}

func (c *countingAuditStore) ListAuditsWithFindings() ([]AuditWithFindings, []domain.FileProblem, error) {
	c.listWithFindings++
	return c.fakeStore.ListAuditsWithFindings()
}

func (c *countingAuditStore) GetAuditByPath(path string) (domain.Audit, string, error) {
	if c.byPath == nil {
		c.byPath = map[string]int{}
	}
	c.byPath[path]++
	return c.fakeStore.GetAuditByPath(path)
}

// TestService_Summary_ReadsEachAuditOnce pins H2: Summary computes the audit
// tallies AND the findings rollup from ONE sweep of the audit bodies — it must
// never re-read a body through GetAuditByPath the way it did before the fix.
func TestService_Summary_ReadsEachAuditOnce(t *testing.T) {
	store := &countingAuditStore{fakeStore: fakeStore{
		audits: []domain.Audit{
			{Slug: "2026-06-14-gateway", Path: "2026-06-14-gateway", Bucket: domain.AuditOpen},
			{Slug: "2026-06-10-ingest", Path: "2026-06-10-ingest", Bucket: domain.AuditClosed},
		},
		auditBodies: map[string]string{
			"2026-06-14-gateway": gatewayBody,
			"2026-06-10-ingest":  ingestBody,
		},
	}}
	s, err := NewService(store).Summary()
	if err != nil {
		t.Fatal(err)
	}
	// One scan, zero re-reads.
	if store.listWithFindings != 1 {
		t.Errorf("ListAuditsWithFindings called %d times, want exactly 1", store.listWithFindings)
	}
	if len(store.byPath) != 0 {
		t.Errorf("Summary must not re-read any audit via GetAuditByPath, got %v", store.byPath)
	}
	// And the rollup is still correct: S1 (gateway, open) + M1 (ingest, open) are
	// actionable; H1 is fixed, so Open == 2 across the two audits.
	if s.Findings.Open != 2 || s.Findings.InProgress != 0 {
		t.Errorf("findings rollup wrong: open=%d in_progress=%d, want 2/0", s.Findings.Open, s.Findings.InProgress)
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

// TestService_Create_SlugifiesHostileTitle pins the 2026-06-25 reversal: a
// filename-hostile title (colon + plus, the motivating case) is no longer
// rejected — it slugifies to a safe id while the FULL original title is kept as
// the body H1, across task / epic / audit create paths. The empty-slug error
// stays the only hard guard (covered by TestService_Create_EmptySlugStillErrors).
func TestService_Create_SlugifiesHostileTitle(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1"}}}
	svc := NewService(fs)

	// task: hostile title accepted, slug derived, full title preserved in the H1.
	tk, err := svc.NewTask(NewTaskParams{Title: "Wire OAuth: PKCE + refresh", Epic: "e1", Tags: []string{"x"}, Tier: 3, Autonomy: 3, Priority: "medium"})
	if err != nil {
		t.Fatalf("task new with a hostile title should succeed, got %v", err)
	}
	if tk.Slug != "wire-oauth-pkce-refresh" {
		t.Errorf("task slug = %q, want wire-oauth-pkce-refresh", tk.Slug)
	}
	if len(fs.createdBodies) != 1 || !strings.Contains(fs.createdBodies[0], "# Wire OAuth: PKCE + refresh") {
		t.Errorf("task body should keep the full title as the H1, got %q", fs.createdBodies)
	}

	// epic: same — slug derived (the store adds the NN- number; the fake passes the
	// slug through as the id), full title in the H1.
	ep, err := svc.NewEpic(NewEpicParams{Title: "Plan: phase — two", Description: "d", Priority: "medium", Status: "active"})
	if err != nil {
		t.Fatalf("epic new with a hostile title should succeed, got %v", err)
	}
	if ep.ID != "plan-phase-two" {
		t.Errorf("epic slug = %q, want plan-phase-two", ep.ID)
	}
	if len(fs.epicCreateBodies) != 1 || !strings.Contains(fs.epicCreateBodies[0], "# Plan: phase — two") {
		t.Errorf("epic body should keep the full title as the H1, got %q", fs.epicCreateBodies)
	}

	// audit: a hostile area slugifies into <date>-<area-slug>, full area in the body.
	au, err := svc.NewAudit(NewAuditParams{Area: "Auth: token / refresh", Date: "2026-06-25"})
	if err != nil {
		t.Fatalf("audit new with a hostile area should succeed, got %v", err)
	}
	if au.Slug != "2026-06-25-auth-token-refresh" {
		t.Errorf("audit slug = %q, want 2026-06-25-auth-token-refresh", au.Slug)
	}
	if au.Area != "Auth: token / refresh" {
		t.Errorf("audit area should keep the full original, got %q", au.Area)
	}
	if len(fs.auditCreateBodies) != 1 || !strings.Contains(fs.auditCreateBodies[0], "Auth: token / refresh") {
		t.Errorf("audit body should keep the full area, got %q", fs.auditCreateBodies)
	}
}

// TestService_Create_EmptySlugStillErrors pins the one remaining hard guard: a
// title/area made only of punctuation slugifies to "" and is rejected as
// ErrValidation, for all three create paths (nothing reaches the store).
func TestService_Create_EmptySlugStillErrors(t *testing.T) {
	fs := &fakeStore{epics: []domain.Epic{{ID: "e1"}}}
	svc := NewService(fs)

	if _, err := svc.NewTask(NewTaskParams{Title: "!!!", Epic: "e1", Tags: []string{"x"}, Tier: 3, Autonomy: 3, Priority: "medium"}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("punctuation-only task title should be ErrValidation, got %v", err)
	}
	if _, err := svc.NewEpic(NewEpicParams{Title: "...", Description: "d", Priority: "medium", Status: "active"}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("punctuation-only epic title should be ErrValidation, got %v", err)
	}
	if _, err := svc.NewAudit(NewAuditParams{Area: "***", Date: "2026-06-25"}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("punctuation-only audit area should be ErrValidation, got %v", err)
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
// a closed vocabulary, and files outside it surface in lint. The "good" epic is
// fully valid (status/priority/description) so only the typo'd one is flagged.
func TestService_Lint_FlagsInvalidEpicStatus(t *testing.T) {
	svc := NewService(&fakeStore{epics: []domain.Epic{
		{ID: "good", Status: "active", Priority: "medium", Description: "a goal"},
		{ID: "weird", Status: "bananas", Priority: "medium", Description: "a goal"},
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
	e, err := svc.NewEpic(NewEpicParams{Title: "My Epic", Description: "d", Priority: "medium", Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	if e.Created == "" || e.Status != "active" {
		t.Errorf("created epic wrong: %+v", e)
	}
}
