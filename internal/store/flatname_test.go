package store

import (
	"errors"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

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

// The three cases must be told apart: a stray belongs in meta/, a mistyped id belongs
// exactly where it is with one character corrected, and a u cannot be corrected at all.
// Collapsing them into "has no leading id" was the reported defect.
func TestEntityNameProblemDistinguishesStrayFromMistypedID(t *testing.T) {
	reason, stray := entityNameProblem("notes.md")
	if !errors.Is(stray, errNotEntity) || !strings.Contains(reason, "meta/") {
		t.Errorf("stray = %v / %q; want the not-an-entity advice", stray, reason)
	}

	reason, bad := entityNameProblem("6fbj87000lt6-bad-id.md")
	if !errors.Is(bad, errBadEntityID) {
		t.Errorf("a mistyped id was classified as %v, not a bad id", bad)
	}
	for _, want := range []string{`"l"`, "position 9", "6fbj870001t6-bad-id.md"} {
		if !strings.Contains(reason, want) {
			t.Errorf("reason %q does not mention %q", reason, want)
		}
	}
	if strings.Contains(reason, "meta/") {
		t.Error("a real entity with a mistyped id must not be told to move to meta/")
	}

	reason, _ = entityNameProblem("6fbj87000ut6-u-id.md")
	if strings.Contains(reason, "rename to") {
		t.Errorf("u has no canonical decode, so no rename should be suggested: %q", reason)
	}
	if !strings.Contains(reason, "replaced deliberately") {
		t.Errorf("the u case should say it needs a deliberate replacement: %q", reason)
	}

	// Both classifications stay ErrValidation so exit-code mapping is unchanged.
	for _, err := range []error{stray, bad} {
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("%v no longer classifies as a validation failure", err)
		}
	}
}

func TestEntityNameProblemGivesIdentityPreservingCaseAdvice(t *testing.T) {
	for _, cand := range []string{strings.ToUpper(goodID), "6Fjvr03mr9zg"} {
		reason, err := entityNameProblem(cand + "-upper-probe.md")
		if !errors.Is(err, errBadEntityID) || !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("case defect lost its validation classification: %v", err)
		}
		if !strings.Contains(reason, "lowercase") || !strings.Contains(reason, goodID+"-upper-probe.md") {
			t.Fatalf("case defect lacks canonical spelling advice: %q", reason)
		}
		for _, wrong := range []string{"Crockford base32 excludes", "u has no canonical", "replaced deliberately", "meta/"} {
			if strings.Contains(reason, wrong) {
				t.Fatalf("misleading case advice %q in %q", wrong, reason)
			}
		}
	}
	for _, cand := range []string{"6fbj87000?t6", "6fbj87000-t6"} {
		reason, _ := entityNameProblem(cand + "-invalid.md")
		if strings.Contains(reason, "u has no canonical") || strings.Contains(reason, "rename to") {
			t.Fatalf("unrecognised character got an invented decode: %q", reason)
		}
	}
}

func TestEntityIDWriteValidationRejectsCaseWithoutNormalizing(t *testing.T) {
	uppercase := strings.ToUpper(goodID)
	for _, err := range []error{validEntityID(uppercase), validateIDUpdate(map[string]any{"id": uppercase})} {
		if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "lowercase") || !strings.Contains(err.Error(), goodID) {
			t.Fatalf("write-time case advice = %v", err)
		}
		if strings.Contains(err.Error(), "excludes") || strings.Contains(err.Error(), "u has no") {
			t.Fatalf("write-time case advice invents an excluded character: %v", err)
		}
	}
	// Mixed case plus a real alias is manual identity-preserving advice, not
	// permission to rewrite an arbitrary invalid character or change Valid.
	reason, _ := entityNameProblem("6FBJ87000LT6-mixed.md")
	if !strings.Contains(reason, "6fbj870001t6-mixed.md") || strings.Contains(reason, "u has no") {
		t.Fatalf("mixed spelling advice = %q", reason)
	}
}
