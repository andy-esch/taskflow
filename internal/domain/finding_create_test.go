package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestCreateFindingAllocatesMonotonicCodeAndInsertsBeforeCandidateSection(t *testing.T) {
	body := "# Audit\n\n## Findings\n\n" +
		"#### H1. First · **Status:** fixed\n\nDone.\n\n" +
		"#### H3. Third · **Status:** open\n\nStill open.\n\n" +
		"## Candidate tasks\n\n" + CandidateTasksMarkerComment() + "\n"
	next, created, err := CreateFinding(body, FindingDraft{
		Band: "h", Title: "New evidence", File: "internal/x.go:12", Component: "core",
		Effort: "xs", Urgency: "Soon", Body: "The invariant is not enforced.",
		Recommendation: "Add the narrow guard.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Code != "H4" || created.Title != "New evidence" || created.Status != "open" {
		t.Fatalf("created finding = %+v", created)
	}
	for _, want := range []string{
		"#### H4. New evidence · **Status:** open",
		"**File:** internal/x.go:12 | **Component:** core\n**Effort:** XS · **Urgency:** soon",
		"The invariant is not enforced.",
		"**Recommendation:** Add the narrow guard.",
	} {
		if !strings.Contains(next, want) {
			t.Errorf("created block missing %q:\n%s", want, next)
		}
	}
	if strings.Index(next, "#### H4.") > strings.Index(next, "## Candidate tasks") {
		t.Fatalf("finding landed after Candidate tasks:\n%s", next)
	}
	if got := len(ParseFindings(next)); got != 3 {
		t.Fatalf("want 3 parsed findings, got %d:\n%s", got, next)
	}
}

func TestCreateFindingUsesSameSpacingForLFAndCRLF(t *testing.T) {
	lf := "## Findings\n\n#### H1. Existing · **Status:** open\n\nEvidence.\n\n## Candidate tasks\n"
	crlf := strings.ReplaceAll(lf, "\n", "\r\n")
	draft := FindingDraft{Band: "H", Title: "New"}
	gotLF, _, err := CreateFinding(lf, draft)
	if err != nil {
		t.Fatal(err)
	}
	gotCRLF, _, err := CreateFinding(crlf, draft)
	if err != nil {
		t.Fatal(err)
	}
	if normalizeNewlines(gotCRLF) != gotLF {
		t.Fatalf("CRLF insertion changed vertical spacing:\nLF:   %q\nCRLF: %q", gotLF, gotCRLF)
	}
}

func TestCreateFindingCreatesMissingFindingsSectionAndIgnoresFencedHeadings(t *testing.T) {
	body := "# Audit\n\n```md\n## Findings\n#### M9. example · **Status:** open\n```\n\n" +
		"## Candidate tasks\n\n" + CandidateTasksMarkerComment() + "\n"
	next, created, err := CreateFinding(body, FindingDraft{
		Band: "M", Title: "Real finding", Body: "A fenced example is fine:\n\n```md\n## Example heading\n**Status:** fixed\n```",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Code != "M1" {
		t.Fatalf("code = %q, want M1", created.Code)
	}
	if got := strings.Count(next, "## Findings"); got != 2 { // one fenced example, one real section
		t.Fatalf("want fenced + real Findings headings, got %d:\n%s", got, next)
	}
	if strings.Index(next, "#### M1.") > strings.Index(next, "## Candidate tasks") {
		t.Fatalf("new Findings section must precede Candidate tasks:\n%s", next)
	}
}

func TestCreateFindingRejectsAmbiguousOrUnsafeInput(t *testing.T) {
	base := "## Findings\n"
	tests := []struct {
		name  string
		body  string
		draft FindingDraft
		want  string
	}{
		{"missing band", base, FindingDraft{Title: "x"}, "band is required"},
		{"unknown band", base, FindingDraft{Band: "P", Title: "x"}, "unknown finding band"},
		{"empty title", base, FindingDraft{Band: "H"}, "title is required"},
		{"title injection", base, FindingDraft{Band: "H", Title: "x\n#### H9. y"}, "title must be one line"},
		{"metadata delimiter", base, FindingDraft{Band: "H", Title: "x", Component: "a | b"}, "delimit metadata fields"},
		{"unknown effort", base, FindingDraft{Band: "H", Title: "x", Effort: "tiny"}, "unknown finding effort"},
		{"unknown urgency", base, FindingDraft{Band: "H", Title: "x", Urgency: "now"}, "unknown finding urgency"},
		{"body heading", base, FindingDraft{Band: "H", Title: "x", Body: "## Escapes"}, "cannot contain an unfenced Markdown heading"},
		{"title metadata injection", base, FindingDraft{Band: "H", Title: "x · **status:** fixed"}, "reserved finding metadata label"},
		{"body metadata injection", base, FindingDraft{Band: "H", Title: "x", Body: "Evidence says **Effort:** L"}, "reserved finding metadata label"},
		{"recommendation metadata injection", base, FindingDraft{Band: "H", Title: "x", Recommendation: "Write **Component:** store"}, "reserved finding metadata label"},
		{"near miss", "## Findings\n\n#### H-1. drift · **Status:** open\n", FindingDraft{Band: "H", Title: "x"}, "run `lint --fix`"},
		{"duplicate code", "## Findings\n\n#### H1. one · **Status:** open\n\n#### H1. two · **Status:** open\n", FindingDraft{Band: "H", Title: "x"}, "duplicate existing finding code"},
		{"multiple sections", "## Findings\n\n## Findings\n", FindingDraft{Band: "H", Title: "x"}, "2 real Findings sections"},
		{"invalid existing finding", "## Findings\n\n#### H1. broken · **Status:** maybe\n", FindingDraft{Band: "H", Title: "x"}, "existing finding H1 is malformed"},
		{"invalid candidate projection", "## Findings\n\n#### H1. one · **Status:** open\n\n## Candidate tasks\n\n" + CandidateTasksMarkerComment() + "\n- ✔ H1 · fixed — stale\n", FindingDraft{Band: "H", Title: "x"}, "Candidate tasks projection is malformed"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := CreateFinding(tc.body, tc.draft)
			if !errors.Is(err, ErrValidation) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("CreateFinding error = %v, want validation containing %q", err, tc.want)
			}
		})
	}
}

func TestCreateFindingPreservesWhitespaceOutsideInsertion(t *testing.T) {
	body := "# Audit\n\n## Findings\n\n#### H1. Existing · **Status:** open\n\nEvidence.\n \n\n## Related\n\nKeep this prose exactly.\n"
	next, _, err := CreateFinding(body, FindingDraft{Band: "H", Title: "New"})
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := "# Audit\n\n## Findings\n\n#### H1. Existing · **Status:** open\n\nEvidence.\n \n\n"
	if !strings.HasPrefix(next, wantPrefix) {
		t.Fatalf("creation changed whitespace before its insertion point:\n%q\nwant prefix:\n%q", next, wantPrefix)
	}
	wantTail := "\n\n## Related\n\nKeep this prose exactly.\n"
	if !strings.HasSuffix(next, wantTail) {
		t.Fatalf("creation changed content after its insertion point:\n%q\nwant suffix:\n%q", next, wantTail)
	}
}

func TestFindingCreationVocabulariesAreStableCopies(t *testing.T) {
	bands := FindingBands()
	efforts := FindingEfforts()
	urgencies := FindingUrgencies()
	if strings.Join(bands, ",") != "H,M,L" || strings.Join(efforts, ",") != "XS,S,M,L" || strings.Join(urgencies, ",") != "acute,soon,eventually" {
		t.Fatalf("unexpected vocabularies: bands=%v efforts=%v urgencies=%v", bands, efforts, urgencies)
	}
	bands[0] = "changed"
	if FindingBands()[0] != "H" {
		t.Fatal("FindingBands exposed mutable registry storage")
	}
}
