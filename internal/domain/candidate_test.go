package domain

import (
	"errors"
	"strings"
	"testing"
)

func managedCandidateBody(rows string) string {
	return "#### H1. First · **Status:** open\n\n" +
		"#### M1. Second · **Status:** deferred (later)\n\n" +
		"## Candidate tasks\n\n" + CandidateTasksMarkerComment() + "\n" +
		"<!-- tool-owned -->\n\n" + rows
}

func TestCandidateTasksMarkerDerivedFromFindingStatuses(t *testing.T) {
	for _, spec := range findingStatusSpecs {
		want := spec.Glyph + " " + spec.Status
		if !strings.Contains(CandidateTasksMarkerComment(), want) {
			t.Errorf("marker missing derived status token %q: %s", want, CandidateTasksMarkerComment())
		}
		if got := FindingStatusGlyph(spec.Status); got != spec.Glyph {
			t.Errorf("FindingStatusGlyph(%q)=%q want %q", spec.Status, got, spec.Glyph)
		}
	}
}

func TestLintCandidateTasksLegacyIsDeliberatelyIgnored(t *testing.T) {
	body := "#### H1. First · **Status:** fixed\n\n## Candidate tasks\n\n- ⏳ free-form legacy text — H1\n"
	if issues := LintCandidateTasks(body, ParseFindings(body)); len(issues) != 0 {
		t.Fatalf("legacy candidate prose must be tolerated, got %+v", issues)
	}
}

func TestLintCandidateTasksManagedRows(t *testing.T) {
	clean := managedCandidateBody("- ○ H1 · open — Create the first task\n- ◌ M1 · deferred — Revisit later\n")
	if issues := LintCandidateTasks(clean, ParseFindings(clean)); len(issues) != 0 {
		t.Fatalf("canonical managed rows should lint clean, got %+v", issues)
	}

	dirty := managedCandidateBody("- ✔ H1 · fixed — stale status\n- ○ H1 · open — duplicate\n- ✔ M1 · deferred — wrong glyph\n- ◌ Z9 · deferred — missing finding\n- ???\n")
	issues := LintCandidateTasks(dirty, ParseFindings(dirty))
	for _, want := range []string{"disagrees with finding status", "expected \"◌\"", "2 candidate rows", "references no parsed finding", "not a canonical candidate row"} {
		found := false
		for _, issue := range issues {
			if strings.Contains(issue.Message, want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing issue containing %q: %+v", want, issues)
		}
	}
}

func TestSetFindingCandidateAddReplaceRemoveAndStatusSync(t *testing.T) {
	body := managedCandidateBody("") + "\n## Related\n\n- context\n"
	got, err := SetFindingCandidate(body, "H1", "Create a focused follow-up")
	if err != nil {
		t.Fatalf("add candidate: %v", err)
	}
	if !strings.Contains(got, "- ○ H1 · open — Create a focused follow-up") {
		t.Fatalf("canonical row not added:\n%s", got)
	}
	if strings.Index(got, "- ○ H1") > strings.Index(got, "## Related") {
		t.Fatalf("candidate escaped its section:\n%s", got)
	}
	if !strings.HasSuffix(got, "Create a focused follow-up\n\n## Related\n\n- context\n") {
		t.Fatalf("candidate insertion should normalize the following section gap:\n%s", got)
	}

	got, err = SetFindingCandidate(got, "H1", "Use the narrower repair")
	if err != nil {
		t.Fatalf("replace candidate: %v", err)
	}
	if strings.Count(got, " H1 · ") != 1 || !strings.Contains(got, "— Use the narrower repair") {
		t.Fatalf("candidate should replace rather than duplicate:\n%s", got)
	}

	got, err = SetFindingStatus(got, "H1", "fixed 2026-09-19")
	if err != nil {
		t.Fatalf("set status: %v", err)
	}
	if !strings.Contains(got, "**Status:** fixed 2026-09-19") || !strings.Contains(got, "- ✔ H1 · fixed — Use the narrower repair") {
		t.Fatalf("finding and managed mirror did not update atomically:\n%s", got)
	}

	got, err = SetFindingCandidate(got, "H1", "")
	if err != nil {
		t.Fatalf("remove candidate: %v", err)
	}
	if strings.Contains(got, " H1 · ") {
		t.Fatalf("candidate row should be removed:\n%s", got)
	}
}

func TestSetFindingCandidateAddRemoveCyclesAreByteStable(t *testing.T) {
	original := strings.TrimRight(managedCandidateBody(""), "\n") + "\n\n## Related\n\n- context\n"
	got := original
	for i := 0; i < 5; i++ {
		var err error
		got, err = SetFindingCandidate(got, "H1", "repeat")
		if err != nil {
			t.Fatalf("cycle %d add: %v", i, err)
		}
		got, err = SetFindingCandidate(got, "H1", "")
		if err != nil {
			t.Fatalf("cycle %d remove: %v", i, err)
		}
	}
	if got != original {
		t.Fatalf("repeated add/remove cycles accumulated whitespace:\n--- got ---\n%q\n--- want ---\n%q", got, original)
	}
}

func TestSetFindingCandidateRefusesLegacyAndUnsafeText(t *testing.T) {
	legacy := "#### H1. First · **Status:** open\n\n## Candidate tasks\n\n- old prose\n"
	if _, err := SetFindingCandidate(legacy, "H1", "new"); !errors.Is(err, ErrValidation) || !strings.Contains(err.Error(), "legacy") {
		t.Fatalf("legacy section should be refused with a remedy, got %v", err)
	}
	managed := managedCandidateBody("")
	if _, err := SetFindingCandidate(managed, "H1", "bad\n- ○ M9 · open — injected"); !errors.Is(err, ErrValidation) {
		t.Fatalf("multiline candidate should be rejected, got %v", err)
	}
	if _, err := SetFindingCandidate(managed, "Z9", "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown finding should be not-found, got %v", err)
	}
	withoutSection := "#### H1. First · **Status:** open\n"
	if _, err := SetFindingCandidate(withoutSection, "H1", "missing section"); !errors.Is(err, ErrValidation) || !strings.Contains(err.Error(), "no Candidate tasks section") {
		t.Fatalf("missing managed section should have a specific remedy, got %v", err)
	}
}

