package core

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// researchStore records what NewResearch hands the store.
type researchStore struct {
	nopStore
	docs        []domain.Research
	created     []domain.Research
	createdBody []string
}

func (f *researchStore) ListResearch() ([]domain.Research, []domain.FileProblem, error) {
	return f.docs, nil, nil
}

func (f *researchStore) CreateResearch(r domain.Research, body string, dryRun bool) (domain.Research, error) {
	if !dryRun {
		f.created = append(f.created, r)
		f.createdBody = append(f.createdBody, body)
	}
	return r, nil
}

func fixedClock(date string) func() time.Time {
	return func() time.Time {
		t, _ := time.Parse("2006-01-02", date)
		return t
	}
}

// The headline property: the id is minted from `created`, NOT from "now", so a
// backdated doc sorts into place chronologically. Ids are lexically ordered by their
// embedded millisecond (ADR-0003 §3), so an older created date must yield a smaller id.
func TestNewResearch_IDMintedFromCreatedNotNow(t *testing.T) {
	store := &researchStore{}
	svc := NewService(store, WithClock(fixedClock("2026-08-14")))

	older, err := svc.NewResearch(NewResearchParams{Title: "Old work", Created: "2026-01-03"})
	if err != nil {
		t.Fatal(err)
	}
	newer, err := svc.NewResearch(NewResearchParams{Title: "Recent work", Created: "2026-06-23"})
	if err != nil {
		t.Fatal(err)
	}
	if older.ID >= newer.ID {
		t.Errorf("id order must follow created order: %q (2026-01-03) should sort before %q (2026-06-23)", older.ID, newer.ID)
	}
	// And the id genuinely decodes back to the created date, not to the clock. Compared
	// in UTC: a date-only string parses to UTC midnight, so the encoded instant is
	// UTC-anchored (id.Time returns it in local zone, which is a day earlier west of
	// Greenwich). This matches flatmigrate's firstDate→UnixMilli minting exactly, so
	// backdated ids stay consistent with the ones already in the tree.
	if got := id.Time(older.ID).UTC().Format("2006-01-02"); got != "2026-01-03" {
		t.Errorf("id encodes %s, want the created date 2026-01-03 (not the 2026-08-14 clock)", got)
	}
}

func TestNewResearch_DefaultsCreatedToToday(t *testing.T) {
	store := &researchStore{}
	svc := NewService(store, WithClock(fixedClock("2026-08-14")))

	r, err := svc.NewResearch(NewResearchParams{Title: "Today's work"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Created != "2026-08-14" {
		t.Errorf("Created = %q, want the clock's date", r.Created)
	}
}

func TestNewResearch_SlugAndBodyFromTitle(t *testing.T) {
	store := &researchStore{}
	svc := NewService(store, WithClock(fixedClock("2026-08-14")))

	r, err := svc.NewResearch(NewResearchParams{Title: "Compare: theming libs → a call"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Slug != "compare-theming-libs-a-call" {
		t.Errorf("slug = %q", r.Slug)
	}
	// The FULL original title survives as the body H1, punctuation and all.
	if len(store.createdBody) != 1 || !strings.Contains(store.createdBody[0], "# Compare: theming libs → a call") {
		t.Errorf("body must keep the full title as H1: %q", store.createdBody)
	}
	// The template's {{date}} is filled with created, not left as a placeholder.
	if !strings.Contains(store.createdBody[0], "as of 2026-08-14") {
		t.Errorf("body must fill {{date}}: %q", store.createdBody[0])
	}
}

func TestNewResearch_Validation(t *testing.T) {
	svc := NewService(&researchStore{}, WithClock(fixedClock("2026-08-14")))
	cases := []struct {
		name string
		p    NewResearchParams
	}{
		{"empty title", NewResearchParams{Title: "   "}},
		{"title slugifies to nothing", NewResearchParams{Title: "!!!"}},
		{"bad date", NewResearchParams{Title: "T", Created: "2026-6-3"}},
		{"body and template together", NewResearchParams{Title: "T", Body: "x", Template: "default"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.NewResearch(tc.p); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("want ErrValidation, got %v", err)
			}
		})
	}
}

func TestNewResearch_DryRunWritesNothing(t *testing.T) {
	store := &researchStore{}
	svc := NewService(store, WithClock(fixedClock("2026-08-14")))

	if _, err := svc.NewResearch(NewResearchParams{Title: "Preview", DryRun: true}); err != nil {
		t.Fatal(err)
	}
	if len(store.created) != 0 {
		t.Errorf("dry run must not create: %+v", store.created)
	}
}

// ListResearch is newest-first with a stable slug tiebreak — same-day docs are the
// common case (dates are day-precision), so the tiebreak is what keeps output stable.
func TestListResearch_NewestFirstStableTiebreak(t *testing.T) {
	store := &researchStore{docs: []domain.Research{
		{Slug: "zebra", Created: "2026-01-03"},
		{Slug: "alpha", Created: "2026-01-03"},
		{Slug: "middle", Created: "2026-06-23"},
	}}
	svc := NewService(store)

	got, _, err := svc.ListResearch("")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"middle", "alpha", "zebra"}
	for i, w := range want {
		if got[i].Slug != w {
			t.Fatalf("order = %v, want %v", slugsOf(got), want)
		}
	}
}

func TestListResearch_TagFilterCaseInsensitive(t *testing.T) {
	store := &researchStore{docs: []domain.Research{
		{Slug: "tui-doc", Created: "2026-01-03", Tags: []string{"TUI", "color"}},
		{Slug: "core-doc", Created: "2026-01-04", Tags: []string{"core"}},
	}}
	svc := NewService(store)

	got, _, err := svc.ListResearch("tui")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Slug != "tui-doc" {
		t.Errorf("tag filter = %v, want [tui-doc]", slugsOf(got))
	}
}

func slugsOf(docs []domain.Research) []string {
	out := make([]string, len(docs))
	for i, r := range docs {
		out[i] = r.Slug
	}
	return out
}

// setStore records what SetResearchFields hands the store.
type setStore struct {
	nopStore
	got map[string]any
}

func (f *setStore) SetResearchFields(_ string, updates map[string]any, _ bool) (domain.Research, error) {
	f.got = updates
	return domain.Research{Slug: "x"}, nil
}

// `created` is the one field that must never be writable: the id is minted from it, so a
// change in place would leave the pair desynced with no way to detect it later.
func TestSetResearchFields_ProtectedFields(t *testing.T) {
	svc := NewService(&setStore{}, WithClock(fixedClock("2026-08-18")))
	for _, field := range []string{"created", "id", "schema", "updated_at"} {
		t.Run(field, func(t *testing.T) {
			_, err := svc.SetResearchFields("x", map[string]any{field: "whatever"}, false, false)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("setting %s must be ErrValidation, got %v", field, err)
			}
			if !strings.Contains(err.Error(), field) {
				t.Errorf("error should name the field: %v", err)
			}
		})
		// Protected on the UNSET path too — removing created is as damaging as changing it.
		t.Run(field+" unset", func(t *testing.T) {
			_, err := svc.SetResearchFields("x", map[string]any{field: domain.UnsetField{}}, false, false)
			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("unsetting %s must be ErrValidation, got %v", field, err)
			}
		})
	}
}

