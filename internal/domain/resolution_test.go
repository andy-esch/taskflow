package domain

import (
	"strings"
	"testing"
)

// The whole point of the shared pool: a word that means the same thing in both vocabularies
// must be SPELLED the same in both. This is the guard against finding M3 of the
// finding-status audit, where `landed` was legal in code, absent from the docs, and
// contradicted by an assertion — three sources of truth with three answers.
func TestSharedResolutionWordsAgree(t *testing.T) {
	findings := make(map[string]bool)
	for _, s := range FindingStatuses() {
		findings[s] = true
	}
	criteria := make(map[string]bool)
	for _, s := range CriterionStates() {
		criteria[string(s)] = true
	}

	// Declared shared words must exist in BOTH vocabularies, spelled identically.
	declared := make(map[string]bool)
	for _, word := range SharedResolutionWords() {
		declared[word] = true
		if !findings[word] {
			t.Errorf("shared word %q is not a finding status — the two have drifted", word)
		}
		if !criteria[word] {
			t.Errorf("shared word %q is not a criterion state — the two have drifted", word)
		}
	}

	// …and the reverse, which is what makes this test more than a restatement of its own
	// constant: ANY word the two vocabularies happen to share must be DECLARED shared.
	// Without this, adding `superseded` to criteria (or `n/a` to findings) would create an
	// undeclared overlap that nothing forces anyone to think about — exactly the silent
	// drift the shared pool exists to prevent, and the reason the first version of this
	// test was tautological.
	for word := range findings {
		if criteria[word] && !declared[word] {
			t.Errorf("%q appears in BOTH vocabularies but is not in SharedResolutionWords — "+
				"either declare it shared or give the two concepts different words", word)
		}
	}

	// A word that is deliberately NOT shared must stay unshared, so an accidental
	// convergence is caught rather than silently blessed.
	for _, only := range []struct {
		word, vocab string
		in          map[string]bool
		notIn       map[string]bool
	}{
		{"superseded", "findings", findings, criteria},
		{"fixed", "findings", findings, criteria},
		{"met", "criteria", criteria, findings},
		{"n/a", "criteria", criteria, findings},
	} {
		if !only.in[only.word] {
			t.Errorf("%q is expected in the %s vocabulary and is missing", only.word, only.vocab)
		}
		if only.notIn[only.word] {
			t.Errorf("%q crossed into the other vocabulary without being declared shared", only.word)
		}
	}
}

// Decoration is handled on purpose rather than discovered in an audit later: this repo's
// own candidate lists use ✅ ⏳ ⛔, and finding M2 records an emoji being captured AS the
// value.
func TestParseCriterionStateToleratesCaseSpaceAndDecoration(t *testing.T) {
	for _, in := range []string{"deferred", "  Deferred ", "⏳ deferred", "✅deferred"} {
		if got, ok := ParseCriterionState(in); !ok || got != CriterionDeferred {
			t.Errorf("ParseCriterionState(%q) = %q,%v; want deferred,true", in, got, ok)
		}
	}
	for _, in := range []string{"met", "not met", "superseded", "nonsense", ""} {
		if got, ok := ParseCriterionState(in); ok {
			// met/not met are what the BRACKET says; allowing them as suffixes too would
			// create two spellings of one thing and a contradiction to lint for.
			t.Errorf("ParseCriterionState(%q) = %q,true; want it rejected as a suffix", in, got)
		}
	}
}

// Only the non-binary states need a why, and only CriterionMet counts as done — a deferred
// criterion is explicitly NOT satisfied, which is the entire point of saying so.
func TestCriterionStateSemantics(t *testing.T) {
	for _, c := range []CriterionState{CriterionDeferred, CriterionWontFix, CriterionNA} {
		if !c.NeedsReason() {
			t.Errorf("%q should require a reason", c)
		}
		if c.Met() {
			t.Errorf("%q must not count as met", c)
		}
	}
	for _, c := range []CriterionState{CriterionMet, CriterionUnmet} {
		if c.NeedsReason() {
			t.Errorf("%q should not require a reason", c)
		}
	}
	if !CriterionMet.Met() || CriterionUnmet.Met() {
		t.Error("Met() must be true for met and false for not met")
	}
}

// Every rejection names the legal set, so a reader can self-correct without the source —
// the omission finding M1 flagged for findings.
func TestInvalidCriterionStateErrorNamesTheLegalSet(t *testing.T) {
	err := InvalidCriterionStateError("bogus")
	for _, want := range append(CriterionSuffixStates(), "bogus") {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error omits %q: %v", want, err)
		}
	}
}