func TestLintCandidateTasksFlagsMarkerDriftAndMultipleManagedSections(t *testing.T) {
	body := managedCandidateBody("")
	drifted := strings.Replace(body, "○ open", "✔ open", 1)
	if issues := LintCandidateTasks(drifted, ParseFindings(drifted)); len(issues) != 1 || !strings.Contains(issues[0].Message, "marker legend drifted") {
		t.Fatalf("drifted generated legend should be one focused issue, got %+v", issues)
	}

	twice := body + "\n## Candidate tasks\n\n" + CandidateTasksMarkerComment() + "\n"
	if issues := LintCandidateTasks(twice, ParseFindings(twice)); len(issues) != 1 || !strings.Contains(issues[0].Message, "2 managed") {
		t.Fatalf("multiple managed sections should be refused unambiguously, got %+v", issues)
	}
}

func TestSetFindingStatusLeavesLegacyCandidateProseUntouched(t *testing.T) {
	body := "#### H1. First · **Status:** open\n\n## Candidate tasks\n\n- ⏳ H1 — handwritten legacy row\n"
	got, err := SetFindingStatus(body, "H1", "fixed")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "**Status:** fixed") || !strings.Contains(got, "- ⏳ H1 — handwritten legacy row") {
		t.Fatalf("status should change without guessing at legacy prose:\n%s", got)
	}
}

func TestManagedCandidateSectionEndsBeforeAppendedFinding(t *testing.T) {
	// `audit append` historically adds a new finding after the scaffold's final Candidate
	// tasks section. Any heading ends the row-only managed section, so that supported write
	// path does not turn the finding body into malformed candidate prose.
	body := "## Candidate tasks\n\n" + CandidateTasksMarkerComment() + "\n\n" +
		"#### H1. Appended finding · **Status:** open\n\nEvidence.\n"
	findings := ParseFindings(body)
	if issues := LintCandidateTasks(body, findings); len(issues) != 0 {
		t.Fatalf("appended finding should sit beyond the managed row section, got %+v", issues)
	}
	got, err := SetFindingCandidate(body, "H1", "Create a follow-up")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(got, "- ○ H1") > strings.Index(got, "#### H1.") {
		t.Fatalf("candidate row should be inserted before the appended finding:\n%s", got)
	}
}

func TestLintCandidateTasksIgnoresFencedManagedExample(t *testing.T) {
	body := "#### H1. First · **Status:** open\n\n```md\n## Candidate tasks\n" + CandidateTasksMarkerComment() + "\n- ✘ H1 · wontfix — example\n```\n"
	if issues := LintCandidateTasks(body, ParseFindings(body)); len(issues) != 0 {
		t.Fatalf("fenced candidate example must not activate managed lint, got %+v", issues)
	}
}

func TestLintCandidateTasksIgnoresFenceInsideManagedSection(t *testing.T) {
	body := managedCandidateBody("```sh\ntskflwctl audit finding sample H1 --candidate \"do it\"\n```\n\n- ○ H1 · open — Real row\n")
	if issues := LintCandidateTasks(body, ParseFindings(body)); len(issues) != 0 {
		t.Fatalf("fenced prose inside a managed section must not activate row lint, got %+v", issues)
	}
	got, err := SetFindingCandidate(body, "H1", "Updated row")
	if err != nil {
		t.Fatalf("fenced prose must not block a managed write: %v", err)
	}
	if !strings.Contains(got, "- ○ H1 · open — Updated row") || !strings.Contains(got, "```sh") {
		t.Fatalf("write should replace the real row and preserve the fence:\n%s", got)
	}
}

func TestLintCandidateTasksDistinguishesLegacyCorruptionAndUnknownVersions(t *testing.T) {
	finding := "#### H1. First · **Status:** open\n\n## Candidate tasks\n\n"
	legacy := finding + "- ⏳ H1 — handwritten legacy row\n"
	if issues := LintCandidateTasks(legacy, ParseFindings(legacy)); len(issues) != 0 {
		t.Fatalf("ordinary legacy prose remains ignored, got %+v", issues)
	}

	for _, marker := range []string{
		"<!--candidate-tasks:v1 · damaged -->",
		"<!-- candidate-task:v1 · damaged -->",
	} {
		body := finding + marker + "\n- ○ H1 · open — Canonical row whose marker was damaged\n"
		issues := LintCandidateTasks(body, ParseFindings(body))
		if len(issues) != 1 || !strings.Contains(issues[0].Message, "unmanaged Candidate tasks section contains 1 canonical row") {
			t.Errorf("damaged marker %q should be diagnosed without guessing at legacy prose, got %+v", marker, issues)
		}
	}

	for _, version := range []string{"candidate-tasks:v2", "candidate-tasks:v10", "candidate-tasks:v1-draft"} {
		body := finding + "<!-- " + version + " · future -->\n- ○ H1 · open — Future row\n"
		issues := LintCandidateTasks(body, ParseFindings(body))
		if len(issues) != 1 || !strings.Contains(issues[0].Message, "unsupported Candidate tasks marker") || !strings.Contains(issues[0].Message, version) {
			t.Errorf("unknown version %q should be distinguished from v1 drift, got %+v", version, issues)
		}
	}
}