func TestSetResearchFields_UnknownFieldNeedsForce(t *testing.T) {
	store := &setStore{}
	svc := NewService(store, WithClock(fixedClock("2026-08-18")))

	if _, err := svc.SetResearchFields("x", map[string]any{"bogus": "1"}, false, false); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("unknown field without --force must be ErrValidation, got %v", err)
	}
	if _, err := svc.SetResearchFields("x", map[string]any{"bogus": "1"}, true, false); err != nil {
		t.Errorf("--force should allow it: %v", err)
	}
	if store.got["bogus"] != "1" {
		t.Errorf("forced field not passed through: %+v", store.got)
	}
}

// tags is the only list field, so `--set tags=a,b` must become a sequence rather than a
// single corrupting string.
func TestSetResearchFields_CoercesTagsToList(t *testing.T) {
	store := &setStore{}
	svc := NewService(store, WithClock(fixedClock("2026-08-18")))

	if _, err := svc.SetResearchFields("x", map[string]any{"tags": "a,b"}, false, false); err != nil {
		t.Fatal(err)
	}
	got, ok := store.got["tags"].([]string)
	if !ok || len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("tags = %#v, want []string{a,b}", store.got["tags"])
	}
}

// updated_at is stamped by the service so every adapter gets it.
func TestSetResearchFields_StampsUpdatedAt(t *testing.T) {
	store := &setStore{}
	svc := NewService(store, WithClock(fixedClock("2026-08-18")))

	if _, err := svc.SetResearchFields("x", map[string]any{"description": "d"}, false, false); err != nil {
		t.Fatal(err)
	}
	if store.got["updated_at"] != "2026-08-18" {
		t.Errorf("updated_at = %v, want the clock's date", store.got["updated_at"])
	}
}

func TestSetResearchFields_RejectsEmptyUpdate(t *testing.T) {
	svc := NewService(&setStore{}, WithClock(fixedClock("2026-08-18")))
	if _, err := svc.SetResearchFields("x", nil, false, false); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("no fields must be ErrValidation, got %v", err)
	}
}

// A too-long description is caught in core, so every adapter inherits the rule.
func TestSetResearchFields_ValidatesDescription(t *testing.T) {
	svc := NewService(&setStore{}, WithClock(fixedClock("2026-08-18")))
	long := strings.Repeat("x", domain.MaxDescriptionLen+1)
	if _, err := svc.SetResearchFields("x", map[string]any{"description": long}, false, false); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("over-long description must be ErrValidation, got %v", err)
	}
}
